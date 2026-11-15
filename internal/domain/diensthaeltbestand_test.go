package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinDienstAlsHalterIstEinKonfliktMitGrund(t *testing.T) {
	t.Parallel()

	grund := errors.New("database is locked")
	fehler := &domain.DienstHaeltBestand{Grund: grund}

	if !errors.Is(fehler, domain.ErrConflict) || !errors.Is(fehler, grund) {
		t.Errorf("Klasse und Grund: %v", fehler)
	}

	if satz := fehler.Error(); !strings.Contains(satz, "Dienst") || !strings.Contains(satz, "Einstellungen") {
		t.Errorf("der Satz: %s", satz)
	}
}
