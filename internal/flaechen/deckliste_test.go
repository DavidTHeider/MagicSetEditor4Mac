package flaechen_test

// **Ein Deck als Liste an jeder Fläche ohne Fenster** — NET-04
// (#1708).
//
// **Und die Liste liest sich zurück**: was `mse deck liste` ausgibt, deutet
// `mse deck-liste` in ein zweites Deck. Ohne diesen Rundlauf wäre „die
// Gestalt, die die Einfuhr versteht" eine Behauptung.

import (
	"net/http"
	"strings"
	"testing"
)

func TestDieDecklisteAnJederFlaeche(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)

	deckID := strings.TrimSpace(mussKommandozeile(t, umgebung, "deck", "neu", "Aggro", "magic"))
	mussKommandozeile(t, umgebung, "deck", "legen", deckID, druckkennung, "4")

	liste := mussKommandozeile(t, umgebung, "deck", "liste", deckID)
	if !strings.HasPrefix(liste, "Deck\n4 ") {
		t.Fatalf("mse deck liste gab %q", liste)
	}

	// **Der Rundlauf**: dieselbe Liste in ein zweites Deck gedeutet.
	zweites := strings.TrimSpace(mussKommandozeile(t, umgebung, "deck", "neu", "Kopie", "magic"))
	gedeutet := mitEingabe(t, umgebung, liste, "deck-liste", zweites)

	if !strings.Contains(gedeutet, "4") {
		t.Errorf("die Deutung der eigenen Liste: %q", gedeutet)
	}

	if roh := mussMCP(t, umgebung, "lesen.deck.liste", map[string]any{"deckID": zahl(t, deckID)}); !strings.Contains(roh, "Deck") {
		t.Errorf("MCP: %s", roh)
	}

	an := starteDienst(t, umgebung)
	if roh := mussREST(t, an, http.MethodGet, "/decks/"+deckID+"/liste", ""); !strings.Contains(roh, `"liste"`) {
		t.Errorf("REST: %s", roh)
	}
}
