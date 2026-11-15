package service_test

// **Die übrigen Ausgänge der Zugdeutung** — P3
// (#2371),
// angestrichen von der Vorprüfung: die Läufe in `zugdeutung_test.go` gingen
// nur Hand → Schlachtfeld und den Ziel-Sitz. Hier stehen die Gesten, die der
// Kern **nicht** als Land oder Zauber deutet, und die, die er abweist.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieUebrigenGestenFolgenDemProfil(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		strenge  domain.Strenge
		nutzlast map[string]any
		grund    string // leer: die Geste geht
		zone     domain.Zone
	}{
		{
			name:    "ohne Objekt wird nichts gedeutet",
			strenge: domain.StrengeBegleitet,
			nutzlast: map[string]any{
				"zone": "tisch", "sitz": "sitz-1",
			},
			grund: "nennt kein Objekt",
		},
		{
			name:    "eine Zone, die es nicht gibt",
			strenge: domain.StrengeBegleitet,
			nutzlast: map[string]any{
				"objekt": "sitz-1/wald/0", "zone": "nirgends", "sitz": "sitz-1",
			},
			grund: "keine Zone",
		},
		{
			// **CR 305.1**: ein Land wird gespielt, nicht gewirkt.
			name:    "ein Land auf den Stapel",
			strenge: domain.StrengeBegleitet,
			nutzlast: map[string]any{
				"objekt": "sitz-1/wald/0", "zone": "stapel", "sitz": "sitz-1",
			},
			grund: "CR 305.1",
		},
		{
			// **Die Ziele gehen mit** — hier eine, die keine Liste ist, und
			// das Wirken sagt es; ohne Weitergabe käme der Zauber ohne Ziel
			// auf den Stapel.
			name:    "ein Zauber mit seinen Zielen",
			strenge: domain.StrengeBegleitet,
			nutzlast: map[string]any{
				"objekt": "sitz-1/hand/0", "zone": "stapel", "sitz": "sitz-1",
				"ziele": "keine Liste",
			},
			grund: "keine Liste",
		},
		{
			name:    "streng deutet eine Geste in den Friedhof nicht",
			strenge: domain.StrengeStreng,
			nutzlast: map[string]any{
				"objekt": "sitz-1/hand/0", "zone": "friedhof", "sitz": "sitz-1",
			},
			grund: "deutet die Geste nichts",
		},
		{
			name:    "begleitet schiebt eine Handkarte in den Friedhof",
			strenge: domain.StrengeBegleitet,
			nutzlast: map[string]any{
				"objekt": "sitz-1/hand/0", "zone": "friedhof", "sitz": "sitz-1",
			},
			zone: domain.ZoneFriedhof,
		},
		{
			name:    "begleitet schiebt eine Karte vom Schlachtfeld",
			strenge: domain.StrengeBegleitet,
			nutzlast: map[string]any{
				"objekt": "sitz-1/wald/1", "zone": "friedhof", "sitz": "sitz-1",
			},
			zone: domain.ZoneFriedhof,
		},
		{
			// **Die Stelle geht mit** — und eine Stelle im Exil ist ein
			// Irrtum, den das Schieben benennt (das Exil führt hier keine
			// Reihenfolge, `Zone.Geordnet`); ohne Weitergabe läge die Karte
			// still darin.
			name:    "eine Stelle im Exil wird benannt",
			strenge: domain.StrengeBegleitet,
			nutzlast: map[string]any{
				"objekt": "sitz-1/hand/0", "zone": "exil", "sitz": "sitz-1",
				"stelle": "unten",
			},
			grund: "keine Stelle",
		},
		{
			// **Die Stelle geht mit** — in die Bibliothek nach oben.
			name:    "begleitet legt mit Stelle in die Bibliothek",
			strenge: domain.StrengeBegleitet,
			nutzlast: map[string]any{
				"objekt": "sitz-1/hand/0", "zone": "bibliothek", "sitz": "sitz-1",
				"stelle": "oben",
			},
			zone: domain.ZoneBibliothek,
		},
	}

	for _, fall := range tests {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			partie := mitStrenge(inDieHauptphase(t, einAuskunftsstand(t)),
				fall.strenge)
			danach, fehler := geste(t, partie, fall.nutzlast)

			if fall.grund != "" {
				if !errors.Is(fehler, domain.ErrValidation) ||
					!strings.Contains(fehler.Error(), fall.grund) {
					t.Errorf("Ablehnung ohne Grund %q: %v", fall.grund, fehler)
				}

				return
			}

			kennung, _ := fall.nutzlast["objekt"].(string) //nolint:errcheck // die Tabelle nennt immer eines
			if fehler != nil || objektZone(danach, kennung) != fall.zone {
				t.Errorf("Zone %q, erwartet %q: %v",
					objektZone(danach, kennung), fall.zone, fehler)
			}
		})
	}
}
