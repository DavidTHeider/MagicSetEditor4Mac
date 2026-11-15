package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinSchrittMeldetDemMelderDesKontexts(t *testing.T) {
	t.Parallel()

	var gemeldet []domain.Fortschritt

	ctx := service.MitFortschritt(context.Background(), func(stand domain.Fortschritt) {
		gemeldet = append(gemeldet, stand)
	})

	if fehler := service.Schritt(ctx, domain.Fortschritt{Erledigt: 1, Gesamt: 3, Einheit: domain.EinheitKarte}); fehler != nil {
		t.Fatalf("Schritt: %v", fehler)
	}

	if len(gemeldet) != 1 || gemeldet[0].Erledigt != 1 || gemeldet[0].Gesamt != 3 ||
		gemeldet[0].Einheit != domain.EinheitKarte {
		t.Errorf("gemeldet: %+v", gemeldet)
	}
}

func TestEinSchrittOhneMelderTutNichts(t *testing.T) {
	t.Parallel()

	if fehler := service.Schritt(context.Background(), domain.Fortschritt{Erledigt: 1}); fehler != nil {
		t.Errorf("ohne Melder: %v", fehler)
	}
}

// **Ein abgebrochener Kontext hält den Lauf an** — und meldet nichts mehr:
// die Spezifikation von MCP verlangt, dass nach dem Ende keine Meldung folgt.
func TestEinSchrittNachDemAbbruchHaeltAn(t *testing.T) {
	t.Parallel()

	gemeldet := 0

	ctx, abbrechen := context.WithCancel(service.MitFortschritt(context.Background(),
		func(domain.Fortschritt) { gemeldet++ }))
	abbrechen()

	fehler := service.Schritt(ctx, domain.Fortschritt{Erledigt: 2, Gesamt: 3})
	if !errors.Is(fehler, context.Canceled) {
		t.Errorf("Fehler: %v", fehler)
	}

	if gemeldet != 0 {
		t.Errorf("nach dem Abbruch gemeldet: %d", gemeldet)
	}
}

// **Der Melder des inneren Kontexts gilt** — ein Werkzeug, das einen Lauf in
// einem anderen fortsetzt, meldet an den, der zuletzt zuhört.
func TestDerInnereMelderGilt(t *testing.T) {
	t.Parallel()

	aussen, innen := 0, 0

	ctx := service.MitFortschritt(context.Background(), func(domain.Fortschritt) { aussen++ })
	ctx = service.MitFortschritt(ctx, func(domain.Fortschritt) { innen++ })

	if fehler := service.Schritt(ctx, domain.Fortschritt{Erledigt: 1}); fehler != nil {
		t.Fatal(fehler)
	}

	if aussen != 0 || innen != 1 {
		t.Errorf("außen %d, innen %d", aussen, innen)
	}
}

// **Ein Abbruch während der Meldung greift sofort** — sonst ließe sich ein
// Lauf mit einem einzigen Schritt nicht mehr anhalten.
func TestEinAbbruchWaehrendDerMeldungGreiftSofort(t *testing.T) {
	t.Parallel()

	ctx, abbrechen := context.WithCancel(context.Background())
	ctx = service.MitFortschritt(ctx, func(domain.Fortschritt) { abbrechen() })

	if fehler := service.Schritt(ctx, domain.Fortschritt{}); !errors.Is(fehler, context.Canceled) {
		t.Errorf("Fehler: %v", fehler)
	}
}
