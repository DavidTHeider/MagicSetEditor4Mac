// **Was freigegeben ist, ist nicht das, was als Nächstes kommt** —
// #1287.
//
// # Der Befund
//
// `docs/ROADMAP.md` sagte am 2026-08-29 in seiner Stand-Tabelle:
//
//	Freigegeben    0.2.0 am 2026-08-26
//	Als Nächstes   0.3.0 — neun Phasen
//
// 0.3.0 war zu diesem Zeitpunkt seit dem 2026-08-28 freigegeben: der Tag
// `v0.3.0` signiert, das Release veröffentlicht, `CHANGELOG.md` mit einer
// datierten Überschrift `## [0.3.0] — 2026-08-28`. Die Datei, die sagt, wo das
// Projekt steht, sagte es einen Stand zu spät.
//
// **Kein Wächter sah es**, weil es keinen gab: `changelog_test.go` hält das
// Protokoll gegen sich selbst, `roadmapoffen_test.go` hält die Phasentabelle
// gegen den Fließtext — die **Stand-Tabelle** hielt niemand.
//
// # Was dieser Wächter hält
//
//  1. Was die Roadmap als freigegeben nennt, ist die **oberste** datierte
//     Überschrift des `CHANGELOG.md`, und die Daten stimmen überein.
//  2. Was sie als Nächstes nennt, hat **keine** — sonst ist es schon
//     erschienen und die Zeile veraltet.
//  3. `docs/RELEASE.md` nennt dieselben zwei Fassungen.
//  4. Die Zahl der Phasen, die die Stand-Zeile nennt, ist die Zahl der Zeilen
//     in der Phasentabelle.
//  5. Was die Stand-Zeile an Handprüfungen offen lässt, steht als Zeile in der
//     Tabelle darunter.
//  6. Das README nennt dieselbe freigegebene Fassung.
//
// **„Die oberste" seit dem 2026-08-29.** Der erste Anlauf fragte nur, ob die
// genannte Fassung *irgendeine* datierte Überschrift hat — damit ging „0.2.0
// freigegeben, 0.4.0 als Nächstes" durch, obwohl 0.3.0 dazwischen liegt. Der
// Zweitblick über Phase 4.0 hat es nachgestellt (Befund 5): ein Wächter, der
// eine Menge prüft, wo eine Reihenfolge gemeint ist, hält die halbe Zusage.
//
// **Und Punkt 3 aus demselben Befund**: `docs/RELEASE.md` sagte „Die nächste
// Freigabe ist 0.3.0", während die Roadmap sie als freigegeben führte. Zwei
// Dateien, zwei Antworten — und die eine, die ein Mensch beim Freigeben liest,
// war die falsche.
//
// Das ist die Zusage, die kaputt war, und sie hängt an keiner Zahl, die
// jemand fortschreiben muss: die datierte Überschrift entsteht bei der
// Freigabe (§9), und der Widerspruch entsteht im selben Moment.
//
// # Was er nicht hält
//
// **Punkt 4 kam vom Zweitblick über Phase 4.0** (Befund 1): die Stand-Zeile
// sagte „acht Phasen", die Tabelle darunter führte acht — und „neun" hätte
// niemanden gestört. Eine Phase, die dazukommt oder wegfällt, ändert beide
// Stellen oder keine.
//
// **Punkt 5 und 6 kamen vom Nachlauf darüber** (Befunde 1 und 5) — dieselbe
// Klasse, zwei Stellen weiter: „sechs der sieben Handprüfungen" ließ sich zu
// „fünf der neun" machen, ohne dass etwas rot wurde, und „Stand: 0.3.0 steht
// bereit" im README war die **dritte** Angabe zur laufenden Freigabe, die mit
// den beiden anderen auseinanderlaufen konnte.
//
// **Die Handprüfungen werden über ihre Differenz gehalten**, nicht über den
// Freigabebeleg: der liegt unter `build/freigabe/` und ist nicht verzeichnet
// (er entsteht beim Bauen). Was in diesem Repositorium steht, ist die Tabelle
// der offenen Punkte — und sieben minus sechs ist genau ihre Zeilenzahl.
//
// # Was er nicht hält
//
//   - **Ob die Freigabe wirklich stattgefunden hat.** Er liest zwei Dateien
//     dieses Repositoriums, keinen Tag und kein Release. Ein Protokolleintrag,
//     der zu früh datiert wird, überzeugt ihn.
//   - **Ob `wails.json` dieselbe Fassung trägt.** Das hält
//     `scripts/releaseversion.sh` im Freigabelauf und `fassung_test.go` für
//     den MCP-Server. Sie hier noch einmal zu prüfen hieße, eine Fassung
//     mitten im Zyklus einzufrieren — die Nummer wird beim Vorbereiten
//     angehoben, und die Roadmap nennt bis dahin die vorige.
package werkzeuge_test

import (
	"regexp"
	"strings"
	"testing"
)

// freigegebenInDerRoadmap findet „| **Freigegeben** | **0.3.0** am 2026-08-28".
var freigegebenInDerRoadmap = regexp.MustCompile(
	`\|\s*\*\*Freigegeben\*\*\s*\|\s*\*\*([0-9.]+)\*\* am (\d{4}-\d\d-\d\d)`)

// alsNaechstesInDerRoadmap findet „| **Als Nächstes** | **0.4.0** — …".
var alsNaechstesInDerRoadmap = regexp.MustCompile(
	`\|\s*\*\*Als Nächstes\*\*\s*\|\s*\*\*([0-9.]+)\*\*`)

// datierteFassung baut das Muster einer veröffentlichten Überschrift.
func datierteFassung(fassung string) *regexp.Regexp {
	return regexp.MustCompile(`(?m)^## \[` + regexp.QuoteMeta(fassung) + `\] — (\d{4}-\d\d-\d\d)`)
}

func TestDieRoadmapNenntDenStandDerFreigabe(t *testing.T) {
	roadmap := lies(t, "docs/ROADMAP.md")
	protokoll := lies(t, "CHANGELOG.md")

	freigegeben := freigegebenInDerRoadmap.FindStringSubmatch(roadmap)
	if freigegeben == nil {
		t.Fatal("die Stand-Tabelle nennt keine freigegebene Fassung mit Datum — " +
			"der Leser greift daneben")
	}

	naechstes := alsNaechstesInDerRoadmap.FindStringSubmatch(roadmap)
	if naechstes == nil {
		t.Fatal("die Stand-Tabelle nennt keine nächste Fassung — der Leser greift daneben")
	}

	if freigegeben[1] == naechstes[1] {
		t.Fatalf("die Roadmap nennt %s zugleich als freigegeben und als Nächstes",
			freigegeben[1])
	}

	// **Die freigegebene ist die oberste datierte Überschrift.**
	oben := obersteDatierteFassung(t, protokoll)
	switch {
	case oben[0] != freigegeben[1]:
		t.Errorf("die Roadmap nennt %s als freigegeben — die oberste datierte "+
			"Überschrift in CHANGELOG.md ist %s (AGENTS.md §9)",
			freigegeben[1], oben[0])
	case oben[1] != freigegeben[2]:
		t.Errorf("die Roadmap datiert die Freigabe %s auf %s, CHANGELOG.md auf %s",
			freigegeben[1], freigegeben[2], oben[1])
	}

	// **Und die nächste trägt keine.** Sonst ist sie erschienen, und die Zeile
	// hat es als Letzte erfahren — genau der Fall vom 2026-08-29.
	if datierteFassung(naechstes[1]).MatchString(protokoll) {
		t.Errorf("die Roadmap nennt %s als Nächstes — CHANGELOG.md führt sie "+
			"bereits mit Datum. Sie ist erschienen (AGENTS.md §9)", naechstes[1])
	}

	pruefeFreigabeanleitung(t, freigegeben[1], naechstes[1])
	pruefePhasenzahl(t, roadmap)
	pruefeHandpruefungen(t, roadmap)
	pruefeReadmestand(t, freigegeben[1])
}

// handpruefungenInDerStandzeile findet „**sechs** der sieben Handprüfungen".
var handpruefungenInDerStandzeile = regexp.MustCompile(
	`\*\*(\p{L}+)\*\* der (\p{L}+) Handprüfungen`)

// offeneZeile findet eine Zeile der Tabelle unter „Offen ist noch".
var offeneZeile = regexp.MustCompile(`(?m)^\| \*\*Handprüfung `)

// pruefeHandpruefungen hält die Differenz gegen die Tabelle der offenen Punkte.
func pruefeHandpruefungen(t *testing.T, roadmap string) {
	t.Helper()

	fund := handpruefungenInDerStandzeile.FindStringSubmatch(roadmap)
	if fund == nil {
		t.Fatal("die Stand-Zeile nennt keine Handprüfungen — der Leser greift daneben")
	}

	erledigt, eins := standzahlwoerter[strings.ToLower(fund[1])]
	insgesamt, zwei := standzahlwoerter[strings.ToLower(fund[2])]

	if !eins || !zwei {
		t.Fatalf("%q von %q sind keine Zahlwörter bis zwölf", fund[1], fund[2])
	}

	offen := len(offeneZeile.FindAllString(roadmap, -1))
	if insgesamt-erledigt != offen {
		t.Errorf("die Stand-Zeile sagt %d von %d Handprüfungen — offen wären %d, "+
			"die Tabelle darunter führt %d",
			erledigt, insgesamt, insgesamt-erledigt, offen)
	}
}

// standImReadme findet „**Stand: 0.3.0 steht bereit**".
var standImReadme = regexp.MustCompile(`\*\*Stand: ([0-9.]+) steht bereit\*\*`)

// pruefeReadmestand hält die dritte Angabe zur laufenden Freigabe mit.
//
// **Die dritte, nicht die zweite.** Roadmap und Freigabeanleitung standen
// schon unter Aufsicht; das README stand daneben und ist die Datei, die ein
// Fremder zuerst liest (Nachlauf über Phase 4.0, Befund 5).
func pruefeReadmestand(t *testing.T, freigegeben string) {
	t.Helper()

	fund := standImReadme.FindStringSubmatch(lies(t, "README.md"))
	if fund == nil {
		t.Fatal("das README nennt keinen Stand — der Leser greift daneben")
	}

	if fund[1] != freigegeben {
		t.Errorf("das README nennt %s als Stand, die Roadmap %s als freigegeben",
			fund[1], freigegeben)
	}
}

// phasenzahlInDerStandzeile findet „**0.4.0** — acht Phasen".
// **Einzahl erlaubt**: seit dem 2026-09-06 ist nur noch eine Phase offen, und
// „eine Phasen" wäre kein Deutsch. Der Wächter greift nach dem Zahlwort, nicht
// nach der Endung.
//
// **Und seit dem 2026-09-07 auch „keine Phase mehr offen"** — der Zustand, in
// dem nur noch Freigabearbeit aussteht.
var phasenzahlInDerStandzeile = regexp.MustCompile(
	`\*\*Als Nächstes\*\*[^\n]*?\*\*[0-9.]+\*\* — (?:\*\*)?(\p{L}+)(?:\*\*)? Phasen?`)

// phasenzeile findet eine Zeile der Phasentabelle: `| **4.0** | … |`.
//
// **Und seit 0.8.0 `| **8.B** | … |`** — die Phasen dieser Fassung tragen
// Buchstaben (#2204). Ohne sie fand der Wächter null Zeilen, und die Tabelle
// blieb leer, während neun Phasen offen waren.
var phasenzeile = regexp.MustCompile(`(?m)^\|\s*\*\*(\d+(?:\.(?:\d+|[A-Z]))?[a-z]?)\*\*\s*\|`)

// pruefePhasenzahl hält die Zahl im Satz gegen die Zeilen der Tabelle.
func pruefePhasenzahl(t *testing.T, roadmap string) {
	t.Helper()

	fund := phasenzahlInDerStandzeile.FindStringSubmatch(roadmap)
	if fund == nil {
		t.Fatal("die Stand-Zeile nennt keine Phasenzahl — der Leser greift daneben")
	}

	genannt, ist := standzahlwoerter[strings.ToLower(fund[1])]
	if !ist {
		t.Fatalf("%q ist kein Zahlwort bis zwölf — die Stand-Zeile ist "+
			"umgeschrieben worden, und dieser Wächter greift daneben", fund[1])
	}

	if gezaehlt := len(phasenzeile.FindAllString(roadmap, -1)); genannt != gezaehlt {
		t.Errorf("die Stand-Zeile nennt %q (%d) Phasen — die Phasentabelle führt %d",
			fund[1], genannt, gezaehlt)
	}
}

// standzahlwoerter ordnet ausgeschriebene Zahlen ihrem Wert zu.
//
// **Eine eigene Liste**, weil die daneben in `paritaet_test.go` im internen
// Paket steht und diese Datei die äußere Sicht prüft. Zwei kurze Listen sind
// hier billiger als eine Kante zwischen den Paketen.
var standzahlwoerter = map[string]int{
	// **`keine` ist seit dem 2026-09-07 dabei** — mit Phase 4.13 ist die
	// Tabelle leer, und das gab es in diesem Projekt noch nie. Ohne den
	// Eintrag wäre die einzige Möglichkeit gewesen, eine erledigte Phase in
	// der Tabelle stehen zu lassen; die Roadmap führt aber nur Offenes (§9).
	"keine": 0,

	"eine": 1, "zwei": 2, "drei": 3, "vier": 4, "fünf": 5, "sechs": 6,
	"sieben": 7, "acht": 8, "neun": 9, "zehn": 10, "elf": 11, "zwölf": 12,
}

// pruefeFreigabeanleitung hält `docs/RELEASE.md` gegen dieselben zwei Fassungen.
//
// **Weil sie die Datei ist, die ein Mensch beim Freigeben liest.** Eine
// Roadmap, die recht hat, nützt nichts, wenn die Anleitung daneben etwas
// anderes sagt.
func pruefeFreigabeanleitung(t *testing.T, freigegeben, naechstes string) {
	t.Helper()

	anleitung := lies(t, "docs/RELEASE.md")

	fund := naechsteFreigabeInDerAnleitung.FindStringSubmatch(anleitung)
	if fund == nil {
		t.Fatal("docs/RELEASE.md nennt keine nächste Freigabe — der Leser greift daneben")
	}

	if fund[1] != naechstes {
		t.Errorf("docs/RELEASE.md nennt %s als nächste Freigabe, die Roadmap %s",
			fund[1], naechstes)
	}

	if !freigegebenInDerAnleitung(freigegeben).MatchString(anleitung) {
		t.Errorf("docs/RELEASE.md führt %s nicht als freigegeben — die Roadmap tut es",
			freigegeben)
	}
}

// naechsteFreigabeInDerAnleitung findet „Die nächste Freigabe ist **0.4.0**".
var naechsteFreigabeInDerAnleitung = regexp.MustCompile(
	`Die nächste Freigabe ist \*\*([0-9.]+)\*\*`)

// freigegebenInDerAnleitung findet „**0.3.0 am 2026-08-28**".
func freigegebenInDerAnleitung(fassung string) *regexp.Regexp {
	return regexp.MustCompile(`\*\*` + regexp.QuoteMeta(fassung) +
		` (?:ist )?am \d{4}-\d\d-\d\d`)
}

// obersteDatierteFassung gibt Fassung und Datum der obersten Überschrift.
//
// **Die oberste, nicht irgendeine.** Das Protokoll ist absteigend geordnet,
// und `## [Unveröffentlicht]` trägt kein Datum — die erste **datierte**
// Überschrift ist damit die zuletzt erschienene Fassung.
func obersteDatierteFassung(t *testing.T, protokoll string) [2]string {
	t.Helper()

	fund := jedeDatierteFassung.FindStringSubmatch(protokoll)
	if fund == nil {
		t.Fatal("CHANGELOG.md hat keine datierte Überschrift — der Leser greift daneben")
	}

	return [2]string{fund[1], fund[2]}
}

// jedeDatierteFassung findet „## [0.3.0] — 2026-08-28".
var jedeDatierteFassung = regexp.MustCompile(
	`(?m)^## \[([0-9.]+)\] — (\d{4}-\d\d-\d\d)`)
