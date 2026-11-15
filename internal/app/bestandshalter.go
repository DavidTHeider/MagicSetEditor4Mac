package app

// **Wer den Bestand sonst offen hält** — M02
// (#2346),
// Befund A-11 des Audits 0.9.0.
//
// Ein Programm, das den Bestand beim Tausch des Bündels offen hält, läuft
// danach mit dem Code von vorher weiter. Die Vorprüfung vor einem Wechsel
// nennt es deshalb als Hindernis — **gefragt bei `lsof`**, denn der Bestand
// selbst weiß nicht, wer ihn hält, und `AlleinAmBestand` sagt nur, dass
// jemand es tut. **Nicht genannt** werden dieser Prozess und der eingerichtete
// Dienst: ihn startet die neue Fassung selbst neu (`dienstnachdemaufstieg.go`).
//
// **Was das nicht kann**: `lsof` sieht nur Prozesse desselben Nutzers; ein
// fremdes Konto, das den Bestand liest, bleibt ungenannt.

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// bestandshalter ist die Halterquelle der Vorprüfung.
type bestandshalter struct {
	pfad   string
	fragen func(ctx context.Context, pfad string) (string, error)
	dienst func(ctx context.Context) ([]int, error)
}

// bestandshalterAn fragt wirklich.
func bestandshalterAn(pfad string) bestandshalter {
	return bestandshalter{pfad: pfad, fragen: lsofFragen, dienst: dienstPIDs}
}

// Halter nennt jedes andere Programm samt Prozessnummer — `mse-rest (4711)`.
func (b bestandshalter) Halter(ctx context.Context) ([]string, error) {
	ausgabe, fehler := b.fragen(ctx, b.pfad)
	if fehler != nil {
		return nil, fmt.Errorf("fragen, wer den Bestand offen hält: %w", fehler)
	}

	// **Ohne Nachweis über den Dienst steht er mit da** — lieber ein Halter
	// zu viel genannt als ein laufender verschwiegen.
	eigene := os.Getpid()
	dienste, _ := b.dienst(ctx) //nolint:errcheck // siehe darüber

	var (
		halter []string
		pid    int
	)

	for _, zeile := range splitZeilen(ausgabe) {
		switch {
		case strings.HasPrefix(zeile, "p"):
			pid, _ = strconv.Atoi(zeile[1:]) //nolint:errcheck // keine Zahl heißt 0, und 0 hält nichts
		case strings.HasPrefix(zeile, "c") && pid != eigene && !slices.Contains(dienste, pid):
			halter = append(halter, fmt.Sprintf("%s (%d)", zeile[1:], pid))
		}
	}

	return halter, nil
}

// halterfrist ist, wie lange `lsof` und `launchctl print` antworten dürfen.
const halterfrist = 5 * time.Second

// lsofFragen fragt `lsof` in seiner Form für Maschinen: je Prozess `p<pid>`
// und `c<befehl>`. **Hält niemand die Datei, endet es mit 1, ohne Ausgabe und
// ohne Meldung** — das ist eine Antwort und kein Fehler. **Mit Meldung ist es
// einer** (Zweitblick über Phase 2, Befund 9): `lsof` endet bei jedem Fehler
// mit 1 (`man lsof`, „DIAGNOSTICS", nachgelesen am 2026-10-14), und eine
// gescheiterte Frage hielt sonst „niemand" für die Antwort. `-w` schaltet
// die Warnungen ab (`lsof -h`: „+|-w Warnings (+)", lsof 4.91 auf macOS 27,
// am 2026-10-14 gefragt), damit auf der Fehlerausgabe nur Fehler stehen.
func lsofFragen(ctx context.Context, pfad string) (string, error) {
	ctx, ende := context.WithTimeout(ctx, halterfrist)
	defer ende()

	// #nosec G204 -- das Programm ist ein Literal, der Pfad der des Bestands
	ausgabe, fehler := exec.CommandContext(ctx, "/usr/sbin/lsof", "-w", "-F", "pc", "--", pfad).Output()

	var beendet *exec.ExitError
	if !errors.As(fehler, &beendet) {
		return string(ausgabe), fehler
	}

	meldung := strings.TrimSpace(string(beendet.Stderr))
	if beendet.ExitCode() == 1 && len(ausgabe) == 0 && meldung == "" && ctx.Err() == nil {
		return "", nil
	}

	return string(ausgabe), fmt.Errorf("%w: %s", fehler, meldung)
}

// dienstPIDs sind die Prozessnummern des eingerichteten Dienstes — **unter
// jedem Etikett, unter dem er läuft**, leer, wenn keiner läuft. Gefragt bei
// `launchd` und nicht verändert.
func dienstPIDs(ctx context.Context) ([]int, error) {
	laeufe, fehler := laufendeDienste(ctx)

	pids := make([]int, 0, len(laeufe))
	for _, lauf := range laeufe {
		pids = append(pids, lauf.pid)
	}

	return pids, fehler
}

// dienstprozess ist ein laufender Dienst: sein Etikett und seine Nummer.
type dienstprozess struct {
	etikett string
	pid     int
}

// laufendeDienste nennt den eingerichteten Dienst **unter dem neuen Etikett
// und unter dem alten** (R02, #2334): wandert nach einem Aufstieg zuerst die
// Kommandozeile, hat das Fenster ihn noch nicht umgezogen, und ein Dienst aus
// 0.8.0 läuft unter dem alten Etikett weiter.
//
// **Jeden, nicht den ersten** (Zweitblick über Phase 3, Befund 8): stehen
// beide da, kann der neue einen anderen Bestand bedienen und der alte diesen
// halten. Ein Fehler beim Fragen bricht ab — ein halber Nachweis ist kein Nein.
func laufendeDienste(ctx context.Context) ([]dienstprozess, error) {
	var laeufe []dienstprozess

	for _, etikett := range []string{dienstlabel, altdienstlabel} {
		pid, fehler := launchctlPID(ctx, etikett)
		if fehler != nil {
			return laeufe, fehler
		}

		if pid != 0 {
			laeufe = append(laeufe, dienstprozess{etikett: etikett, pid: pid})
		}
	}

	return laeufe, nil
}

// launchctlDrucken fragt `launchctl print` nach einem Ziel — eine Naht für
// den Lauf, der eine Antwort über einen laufenden Dienst braucht. **Das
// Programm bleibt ein Literal** (`internal/schichten`, `unterprozess_test.go`).
var launchctlDrucken = func(ctx context.Context, ziel string) ([]byte, error) {
	// #nosec G204 -- das Programm ist ein Literal, das Ziel setzt diese Anwendung zusammen
	return exec.CommandContext(ctx, "/bin/launchctl", "print", ziel).Output()
}

// nichtGeladenCode ist, womit `launchctl print` einen Dienst meldet, den es
// nicht kennt — 113, am 2026-10-14 gefragt („Could not find service").
const nichtGeladenCode = 113

// launchctlPID fragt `launchctl print` nach einem Etikett.
//
// **Ein Nachweis, der scheitert, ist kein „läuft nicht"** — Nachlauf über
// Phase 2 von 0.9.0, Befund 1 (#2323): bis dahin hieß jeder Fehler und jede
// Ausgabe ohne Prozessnummer „kein Dienst", und gewandert wurde ohne
// Nachweis. Nur der Ausgang 113 heißt „nicht geladen".
func launchctlPID(ctx context.Context, etikett string) (int, error) {
	ctx, ende := context.WithTimeout(ctx, halterfrist)
	defer ende()

	ausgabe, fehler := launchctlDrucken(ctx, "gui/"+strconv.Itoa(os.Getuid())+"/"+etikett)

	var beendet *exec.ExitError
	if errors.As(fehler, &beendet) && beendet.ExitCode() == nichtGeladenCode && ctx.Err() == nil {
		return 0, nil
	}

	if fehler != nil {
		return 0, fmt.Errorf("launchctl print: %w", fehler)
	}

	return pidAus(string(ausgabe))
}

// pidMuster und zustandsmuster lesen `pid = 4711` und `state = running` aus
// der Ausgabe von `launchctl print` — **die erste Fundstelle** ist die des
// Dienstes selbst; tiefer eingerückte gehören seinen Teilen.
var (
	pidMuster      = regexp.MustCompile(`(?m)^\s*pid = (\d+)$`)
	zustandsmuster = regexp.MustCompile(`(?m)^\s*state = (.+)$`)
)

// pidAus gibt die Prozessnummer — 0 für einen geladenen Dienst, der nicht
// läuft. **`print` ist nach seiner Handbuchseite keine Schnittstelle**: eine
// Ausgabe ohne Zustand oder ein laufender Dienst ohne lesbare Nummer ist
// deshalb ein Fehler und kein Nein.
func pidAus(ausgabe string) (int, error) {
	zustand := zustandsmuster.FindStringSubmatch(ausgabe)
	if zustand == nil {
		return 0, errors.New("launchctl print nennt keinen Zustand")
	}

	if treffer := pidMuster.FindStringSubmatch(ausgabe); treffer != nil {
		pid, _ := strconv.Atoi(treffer[1]) //nolint:errcheck // das Muster sind Ziffern

		return pid, nil
	}

	if strings.TrimSpace(zustand[1]) == "running" {
		return 0, errors.New("launchctl print nennt einen laufenden Dienst ohne Prozessnummer")
	}

	return 0, nil
}

// splitZeilen zerlegt eine Ausgabe in Zeilen.
func splitZeilen(text string) []string { return strings.Split(text, "\n") }
