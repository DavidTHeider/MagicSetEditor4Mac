package domain_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Paket ohne Art oder Name ist keines.**
//
// Der Katalog führt beide als Schlüssel (`UNIQUE (art, name)`); ein Eintrag
// ohne sie ließe sich weder wiederfinden noch entfernen.
func TestPaketValidateBefunde(t *testing.T) {
	faelle := map[string]struct {
		paket domain.Paket
		feld  string
	}{
		"ohne Art":      {paket: domain.Paket{Name: "magic", Pfad: "/p"}, feld: "art"},
		"ohne Name":     {paket: domain.Paket{Art: "mse-game", Pfad: "/p"}, feld: "name"},
		"nur Leerraum":  {paket: domain.Paket{Art: "mse-game", Name: "  ", Pfad: "/p"}, feld: "name"},
		"ohne Pfad":     {paket: domain.Paket{Art: "mse-game", Name: "magic"}, feld: "pfad"},
		"Art unbekannt": {paket: domain.Paket{Art: "mse-erfunden", Name: "x", Pfad: "/p"}, feld: "art"},
		"Bedarf ohne Namen": {paket: domain.Paket{
			Art: "mse-style", Name: "x", Pfad: "/p",
			Braucht: []domain.Bedarf{{Version: "2026-01-01"}},
		}, feld: "braucht"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			err := fall.paket.Validate()
			if err == nil {
				t.Fatal("kein Befund")
			}

			var feldfehler *domain.FieldError
			if !errors.As(err, &feldfehler) {
				t.Fatalf("kein Feldfehler: %v", err)
			}

			if feldfehler.Field != fall.feld {
				t.Errorf("Feld %q, erwartet %q", feldfehler.Field, fall.feld)
			}
		})
	}
}

// **Eine leere Version und eine leere Herkunft bleiben erlaubt**, und das ist
// keine Nachlässigkeit: `docs/FORMATE.md` führt Pakete ohne `version:`, und ein
// Paket, das jemand von Hand ins Verzeichnis gelegt hat, kommt aus keiner
// benennbaren Quelle. Wer beides erzwänge, machte ein vorhandenes Paket
// unbenutzbar — auch eine Ablehnung braucht Deckung (§10).
func TestEinPaketOhneVersionUndHerkunftGehtDurch(t *testing.T) {
	paket := domain.Paket{Art: "mse-game", Name: "probe", Pfad: "/pakete/probe.mse-game"}

	if err := paket.Validate(); err != nil {
		t.Errorf("abgewiesen: %v", err)
	}
}

// **Eine Stiloption ohne Namen ist keine, und eine Auswahl ohne Möglichkeiten
// auch nicht.**
func TestStyleFieldValidateBefunde(t *testing.T) {
	faelle := map[string]struct {
		feld     domain.StyleField
		erwartet string
	}{
		"ohne Namen": {feld: domain.StyleField{Kind: domain.KindBoolean}, erwartet: "name"},
		"Auswahl ohne Möglichkeiten": {
			feld: domain.StyleField{Name: "rand", Kind: domain.KindChoice}, erwartet: "choices",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			err := fall.feld.Validate()
			if err == nil {
				t.Fatal("kein Befund")
			}

			var feldfehler *domain.FieldError
			if !errors.As(err, &feldfehler) || feldfehler.Field != fall.erwartet {
				t.Errorf("Befund: %v", err)
			}
		})
	}

	gut := domain.StyleField{Name: "rand", Kind: domain.KindChoice, Choices: []string{"hell"}}
	if err := gut.Validate(); err != nil {
		t.Errorf("abgewiesen: %v", err)
	}
}
