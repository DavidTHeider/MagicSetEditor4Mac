package app

// **Das Bild einer Karte für die KI wird mit Budget gelesen** — Befund 2 des
// Nachlaufs über Phase 6.1.
//
// Der Wächter über ganzes Lesen führte `sichtbildLesen` als „eine Datei, die
// ein Mensch im Dialog gewählt hat". Den Pfad nehmen aber auch MCP und der
// Dienst entgegen — ein Aufrufer dort nennt, was er will, und eine Datei über
// jeder Grenze belegte den Speicher, bevor irgendetwas geprüft war.

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Nicht parallel**, und das ist Absicht: gemessen wird die Zuteilung des
// ganzen Prozesses, und parallele Tests warten, bis die folgenden fertig sind
// (`internal/werkzeuge/zuteilungsmessung_test.go`).
func TestEinUebergrossesSichtbildWirdNichtGanzGelesen(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "riesig.png")
	if err := os.WriteFile(pfad, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	// **Eine Datei mit Loch** — größer als die Grenze, ohne die Platte zu füllen.
	if err := os.Truncate(pfad, msepaket.GrenzePaketdatei+1); err != nil {
		t.Fatal(err)
	}

	var vorher, nachher runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&vorher)

	_, _, fehler := sichtbildLesen(pfad)

	runtime.ReadMemStats(&nachher)

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet eine Abweisung", fehler)
	}

	if belegt := nachher.TotalAlloc - vorher.TotalAlloc; belegt > 64<<20 {
		t.Errorf("%d MiB zugeteilt — die Datei wurde ganz gelesen", belegt>>20)
	}
}
