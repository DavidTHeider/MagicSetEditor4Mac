package domain_test

// **Eine Quelle und ihre Familie** — Werkstück A3 von Phase 5.3
// (#1442).
//
// # Warum es eine Familie gibt
//
// MTGJSON ist ein **Sammler**, kein Händler: eine Antwort trägt die Preise von
// `cardkingdom`, `cardmarket`, `cardsphere`, `tcgplayer` und `cardhoarder`.
// Jeder davon nennt einen anderen Preis, und eine Quelle für alle wäre eine
// Reihe für fünf Aussagen — deshalb heißen sie `mtgjson:cardkingdom` und so
// fort.
//
// **Der Widerruf gilt aber der Antwort**, nicht dem Anbieter: wenn MTGJSON für
// einen Druck nichts mehr nennt, ist **jede** seiner Reihen weggefallen. Ohne
// diesen Begriff verglich der Dienst `mtgjson` mit `mtgjson:cardkingdom`,
// fand sie ungleich und widerrief nichts.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineQuelleGehoertZuIhrerFamilie(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		quelle  domain.Preisquelle
		familie domain.Preisquelle
		dazu    bool
	}{
		"sie selbst":          {"mtgjson", "mtgjson", true},
		"ein Anbieter darin":  {"mtgjson:cardkingdom", "mtgjson", true},
		"ein zweiter":         {"mtgjson:cardhoarder", "mtgjson", true},
		"eine fremde Quelle":  {"scryfall", "mtgjson", false},
		"ein ähnlicher Name":  {"mtgjsonx", "mtgjson", false},
		"die andere Richtung": {"mtgjson", "mtgjson:cardkingdom", false},
		"leer":                {"", "mtgjson", false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fall.quelle.GehoertZu(fall.familie) != fall.dazu {
				t.Errorf("%q zu %q: erwartet %v",
					string(fall.quelle), string(fall.familie), fall.dazu)
			}
		})
	}
}
