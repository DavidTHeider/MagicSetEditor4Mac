package app_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Schreibnaht des MCP-Servers entsteht hier** —
// #710.
//
// # Warum sie hier geprüft wird und nicht nur dort
//
// `cmd/mse-mcp` prüft den **Weg** — dass ein berechnetes Feld über das
// Protokoll in seinen Teilen landet. Hier steht die Frage davor: dass die Naht
// überhaupt entsteht und dass ihre Fehlerwege benannt sind. Ohne das bliebe
// der Composition Root der Anwendung an der Stelle ungeprüft, an der der
// Befund gefunden wurde.
func TestDieMCPSchreibnahtEntsteht(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	pfad, err := app.Datenbankpfad()
	if err != nil {
		t.Fatalf("Pfad: %v", err)
	}

	if fehler := os.MkdirAll(filepath.Dir(pfad), 0o750); fehler != nil {
		t.Fatalf("Verzeichnis: %v", fehler)
	}

	db, err := storage.Oeffnen(probekontext(t), "file:"+pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := db.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	schreiber, err := app.Schreibnaht(db, domain.UrheberServer)
	if err != nil {
		t.Fatalf("Naht: %v", err)
	}

	// **Der Weg geht bis zum Bestand.** Eine Karte gibt es nicht; die Antwort
	// ist deshalb ein Fehler — und genau der belegt, dass die Naht steht statt
	// stillschweigend nichts zu tun.
	if _, _, fehler := schreiber.FeldSetzenMitVorwert(
		probekontext(t), 1, 999, "name", "Eisball",
	); fehler == nil {
		t.Error("eine Karte, die es nicht gibt, wurde beschrieben")
	}
}

// **Ohne Benutzerverzeichnis wird gemeldet statt geraten.**
func TestDieMCPSchreibnahtMeldetEinenFehlendenOrt(t *testing.T) {
	t.Setenv("HOME", "")

	if _, err := app.Schreibnaht(nil, domain.UrheberServer); err == nil {
		t.Error("kein Fehler")
	}
}

// **Die beiden anderen Nähte für Flächen ohne Fenster** (#886).
//
// Sie werden von `cmd/mse` gerufen und blieben hier sonst ungemessen — genau
// die Sorte Composition-Root-Code, in dem in Phase 5 der Fehler entstand, den
// kein Test sah (#130).
func TestDieNaehteFuerFlaechenOhneFensterStehen(t *testing.T) {
	db := offeneProbedatenbank(t)

	if app.Einfuhrnaht(db) == nil {
		t.Error("die Einfuhrnaht kam nicht zustande")
	}

	schreiber, err := app.Setschreibnaht(db)
	if err != nil {
		t.Fatalf("die Setschreibnaht: %v", err)
	}

	if schreiber == nil {
		t.Error("die Setschreibnaht kam nicht zustande")
	}
}

// offeneProbedatenbank gibt eine migrierte Datenbank in einem eigenen
// Verzeichnis.
func offeneProbedatenbank(t *testing.T) *sql.DB {
	t.Helper()

	db, err := storage.Oeffnen(probekontext(t), filepath.Join(t.TempDir(), "bestand.db"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := db.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	return db
}

// **Ohne Benutzerverzeichnis kommt die Setschreibnaht nicht zustande.**
//
// Sie braucht den Ort der eingesetzten Bilder; ohne ihn schriebe ein Set seine
// Bilder nicht mit, und die Datei wäre eine andere als die des Fensters. Ein
// stiller Weiterlauf wäre genau der Unterschied, der #710 gekostet hat.
func TestOhneBenutzerverzeichnisGibtEsKeineSetschreibnaht(t *testing.T) {
	db := offeneProbedatenbank(t)

	t.Setenv("HOME", "")

	if _, err := app.Setschreibnaht(db); err == nil {
		t.Error("die Naht kam ohne Benutzerverzeichnis zustande")
	}
}
