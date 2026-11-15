package domain_test

// **Grenzfälle, die erst die Zufallsmutation zeigte** — B05
// (#2224).
// Je Fall blieb ein Mutant am Leben, weil kein Test die Grenze betrat; die
// Durchsicht steht in `docs/belege/zufallsmutation-durchsicht.json`.

import (
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Null ist ein Betrag** — „ein Betrag ist nicht negativ" (`kostenjournal.go`).
func TestNullIstEinBetrag(t *testing.T) {
	t.Parallel()

	if fehler := (domain.Betrag{Wert: 0, Waehrung: "EUR"}).Pruefe(); fehler != nil {
		t.Errorf("ein Betrag von null wurde abgewiesen: %v", fehler)
	}
}

// **Genau sieben Tage gelten noch** — abgewiesen wird, was **länger** gilt.
func TestEineEinladungVonGenauSiebenTagenGilt(t *testing.T) {
	t.Parallel()

	einladung := eineKontakteinladung()
	einladung.Ablauf = einladungsjetzt.Add(domain.KontakteinladungHoechstens)

	if fehler := einladung.Pruefe(einladungsjetzt); fehler != nil {
		t.Errorf("eine Einladung von genau sieben Tagen wurde abgewiesen: %v", fehler)
	}
}

// **Eine Seite eins hinter der letzten fällt auf die Vorderseite** — CR
// 701.27a, wie in `objekteigenschaften.go` belegt. Ein `>=`, das zum `>`
// wurde, griff dort hinter das Ende der Liste.
func TestEineSeiteHinterDerLetztenFaelltAufDieVorderseite(t *testing.T) {
	t.Parallel()

	partie := eineWirkpartie(t, "")
	objekt := objektMit(t, partie, "sitz-1/blitz/0")
	objekt.Seite = 1

	eigenschaften, gefunden := partie.Eigenschaften(objekt)
	if !gefunden {
		t.Fatal("die Karte wurde nicht gefunden")
	}

	vorn, _ := partie.Eigenschaften(objektMit(t, partie, "sitz-1/blitz/0"))
	if !reflect.DeepEqual(eigenschaften, vorn) {
		t.Errorf("Seite 1 einer einseitigen Karte gab %+v, die Vorderseite %+v", eigenschaften, vorn)
	}
}

// **Vor dem ersten Zug wirkt, wer will** — die Prioritätsregel gilt ab Zug 1
// (`wirken.go`); ein `>`, das zum `>=` wurde, sperrte den Sandkasten vor dem
// Zug.
func TestVorDemErstenZugBrauchtEsKeinePrioritaet(t *testing.T) {
	t.Parallel()

	partie := eineWirkpartie(t, "")
	partie.Zug = domain.Zugstand{}

	if _, fehler := partie.Wirken("sitz-1", "sitz-1/blitz/0", nil); fehler != nil {
		t.Errorf("vor dem ersten Zug wurde das Wirken abgewiesen: %v", fehler)
	}
}
