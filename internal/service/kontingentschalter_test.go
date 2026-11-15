package service_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ohne Kontingente geht jeder Ruf durch** — QA-08.AK-03 (#1739): für eine
// kontrollierte Umgebung, in der niemand einem anderen etwas wegnimmt.
func TestOhneKontingenteGehtJederRufDurch(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.October, 2, 12, 0, 0, 0, time.UTC)
	grenzen := kleineGrenzen()
	grenzen.OhneKontingente = true
	kontingente := service.NeueKontingente(grenzen, func() time.Time { return start })

	// **Weit über Principal, Global und Reserve** — und mit der teuersten
	// Klasse, je Principal und quer über viele.
	for stelle := range 1000 {
		ok, warten := kontingente.Abbuchen(fmt.Sprintf("p%d", stelle%300), domain.KostenMasse, false)
		if !ok || warten != 0 {
			t.Fatalf("Ruf %d: ok=%v warten=%v", stelle, ok, warten)
		}
	}

	// **Und die Tabelle bleibt leer** — wer nichts zählt, merkt sich niemanden.
	if umfang := kontingente.Umfang(); umfang != 0 {
		t.Errorf("Tabelle: %d Einträge", umfang)
	}
}

// **Die Gegenprobe an denselben Grenzen**: mit Kontingenten hält der fünfte
// Ruf an — sonst bewiese der Fall darüber nur, dass die Grenzen weit sind.
func TestMitKontingentenHaeltDerselbeLaufAn(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.October, 2, 12, 0, 0, 0, time.UTC)
	kontingente := service.NeueKontingente(kleineGrenzen(), func() time.Time { return start })

	angehalten := 0

	for range 10 {
		if ok, _ := kontingente.Abbuchen("p", domain.KostenLesen, false); !ok {
			angehalten++
		}
	}

	if angehalten != 6 {
		t.Errorf("angehalten: %d von 10", angehalten)
	}
}
