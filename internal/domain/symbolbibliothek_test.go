package domain_test

// **Ein Symbol der Bibliothek** — DOM-02.AK-01
// (#1697).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinSymbolnameWirdGeprueft(t *testing.T) {
	t.Parallel()

	if name, err := domain.PruefeSymbolname("  Krone  "); err != nil || name != "Krone" {
		t.Errorf("ein gewöhnlicher Name: %q, %v", name, err)
	}

	for name, roh := range map[string]string{
		"leer":          "   ",
		"zu lang":       strings.Repeat("ä", domain.Symbolnamengrenze+1),
		"Zeilenumbruch": "Kro\nne",
	} {
		var feldfehler *domain.FieldError
		if _, err := domain.PruefeSymbolname(roh); !errors.Is(err, domain.ErrValidation) ||
			!errors.As(err, &feldfehler) || feldfehler.Field != "name" {
			t.Errorf("%s: %v", name, err)
		}
	}

	if _, err := domain.PruefeSymbolname(strings.Repeat("ä", domain.Symbolnamengrenze)); err != nil {
		t.Errorf("genau an der Grenze: %v", err)
	}
}

// **Ein Symbol trägt Marken wie jedes Fachobjekt** — seine Kennung ist eine
// Zahl.
func TestEinSymbolTraegtMarkenUnterSeinerZahl(t *testing.T) {
	t.Parallel()

	if kennung, err := domain.MarkenartSymbol.Kennung("007"); err != nil || kennung != "7" {
		t.Errorf("%q, %v", kennung, err)
	}
}
