package werkzeuge_test

// **Ein Bild wird an genau einer Stelle ausgepackt** — FMT-06.AK-01
// (#1695).
//
// # Warum es diesen Wächter gibt
//
// **Weil die Maßprüfung dort entsteht, wo ausgepackt wird.** Ein PNG von ein
// paar hundert Bytes darf 40.000 × 40.000 Bildpunkte behaupten, und
// `image.Decode` teilt dafür 6,4 GB zu, bevor es merkt, dass die Daten fehlen.
// Am 2026-09-23 dekodierten **sieben** Stellen ohne Vorprüfung; geprüft hat
// eine einzige, und zwar mit Grenzen für KI-Bilder. `msepaket.DekodiereBild`
// vermisst erst und packt dann aus — wer daneben `image.Decode` ruft, bekommt
// das nicht.
//
// # Was er hält
//
// **Kein Quelltext außerhalb von `internal/msepaket/bilddekodieren.go` ruft
// `image.Decode`, `png.Decode`, `jpeg.Decode` oder `gif.Decode`.**
//
// # Was er nicht hält
//
// **Ob die Grenze die richtige ist** — sie ist gemessen und steht mit ihrer
// Messung an `msepaket.GrenzeBildpunkte`. Und keine anderen Dekodierer: ein
// neues Format, das jemand registriert, liefe durch `image.Decode` und damit
// durch die Prüfung; einer, den jemand unmittelbar ruft, stünde unter einem
// Namen, den dieser Ausdruck nicht kennt.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// bilddekodierstelle ist die eine Datei, die auspacken darf.
const bilddekodierstelle = "internal/msepaket/bilddekodieren.go"

// dekodiermuster findet jeden unmittelbaren Ruf eines Bilddekodierers.
var dekodiermuster = regexp.MustCompile(`\b(image|png|jpeg|gif)\.Decode\(`)

func TestEinBildWirdNurAnEinerStelleAusgepackt(t *testing.T) {
	t.Parallel()

	gefunden := make([]string, 0, 8)

	for _, wurzel := range []string{"internal", "cmd"} {
		ort := filepath.Join("..", "..", wurzel)

		fehler := filepath.Walk(ort, func(pfad string, eintrag os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if eintrag.IsDir() || !strings.HasSuffix(pfad, ".go") || strings.HasSuffix(pfad, "_test.go") {
				return nil
			}

			roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // Quelltext im Baum
			if lesefehler != nil {
				return lesefehler
			}

			if dekodiermuster.Match(roh) {
				rel, _ := filepath.Rel(filepath.Join("..", ".."), pfad) //nolint:errcheck // unter der Wurzel
				gefunden = append(gefunden, filepath.ToSlash(rel))
			}

			return nil
		})
		if fehler != nil {
			t.Fatal(fehler)
		}
	}

	stelleGesehen := false

	for _, datei := range gefunden {
		if datei == bilddekodierstelle {
			stelleGesehen = true

			continue
		}

		t.Errorf("%s packt ein Bild selbst aus — ohne die Maßprüfung von "+
			"`msepaket.DekodiereBild`", datei)
	}

	// **Die Reichweite der Messung** (§10): ein Ausdruck, der nichts findet,
	// wäre grün. Die eine Stelle, die es geben muss, ist die erlaubte.
	if !stelleGesehen {
		t.Fatalf("der Ausdruck findet nicht einmal %s — dann sagt dieser Lauf nichts", bilddekodierstelle)
	}
}
