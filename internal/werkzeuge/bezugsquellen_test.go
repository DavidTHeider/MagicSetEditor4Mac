// Der Wächter über die Bezugsquellen —
// #263.
//
// Die Anwendung bietet Quellen zum Anklicken an. `DRITTANBIETER.md` sagt, wem
// deren Inhalte gehören und unter welcher Lizenz sie stehen — bei allen vier
// heißt die Antwort **keine**, und genau deshalb muss sie dastehen: eine
// Quelle, die die Anwendung anbietet und über die niemand nachlesen kann, ist
// die Lage, die §11 verhindern soll.
//
// # Was er nicht prüft
//
// **Ob die Lizenzangabe stimmt.** Er sucht den Namen und das Wort; ob das Wort
// heute noch zutrifft, weiß nur, wer bei GitHub nachgesehen hat. Zuletzt
// geschehen am 2026-08-19: zehn Repositorien des Kontos
// `MagicSetEditorPacks`, keines mit Lizenz.
//
// **Ob der Stand erreichbar ist.** Ein Commit, den es nicht mehr gibt, fällt
// erst beim Bezug auf. Ein Netzzugriff in der Pipeline wäre ein Test, der bei
// schlechtem WLAN rot wird.
package werkzeuge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
)

func TestJedeBezugsquelleStehtInDrittanbieter(t *testing.T) {
	t.Parallel()

	inhalt, err := os.ReadFile(filepath.Join("..", "..", "DRITTANBIETER.md"))
	if err != nil {
		t.Fatalf("DRITTANBIETER.md lesen: %v", err)
	}

	text := string(inhalt)

	for _, quelle := range bezug.Quellen() {
		t.Run(quelle.Name, func(t *testing.T) {
			t.Parallel()

			if !strings.Contains(text, quelle.Name) {
				t.Errorf("DRITTANBIETER.md nennt %q nicht — die Anwendung bietet es an",
					quelle.Name)
			}

			// **Die Adresse dazu.** Der Name allein steht auch in einem
			// Nebensatz; die Adresse ist die Zusage, dass die Zeile *diese*
			// Quelle meint und nicht eine gleichnamige.
			if !strings.Contains(text, quelle.Adresse) {
				t.Errorf("DRITTANBIETER.md nennt die Adresse von %q nicht", quelle.Name)
			}
		})
	}
}

// **Und der Abschnitt sagt, dass keine Lizenz dasteht.**
//
// Der Abnahmekasten aus #263 lautet: „nennt jede Quelle samt ihrer Lizenzlage;
// wo keine steht, steht **das** dort". Ein Abschnitt, der die Quelle nennt und
// über ihre Lizenz schweigt, erfüllt ihn nicht — und schweigen ist hier die
// wahrscheinlichste Art zu irren, weil „keine Lizenz" wie eine Auslassung
// aussieht.
func TestDerAbschnittNenntDieLizenzlage(t *testing.T) {
	t.Parallel()

	inhalt, err := os.ReadFile(filepath.Join("..", "..", "DRITTANBIETER.md"))
	if err != nil {
		t.Fatalf("DRITTANBIETER.md lesen: %v", err)
	}

	abschnitte := strings.Split(string(inhalt), "\n### ")

	for _, quelle := range bezug.Quellen() {
		t.Run(quelle.Name, func(t *testing.T) {
			t.Parallel()

			for _, abschnitt := range abschnitte {
				if !strings.HasPrefix(abschnitt, quelle.Name) {
					continue
				}

				if !strings.Contains(abschnitt, "Lizenz") {
					t.Errorf("der Abschnitt zu %q sagt nichts über die Lizenz", quelle.Name)
				}

				return
			}

			t.Errorf("DRITTANBIETER.md hat keinen Abschnitt `### %s`", quelle.Name)
		})
	}
}
