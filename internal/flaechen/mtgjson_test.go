// **MTG JSON an jeder Fläche** —
// #1291.
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mtgjsonprobe gibt den Pfad zu einer eingecheckten MTG-JSON-Datei.
func mtgjsonprobe(t *testing.T) string {
	t.Helper()

	// **Aus dem Vorrat**, den der Dienst annimmt (C06) — siehe [vorrat].
	pfad := filepath.Join(vorrat, "testdaten", "mtgjson", "probe.json")
	if _, err := os.Stat(pfad); err != nil {
		t.Fatalf("die Probedatei fehlt: %v", err)
	}

	return pfad
}

// **Wer über die Kommandozeile, MCP oder REST MTG JSON liest, füllt den Bestand.**
//
// # Der Befund
//
// Zwei Zeilen von `docs/FAEHIGKEITEN.md` standen so da:
//
//	| MTG JSON lesen     | `MTGJSONOeffnen`, … | — | — | — | ✋ |
//	| MTG JSON schreiben | `MTGJSONSchreiben`  | — | — | — | ✋ |
//
// Die erhobene Hand meint einen Dateidialog, und der steckt allein in der
// Bindung. **Gerade dieses Format ist das, was ein Skript braucht**: MTG JSON
// ist der Weg zwischen diesem Programm und allem anderen
// ([0009](../../docs/entscheidungen/0009-phase-9-fremdformate-und-schnittstelle.md)).
func TestMTGJSONWirdAnJederFlaecheGelesen(t *testing.T) {
	quelle := mtgjsonprobe(t)

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)

		aus, fehler, code := kommandozeile(t, umgebung,
			"mtgjson-lesen", quelle, "magic", "--json")
		if code != 0 {
			t.Fatalf("lesen: %d (%s)", code, fehler)
		}

		nenntEinSet(t, "die Kommandozeile", aus)
		bestandHatEinSet(t, umgebung, "die Kommandozeile")
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)

		text, istFehler := mcpRuf(t, umgebung, "schreiben.mtgjson.lesen",
			map[string]any{"pfad": quelle, "spiel": "magic"})
		if istFehler {
			t.Fatalf("lesen: %s", text)
		}

		nenntEinSet(t, "MCP", text)
		bestandHatEinSet(t, umgebung, "MCP")
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/mtgjson",
			`{"pfad":`+alsJSONText(t, quelle)+`,"spiel":"magic"}`)
		if status != http.StatusCreated {
			t.Fatalf("POST: %d (%s)", status, rumpf)
		}

		nenntEinSet(t, "REST", rumpf)
		bestandHatEinSet(t, umgebung, "REST")
	})
}

// **Und wer schreibt, bekommt MTG JSON zurück — mit seinem Umschlag.**
func TestMTGJSONWirdAnJederFlaecheGeschrieben(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	ziel := t.TempDir()

	setID := setkennungAus(t,
		mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))

	t.Run("Kommandozeile", func(t *testing.T) {
		pfad := filepath.Join(ziel, "cli.json")

		aus, fehler, code := kommandozeile(t, umgebung,
			"mtgjson-schreiben", setID, "magic", pfad)
		if code != 0 {
			t.Fatalf("schreiben: %d (%s)", code, fehler)
		}

		istMTGJSON(t, "die Kommandozeile", pfad, aus)
	})

	t.Run("MCP", func(t *testing.T) {
		pfad := filepath.Join(ziel, "mcp.json")

		text, istFehler := mcpRuf(t, umgebung, "schreiben.mtgjson.schreiben",
			map[string]any{"setID": zahl(t, setID), "spiel": "magic", "ziel": pfad})
		if istFehler {
			t.Fatalf("schreiben: %s", text)
		}

		istMTGJSON(t, "MCP", pfad, text)
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)
		pfad := filepath.Join(ziel, "rest.json")

		status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/mtgjson",
			`{"spiel":"magic","ziel":`+alsJSONText(t, pfad)+`}`)
		if status != http.StatusCreated {
			t.Fatalf("POST: %d (%s)", status, rumpf)
		}

		istMTGJSON(t, "REST", pfad, rumpf)
	})
}

// **Und dieselbe fachliche Abweisung: ein Spiel ohne Zuordnung** (§16).
func TestEinSpielOhneZuordnungWirdAnJederFlaecheAbgewiesen(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	quelle := mtgjsonprobe(t)

	t.Run("Kommandozeile", func(t *testing.T) {
		if _, _, code := kommandozeile(t, umgebung,
			"mtgjson-lesen", quelle, "kein-spiel"); code == 0 {
			t.Error("die Kommandozeile hat es hingenommen")
		}
	})

	t.Run("MCP", func(t *testing.T) {
		if text, istFehler := mcpRuf(t, umgebung, "schreiben.mtgjson.lesen",
			map[string]any{"pfad": quelle, "spiel": "kein-spiel"}); !istFehler {
			t.Errorf("MCP hat es hingenommen: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/mtgjson",
			`{"pfad":`+alsJSONText(t, quelle)+`,"spiel":"kein-spiel"}`)
		if status < 400 {
			t.Errorf("REST hat es hingenommen: %d (%s)", status, rumpf)
		}
	})
}

// nenntEinSet prüft, dass die Antwort eine Setkennung führt.
func nenntEinSet(t *testing.T, wer, antwort string) {
	t.Helper()

	if !strings.Contains(antwort, "setID") && !strings.Contains(antwort, "angelegt") {
		t.Errorf("%s nennt kein angelegtes Set: %s", wer, antwort)
	}
}

// bestandHatEinSet sieht nach, dass wirklich etwas im Bestand steht.
//
// **Zwei grüne Seiten einer Grenze sind kein Nachweis für die Grenze**: eine
// Antwort mit einer Kennung sagt nicht, dass das Set danach da ist.
func bestandHatEinSet(t *testing.T, umgebung []string, wer string) {
	t.Helper()

	if liste := mussKommandozeile(t, umgebung, "sets"); strings.TrimSpace(liste) == "" {
		t.Errorf("nach %s ist der Bestand leer", wer)
	}
}

// istMTGJSON sieht nach, dass am Pfad MTG JSON liegt — mit seinem Umschlag.
func istMTGJSON(t *testing.T, wer, pfad, antwort string) {
	t.Helper()

	roh, err := os.ReadFile(pfad) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("%s hat keine Datei hinterlassen: %v", wer, err)
	}

	// **Der Umschlag ist das, woran ein anderes Werkzeug es erkennt** — eine
	// Datei mit Inhalt sähe von außen genauso aus.
	var geschrieben struct {
		Meta map[string]any `json:"meta"`
		Data map[string]any `json:"data"`
	}

	if err := json.Unmarshal(roh, &geschrieben); err != nil {
		t.Fatalf("%s hat kein JSON geschrieben: %v", wer, err)
	}

	if len(geschrieben.Meta) == 0 || len(geschrieben.Data) == 0 {
		t.Errorf("%s: der Umschlag fehlt: %s", wer, roh[:min(120, len(roh))])
	}

	if !strings.Contains(antwort, pfad) {
		t.Errorf("%s nennt den Pfad nicht: %s", wer, antwort)
	}
}
