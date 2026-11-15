package domain_test

// **Ein Abzug hat eine Obergrenze** — Zweitblick über Phase 7, Befund 2
// (S1): ein eingeladener Gast schickt einen Abzug, und je Exemplar entsteht
// ein Objekt. Ohne Grenze legte ein Platz mit einer Milliarde Exemplaren den
// Wirt lahm — die Grenze der Leitung begrenzt die Nachricht, nicht ihre
// Wirkung.

import (
	"errors"
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinAbzugHatEineObergrenze(t *testing.T) {
	t.Parallel()

	karte := func(kennung string, menge int) domain.Abzugskarte {
		return domain.Abzugskarte{Kennung: kennung, Name: kennung, Menge: menge}
	}

	faelle := map[string]struct {
		karten []domain.Abzugskarte
		gilt   bool
	}{
		"an der Grenze":     {[]domain.Abzugskarte{karte("a", domain.Abzuggrenze)}, true},
		"ein Platz darüber": {[]domain.Abzugskarte{karte("a", domain.Abzuggrenze+1)}, false},
		"zusammen darüber":  {[]domain.Abzugskarte{karte("a", domain.Abzuggrenze), karte("b", 1)}, false},
		"überlaufend groß":  {[]domain.Abzugskarte{karte("a", math.MaxInt), karte("b", math.MaxInt)}, false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := domain.Partiedeck{Abdruck: "x", Karten: fall.karten}.Pruefe()
			if (fehler == nil) != fall.gilt {
				t.Fatalf("gilt %v, Prüfung sagt %v", fall.gilt, fehler)
			}

			if fehler != nil && !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("kein Validierungsfehler: %v", fehler)
			}
		})
	}
}

func TestEinAbzugHatHoechstensSovielePlaetze(t *testing.T) {
	t.Parallel()

	karten := make([]domain.Abzugskarte, 0, domain.Abzuggrenze+1)
	for stelle := range domain.Abzuggrenze + 1 {
		karten = append(karten, domain.Abzugskarte{Kennung: string(rune('a'+stelle%26)) + string(rune(stelle)), Menge: 1})
	}

	if fehler := (domain.Partiedeck{Abdruck: "x", Karten: karten}).Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}
