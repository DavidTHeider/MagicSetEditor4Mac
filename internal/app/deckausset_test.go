package app_test

// **Aus einem Set ein Deck — die Übersetzung an der Fläche** — FMT-04
// (#1693).

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

type setdeckprobe struct {
	setID  int64
	name   string
	menge  int
	fehler error
}

func (s *setdeckprobe) Anlegen(_ context.Context, setID int64, name string, menge int) (domain.Deck, error) {
	s.setID, s.name, s.menge = setID, name, menge

	deck := domain.Deck{ID: 5, Name: "Werkstatt", Spiel: "magic", Plaetze: []domain.Deckplatz{
		{ID: 1, KarteID: 11, Menge: 2, Bereich: domain.BereichHauptdeck},
	}}
	deck.Zaehlen() // wie `storage.Deckablage.Uebernehmen`

	return deck, s.fehler
}

func TestEinDeckEntstehtAusEinemSet(t *testing.T) {
	anwendung := app.NeueApp("")
	probe := &setdeckprobe{}
	anwendung.MitSetdeck(probe)

	antwort := app.NeueBindungen(anwendung).DeckAusSet(t.Context(), "4", "", 2)
	if antwort.Fehler != nil || antwort.Deck == nil || antwort.Deck.ID != "5" || antwort.Deck.Hauptdeck != 2 {
		t.Fatalf("%+v %+v", antwort.Deck, antwort.Fehler)
	}

	if probe.setID != 4 || probe.name != "" || probe.menge != 2 {
		t.Errorf("an den Dienst ging %+v", probe)
	}

	if antwort := anwendung.DeckAusSet(t.Context(), "x", "", 1); antwort.Fehler == nil || antwort.Fehler.Feld != "setID" {
		t.Errorf("die Kennung „x“ gab %+v", antwort.Fehler)
	}

	probe.fehler = domain.ErrNotFound

	if antwort := anwendung.DeckAusSet(t.Context(), "4", "", 1); antwort.Fehler == nil || antwort.Fehler.Code != "nichtGefunden" {
		t.Errorf("ein fehlendes Set gab %+v", antwort.Fehler)
	}
}
