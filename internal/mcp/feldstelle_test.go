// Die Rücknahme über den MCP-Server — vollständig, nicht halb (#160).
package mcp_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **`schreiben.feld` nennt die Stelle, an der das Feld stand** — und nimmt sie
// zurück.
//
// # Warum beides zusammengehört
//
// Eine Antwort, die eine Zahl nennt, die niemand zurückgeben kann, ist eine
// Zusage ohne Deckung. Erst mit dem Eingang wird sie zu einer Rücknahme, die
// **vollständig** ist: das Feld kommt an seine Stelle zurück und nicht ans Ende
// (#160).
func TestSchreibenFeldNenntUndNimmtDieStelle(t *testing.T) {
	bestand := &bestandsprobe{karten: []*domain.CardRow{{
		ID: 1, Number: 1, Values: []domain.Value{
			{Field: "name", Text: "Feuerball"},
			{Field: "farbe", Text: "rot"},
			{Field: "text", Text: "Drei Schaden"},
		},
	}}}

	werkzeug := werkzeugMitNamen(t, bestand, "schreiben.feld")

	// **Das mittlere Feld**, sonst wäre „ans Ende" dieselbe Stelle.
	antwort := rufeWerkzeug(t, werkzeug,
		`{"setID": 7, "karteID": 1, "feld": "farbe", "entfernen": true}`)

	if antwort["reihenfolge"] != float64(1) {
		t.Fatalf("Reihenfolge %v, erwartet 1 — %v", antwort["reihenfolge"], antwort)
	}

	hinweis, istText := antwort["zuruecknahme"].(string)
	if !istText || !strings.Contains(hinweis, "reihenfolge") {
		t.Errorf("die Zurücknahme nennt die Stelle nicht: %v", antwort["zuruecknahme"])
	}

	// Und zurück an ihre Stelle.
	rufeWerkzeug(t, werkzeug,
		`{"setID": 7, "karteID": 1, "feld": "farbe", "wert": "rot", "reihenfolge": 1}`)

	namen := make([]string, 0, len(bestand.karten[0].Values))
	for _, wert := range bestand.karten[0].Values {
		namen = append(namen, wert.Field)
	}

	if strings.Join(namen, ",") != "name,farbe,text" {
		t.Errorf("Reihenfolge %v", namen)
	}
}

// **Ein Feld, das es nicht gab, nennt keine Stelle.**
//
// `null` im JSON hieße „hatte keine" und läse sich wie „unbekannt"; wer den
// Schlüssel weglässt, sagt dasselbe ohne die zweite Lesart.
func TestEinFeldOhneStelleNenntKeine(t *testing.T) {
	bestand := &bestandsprobe{karten: []*domain.CardRow{{
		ID: 1, Number: 1, Values: []domain.Value{{Field: "name", Text: "Feuerball"}},
	}}}

	antwort := rufeWerkzeug(t, werkzeugMitNamen(t, bestand, "schreiben.feld"),
		`{"setID": 7, "karteID": 1, "feld": "gibtsnicht", "entfernen": true}`)

	if _, da := antwort["reihenfolge"]; da {
		t.Errorf("eine Stelle für ein Feld, das es nicht gab: %v", antwort)
	}
}

// werkzeugMitNamen sucht ein Werkzeug in der Liste.
func werkzeugMitNamen(t *testing.T, bestand *bestandsprobe, name string) mcp.Werkzeug {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(probenaehte(bestand)) {
		if werkzeug.Name == name {
			return werkzeug
		}
	}

	t.Fatalf("kein Werkzeug %q", name)

	return mcp.Werkzeug{}
}

// rufeWerkzeug ruft ein Werkzeug und zerlegt seine Antwort.
func rufeWerkzeug(t *testing.T, werkzeug mcp.Werkzeug, argumente string) map[string]any {
	t.Helper()

	roh, err := werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
	if err != nil {
		t.Fatalf("rufen: %v", err)
	}

	var antwort map[string]any
	if fehler := json.Unmarshal([]byte(roh), &antwort); fehler != nil {
		t.Fatalf("die Antwort ist kein JSON: %v — %s", fehler, roh)
	}

	return antwort
}
