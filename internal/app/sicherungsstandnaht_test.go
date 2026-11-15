package app_test

// **Das Alter über die Naht ohne Fenster** — OPS01
// (#2347):
// dieselbe Antwort, die Kommandozeile, Dienst und Werkzeugserver geben.

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieNahtNenntDasAlterDerSicherung(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	ort := filepath.Join(t.TempDir(), "bestand.db")

	db, fehler := storage.Oeffnen(t.Context(), ort)
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Aufräumen

	if fehler := storage.Migriere(t.Context(), db, nil); fehler != nil {
		t.Fatal(fehler)
	}

	naht := app.BestandsnahtUeber(&db, ort)

	if vorher, err := naht.Sicherungsstand(probekontext(t)); err != nil || vorher.Stufe != "unbekannt" || vorher.Gesichert != "" {
		t.Fatalf("vorher: %+v, %v", vorher, err)
	}

	ziel := filepath.Join(t.TempDir(), "naht"+sicherung.Endung)
	if fehler := naht.Sichern(app.MitSicherungswunsch(t.Context(), domain.FlaecheCLI, "", false), ziel); fehler != nil {
		t.Fatalf("sichern: %v", fehler)
	}

	nachher, err := naht.Sicherungsstand(probekontext(t))
	if err != nil || nachher.Stufe != "frisch" || nachher.Flaeche != "cli" || nachher.Datei != filepath.Base(ziel) {
		t.Errorf("nachher: %+v, %v", nachher, err)
	}
}
