//go:build fremdpakete

// **Unsere Vorschübe treffen die von wx** — die Naht, an der #598 hängt.
//
// # Warum das hier steht
//
// #598 sagt,
// MSE rastere auch die **Breite** auf Viertelpunkte. Bevor jemand das umsetzt,
// gehört belegt, dass unsere Vorschübe überhaupt dieselben sind wie wx' —
// sonst rastert man eine falsche Zahl.
//
// **Sie sind es**, gemessen am 2026-08-20 gegen wx 3.3.3: neun Zeichenketten
// bei zwei Punktgrößen, und `ceil(unsere Summe)` trifft **18 von 18**.
// `round` verfehlt drei.
//
// **In dieser Schrift ist kein Kerning im Spiel.** Unsere Summe liegt bei jeder
// der achtzehn knapp **unter** wx' ganzer Zahl; gäbe es Kerning, läge sie mal
// darüber.
//
// **Und das ist keine Aussage über andere Schriften** —
// #752:
// `mplantin.ttf` trägt **gar keine** Kerningtabelle, die Messung konnte darüber
// also nichts sagen. Welche Schriften des Packs kernen, steht in
// `kerningreichweite_intern_test.go`.
//
// # Was er nicht sagt
//
// **Nicht, dass die Rasterung an dieser Stelle richtig wäre.** Sie ist am
// 2026-08-20 versucht und **gemessen zurückgenommen** worden: über den
// Textlauf angewandt kostete sie 20,679 → 24,728. Die Regel stimmt, die Ebene
// nicht — der Stand steht an #598.
package render

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

// **Die Breite eines Zuges, wie wx sie misst.**
func TestDieBreiteTrifftWx(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	pfad := filepath.Join(filepath.Dir(wurzel), "Magic - Fonts", "mplantin.ttf")

	roh, err := os.ReadFile(pfad) //nolint:gosec // der Pfad kommt aus der Umgebung des Laufs
	if err != nil {
		t.Skipf("keine Schriftdatei unter %s: %v", pfad, err)
	}

	schrift, err := SchriftAusDatei(roh)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// Am 2026-08-20 mit `wxMemoryDC::GetTextExtent` erhoben.
	faelle := []struct {
		text                                    string
		beiSechsundfuenfzig, beiSechsundneunzig int
	}{
		{"H", 48, 82},
		{"Hg", 76, 130},
		{"Hallo", 136, 233},
		{"Wolkenbruch", 331, 567},
		{"Fuege drei Schaden zu.", 565, 968},
		{"AV", 80, 137},
		{"To", 73, 125},
		{"iiiii", 76, 131},
		{"MMMMM", 283, 485},
	}

	for _, fall := range faelle {
		verhaeltnis := 0.0
		for _, zeichen := range fall.text {
			verhaeltnis += schrift.Breite(zeichen)
		}

		for _, groesse := range []struct {
			punkte int
			wx     int
		}{{56, fall.beiSechsundfuenfzig}, {96, fall.beiSechsundneunzig}} {
			unsere := verhaeltnis * float64(groesse.punkte)

			if gerastert := int(math.Ceil(unsere)); gerastert != groesse.wx {
				t.Errorf("%q bei %d: ceil(%.3f) = %d, wx misst %d",
					fall.text, groesse.punkte, unsere, gerastert, groesse.wx)
			}

			// **Die Gegenprobe auf Kerning**: läge unsere Summe über wx' Zahl,
			// zöge wx Zeichen zusammen und `ceil` wäre die falsche Rechnung.
			if unsere > float64(groesse.wx) {
				t.Errorf("%q bei %d: unsere Summe %.3f liegt über wx' %d — "+
					"das sähe nach Kerning aus", fall.text, groesse.punkte, unsere, groesse.wx)
			}
		}
	}
}
