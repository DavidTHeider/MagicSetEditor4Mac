package service_test

// **Vor dem Senden** — Werkstück A16 von Phase 5.3
// (#1515).
//
// **Ein Paket ist eine Momentaufnahme.** Zwischen Zusammenstellen und Senden
// vergeht Zeit, und das Issue zählt auf, was sich dabei ändern kann.

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einGesendetesPaket stellt eines zusammen.
func einGesendetesPaket(t *testing.T) domain.Kontextmanifest {
	t.Helper()

	paket, fehler := einKontextdienst(&recherchequellendoppel{
		stand:   einRegelstand(),
		treffer: []domain.Retrievaltreffer{einRegeltreffer("702.9b", "Ein Text.", 1)},
	}).Zusammenstellen(t.Context(), einKontextwunsch())
	if fehler != nil {
		t.Fatalf("Zusammenstellen: %v", fehler)
	}

	return paket.Manifest
}

// dasZiel ist das, für das budgetiert wurde.
func dasZiel() service.Sendeziel {
	return service.Sendeziel{
		Anbieter: "ollama", Modell: "gemma3:270m",
		Laufzeit: "ollama 0.33.3", Zielrevision: "rev-1",
	}
}

// **Ein frisches Paket an sein Ziel geht durch.**
func TestEinFrischesPaketAnSeinZielGehtDurch(t *testing.T) {
	t.Parallel()

	dienst := einKontextdienst(&recherchequellendoppel{stand: einRegelstand()})

	if fehler := dienst.PruefeVorDemSenden(t.Context(),
		einGesendetesPaket(t), dasZiel()); fehler != nil {
		t.Fatalf("PruefeVorDemSenden: %v", fehler)
	}
}

// **Ein abgelaufenes Paket geht nicht mehr hinaus.**
func TestEinAbgelaufenesPaketGehtNichtMehrHinaus(t *testing.T) {
	t.Parallel()

	manifest := einGesendetesPaket(t)

	spaeter := dieRechercheuhr().Add(domain.Kontextfrist + time.Minute)

	dienst := service.NeuerKontextdienst(
		einRecherchedienst(&recherchequellendoppel{stand: einRegelstand()}, nil),
		profildoppel{}).MitUhr(func() time.Time { return spaeter })

	fehler := dienst.PruefeVorDemSenden(t.Context(), manifest, dasZiel())

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("das abgelaufene Paket kam durch: %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "abgelaufen") {
		t.Errorf("die Meldung ist %q", fehler.Error())
	}
}

// **Ein anderes Ziel bekommt kein altes Paket.**
func TestEinAnderesZielBekommtKeinAltesPaket(t *testing.T) {
	t.Parallel()

	dienst := einKontextdienst(&recherchequellendoppel{stand: einRegelstand()})
	manifest := einGesendetesPaket(t)

	for name, aendern := range map[string]func(*service.Sendeziel){
		"anderer Anbieter": func(z *service.Sendeziel) { z.Anbieter = "anthropic" },
		"anderes Modell":   func(z *service.Sendeziel) { z.Modell = "ein-anderes" },
		"andere Laufzeit":  func(z *service.Sendeziel) { z.Laufzeit = "ollama 0.34.0" },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ziel := dasZiel()
			aendern(&ziel)

			if fehler := dienst.PruefeVorDemSenden(t.Context(), manifest,
				ziel); !errors.Is(fehler, domain.ErrConflict) {
				t.Fatalf("%s kam durch: %v", name, fehler)
			}
		})
	}
}

// **Eine Karte, die sich geändert hat, bekommt kein altes Paket.**
func TestEineGeaenderteZielkarteBekommtKeinAltesPaket(t *testing.T) {
	t.Parallel()

	dienst := einKontextdienst(&recherchequellendoppel{stand: einRegelstand()})

	ziel := dasZiel()
	ziel.Zielrevision = "rev-2"

	if fehler := dienst.PruefeVorDemSenden(t.Context(), einGesendetesPaket(t),
		ziel); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("die geänderte Karte kam durch: %v", fehler)
	}

	// **Ohne genannte Revision wird nicht geprüft** — ein Auftrag ohne
	// Zielkarte hat keine.
	ohne := dasZiel()
	ohne.Zielrevision = ""

	if fehler := dienst.PruefeVorDemSenden(t.Context(), einGesendetesPaket(t),
		ohne); fehler != nil {
		t.Fatalf("ohne Revision: %v", fehler)
	}
}

// **Ein fremdes Profil bekommt kein Paket.**
func TestEinFremdesProfilBekommtKeinPaket(t *testing.T) {
	t.Parallel()

	manifest := einGesendetesPaket(t)

	fremd := service.NeuerKontextdienst(
		einRecherchedienst(&recherchequellendoppel{stand: einRegelstand()}, nil),
		profildoppel{profil: 2}).MitUhr(dieRechercheuhr)

	fehler := fremd.PruefeVorDemSenden(t.Context(), manifest, dasZiel())

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("das fremde Profil kam durch: %v", fehler)
	}

	// **Und ein unlesbares Profil ist ein anderer Fehler.**
	ohne := service.NeuerKontextdienst(
		einRecherchedienst(&recherchequellendoppel{stand: einRegelstand()}, nil),
		profildoppel{fehler: domain.MitGrund(domain.ErrValidation, nil,
			"kein Profil im Kontext")}).MitUhr(dieRechercheuhr)

	if fehler := ohne.PruefeVorDemSenden(t.Context(), manifest,
		dasZiel()); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ohne Profil: %v", fehler)
	}
}
