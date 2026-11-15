package service_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Deck aus einem anderen Spiel gehört nicht in dieses Turnier.**
//
// Ein Turnier hat ein Spiel, ein Deck auch — und bis zum Nachlauf über Phase
// 4.7 (Befund 3) hat niemand die beiden verglichen. Ein Vanguard-Deck ließ
// sich in ein Magic-Turnier anmelden, und der Abzug stand daneben, als gehöre
// er dazu.
func TestEinDeckAusEinemAnderenSpielLaesstSichNichtAnmelden(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)
	turnier := werk.turnierAnlegen(t)

	werk.decks.deck = domain.Deck{
		ID: 5, Name: "Fremdling", Spiel: "vanguard",
	}

	_, err := werk.leitung.Anmelden(t.Context(), turnier.ID,
		domain.Teilnehmer{Name: "Ada", DeckID: 5})

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler ist %v, erwartet eine Validierung", err)
	}

	// **Und die Meldung nennt beide Spiele** — sonst rät, wer sie liest.
	if !strings.Contains(err.Error(), "vanguard") ||
		!strings.Contains(err.Error(), turnier.Spiel) {
		t.Errorf("Meldung %q nennt nicht beide Spiele", err.Error())
	}
}
