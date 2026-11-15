package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinVeralteterProzessIstBelegt(t *testing.T) {
	t.Parallel()

	fehler := &domain.ProzessVeraltet{Fassung: &domain.NeuereFassung{Unbekannt: []string{"0099_kuenftig.sql"}, Stand: "0098_sicherungsvermerke.sql"}}

	if !errors.Is(fehler, domain.ErrBelegt) {
		t.Error("ein veralteter Prozess ist nicht belegt")
	}

	if satz := fehler.Error(); !strings.Contains(satz, "endet") || !strings.Contains(satz, "0099_kuenftig.sql") {
		t.Errorf("der Satz: %s", satz)
	}
}
