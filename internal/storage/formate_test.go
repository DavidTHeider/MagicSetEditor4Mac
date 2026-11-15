package storage_test

// **Der Formatspeicher** — Werkstück 4 von Phase 4.6
// (#1293),
// nach Vermerk 0061.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Format lässt sich anlegen und wiederfinden.**
func TestEinFormatLaesstSichAnlegenUndWiederfinden(t *testing.T) {
	ablage := storage.NeueFormatablage(frischeDatenbank(t))

	angelegt, err := ablage.Anlegen(probekontext(t), domain.Format{
		Name: "Standard", Spiel: "magic",
		Mindestgroesse: 60, Sideboardgroesse: 15, HoechstzahlGleich: 4,
		Scryfallname: "standard",
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if angelegt.ID == 0 || angelegt.Fassung != 1 {
		t.Errorf("%+v", angelegt)
	}

	gefunden, err := ablage.Format(probekontext(t), angelegt.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gefunden.Name != "Standard" || gefunden.Mindestgroesse != 60 ||
		gefunden.Sideboardgroesse != 15 || gefunden.HoechstzahlGleich != 4 ||
		gefunden.Scryfallname != "standard" {
		t.Errorf("%+v", gefunden)
	}
}

// **Ein Format, das es nicht gibt, ist ein fachliches Nein.**
func TestEinFehlendesFormatIstEinNein(t *testing.T) {
	if _, err := storage.NeueFormatablage(frischeDatenbank(t)).Format(
		probekontext(t), 999); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}

// **Zwei Formate desselben Spiels mit demselben Namen sind ein Konflikt.**
func TestZweiFormateMitDemselbenNamenSindEinKonflikt(t *testing.T) {
	ablage := storage.NeueFormatablage(frischeDatenbank(t))

	format := domain.Format{Name: "Standard", Spiel: "magic"}

	if _, err := ablage.Anlegen(probekontext(t), format); err != nil {
		t.Fatalf("das erste: %v", err)
	}

	if _, err := ablage.Anlegen(probekontext(t), format); !errors.Is(
		err, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet ErrConflict", err)
	}
}

// **Die Liste gibt die Formate eines Spiels — ein leeres Spiel heißt: alle.**
func TestDieFormatlisteEinesSpiels(t *testing.T) {
	ablage := storage.NeueFormatablage(frischeDatenbank(t))

	for _, format := range []domain.Format{
		{Name: "Standard", Spiel: "magic"},
		{Name: "Commander", Spiel: "magic"},
		{Name: "Golden Age", Spiel: "vs"},
	} {
		if _, err := ablage.Anlegen(probekontext(t), format); err != nil {
			t.Fatalf("anlegen: %v", err)
		}
	}

	magic, err := ablage.Formate(probekontext(t), "magic")
	if err != nil || len(magic) != 2 {
		t.Fatalf("die Liste: %v, %d Formate", err, len(magic))
	}

	// **Geordnet vom Bestand**, sonst sortierte jede Fläche für sich.
	if magic[0].Name != "Commander" || magic[1].Name != "Standard" {
		t.Errorf("%q, %q", magic[0].Name, magic[1].Name)
	}

	alle, err := ablage.Formate(probekontext(t), "")
	if err != nil || len(alle) != 3 {
		t.Errorf("alle: %v, %d Formate", err, len(alle))
	}
}

// **Ein Format lässt sich ändern, und die Fassung ist die Bedingung.**
func TestEinFormatLaesstSichAendern(t *testing.T) {
	ablage := storage.NeueFormatablage(frischeDatenbank(t))

	angelegt, err := ablage.Anlegen(probekontext(t), domain.Format{
		Name: "Standard", Spiel: "magic", Mindestgroesse: 60,
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	angelegt.Mindestgroesse = 40

	geaendert, err := ablage.Aendern(probekontext(t), angelegt)
	if err != nil {
		t.Fatalf("ändern: %v", err)
	}

	if geaendert.Fassung != 2 {
		t.Errorf("Fassung %d, erwartet 2", geaendert.Fassung)
	}

	// **Derselbe Stand ein zweites Mal** — der Aufrufer hat die erste
	// Änderung nicht gesehen.
	if _, err := ablage.Aendern(probekontext(t), angelegt); !errors.Is(
		err, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet ErrConflict", err)
	}
}

// **Ein Format, das es nicht gibt, lässt sich nicht ändern.**
func TestEinFehlendesFormatLaesstSichNichtAendern(t *testing.T) {
	if _, err := storage.NeueFormatablage(frischeDatenbank(t)).Aendern(
		probekontext(t), domain.Format{ID: 999, Fassung: 1, Name: "X", Spiel: "magic"},
	); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}

// **Ein Format lässt sich löschen — und das Deck bleibt.**
//
// `ON DELETE SET NULL` und nicht `CASCADE`: wer ein Format löscht, wirft keine
// Decks weg. Das Deck ist danach **ungeprüft**, nicht verschwunden.
func TestEinGeloeschtesFormatLaesstDasDeckStehen(t *testing.T) {
	db := frischeDatenbank(t)
	formate := storage.NeueFormatablage(db)
	decks := storage.NeueDeckablage(db)

	format, err := formate.Anlegen(probekontext(t), domain.Format{
		Name: "Standard", Spiel: "magic",
	})
	if err != nil {
		t.Fatalf("das Format: %v", err)
	}

	deck, err := decks.Anlegen(probekontext(t), domain.Deck{
		Name: "Aggro", Spiel: "magic", FormatID: format.ID,
	})
	if err != nil {
		t.Fatalf("das Deck: %v", err)
	}

	if loeschfehler := formate.Loeschen(probekontext(t), format.ID); loeschfehler != nil {
		t.Fatalf("löschen: %v", loeschfehler)
	}

	geblieben, lesefehler := decks.Deck(probekontext(t), deck.ID)
	if lesefehler != nil {
		t.Fatalf("das Deck ist weg: %v", lesefehler)
	}

	if geblieben.FormatID != 0 {
		t.Errorf("das Deck hält noch das Format %d", geblieben.FormatID)
	}
}

// **Ein Format, das es nicht gibt, lässt sich nicht löschen.**
func TestEinFehlendesFormatLaesstSichNichtLoeschen(t *testing.T) {
	if err := storage.NeueFormatablage(frischeDatenbank(t)).Loeschen(
		probekontext(t), 999); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}

// **Ein Deck lässt sich auf ein Format stellen — und wieder herunter.**
//
// Ohne diesen Weg trüge `decks.format_id` einen Wert, den nur das Anlegen
// setzen kann: die Wahl wäre einmalig und unwiderruflich.
func TestEinDeckLaesstSichAufEinFormatStellen(t *testing.T) {
	db := frischeDatenbank(t)
	formate := storage.NeueFormatablage(db)
	decks := storage.NeueDeckablage(db)

	format, err := formate.Anlegen(probekontext(t), domain.Format{
		Name: "Standard", Spiel: "magic",
	})
	if err != nil {
		t.Fatalf("das Format: %v", err)
	}

	deck, err := decks.Anlegen(probekontext(t), domain.Deck{Name: "Aggro", Spiel: "magic"})
	if err != nil {
		t.Fatalf("das Deck: %v", err)
	}

	deck.FormatID = format.ID

	gestellt, err := decks.Aendern(probekontext(t), deck)
	if err != nil {
		t.Fatalf("stellen: %v", err)
	}

	if gelesen := deckLesen(t, decks, deck.ID); gelesen.FormatID != format.ID {
		t.Errorf("FormatID %d, erwartet %d", gelesen.FormatID, format.ID)
	}

	// **Und wieder herunter** — null heißt: kein Format, nicht „behalte das
	// alte".
	gestellt.FormatID = 0

	if _, err := decks.Aendern(probekontext(t), gestellt); err != nil {
		t.Fatalf("herunterstellen: %v", err)
	}

	if gelesen := deckLesen(t, decks, deck.ID); gelesen.FormatID != 0 {
		t.Errorf("das Format blieb: %d", gelesen.FormatID)
	}
}

// **Jeder Weg meldet einen kaputten Bestand.**
func TestJederFormatwegMeldetEinenKaputtenBestand(t *testing.T) {
	ablage := storage.NeueFormatablage(kaputterBestand(t))

	if _, err := ablage.Anlegen(probekontext(t), domain.Format{
		Name: "F", Spiel: "magic",
	}); err == nil {
		t.Error("anlegen schwieg")
	}

	if _, err := ablage.Format(probekontext(t), 1); err == nil {
		t.Error("lesen schwieg")
	}

	if _, err := ablage.Formate(probekontext(t), "magic"); err == nil {
		t.Error("die Liste schwieg")
	}

	if _, err := ablage.Aendern(probekontext(t), domain.Format{
		ID: 1, Fassung: 1, Name: "F", Spiel: "magic",
	}); err == nil {
		t.Error("ändern schwieg")
	}

	if err := ablage.Loeschen(probekontext(t), 1); err == nil {
		t.Error("löschen schwieg")
	}
}
