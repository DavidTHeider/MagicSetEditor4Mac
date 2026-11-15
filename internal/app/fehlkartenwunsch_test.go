package app_test

// **Was einem Deck fehlt, auf die Wunschliste — an der Naht** — DOM-03.AK-02
// (#1698).

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

type wunschprobe struct {
	ergebnis service.Fehlkartenwuensche
	fehler   error
	deckID   int64
}

func (w *wunschprobe) Wuenschen(_ context.Context, deckID int64) (service.Fehlkartenwuensche, error) {
	w.deckID = deckID

	return w.ergebnis, w.fehler
}

func TestFehlkartenGehenUeberDieNahtAufDieWunschliste(t *testing.T) {
	probe := &wunschprobe{ergebnis: service.Fehlkartenwuensche{
		Neu: []domain.Exemplar{{
			ID: 5, Fassung: 1, Printing: "a", Menge: 3, Wunsch: true,
			Zustand: domain.ZustandNearMint, Sprache: "en", Name: "Blitz",
		}},
		Vorhanden: []service.Vorhandenerwunsch{{Habenzeile: service.Habenzeile{
			Printing: "b", Bezeichnung: "Glut", Gebraucht: 2, Fehlend: 2, Gewuenscht: 1,
		}}},
		Unbekannt: 1,
	}}

	anwendung := app.NeueApp("")
	anwendung.MitFehlkartenwunsch(probe)

	antwort := anwendung.DeckFehlkartenWuenschen(t.Context(), "7")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.deckID != 7 {
		t.Errorf("gefragt wurde Deck %d", probe.deckID)
	}

	if len(antwort.Neu) != 1 || antwort.Neu[0].ID != 5 || !antwort.Neu[0].Wunsch ||
		antwort.Neu[0].Menge != 3 || antwort.Neu[0].Name != "Blitz" {
		t.Errorf("neu: %+v", antwort.Neu)
	}

	if len(antwort.Vorhanden) != 1 || antwort.Vorhanden[0].Printing != "b" ||
		antwort.Vorhanden[0].Fehlend != 2 || antwort.Vorhanden[0].Gewuenscht != 1 {
		t.Errorf("vorhanden: %+v", antwort.Vorhanden)
	}

	if antwort.Unbekannt != 1 {
		t.Errorf("unbekannt: %d", antwort.Unbekannt)
	}
}

// **Eine unlesbare Kennung ist ein Feldfehler, ein Nein des Dienstes wird
// übersetzt — und die Listen bleiben Listen**, auch im Fehlerfall.
func TestDerWunschwegMeldetKennungUndDienstfehler(t *testing.T) {
	probe := &wunschprobe{}

	anwendung := app.NeueApp("")
	anwendung.MitFehlkartenwunsch(probe)

	if antwort := anwendung.DeckFehlkartenWuenschen(t.Context(), "x"); antwort.Fehler == nil ||
		antwort.Fehler.Feld != "deckID" || antwort.Neu == nil || antwort.Vorhanden == nil {
		t.Errorf("die Kennung „x“ gab %+v", antwort)
	}

	probe.fehler = domain.ErrNotFound

	if antwort := anwendung.DeckFehlkartenWuenschen(t.Context(), "7"); antwort.Fehler == nil ||
		antwort.Fehler.Code != app.CodeNichtGefunden || antwort.Neu == nil {
		t.Errorf("ein fehlendes Deck gab %+v", antwort)
	}
}
