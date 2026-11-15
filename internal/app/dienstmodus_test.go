// **Die Fläche im Kontext und der Riegel an der Wahl** — Werkstück O1 von
// Phase 5.2 (#1425).
//
// **Hier und nicht in `cmd`**, weil die drei Wurzeln ohne Fenster in ihren
// eigenen Paketen laufen und ein Ruf von dort in der Abdeckung dieses Pakets
// nicht zählt (§3).

package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Was die Wurzel einträgt, liest der Dienst.**
func TestDieFlaecheReistImKontext(t *testing.T) {
	t.Parallel()

	ctx := app.MitDienstmodus(context.Background(), domain.ModusKommandozeile)

	if modus := service.DienstmodusAus(ctx); modus != domain.ModusKommandozeile {
		t.Errorf("Modus %q, erwartet %q", modus, domain.ModusKommandozeile)
	}
}

// **Eine Wahl ohne Riegel lässt jeden nach draußen** — das ist der Stand
// einer Wahl, die nur die Einstellungen kennt, etwa im Test einer Naht.
func TestEineWahlOhneRiegelFragtNiemanden(t *testing.T) {
	t.Parallel()

	wahl := app.NeueEinstellungswahl(einstellungsleser{})

	if fehler := wahl.NachAussen(context.Background()); fehler != nil {
		t.Errorf("ohne Riegel: %v", fehler)
	}
}

// einstellungsleser ist ein Leser ohne Einstellungen.
type einstellungsleser struct{}

func (einstellungsleser) Alle(context.Context) (map[string]string, error) {
	return map[string]string{}, nil
}
