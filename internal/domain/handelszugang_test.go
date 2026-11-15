package domain_test

// **Was die empfangende Seite verbuchen kann** — #2495.
//
// Eine Zeile ohne Sprache — „Wunsch aus einem Deck eintragen" legte solche
// an — ließ sich anbieten, annehmen und übergeben und scheiterte erst beim
// Verbuchen an [domain.Exemplar.Validate]. Hier steht **eine** Regel für
// beides: die Zeile, die die Verbuchung bucht, und die Prüfung davor.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func einTausch(positionen ...domain.Handelsposition) domain.Handelsangebot {
	return domain.Handelsangebot{
		Vorgang: "h-2495", Art: domain.HandelTausch,
		Parteien: []string{"anna", "bert"}, Positionen: positionen,
	}
}

var verbuchbar = domain.Handelsposition{
	Von: "anna", Printing: "p-glut", Menge: 1,
	Guete: "nearMint", Sprache: "de", Finish: "nonfoil",
}

func TestEineZeileOhneSpracheIstNichtVerbuchbar(t *testing.T) {
	t.Parallel()

	ohneSprache := verbuchbar
	ohneSprache.Von, ohneSprache.Sprache = "bert", ""

	fehler := einTausch(verbuchbar, ohneSprache).PruefeVerbuchbar()

	var feld *domain.FieldError
	if !errors.As(fehler, &feld) || feld.Key != domain.SchluesselZeileNichtVerbuchbar ||
		feld.Werte["zeile"] != "2" || !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %#v", fehler)
	}
}

func TestVerbuchbarIstDieZeileDieDasVerbuchenBucht(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name  string
		teil  func(*domain.Handelsposition)
		passt bool
	}{
		{"vollständig", func(*domain.Handelsposition) {}, true},
		{"ohne Zustand — drüben entstünde ein Exemplar ohne Zustand", func(p *domain.Handelsposition) { p.Guete = "" }, false},
		{"ohne Ausführung ist es nonfoil", func(p *domain.Handelsposition) { p.Finish = "" }, true},
		{"ein Zustand, den es nicht gibt", func(p *domain.Handelsposition) { p.Guete = "NM" }, false},
		{"eine Ausführung, die es nicht gibt", func(p *domain.Handelsposition) { p.Finish = "gold" }, false},
		{"ohne Druck — eine Druckzeile ist drüben keine", func(p *domain.Handelsposition) { p.Printing, p.DruckID = "", 7 }, false},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			stelle := verbuchbar
			fall.teil(&stelle)

			zeile := stelle.Zugangszeile("h-2495")
			if zeile.Art != domain.JournalZugang || zeile.Vorgang != "h-2495" || zeile.Printing != stelle.Printing {
				t.Errorf("die Zugangszeile: %+v", zeile)
			}

			if fehler := einTausch(stelle).PruefeVerbuchbar(); (fehler == nil) != fall.passt {
				t.Errorf("verbuchbar %v, erwartet %v: %v", fehler == nil, fall.passt, fehler)
			}
		})
	}
}

// **Ein Handel, der keinen Bestand bewegt, bucht nichts** — und hat deshalb
// auch nichts, das verbuchbar sein müsste.
func TestEinHandelOhneBestandIstImmerVerbuchbar(t *testing.T) {
	t.Parallel()

	ohneSprache := verbuchbar
	ohneSprache.Sprache = ""

	teilen := einTausch(ohneSprache)
	teilen.Art = domain.HandelTeilen

	if fehler := teilen.PruefeVerbuchbar(); fehler != nil {
		t.Errorf("es kam %v", fehler)
	}
}
