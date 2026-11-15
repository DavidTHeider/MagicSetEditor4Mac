package domain_test

// **Die benannten Wege, über die eine Fähigkeit den Zustand ändert** —
// Werkstück M8 von Phase 5.5
// (#1610).
//
// # Warum das die Zusage des Werkstücks ist
//
// **Das Issue verlangt „einen Sandkasten mit Grenzen": eine Fähigkeit ändert
// den Zustand nur über benannte Wege.** Genau das ist [domain.Spielwirkung] —
// ein Skript kann sie **erzeugen**, aber nicht anwenden, und anwenden kann sie
// nur die Domäne.
//
// **Und CR 101.2 steht darin**: was nicht geschehen kann, wird ausgelassen.
// Eine Wirkung, die ins Leere geht, gibt `false` zurück und ändert nichts.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jede benannte Wirkung ändert genau das, was sie sagt.**
func TestJedeWirkungAendertWasSieSagt(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name    string
		wirkung domain.Spielwirkung
		pruefe  func(t *testing.T, danach domain.Partie)
	}{
		{
			name: "ziehen",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungZiehen, Wer: "sitz-1", Menge: 2,
			},
			pruefe: func(t *testing.T, danach domain.Partie) {
				if hand := len(handVon(danach, "sitz-1")); hand != 2 {
					t.Errorf("%d Karten auf der Hand", hand)
				}
			},
		},
		{
			name: "leben",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungLeben, Wer: "sitz-2", Menge: -3,
			},
			pruefe: func(t *testing.T, danach domain.Partie) {
				if leben := lebenVon(t, danach, "sitz-2"); leben != 17 {
					t.Errorf("%d Leben", leben)
				}
			},
		},
		{
			name: "schaden",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungSchaden, Objekt: "sitz-1/baer/0", Menge: 1,
			},
			pruefe: func(t *testing.T, danach domain.Partie) {
				if schaden := objektVon(t, danach,
					"sitz-1/baer/0").Schaden; schaden != 1 {
					t.Errorf("%d Schaden", schaden)
				}
			},
		},
		{
			name: "zerstoeren",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungZerstoeren, Objekt: "sitz-1/baer/0",
			},
			pruefe: func(t *testing.T, danach domain.Partie) {
				if zone := objektVon(t, danach,
					"sitz-1/baer/0").Zone; zone != domain.ZoneFriedhof {
					t.Errorf("der Bär liegt in %q", zone)
				}
			},
		},
		{
			name: "marke",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungMarke, Objekt: "sitz-1/baer/0",
				Marke: domain.MarkePlusEins, Menge: 2,
			},
			pruefe: func(t *testing.T, danach domain.Partie) {
				staerke, _ := danach.Werte(objektVon(t, danach, "sitz-1/baer/0"))
				if staerke.Zahl != 4 {
					t.Errorf("Stärke %d", staerke.Zahl)
				}
			},
		},
		{
			name: "mahlen",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungMahlen, Wer: "sitz-1", Menge: 2,
			},
			pruefe: func(t *testing.T, danach domain.Partie) {
				friedhof := 0

				for _, objekt := range danach.Objekte {
					if objekt.Zone == domain.ZoneFriedhof {
						friedhof++
					}
				}

				if friedhof != 2 {
					t.Errorf("%d Karten im Friedhof", friedhof)
				}
			},
		},
		{
			name: "exilieren",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungExilieren, Objekt: "sitz-1/baer/0",
			},
			pruefe: func(t *testing.T, danach domain.Partie) {
				if zone := objektVon(t, danach,
					"sitz-1/baer/0").Zone; zone != domain.ZoneExil {
					t.Errorf("der Bär liegt in %q", zone)
				}
			},
		},
		{
			name: "enttappen",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungEnttappen, Objekt: "sitz-1/baer/0",
			},
			pruefe: func(t *testing.T, danach domain.Partie) {
				if objektVon(t, danach, "sitz-1/baer/0").Getappt {
					t.Error("der Bär ist getappt")
				}
			},
		},
		{
			name: "eine Marke nehmen",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungMarke, Objekt: "sitz-1/baer/0",
				Marke: domain.MarkePlusEins, Menge: -3,
			},
			pruefe: func(t *testing.T, danach domain.Partie) {
				// **Unter null gibt es keine** — CR 122.1.
				staerke, _ := danach.Werte(objektVon(t, danach, "sitz-1/baer/0"))
				if staerke.Zahl != 2 {
					t.Errorf("Stärke %d", staerke.Zahl)
				}
			},
		},
		{
			// **Eine zweite Marke lässt die erste liegen** — der Fall, in
			// dem die Abbildung wirklich kopiert wird.
			name: "eine zweite Markenart",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungMarke, Objekt: "sitz-1/marken/0",
				Marke: domain.MarkeMinusEins, Menge: 1,
			},
			pruefe: func(t *testing.T, danach domain.Partie) {
				objekt := objektVon(t, danach, "sitz-1/marken/0")

				if objekt.Zaehler[domain.MarkePlusEins] != 2 ||
					objekt.Zaehler[domain.MarkeMinusEins] != 1 {
					t.Errorf("Zähler %v", objekt.Zaehler)
				}
			},
		},
		{
			name: "tappen",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungTappen, Objekt: "sitz-1/baer/0",
			},
			pruefe: func(t *testing.T, danach domain.Partie) {
				if !objektVon(t, danach, "sitz-1/baer/0").Getappt {
					t.Error("der Bär ist nicht getappt")
				}
			},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			danach, ging := eineWirkungspartie(t).Anwenden(fall.wirkung)
			if !ging {
				t.Fatalf("die Wirkung %+v ging nicht", fall.wirkung)
			}

			fall.pruefe(t, danach)
		})
	}
}

// **Was nicht geschehen kann, wird ausgelassen** — CR 101.2.
func TestWasNichtGeschehenKannWirdAusgelassen(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name    string
		wirkung domain.Spielwirkung
	}{
		{
			name: "ein Sitz, den es nicht gibt",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungLeben, Wer: "weg", Menge: 1,
			},
		},
		{
			name: "ein Objekt, das es nicht gibt",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungZerstoeren, Objekt: "weg",
			},
		},
		{
			name: "eine Art, die niemand kennt",
			wirkung: domain.Spielwirkung{
				Art: "fliegen lernen", Wer: "sitz-1",
			},
		},
		{
			name: "null von etwas",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungZiehen, Wer: "sitz-1", Menge: 0,
			},
		},
		{
			name: "mahlen an einem Sitz, den es nicht gibt",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungMahlen, Wer: "weg", Menge: 1,
			},
		},
		{
			name: "null Schaden",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungSchaden, Objekt: "sitz-1/baer/0", Menge: 0,
			},
		},
		{
			name: "eine Marke ohne Namen",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungMarke, Objekt: "sitz-1/baer/0", Menge: 1,
			},
		},
		{
			name: "null Marken",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungMarke, Objekt: "sitz-1/baer/0",
				Marke: domain.MarkePlusEins,
			},
		},
		{
			name: "exilieren, was nicht da ist",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungExilieren, Objekt: "weg",
			},
		},
		{
			name: "tappen, was nicht da ist",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungTappen, Objekt: "weg",
			},
		},
		{
			name: "null mahlen",
			wirkung: domain.Spielwirkung{
				Art: domain.WirkungMahlen, Wer: "sitz-1",
			},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			vorher := eineWirkungspartie(t)

			danach, ging := vorher.Anwenden(fall.wirkung)
			if ging {
				t.Fatalf("die Wirkung %+v ging", fall.wirkung)
			}

			if len(danach.Objekte) != len(vorher.Objekte) {
				t.Errorf("die Partie hat sich geändert")
			}
		})
	}
}

// eineWirkungspartie gibt eine Partie mit einem Bären und vier Karten in der
// Bibliothek.
func eineWirkungspartie(t *testing.T) domain.Partie {
	t.Helper()

	partie := eineAusloeserpartie(t)

	// **Ein Objekt, das schon Marken trägt** — damit die Kopie der
	// Abbildung einen Fall hat.
	partie.Objekte = append(append([]domain.Spielobjekt{}, partie.Objekte...),
		domain.Spielobjekt{
			Kennung: "sitz-1/marken/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneTisch, Karte: "baer",
			Zaehler: map[string]int{domain.MarkePlusEins: 2},
		})

	return partie
}
