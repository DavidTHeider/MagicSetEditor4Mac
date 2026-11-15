//go:build fremdpakete

package render

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

// **Die Namensschrift, gesetzt wie von Core Text** — QA-03.AK-02
// (#1656)
// und die Ligaturen
// (#1805).
//
// MSE2 setzt unter macOS über wxWidgets und Core Text, und wxWidgets stellt
// dabei kein Merkmal um (`docs/SCHRIFTUNTERSUCHUNG.md`). Breite und
// Glyphenzahl unten hat Core Text am 2026-09-23 auf diesem Rechner gegeben —
// `CTLineGetTypographicBounds` über eine Zeile bei 16 Punkt, dem Aufruf, mit
// dem wxWidgets die Breite einer Zeichenkette misst
// (`scripts/coretext/breiten.swift`).
//
// **Gewählt sind die Namen, an denen es hing**: die vier mit der größten
// Abweichung, solange die Klassenpaare fehlten, die Beispiele aus der
// Untersuchung — einer mit positivem Kerning (`Kyscu Drake`), einer ganz ohne
// (`Doom Blade`) — und sieben mit `ffi` oder `ffl`, die Core Text als **eine**
// Glyphe setzt. **Gemessen wird über [Textbreite] und [gesetzt]**, den Weg,
// den der Rasterer geht, und nicht über eine Summe daneben.
//
// **Beide Beleren-Schnitte**, denn beide führen die zwei Ligaturen.
func TestDieNamensschriftSetztWieCoreText(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	const groesse = 16

	type fall struct {
		name     string
		coreText float64
		glyphen  int
	}

	for datei, faelle := range map[string][]fall{
		"beleren-bold_P1.01.ttf": {
			{"Multani, Yavimaya’s Avatar", 176.6250, 26},
			{"Your Will Is Not Your Own", 174.7422, 25},
			{"When We Were Young", 143.7031, 18},
			{"Towering-Wave Mystic", 149.5547, 20},
			{"Shivan Dragon", 103.0391, 13},
			{"Force of Will", 86.1719, 13},
			{"Kraken Hatchling", 122.0156, 16},
			{"Kyscu Drake", 91.7344, 11},
			{"Doom Blade", 86.4844, 10},
			{"Armored Griffin", 114.8594, 13},
			{"Efficient Construction", 154.4453, 20},
			{"Trufflesnout", 88.8828, 10},
			{"Pharika, God of Affliction", 179.6016, 24},
			{"Syndicate Trafficker", 140.7031, 18},
			{"Coffin Queen", 93.8672, 10},
			{"Raffine’s Guidance", 131.0391, 16},
		},
		"belerensmallcaps-bold.ttf": {
			{"Shivan Dragon", 108.1016, 13},
			{"Armored Griffin", 125.0000, 13},
			{"Trufflesnout", 97.7891, 10},
			{"Pharika, God of Affliction", 198.5391, 24},
		},
	} {
		pfad := filepath.Join(filepath.Dir(wurzel), "Magic - Fonts", datei)

		roh, err := os.ReadFile(pfad) //nolint:gosec // Pfad aus der Umgebung
		if err != nil {
			t.Fatalf("die Namensschrift des Packs: %v", err)
		}

		schrift, err := SchriftAusDatei(roh)
		if err != nil {
			t.Fatalf("lesen: %v", err)
		}

		for _, f := range faelle {
			breite := Textbreite(f.name, groesse, schrift)
			glyphen := len(gesetzt(f.name, Satz{Groesse: groesse}, schrift))

			if abstand := math.Abs(breite - f.coreText); abstand > 0.001 || glyphen != f.glyphen {
				t.Errorf("%s, %s: %.4f in %d Glyphen, Core Text %.4f in %d — %.4f Punkt daneben",
					datei, f.name, breite, glyphen, f.coreText, f.glyphen, abstand)
			}
		}
	}
}
