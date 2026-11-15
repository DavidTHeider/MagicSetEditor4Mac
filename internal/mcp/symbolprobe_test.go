// **Die Probe des Symboleditors** — Werkstück
// #1391.
//
// # Sie rechnet nicht, und das ist keine Wahl
//
// Der erste Anlauf rief `app.SymbolRechnen` — und der Schichtenwächter hat
// ihn abgewiesen: `internal/mcp` ist ein **Blatt** (§2) und kennt
// `internal/domain` und sonst nichts aus diesem Projekt. **Auch nicht in
// einem Test**, und das ist richtig so: eine Unabhängigkeit, die nur für den
// Produktivcode gilt, ist keine.
//
// # Was diese Probe deshalb belegt — und was nicht
//
// **Sie belegt die Verdrahtung**: dass ein Werkzeug seine Argumente liest,
// die Naht ruft und ihre Antwort weiterreicht. Sie gibt die Argumente
// zurück, statt eine feste Antwort zu geben — damit ein Werkzeug, das seine
// Argumente fallen lässt, auffällt.
//
// **Sie belegt nicht, dass die Rechnung stimmt.** Das tut
// `internal/flaechen/symbol_test.go`: dort laufen die **gebauten** Programme,
// und dieselbe Kette geht über alle drei Flächen. Genau dafür gibt es diese
// Läufe (§16).
package mcp_test

import (
	"context"
	"encoding/json"
)

// symbolprobe bedient die Symbolwerkzeuge im Test.
type symbolprobe struct {
	// abgelegt hält fest, was `schreiben.symbol` weiterreichte.
	abgelegt string

	// fehler lässt **jeden** Weg scheitern.
	//
	// **Damit die Fehlerwege der Werkzeuge betreten werden.** Sie sind je
	// zwei Zeilen und sehen aus wie eine Formalie; ein Zweig, den kein Lauf
	// begeht, ist eine Behauptung darüber, was im Ernstfall geschieht (§3).
	fehler error
}

// SymbolLesen gibt ein festes Teil.
func (s *symbolprobe) SymbolLesen(_ context.Context, setID int64, name string) (any, error) {
	if s.fehler != nil {
		return nil, s.fehler
	}

	return []map[string]any{{"art": "Form", "name": name, "setID": setID}}, nil
}

// SymbolAblegen nimmt die Teile entgegen und gibt einen Namen.
func (s *symbolprobe) SymbolAblegen(_ context.Context, _ int64, teile json.RawMessage, unter string) (string, error) {
	if s.fehler != nil {
		return "", s.fehler
	}

	s.abgelegt = string(teile)

	if unter != "" {
		return unter, nil
	}

	return "symbol.mse-symbol", nil
}

// SymbolRechnen gibt zurück, was hereinkam.
//
// **Als Echo und nicht als feste Antwort.** Ein Werkzeug, das seine
// Argumente nicht weiterreicht, gäbe hier ein leeres Ergebnis — und der
// Unterschied ist zu sehen. Eine feste Antwort machte jedes Werkzeug gleich.
func (s *symbolprobe) SymbolRechnen(was string, argumente json.RawMessage) (any, error) {
	if s.fehler != nil {
		return nil, s.fehler
	}

	// **Sie liest die Argumente wirklich.** Der echte Rechner tut es auch,
	// und sein Fehlerweg ist der einzige, über den ein Symbolwerkzeug
	// scheitern kann — eine Probe, die jedes JSON annimmt, ließe ihn
	// unbetreten.
	var gelesen map[string]any
	if err := json.Unmarshal(argumente, &gelesen); err != nil {
		return nil, err
	}

	return map[string]any{"was": was, "argumente": gelesen}, nil
}

// Palette gibt zwei Einträge.
func (s *symbolprobe) Palette(_ context.Context, _ int64, feld string) (any, error) {
	if s.fehler != nil {
		return nil, s.fehler
	}

	return []map[string]any{
		{"art": "code", "code": "W", "beschriftung": feld},
		{"art": "line"},
	}, nil
}
