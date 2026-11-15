package flaechen_test

// **Das Symbol eines fremden Sets ist an keiner Fläche zu erreichen** —
// #1815.
//
// # Der Befund
//
// Die Symbolzeichnung eines Sets liegt als Datei neben ihm, und die Werkstatt
// fasste sie über die Setkennung an, **ohne den Bestand zu fragen**: wer die
// Kennung eines fremden Sets nannte, las dessen Symbol oder legte eines
// hinein — und das reist mit dem Set, wenn sein Eigner es sichert. MCP und
// Dienst riefen die Werkstatt dazu mit einem Kontext **ohne** Profil.
//
// # Was er zeigt
//
// Der erste Nutzer legt ein Symbol in sein Set; Bob, im eigenen Profil, liest
// und schreibt es an Kommandozeile, MCP und Dienst nicht — „nicht gefunden",
// obwohl die Datei dasteht. Die Gegenprobe steht davor: der Eigner schreibt
// und liest.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// symbolteile ist eine kleine, gültige Zeichnung.
const symbolteile = `[{"art":"Form","name":"Probe","kombination":"merge","punkte":[` +
	`{"x":0,"y":0},{"x":1,"y":0},{"x":1,"y":1}]}]`

func TestDasSymbolEinesFremdenSetsTrifftNichts(t *testing.T) {
	umgebung, setID, _ := setMitKarten(t)

	// **Die Gegenprobe zuerst**: der Eigner legt ab und liest zurück — ohne
	// sie hieße ein „nicht gefunden" unten nur, dass es kein Symbol gibt.
	abgelegt, fehler, code := mitEingabeUndCode(t, umgebung, symbolteile, "symbol-setzen", setID, "--json")
	if code != 0 {
		t.Fatalf("der Eigner legt sein Symbol nicht ab: %d (%s)", code, fehler)
	}

	var datei struct {
		Name string `json:"name"`
	}
	if fehler := json.Unmarshal([]byte(abgelegt), &datei); fehler != nil || datei.Name == "" {
		t.Fatalf("kein Name in %q (%v)", abgelegt, fehler)
	}

	if aus := mussKommandozeile(t, umgebung, "symbol", setID, datei.Name); !strings.Contains(aus, "Probe") {
		t.Fatalf("der Eigner liest sein Symbol nicht: %s", aus)
	}

	mussKommandozeile(t, umgebung, "nutzer-neu", "Bob")
	mussKommandozeile(t, umgebung, "nutzer-wechseln", "2")

	t.Run("Kommandozeile", func(t *testing.T) {
		for name, ruf := range map[string]func() (string, int){
			"lesen": func() (string, int) {
				_, fehler, code := mitEingabeUndCode(t, umgebung, "", "symbol", setID, datei.Name)
				return fehler, code
			},
			"schreiben": func() (string, int) {
				_, fehler, code := mitEingabeUndCode(t, umgebung, symbolteile, "symbol-setzen", setID)
				return fehler, code
			},
			"nachzeichnen": func() (string, int) {
				_, fehler, code := mitEingabeUndCode(t, umgebung, "", "symbol-nachzeichnen", setID, kreisbild(t))
				return fehler, code
			},
		} {
			if fehler, code := ruf(); code == 0 || !strings.Contains(fehler, "nicht gefunden") {
				t.Errorf("Bob beim %s: %d (%s), erwartet „nicht gefunden“", name, code, fehler)
			}
		}
	})

	t.Run("MCP", func(t *testing.T) {
		for werkzeug, argumente := range map[string]map[string]any{
			"lesen.symbol":                 {"setID": zahl(t, setID), "name": datei.Name},
			"schreiben.symbol":             {"setID": zahl(t, setID), "teile": rohTeile(t)},
			"schreiben.symbolnachzeichnen": {"setID": zahl(t, setID), "bildpfad": kreisbild(t)},
		} {
			if text, istFehler := mcpRuf(t, umgebung, werkzeug, argumente); !istFehler ||
				!strings.Contains(text, "nicht gefunden") {
				t.Errorf("%s: %s, erwartet „nicht gefunden“", werkzeug, text)
			}
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		for methode, rumpf := range map[string]string{
			http.MethodGet: "",
			http.MethodPut: `{"teile":` + symbolteile + `,"unter":""}`,
		} {
			if status, antwort := an.ruf(t, methode, "/sets/"+setID+"/symbol?name="+datei.Name, rumpf); status != http.StatusNotFound {
				t.Errorf("%s: %d (%s), erwartet 404", methode, status, antwort)
			}
		}

		bild, fehler := json.Marshal(map[string]string{"bildpfad": kreisbild(t)})
		if fehler != nil {
			t.Fatalf("den Rumpf bauen: %v", fehler)
		}

		if status, antwort := an.ruf(t, http.MethodPost, "/sets/"+setID+"/symbol-nachzeichnen", string(bild)); status != http.StatusNotFound {
			t.Errorf("nachzeichnen: %d (%s), erwartet 404", status, antwort)
		}
	})
}

// rohTeile gibt die Zeichnung als Liste, wie MCP sie im Argument erwartet.
func rohTeile(t *testing.T) []any {
	t.Helper()

	var teile []any
	if fehler := json.Unmarshal([]byte(symbolteile), &teile); fehler != nil {
		t.Fatalf("die Zeichnung lesen: %v", fehler)
	}

	return teile
}
