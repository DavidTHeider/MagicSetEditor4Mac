package app

// **Die Naht, die den Horcher öffnet** — Werkstück 4 von Phase 4.12
// (#1351).
//
// **Paketintern**, weil die Naht es ist: ein Horchen, das scheitert, kommt auf
// einem gesunden Rechner nicht zustande — und einen Setzer dafür zu
// exportieren hieße, Fläche für einen Testfall zu bauen.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anmeldung"
)

// **Ein Horcher, der nicht aufgeht, beendet den Fluss.**
//
// **Und der Browser geht nicht auf**: ohne Rückadresse gibt es nichts, wohin
// jemand zurückkäme — ein Browserfenster wäre dann eine Sackgasse.
func TestEinHorcherDerNichtAufgehtBeendetDenFluss(t *testing.T) {
	geoeffnet := false

	naht := NeueAnmeldenaht(nil,
		func(context.Context, string) (Anbieterzugang, error) {
			return Anbieterzugang{Kennung: "k", Geheimnis: "g"}, nil
		},
		func(context.Context, string) error {
			geoeffnet = true

			return nil
		},
		nil)

	naht.horche = func() (*anmeldung.Horcher, error) {
		return nil, errors.New("kein Port frei")
	}

	if _, fehler := naht.Anmelden(t.Context(), "github"); fehler == nil {
		t.Fatal("der gescheiterte Horcher kam durch")
	}

	if geoeffnet {
		t.Error("der Browser ging trotzdem auf")
	}
}
