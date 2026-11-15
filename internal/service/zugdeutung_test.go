package service_test

// **Eine Geste wird im Kern gedeutet** — P3
// (#2371),
// Befund C-03: eine Karte von der Hand aufs Schlachtfeld zu ziehen stellte
// `objekt.schieben` — ohne Landzählung und ohne Stapel, und die Sitzangabe
// des Ziels fiel weg. `objekt.ziehen` nennt Objekt, Zone und Ziel-Sitz; was
// daraus wird, entscheidet der Kern nach dem Profil der Partie.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// geste stellt `objekt.ziehen` als `sitz-1`.
func geste(
	t *testing.T, partie domain.Partie, nutzlast map[string]any,
) (domain.Partie, error) {
	t.Helper()

	kasten := service.NeuerSandkasten(festeMischung{})
	_, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "g1",
		Art: service.BefehlZiehen, Nutzlast: nutzlast,
	})

	return danach, fehler
}

func mitStrenge(partie domain.Partie, strenge domain.Strenge) domain.Partie {
	partie.Regelwerk.Strenge = strenge

	return partie
}

// inDieHauptphase passt bis zur ersten Hauptphase — über den Weg (§3), denn
// CR 116.2a lässt ein Land nur dort zu.
func inDieHauptphase(t *testing.T, partie domain.Partie) domain.Partie {
	t.Helper()

	for range 20 {
		if partie.Zug.Abschnitt.Phase == domain.PhaseHauptVorKampf &&
			partie.StapelOben() == nil {
			return partie
		}

		danach, fehler := partie.Passen(partie.Zug.Prioritaet)
		if fehler != nil {
			t.Fatalf("passen: %v", fehler)
		}

		partie = danach
	}

	t.Fatal("die Hauptphase kam nicht")

	return partie
}

func TestEinGezogenesLandWirdGespielt(t *testing.T) {
	t.Parallel()

	partie := mitStrenge(inDieHauptphase(t, einAuskunftsstand(t)),
		domain.StrengeBegleitet)

	danach, fehler := geste(t, partie, map[string]any{
		"objekt": "sitz-1/wald/0", "zone": "tisch", "sitz": "sitz-1",
	})
	if fehler != nil {
		t.Fatalf("das Land ging nicht: %v", fehler)
	}

	if objektZone(danach, "sitz-1/wald/0") != domain.ZoneTisch ||
		danach.Zug.Laender != 1 {
		t.Errorf("Zone %q, gezählt %d Länder",
			objektZone(danach, "sitz-1/wald/0"), danach.Zug.Laender)
	}

	// **Ein zweites wird abgewiesen, mit Grund** — CR 305.2. Heute legte
	// `objekt.schieben` es ungezählt hin.
	danach.Objekte = append(danach.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/wald/2", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneHand, Karte: "wald",
	})

	_, fehler = geste(t, danach, map[string]any{
		"objekt": "sitz-1/wald/2", "zone": "tisch", "sitz": "sitz-1",
	})
	if !errors.Is(fehler, domain.ErrValidation) ||
		!strings.Contains(fehler.Error(), "CR 305.2") {
		t.Errorf("das zweite Land: %v", fehler)
	}
}

// **Ein Zauber geht auf den Stapel, nicht aufs Feld** — CR 601.2a.
func TestEinGezogenerZauberWirdGewirkt(t *testing.T) {
	t.Parallel()

	partie := mitStrenge(inDieHauptphase(t, einAuskunftsstand(t)),
		domain.StrengeBegleitet)

	danach, fehler := geste(t, partie, map[string]any{
		"objekt": "sitz-1/hand/0", "zone": "tisch", "sitz": "sitz-1",
	})
	if fehler != nil {
		t.Fatalf("das Wirken ging nicht: %v", fehler)
	}

	if objektZone(danach, "sitz-1/hand/0") != domain.ZoneStapel {
		t.Errorf("der Zauber liegt in %q", objektZone(danach, "sitz-1/hand/0"))
	}
}

func TestEineGesteOhnePassendenZielSitzWirdErklaert(t *testing.T) {
	t.Parallel()

	partie := mitStrenge(einAuskunftsstand(t), domain.StrengeBegleitet)

	tests := []struct {
		name     string
		nutzlast map[string]any
		grund    string
	}{
		{
			name: "ohne Ziel-Sitz",
			nutzlast: map[string]any{
				"objekt": "sitz-1/wald/0", "zone": "tisch",
			},
			grund: "Ziel-Sitz",
		},
		{
			name: "mit einem Sitz, den es nicht gibt",
			nutzlast: map[string]any{
				"objekt": "sitz-1/wald/0", "zone": "tisch", "sitz": "sitz-9",
			},
			grund: "sitz-9",
		},
		{
			name: "in die Hand eines anderen Sitzes",
			nutzlast: map[string]any{
				"objekt": "sitz-1/wald/0", "zone": "hand", "sitz": "sitz-2",
			},
			grund: "CR 400.3",
		},
		{
			name: "auf den Tisch eines anderen Sitzes",
			nutzlast: map[string]any{
				"objekt": "sitz-1/wald/1", "zone": "tisch", "sitz": "sitz-2",
			},
			grund: "CR 110.2",
		},
	}

	for _, fall := range tests {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, fehler := geste(t, partie, fall.nutzlast)
			if !errors.Is(fehler, domain.ErrValidation) ||
				!strings.Contains(fehler.Error(), fall.grund) {
				t.Errorf("Ablehnung ohne Grund %q: %v", fall.grund, fehler)
			}
		})
	}
}

// **Freies Schieben bleibt** — ausdrücklich, und im Sandkasten ohnehin. Unter
// strenger Prüfung gibt es es nicht.
func TestFreiesSchiebenFolgtDemProfil(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		strenge domain.Strenge
		frei    bool
		geht    bool
	}{
		{"im Sandkasten ist jede Geste ein Schieben", domain.StrengeSandkasten, false, true},
		{"begleitet schiebt, wer es ausdrücklich will", domain.StrengeBegleitet, true, true},
		{"streng gibt es kein freies Schieben", domain.StrengeStreng, true, false},
	}

	for _, fall := range tests {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			partie := mitStrenge(einAuskunftsstand(t), fall.strenge)
			danach, fehler := geste(t, partie, map[string]any{
				"objekt": "sitz-1/wald/0", "zone": "tisch", "sitz": "sitz-1",
				"frei": fall.frei,
			})

			if !fall.geht {
				if !errors.Is(fehler, domain.ErrValidation) {
					t.Fatalf("streng ging es: %v", fehler)
				}

				return
			}

			if fehler != nil {
				t.Fatalf("das Schieben ging nicht: %v", fehler)
			}

			// **Geschoben und nicht gespielt** — kein Land gezählt.
			if objektZone(danach, "sitz-1/wald/0") != domain.ZoneTisch ||
				danach.Zug.Laender != 0 {
				t.Errorf("Zone %q, gezählt %d",
					objektZone(danach, "sitz-1/wald/0"), danach.Zug.Laender)
			}
		})
	}
}
