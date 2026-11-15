package domain_test

// **Wer eine Karte kennt, und ob ein Land Mana gibt** — die zwei Fragen, die
// die Handlungsauskunft aus P3 an die Domäne stellt
// (#2371).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestWerEineKarteKennt(t *testing.T) {
	t.Parallel()

	objekt := func(zone domain.Zone, wer string, verdeckt bool) domain.Spielobjekt {
		return domain.Spielobjekt{
			Kennung: "o", Art: domain.ObjektKarte, Beherrscher: wer,
			Eigner: wer, Zone: zone, Karte: "wald", Verdeckt: verdeckt,
		}
	}

	tests := []struct {
		name   string
		objekt domain.Spielobjekt
		kennt  bool
	}{
		{"die eigene Hand", objekt(domain.ZoneHand, "sitz-1", false), true},
		{"eine fremde Hand", objekt(domain.ZoneHand, "sitz-2", false), false},
		{"die eigene Bibliothek", objekt(domain.ZoneBibliothek, "sitz-1", false), false},
		{"ein fremdes Objekt auf dem Tisch", objekt(domain.ZoneTisch, "sitz-2", false), true},
		{"ein fremdes verdecktes Objekt", objekt(domain.ZoneTisch, "sitz-2", true), false},
		{"ein eigenes verdecktes Objekt", objekt(domain.ZoneTisch, "sitz-1", true), true},
	}

	for _, fall := range tests {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if domain.KennenDieKarte(fall.objekt, "sitz-1") != fall.kennt {
				t.Errorf("kennt sollte %v sein", fall.kennt)
			}
		})
	}
}

func TestObEinLandEinenGrundlandtypTraegt(t *testing.T) {
	t.Parallel()

	partie := eineLandpartie(t)

	tests := []struct {
		name  string
		karte string
		hat   bool
	}{
		{"ein Wald", "wald", true},
		{"ein Spontanzauber", "blitz", false},
		{"eine Karte, die im Abzug fehlt", "gibtesnicht", false},
	}

	for _, fall := range tests {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			objekt := domain.Spielobjekt{
				Kennung: "o", Art: domain.ObjektKarte, Beherrscher: "sitz-1",
				Eigner: "sitz-1", Zone: domain.ZoneTisch, Karte: fall.karte,
			}

			if partie.HatGrundlandart(objekt) != fall.hat {
				t.Errorf("hat sollte %v sein", fall.hat)
			}
		})
	}
}
