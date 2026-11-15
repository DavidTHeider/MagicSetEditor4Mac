package domain_test

// **Die Legendenregel trifft nur gleichnamige Legenden** — CR 704.5j, wie in
// `legendenregel.go` belegt. Gefunden von der Zufallsmutation (B05, #2224): ein
// `&&`, das zum `||` wurde, legte **jede** andere Legende des Spielers in den
// Friedhof, und kein Test sah es.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineLegendeAnderenNamensBleibtLiegen(t *testing.T) {
	t.Parallel()

	partie := eineLegendenpartie(t)
	deck := partie.Sitze[0].Deck
	deck.Karten = append(append([]domain.Abzugskarte{}, deck.Karten...), domain.Abzugskarte{
		Kennung: "heldin", Name: "Heldin", Menge: 1,
		Seiten: []domain.Abzugsseite{{Felder: map[string]string{
			"name": "Heldin", "super type": "Legendary Creature",
			"power": "1", "toughness": "1",
		}}},
	})
	partie.Sitze[0].Deck = deck
	partie.Objekte = append(append([]domain.Spielobjekt{}, partie.Objekte...), domain.Spielobjekt{
		Kennung: "sitz-1/heldin/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneTisch, Karte: "heldin",
	})

	danach, fehler := partie.Zustandspruefung().Legende("sitz-1", "sitz-1/held/0")
	if fehler != nil {
		t.Fatalf("legende: %v", fehler)
	}

	if zone := objektMit(t, danach, "sitz-1/heldin/0").Zone; zone != domain.ZoneTisch {
		t.Errorf("die Legende anderen Namens liegt in %q", zone)
	}

	if zone := objektMit(t, danach, "sitz-1/held/1").Zone; zone != domain.ZoneFriedhof {
		t.Errorf("die gleichnamige Legende liegt in %q", zone)
	}
}
