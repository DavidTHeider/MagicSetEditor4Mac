package storage_test

// **Ein Deck ändern, und wissen, zu welchem Spiel eine Karte gehört** —
// Werkstück 3 von Phase 4.6
// (#1293).
//
// Beides braucht der Deckbau: das eine, um umzubenennen, das andere, um eine
// **eigene** Karte abzuweisen, die zu einem anderen Spiel gehört.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Deck lässt sich umbenennen, und die Fassung steigt.**
func TestEinDeckLaesstSichAendern(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	angelegt, err := ablage.Anlegen(probekontext(t), domain.Deck{
		Name: "Erst so", Spiel: "magic",
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	angelegt.Name = "Dann so"
	angelegt.Notiz = "umbenannt"

	geaendert, err := ablage.Aendern(probekontext(t), angelegt)
	if err != nil {
		t.Fatalf("ändern: %v", err)
	}

	if geaendert.Fassung != 2 {
		t.Errorf("Fassung %d, erwartet 2", geaendert.Fassung)
	}

	gelesen, err := ablage.Deck(probekontext(t), angelegt.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen.Name != "Dann so" || gelesen.Notiz != "umbenannt" {
		t.Errorf("%+v", gelesen)
	}
}

// **Eine überholte Fassung ist ein Konflikt** — §7, Regel 2.
func TestEineUeberholteDeckfassungIstEinKonflikt(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	angelegt, err := ablage.Anlegen(probekontext(t), domain.Deck{
		Name: "Aggro", Spiel: "magic",
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if _, err := ablage.Aendern(probekontext(t), angelegt); err != nil {
		t.Fatalf("die erste Änderung: %v", err)
	}

	// **Derselbe Stand ein zweites Mal** — der Aufrufer hat die erste
	// Änderung nicht gesehen.
	if _, err := ablage.Aendern(probekontext(t), angelegt); !errors.Is(
		err, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet ErrConflict", err)
	}
}

// **Ein Deck, das es nicht gibt, ist ein fachliches Nein** — und nicht
// dasselbe wie ein Konflikt.
func TestEinFehlendesDeckAendernIstEinNein(t *testing.T) {
	db := frischeDatenbank(t)

	if _, err := storage.NeueDeckablage(db).Aendern(probekontext(t), domain.Deck{
		ID: 999, Fassung: 1, Name: "Nichts", Spiel: "magic",
	}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}

// **Eine eigene Karte gehört zum Spiel ihres Sets.**
func TestSpielZurKarte(t *testing.T) {
	db := frischeDatenbank(t)
	karte := kartenkennungFuerDeck(t, db)

	spiel, err := storage.NeuerSetLeser(db).SpielZurKarte(probekontext(t), karte)
	if err != nil {
		t.Fatalf("das Spiel lesen: %v", err)
	}

	if spiel != "magic" {
		t.Errorf("Spiel %q, erwartet magic", spiel)
	}
}

// **Eine Karte, die es nicht gibt, hat kein Spiel.**
func TestSpielZuEinerFehlendenKarte(t *testing.T) {
	db := frischeDatenbank(t)

	if _, err := storage.NeuerSetLeser(db).SpielZurKarte(
		probekontext(t), 999); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}
