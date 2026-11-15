package storage_test

// **Ein Exemplar nach seiner Kennung** — Werkstück U9 von Phase 5.1
// (#1420):
// die Zuordnung nennt ein Exemplar, und der Dienst muss es nachschlagen —
// im eigenen Profil, sonst gibt es keins.
//
// **Kein `t.Parallel()`**, aus dem Grund, der in `sammlung_test.go` steht.

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinExemplarLaesstSichNachKennungLesen(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	id, err := speicher.Anlegen(probekontext(t), probeexemplar())
	if err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	gelesen, err := speicher.Einzeln(probekontext(t), id)
	if err != nil {
		t.Fatalf("Einzeln: %v", err)
	}

	if gelesen.ID != id || gelesen.Printing != probedruck || gelesen.Menge != 3 || gelesen.DruckID == 0 {
		t.Errorf("gelesen: %+v", gelesen)
	}

	if _, err := speicher.Einzeln(probekontext(t), id+1); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein fehlendes Exemplar: %v", err)
	}

	if _, err := speicher.Einzeln(context.Background(), id); err == nil {
		t.Error("ohne Profil kam es durch")
	}

	if _, err := storage.NeuerSammlungsspeicher(geschlossen(t)).Einzeln(probekontext(t), id); err == nil {
		t.Error("der geschlossene Bestand kam durch")
	}
}

// geschlossen gibt einen Bestand, der zu ist.
func geschlossen(t *testing.T) *sql.DB {
	t.Helper()

	db, err := storage.FuerLaeufe("file:" + t.TempDir() + "/zu.db")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	return db
}
