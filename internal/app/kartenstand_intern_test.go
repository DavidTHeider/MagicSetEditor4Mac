package app

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Ausgang von `kartenstandUeber` trägt einen Fall**
// (#1005).
//
// # Warum ein interner Test
//
// Weil die Quelle eingesetzt wird und `Kartenquelle` nicht exportiert werden
// soll — sie ist eine Eigenschaft dieser Naht, keine Fähigkeit für Aufrufer.
//
// # Warum die Reihenfolge im Code zählt
//
// Erst das Set, dann die Karte. Andersherum wäre der Ausgang „das Set gibt es
// nicht" **unerreichbar**: wer die Karte gelesen und ihr Set geprüft hat, weiß
// danach, dass es das Set gibt. §3 nimmt eine unbetretene Verzweigung nicht
// als Ausnahme an.
func TestJederAusgangDesKartenstandsTraegtEinenFall(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		quelle Kartenquelle
		setID  int64
		gut    bool
	}{
		"alles da": {
			quelle: quellemit{spiel: "magic", setID: 1}, setID: 1, gut: true,
		},
		"die Setköpfe lassen sich nicht lesen": {
			quelle: quellemit{kopffehler: errors.New("der Bestand antwortet nicht")}, setID: 1,
		},
		"das Set gibt es nicht": {
			quelle: quellemit{spiel: "magic", setID: 1}, setID: 4711,
		},
		"die Karte lässt sich nicht lesen": {
			quelle: quellemit{
				spiel: "magic", setID: 1, kartenfehler: errors.New("die Karte fehlt"),
			},
			setID: 1,
		},
		"die Karte gehört zu einem anderen Set": {
			quelle: quellemit{spiel: "magic", setID: 1, karteAusSet: 9}, setID: 1,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			werte, spiel, fehler := kartenstandUeber(probekontext(t), fall.quelle, fall.setID, 2)

			if fall.gut {
				if fehler != nil {
					t.Fatalf("Fehler: %v", fehler)
				}

				if spiel != "magic" || werte["name"] != "Alt" {
					t.Errorf("Spiel %q, Werte %v", spiel, werte)
				}

				return
			}

			if fehler == nil {
				t.Error("der Fall kam durch")
			}
		})
	}
}

// quellemit gibt genau das, was der Fall braucht.
type quellemit struct {
	spiel        string
	setID        int64
	karteAusSet  int64
	kopffehler   error
	kartenfehler error
}

func (q quellemit) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	if q.kopffehler != nil {
		return nil, q.kopffehler
	}

	return []domain.SetHead{{ID: q.setID, Game: q.spiel}}, nil
}

func (q quellemit) KarteLesen(context.Context, int64) (*domain.CardRow, error) {
	if q.kartenfehler != nil {
		return nil, q.kartenfehler
	}

	gehoertZu := q.setID
	if q.karteAusSet != 0 {
		gehoertZu = q.karteAusSet
	}

	return &domain.CardRow{
		ID: 2, SetID: gehoertZu,
		Values: []domain.Value{{Field: "name", Text: "Alt"}},
	}, nil
}
