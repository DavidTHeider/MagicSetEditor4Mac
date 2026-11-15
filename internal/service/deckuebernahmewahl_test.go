package service_test

// **Die Decks eines anderen Profils zur Wahl** — UX-08.AK-04
// (#1680).
//
// **Dieselbe Erlaubnis wie die Übernahme, und davor** (Vermerk 0071): wer
// nicht hinein darf, erfährt auch nicht, welche Decks es dort gibt. Wer es
// darf, bekommt die Liste — und tippt keine Kennung mehr ab.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestMitFreigabeStehenDieFremdenDecksZurWahl(t *testing.T) {
	t.Parallel()

	uebernahme, _ := zweiProfileMitDeck(t, domain.RolleLeser)

	decks, fehler := uebernahme.DeckeZurUebernahme(t.Context(), 2)
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if len(decks) != 1 || decks[0].Name != "Nachtlauf" {
		t.Errorf("zur Wahl stehen %+v", decks)
	}
}

func TestOhneFreigabeStehtNichtsZurWahl(t *testing.T) {
	t.Parallel()

	uebernahme, doppel := zweiProfileMitDeck(t, "")

	if decks, fehler := uebernahme.DeckeZurUebernahme(
		t.Context(), 2); fehler == nil || decks != nil {
		t.Fatalf("es kam %v und %v", decks, fehler)
	}

	if doppel.entnommen {
		t.Error("die fremden Decks wurden trotzdem gelesen")
	}
}

// **Ein Leser drüben darf holen, nicht nur wählen.** Seit Vermerk 0073 reicht
// `RolleLeser` zum Hineingehen, und die Übernahme stellt dieselbe Frage. Bis
// zum 2026-09-25 stand im Fenster und im MCP-Werkzeug „mindestens
// »Mitarbeiter«" — eine Zusage, die seit Phase 4.12 niemand mehr hielt;
// gefunden bei der Sichtprüfung der Wahl (UX-08.AK-04).
func TestEinLeserDruebenDarfHolen(t *testing.T) {
	t.Parallel()

	uebernahme, _ := zweiProfileMitDeck(t, domain.RolleLeser)

	if _, _, fehler := uebernahme.DeckUebernehmen(t.Context(), 2, 7); fehler != nil {
		t.Fatalf("ein Leser drüben durfte nicht holen: %v", fehler)
	}
}
