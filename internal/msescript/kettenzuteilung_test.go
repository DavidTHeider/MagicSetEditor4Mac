package msescript_test

import (
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Ein Glied einer Kette kopiert keine Argumente** — E03
// (#2240).
//
// `text_filter` ist eine Kette aus `+` und läuft für jedes Feld jeder Karte.
// Bis E03 kopierte jedes Glied die benannten Argumente seines Rufs zweimal
// und ließ die Liste beim Anhängen von `input` nachwachsen — über eine
// Ausfuhr der 70 Archenemy-Karten rund 1,1 GB von 1,6 GB
// (`docs/LEISTUNGSSCHNITTE.md` §7.1).
//
// **Was ein Glied braucht, sind zwei Zuteilungen**: das Stellungsargument
// und die benannten Argumente samt `input` für die rechte Seite. Gemessen
// wird der Unterschied zwischen acht Gliedern und einem — die Zuteilungen des
// Rufs selbst fallen dabei heraus, und mit ihnen, was der Race-Detektor
// hinzufügt.
func TestEinGliedEinerKetteKopiertKeineArgumente(t *testing.T) {
	zuteilungen := func(t *testing.T, glieder int, ruf string) float64 {
		t.Helper()

		umgebung := msescript.Bibliothek()

		kette, err := msescript.AuswertenQuelle(t.Context(), strings.Repeat("to_lower + ", glieder-1)+"to_lower", umgebung)
		if err != nil {
			t.Fatal(err)
		}

		umgebung.Setze("f", kette)

		knoten, err := msescript.Zerteile(ruf)
		if err != nil {
			t.Fatal(err)
		}

		return testing.AllocsPerRun(100, func() {
			if wert, err := msescript.Auswerten(t.Context(), knoten, umgebung); err != nil || wert.Text != "ab" {
				t.Fatalf("%v, %v", wert, err)
			}
		})
	}

	for _, ruf := range []string{`f("ab")`, `f(input: "ab")`, `f("ab", extra: 1)`} {
		t.Run(ruf, func(t *testing.T) {
			eins, acht := zuteilungen(t, 1, ruf), zuteilungen(t, 8, ruf)

			if jeGlied := (acht - eins) / 7; jeGlied > 2 {
				t.Errorf("ein Glied teilt %.1f-mal zu, erwartet höchstens zweimal (ein Glied %.0f, acht %.0f)", jeGlied, eins, acht)
			}
		})
	}
}

// **Und die Liste eines Glieds ist so lang, wie sie sein muss.**
//
// Ein `append` an eine gekappte Liste verdoppelt ihre Kapazität: aus drei
// benannten Argumenten und `input` wurden sechs Plätze zu je 176 Byte. Erlaubt
// ist, was die Liste und das Stellungsargument brauchen, und ein Achtel für
// die Größenklassen der Zuteilung.
func TestDieListeEinesGliedsWaechstNichtNach(t *testing.T) {
	bytesJeRuf := func(t *testing.T, glieder int) float64 {
		t.Helper()

		umgebung := msescript.Bibliothek()

		kette, err := msescript.AuswertenQuelle(t.Context(), strings.Repeat("to_lower + ", glieder-1)+"to_lower", umgebung)
		if err != nil {
			t.Fatal(err)
		}

		umgebung.Setze("f", kette)

		knoten, err := msescript.Zerteile(`f(input: "ab", a: 1, b: 2)`)
		if err != nil {
			t.Fatal(err)
		}

		const rufe = 1000

		var vorher, nachher runtime.MemStats

		runtime.ReadMemStats(&vorher)

		for range rufe {
			if _, err := msescript.Auswerten(t.Context(), knoten, umgebung); err != nil {
				t.Fatal(err)
			}
		}

		runtime.ReadMemStats(&nachher)

		return float64(nachher.TotalAlloc-vorher.TotalAlloc) / rufe
	}

	eintrag, wert := float64(unsafe.Sizeof(msescript.Eintrag{})), float64(unsafe.Sizeof(msescript.Wert{}))
	erlaubt := (4*eintrag + wert) * 9 / 8

	if jeGlied := (bytesJeRuf(t, 8) - bytesJeRuf(t, 1)) / 7; jeGlied > erlaubt {
		t.Errorf("ein Glied teilt %.0f Byte zu, erlaubt %.0f", jeGlied, erlaubt)
	}
}
