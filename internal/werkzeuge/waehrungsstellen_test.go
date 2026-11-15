package werkzeuge_test

// **Das Fenster teilt durch dieselben Stellen, die Go speichert** —
// LOC-03.AK-02 (#1689).
//
// # Wozu dieser Wächter da ist
//
// **Weil die Skala eines Betrags an zwei Stellen steht.** Go legt in
// `internal/domain/geld.go` fest, was eine Zahl im Bestand bedeutet; das
// Fenster teilt in `frontend/src/waehrungsstellen.ts` durch dieselbe
// Stellenzahl, bevor es zeigt. Laufen die beiden auseinander, steht ein
// Betrag hundertfach im Fenster — und kein Lauf wird rot.
//
// **Das ist schon einmal geschehen**: das Fenster fragte `Intl`, und `Intl`
// antwortet nach CLDR statt nach ISO 4217 — gemessen am 2026-09-22 bei 17
// von 162 Codes anders, darunter HUF und IQD.
//
// # Was er nicht hält
//
// **Ob ISO 4217 sich geändert hat.** Er vergleicht zwei Listen miteinander,
// nicht eine mit der Norm.

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Für jeden möglichen Code sagen beide dasselbe.**
//
// **Alle 17.576 und nicht die bekannten**: eine Liste der bekannten stünde
// selbst unter keinem Wächter.
func TestDasFensterKenntDieStellenWieGo(t *testing.T) {
	t.Parallel()

	fenster := fensterstellen(t)
	abweichend := []string{}

	for _, erster := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		for _, zweiter := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
			for _, dritter := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
				code := string([]rune{erster, zweiter, dritter})

				gespeichert, _ := domain.Waehrungsstellen(code)

				gezeigt, bekannt := fenster[code]
				if !bekannt {
					gezeigt = 2
				}

				if gespeichert != gezeigt {
					abweichend = append(abweichend, code)
				}
			}
		}
	}

	if len(abweichend) > 0 {
		t.Errorf("Go und frontend/src/waehrungsstellen.ts nennen für diese "+
			"Codes verschiedene Stellen: %s", strings.Join(abweichend, ", "))
	}
}

// fensterstellen liest die beiden Mengen aus der Datei des Fensters.
func fensterstellen(t *testing.T) map[string]int {
	t.Helper()

	roh, fehler := os.ReadFile("../../frontend/src/waehrungsstellen.ts")
	if fehler != nil {
		t.Fatalf("die Datei des Fensters lesen: %v", fehler)
	}

	stellen := map[string]int{}

	for name, zahl := range map[string]int{"nullstellige": 0, "dreistellige": 3} {
		menge := regexp.MustCompile(
			`(?s)const ` + name + ` = new Set\(\[(.*?)\]\)`).
			FindStringSubmatch(string(roh))

		// **Eine umbenannte Menge macht den Wächter nicht leer** — sonst
		// stimmte er jedem Fenster zu, das gar keine Liste mehr hat.
		if menge == nil {
			t.Fatalf("die Menge %s steht nicht in waehrungsstellen.ts", name)
		}

		for _, code := range regexp.MustCompile(`'([A-Z]{3})'`).
			FindAllStringSubmatch(menge[1], -1) {
			stellen[code[1]] = zahl
		}
	}

	return stellen
}
