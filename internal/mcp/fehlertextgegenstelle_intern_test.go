package mcp

// **Ein fremder Dienst, der nicht antwortet, sagt das auch dem Modell** — D02
// (#2233).
// Bis dahin bekam es „der Aufruf ist gescheitert" und konnte nicht wissen,
// dass ein zweiter Versuch später der richtige Schritt ist.

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinFremderDienstBehaeltSeinenSatz(t *testing.T) {
	t.Parallel()

	fehler := fmt.Errorf("Scryfall fragen: %w", domain.NeuerGegenstellenfehler(
		domain.ErrKeineAntwort, errors.New("net/http: timeout awaiting response headers"),
		"api.scryfall.com", 30*time.Second))

	text := fehlertext(fehler)
	if text == Gescheitert || !strings.Contains(text, "api.scryfall.com") || strings.Contains(text, "net/http") {
		t.Errorf("das Gegenüber liest %q", text)
	}
}
