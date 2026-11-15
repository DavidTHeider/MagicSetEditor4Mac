package service

// **Ein Abzug wird geprüft, bevor er Objekte wird** — Nachlauf zum Zweitblick
// über Phase 7, Rest R1 (S1). Befund 2 hatte die Grenze an `Platznehmen`
// gesetzt; der Partieaufbau legte weiter zuerst und prüfte danach. Die
// Prüfung steht jetzt in der **einen** Funktion, die Objekte erzeugt — kein
// Aufrufer kann sie auslassen.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestBibliothekAusPrueftVorDemLegen(t *testing.T) {
	t.Parallel()

	// **Eine Menge, die niemand legen könnte** — ginge sie in die Schleife,
	// liefe dieser Test über Minuten und Gigabytes.
	riesig := domain.Partiedeck{Abdruck: "a", Karten: []domain.Abzugskarte{{
		Kennung: "p1", Name: "Wald", Menge: 1 << 40,
		Seiten: []domain.Abzugsseite{{Name: "Wald", KarteID: 1}},
	}}}

	gelegt, fehler := bibliothekAus(riesig, "sitz-1")
	if !errors.Is(fehler, domain.ErrValidation) || len(gelegt) != 0 {
		t.Errorf("gelegt %d, Fehler %v", len(gelegt), fehler)
	}
}
