package main

// **Aus einem Set ein Deck an der Kommandozeile** — FMT-04
// (#1693).

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAusEinemSetWirdAnDerKommandozeileEinDeck(t *testing.T) {
	ort, setID, _ := eingefahrenesSet(t)

	deckID := strings.TrimSpace(mussLaufen(t, ort, "deck", "aus-set", setID, "2", "Probedeck"))
	if deckID == "" {
		t.Fatal("keine Kennung")
	}

	var deck struct {
		Deck struct {
			Name    string `json:"name"`
			Plaetze []struct {
				Menge int `json:"menge"`
			} `json:"plaetze"`
		} `json:"deck"`
	}

	if err := json.Unmarshal([]byte(mussLaufen(t, ort, "deck", deckID, "--json")), &deck); err != nil ||
		deck.Deck.Name != "Probedeck" || len(deck.Deck.Plaetze) == 0 || deck.Deck.Plaetze[0].Menge != 2 {
		t.Errorf("das Deck: %+v, %v", deck, err)
	}

	// **Ohne Namen heißt es wie das Set**, und JSON gibt es auch hier.
	if aus := mussLaufen(t, ort, "deck", "aus-set", setID, "1", "--json"); !strings.HasPrefix(aus, "{") {
		t.Errorf("als JSON kam %q", aus)
	}

	for name, fall := range map[string]struct {
		argumente []string
		code      int
	}{
		"ohne Menge":           {[]string{"deck", "aus-set", setID}, CodeBedienung},
		"eine Menge ohne Zahl": {[]string{"deck", "aus-set", setID, "viele"}, CodeBedienung},
		"eine Menge null":      {[]string{"deck", "aus-set", setID, "0"}, CodeFachlich},
		"ein fehlendes Set":    {[]string{"deck", "aus-set", "999", "1"}, CodeFachlich},
	} {
		if code, _, _ := fahreAn(t, ort, fall.argumente...); code != fall.code {
			t.Errorf("%s: %d, erwartet %d", name, code, fall.code)
		}
	}
}
