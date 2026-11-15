package app_test

// **Ein Deck als Liste über jede Fläche** — NET-04
// (#1708).
// Die Gestalt steht in `service.Deckliste`; hier wird die Verdrahtung
// geprüft — Fläche der Kommandozeile und Werkstatt, die MCP und REST
// bedienen.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestDieDecklisteUeberDieFlaechen(t *testing.T) {
	flaeche := deckflaeche(t)

	deck := flaeche.DeckAnlegen(probekontext(t), "Aggro", "magic", "")
	if deck.Fehler != nil {
		t.Fatalf("anlegen: %v", deck.Fehler)
	}

	if platz := flaeche.DeckplatzLegen(probekontext(t), deck.Deck.ID, "",
		"a1b2c3d4-0000-4000-8000-00000000000a", 4, "hauptdeck"); platz.Fehler != nil {
		t.Fatalf("legen: %v", platz.Fehler)
	}

	liste := flaeche.DecklisteAusgeben(probekontext(t), deck.Deck.ID)
	if liste.Fehler != nil || !strings.HasPrefix(liste.Liste, "Deck\n4 ") {
		t.Errorf("die Liste: %+v", liste)
	}

	// **Ein Deck, das es nicht gibt, ist ein Nein** — und eine Kennung, die
	// keine ist, auch.
	if fehler := flaeche.DecklisteAusgeben(probekontext(t), "999").Fehler; fehler == nil ||
		fehler.Code != app.CodeNichtGefunden {
		t.Errorf("ein fremdes Deck: %+v", fehler)
	}

	if fehler := flaeche.DecklisteAusgeben(probekontext(t), "keine").Fehler; fehler == nil {
		t.Error("eine Kennung, die keine ist, kam durch")
	}

	werkstatt := app.NeueDeckwerkstatt(frischerBestand(t))

	angelegt, fehler := werkstatt.DeckAnlegen(probekontext(t), "Leer", "magic", "")
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	mussWerkstatt(t)(werkstatt.Deckliste(probekontext(t), kennungAusWerkstatt(t, angelegt)))
}
