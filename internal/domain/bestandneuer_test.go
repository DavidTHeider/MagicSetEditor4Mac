package domain_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Abweisung des Bestands, den ein Prozess öffnet** — Nachlauf über
// Phase 2 von 0.9.0, Befund 2 (#2323): sie ist dieselbe neuere Fassung wie die
// einer Sicherung, und nur sie heißt für `mse-rest`: ohne Neustart enden.
func TestEinNeuererBestandIstEineNeuereFassung(t *testing.T) {
	t.Parallel()

	fassung := &domain.NeuereFassung{Unbekannt: []string{"0099_kuenftig.sql"}, Stand: "0098_sicherungsvermerke.sql"}
	bestand := &domain.BestandNeuer{Grund: fassung}

	var neuer *domain.NeuereFassung
	if !errors.As(bestand, &neuer) || neuer != fassung {
		t.Errorf("die neuere Fassung ging verloren: %v", bestand)
	}

	if bestand.Error() != fassung.Error() {
		t.Errorf("Satz %q, erwartet %q", bestand.Error(), fassung.Error())
	}

	var nurSicherung *domain.BestandNeuer
	if errors.As(fassung, &nurSicherung) {
		t.Error("eine Sicherung aus einer neueren Fassung gilt als Bestand")
	}
}
