package laufbeleg

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

// Einordnung sagt, ob ein Beleg zum heutigen Stand passt — QA-01.AK-04:
// „Überfällige oder zum Commit unpassende Nachweise sind sichtbar".
type Einordnung string

// Die vier Einordnungen.
//
// **`OhneEingaben` ist kein »aktuell«**: der Inhalt passt, aber ob die
// Pflichteingaben noch dieselben sind, hat niemand gerechnet.
const (
	Aktuell      Einordnung = "aktuell"
	AndererStand Einordnung = "anderer Stand"
	Ueberfaellig Einordnung = "überfällig"
	OhneEingaben Einordnung = "ohne Eingaben"
)

// Heute ist der Stand, gegen den ein Beleg gehalten wird — **alles, was in
// eine Messung eingeht**, und nicht nur der Index.
type Heute struct {
	// Inhalt ist der Hash über das, was der Lauf misst — gerechnet wie beim
	// Messen, also über den Index.
	Inhalt string
	// Offen sind die Zeilen von `git status --porcelain` über dieselben
	// Pfade: was im Arbeitsbaum anders ist als im Index, und was dort neu
	// liegt. **Gebaut wird aus dem Arbeitsbaum**, und der Index sähe es nicht.
	Offen []string
	// Eingaben sind die Pflichteingaben des Plans, heute erhoben.
	Eingaben []Eingabe
	// OhneEingaben heißt: sie wurden mit Absicht nicht erhoben — die
	// Reichweite am Ende von `make ci` kennt sie nicht. **Die Schalter dann
	// auch nicht**: sie kommen aus derselben Umgebung.
	OhneEingaben bool
	// Werkzeuge sind die Fassungen, mit denen heute gemessen würde — `go`
	// und `plattform`, wie der Beleg sie führt.
	Werkzeuge map[string]string
	// Schalter sind die Umgebungswerte aus dem Plan, heute gelesen.
	Schalter map[string]string
}

// verglicheneWerkzeuge sind die Fassungen, die zum Stand gehören — **die
// Systemfassung nicht**: gerastert wird in Go, und eine Aktualisierung von
// macOS änderte sonst jeden Beleg.
var verglicheneWerkzeuge = []string{"go", "plattform"}

// Ordne sagt, ob ein Beleg zum heutigen Inhalt passt.
//
// **Verglichen wird der Inhalt und nicht der Commit** — `heute.Inhalt` ist
// der Hash über das, was der Lauf misst, gerechnet wie beim Messen. Derselbe
// Inhalt an einem anderen Commit ist derselbe Stand; ein Squash-Merge ändert
// den Commit und sonst nichts.
//
// **Ein Beleg von einem Baum mit ungesicherten Änderungen passt zu keinem
// Stand** — welcher gemessen wurde, weiß niemand. **Ein heutiger Baum mit
// solchen Änderungen passt zu keinem Beleg**, aus demselben Grund in der
// anderen Richtung. **Und eine Eingabe, die heute anders ist oder sich nicht
// erheben lässt, ist ein anderer Stand** — bis zum Zweitblick über Phase 6.0
// sah die Prüfung weder das eine noch das andere. **`frist` null heißt:
// keine Frist.**
func Ordne(beleg Beleg, heute Heute, jetzt time.Time, frist time.Duration) (einordnung Einordnung, grund string) {
	kurz := kurzform(beleg.Commit)

	switch {
	case !beleg.BaumSauber:
		return AndererStand, "gemessen an " + kurz +
			", aber an einem Baum mit ungesicherten Änderungen"
	case beleg.Inhalt == "":
		return AndererStand, "ein Beleg ohne Inhaltshash — ob er passt, sagt nichts"
	case len(heute.Offen) > 0:
		return AndererStand, "der Arbeitsbaum weicht ab, wo der Lauf misst: " + pfadeAus(heute.Offen)
	case beleg.Inhalt != heute.Inhalt:
		return AndererStand, "seit " + kurz + " hat sich geändert, was der Lauf misst"
	}

	if anders := anderesWerkzeug(beleg.Werkzeuge, heute.Werkzeuge); anders != "" {
		return AndererStand, anders
	}

	if !heute.OhneEingaben {
		if anders := andereEingabe(beleg.Eingaben, heute.Eingaben); anders != "" {
			return AndererStand, anders
		}

		if anders := andererSchalter(beleg.Schalter, heute.Schalter); anders != "" {
			return AndererStand, anders
		}
	}

	if frist > 0 && jetzt.Sub(beleg.Ende) > frist {
		return Ueberfaellig, fmt.Sprintf("gemessen am %s, älter als die Frist von %s",
			beleg.Ende.Format(time.DateOnly), frist)
	}

	gemessen := "gemessen an " + kurz + " am " + beleg.Ende.Format(time.DateOnly)
	if heute.OhneEingaben {
		return OhneEingaben, gemessen + " — ob die Eingaben dieselben sind, sagt make belege"
	}

	return Aktuell, gemessen
}

// andereEingabe nennt die erste Eingabe, die heute nicht dieselbe ist.
//
// **Was beim Messen fehlte und heute noch fehlt, ist dasselbe** — beide ohne
// Hash. Ein BLOCKED bleibt so ein BLOCKED, statt zusätzlich »anders« zu heißen.
func andereEingabe(damals, heute []Eingabe) string {
	jetzt := make(map[string]Eingabe, len(heute))
	for _, eingabe := range heute {
		jetzt[eingabe.Name] = eingabe
	}

	for _, eingabe := range damals {
		heutige, erhoben := jetzt[eingabe.Name]

		switch {
		case !erhoben:
			return "die Eingabe " + eingabe.Name + " ist heute nicht zu prüfen: nicht erhoben"
		case heutige.Fehlt != "" && eingabe.Fehlt == "":
			return "die Eingabe " + eingabe.Name + " ist heute nicht zu prüfen: " + heutige.Fehlt
		case heutige.Hash != eingabe.Hash:
			return "die Eingabe " + eingabe.Name + " hat heute einen anderen Inhalt"
		}
	}

	return ""
}

// genannt ist, wie viele geänderte Pfade eine Einordnung beim Namen nennt.
const genannt = 5

// anderesWerkzeug nennt die erste Fassung, die heute eine andere ist —
// **gefragt wird nur, was beide Seiten kennen**: ein Beleg ohne Plattform
// sagt über sie nichts, und ein heutiger Stand ohne Werkzeuge ebenso.
func anderesWerkzeug(damals, heute map[string]string) string {
	for _, name := range verglicheneWerkzeuge {
		alt, neu := damals[name], heute[name]
		if alt != "" && neu != "" && alt != neu {
			return "gemessen mit " + name + " " + alt + ", heute " + neu
		}
	}

	return ""
}

// andererSchalter nennt den ersten Schalter, der heute anders steht — über
// beide Seiten, **ein fehlender zählt als leer**: ein Beleg aus der Zeit vor
// den Schaltern und ein heute ungesetzter sind derselbe Stand.
func andererSchalter(damals, heute map[string]string) string {
	namen := make([]string, 0, len(damals)+len(heute))
	for name := range damals {
		namen = append(namen, name)
	}

	for name := range heute {
		namen = append(namen, name)
	}

	slices.Sort(namen)

	for _, name := range slices.Compact(namen) {
		alt, neu := damals[name], heute[name]

		switch {
		case alt == neu:
			continue
		case neu == "":
			return "gemessen mit " + name + "=" + alt + ", heute ohne"
		case alt == "":
			return "gemessen ohne " + name + ", heute " + name + "=" + neu
		default:
			return "gemessen mit " + name + "=" + alt + ", heute " + name + "=" + neu
		}
	}

	return ""
}

// pfadeAus nimmt `git status --porcelain` den Zustand und lässt die Pfade —
// **fünf beim Namen, der Rest als Zahl**: am ersten echten Lauf standen dort
// achtzehn, und die Zeile war nicht mehr zu lesen.
//
// **Getrennt am ersten Leerzeichen und nicht an der dritten Stelle**: der
// Werkzeugkasten schneidet die Ausgabe zu, und die erste Zeile verliert
// dabei ihren führenden Leerraum — aus » M k.go« wird »M k.go«.
func pfadeAus(zeilen []string) string {
	pfade := make([]string, 0, min(len(zeilen), genannt))

	for _, zeile := range zeilen[:min(len(zeilen), genannt)] {
		pfad := strings.TrimSpace(zeile)
		if _, rest, getrennt := strings.Cut(pfad, " "); getrennt {
			pfad = strings.TrimSpace(rest)
		}

		pfade = append(pfade, pfad)
	}

	liste := strings.Join(pfade, ", ")
	if weitere := len(zeilen) - genannt; weitere > 0 {
		liste += fmt.Sprintf(" und %d weitere", weitere)
	}

	return liste
}

// kurzform gibt die ersten sieben Zeichen eines Commits — wie `git` sie zeigt.
func kurzform(commit string) string {
	if len(commit) > 7 {
		return commit[:7]
	}

	return commit
}
