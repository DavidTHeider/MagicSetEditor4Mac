package mcp_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Vorschau einer Massenänderung über MCP** — Werkstück U6 von Phase
// 5.1 (#1417): lesend, idempotent, und die Antwort nennt je Karte Alt, Neu
// und Ausschluss.
func TestDieVorschauNenntJeKarteAltNeuUndAusschluss(t *testing.T) {
	probe := probenbestand()
	werkzeug := werkzeugMitNamen(t, probe, "lesen.felder.vorschau")

	antwort, fehler := werkzeug.Ruf(t.Context(), json.RawMessage(
		`{"setID": 1, "karteIDs": [7, 8], "feld": "artist", "wert": "X"}`))
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	var gelesen struct {
		Feld   string `json:"feld"`
		Wert   string `json:"wert"`
		Zeilen []struct {
			KarteID    int64  `json:"karteID"`
			Name       string `json:"name"`
			Alt        string `json:"alt"`
			GabAlt     bool   `json:"gabAlt"`
			Neu        string `json:"neu"`
			Ausschluss string `json:"ausschluss"`
		} `json:"zeilen"`
	}
	if err := json.Unmarshal([]byte(antwort), &gelesen); err != nil {
		t.Fatalf("Antwort %q: %v", antwort, err)
	}

	if gelesen.Feld != "artist" || gelesen.Wert != "X" || len(gelesen.Zeilen) != 2 {
		t.Fatalf("Antwort: %+v", gelesen)
	}

	if gelesen.Zeilen[0].KarteID != 7 || gelesen.Zeilen[0].Alt != "Alt" || !gelesen.Zeilen[0].GabAlt ||
		gelesen.Zeilen[0].Neu != "X" || gelesen.Zeilen[0].Ausschluss != "" {
		t.Errorf("erste Zeile: %+v", gelesen.Zeilen[0])
	}

	if gelesen.Zeilen[1].KarteID != 8 || gelesen.Zeilen[1].Ausschluss != domain.AusschlussGleich {
		t.Errorf("zweite Zeile: %+v", gelesen.Zeilen[1])
	}

	if len(probe.massengesetzt) != 0 {
		t.Errorf("die Vorschau hat geschrieben: %v", probe.massengesetzt)
	}
}

// **`schreiben.felder` nimmt eine Erwartung an** — und geht dann den Weg,
// der vergleicht, bevor er schreibt.
func TestFelderSetzenMitErwartungGehtDenVergleichendenWeg(t *testing.T) {
	probe := probenbestand()
	werkzeug := werkzeugMitNamen(t, probe, "schreiben.felder")

	antwort, fehler := werkzeug.Ruf(t.Context(), json.RawMessage(
		`{"setID": 1, "karteIDs": [7, 8], "feld": "artist", "wert": "X",
		  "erwartet": [{"karteID": 7, "wert": "Alt", "gabEs": true}, {"karteID": 8, "gabEs": false}]}`))
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if !strings.Contains(antwort, `"vorgang8"`) {
		t.Errorf("Antwort %q nennt den Vorgang des vergleichenden Wegs nicht", antwort)
	}

	if len(probe.erwartet) != 2 || probe.erwartet[0] != (domain.Feldvorwert{KarteID: 7, Wert: "Alt", GabEs: true}) ||
		probe.erwartet[1] != (domain.Feldvorwert{KarteID: 8}) {
		t.Errorf("Erwartung: %+v", probe.erwartet)
	}
}

// **Eine Erwartung mit einer Kennung, die nicht in `karteIDs` steht, ist ein
// Bedienfehler** — sonst schriebe der Ruf Karten, die er gar nicht nennt.
func TestEineErwartungAusserhalbDerAuswahlWirdAbgewiesen(t *testing.T) {
	probe := probenbestand()
	werkzeug := werkzeugMitNamen(t, probe, "schreiben.felder")

	_, fehler := werkzeug.Ruf(t.Context(), json.RawMessage(
		`{"setID": 1, "karteIDs": [7], "feld": "artist", "wert": "X",
		  "erwartet": [{"karteID": 9, "wert": "Alt", "gabEs": true}]}`))
	if fehler == nil {
		t.Fatal("kein Fehler")
	}

	if len(probe.erwartet) != 0 || len(probe.massengesetzt) != 0 {
		t.Errorf("geschrieben wurde trotzdem: %v / %v", probe.erwartet, probe.massengesetzt)
	}
}
