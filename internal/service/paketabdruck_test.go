package service_test

// **Lokale Änderungen an einem Paket** — DOM-01.AK-01
// (#1696).
//
// Beim Einbinden hält der Katalog einen **Abdruck** fest; die Liste
// vergleicht ihn mit dem, was heute dasteht. Ein Paket ohne Abdruck — die
// Saat, ein selbst angelegtes — ist nie „geändert": dort gibt es keinen
// Stand, von dem es abweichen könnte.

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinGeaendertesPaketFaelltAuf(t *testing.T) {
	quelle := t.TempDir()
	paket(t, quelle, "fremd.mse-style",
		"mse version: 2.0.0\ngame: probe\nshort name: Fremd\nversion: 2026-02-02\n")

	ziel := t.TempDir()
	saat := t.TempDir()
	paket(t, saat, "probe.mse-game", "mse version: 2.0.0\nshort name: Probe\n")

	verwalter := service.NeuerPaketverwalter(&ablage{}, ziel, ziel, saat)

	if _, err := verwalter.Installieren(t.Context(), filepath.Join(quelle, "fremd.mse-style")); err != nil {
		t.Fatalf("installieren: %v", err)
	}

	liste, err := verwalter.Liste(t.Context())
	if err != nil {
		t.Fatalf("liste: %v", err)
	}

	for _, eines := range liste {
		switch eines.Name {
		case "fremd":
			if eines.Abdruck == "" || eines.Geaendert {
				t.Errorf("frisch eingebunden: Abdruck %q, geändert %v", eines.Abdruck, eines.Geaendert)
			}
		case "probe":
			if eines.Abdruck != "" || eines.Geaendert {
				t.Errorf("die Saat: Abdruck %q, geändert %v", eines.Abdruck, eines.Geaendert)
			}
		}
	}

	// **Eine Datei wird bearbeitet** — länger und später.
	datei := filepath.Join(ziel, "fremd.mse-style", "style")
	if schreibfehler := os.WriteFile(datei, []byte("mse version: 2.0.0\ngame: probe\nshort name: Fremd, bearbeitet\n"), 0o600); schreibfehler != nil {
		t.Fatalf("bearbeiten: %v", schreibfehler)
	}

	spaeter := time.Now().Add(time.Hour)
	if zeitfehler := os.Chtimes(datei, spaeter, spaeter); zeitfehler != nil {
		t.Fatalf("Zeit: %v", zeitfehler)
	}

	liste, err = verwalter.Liste(t.Context())
	if err != nil {
		t.Fatalf("liste: %v", err)
	}

	for _, eines := range liste {
		if eines.Name == "fremd" && !eines.Geaendert {
			t.Error("die Bearbeitung fiel nicht auf")
		}
	}
}

// **Ein Paket, das nicht mehr dasteht, gilt als geändert** — es steht
// jedenfalls nicht mehr so da wie beim Einbinden.
func TestEinVerschwundenesPaketGiltAlsGeaendert(t *testing.T) {
	verwalter := service.NeuerPaketverwalter(&ablage{pakete: []domain.Paket{{
		Art: "mse-style", Name: "weg", Pfad: filepath.Join(t.TempDir(), "weg.mse-style"), Abdruck: "abdruck:alt",
	}}}, "")

	liste, err := verwalter.Liste(t.Context())
	if err != nil || len(liste) != 1 || !liste[0].Geaendert {
		t.Errorf("das verschwundene Paket: %+v, %v", liste, err)
	}
}
