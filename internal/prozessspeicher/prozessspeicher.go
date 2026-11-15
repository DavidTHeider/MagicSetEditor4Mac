// Package prozessspeicher misst den Speicher dieses Programms — **und den
// seiner WebKit-Prozesse**, die außerhalb liegen — OPS02
// (#2348),
// Befund A-07 des Audits 0.9.0.
//
// # Warum es dieses Paket gibt
//
// **Der größte Verbraucher war unsichtbar.** Die Metriken nannten den Go-Heap;
// der Webprozess des Fensters hält im Dauerlauf 450 bis 630 MB
// (`docs/belege/huellenprobe/dauerbetrieb.md`) und gehört nicht zum
// Prozess, denn WebKit startet ihn über launchd. `ps` und `footprint` sagen
// nicht, wem er gehört.
//
// # Wie es zuordnet
//
// **Über den verantwortlichen Prozess**, den macOS jedem Prozess mitgibt und
// den die Aktivitätsanzeige zeigt (`responsibility_get_pid_responsible_for_pid`
// aus libSystem — am 2026-10-15 auf macOS 27.0.1 gemessen: WebContent und GPU
// gehören der Anwendung, die die Webansicht hält, auch
// `WebContent.EnhancedSecurity`). **Die Funktion ist nicht öffentlich**; fehlt
// sie, wird der Webspeicher **als nicht gemessen** genannt und nicht als null.
//
// # Was gemessen wird
//
// **Der physische Fußabdruck** (`ri_phys_footprint` aus `proc_pid_rusage`,
// `RUSAGE_INFO_V4`, nachgelesen in `sys/resource.h` des SDK) — die Zahl, die
// auch `footprint` und die Aktivitätsanzeige nennen, nicht der residente
// Speicher.
//
// # Ein Blatt (§2)
//
// Es kennt **nichts** aus diesem Projekt.
package prozessspeicher

import (
	"errors"
	"fmt"
	"strings"
)

// Art sagt, was für ein Prozess es ist.
type Art string

// Die Arten.
const (
	ArtAnwendung Art = "anwendung"
	ArtWebinhalt Art = "webinhalt"
	ArtGrafik    Art = "gpu"
	ArtNetz      Art = "netz"
)

// Prozess ist ein gemessener Prozess.
type Prozess struct {
	PID              int
	Art              Art
	FussabdruckBytes uint64
}

// Stand ist der Speicher dieses Programms und seiner WebKit-Prozesse.
type Stand struct {
	Anwendung Prozess
	Web       []Prozess

	// Ungemessen nennt, warum die WebKit-Prozesse nicht gemessen sind — leer,
	// wenn sie es sind. **Leer und ohne Prozesse heißt: es gibt keine.**
	Ungemessen string
}

// quelle sind die Rufe ans System — eingesetzt, damit die Zuordnung ohne
// echte WebKit-Prozesse prüfbar ist.
type quelle struct {
	pids           func() ([]int, error)
	pfad           func(pid int) (string, error)
	verantwortlich func(pid int) (int, error)
	fussabdruck    func(pid int) (uint64, error)

	// lebt sagt, ob es einen Prozess noch gibt — auch einen fremden.
	lebt func(pid int) bool
}

// errNichtGemessen sagt, dass dieses System den Speicher nicht hergibt.
var errNichtGemessen = errors.New("der Speicher der Prozesse lässt sich auf diesem System nicht messen")

// artAus ordnet einen Programmpfad einer Art zu — leer heißt: kein
// WebKit-Prozess.
func artAus(pfad string) Art {
	name := pfad[strings.LastIndex(pfad, "/")+1:]

	switch {
	case strings.HasPrefix(name, "com.apple.WebKit.WebContent"):
		return ArtWebinhalt
	case strings.HasPrefix(name, "com.apple.WebKit.GPU"):
		return ArtGrafik
	case strings.HasPrefix(name, "com.apple.WebKit.Networking"):
		return ArtNetz
	}

	return ""
}

// messenUeber misst `pid` und die WebKit-Prozesse, für die er verantwortlich
// ist.
//
// **Der eigene Fußabdruck muss gelingen** — ohne ihn gibt es keinen Stand.
func messenUeber(q quelle, pid int) (Stand, error) {
	eigen, fehler := q.fussabdruck(pid)
	if fehler != nil {
		return Stand{}, fmt.Errorf("den eigenen Speicher messen: %w", fehler)
	}

	web, ungemessen := webprozesse(q, pid)

	return Stand{Anwendung: Prozess{PID: pid, Art: ArtAnwendung, FussabdruckBytes: eigen}, Web: web, Ungemessen: ungemessen}, nil
}

// webprozesse gibt die WebKit-Prozesse, für die `pid` verantwortlich ist —
// und, warum sie sich nicht oder nicht alle fragen ließen. **Ein einzelner
// darf fehlen, wenn er verschwand**: er kann zwischen Aufzählen und Messen
// enden, und dann zählt er nicht.
func webprozesse(q quelle, pid int) (web []Prozess, ungemessen string) {
	pids, fehler := q.pids()
	if fehler != nil {
		return nil, fehler.Error()
	}

	var unvollstaendig error

	for _, kandidat := range pids {
		if kandidat == pid {
			continue
		}

		// **Zuerst, wer verantwortlich ist** (Nachlauf über Phase 4, Befund
		// 10): ein Pfadfehler eines fremden Prozesses ist kein Befund, einer
		// eines eigenen Kindes eine Lücke. `responsibility_get_…` gibt für
		// einen fremden einen anderen Prozess oder −1, einen Fehler nur aus
		// dem Riegel.
		verantwortlich, fehler := q.verantwortlich(kandidat)
		if fehler != nil {
			return nil, fehler.Error()
		}

		if verantwortlich != pid {
			continue
		}

		pfad, pfadfehler := q.pfad(kandidat)

		// **Ein Fehler, der jeden trifft, ist keine Antwort** (Zweitblick
		// über Phase 4, Befund 9): ein fehlendes Symbol fängt der Riegel,
		// und hier verschwand es — die Oberfläche las „keine Webansicht".
		if errors.Is(pfadfehler, errNichtGemessen) {
			return nil, pfadfehler.Error()
		}

		// **Ein eigenes Kind ohne Pfad, das lebt, ist eine Lücke** — eines,
		// das verschwand, zählt nicht.
		if pfadfehler != nil {
			if q.lebt(kandidat) && unvollstaendig == nil {
				unvollstaendig = pfadfehler
			}

			continue
		}

		art := artAus(pfad)
		if art == "" {
			continue
		}

		bytes, fehler := q.fussabdruck(kandidat)
		if fehler == nil {
			web = append(web, Prozess{PID: kandidat, Art: art, FussabdruckBytes: bytes})

			continue
		}

		// **Nur ein verschwundener zählt nicht** — einer, der lebt und sich
		// nicht messen lässt, macht die Summe unvollständig. Genannt wird der
		// erste Grund: derselbe Satz je Prozess wäre kein zweiter.
		if (errors.Is(fehler, errNichtGemessen) || q.lebt(kandidat)) && unvollstaendig == nil {
			unvollstaendig = fehler
		}
	}

	if unvollstaendig != nil {
		return web, unvollstaendig.Error()
	}

	return web, ""
}

// Messen misst `pid` und seine WebKit-Prozesse.
func Messen(pid int) (Stand, error) { return messenUeber(echteQuelle(), pid) }
