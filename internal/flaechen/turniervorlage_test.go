// **Ein Turnier nach einer Vorlage — an jeder Fläche** — DOM-04.AK-01 und
// AK-03 (#1699).
//
// # Der Ablauf
//
// Ein Format → ein Turnier in diesem Format mit Anna samt Deck und Bert, eine
// Runde gespielt → eine Kopie mit Teilnehmenden. **Dieselbe Antwort an jeder
// Fläche** (§16): die Kopie trägt Format und Namen, aber **kein Deck, keine
// Liste und keine Runde** — eine eingereichte Liste gilt für ein Turnier.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

// kopiertesTurnier ist, was jede Fläche über die Kopie sagen muss.
type kopiertesTurnier struct {
	ID         string `json:"id"`
	FormatID   string `json:"formatID"`
	Teilnehmer []struct {
		Name      string            `json:"name"`
		DeckID    string            `json:"deckID"`
		Deckabzug []json.RawMessage `json:"deckabzug"`
	} `json:"teilnehmer"`
	Runden []json.RawMessage `json:"runden"`
}

// vorlagenturnier legt über die Kommandozeile an, was jede Fläche kopiert.
func vorlagenturnier(t *testing.T, umgebung []string) (turnierID, formatID string) {
	t.Helper()

	formatID = strings.TrimSpace(mussKommandozeile(t, umgebung,
		"format", "neu", "konstruiert", "Pauper", "magic"))
	deckID, _ := zweiDecks(t, umgebung)

	turnierID = strings.TrimSpace(mussKommandozeile(t, umgebung,
		"turnier", "neu", "Freitag", "magic", "--format", formatID))
	mussKommandozeile(t, umgebung, "turnier", "anmelden", turnierID, "Anna", deckID)
	mussKommandozeile(t, umgebung, "turnier", "anmelden", turnierID, "Bert")
	mussKommandozeile(t, umgebung, "turnier", "runde", turnierID)

	return turnierID, formatID
}

// pruefeTurnierkopie hält die Zusage an der Antwort einer Fläche.
func pruefeTurnierkopie(t *testing.T, flaeche, roh, vorlage, formatID string) {
	t.Helper()

	var kopie kopiertesTurnier
	if err := json.Unmarshal([]byte(roh), &kopie); err != nil {
		t.Fatalf("%s: die Antwort ist kein gültiges JSON: %v (%s)", flaeche, err, roh)
	}

	if kopie.ID == "" || kopie.ID == vorlage || kopie.FormatID != formatID {
		t.Errorf("%s: die Kopie %+v, erwartet eine neue Kennung und Format %s", flaeche, kopie, formatID)
	}

	if len(kopie.Runden) != 0 || len(kopie.Teilnehmer) != 2 {
		t.Fatalf("%s: %d Runden und %d Teilnehmende, erwartet 0 und 2",
			flaeche, len(kopie.Runden), len(kopie.Teilnehmer))
	}

	for _, teilnehmer := range kopie.Teilnehmer {
		if teilnehmer.DeckID != "" || len(teilnehmer.Deckabzug) != 0 {
			t.Errorf("%s: %s bringt ein Deck mit", flaeche, teilnehmer.Name)
		}
	}
}

func TestEinTurnierLaesstSichUeberallNachVorlageAnlegen(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := deckumgebung(t)
		vorlage, formatID := vorlagenturnier(t, umgebung)

		kopie := strings.TrimSpace(mussKommandozeile(t, umgebung,
			"turnier", "kopieren", vorlage, "Freitag II", "--mit-teilnehmenden"))

		// **Die Kommandozeile umschließt die Antwort** — `{"turnier": …}`.
		var umschlag struct {
			Turnier json.RawMessage `json:"turnier"`
		}

		roh := mussKommandozeile(t, umgebung, "turnier", kopie, "--json")
		if err := json.Unmarshal([]byte(roh), &umschlag); err != nil {
			t.Fatalf("die Antwort: %v (%s)", err, roh)
		}

		pruefeTurnierkopie(t, "Kommandozeile", string(umschlag.Turnier), vorlage, formatID)
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := deckumgebung(t)
		vorlage, formatID := vorlagenturnier(t, umgebung)

		kennung, err := strconv.ParseInt(vorlage, 10, 64)
		if err != nil {
			t.Fatalf("die Kennung: %v", err)
		}

		pruefeTurnierkopie(t, "MCP", mussMCP(t, umgebung, "schreiben.turnier.kopieren", map[string]any{
			"turnierID": kennung, "name": "Freitag II", "mitTeilnehmenden": true,
		}), vorlage, formatID)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := deckumgebung(t)
		vorlage, formatID := vorlagenturnier(t, umgebung)
		an := starteDienst(t, umgebung)

		pruefeTurnierkopie(t, "REST", mussREST(t, an, http.MethodPost, "/turniere/"+vorlage+"/kopie",
			`{"name":"Freitag II","mitTeilnehmenden":true}`), vorlage, formatID)
	})
}
