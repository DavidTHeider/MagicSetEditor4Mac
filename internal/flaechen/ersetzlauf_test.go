// **Suchen und Ersetzen über jede Fläche ohne Fenster** — Werkstück 2 von
// Phase 3.7
// (#1168).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// **Zeigen, ersetzen, zurücknehmen — über jede Fläche.**
//
// # Was hier die Zusage ist
//
// Drei Dinge, und das mittlere ist das, wofür es das Werkstück gibt:
//
//  1. Die **Vorschau ändert nichts** — danach steht der alte Wert noch da.
//  2. Das Ausführen ändert **jede** Karte, die die Vorschau genannt hat.
//  3. **Ein** ⌘Z nimmt den ganzen Zug zurück, nicht eine Karte davon.
func TestErsetzenUeberJedeFlaecheOhneFenster(t *testing.T) {
	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := zweiFliegendeKarten(t, umgebung)

		schau := mussKommandozeile(t, umgebung,
			"ersetzen", setID, "Fliegend", "Rennend", "--json")
		if treffer := trefferAus(t, schau); treffer != 2 {
			t.Fatalf("%d Treffer in der Vorschau", treffer)
		}

		if namen := kartennamen(t, umgebung, setID); namen[0] != "Fliegend hoch" {
			t.Fatalf("die Vorschau hat geschrieben: %v", namen)
		}

		mussKommandozeile(t, umgebung, "ersetzen", setID, "Fliegend", "Rennend", "--tun")

		if namen := kartennamen(t, umgebung, setID); namen[0] != "Rennend hoch" ||
			namen[1] != "Rennend tief" {
			t.Fatalf("nach dem Ersetzen %v", namen)
		}

		// **Ein Zug zurück, und beide sind wieder da.**
		mussKommandozeile(t, umgebung, "zuruecknehmen", setID)

		if namen := kartennamen(t, umgebung, setID); namen[0] != "Fliegend hoch" ||
			namen[1] != "Fliegend tief" {
			t.Errorf("nach der Rücknahme %v", namen)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := zweiFliegendeKarten(t, umgebung)

		argumente := map[string]any{
			"setID": zahl(t, setID), "text": "Fliegend", "ersatz": "Rennend",
		}

		schau, istFehler := mcpRuf(t, umgebung, "lesen.ersetzen", argumente)
		if istFehler || trefferAus(t, schau) != 2 {
			t.Fatalf("lesen.ersetzen: %s", schau)
		}

		if namen := kartennamen(t, umgebung, setID); namen[0] != "Fliegend hoch" {
			t.Fatalf("die Vorschau hat geschrieben: %v", namen)
		}

		if text, fehlerhaft := mcpRuf(
			t, umgebung, "schreiben.ersetzen", argumente); fehlerhaft {
			t.Fatalf("schreiben.ersetzen: %s", text)
		}

		beideErsetztUndZurueck(t, umgebung, setID, func() {
			if text, fehlerhaft := mcpRuf(t, umgebung, "schreiben.zuruecknehmen",
				map[string]any{"setID": zahl(t, setID)}); fehlerhaft {
				t.Fatalf("schreiben.zuruecknehmen: %s", text)
			}
		})
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := zweiFliegendeKarten(t, umgebung)
		an := starteDienst(t, umgebung)

		rumpf := `{"text":"Fliegend","ersatz":"Rennend"}`

		status, schau := an.ruf(t, http.MethodPost,
			"/sets/"+setID+"/ersetzung/vorschau", rumpf)
		if status != http.StatusOK || trefferAus(t, schau) != 2 {
			t.Fatalf("POST vorschau: %d (%s)", status, schau)
		}

		if namen := kartennamen(t, umgebung, setID); namen[0] != "Fliegend hoch" {
			t.Fatalf("die Vorschau hat geschrieben: %v", namen)
		}

		status, getan := an.ruf(t, http.MethodPost, "/sets/"+setID+"/ersetzung", rumpf)
		if status != http.StatusOK {
			t.Fatalf("POST ersetzung: %d (%s)", status, getan)
		}

		beideErsetztUndZurueck(t, umgebung, setID, func() {
			if status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/zuruecknehmen", ""); status != http.StatusOK {
				t.Fatalf("POST zuruecknehmen: %d (%s)", status, rumpf)
			}
		})
	})
}

// beideErsetztUndZurueck prüft, was der Fall der Kommandozeile prüft: **beide**
// Karten ersetzt, und **ein** Zug zurück über dieselbe Fläche bringt beide
// wieder. MCP und Dienst sahen hier nur die zweite Karte und keine Rücknahme
// (Zweitblick über Phase C, Befund 3).
func beideErsetztUndZurueck(t *testing.T, umgebung []string, setID string, zuruecknehmen func()) {
	t.Helper()

	if namen := kartennamen(t, umgebung, setID); namen[0] != "Rennend hoch" || namen[1] != "Rennend tief" {
		t.Fatalf("nach dem Ersetzen %v", namen)
	}

	zuruecknehmen()

	if namen := kartennamen(t, umgebung, setID); namen[0] != "Fliegend hoch" || namen[1] != "Fliegend tief" {
		t.Errorf("nach der Rücknahme %v", namen)
	}
}

// **Ein Wahlfeld bleibt außen vor — an jeder Fläche.**
//
// # Warum gerade dieser Fall
//
// Er ist die **Grenze** des Ersetzens: `TextValueEditor` ist der einzige
// Editor, der `search` überschreibt (`src/gui/value/editor.hpp`, Zeile 115,
// Fassung f9d9356 vom 2021-07-06). Ginge eine Ersetzung durch ein
// `choice:`-Feld, entstünde eine Wahl, die die Vorlage nicht kennt — und die
// Karte zeichnete nichts mehr.
func TestEinWahlfeldBleibtBeimErsetzenAussenVor(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := zweiFliegendeKarten(t, umgebung)

	// Die Saat führt `art` als `choice:` — der Wert steht auf einer der Wahlen.
	karten := kartenzeilenAus(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))
	mussKommandozeile(t, umgebung, "feld", setID, karten[0].ID, "art", "Kreatur")

	schau := mussKommandozeile(t, umgebung,
		"ersetzen", setID, "Kreatur", "Artefakt", "--json")
	if treffer := trefferAus(t, schau); treffer != 0 {
		t.Fatalf("%d Treffer in einem Wahlfeld:\n%s", treffer, schau)
	}
}

// **Eine Frage ohne Suchtext wird überall abgewiesen.**
func TestEineLeereErsetzungWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := zweiFliegendeKarten(t, umgebung)

	t.Run("Kommandozeile", func(t *testing.T) {
		if aus, _, code := kommandozeile(
			t, umgebung, "ersetzen", setID, "  ", "X"); code == 0 {
			t.Fatalf("die leere Frage kam durch: %s", aus)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "lesen.ersetzen",
			map[string]any{"setID": zahl(t, setID), "text": "  ", "ersatz": "X"})
		if !istFehler {
			t.Fatalf("die leere Frage kam durch: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/ersetzung",
			`{"text":"  ","ersatz":"X"}`)
		if status == http.StatusOK {
			t.Fatalf("die leere Frage kam durch: %s", rumpf)
		}
	})
}

// zweiFliegendeKarten legt ein Set mit zwei Karten an, die beide »Fliegend«
// im Namen tragen.
func zweiFliegendeKarten(t *testing.T, umgebung []string) string {
	t.Helper()

	setID := leeresSet(t, umgebung, "Zum Ersetzen")

	for _, name := range []string{"Fliegend hoch", "Fliegend tief"} {
		mussKommandozeile(t, umgebung, "karte", setID, "anlegen")

		karten := kartenzeilenAus(t,
			mussKommandozeile(t, umgebung, "karten", setID, "--json"))

		mussKommandozeile(t, umgebung, "feld", setID,
			karten[len(karten)-1].ID, "name", name)
	}

	return setID
}

// kartennamen liest die Namen der Karten eines Sets, in ihrer Reihenfolge.
func kartennamen(t *testing.T, umgebung []string, setID string) []string {
	t.Helper()

	var antwort struct {
		Karten []struct {
			Werte map[string]string `json:"werte"`
		} `json:"karten"`
	}

	roh := mussKommandozeile(t, umgebung, "karten", setID, "--json")
	if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	namen := make([]string, 0, len(antwort.Karten))
	for _, karte := range antwort.Karten {
		namen = append(namen, karte.Werte["name"])
	}

	return namen
}

// trefferAus liest `treffer` aus einer Ersetzungsantwort.
func trefferAus(t *testing.T, roh string) int {
	t.Helper()

	var antwort struct {
		Treffer int `json:"treffer"`
	}

	if err := json.Unmarshal([]byte(strings.TrimSpace(roh)), &antwort); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	return antwort.Treffer
}
