package service_test

// **Der Abgleich rechnet je Druckzeile, nicht je Alias** — Werkstück U9 von
// Phase 5.1 (#1420),
// seit A2 (#1441) einen Platz ohne Alias geben kann: ein vorläufiger Druck
// zählt die Exemplare **seiner Zeile**, und er heißt ungeklärt — nicht
// unbekannt, das ist die eigene Karte ohne Verweis.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinVorlaeufigerPlatzZaehltDieExemplareSeinerZeile(t *testing.T) {
	werk := neuesPruefwerk(t)

	// **Zwei vorläufige Exemplare der Zeile 9**, ohne Alias.
	werk.sammlung.jeZeile[9] = []*domain.Exemplar{{DruckID: 9, Menge: 2, Name: "Eigenbau"}}

	if _, err := werk.bau.PlatzLegen(t.Context(), werk.deck.ID, domain.Deckplatz{
		DruckID: 9, Menge: 3, Bereich: domain.BereichHauptdeck,
	}); err != nil {
		t.Fatalf("legen: %v", err)
	}

	abgleich, err := werk.abgleich.Fuer(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("abgleichen: %v", err)
	}

	if len(abgleich.Zeilen) != 1 {
		t.Fatalf("%d Zeilen, erwartet 1: %+v", len(abgleich.Zeilen), abgleich.Zeilen)
	}

	zeile := abgleich.Zeilen[0]
	if zeile.DruckID != 9 || !zeile.Vorlaeufig || zeile.Gebraucht != 3 || zeile.Vorhanden != 2 || zeile.Fehlend != 1 {
		t.Errorf("%+v", zeile)
	}

	if abgleich.Unbekannt != 0 || abgleich.Fehlend != 1 {
		t.Errorf("unbekannt %d, fehlend %d", abgleich.Unbekannt, abgleich.Fehlend)
	}
}

// **Ein Platz mit Alias trägt seine Druckzeile ebenso** — und wird weiter
// über den Alias gezählt.
func TestEinPlatzMitAliasTraegtSeineZeile(t *testing.T) {
	werk := neuesPruefwerk(t)
	werk.besitze(druckA, 1)

	if _, err := werk.bau.PlatzLegen(t.Context(), werk.deck.ID, domain.Deckplatz{
		Printing: druckA, DruckID: 4, Menge: 1, Bereich: domain.BereichHauptdeck,
	}); err != nil {
		t.Fatalf("legen: %v", err)
	}

	abgleich, err := werk.abgleich.Fuer(t.Context(), werk.deck.ID)
	if err != nil {
		t.Fatalf("abgleichen: %v", err)
	}

	if abgleich.Zeilen[0].DruckID != 4 || abgleich.Zeilen[0].Vorlaeufig || abgleich.Zeilen[0].Vorhanden != 1 {
		t.Errorf("%+v", abgleich.Zeilen[0])
	}
}

// **Und die Zeile ohne Alias meldet ihren Lesefehler.**
func TestEinLesefehlerDerZeileHaeltDenAbgleichAn(t *testing.T) {
	werk := neuesPruefwerk(t)
	werk.sammlung.zeilenfehler = errKaputt

	if _, err := werk.bau.PlatzLegen(t.Context(), werk.deck.ID, domain.Deckplatz{
		DruckID: 9, Menge: 1, Bereich: domain.BereichHauptdeck,
	}); err != nil {
		t.Fatalf("legen: %v", err)
	}

	if _, err := werk.abgleich.Fuer(t.Context(), werk.deck.ID); err == nil {
		t.Error("der Lesefehler kam durch")
	}
}
