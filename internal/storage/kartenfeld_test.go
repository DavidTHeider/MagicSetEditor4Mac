package storage_test

// **Welchen Wert trägt ein benanntes Feld einer Karte?** —
// #1376.
//
// `NameZurKarte` beantwortete das für genau ein Feld. Die Deckliste braucht
// daneben die Manakosten, und die Frage ist dieselbe — deshalb eine Naht mit
// dem Feldnamen als Argument statt einer zweiten Funktion, die mit der ersten
// auseinanderliefe.

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein gefülltes Feld gibt seinen Wert.**
func TestEinGefuelltesFeldGibtSeinenWert(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := karteMitNamen(t, db, "Lightning Bolt")
	feldSetzen(t, db, karteID, "casting cost", "R")

	wert, err := storage.NeuerSetLeser(db).FeldZurKarte(
		probekontext(t), karteID, "casting cost")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if wert != "R" {
		t.Errorf("der Wert ist %q, erwartet war %q", wert, "R")
	}
}

// **Ein Feld, das die Karte nicht führt, gibt leer — und keinen Fehler.**
//
// Nicht jedes Spiel hat „casting cost". Ein Fehler dafür hieße, dass eine
// Deckliste in jedem anderen Spiel gar nicht käme.
func TestEinUngefuehrtesFeldGibtLeer(t *testing.T) {
	db := frischeDatenbank(t)

	wert, err := storage.NeuerSetLeser(db).FeldZurKarte(
		probekontext(t), karteMitNamen(t, db, "Lightning Bolt"), "casting cost")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if wert != "" {
		t.Errorf("der Wert ist %q, erwartet war leer", wert)
	}
}

// **Eine Karte, die es nicht gibt, führt kein Feld — und das ist ein Nein.**
//
// Der Unterschied zum Fall darüber ist der, auf den es ankommt: leer heißt
// „die Karte hat das Feld nicht", nicht „die Karte gibt es nicht".
func TestEineFehlendeKarteFuehrtKeinFeld(t *testing.T) {
	if _, err := storage.NeuerSetLeser(frischeDatenbank(t)).FeldZurKarte(
		probekontext(t), 999, "casting cost"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}
}

// feldSetzen legt einen Feldwert an einer Karte ab.
func feldSetzen(t *testing.T, db *sql.DB, karteID int64, feld, wert string) {
	t.Helper()

	if _, err := db.ExecContext(probekontext(t),
		`INSERT INTO card_values (card_id, feld, wert, reihenfolge)
		 VALUES (?, ?, ?, 1)`, karteID, feld, wert); err != nil {
		t.Fatalf("das Feld %q setzen: %v", feld, err)
	}
}
