package msescript_test

// **Eine Kennung allein trennt** —
// #1366.
//
// # Was hier stand und falsch war
//
// „Tragen **beide** eine Kennung, entscheidet sie" — mit dem Zusatz „eine
// Kennung allein trennt nicht. MSE fällt dort auf den Zeichenkettenvergleich
// zurück." Das ist beim Nachlesen nicht haltbar:
//
//	} else if (aw == COMPARE_AS_POINTER || bw == COMPARE_AS_POINTER) {
//	  return ap == bp;
//	}
//
// `src/script/value.cpp`, Zeilen 125–127, Fassung `f9d9356` vom 2021-07-06.
// **Ein `||`, kein `&&`**: eine Seite mit Gegenstandsidentität genügt, und
// dann entscheiden die Zeiger. Ein Nachbau ohne Identität hat einen anderen
// Zeiger und ist damit ungleich.
//
// Der Zeichenkettenvergleich steht **davor** und greift nur, wenn eine Seite
// `COMPARE_AS_STRING` meldet — das sind Farben und Kartenwerte, keine
// Gegenstände.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestEinNachbauOhneKennungIstNichtDerGegenstand(t *testing.T) {
	t.Parallel()

	gegenstand := msescript.Wert{
		Art:       msescript.ArtAbbildung,
		Kennung:   "set",
		Eintraege: []msescript.Eintrag{{Name: "titel", Wert: msescript.Text("Probe")}},
	}

	nachbau := gegenstand
	nachbau.Kennung = ""

	umgebung := msescript.NeueUmgebung(map[string]msescript.Wert{
		"gegenstand": gegenstand,
		"nachbau":    nachbau,
	})

	for _, fall := range []struct {
		quelle   string
		erwartet string
	}{
		{"gegenstand == gegenstand", "true"},
		{"gegenstand == nachbau", "false"},
		{"nachbau == gegenstand", "false"},

		// **Und zwei ohne Kennung entscheiden weiter über die Gestalt** —
		// das sind Sammlungen, und die vergleicht `equal` gliedweise, bevor
		// `compareAs` überhaupt drankommt (`value.cpp`, Zeilen 107–117).
		{"nachbau == nachbau", "true"},
	} {
		t.Run(fall.quelle, func(t *testing.T) {
			t.Parallel()

			wert, fehler := msescript.AuswertenQuelle(t.Context(), fall.quelle, umgebung)
			if fehler != nil {
				t.Fatalf("auswerten: %v", fehler)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%s ergibt %q, MSE gibt %q",
					fall.quelle, wert.AlsText(), fall.erwartet)
			}
		})
	}
}
