package msepaket_test

// **Der Vertrag eines Paketverweises** — FZ-01
// (#2135).
// Leerraum am Rand eines Teils ist mehrdeutig und wird abgewiesen; Leerraum
// im Namen und ein doppelter Schrägstrich bleiben, und was angenommen wird,
// kommt zurück, wie es ging.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestDerVertragEinesPaketverweises(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		verweis string
		gilt    msepaket.Paketpfad
		falsch  bool
	}{
		"gewöhnlich":               {verweis: "/magic.mse-game/card", gilt: msepaket.Paketpfad{Paket: "magic.mse-game", Datei: "card"}},
		"Leerraum außen":           {verweis: "  /magic.mse-game/card ", gilt: msepaket.Paketpfad{Paket: "magic.mse-game", Datei: "card"}},
		"Leerraum im Namen":        {verweis: "/magic.mse-game/zwei worte", gilt: msepaket.Paketpfad{Paket: "magic.mse-game", Datei: "zwei worte"}},
		"doppelter Schrägstrich":   {verweis: "/magic.mse-game//card", gilt: msepaket.Paketpfad{Paket: "magic.mse-game", Datei: "card"}},
		"Leerraum als Datei":       {verweis: "/magic.mse-game/ /", falsch: true},
		"Leerraum am Ende innen":   {verweis: "/magic.mse-game/card /x", falsch: true},
		"Leerraum am Anfang innen": {verweis: "/magic.mse-game/ card", falsch: true},
		"Leerraum am Paketnamen":   {verweis: "/ magic.mse-game/card", falsch: true},
		"Elternteil":               {verweis: "/magic.mse-game/../card", falsch: true},
		"Punkt":                    {verweis: "/magic.mse-game/./card", falsch: true},
		"kein UTF-8":               {verweis: "/magic.mse-game/\xff", falsch: true},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pfad, fehler := msepaket.ZerlegePaketpfad(fall.verweis)
			if fall.falsch {
				if !errors.Is(fehler, msepaket.ErrPaketpfad) {
					t.Errorf("angenommen: %+v (%v)", pfad, fehler)
				}

				return
			}

			if fehler != nil || pfad != fall.gilt {
				t.Fatalf("%+v (%v), erwartet %+v", pfad, fehler, fall.gilt)
			}

			if zurueck, err := msepaket.ZerlegePaketpfad("/" + pfad.Paket + "/" + pfad.Datei); err != nil || zurueck != pfad {
				t.Errorf("kommt nicht zurück: %+v (%v)", zurueck, err)
			}
		})
	}
}

func TestEinEigenverweisHatDenselbenVertrag(t *testing.T) {
	t.Parallel()

	if _, fehler := msepaket.ZerlegeEigenpfad("magic.mse-game", "watermarks /x"); !errors.Is(fehler, msepaket.ErrPaketpfad) {
		t.Errorf("Leerraum am Rand eines Teils angenommen: %v", fehler)
	}

	if pfad, fehler := msepaket.ZerlegeEigenpfad("magic.mse-game", "zwei worte/x"); fehler != nil || pfad.Datei != "zwei worte/x" {
		t.Errorf("Leerraum im Namen: %+v (%v)", pfad, fehler)
	}
}
