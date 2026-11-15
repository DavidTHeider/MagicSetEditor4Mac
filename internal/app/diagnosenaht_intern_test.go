package app

// **Die Naht zum Schemastand** — Befund 2 des Zweitblicks über Phase 3.5.
//
// **Ein interner Lauf**, weil `schemanaht` intern ist: sie ist die eine Stelle,
// an der die Diagnose `internal/storage` erreicht, und sie soll nicht weiter
// sichtbar sein als nötig (§2).

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Über einer migrierten Datenbank nennt sie die verbuchten Migrationen.**
func TestDieSchemanahtNenntDieMigrationen(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "bestand.db")

	db, err := storage.FuerLaeufe("file:" + pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	namen, lesefehler := SchemastandUeber(db).AngewendeteMigrationen(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	// **Mehr als eine, und sortiert.** Die Zahl selbst steht hier nicht: sie
	// wächst mit jeder Migration, und ein Lauf, der sie festschriebe, wäre bei
	// der nächsten rot, ohne dass etwas kaputt wäre.
	if len(namen) < 2 {
		t.Fatalf("es sind %d Migrationen: %v", len(namen), namen)
	}

	for i := 1; i < len(namen); i++ {
		if namen[i-1] >= namen[i] {
			t.Errorf("die Reihenfolge stimmt nicht: %v", namen)

			break
		}
	}
}
