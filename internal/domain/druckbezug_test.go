package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Bezug auf einen Druck an den fünf Nähten** — Werkstück A2 von
// Phase 5.3 (#1441): ein Exemplar, ein Verweis und ein Deckplatz dürfen einen
// Druck ohne Scryfall-Alias nennen, und dann sind sie **vorläufig**.

func TestEinExemplarDarfEinenDruckOhneAliasNennen(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name       string
		exemplar   domain.Exemplar
		befund     string
		vorlaeufig bool
	}{
		{
			name:     "mit Alias ist nichts vorläufig",
			exemplar: domain.Exemplar{Printing: "8a1b0d5e-1e0f-4c0f-8b39-3d3a2f9a1c2d", Menge: 1},
		},
		{
			name:       "ein Druck ohne Alias ist vorläufig",
			exemplar:   domain.Exemplar{DruckID: 7, Menge: 1},
			vorlaeufig: true,
		},
		{
			name:     "weder Alias noch Druck ist kein Exemplar",
			exemplar: domain.Exemplar{Menge: 1},
			befund:   "printing",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fall.exemplar.Normalisieren()
			befunde := fall.exemplar.Validate()

			pruefeBefund(t, befunde, fall.befund)

			if fall.exemplar.Vorlaeufig() != fall.vorlaeufig {
				t.Errorf("Vorlaeufig = %v, erwartet %v", fall.exemplar.Vorlaeufig(), fall.vorlaeufig)
			}
		})
	}
}

func TestEinVerweisDarfEinenDruckOhneAliasNennen(t *testing.T) {
	t.Parallel()

	ohne := domain.Kartenverweis{DruckID: 3, Name: "Lightning Bolt"}
	ohne.Normalisieren()

	pruefeBefund(t, ohne.Validate(), "")

	if !ohne.Vorlaeufig() {
		t.Error("ein Verweis ohne Alias ist vorläufig")
	}

	mit := domain.Kartenverweis{Printing: "8a1b0d5e-1e0f-4c0f-8b39-3d3a2f9a1c2d", Name: "Lightning Bolt"}
	if mit.Vorlaeufig() {
		t.Error("ein Verweis mit Alias ist nicht vorläufig")
	}

	keiner := domain.Kartenverweis{Name: "Lightning Bolt"}
	pruefeBefund(t, keiner.Validate(), "printing")
}

func TestEinDeckplatzDarfEinenDruckOhneAliasNennen(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name       string
		platz      domain.Deckplatz
		befund     string
		vorlaeufig bool
	}{
		{
			name:       "ein Druck ohne Alias ist ein Platz und vorläufig",
			platz:      domain.Deckplatz{DruckID: 9, Menge: 1, Bereich: domain.BereichHauptdeck},
			vorlaeufig: true,
		},
		{
			name:   "eigene Karte und Druck zugleich geht nicht",
			platz:  domain.Deckplatz{KarteID: 1, DruckID: 9, Menge: 1, Bereich: domain.BereichHauptdeck},
			befund: "printing",
		},
		{
			name:  "eine eigene Karte ist nicht vorläufig",
			platz: domain.Deckplatz{KarteID: 1, Menge: 1, Bereich: domain.BereichHauptdeck},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			pruefeBefund(t, fall.platz.Validate(), fall.befund)

			if fall.platz.Vorlaeufig() != fall.vorlaeufig {
				t.Errorf("Vorlaeufig = %v, erwartet %v", fall.platz.Vorlaeufig(), fall.vorlaeufig)
			}
		})
	}
}

func TestDerDruckschluesselNenntAliasOderZeile(t *testing.T) {
	t.Parallel()

	if got := (domain.Exemplar{DruckID: 4}).Druckschluessel(); got != "druck:4" {
		t.Errorf("Exemplar.Druckschluessel = %q", got)
	}

	faelle := []struct {
		name     string
		druckID  int64
		printing string
		erwartet string
	}{
		{"der Alias geht vor", 4, "abc", "abc"},
		{"ohne Alias die Zeile", 4, "", "druck:4"},
		{"ohne beides nichts", 0, "", ""},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if got := domain.Druckschluessel(fall.druckID, fall.printing); got != fall.erwartet {
				t.Errorf("Druckschluessel = %q, erwartet %q", got, fall.erwartet)
			}
		})
	}
}

// pruefeBefund erwartet genau den genannten Befund — oder keinen.
func pruefeBefund(t *testing.T, befunde []*domain.FieldError, feld string) {
	t.Helper()

	for _, befund := range befunde {
		if befund.Field == feld {
			return
		}
	}

	if feld == "" && len(befunde) == 0 {
		return
	}

	t.Errorf("Befunde %v, erwartet Feld %q", befunde, feld)
}
