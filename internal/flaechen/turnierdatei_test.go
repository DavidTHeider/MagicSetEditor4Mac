// **Ein Turnier als Datei und als Bericht — an jeder Fläche** —
// DOM-04.AK-03 (#1699),
// Vermerk 0140.
//
// # Der Ablauf
//
// Zwei Namen, eine gespielte Runde → ausgeben → unter anderem Namen
// einlesen → das neue Turnier trägt **denselben Verlauf** → der Bericht
// nennt die Runde. **Dieselbe Antwort an jeder Fläche** (§16).
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// umbenannt gibt die Ausgabe unter dem Namen »Samstag« — derselbe Name im
// selben Spiel wäre ein Konflikt.
func umbenannt(t *testing.T, flaeche, ausgabe string) string {
	t.Helper()

	if !strings.Contains(ausgabe, `"name": "Freitag"`) || strings.Contains(ausgabe, `"id"`) {
		t.Fatalf("%s: die Ausgabe lautet %q", flaeche, ausgabe)
	}

	return strings.Replace(ausgabe, `"Freitag"`, `"Samstag"`, 1)
}

// pruefeEingelesen hält, dass das eingelesene Turnier den Verlauf trägt.
func pruefeEingelesen(t *testing.T, flaeche, roh string) {
	t.Helper()

	var turnier struct {
		Name       string            `json:"name"`
		Teilnehmer []json.RawMessage `json:"teilnehmer"`
		Runden     []struct {
			Paarungen []struct {
				Ergebnis *struct {
					SpieleA int `json:"spieleA"`
				} `json:"ergebnis"`
			} `json:"paarungen"`
		} `json:"runden"`
	}

	if err := json.Unmarshal([]byte(roh), &turnier); err != nil {
		t.Fatalf("%s: das Turnier ist kein gültiges JSON: %v (%s)", flaeche, err, roh)
	}

	if turnier.Name != "Samstag" || len(turnier.Teilnehmer) != 2 || len(turnier.Runden) != 1 ||
		turnier.Runden[0].Paarungen[0].Ergebnis == nil || turnier.Runden[0].Paarungen[0].Ergebnis.SpieleA != 2 {
		t.Errorf("%s: eingelesen wurde %s", flaeche, roh)
	}
}

// pruefeBericht hält den Kopf und die Runde des Berichts.
func pruefeBericht(t *testing.T, flaeche, bericht string) {
	t.Helper()

	if !strings.HasPrefix(bericht, "# Freitag\n") || !strings.Contains(bericht, "## Runde 1") ||
		!strings.Contains(bericht, "| Anna | Bert | 2–0 |") {
		t.Errorf("%s: der Bericht lautet %q", flaeche, bericht)
	}
}

// ausDemFeld liest ein Textfeld aus einer JSON-Antwort.
func ausDemFeld(t *testing.T, roh, feld string) string {
	t.Helper()

	var umschlag map[string]any
	if err := json.Unmarshal([]byte(roh), &umschlag); err != nil {
		t.Fatalf("die Antwort: %v (%s)", err, roh)
	}

	text, _ := umschlag[feld].(string) //nolint:errcheck // ein fehlendes Feld prüft der Aufrufer

	return text
}

func TestEinTurnierReistUeberallGleich(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := deckumgebung(t)
		turnierID, _ := gespielteRunde(t, umgebung)

		datei := filepath.Join(t.TempDir(), "t.mseturnier")
		inhalt := umbenannt(t, "Kommandozeile", mussKommandozeile(t, umgebung, "turnier", "ausgeben", turnierID))

		if err := os.WriteFile(datei, []byte(inhalt), 0o600); err != nil {
			t.Fatalf("schreiben: %v", err)
		}

		kennung := ausDemFeld(t, mussKommandozeile(t, umgebung, "turnier", "einlesen", datei, "--json"), "kennung")
		pruefeEingelesen(t, "Kommandozeile",
			ausDemFeldRoh(t, mussKommandozeile(t, umgebung, "turnier", kennung, "--json"), "turnier"))
		pruefeBericht(t, "Kommandozeile", mussKommandozeile(t, umgebung, "turnier", "bericht", turnierID))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := deckumgebung(t)
		turnierID, _ := gespielteRunde(t, umgebung)

		kennung, err := strconv.ParseInt(turnierID, 10, 64)
		if err != nil {
			t.Fatalf("die Kennung: %v", err)
		}

		inhalt := umbenannt(t, "MCP", ausDemFeld(t,
			mussMCP(t, umgebung, "lesen.turnier.ausgabe", map[string]any{"turnierID": kennung}), "inhalt"))
		neu, err := strconv.ParseInt(ausDemFeld(t,
			mussMCP(t, umgebung, "schreiben.turnier.einlesen", map[string]any{"inhalt": inhalt}), "kennung"), 10, 64)
		if err != nil {
			t.Fatalf("die neue Kennung: %v", err)
		}

		pruefeEingelesen(t, "MCP", mussMCP(t, umgebung, "lesen.turnier", map[string]any{"turnierID": neu}))
		pruefeBericht(t, "MCP", ausDemFeld(t,
			mussMCP(t, umgebung, "lesen.turnier.bericht", map[string]any{"turnierID": kennung}), "inhalt"))
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := deckumgebung(t)
		turnierID, _ := gespielteRunde(t, umgebung)
		an := starteDienst(t, umgebung)

		inhalt := umbenannt(t, "REST", ausDemFeld(t,
			mussREST(t, an, http.MethodGet, "/turniere/"+turnierID+"/ausgabe", ""), "inhalt"))

		rumpf, err := json.Marshal(map[string]string{"inhalt": inhalt})
		if err != nil {
			t.Fatalf("der Rumpf: %v", err)
		}

		neu := ausDemFeld(t, mussREST(t, an, http.MethodPost, "/turniere/einlesen", string(rumpf)), "kennung")
		pruefeEingelesen(t, "REST", mussREST(t, an, http.MethodGet, "/turniere/"+neu, ""))
		pruefeBericht(t, "REST", ausDemFeld(t,
			mussREST(t, an, http.MethodGet, "/turniere/"+turnierID+"/bericht", ""), "inhalt"))
	})
}

// ausDemFeldRoh gibt ein Feld einer JSON-Antwort als rohes JSON.
func ausDemFeldRoh(t *testing.T, roh, feld string) string {
	t.Helper()

	var umschlag map[string]json.RawMessage
	if err := json.Unmarshal([]byte(roh), &umschlag); err != nil {
		t.Fatalf("die Antwort: %v (%s)", err, roh)
	}

	return string(umschlag[feld])
}
