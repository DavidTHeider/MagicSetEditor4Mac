package domain_test

// **Ein Sitz, der auf das Deck seines Gastes wartet** — P9b
// (#2377).
// Der Wirt legt die Partie an, bevor der Gast da ist; das Deck bringt der
// Gast beim Hinsetzen mit. Bis dahin steht der Sitz ohne Deck da — **nur**
// er, und nur, wenn er jemanden erwartet.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineEinladung ist die kleinste Partie mit einem wartenden Sitz.
func eineEinladung() domain.Partie {
	partie := einePartie()
	partie.Sitze[1].Deck = domain.Partiedeck{}
	partie.Sitze[1].Eingeladen = "peer-bo"

	return partie
}

func TestEinEingeladenerSitzDarfOhneDeckDastehen(t *testing.T) {
	t.Parallel()

	if fehler := eineEinladung().Pruefe(); fehler != nil {
		t.Fatalf("abgewiesen: %v", fehler)
	}
}

func TestEinSitzOhneDeckUndOhneEinladungBleibtFalsch(t *testing.T) {
	t.Parallel()

	partie := eineEinladung()
	partie.Sitze[1].Eingeladen = ""

	if fehler := partie.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Der Wirt wartet nicht auf sich selbst** — sein Deck liegt beim Anlegen.
func TestDerWirtWirdNichtEingeladen(t *testing.T) {
	t.Parallel()

	partie := eineEinladung()
	partie.Sitze[0].Deck = domain.Partiedeck{}
	partie.Sitze[0].Eingeladen = "peer-ada"

	if fehler := partie.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Die Sicht sagt, dass der Sitz wartet — nicht, auf wen.** Eine
// Peerkennung ist kein Name am Tisch, und wer zuschaut, braucht sie nicht.
func TestDieSichtSagtDassEinSitzWartet(t *testing.T) {
	t.Parallel()

	for _, fuer := range []string{"sitz-1", "sitz-2", "zuschauer"} {
		sicht := eineEinladung().Sicht(fuer)

		if !sicht.Sitze[1].Wartet || sicht.Sitze[0].Wartet {
			t.Errorf("für %s: %+v", fuer, sicht.Sitze)
		}
	}
}

// **Mit Deck wartet er nicht mehr** — auch wenn er eingeladen war.
func TestMitDeckWartetErNichtMehr(t *testing.T) {
	t.Parallel()

	partie := eineEinladung()
	partie.Sitze[1].Deck = einDeckabzug()

	if partie.Sicht("sitz-1").Sitze[1].Wartet {
		t.Error("der Sitz wartet trotz Deck")
	}
}

// **Wen er eingeladen hat, sieht nur der Wirt** — er gibt den Sitz frei und
// muss ihn dem Richtigen geben. Alle anderen lesen nur, dass jemand kommt.
func TestNurDerWirtSiehtWenErEingeladenHat(t *testing.T) {
	t.Parallel()

	partie := eineEinladung()

	if sicht := partie.Sicht("sitz-1"); sicht.Sitze[1].Eingeladen != "peer-bo" {
		t.Errorf("der Wirt sieht %q", sicht.Sitze[1].Eingeladen)
	}

	for _, fuer := range []string{"sitz-2", "zuschauer"} {
		if sicht := partie.Sicht(fuer); sicht.Sitze[1].Eingeladen != "" {
			t.Errorf("%s sieht %q", fuer, sicht.Sitze[1].Eingeladen)
		}
	}
}
