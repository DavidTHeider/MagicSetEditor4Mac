package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Fehler nennt die Karten**, sonst probiert jemand zwanzig einzeln durch.
func TestBulkErrorNenntDieKarten(t *testing.T) {
	fehler := &domain.BulkError{
		CardIDs: []int64{4711, 4712},
		Err:     domain.ErrNotFound,
	}

	if !strings.Contains(fehler.Error(), "4711") || !strings.Contains(fehler.Error(), "4712") {
		t.Errorf("%q", fehler.Error())
	}

	// Und er bleibt ein `ErrNotFound` — geprüft mit errors.Is, nicht am Text.
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("%v ist kein ErrNotFound", fehler)
	}

	var gefunden *domain.BulkError
	if !errors.As(error(fehler), &gefunden) || len(gefunden.CardIDs) != 2 {
		t.Errorf("%v", gefunden)
	}
}

func TestCardRowValueUnterscheidetLeerVonFehlend(t *testing.T) {
	karte := &domain.CardRow{Values: []domain.Value{{Field: "name", Text: ""}}}

	if wert, da := karte.Value("name"); !da || wert != "" {
		t.Errorf("%q / %v — ein geleertes Feld ist da", wert, da)
	}

	if _, da := karte.Value("typ"); da {
		t.Error("ein Feld, das es nicht gibt, wurde gefunden")
	}
}
