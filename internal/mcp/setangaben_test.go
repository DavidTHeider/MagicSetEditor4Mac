// **Was die vier Setangabenwerkzeuge antworten — und wann sie abweisen**
// (#1107).
package mcp_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// **Die vier Werkzeuge reichen den Fehler ihrer Naht durch.**
//
// **Je Werkzeug und nicht einmal**: jedes hat sein eigenes `if fehler != nil`,
// und ein vergessenes gäbe eine leere Antwort statt einer Meldung.
func TestDieSetangabenwerkzeugeReichenIhrenFehlerDurch(t *testing.T) {
	faelle := map[string]string{
		"lesen.setangaben":    `{"setID": 1}`,
		"lesen.vorbehalte":    `{"setID": 1}`,
		"lesen.beispielsets":  `{}`,
		"schreiben.setangabe": `{"setID": 1, "feld": "title", "wert": "x"}`,
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			probe := neueAngabenprobe()
			probe.fehler = errors.New("die Naht ist zu")

			werkzeug := werkzeugMitAngaben(t, probe, name)

			if _, err := werkzeug.Ruf(t.Context(), json.RawMessage(argumente)); err == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}
}

// **Eine unbrauchbare Kennung wird abgewiesen, bevor die Naht sie sieht.**
func TestDieSetangabenwerkzeugeWeisenEineNullkennungAb(t *testing.T) {
	faelle := map[string]string{
		"lesen.setangaben":    `{"setID": 0}`,
		"lesen.vorbehalte":    `{"setID": 0}`,
		"schreiben.setangabe": `{"setID": 0, "feld": "title", "wert": "x"}`,
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			probe := neueAngabenprobe()
			werkzeug := werkzeugMitAngaben(t, probe, name)

			_, err := werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
			if err == nil {
				t.Fatal("die Null lief durch")
			}

			if !strings.Contains(err.Error(), "setID") {
				t.Errorf("die Meldung nennt das Feld nicht: %v", err)
			}
		})
	}
}

// **Ein Rumpf, der kein JSON ist, wird abgewiesen.**
func TestDasSetzenWeistUnlesbareArgumenteAb(t *testing.T) {
	werkzeug := werkzeugMitAngaben(t, neueAngabenprobe(), "schreiben.setangabe")

	if _, err := werkzeug.Ruf(t.Context(), json.RawMessage(`{`)); err == nil {
		t.Error("ein unlesbarer Rumpf lief durch")
	}
}

// **Und das Setzen nennt den vorigen Wert.**
func TestDasSetzenNenntDenVorigenWert(t *testing.T) {
	probe := neueAngabenprobe()
	werkzeug := werkzeugMitAngaben(t, probe, "schreiben.setangabe")

	aus, err := werkzeug.Ruf(t.Context(),
		json.RawMessage(`{"setID": 1, "feld": "title", "wert": "Neu"}`))
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	var antwort struct {
		Vorher string `json:"vorher"`
		Hatte  bool   `json:"hatte"`
	}

	if fehler := json.Unmarshal([]byte(aus), &antwort); fehler != nil {
		t.Fatalf("JSON lesen: %v (%s)", fehler, aus)
	}

	if !antwort.Hatte || antwort.Vorher != "Probeset" {
		t.Errorf("der vorige Wert lautet %+v statt »Probeset«", antwort)
	}
}
