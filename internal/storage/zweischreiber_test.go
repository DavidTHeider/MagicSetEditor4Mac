package storage_test

import (
	"database/sql"
	"path/filepath"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Zwei Aufrufer, ein Bestand** — die Lage, die der MCP-Server herstellt.
//
// # Warum das hier steht und nicht in `internal/mcp`
//
// Es ist eine Eigenschaft des Bestands, keine des Protokolls: seit
// [0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md) schreiben
// **zwei Prozesse** auf dieselbe SQLite-Datei — das Fenster und der Server.
// Bisher schrieb genau eines.
//
// # Warum eine Datei und kein Arbeitsspeicher
//
// `ImArbeitsspeicher` gibt je Verbindung eine eigene Datenbank; zwei Verbindungen
// darauf sähen einander nie. Die Lage, um die es geht, entsteht erst an einer
// **Datei** — mit WAL, mit `busy_timeout` und mit allem, was SQLite dazu sagt.
//
// Läuft unter `-race` wie die ganze Stufe `test` (`scripts/ci.sh`): #143
// verlangt, dass nebenläufiges Schreiben **geprüft** ist, nicht behauptet.
func TestZweiSchreiberAufEinerDatei(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bestand.db")

	// Zwei **getrennte** Verbindungen — wie zwei Prozesse. Ein geteiltes
	// `*sql.DB` wäre die Lage, die es schon gab: ein Verbindungsvorrat, der
	// selbst reihum lässt.
	erste, zweite := verbindung(t, ort), verbindung(t, ort)

	if err := storage.Migriere(probekontext(t), erste, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	speicher := storage.NeuerSetSpeicher(erste)

	setID, err := speicher.SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "magic", Origin: ort,
		Cards: []*domain.Card{
			{Values: []domain.Value{{Field: "name", Text: "Eine"}}},
			{Values: []domain.Value{{Field: "name", Text: "Zwei"}}},
		},
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	karten, err := storage.NeuerSetLeser(erste).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(karten) != 2 {
		t.Fatalf("%d Karten", len(karten))
	}

	// **Jeder Schreiber seine eigene Karte.** Zwei Schreiber auf **derselben**
	// Karte wären ein Wettlauf um den letzten Wert — eine andere Frage, und
	// eine, die §7 Regel 4 beantwortet.
	schreiber := []*storage.Kartenschreiber{
		storage.NeuerKartenschreiber(erste),
		storage.NeuerKartenschreiber(zweite),
	}

	var (
		warten  sync.WaitGroup
		schloss sync.Mutex
		fehler  []error
	)

	for lauf := range 20 {
		warten.Add(1)

		go func() {
			defer warten.Done()

			fehlschlag := schreiber[lauf%2].FeldSetzen(
				probekontext(t), karten[lauf%2].ID, "rule text", "Lauf "+string(rune('a'+lauf)))
			if fehlschlag != nil {
				schloss.Lock()
				fehler = append(fehler, fehlschlag)
				schloss.Unlock()
			}
		}()
	}

	warten.Wait()

	// **Kein `SQLITE_BUSY` ist die Zusage**, und sie hängt an den Pragmas:
	// WAL erlaubt Leser neben einem Schreiber, `busy_timeout` lässt den zweiten
	// warten statt scheitern.
	if len(fehler) > 0 {
		t.Errorf("%d von 20 Schreibvorgängen scheiterten, erster: %v", len(fehler), fehler[0])
	}

	// Und beide Karten tragen danach einen Wert — geschrieben ist geschrieben.
	nachher, err := storage.NeuerSetLeser(zweite).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	for nummer, karte := range nachher {
		if wert, da := karte.Value("rule text"); !da || wert == "" {
			t.Errorf("Karte %d trägt keinen Regeltext", nummer+1)
		}
	}
}

// verbindung öffnet eine eigene Verbindung auf dieselbe Datei.
func verbindung(t *testing.T, ort string) *sql.DB {
	t.Helper()

	db, err := storage.Oeffnen(probekontext(t), ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := db.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	return db
}
