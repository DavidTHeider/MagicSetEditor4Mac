package storage_test

// **Das Profil eines Aufrufs** — Werkstück O1 von Phase 5.2
// (#1425).

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDasArbeitsprofilKommtAusDemKontextSonstVomZeiger(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	// **Der Kontext gewinnt** — er sagt, was ein Token oder ein Wechsel
	// gesagt haben, auch wenn der Zeiger woanders steht.
	ausKontext, fehler := ablage.Arbeitsprofil(storage.MitProfil(context.Background(), 42))
	if fehler != nil || ausKontext != 42 {
		t.Errorf("aus dem Kontext: %d, %v — erwartet 42", ausKontext, fehler)
	}

	// **Ohne Kontext der Zeiger** — die Wanderung zeigt auf das erste Profil.
	zeiger, zeigerfehler := ablage.Profilzeiger(context.Background())
	if zeigerfehler != nil {
		t.Fatalf("Zeiger: %v", zeigerfehler)
	}

	vomZeiger, fehler := ablage.Arbeitsprofil(context.Background())
	if fehler != nil || vomZeiger != zeiger {
		t.Errorf("vom Zeiger: %d, %v — erwartet %d", vomZeiger, fehler, zeiger)
	}
}

// **Und ein kaputter Bestand ohne Kontext ist ein Fehler.**
func TestDasArbeitsprofilMeldetEinenKaputtenBestand(t *testing.T) {
	t.Parallel()

	if _, fehler := storage.NeueNutzerablage(kaputterBestand(t)).Arbeitsprofil(context.Background()); fehler == nil {
		t.Error("das Arbeitsprofil kam aus dem kaputten Bestand")
	}
}
