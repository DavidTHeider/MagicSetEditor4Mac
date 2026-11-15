package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Feld auf einen gelesenen Stand schreiben** — Befund 1 des Nachlaufs
// über E/F (#2298).
//
// Die Zusatzwerte einer Karte stehen zusammen in einem Feld `extra data`; wer
// eines setzt, liest den Block, baut ihn neu und schreibt ihn ganz. Ohne
// Bedingung schrieb ein zweiter Schreiber den Block von vorhin zurück, und
// eine Änderung dazwischen ging verloren. Mit dem erwarteten Stand der Karte
// — geprüft in derselben Transaktion — wird er abgewiesen.
func TestEinFeldWirdNurAufDenGelesenenStandGeschrieben(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "extra data", Text: "a"})
	schreiber := storage.NeuerKartenschreiber(db)

	gelesen, err := storage.NeuerSetLeser(db).KarteLesen(probekontext(t), karteID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	stand := gelesen.Stand()

	vorher, hatte, err := schreiber.FeldSetzenAufStand(probekontext(t), karteID, "extra data", "b", stand)
	if err != nil || !hatte || vorher != "a" {
		t.Fatalf("auf dem Stand: vorher %q (da %v), Fehler %v", vorher, hatte, err)
	}

	// Derselbe Stand ein zweites Mal: die Karte steht inzwischen auf »b«.
	_, _, err = schreiber.FeldSetzenAufStand(probekontext(t), karteID, "extra data", "c", stand)
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("auf einem alten Stand: Fehler %v, erwartet ErrConflict", err)
	}

	karte, err := storage.NeuerSetLeser(db).KarteLesen(probekontext(t), karteID)
	if wert, _ := karte.Value("extra data"); err != nil || wert != "b" {
		t.Errorf("im Bestand %q, Fehler %v — der abgewiesene Wert darf nicht stehen", wert, err)
	}
}
