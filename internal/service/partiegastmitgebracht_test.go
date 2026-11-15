package service_test

// **Der Gast bringt sein Deck beim ersten Mal mit** — P9b
// (#2377).
// Nur beim ersten Mal: wer schon dabei war, sitzt mit seinem Deck, und ein
// zweites wäre eine Frage, die der Wirt nicht mehr stellt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerGastBringtBeimErstenMalSeinDeckMit(t *testing.T) {
	t.Parallel()

	gast, naht := einGast(&nehmerdoppelGast{})
	gast.MitMitgebracht(&domain.Mitgebracht{Leihe: "l1"})

	if fehler := gast.Kehrezurueck(t.Context()); fehler != nil {
		t.Fatalf("Kehrezurueck: %v", fehler)
	}

	if wunsch := derWunsch(t, naht); wunsch.Mitgebracht == nil || wunsch.Mitgebracht.Leihe != "l1" {
		t.Errorf("der Wunsch ist %+v", wunsch)
	}
}

func TestWerSchonDabeiWarBringtNichtsMehrMit(t *testing.T) {
	t.Parallel()

	gast, naht := einGast(&nehmerdoppelGast{})
	gast.MitStandhalter(&standdoppel{folge: 4, abdruck: "vier"})
	gast.MitMitgebracht(&domain.Mitgebracht{Leihe: "l1"})

	if fehler := gast.Kehrezurueck(t.Context()); fehler != nil {
		t.Fatalf("Kehrezurueck: %v", fehler)
	}

	if wunsch := derWunsch(t, naht); wunsch.Mitgebracht != nil {
		t.Errorf("der Wunsch ist %+v", wunsch)
	}
}
