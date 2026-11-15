package mcp_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// entpacke liest eine Antwortzeile.
func entpacke(t *testing.T, zeile string) map[string]any {
	t.Helper()

	var antwort map[string]any
	if err := json.Unmarshal([]byte(zeile), &antwort); err != nil {
		t.Fatalf("Antwort ist kein JSON: %v\n%s", err, zeile)
	}

	return antwort
}

// Der Server stellt sich vor und nennt die Protokollfassung.
func TestDerServerStelltSichVor(t *testing.T) {
	zeilen := sprich(t, probenbestand(), `{"jsonrpc":"2.0","id":1,"method":"initialize"}`)

	if len(zeilen) != 1 {
		t.Fatalf("%d Antworten, erwartet eine", len(zeilen))
	}

	ergebnis := teilAntwort(t, zeilen[0])
	if ergebnis["protocolVersion"] != "2025-06-18" {
		t.Errorf("Fassung: %v", ergebnis["protocolVersion"])
	}

	// **Der Hinweis steht in der Vorstellung**, bevor ein Gegenüber das erste
	// Werkzeug sieht.
	anweisung := zeichenkette(t, ergebnis["instructions"])
	if !strings.Contains(anweisung, "schreiben.") {
		t.Errorf("die Vorstellung sagt nichts über die getrennten Reihen: %q", anweisung)
	}
}

// **Eine Meldung bekommt keine Antwort.**
//
// Sie hat keine Kennung; wer ihr trotzdem antwortet, schickt eine Antwort auf
// nichts — manche Gegenstellen brechen daran ab.
func TestEineMeldungBekommtKeineAntwort(t *testing.T) {
	zeilen := sprich(t, probenbestand(),
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":null,"method":"irgendwas"}`)

	if len(zeilen) != 0 {
		t.Errorf("Antworten: %v, erwartet keine", zeilen)
	}
}

// Eine Kennung kommt in der Gestalt zurück, in der sie kam.
//
// JSON-RPC erlaubt Zahlen **und** Zeichenketten. Wer sie in einen Typ zwingt,
// antwortet dem einen Gegenüber mit `"1"`, wo es `1` geschickt hat.
func TestDieKennungKommtInIhrerGestaltZurueck(t *testing.T) {
	faelle := map[string]string{
		"eine Zahl":         `1`,
		"eine Zeichenkette": `"abc"`,
	}

	for name, kennung := range faelle {
		t.Run(name, func(t *testing.T) {
			zeilen := sprich(t, probenbestand(),
				`{"jsonrpc":"2.0","id":`+kennung+`,"method":"ping"}`)

			if len(zeilen) != 1 {
				t.Fatalf("%d Antworten", len(zeilen))
			}

			if !strings.Contains(zeilen[0], `"id":`+kennung) {
				t.Errorf("Antwort: %s", zeilen[0])
			}
		})
	}
}

// Was der Server nicht kann, sagt er.
func TestUnbekanntesWirdBenannt(t *testing.T) {
	faelle := map[string]struct {
		zeile    string
		enthaelt string
	}{
		"kein JSON": {
			zeile: `{ kaputt`, enthaelt: "kein JSON",
		},
		"eine unbekannte Methode": {
			zeile:    `{"jsonrpc":"2.0","id":1,"method":"gibtsnicht"}`,
			enthaelt: "unbekannte Methode",
		},
		"ein unbekanntes Werkzeug": {
			zeile:    `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"gibtsnicht"}}`,
			enthaelt: "unbekanntes Werkzeug",
		},
		"Argumente, die kein JSON sind": {
			zeile:    `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":"kein Objekt"}`,
			enthaelt: "kein JSON",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			zeilen := sprich(t, probenbestand(), fall.zeile)

			if len(zeilen) != 1 {
				t.Fatalf("%d Antworten", len(zeilen))
			}

			if !strings.Contains(zeilen[0], fall.enthaelt) {
				t.Errorf("Antwort: %s", zeilen[0])
			}
		})
	}
}

// **Ein fachlicher Fehler ist kein Protokollfehler.**
//
// MCP will ihn im Ergebnis mit `isError`, damit das Gegenüber ihn dem Modell
// zeigen kann — statt die Verbindung für kaputt zu halten.
func TestEinFachlicherFehlerKommtAlsErgebnis(t *testing.T) {
	bestand := probenbestand()
	bestand.fehler = errors.New("die Platte ist voll")

	zeilen := sprich(t, bestand,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lesen.sets","arguments":{}}}`)

	antwort := entpacke(t, zeilen[0])
	if antwort["error"] != nil {
		t.Errorf("als Protokollfehler geantwortet: %v", antwort["error"])
	}

	ergebnis := abbildung(t, antwort["result"])
	if ergebnis["isError"] != true {
		t.Errorf("Ergebnis: %v — erwartet war isError", ergebnis)
	}
}

// Ein Strom, der endet, ist kein Fehler.
func TestEinEndenderStromIstKeinFehler(t *testing.T) {
	if zeilen := sprich(t, probenbestand()); len(zeilen) != 0 {
		t.Errorf("Antworten: %v", zeilen)
	}
}

// abbildung liest einen Wert als Abbildung — oder bricht ab.
//
// **Mit Abbruch statt mit Nullwert**: eine Zusicherung, die still misslingt,
// macht aus einem Formfehler eine leere Abbildung, und der Test prüft danach
// nichts mehr.
func abbildung(t *testing.T, wert any) map[string]any {
	t.Helper()

	gefunden, ist := wert.(map[string]any)
	if !ist {
		t.Fatalf("%v ist keine Abbildung", wert)
	}

	return gefunden
}

// zeichenkette liest einen Wert als Zeichenkette — oder bricht ab.
func zeichenkette(t *testing.T, wert any) string {
	t.Helper()

	gefunden, ist := wert.(string)
	if !ist {
		t.Fatalf("%v ist keine Zeichenkette", wert)
	}

	return gefunden
}

// folge liest einen Wert als Liste — oder bricht ab.
func folge(t *testing.T, wert any) []any {
	t.Helper()

	gefunden, ist := wert.([]any)
	if !ist {
		t.Fatalf("%v ist keine Liste", wert)
	}

	return gefunden
}

// teilAntwort liest das Ergebnis einer Antwortzeile.
func teilAntwort(t *testing.T, zeile string) map[string]any {
	t.Helper()

	return abbildung(t, entpacke(t, zeile)["result"])
}
