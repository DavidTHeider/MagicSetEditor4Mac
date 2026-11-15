// **Der Zugang zum Befund eines Fehlers** — Werkstück B von Phase 4.9
// (#1324).
package msescript_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Ein Zerteilfehler kommt mit Stelle und Schlüssel zurück.**
func TestBefundGibtDenZerteilfehlerZurueck(t *testing.T) {
	_, err := msescript.Zerteile("f(1, 2")
	if err == nil {
		t.Fatal("kein Fehler")
	}

	befund := msescript.Befund(err)

	if befund.Zeile != 1 || befund.Spalte != 2 {
		t.Errorf("Stelle %d:%d, erwartet 1:2", befund.Zeile, befund.Spalte)
	}

	if befund.Schluessel != msescript.SchluesselKlammerOffen {
		t.Errorf("Schlüssel %q", befund.Schluessel)
	}
}

// **Auch aus einer Kette.** Ein Dienst umschließt ihn mit `%w`; ohne
// `errors.As` fände ihn niemand mehr.
func TestBefundFindetIhnInEinerKette(t *testing.T) {
	_, err := msescript.Zerteile("f(1, 2")

	umschlossen := fmt.Errorf("das Feld prüfen: %w", err)

	if befund := msescript.Befund(umschlossen); befund.Spalte != 2 {
		t.Errorf("Stelle %d:%d", befund.Zeile, befund.Spalte)
	}
}

// **Und ein fremder Fehler bekommt die erste Zeile und seinen Satz.**
//
// Das ist der Zweig, um dessentwillen diese Funktion hier steht: bei jedem
// Aufrufer bliebe er unbetreten, weil der Zerteiler nichts anderes gibt.
// **Hier** ist er zu begehen — und die Zusage ist, dass nichts still
// verschwindet.
func TestBefundGibtEinemFremdenFehlerDieErsteZeile(t *testing.T) {
	befund := msescript.Befund(errors.New("etwas ganz anderes"))

	if befund == nil {
		t.Fatal("nil — die Zusage ist »nie nil«")
	}

	if befund.Zeile != 1 || befund.Spalte != 1 {
		t.Errorf("Stelle %d:%d, erwartet 1:1", befund.Zeile, befund.Spalte)
	}

	if befund.Schluessel != msescript.SchluesselDurchgereicht {
		t.Errorf("Schlüssel %q, erwartet %q",
			befund.Schluessel, msescript.SchluesselDurchgereicht)
	}

	if befund.Meldung != "etwas ganz anderes" {
		t.Errorf("Meldung %q", befund.Meldung)
	}
}
