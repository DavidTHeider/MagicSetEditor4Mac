package storage_test

// **Der lesbare Bezug eines Exemplars** —
// #1200.
//
// **Eine eigene Datei**, weil `sammlung_test.go` mit ihr über 400 Zeilen ginge
// (§2). Der Schnitt folgt der Frage: dort steht, wie ein Exemplar entsteht und
// vergeht, hier, woran ein Mensch es erkennt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Der lesbare Bezug reist mit** —
// #1200.
//
// **§16, Regel 5: ohne Netz arbeitet die Anwendung weiter.** Wer seine
// Sammlung im Zug ansieht, soll Namen lesen und keine Kennungen.
func TestDerLesbareBezugReistMit(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	id, err := speicher.Anlegen(probekontext(t), &domain.Exemplar{
		Printing: "8a1b0d5e", Menge: 1, Zustand: domain.ZustandNearMint,
		Sprache: "en", Name: "Lightning Bolt", Setcode: "m11", Nummer: "149",
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	gelesen := einExemplar(t, speicher)
	if gelesen.Name != "Lightning Bolt" || gelesen.Setcode != "m11" ||
		gelesen.Nummer != "149" {
		t.Fatalf("gelesen: %+v", gelesen)
	}

	// **Und er lässt sich ändern** — eine falsch erfasste Angabe muss sich
	// berichtigen lassen, ohne die Zeile zu löschen (Befund 1).
	gelesen.Nummer = "150"

	if err := speicher.Aendern(probekontext(t), gelesen); err != nil {
		t.Fatalf("ändern: %v", err)
	}

	if danach := einExemplar(t, speicher); danach.Nummer != "150" {
		t.Errorf("nach dem Ändern: %+v", danach)
	}

	if id == 0 {
		t.Error("keine Kennung")
	}
}

// einExemplar liest das einzige Exemplar der Sammlung.
func einExemplar(t *testing.T, speicher *storage.Sammlungsspeicher) *domain.Exemplar {
	t.Helper()

	liste, err := speicher.Liste(probekontext(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(liste) != 1 {
		t.Fatalf("%d Exemplare", len(liste))
	}

	return liste[0]
}
