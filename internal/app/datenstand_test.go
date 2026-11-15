package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// stolperstand meldet jede Frage als Fehler.
type stolperstand struct{}

func (stolperstand) Datenstand(context.Context) (int64, error) {
	return 0, errors.New("die Platte ist fort")
}

// **Ein Fehler am Datenstand geht als `Fehler` hinaus** (§8), nicht als
// Wails-Zeichenkette.
//
// Der Weg ist der stillste im ganzen Fenster: das Frontend fragt alle zwei
// Sekunden und zeigt nichts an, wenn etwas schiefgeht
// ([0017](../../docs/entscheidungen/0017-die-prozesse-bemerken-einander.md)).
// Gerade deshalb muss er einen Befund tragen — sonst gäbe es keinen Ort, an
// dem jemand ihn je sähe.
func TestEinFehlerAmDatenstandKommtAlsBefund(t *testing.T) {
	anwendung := app.NeueApp(":memory:")
	anwendung.MitStand(stolperstand{})

	antwort := anwendung.Datenstand(t.Context())
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if antwort.Stand != 0 {
		t.Errorf("Stand %d, erwartet 0", antwort.Stand)
	}
}
