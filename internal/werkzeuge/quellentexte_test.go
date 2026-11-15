// Der Wächter über die Sätze zu den Bezugsquellen —
// #263.
//
// `internal/bezug/quellen.go` gibt je Quelle einen **Schlüssel**, die
// Sprachkataloge geben den Satz dazu (§9). Zwei Seiten, die niemand
// gegeneinander hält, laufen auseinander — und der Ausfall ist still: die
// Kachel zeigte den Namen und darunter nichts.
//
// **Beide Sprachen, denn genau eine zu vergessen ist der Normalfall.**
//
// # Was er nicht prüft
//
// **Ob der Satz stimmt.** Er sucht den Schlüssel. Ob darunter „Vorlagen für
// andere Spiele" oder „Rezept für Kartoffelsalat" steht, sieht nur, wer liest.
package werkzeuge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
)

func TestJedeQuelleHatIhrenSatzInBeidenSprachen(t *testing.T) {
	t.Parallel()

	for _, sprache := range []string{"de", "en"} {
		pfad := filepath.Join("..", "..", "frontend", "src", "sprache", sprache, "seiten.ts")

		inhalt, err := os.ReadFile(pfad) // #nosec G304 -- der Pfad steht hier, er kommt nicht von außen
		if err != nil {
			t.Fatalf("%s lesen: %v", pfad, err)
		}

		text := string(inhalt)

		for _, quelle := range bezug.Quellen() {
			t.Run(sprache+"/"+quelle.Schluessel, func(t *testing.T) {
				t.Parallel()

				// **Mit Doppelpunkt gesucht.** Ein bloßes `m15` steht in jeder
				// zweiten Zeile dieser Datei; `m15:` ist ein Katalogeintrag.
				if !strings.Contains(text, quelle.Schluessel+":") {
					t.Errorf("%s nennt %q nicht — die Kachel bliebe ohne Satz",
						pfad, quelle.Schluessel)
				}
			})
		}
	}
}
