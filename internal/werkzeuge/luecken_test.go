// **Eine Lücke, die es nicht mehr gibt, ist eine Falschaussage**
// (Audit 0.3.0 (Entwicklungsarchiv), Befund A0-01).
//
// # Der Befund
//
// `docs/MSE2-OBERFLAECHE.md` und `docs/PARITAET.md` beantworten die Frage
// „was fehlt hier gegenüber MSE2?". Am 2026-08-26 waren **acht von 22**
// „fehlt"-Zeilen falsch und **neun von 14** offenen Paritätszeilen nannten
// eine Phase, die längst abgeschlossen ist.
//
// **Und die vorhandenen Wächter hielten nur die Zahlen.**
// `paritaet_test.go` zählt die Marken aus den Tabellen nach; dass eine Zeile
// ihren **Inhalt** verloren hat, sieht er nicht. Das ist die Gestalt aus §10:
// ein Wächter, der die halbe Zusage hält, macht die andere unsichtbar.
//
// **Warum das teuer ist**: aus diesen beiden Dateien wird geplant. Eine
// Planung auf falschen Zeilen priorisiert Arbeit, die getan ist, und übersieht,
// was danebensteht.
//
// # Was er hält, und was nicht
//
// Er hält **zwei mechanische Widersprüche**:
//
//  1. eine offene Paritätszeile ohne Plan — oder mit einem, der auf eine
//     **abgeschlossene** Phase zeigt,
//  2. eine „fehlt"-Zeile, die ein erledigtes Issue nennt — seit 1.0.0 jedes
//     Issue aus dem Entwicklungsarchiv (Vermerk 0205).
//
// **Was er nicht hält**: ob eine Zeile ohne Phasen- und Issueangabe noch
// stimmt. Das liest ein Mensch — und dieser Wächter macht seine Arbeit
// kleiner, nicht überflüssig.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// phasenInDerChronik gibt die Phasen, die als **fertig** geführt sind.
//
// **Aus `CHRONIK.md` und nicht aus einer Liste daneben.** §9: dort steht, was
// erledigt ist — eine zweite Aufzählung liefe mit ihr auseinander.
//
// **Seit 1.0.0 aus der Liste „Bis 1.0.0 — die abgeschlossenen Phasen"**: die
// Einträge selbst liegen im Entwicklungsarchiv (Vermerk 0205), ihre
// Überschriften stehen dort als Aufzählung.
func phasenInDerChronik(t *testing.T) map[string]bool {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join("..", "..", "docs", "CHRONIK.md"))
	if err != nil {
		t.Fatalf("die Chronik lesen: %v", err)
	}

	muster := regexp.MustCompile(`(?m)^(?:## |- )[\d-]+ — Phase ([A-ZΩ\d]+)`)

	fertig := map[string]bool{}
	for _, treffer := range muster.FindAllStringSubmatch(string(roh), -1) {
		fertig[treffer[1]] = true
	}

	if len(fertig) < 20 {
		t.Fatalf("die Chronik führt nur %d Phasen — das Muster trifft nicht", len(fertig))
	}

	return fertig
}

// **Jede offene Paritätszeile trägt einen Plan, und der zeigt nach vorn.**
//
// # Warum ein Feld und keine Prosa
//
// Der erste Anlauf suchte „Phase X" **irgendwo** in der Zeile — und strich
// damit auch die Zeilen an, die eine abgeschlossene Phase zu Recht nennen:
// „das ist gebaut (Phase X)", „Phase Y sollte es entscheiden und hat nicht".
// Ein Wächter, der Richtiges anstreicht, wird stummgeschaltet (§10).
//
// **Also wird der Plan ein Feld.** Eine offene Zeile endet auf
// `**Plan:** …`; alles davor ist Erklärung und darf jede Phase nennen. Der
// Wächter liest **nur** den Plan.
//
// Das hält zwei Dinge statt einem: dass eine offene Zeile überhaupt einen
// Plan hat — **das ist die Zeile, die verrottet** —, und dass er nicht auf
// etwas zeigt, das längst geschehen ist.
//
// # Und seit dem 2026-08-30 auch die Fassungen
//
// Er sah nur „Phase X". **Neun Pläne zeigten auf „0.3.0" oder „nach 0.3.0"**,
// und `v0.3.0` ist am 2026-08-28 freigegeben worden — der Lauf war grün, und
// neun offene Zeilen planten eine Vergangenheit
// (#1291,
// Werkstück C). Welche Fassungen erschienen sind, steht in `CHANGELOG.md`:
// eine Überschrift `## [0.3.0] — …` ist eine, `## [Unveröffentlicht]` nicht.
//
// **„nach 0.3.0" wird mitgezählt**, und das ist Absicht: nach einer
// erschienenen Fassung ist **jetzt**. Ein Plan, der das sagt, sagt nichts.
func TestJedeOffeneParitaetszeileTraegtEinenPlanNachVorn(t *testing.T) {
	t.Parallel()

	fertig := phasenInDerChronik(t)

	roh, err := os.ReadFile(filepath.Join("..", "..", "docs", "PARITAET.md"))
	if err != nil {
		t.Fatalf("die Paritätsmatrix lesen: %v", err)
	}

	phase := regexp.MustCompile(`Phase ([A-ZΩ0-9]+)`)
	fassung := regexp.MustCompile(`\b(\d+\.\d+\.\d+)\b`)
	erschienen := fassungenImProtokoll(t)

	offen := 0

	for nummer, zeile := range strings.Split(string(roh), "\n") {
		if !strings.HasPrefix(zeile, "|") {
			continue
		}

		if !strings.Contains(zeile, "| deferred |") && !strings.Contains(zeile, "| uncertain |") {
			continue
		}

		offen++

		_, plan, gefunden := strings.Cut(zeile, "**Plan:**")
		if !gefunden {
			t.Errorf("PARITAET.md:%d ist offen und nennt keinen Plan — "+
				"eine offene Zeile ohne Plan ist die, die verrottet", nummer+1)

			continue
		}

		for _, treffer := range fassung.FindAllStringSubmatch(plan, -1) {
			if istErschienen(erschienen, treffer[1]) {
				t.Errorf("PARITAET.md:%d plant die Fassung %s — sie ist längst "+
					"erschienen", nummer+1, treffer[1])
			}
		}

		for _, treffer := range phase.FindAllStringSubmatch(plan, -1) {
			if fertig[treffer[1]] {
				t.Errorf("PARITAET.md:%d plant Phase %s — sie steht längst in "+
					"der Chronik", nummer+1, treffer[1])
			}
		}
	}

	// **Und die Gegenprobe, dass überhaupt etwas geprüft wurde.** Ohne sie
	// wäre dieser Wächter grün, wenn das Muster für „deferred" nicht mehr
	// trifft — und niemand sähe es.
	if offen == 0 {
		t.Error("keine einzige offene Zeile gefunden — trifft das Muster noch?")
	}
}

// **Keine „fehlt"-Zeile nennt ein Issue aus dem Entwicklungsarchiv.**
//
// # Warum das Archiv und nicht mehr der CHANGELOG
//
// Bis 1.0.0 las dieser Wächter die erledigten Issues aus dem `CHANGELOG.md`.
// Seit 1.0.0 fasst der eine Fassung zusammen, und die Issues bis dahin liegen
// im privaten Entwicklungsarchiv, **alle geschlossen** bis auf das Dashboard
// der Abhängigkeiten (Vermerk 0205). Ein Verweis darauf steht als Klartext
// `#N`; ein Issue dieses Repositoriums steht als Verweis mit Adresse. Eine
// „fehlt"-Zeile, die ein `#N` aus dem Archiv nennt, plant also mit einem
// geschlossenen Issue — derselbe Widerspruch wie vorher, nur aus einer
// Quelle, die sich nicht mehr ändert.
func TestKeineLueckenzeileNenntEinIssueAusDemArchiv(t *testing.T) {
	t.Parallel()

	roh, err := os.ReadFile(filepath.Join("..", "..", "docs", "MSE2-OBERFLAECHE.md"))
	if err != nil {
		t.Fatalf("die Oberflächenkarte lesen: %v", err)
	}

	zeilen := 0

	for zeilennummer, zeile := range strings.Split(string(roh), "\n") {
		if !strings.HasPrefix(zeile, "|") {
			continue
		}

		// **Nur die Spalte „Hier".**
		//
		// Eine Zeile darf ein Issue **vor** dem Befund nennen, um etwas
		// abzugrenzen: `card_angle` sagt „nicht zu verwechseln mit `angle:`
		// am Feldstil, das gebaut ist (#486)" — und meint damit das
		// **Gegenteil** einer Lücke. Der erste Anlauf las die ganze Zeile und
		// strich sie an; ein Wächter, der Richtiges anstreicht, wird
		// stummgeschaltet (§10).
		_, hier, gefunden := strings.Cut(zeile, "**fehlt**")
		if !gefunden {
			continue
		}

		zeilen++

		for _, nummer := range archivverweise(hier) {
			t.Errorf("MSE2-OBERFLAECHE.md:%d sagt „fehlt\" und nennt #%s — "+
				"ein Issue aus dem Entwicklungsarchiv, und dort ist es geschlossen", zeilennummer+1, nummer)
		}
	}

	// **Ohne diese Zeile wäre der Lauf grün, wenn „fehlt" nicht mehr trifft.**
	if zeilen == 0 {
		t.Fatal("keine einzige „fehlt\"-Zeile gefunden — trifft das Muster noch?")
	}
}

// archivverweise gibt die Nummern der Verweise `#N`, die nicht Text eines
// Verweises mit Adresse sind — die Verweise ins Entwicklungsarchiv.
func archivverweise(text string) []string {
	ohneVerweise := mitAdresse.ReplaceAllString(text, "")

	funde := klartextnummer.FindAllStringSubmatch(ohneVerweise, -1)

	nummern := make([]string, 0, len(funde))
	for _, treffer := range funde {
		nummern = append(nummern, treffer[1])
	}

	return nummern
}

var (
	mitAdresse     = regexp.MustCompile(`\[[^\]]*\]\(https://[^)]*\)`)
	klartextnummer = regexp.MustCompile(`(?:^|[^\w&/])#(\d+)\b`)
)

// **Die Leserproben**: dass der Leser greift, wo er soll, und nicht, wo er
// nicht soll.
func TestArchivverweiseGreifenWirklich(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name string
		text string
		will []string
	}{
		{"Klartext", "plant mit #2255", []string{"2255"}},
		{"zwei", "#12 und (#345)", []string{"12", "345"}},
		// **Zusammengesetzt**, damit ein Werkzeug, das Verweise ins Archiv in
		// Klartext wandelt, die Probe nicht mit umschreibt — so geschehen beim
		// Wechsel auf 1.0.0.
		{"Verweis mit Adresse", "[#12](https://github.com/DavidTHeider/MagicSetEditor4Mac" + "/issues/12)", nil},
		{"Anker im Text", "siehe `datei.md#abschnitt`", nil},
		{"HTML-Entität", "&#8212;", nil},
	}

	for _, fall := range faelle {
		if hat := archivverweise(fall.text); !slices.Equal(hat, fall.will) {
			t.Errorf("%s: %q gibt %v, erwartet %v", fall.name, fall.text, hat, fall.will)
		}
	}
}

// istErschienen sagt, ob eine Fassung erschienen ist — sie selbst oder eine
// spätere hat eine datierte Überschrift.
//
// **Seit 1.0.0 über die Reihenfolge und nicht mehr über die Menge**: die
// Fassungen bis 0.9.0 stehen nicht mehr mit eigener Überschrift im
// Protokoll, erschienen sind sie trotzdem (Vermerk 0205).
func istErschienen(erschienen map[string]bool, fassung string) bool {
	for datiert := range erschienen {
		if vergleicheFassung(fassung, datiert) <= 0 {
			return true
		}
	}

	return false
}

// vergleicheFassung vergleicht zwei Fassungen der Gestalt `a.b.c`.
func vergleicheFassung(a, b string) int {
	teileA, teileB := strings.Split(a, "."), strings.Split(b, ".")
	for i := range min(len(teileA), len(teileB)) {
		// **Ein Teil, der keine Zahl ist, zählt als 0** — die Fassungen kommen aus
		// einem Muster, das nur Ziffern trifft; der Fehler kann nicht eintreten.
		x, fehlerA := strconv.Atoi(teileA[i])
		y, fehlerB := strconv.Atoi(teileB[i])

		if fehlerA != nil || fehlerB != nil {
			return strings.Compare(teileA[i], teileB[i])
		}

		if x != y {
			return x - y
		}
	}

	return len(teileA) - len(teileB)
}

// fassungenImProtokoll nennt die Fassungen, die eine datierte Überschrift im
// Änderungsprotokoll haben.
//
// **`## [Unveröffentlicht]` zählt nicht** — das ist der Ort, an dem die
// nächste entsteht, und ein Plan darf dorthin zeigen.
func fassungenImProtokoll(t *testing.T) map[string]bool {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join("..", "..", "CHANGELOG.md"))
	if err != nil {
		t.Fatalf("das Änderungsprotokoll lesen: %v", err)
	}

	ueberschrift := regexp.MustCompile(`(?m)^## \[(\d+\.\d+\.\d+)\] — \d`)

	erschienen := map[string]bool{}
	for _, treffer := range ueberschrift.FindAllStringSubmatch(string(roh), -1) {
		erschienen[treffer[1]] = true
	}

	// **Ohne diese Zeile wäre der Lauf grün, wenn das Muster nie greift.**
	if len(erschienen) == 0 {
		t.Fatal("keine einzige datierte Fassung im Protokoll — das Muster trifft nicht")
	}

	return erschienen
}
