package domain_test

// **Ein Anbieterkonto** — Werkstück 4 von Phase 4.12
// (#1351).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Konto braucht einen Anbieter, ein Subjekt und einen Nutzer.**
//
// # Warum das Subjekt und nicht die Adresse
//
// **Weil eine Adresse sich ändert und ein Subjekt nicht.** Wer bei GitHub
// seine Mailadresse tauscht, ist derselbe Mensch; wer eine Adresse als
// Schlüssel nähme, machte aus ihm einen zweiten — und der erste hätte alle
// Daten.
func TestEinAnbieterkontoBrauchtSeineDreiAngaben(t *testing.T) {
	t.Parallel()

	faelle := map[string]domain.Anbieterkonto{
		"ohne Nutzer":   {Anbieter: "github", Subjekt: "4711"},
		"ohne Anbieter": {NutzerID: 1, Subjekt: "4711"},
		"ohne Subjekt":  {NutzerID: 1, Anbieter: "github"},
		"leerer Anbieter aus Leerraum": {
			NutzerID: 1, Anbieter: "  ", Subjekt: "4711",
		},
		"leeres Subjekt aus Leerraum": {
			NutzerID: 1, Anbieter: "github", Subjekt: "\t",
		},
	}

	for name, konto := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := konto.Validate(); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Fehler %v, erwartet war ErrValidation", err)
			}
		})
	}
}

// **Ein vollständiges Konto geht durch.**
func TestEinVollstaendigesAnbieterkontoGehtDurch(t *testing.T) {
	t.Parallel()

	konto := domain.Anbieterkonto{NutzerID: 1, Anbieter: "github", Subjekt: "4711"}

	if err := konto.Validate(); err != nil {
		t.Errorf("abgewiesen: %v", err)
	}
}

// **Ein unbekannter Anbieter wird abgewiesen.**
//
// **Und das ist eine Aussage über die Fläche**, nicht über die Schreibweise:
// wer „Github" schreibt, meint denselben Anbieter, und zwei Zeilen für ihn
// wären zwei Menschen. Deshalb ist die Liste geschlossen und die Schreibweise
// festgelegt.
func TestEinUnbekannterAnbieterWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	konto := domain.Anbieterkonto{NutzerID: 1, Anbieter: "Github", Subjekt: "4711"}

	if err := konto.Validate(); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", err)
	}
}

// **Die vier Anbieter stehen fest** — und Apple ist keiner.
//
// Warum, steht in Vermerk 0066: der Weg verlangt einen Developer-Account und
// eine App im App Store. Dieses Projekt hat beides nicht und soll es nicht
// bekommen ([0003](../../docs/entscheidungen/0003-keine-notarisierung.md)).
func TestDieVierAnbieterStehenFest(t *testing.T) {
	t.Parallel()

	if gesehen := len(domain.Anbieter()); gesehen != 4 {
		t.Fatalf("%d Anbieter, erwartet waren vier: %v", gesehen, domain.Anbieter())
	}

	for _, name := range []string{"github", "google", "microsoft", "x"} {
		if !domain.KenntAnbieter(name) {
			t.Errorf("%q fehlt", name)
		}
	}

	if domain.KenntAnbieter("apple") {
		t.Error("Apple steht in der Liste — siehe Vermerk 0066")
	}
}
