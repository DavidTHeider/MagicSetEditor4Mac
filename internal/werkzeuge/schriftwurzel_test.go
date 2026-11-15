// **Die Schriftwurzel wird an genau einer Stelle gelesen** — gefunden beim
// Fremdpaketlauf zur Freigabe 0.3.0
// (#1170).
//
// # Der Befund
//
// Am 2026-08-28 sind **vier** Wege behoben worden, die ihren Pfad aus `HOME`
// bauten, obwohl `TestMain` dem Lauf ein eigenes Zuhause gibt. Der fünfte war
// nicht dabei: `TestDieKanonischenSetsZeichnen` liest `MSE_FREMDPAKETE`, aber
// **nicht** `MSE_SCHRIFTWURZEL` — die Schriften liegen **neben** den Paketen,
// und ohne die Angabe findet er keine.
//
// Der Lauf sprang nicht ab; er wurde **rot**, mit 8.325 Befunden, davon 2.432
// „die Schrift Beleren Bold ist nicht zu finden". Das ist die laute Richtung —
// und trotzdem unbrauchbar: er ist rot aus einem Grund, der nichts mit dem
// Code zu tun hat.
//
// # Was hier gehalten wird
//
// `MSE_SCHRIFTWURZEL` steht in **einer** Testdatei, und das ist die, in der
// jeder dieser Läufe die Anwendung startet. Ein sechster Weg kann sie damit
// nicht vergessen — er bekommt sie, indem er `gestartetMit` ruft.
//
// # Was er nicht hält
//
// Dass ein Lauf `gestartetMit` **benutzt**. Wer sich eine eigene Anwendung
// baut, geht daran vorbei — dann fehlt ihm allerdings auch alles andere, was
// dort verdrahtet wird.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/werkzeuge"
)

func TestDieSchriftwurzelStehtAnEinerStelle(t *testing.T) {
	t.Parallel()

	eintraege, err := os.ReadDir(filepath.Join("..", "app"))
	if err != nil {
		t.Fatalf("internal/app lesen: %v", err)
	}

	var stellen []string

	for _, eintrag := range eintraege {
		if !strings.HasSuffix(eintrag.Name(), "_test.go") {
			continue
		}

		// **Gesucht wird der Zugriff, nicht die Erwähnung.** Der erste Anlauf
		// suchte den Namen — und fand ihn im Kommentar, der erklärt, warum er
		// hier **nicht** mehr steht. Derselbe Fehler wie am Flächenwächter,
		// am selben Tag (#1170, Befund 12).
		//
		// **Und die zweite Hälfte davon fiel erst am 2026-08-31**
		// (#1346):
		// gesucht wurde der ganze Ausdruck, aber in **rohem** Quelltext. Wer
		// den Zugriff entfernte und den Kommentar stehen ließ, hatte einen
		// grünen Lauf. Jetzt liest ein Zerteiler.
		if strings.Contains(
			werkzeuge.GoOhneKommentare(t, filepath.Join("..", "app", eintrag.Name())),
			`Getenv("MSE_SCHRIFTWURZEL")`) {
			stellen = append(stellen, eintrag.Name())
		}
	}

	if len(stellen) != 1 || stellen[0] != "saatverdrahtung_intern_test.go" {
		t.Errorf("MSE_SCHRIFTWURZEL steht in %v — erwartet war allein "+
			"saatverdrahtung_intern_test.go, wo jeder Fremdpaketlauf die "+
			"Anwendung startet. Ein zweiter Ort ist ein Weg, der sie "+
			"vergessen kann", stellen)
	}
}
