package service_test

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// quelleGibtUnlesbar lässt die Quelle dieselbe Reihe nennen, aber in einer
// Form, die sich nicht deuten lässt.
func (v *preisvorrichtung) quelleGibtUnlesbar() {
	v.quelle.preise = nil
	v.quelle.ungeprueft = []scryfall.Preis{{
		Waehrung: "eur", Finish: domain.FinishNonfoil,
	}}
	v.quelle.fehler = nil
}

// **Ein unlesbarer Preis verschlechtert den gültigen Stand nicht** — F04
// (#2246).
//
// Die Quelle nennt die Reihe, nur nicht lesbar. Das ist weder ein Wegfall
// noch ein neuer Wert: der letzte bekannte bleibt, und er sagt, wie alt er
// ist (§16, Satz 5).
func TestEinUnlesbarerPreisLaesstDenLetztenStehen(t *testing.T) {
	t.Parallel()

	vor := neuePreisvorrichtung(t)
	vor.quelleGibt(417)
	mussHolen(t, vor)

	vor.jetzt = vor.jetzt.Add(time.Hour)
	vor.quelleGibtUnlesbar()

	staende, _, _, fehler := vor.pflege.Holen(t.Context(), "druck-1")
	if fehler != nil {
		t.Fatalf("zweiter Abruf: %v", fehler)
	}

	// **Was der Abruf zurückgibt, nennt keinen Wert**, den es nicht gibt.
	if len(staende) != 0 {
		t.Errorf("der Abruf gab %+v", staende)
	}

	neueste, err := vor.pflege.Neueste(t.Context(), "druck-1")
	if err != nil {
		t.Fatalf("Neueste: %v", err)
	}

	if len(neueste) != 1 || neueste[0].Lage != domain.PreisGenannt || neueste[0].Wert != 417 {
		t.Errorf("der neueste Stand ist %+v — erwartet der genannte von vorhin", neueste)
	}
}
