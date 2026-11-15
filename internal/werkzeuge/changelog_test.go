// **Kein Eintrag steht zweimal im Änderungsprotokoll** —
// gefunden vom Zweitblick über Phase V am 2026-08-24.
//
// # Der Fall
//
// Sieben Einträge standen **dreimal** darin: einmal richtig unter
// `## [Unveröffentlicht]`, zweimal fälschlich in der bereits veröffentlichten
// Fassung `0.1.0`. Sie behauptete damit Fähigkeiten, die es am 2026-08-22
// nicht gab.
//
// # Die Klasse
//
// Ein Einfügen an einer Zwischenüberschrift, die **mehrfach** vorkommt.
// `### Geändert` steht in jeder Fassung; wer ohne Rücksicht auf den Abschnitt
// einsetzt, trifft alle. Der Fehler ist nicht das Vergessen, sondern die
// Annahme, eine Überschrift sei eindeutig.
//
// # Was dieser Wächter nicht prüft
//
// **Ob der Eintrag an der richtigen Stelle steht.** Er findet nur, dass
// derselbe Satz mehrfach dasteht. Ein Eintrag, der **einmal** und im falschen
// Abschnitt steht, kommt durch — dagegen hilft der zweite Lauf darunter, und
// auch der nur für die veröffentlichten Fassungen.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// changelog liest das Änderungsprotokoll.
func changelog(t *testing.T) []string {
	t.Helper()

	roh, fehler := os.ReadFile(filepath.Join("..", "..", "CHANGELOG.md"))
	if fehler != nil {
		t.Fatalf("CHANGELOG.md lesen: %v", fehler)
	}

	return strings.Split(string(roh), "\n")
}

// istEintrag sagt, ob eine Zeile einen Punkt eröffnet.
//
// **Nur die erste Zeile eines Punktes.** Fortsetzungszeilen sind eingerückt
// und können sich wiederholen, ohne dass etwas falsch wäre — „Sie gehen über
// eine eigene Drossel" etwa.
func istEintrag(zeile string) bool {
	return strings.HasPrefix(zeile, "- **") && len(zeile) > 20
}

func TestKeinEintragStehtZweimalImProtokoll(t *testing.T) {
	t.Parallel()

	gesehen := map[string]int{}

	for _, zeile := range changelog(t) {
		if istEintrag(zeile) {
			gesehen[strings.TrimSpace(zeile)]++
		}
	}

	for eintrag, wieOft := range gesehen {
		if wieOft > 1 {
			t.Errorf("dieser Eintrag steht %dmal im Protokoll:\n  %s", wieOft, eintrag)
		}
	}
}

// **Eine veröffentlichte Fassung wächst nicht mehr.**
//
// # Warum das über die Zeilenzahl geht und nicht über den Inhalt
//
// Weil eine Fassung inhaltlich beliebig aussehen darf — was sie **nicht** darf,
// ist sich nachträglich ändern. Der Wächter hält deshalb fest, wie lang jeder
// veröffentlichte Abschnitt ist; wer einen Eintrag hineinschreibt, macht ihn
// länger, und der Lauf wird rot.
//
// **Was er nicht prüft**: eine Änderung, die genauso viele Zeilen entfernt wie
// sie einfügt. Dagegen hilft nur, den Abschnitt zu lesen.
func TestEineVeroeffentlichteFassungWaechstNichtMehr(t *testing.T) {
	t.Parallel()

	// **Gezählt, nicht geschätzt** — und beim Ändern einer veröffentlichten
	// Fassung ist die Frage nicht „welche Zahl steht hier jetzt", sondern
	// „warum ändert sie sich".
	// **Seit 1.0.0 hält er die Fassung 1.0.0** — bis dahin `0.1.0` mit 1.622
	// Zeilen; die Abschnitte der Entwicklungsfassungen liegen seither im
	// Entwicklungsarchiv (Vermerk 0205), und der Abschnitt 1.0.0 reicht bis
	// zum Ende der Datei, die Tafel der Entwicklungsfassungen eingeschlossen.
	erwartet := map[string]int{"## [1.0.0] — 2026-11-15": 78}

	gemessen := map[string]int{}

	var offen string

	for _, zeile := range changelog(t) {
		if strings.HasPrefix(zeile, "## [") {
			offen = strings.TrimSpace(zeile)

			continue
		}

		if offen != "" {
			gemessen[offen]++
		}
	}

	for abschnitt, zahl := range erwartet {
		if gemessen[abschnitt] != zahl {
			t.Errorf("%s hat %d Zeilen, erwartet %d — eine veröffentlichte "+
				"Fassung ändert sich nicht mehr", abschnitt, gemessen[abschnitt], zahl)
		}
	}
}
