// **Aus einem Set ein Deck — an jeder Fläche** — FMT-04
// (#1693).
//
// # Der Ablauf
//
// Ein Set einfahren → daraus ein Deck mit zwei von jeder Karte → das Deck
// lesen → noch einmal mit der Menge null.
//
// **Dasselbe Ergebnis und dasselbe Nein**: drei Plätze zu je zwei, sechs
// Karten im Hauptdeck, der Name des Sets; eine Menge unter eins ist an jeder
// Fläche ein fachliches Nein, und kein zweites Deck entsteht (§16).
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestAusEinemSetWirdUeberallEinDeck(t *testing.T) {
	t.Parallel()

	einfahren := func(t *testing.T, umgebung []string) string {
		t.Helper()

		aus, fehler, code := kommandozeile(t, umgebung, "oeffnen", probeset(t), "--json")
		if code != 0 {
			t.Fatalf("einfahren: %d (%s)", code, fehler)
		}

		return setkennungAus(t, aus)
	}

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := deckumgebung(t)
		setID := einfahren(t, umgebung)

		deckID := strings.TrimSpace(mussKommandozeile(t, umgebung, "deck", "aus-set", setID, "2"))
		pruefeSetdeck(t, mussKommandozeile(t, umgebung, "deck", deckID, "--json"))

		if _, _, code := kommandozeile(t, umgebung, "deck", "aus-set", setID, "0"); code != 1 {
			t.Errorf("die Menge null gab %d, erwartet 1 (fachliches Nein)", code)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := deckumgebung(t)
		setID, _ := strconv.Atoi(einfahren(t, umgebung)) //nolint:errcheck // eine Kennung aus dem Bestand

		deckID := kennungAusMCP(t, umgebung, "schreiben.deck.ausset", map[string]any{"setID": setID, "menge": 2})
		pruefeSetdeck(t, mussMCP(t, umgebung, "lesen.deck", map[string]any{"deckID": deckID}))

		if roh, istFehler := mcpRuf(t, umgebung, "schreiben.deck.ausset",
			map[string]any{"setID": setID, "menge": 0}); !istFehler {
			t.Errorf("die Menge null kam durch: %s", roh)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := deckumgebung(t)
		setID := einfahren(t, umgebung)
		an := starteDienst(t, umgebung)

		deckID := kennungAusREST(t, an, "/decks/aus-set", `{"setID":`+setID+`,"menge":2}`)
		pruefeSetdeck(t, mussREST(t, an, http.MethodGet, "/decks/"+deckID, ""))

		if status, rumpf := an.ruf(t, http.MethodPost, "/decks/aus-set", `{"setID":`+setID+`,"menge":0}`); status < 400 || status >= 500 {
			t.Errorf("die Menge null gab %d (%s), erwartet ein fachliches Nein", status, rumpf)
		}
	})
}

// pruefeSetdeck sichert zu, dass das Deck jede Karte des Probesets zweimal trägt.
func pruefeSetdeck(t *testing.T, roh string) {
	t.Helper()

	var deck struct {
		Name      string `json:"name"`
		Hauptdeck int    `json:"hauptdeck"`
		Plaetze   []struct {
			KarteID string `json:"karteID"`
			Menge   int    `json:"menge"`
		} `json:"plaetze"`
	}

	var umschlossen struct {
		Deck json.RawMessage `json:"deck"`
	}

	daten := []byte(roh)
	if err := json.Unmarshal(daten, &umschlossen); err == nil && len(umschlossen.Deck) > 0 {
		daten = umschlossen.Deck
	}

	if err := json.Unmarshal(daten, &deck); err != nil {
		t.Fatalf("das Deck ist kein gültiges JSON: %v (%s)", err, roh)
	}

	if deck.Name != "Mein Probeset" || deck.Hauptdeck != 6 || len(deck.Plaetze) != 3 {
		t.Fatalf("das Deck ist %s", roh)
	}

	for _, platz := range deck.Plaetze {
		if platz.KarteID == "" || platz.Menge != 2 {
			t.Errorf("ein Platz ist %+v", platz)
		}
	}
}
