package domain_test

// **Aus `angeboten` schlägt nur die Gegenseite vor** — die Rolle, die die
// Matrix für `gegenangebot` nennt
// (#1975).
//
// Bis hierher hielt sie nur das Fenster: es bot dem Initiator den Knopf nicht
// an. Kommandozeile, REST, MCP und die Leitung kamen an der Matrix vorbei,
// weil der Vorschlag nicht wusste, wer ihn machte.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieRolleBeimVorschlag(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name    string
		zustand domain.Handelszustand
		wer     string
		geht    bool
	}{
		{"die Gegenseite aus angeboten", domain.HandelAngeboten, "bbert", true},
		{"der Initiator aus angeboten", domain.HandelAngeboten, "banna", false},
		{"der Initiator in der Verhandlung", domain.HandelVerhandlung, "banna", true},
		{"die Gegenseite in der Verhandlung", domain.HandelVerhandlung, "bbert", true},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			heute := einHandelsstand(fall.zustand)

			fehler := heute.NimmtVorschlagVon(einVorschlag(heute, 2), fall.wer)
			if fall.geht && fehler != nil {
				t.Errorf("abgewiesen: %v", fehler)
			}

			var befund *domain.FieldError
			if !fall.geht && (!errors.Is(fehler, domain.ErrValidation) || !errors.As(fehler, &befund) ||
				befund.Key != domain.SchluesselVorschlagNurGegenseite) {
				t.Errorf("angenommen oder falsch abgewiesen: %v", fehler)
			}
		})
	}
}

// **Über die Leitung gilt dieselbe Regel** — der Absender ist, wer spricht.
func TestDieLeitungHaeltDieRolleBeimVorschlag(t *testing.T) {
	t.Parallel()

	heute := einHandelsstand(domain.HandelAngeboten)

	if fehler := heute.NimmtVon(einVorschlag(heute, 2), "banna"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("der Initiator schlug über die Leitung vor: %v", fehler)
	}

	if fehler := heute.NimmtVon(einVorschlag(heute, 2), "bbert"); fehler != nil {
		t.Errorf("die Gegenseite wurde abgewiesen: %v", fehler)
	}
}

// **Ein Gegenangebot nennt nur die Gegenseite oder jede** — die Annahme,
// auf der der Satz von `vorschlagNurGegenseite` steht. Eine Zeile mit dem
// Initiator machte ihn falsch.
func TestEinGegenangebotNenntNurGegenseiteOderJede(t *testing.T) {
	t.Parallel()

	for _, uebergang := range domain.Handelsmatrix() {
		if uebergang.Schritt == domain.SchrittGegenangebot &&
			uebergang.Wer != domain.HandelnderGegenseite && uebergang.Wer != domain.HandelnderJede {
			t.Errorf("aus %q schlägt %q vor", uebergang.Von, uebergang.Wer)
		}
	}
}
