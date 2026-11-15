// **Was einem Deck fehlt, auf die Wunschliste — an jeder Fläche** —
// DOM-03.AK-02
// (#1698).
//
// # Der Ablauf
//
// Deck anlegen → vier Karten legen → wünschen → noch einmal wünschen.
//
// **Dieselbe Antwort an jeder Fläche**: beim ersten Mal ein neuer Wunsch über
// vier Stück, beim zweiten keiner — der Druck steht schon darauf, und die
// Antwort nennt ihn. Dass der zweite Ruf ihn **findet**, belegt, dass der
// erste geschrieben hat (§16).
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestFehlkartenKommenUeberallAufDieWunschliste(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := deckumgebung(t)

		deckID := strings.TrimSpace(mussKommandozeile(t, umgebung,
			"deck", "neu", "Aggro", "magic"))
		mussKommandozeile(t, umgebung, "deck", "legen", deckID, druckkennung, "4")

		pruefeWunsch(t, mussKommandozeile(t, umgebung, "deck", "wuenschen", deckID, "--json"), 1, 0)
		pruefeWunsch(t, mussKommandozeile(t, umgebung, "deck", "wuenschen", deckID, "--json"), 0, 1)
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := deckumgebung(t)

		deckID := kennungAusMCP(t, umgebung, "schreiben.deck.anlegen",
			map[string]any{"name": "Aggro", "spiel": "magic"})
		mussMCP(t, umgebung, "schreiben.deckplatz.legen", map[string]any{
			"deckID": deckID, "printing": druckkennung, "menge": 4, "bereich": "hauptdeck",
		})

		for _, erwartet := range [][2]int{{1, 0}, {0, 1}} {
			pruefeWunsch(t, mussMCP(t, umgebung, "schreiben.deck.fehlkarten.wuenschen",
				map[string]any{"deckID": deckID}), erwartet[0], erwartet[1])
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := deckumgebung(t)
		an := starteDienst(t, umgebung)

		deckID := kennungAusREST(t, an, "/decks", `{"name":"Aggro","spiel":"magic"}`)
		mussREST(t, an, http.MethodPut, "/decks/"+deckID+"/plaetze",
			`{"printing":"`+druckkennung+`","menge":4,"bereich":"hauptdeck"}`)

		for _, erwartet := range [][2]int{{1, 0}, {0, 1}} {
			pruefeWunsch(t, mussREST(t, an, http.MethodPost,
				"/decks/"+deckID+"/fehlkarten/wunschliste", ""), erwartet[0], erwartet[1])
		}
	})
}

// pruefeWunsch sichert zu, wie viele Wünsche neu und wie viele schon da sind.
func pruefeWunsch(t *testing.T, roh string, neu, vorhanden int) {
	t.Helper()

	var antwort struct {
		Neu []struct {
			Printing string `json:"printing"`
			Menge    int    `json:"menge"`
			Wunsch   bool   `json:"wunsch"`
		} `json:"neu"`
		Vorhanden []struct {
			Fehlend    int `json:"fehlend"`
			Gewuenscht int `json:"gewuenscht"`
		} `json:"vorhanden"`
	}

	if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	if len(antwort.Neu) != neu || len(antwort.Vorhanden) != vorhanden {
		t.Fatalf("neu %d, vorhanden %d — erwartet %d und %d (%s)",
			len(antwort.Neu), len(antwort.Vorhanden), neu, vorhanden, roh)
	}

	for _, wunsch := range antwort.Neu {
		if wunsch.Printing != druckkennung || wunsch.Menge != 4 || !wunsch.Wunsch {
			t.Errorf("neuer Wunsch %+v", wunsch)
		}
	}

	for _, schon := range antwort.Vorhanden {
		if schon.Fehlend != 4 || schon.Gewuenscht != 4 {
			t.Errorf("vorhandener Wunsch %+v", schon)
		}
	}
}
