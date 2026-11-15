package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine Massenänderung läuft in einer Transaktion.**
func TestFelderSetzenAendertAlleAufEinmal(t *testing.T) {
	leser, db := suchbestand(t)
	schreiber := storage.NeuerKartenschreiber(db)

	if _, err := schreiber.FelderSetzen(probekontext(t), []int64{1, 2}, "typ", "Verzauberung"); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	for _, karte := range []int64{1, 2} {
		gelesen, err := leser.KarteLesen(probekontext(t), karte)
		if err != nil {
			t.Fatalf("lesen: %v", err)
		}

		if wert, da := gelesen.Value("typ"); !da || wert != "Verzauberung" {
			t.Errorf("Karte %d: %q / %v", karte, wert, da)
		}
	}
}

// Ein Feld, das eine Karte noch nicht hat, kommt hinzu.
func TestFelderSetzenLegtFehlendeFelderAn(t *testing.T) {
	leser, db := suchbestand(t)

	if _, err := storage.NeuerKartenschreiber(db).
		FelderSetzen(probekontext(t), []int64{2}, "text", "Neu"); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	gelesen, err := leser.KarteLesen(probekontext(t), 2)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if wert, da := gelesen.Value("text"); !da || wert != "Neu" {
		t.Errorf("%q / %v", wert, da)
	}
}

// **Eine gescheiterte Änderung sagt, welche Karten betroffen sind** — und
// ändert nichts. Alles oder nichts: sonst stünde die Hälfte eines Vorgangs im
// Bestand, und niemand wüsste welche.
func TestFelderSetzenNenntDieFehlendenKarten(t *testing.T) {
	leser, db := suchbestand(t)
	schreiber := storage.NeuerKartenschreiber(db)

	_, err := schreiber.FelderSetzen(probekontext(t), []int64{1, 4711, 2, 4712}, "typ", "Artefakt")
	if err == nil {
		t.Fatal("die Änderung ging ohne Fehler durch")
	}

	var massenfehler *domain.BulkError
	if !errors.As(err, &massenfehler) {
		t.Fatalf("Fehler %v — erwartet war ein BulkError", err)
	}

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v — erwartet war ErrNotFound", err)
	}

	// **Alle fehlenden**, nicht nur die erste: wer eine Auswahl von zwanzig
	// Karten anstößt, will nicht zwanzigmal nacheinander eine erfahren.
	if len(massenfehler.CardIDs) != 2 ||
		massenfehler.CardIDs[0] != 4711 || massenfehler.CardIDs[1] != 4712 {
		t.Errorf("%v", massenfehler.CardIDs)
	}

	// Und die vorhandenen Karten sind unverändert.
	gelesen, err := leser.KarteLesen(probekontext(t), 1)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if wert, _ := gelesen.Value("typ"); wert != "Zauber" {
		t.Errorf("Karte 1 wurde geändert: %q", wert)
	}
}

// Eine leere Auswahl ist kein Fehler, sondern nichts zu tun.
func TestFelderSetzenOhneKartenTutNichts(t *testing.T) {
	_, db := suchbestand(t)

	if _, err := storage.NeuerKartenschreiber(db).
		FelderSetzen(probekontext(t), nil, "typ", "Artefakt"); err != nil {
		t.Errorf("leere Auswahl: %v", err)
	}
}

func TestFelderSetzenMeldetEineGeschlosseneDatenbank(t *testing.T) {
	_, db := suchbestand(t)

	if err := db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	if _, err := storage.NeuerKartenschreiber(db).
		FelderSetzen(probekontext(t), []int64{1}, "typ", "Artefakt"); err == nil {
		t.Error("die Änderung kam ohne Fehler zurück")
	}
}
