package domain_test

// **Die zweite öffentliche Seite** — Werkstück G2 von Phase 5.5
// (#1459).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Karte mit `name 2` hat zwei Seiten, eine ohne nicht.**
func TestHatZweiteSeite(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		werte    []domain.Value
		erwartet bool
	}{
		{
			name: "nichtleeres `name 2`",
			werte: []domain.Value{
				{Field: "name", Text: "Delver of Secrets"},
				{Field: "name 2", Text: "Insectile Aberration"},
			},
			erwartet: true,
		},
		{
			name:  "gar kein `name 2`",
			werte: []domain.Value{{Field: "name", Text: "Wald"}},
		},
		{
			name: "leeres `name 2`",
			werte: []domain.Value{
				{Field: "name", Text: "Wald"},
				{Field: "name 2", Text: ""},
			},
		},
		{
			// **Der Fall, für den die Regel am Namen hängt** — ein
			// Planeswalker führt `loyalty cost 2` und hat eine Seite.
			name: "ein anderes Feld mit ` 2`",
			werte: []domain.Value{
				{Field: "name", Text: "Chandra"},
				{Field: "loyalty cost 2", Text: "-3"},
			},
		},
		{
			name: "nur Leerraum in `name 2`",
			werte: []domain.Value{
				{Field: "name", Text: "Wald"},
				{Field: "name 2", Text: "   "},
			},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			karte := &domain.CardRow{Values: fall.werte}

			if karte.HatZweiteSeite() != fall.erwartet {
				t.Errorf("HatZweiteSeite = %v, erwartet %v",
					karte.HatZweiteSeite(), fall.erwartet)
			}

			if domain.WerteHabenZweiteSeite(fall.werte) != fall.erwartet {
				t.Errorf("WerteHabenZweiteSeite = %v, erwartet %v",
					domain.WerteHabenZweiteSeite(fall.werte), fall.erwartet)
			}
		})
	}
}

// **Und eine Karte, die es gar nicht gibt, hat keine zweite Seite.**
//
// **Der Weg kommt vor**: der Rücken wird ohne Karte gezeichnet, und wer die
// Frage dort stellte, bekäme sonst einen Sturz statt einer Antwort.
func TestKeineKarteHatKeineZweiteSeite(t *testing.T) {
	t.Parallel()

	var keine *domain.CardRow

	if keine.HatZweiteSeite() {
		t.Error("eine Karte, die es nicht gibt, hat zwei Seiten")
	}
}

// **Der Zusatz ist ` 2` mit Leerzeichen.**
func TestZweiteSeiteHaengtEinLeerzeichenAn(t *testing.T) {
	t.Parallel()

	if gesehen := domain.ZweiteSeite("casting cost"); gesehen != "casting cost 2" {
		t.Errorf("%q", gesehen)
	}
}
