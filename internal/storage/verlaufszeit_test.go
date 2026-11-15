package storage_test

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Der Verlauf sagt, wann etwas geschah** — Werkstück U6 von Phase 5.1
// (#1417). Die Spalte stand seit Phase 2 in der Tafel und erreichte den
// Eintrag nie; eine Liste, die nicht sagt, wann, ist ein Protokoll ohne Uhr.
func TestDerVerlaufseintragTraegtSeinenZeitpunkt(t *testing.T) {
	_, db := suchbestand(t)

	vorher := time.Now().UTC().Add(-time.Second)

	if _, _, err := storage.NeuerKartenschreiber(db).FeldSetzenMitVorwert(
		probekontext(t), 1, "typ", "Artefakt"); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	verlauf, err := storage.NeuerVerlaufsleser(db).Verlauf(probekontext(t), 1)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(verlauf) == 0 {
		t.Fatal("kein Eintrag")
	}

	juengster := verlauf[len(verlauf)-1]

	zeit, fehler := time.Parse(time.RFC3339Nano, juengster.Zeitpunkt)
	if fehler != nil {
		t.Fatalf("Zeitpunkt %q: %v", juengster.Zeitpunkt, fehler)
	}

	if zeit.Before(vorher) || zeit.After(time.Now().UTC().Add(time.Second)) {
		t.Errorf("Zeitpunkt %s liegt nicht bei jetzt", zeit)
	}
}
