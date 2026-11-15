package mcp_test

import (
	"strings"
	"testing"
)

// **Der Sitzungsschalter des Gerätemodells** — AI-05 (#2042): der Stand
// kommt zurück, ein Fehler der Werkstatt ist einer, und eine unlesbare
// Eingabe wird abgewiesen, bevor jemand gefragt wird.
func TestDerSitzungsschalterAmWerkzeug(t *testing.T) {
	t.Parallel()

	antwort, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.ki.geraetemodell.sitzung", `{"an": false}`)
	if err != nil || !strings.Contains(antwort, `"an": false`) {
		t.Errorf("%s %v", antwort, err)
	}

	if _, err := rufeSammlungswerkzeug(t, nilprobe{}, stolpernderPreis{},
		"schreiben.ki.geraetemodell.sitzung", `{"an": false}`); err == nil {
		t.Error("der Fehler der Werkstatt kam nicht an")
	}

	if _, err := rufeSammlungswerkzeug(t, nilprobe{}, neuePreisprobe(),
		"schreiben.ki.geraetemodell.sitzung", `{"an": "vielleicht"}`); err == nil {
		t.Error("eine unlesbare Eingabe ging durch")
	}
}
