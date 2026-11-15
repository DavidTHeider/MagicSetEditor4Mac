// **Ein Deck kopieren — an jeder Fläche** — DOM-03.AK-02
// (#1698).
//
// # Der Ablauf
//
// Deck anlegen → vier Karten legen → unter neuem Namen kopieren → die Kopie
// lesen → noch einmal unter dem **alten** Namen kopieren.
//
// **Dasselbe Ergebnis und dasselbe Nein**: die Kopie hat ihre eigene Kennung
// und dieselben vier Karten; ein Name, den es im Spiel schon gibt, ist an
// jeder Fläche ein fachliches Nein und nichts entsteht (§16).
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestEinDeckLaesstSichUeberallKopieren(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := deckumgebung(t)

		deckID := strings.TrimSpace(mussKommandozeile(t, umgebung,
			"deck", "neu", "Aggro", "magic"))
		mussKommandozeile(t, umgebung, "deck", "legen", deckID, druckkennung, "4")

		kopieID := strings.TrimSpace(mussKommandozeile(t, umgebung,
			"deck", "kopieren", deckID, "Aggro II"))
		if kopieID == "" || kopieID == deckID {
			t.Fatalf("die Kopie hat die Kennung %q, das Original %q", kopieID, deckID)
		}

		pruefeKopie(t, mussKommandozeile(t, umgebung, "deck", kopieID, "--json"))

		_, _, code := kommandozeile(t, umgebung, "deck", "kopieren", deckID, "Aggro")
		if code != 1 {
			t.Errorf("ein vergebener Name gab %d, erwartet 1 (fachliches Nein)", code)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := deckumgebung(t)

		deckID := kennungAusMCP(t, umgebung, "schreiben.deck.anlegen",
			map[string]any{"name": "Aggro", "spiel": "magic"})
		mussMCP(t, umgebung, "schreiben.deckplatz.legen", map[string]any{
			"deckID": deckID, "printing": druckkennung, "menge": 4, "bereich": "hauptdeck",
		})

		kopieID := kennungAusMCP(t, umgebung, "schreiben.deck.kopieren",
			map[string]any{"deckID": deckID, "name": "Aggro II"})
		if kopieID == deckID {
			t.Fatalf("die Kopie teilt die Kennung %d mit dem Original", kopieID)
		}

		pruefeKopie(t, mussMCP(t, umgebung, "lesen.deck", map[string]any{"deckID": kopieID}))

		roh, istFehler := mcpRuf(t, umgebung, "schreiben.deck.kopieren",
			map[string]any{"deckID": deckID, "name": "Aggro"})
		if !istFehler {
			t.Errorf("ein vergebener Name kam durch: %s", roh)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := deckumgebung(t)
		an := starteDienst(t, umgebung)

		deckID := kennungAusREST(t, an, "/decks", `{"name":"Aggro","spiel":"magic"}`)
		mussREST(t, an, http.MethodPut, "/decks/"+deckID+"/plaetze",
			`{"printing":"`+druckkennung+`","menge":4,"bereich":"hauptdeck"}`)

		kopieID := kennungAusREST(t, an, "/decks/"+deckID+"/kopie", `{"name":"Aggro II"}`)
		if kopieID == deckID {
			t.Fatalf("die Kopie teilt die Kennung %s mit dem Original", kopieID)
		}

		pruefeKopie(t, mussREST(t, an, http.MethodGet, "/decks/"+kopieID, ""))

		status, rumpf := an.ruf(t, http.MethodPost, "/decks/"+deckID+"/kopie", `{"name":"Aggro"}`)
		if status < 400 || status >= 500 {
			t.Errorf("ein vergebener Name gab %d (%s), erwartet ein fachliches Nein", status, rumpf)
		}
	})
}

// pruefeKopie sichert zu, dass die Kopie Namen und Karten des Originals trägt.
func pruefeKopie(t *testing.T, roh string) {
	t.Helper()

	var kopie struct {
		Name      string `json:"name"`
		Hauptdeck int    `json:"hauptdeck"`
		Plaetze   []struct {
			Printing string `json:"printing"`
			Menge    int    `json:"menge"`
		} `json:"plaetze"`
	}

	// **Die Kommandozeile umschließt das Deck** — dieselbe Antwort wie die
	// Bindung, mit `deck` als Feld.
	var umschlossen struct {
		Deck json.RawMessage `json:"deck"`
	}

	daten := []byte(roh)
	if err := json.Unmarshal(daten, &umschlossen); err == nil && len(umschlossen.Deck) > 0 {
		daten = umschlossen.Deck
	}

	if err := json.Unmarshal(daten, &kopie); err != nil {
		t.Fatalf("die Kopie ist kein gültiges JSON: %v (%s)", err, roh)
	}

	if kopie.Name != "Aggro II" || kopie.Hauptdeck != 4 || len(kopie.Plaetze) != 1 ||
		kopie.Plaetze[0].Printing != druckkennung || kopie.Plaetze[0].Menge != 4 {
		t.Errorf("die Kopie ist %s", roh)
	}
}
