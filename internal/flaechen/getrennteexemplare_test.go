// **Getrennte physische Exemplare bleiben getrennt** — FMT-04.AK-04
// (#1693).
//
// Zwei Exemplare **desselben Drucks** in verschiedenem Zustand sind zwei
// Stücke, keine Menge von zwei. Durch die CSV hinaus und in einen frischen
// Bestand zurück bleiben sie es — mit ihrer Menge und ihrem Zustand.
package flaechen_test

import (
	"encoding/json"
	"sort"
	"testing"
)

func TestGetrennteExemplareBleibenImRundlaufGetrennt(t *testing.T) {
	t.Parallel()

	quelle := eigenerBestand(t)
	mussKommandozeile(t, quelle, "exemplar", "probe-druck", "1", "nearMint", "de", "nonfoil",
		"2026-08-30", "4.17", "EUR", "aus dem Booster")
	mussKommandozeile(t, quelle, "exemplar", "probe-druck", "3", "played", "de", "nonfoil",
		"2026-08-31", "2.50", "EUR", "getauscht")

	text := mussKommandozeile(t, quelle, "sammlung-ausgeben", "csv")

	ziel := eigenerBestand(t)
	mitEingabe(t, ziel, text, "sammlung-einlesen")

	var sammlung struct {
		Exemplare []struct {
			Printing string `json:"printing"`
			Menge    int    `json:"menge"`
			Zustand  string `json:"zustand"`
			Notiz    string `json:"notiz"`
		} `json:"exemplare"`
	}

	roh := mussKommandozeile(t, ziel, "sammlung", "--json")
	if err := json.Unmarshal([]byte(roh), &sammlung); err != nil {
		t.Fatalf("die Sammlung lesen: %v (%s)", err, roh)
	}

	gelesen := make([]string, 0, len(sammlung.Exemplare))
	for _, eines := range sammlung.Exemplare {
		gelesen = append(gelesen, eines.Printing+"·"+eines.Zustand+"·"+eines.Notiz+"·"+string(rune('0'+eines.Menge)))
	}

	sort.Strings(gelesen)

	if len(gelesen) != 2 || gelesen[0] != "probe-druck·nearMint·aus dem Booster·1" ||
		gelesen[1] != "probe-druck·played·getauscht·3" {
		t.Errorf("zurück kamen %v", gelesen)
	}
}
