package service_test

// **Kein Weg zur Einigung über eine Zeile, die drüben nicht zu buchen ist** —
// #2495.
//
// „Wunsch aus einem Deck eintragen" legte Zeilen ohne Zustand, Sprache und
// Ausführung an. Sie ließen sich anbieten, annehmen und übergeben und
// scheiterten erst beim Verbuchen — nach der Übergabe. **Jetzt sagt es der
// Schritt, der auf eine Einigung zuführt**, und er schreibt nichts.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ohneSprache ist ein Tausch, dessen zweite Zeile keine Sprache nennt.
func ohneSprache() domain.Handelsangebot {
	angebot := einHandelsangebot()
	angebot.Positionen = append(angebot.Positionen, domain.Handelsposition{
		Von: "bert", Printing: "p-wunsch", Menge: 1,
	})

	return angebot
}

func TestEinTauschMitZeileOhneSpracheWirdNichtAngeboten(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()

	if _, fehler := leitung.Eroeffnen(t.Context(), ohneSprache()); fehler != nil {
		t.Fatalf("Eroeffnen: %v — ein Entwurf darf unfertig sein", fehler)
	}

	vorher := ablage.geschrieben

	_, fehler := leitung.Schreiten(t.Context(), "h1", domain.SchrittAnbieten, "anna")

	var feld *domain.FieldError
	if !errors.As(fehler, &feld) || feld.Key != domain.SchluesselZeileNichtVerbuchbar || feld.Werte["zeile"] != "2" {
		t.Fatalf("es kam %v", fehler)
	}

	if ablage.geschrieben != vorher || ablage.staende["h1"].Zustand != domain.HandelEntwurf {
		t.Errorf("geschrieben %d, Zustand %q", ablage.geschrieben-vorher, ablage.staende["h1"].Zustand)
	}
}

// **Und angenommen wird es auch nicht**, wenn es so von drüben kam.
func TestEinAngebotMitZeileOhneSpracheWirdNichtAngenommen(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()

	angeboten := ohneSprache()
	angeboten.Zustand = domain.HandelAngeboten
	ablage.staende["h1"] = angeboten

	// **Bert ist die Gegenseite** — eine fehlende Rolle wäre auch eine
	// Validierung, deshalb zählt hier nur der Schlüssel.
	_, fehler := leitung.Schreiten(t.Context(), "h1", domain.SchrittZustimmen, "bert")

	var feld *domain.FieldError
	if !errors.As(fehler, &feld) || feld.Key != domain.SchluesselZeileNichtVerbuchbar {
		t.Fatalf("es kam %v", fehler)
	}

	if ablage.geschrieben != 0 {
		t.Errorf("es wurde %d-mal geschrieben", ablage.geschrieben)
	}

	// **Ein Ausweg bleibt offen**: ablehnen geht.
	if _, fehler := leitung.Schreiten(t.Context(), "h1", domain.SchrittAblehnen, "bert"); fehler != nil {
		t.Errorf("ablehnen: %v", fehler)
	}
}

// **Ein Schritt gilt der Revision, die gesehen wurde** — #2496. Zwischen
// dem Nachlesen im Fenster und dem Schritt kam ein Gegenangebot an; die
// Zustimmung galt dessen Revision, die niemand gesehen hatte.
func TestEinSchrittMitAltemAbdruckWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()
	gesehen := angebotenerHandel(t, leitung)

	// **Dazwischen**: eine neue Revision, wie sie ein Gegenangebot schreibt.
	gegen := gesehen.Fortgeschrieben(gesehen.Positionen)
	gegen.Positionen[0].Menge = 1
	ablage.staende["h1"] = gegen
	vorher := ablage.geschrieben

	_, fehler := leitung.SchreitenAls(t.Context(), "h1", gesehen.Abdruck(),
		domain.SchrittZustimmen, "bert", "bert")
	if !errors.Is(fehler, domain.ErrUeberholt) {
		t.Fatalf("es kam %v", fehler)
	}

	if ablage.geschrieben != vorher {
		t.Errorf("es wurde %d-mal geschrieben", ablage.geschrieben-vorher)
	}

	// **Mit dem Abdruck, der gilt, geht es.**
	if _, fehler := leitung.SchreitenAls(t.Context(), "h1", gegen.Abdruck(),
		domain.SchrittZustimmen, "bert", "bert"); fehler != nil {
		t.Errorf("mit dem neuen Abdruck: %v", fehler)
	}
}
