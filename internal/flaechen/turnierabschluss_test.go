// **Abschluss und Protokoll eines Turniers — an jeder Fläche** —
// DOM-04.AK-04 (#1699),
// Vermerk 0139.
//
// # Der Ablauf
//
// Zwei Namen, eine gespielte Runde → abschließen → ein Anmelden wird
// abgewiesen → wiedereröffnen → das Protokoll nennt beides. **Dieselbe
// Antwort an jeder Fläche** (§16).
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

// pruefeProtokoll hält, dass Abschluss und Wiedereröffnen darin stehen.
func pruefeProtokoll(t *testing.T, flaeche, roh string) {
	t.Helper()

	var eintraege []struct {
		Vorgang string `json:"vorgang"`
	}

	if err := json.Unmarshal([]byte(roh), &eintraege); err != nil {
		t.Fatalf("%s: das Protokoll ist kein gültiges JSON: %v (%s)", flaeche, err, roh)
	}

	vorgaenge := make([]string, 0, len(eintraege))
	for _, eintrag := range eintraege {
		vorgaenge = append(vorgaenge, eintrag.Vorgang)
	}

	if liste := strings.Join(vorgaenge, " "); !strings.HasSuffix(liste, "abgeschlossen wiedereroeffnet") ||
		!strings.HasPrefix(liste, "angelegt angemeldet angemeldet runde ergebnis") {
		t.Errorf("%s: das Protokoll lautet %q", flaeche, liste)
	}
}

func TestEinTurnierSchliesstUeberallGleichAb(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := deckumgebung(t)
		turnierID, _ := gespielteRunde(t, umgebung)

		mussKommandozeile(t, umgebung, "turnier", "abschliessen", turnierID)
		if _, _, code := kommandozeile(t, umgebung, "turnier", "anmelden", turnierID, "Cleo"); code == 0 {
			t.Error("nach dem Abschluss angemeldet")
		}

		mussKommandozeile(t, umgebung, "turnier", "wiedereroeffnen", turnierID)

		var umschlag struct {
			Eintraege json.RawMessage `json:"eintraege"`
		}

		roh := mussKommandozeile(t, umgebung, "turnier", "protokoll", turnierID, "--json")
		if err := json.Unmarshal([]byte(roh), &umschlag); err != nil {
			t.Fatalf("die Antwort: %v (%s)", err, roh)
		}

		pruefeProtokoll(t, "Kommandozeile", string(umschlag.Eintraege))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := deckumgebung(t)
		turnierID, _ := gespielteRunde(t, umgebung)

		kennung, err := strconv.ParseInt(turnierID, 10, 64)
		if err != nil {
			t.Fatalf("die Kennung: %v", err)
		}

		mussMCP(t, umgebung, "schreiben.turnier.abschliessen", map[string]any{"turnierID": kennung})
		mussMCP(t, umgebung, "schreiben.turnier.wiedereroeffnen", map[string]any{"turnierID": kennung})
		pruefeProtokoll(t, "MCP", mussMCP(t, umgebung, "lesen.turnier.protokoll",
			map[string]any{"turnierID": kennung}))
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := deckumgebung(t)
		turnierID, _ := gespielteRunde(t, umgebung)
		an := starteDienst(t, umgebung)

		mussREST(t, an, http.MethodPost, "/turniere/"+turnierID+"/abschluss", "")
		mussREST(t, an, http.MethodDelete, "/turniere/"+turnierID+"/abschluss", "")
		pruefeProtokoll(t, "REST", mussREST(t, an, http.MethodGet, "/turniere/"+turnierID+"/protokoll", ""))
	})
}
