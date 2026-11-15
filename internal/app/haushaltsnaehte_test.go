package app_test

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Guthaben und Sperre für die Wurzelknoten** — Werkstück O5 von Phase 5.2 (#1429).
func TestDieWurzelknotenBekommenGuthabenUndSperreMitDenVorgabegrenzen(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	uhr := func() time.Time { return jetzt }

	kontingente := app.NeueKontingente(domain.Vorgabegrenzen(), uhr)
	for range domain.Vorgabegrenzen().Principal.Burst {
		if ok, _ := kontingente.Abbuchen("a", domain.KostenLesen, false); !ok {
			t.Fatal("Guthaben zu früh leer")
		}
	}

	if ok, _ := kontingente.Abbuchen("a", domain.KostenLesen, false); ok {
		t.Error("über dem Burst durchgelassen")
	}

	sperre := app.NeueZugangssperre(uhr)
	if gesperrt, _ := sperre.Gesperrt("x"); gesperrt {
		t.Error("frisch gesperrt")
	}
}
