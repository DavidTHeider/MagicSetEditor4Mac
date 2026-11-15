package storage_test

// **Zwei Prozesse ändern dieselbe Übersetzung, und keiner verliert etwas** —
// LOC-02 (#1688).
//
// Die Sätze stehen als ein JSON-Objekt in einer Zeile. Wer liest, ändert und
// zurückschreibt, überschreibt, was ein zweiter inzwischen geschrieben hat —
// das Fenster und der MCP-Server sind zwei Prozesse über derselben Datei.
// **Deshalb geschieht es in einem Zug** (`_txlock=immediate`).

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestZweiProzesseAendernDieselbeUebersetzungOhneVerlust(t *testing.T) {
	t.Parallel()

	ctx := probekontext(t)
	ort := filepath.Join(t.TempDir(), "bestand.db")
	ablagen := make([]*storage.Sprachpaketablage, 2)

	for i := range ablagen {
		db, fehler := storage.Oeffnen(ctx, ort)
		if fehler != nil {
			t.Fatalf("öffnen: %v", fehler)
		}

		t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // am Ende eines Laufs

		if fehler := storage.Migriere(ctx, db, nil); fehler != nil {
			t.Fatalf("migrieren: %v", fehler)
		}

		ablagen[i] = storage.NeueSprachpaketablage(db)
	}

	if err := ablagen[0].SprachpaketAnlegen(ctx, sprachpaket("da")); err != nil {
		t.Fatal(err)
	}

	const jeProzess = 12

	var fertig sync.WaitGroup

	fehler := make(chan error, 2*jeProzess)

	for prozess, ablage := range ablagen {
		for satz := range jeProzess {
			fertig.Go(func() {
				pfad := fmt.Sprintf("p%d.s%d", prozess, satz)

				fehler <- ablage.SprachpaketAendern(ctx, "da", func(paket *domain.Sprachpaket) error {
					paket.Texte[pfad] = domain.Uebersetzungstext{Text: pfad, Quelle: "q"}

					return nil
				})
			})
		}
	}

	fertig.Wait()
	close(fehler)

	for err := range fehler {
		if err != nil {
			t.Errorf("ändern: %v", err)
		}
	}

	paket, err := ablagen[1].Sprachpaket(ctx, "da")
	if err != nil || len(paket.Texte) != 1+2*jeProzess {
		t.Errorf("%d Sätze statt %d: %v", len(paket.Texte), 1+2*jeProzess, err)
	}
}

// **Was die Änderung ablehnt, wird nicht geschrieben** — und ein fremdes
// Paket gibt es nicht.
func TestEineAbgelehnteAenderungSchreibtNichts(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueSprachpaketablage(frischeDatenbank(t))
	ctx := probekontext(t)

	if err := ablage.SprachpaketAnlegen(ctx, sprachpaket("da")); err != nil {
		t.Fatal(err)
	}

	nein := errors.New("nein")

	if err := ablage.SprachpaketAendern(ctx, "da", func(paket *domain.Sprachpaket) error {
		paket.Name = "Fort"

		return nein
	}); !errors.Is(err, nein) {
		t.Errorf("abgelehnt: %v", err)
	}

	if paket, _ := ablage.Sprachpaket(ctx, "da"); paket.Name != "Português" { //nolint:errcheck // der Name sagt es
		t.Errorf("geschrieben wurde %q", paket.Name)
	}

	if err := ablage.SprachpaketAendern(ctx, "xx", func(*domain.Sprachpaket) error { return nil }); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein fremdes: %v", err)
	}
}
