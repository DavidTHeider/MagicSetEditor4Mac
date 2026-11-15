package app_test

// **Ein Deck kopieren, an der Naht** — DOM-03.AK-02
// (#1698).

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Kopieren gibt das Deck unter neuem Namen und neuer Kennung zurück.
func (d *deckprobe) Kopieren(_ context.Context, quelle int64, name string) (
	domain.Deck, error,
) {
	if d.fehler != nil {
		return domain.Deck{}, d.fehler
	}

	kopie := d.deck
	kopie.ID, kopie.Name, kopie.Fassung = quelle+1, name, 1
	kopie.Zaehlen()

	return kopie, nil
}

func TestEinDeckLaesstSichKopieren(t *testing.T) {
	anwendung, decks, _, _, _ := deckanwendung(t)

	decks.deck = domain.Deck{ID: 7, Name: "Aggro", Spiel: "magic", Plaetze: []domain.Deckplatz{
		{ID: 1, Printing: "a", Menge: 4, Bereich: domain.BereichHauptdeck},
	}}

	antwort := anwendung.DeckKopieren(t.Context(), "7", "Aggro II")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Deck.ID != "8" || antwort.Deck.Name != "Aggro II" || antwort.Deck.Hauptdeck != 4 {
		t.Errorf("%+v", antwort.Deck)
	}
}

// **Eine unlesbare Kennung ist ein Feldfehler, ein Nein des Dienstes wird
// übersetzt** — dieselben zwei Wege wie bei jeder Deckbindung.
func TestEineKopieMeldetIhreFehler(t *testing.T) {
	anwendung, decks, _, _, _ := deckanwendung(t)

	if antwort := anwendung.DeckKopieren(t.Context(), "x", "N"); antwort.Fehler == nil ||
		antwort.Fehler.Feld != "deckID" {
		t.Errorf("die Kennung „x“ gab %+v", antwort.Fehler)
	}

	decks.fehler = domain.ErrNotFound

	if antwort := anwendung.DeckKopieren(t.Context(), "7", "N"); antwort.Fehler == nil ||
		antwort.Fehler.Code != "nichtGefunden" {
		t.Errorf("ein fehlendes Deck gab %+v", antwort.Fehler)
	}
}
