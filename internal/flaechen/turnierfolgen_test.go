// **Die Folgen einer Berichtigung — an jeder Fläche** — DOM-04.AK-02
// (#1699).
//
// # Der Ablauf
//
// Zwei Namen, eine Runde, Anna gewinnt → gefragt wird, was *Bert gewinnt*
// verschöbe. **Dieselbe Antwort an jeder Fläche** (§16): beide tauschen die
// Plätze — und eingetragen ist danach nichts.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

// gespielteRunde legt zwei Namen und eine Runde an, in der Anna gewinnt.
func gespielteRunde(t *testing.T, umgebung []string) (turnierID, paarungID string) {
	t.Helper()

	turnierID = strings.TrimSpace(mussKommandozeile(t, umgebung, "turnier", "neu", "Freitag", "magic"))
	mussKommandozeile(t, umgebung, "turnier", "anmelden", turnierID, "Anna")
	mussKommandozeile(t, umgebung, "turnier", "anmelden", turnierID, "Bert")
	mussKommandozeile(t, umgebung, "turnier", "runde", turnierID)

	paarungID = offenePaarungen(t, umgebung, turnierID)[0]
	mussKommandozeile(t, umgebung, "turnier", "ergebnis", turnierID, paarungID, "2", "0")

	return turnierID, paarungID
}

// pruefeFolgen hält die Zusage an der Antwort einer Fläche — und dass nichts
// eingetragen wurde.
func pruefeFolgen(t *testing.T, flaeche, roh string, umgebung []string, turnierID string) {
	t.Helper()

	var folgen struct {
		Runde      int `json:"runde"`
		Verschoben []struct {
			Name         string `json:"name"`
			PlatzVorher  int    `json:"platzVorher"`
			PlatzNachher int    `json:"platzNachher"`
		} `json:"verschoben"`
	}

	if err := json.Unmarshal([]byte(roh), &folgen); err != nil {
		t.Fatalf("%s: die Antwort ist kein gültiges JSON: %v (%s)", flaeche, err, roh)
	}

	if folgen.Runde != 1 || len(folgen.Verschoben) != 2 {
		t.Fatalf("%s: die Folgen %+v, erwartet zwei Wechsel in Runde 1", flaeche, folgen)
	}

	for _, wechsel := range folgen.Verschoben {
		if wechsel.Name == "Bert" && (wechsel.PlatzVorher != 2 || wechsel.PlatzNachher != 1) {
			t.Errorf("%s: Bert %+v, erwartet 2 → 1", flaeche, wechsel)
		}
	}

	if rangliste := mussKommandozeile(t, umgebung, "turnier", "rangliste", turnierID); !strings.HasPrefix(rangliste, "1.\tAnna") {
		t.Errorf("%s: die Frage hat geschrieben — %q", flaeche, rangliste)
	}
}

func TestDieFolgenEinerBerichtigungStehenUeberallGleich(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := deckumgebung(t)
		turnierID, paarungID := gespielteRunde(t, umgebung)

		var umschlag struct {
			Folgen json.RawMessage `json:"folgen"`
		}

		roh := mussKommandozeile(t, umgebung, "turnier", "folgen", turnierID, paarungID, "0", "2", "--json")
		if err := json.Unmarshal([]byte(roh), &umschlag); err != nil {
			t.Fatalf("die Antwort: %v (%s)", err, roh)
		}

		pruefeFolgen(t, "Kommandozeile", string(umschlag.Folgen), umgebung, turnierID)
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := deckumgebung(t)
		turnierID, paarungID := gespielteRunde(t, umgebung)

		turnier, err := strconv.ParseInt(turnierID, 10, 64)
		if err != nil {
			t.Fatalf("die Kennung: %v", err)
		}

		paarung, err := strconv.ParseInt(paarungID, 10, 64)
		if err != nil {
			t.Fatalf("die Paarung: %v", err)
		}

		pruefeFolgen(t, "MCP", mussMCP(t, umgebung, "lesen.turnier.folgen", map[string]any{
			"turnierID": turnier, "paarungID": paarung, "spieleB": 2,
		}), umgebung, turnierID)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := deckumgebung(t)
		turnierID, paarungID := gespielteRunde(t, umgebung)
		an := starteDienst(t, umgebung)

		pruefeFolgen(t, "REST", mussREST(t, an, http.MethodPost,
			"/turniere/"+turnierID+"/paarungen/"+paarungID+"/folgen", `{"spieleB":2}`),
			umgebung, turnierID)
	})
}
