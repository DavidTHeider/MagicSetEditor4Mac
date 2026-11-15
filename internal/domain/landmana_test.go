package domain_test

// **Ein Grundland tappt für Mana** — Werkstück M3 der Regelmaschine
// (#1605).
//
// # Wogegen hier geprüft wird
//
// **CR 305.6** (Fassung 2026-08-19): ein Land mit einem Grundlandtyp hat die
// innewohnende Fähigkeit, für sein Mana zu tappen — auch wenn kein Text
// darauf steht. Das ist die **einzige** Manafähigkeit, die dieses Programm
// ohne `M8` kennt: sie steht in der Regel und nicht im Kartentext.
//
// **Und CR 605.3b**: eine Manafähigkeit geht nicht auf den Stapel. Hier
// heißt das: der Vorrat steht **sofort** da, und niemand kann antworten.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einLandtisch gibt eine Partie mit einem Grundland auf dem Tisch.
func einLandtisch(t *testing.T, untertyp string) domain.Partie {
	t.Helper()

	partie := einePartie().ZugBeginnen("sitz-1")

	partie.Sitze[0].Deck = domain.Partiedeck{
		Abdruck: "abcd",
		Karten: []domain.Abzugskarte{{
			Kennung: "land", Name: untertyp, Menge: 1,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"super type": "Basic Land", "sub type": untertyp,
			}}},
		}},
	}

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/land/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneTisch, Karte: "land",
	})

	return partie
}

// **Jedes der fünf Grundländer gibt sein Mana** — CR 305.6.
func TestJedesGrundlandGibtSeinMana(t *testing.T) {
	t.Parallel()

	faelle := map[string]domain.Manaart{
		"Plains":   domain.ManaWeiss,
		"Island":   domain.ManaBlau,
		"Swamp":    domain.ManaSchwarz,
		"Mountain": domain.ManaRot,
		"Forest":   domain.ManaGruen,
	}

	for untertyp, erwartet := range faelle {
		t.Run(untertyp, func(t *testing.T) {
			t.Parallel()

			partie := einLandtisch(t, untertyp)

			danach, fehler := partie.LandFuerMana("sitz-1", "sitz-1/land/0")
			if fehler != nil {
				t.Fatalf("tappen: %v", fehler)
			}

			if zahl := danach.Sitze[0].Manavorrat.Zahl(erwartet); zahl != 1 {
				t.Errorf("%q gab %v", untertyp, danach.Sitze[0].Manavorrat)
			}

			// **Und es liegt getappt da** — die Kosten der Fähigkeit.
			if !danach.Objekte[len(danach.Objekte)-1].Getappt {
				t.Error("das Land steht ungetappt da")
			}
		})
	}
}

// **Ein getapptes Land gibt nichts mehr** — CR 118.3.
func TestEinGetapptesLandGibtNichtsMehr(t *testing.T) {
	t.Parallel()

	partie := einLandtisch(t, "Forest")

	partie, fehler := partie.LandFuerMana("sitz-1", "sitz-1/land/0")
	if fehler != nil {
		t.Fatalf("erstes tappen: %v", fehler)
	}

	if _, fehler = partie.LandFuerMana("sitz-1", "sitz-1/land/0"); fehler == nil {
		t.Error("dasselbe Land hat zweimal Mana gegeben")
	}
}

// **Ein Land ohne Grundlandtyp hat keine innewohnende Fähigkeit** — CR 305.6.
//
// **Das ist die Gegenprobe** (§10): ohne sie sagte der Lauf darüber nur, dass
// **irgendetwas** Mana gibt — und ein Weg, der jede Karte tappen ließe, käme
// genauso durch.
func TestEinLandOhneGrundtypGibtKeinMana(t *testing.T) {
	t.Parallel()

	partie := einLandtisch(t, "Gate")

	if _, fehler := partie.LandFuerMana("sitz-1", "sitz-1/land/0"); fehler == nil {
		t.Error("ein Land ohne Grundlandtyp hat Mana gegeben")
	}
}

// **Und was fremd ist, tappt niemand** — CR 110.2a.
func TestEinFremdesLandTapptNiemand(t *testing.T) {
	t.Parallel()

	partie := einLandtisch(t, "Forest")

	if _, fehler := partie.LandFuerMana("sitz-2", "sitz-1/land/0"); fehler == nil {
		t.Error("ein fremdes Land wurde getappt")
	}
}
