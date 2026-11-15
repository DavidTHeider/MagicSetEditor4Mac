package domain_test

// **Die Bereitschaft ist gerechnet, nicht behauptet** — Werkstück A1 von
// Phase 5.3 (#1440).
//
// Eine Fähigkeit ist erst »bereit«, wenn nichts dagegen steht — und wenn
// mehreres dagegen steht, nennt der Status das, was zuerst zu beheben ist:
// ein abgeschalteter Dienst braucht keinen Schlüssel, ein fehlender
// Schlüssel keinen Riegel.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerStatusNenntDasErsteHindernis(t *testing.T) {
	t.Parallel()

	stoerung := errors.New("der Schlüsselbund antwortet nicht")
	sperre := errors.New("der Leser darf nicht nach draußen")

	faelle := []struct {
		name     string
		lage     domain.Dienstlage
		erwartet string
	}{
		{"an, ohne Wahl, nichts dagegen", domain.Dienstlage{An: true}, domain.StatusBereit},
		{"an und gewählt", domain.Dienstlage{An: true, HatWahl: true, Gewaehlt: true}, domain.StatusBereit},
		{"aus schlägt alles", domain.Dienstlage{
			FehlendeAngabe: "marke", Stoerung: stoerung, Sperre: sperre,
		}, domain.StatusAus},
		{"der Zugang fehlt", domain.Dienstlage{An: true, FehlendeAngabe: "marke", Sperre: sperre}, domain.StatusZugangFehlt},
		{"der Schlüsselbund ist gestört", domain.Dienstlage{An: true, Stoerung: stoerung, Sperre: sperre}, domain.StatusGestoert},
		{"der Riegel sperrt", domain.Dienstlage{An: true, Sperre: sperre, HatWahl: true}, domain.StatusGesperrt},
		{"ein anderer ist gewählt", domain.Dienstlage{An: true, HatWahl: true}, domain.StatusNichtGewaehlt},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if status := fall.lage.Status(); status != fall.erwartet {
				t.Errorf("der Status ist %q, erwartet %q", status, fall.erwartet)
			}
		})
	}
}

// **Der Grund gehört zum Status**: wer »gestört« liest, will wissen, was.
func TestDerGrundNenntDasHindernis(t *testing.T) {
	t.Parallel()

	stoerung := errors.New("der Schlüsselbund antwortet nicht")

	faelle := []struct {
		name string
		lage domain.Dienstlage
		will string
	}{
		{"bereit hat keinen Grund", domain.Dienstlage{An: true}, ""},
		{"aus nennt den Schalter", domain.Dienstlage{}, "abgeschaltet"},
		{"fehlender Zugang nennt die Angabe", domain.Dienstlage{An: true, FehlendeAngabe: "team"}, "team"},
		{"eine Störung nennt sich", domain.Dienstlage{An: true, Stoerung: stoerung}, stoerung.Error()},
		{"eine Sperre nennt sich", domain.Dienstlage{An: true, Sperre: errors.New("gesperrt")}, "gesperrt"},
		{"nicht gewählt sagt es", domain.Dienstlage{An: true, HatWahl: true}, "gewählt"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			grund := fall.lage.Grund()
			if fall.will == "" && grund != "" {
				t.Errorf("der Grund ist %q, erwartet keiner", grund)
			}

			if fall.will != "" && !strings.Contains(grund, fall.will) {
				t.Errorf("der Grund %q nennt %q nicht", grund, fall.will)
			}
		})
	}
}
