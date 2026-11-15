package service_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Sperre je vorgezeigtem Nachweis** — Werkstück O5 von Phase 5.2 (#1429).

func TestDieZugangssperreTrifftNurDenRatenden(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	sperre := service.NeueZugangssperre(func() time.Time { return jetzt })

	for range domain.SperreFreieVersuche + 1 {
		sperre.Fehlgeschlagen("mse_falsch")
	}

	gesperrt, rest := sperre.Gesperrt("mse_falsch")
	if !gesperrt || rest != domain.SperreGrund {
		t.Errorf("der Ratende: gesperrt=%v rest=%v", gesperrt, rest)
	}

	// **Der Richtige bleibt frei** — die Sperre hängt am Nachweis, nicht am Rechner.
	if gesperrt, _ := sperre.Gesperrt("mse_richtig"); gesperrt {
		t.Error("der gültige Client ist mitgesperrt")
	}

	// Ein Erfolg löscht die Lage.
	sperre.Gelungen("mse_falsch")

	if gesperrt, _ := sperre.Gesperrt("mse_falsch"); gesperrt {
		t.Error("nach Erfolg noch gesperrt")
	}

	// Nach Ablauf ist sie vorbei, ohne dass jemand aufräumt.
	sperre.Fehlgeschlagen("mse_x")
	sperre.Fehlgeschlagen("mse_x")
	sperre.Fehlgeschlagen("mse_x")
	sperre.Fehlgeschlagen("mse_x")
	jetzt = jetzt.Add(2 * time.Second)

	if gesperrt, _ := sperre.Gesperrt("mse_x"); gesperrt {
		t.Error("nach Ablauf gesperrt")
	}
}

func TestZufaelligeNachweiseWachsenNichtUnbegrenzt(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	sperre := service.NeueZugangssperre(func() time.Time { return jetzt })

	for i := range service.HoechstensNachweise * 4 {
		sperre.Fehlgeschlagen(fmt.Sprintf("mse_zufall_%d", i))
	}

	if umfang := sperre.Umfang(); umfang != service.HoechstensNachweise {
		t.Errorf("Tabelle: %d Einträge", umfang)
	}
}
