package domain_test

// **Der Vermerk über die eigene Peerkennung** — Werkstück N1 von Phase 5.4
// (#1454).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **`Gibts` unterscheidet „es gab noch keine" von „sie ist da".**
//
// **Und das ist keine Formsache**: an dieser Frage hängt, ob ein fehlender
// Schlüssel ein **Verlust** ist oder ein erster Start.
func TestEinVermerkSagtObEsIhnGibt(t *testing.T) {
	t.Parallel()

	if (domain.Peervermerk{}).Gibts() {
		t.Error("ein leerer Vermerk gibt es angeblich")
	}

	if !(domain.Peervermerk{Kennung: "bwas"}).Gibts() {
		t.Error("ein Vermerk mit Kennung gibt es angeblich nicht")
	}
}
