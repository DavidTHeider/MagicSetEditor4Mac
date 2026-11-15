package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func hinweis(art domain.Hinweisart, name string, anzahl int) domain.Hinweis {
	return domain.Hinweis{
		Art:    art,
		Objekt: domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: name, Name: name},
		Anzahl: anzahl,
	}
}

// **Nach Art, dann die größte Zahl, dann der Name — und nichts ohne Zahl.**
func TestHinweiseWerdenGeordnetUndLeereFallen(t *testing.T) {
	t.Parallel()

	geordnet := domain.OrdneHinweise([]domain.Hinweis{
		hinweis(domain.HinweisPaketBefund, "Paket", 1),
		hinweis(domain.HinweisSetVorbehalte, "Beta", 2),
		hinweis(domain.HinweisDeckFehlend, "Deck", 4),
		hinweis(domain.HinweisSetVorbehalte, "Alpha", 2),
		hinweis(domain.HinweisSetVorbehalte, "Gamma", 5),
		hinweis(domain.HinweisSetVorbehalte, "Leer", 0),
	})

	namen := make([]string, 0, len(geordnet))
	for _, eines := range geordnet {
		namen = append(namen, eines.Objekt.Name)
	}

	erwartet := []string{"Gamma", "Alpha", "Beta", "Deck", "Paket"}
	if len(namen) != len(erwartet) {
		t.Fatalf("geordnet %v, erwartet %v", namen, erwartet)
	}

	for i := range erwartet {
		if namen[i] != erwartet[i] {
			t.Errorf("Stelle %d: %q, erwartet %q", i, namen[i], erwartet[i])
		}
	}

	if leer := domain.OrdneHinweise(nil); leer == nil || len(leer) != 0 {
		t.Errorf("ohne Hinweise: %v", leer)
	}
}
