// **Die Gestalt der Kartenliste über jede Fläche ohne Fenster** —
// Werkstück 3 von Phase 3.7
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

// **Wählen, ordnen — und die Wahl überlebt das Set.**
//
// # Was hier die Zusage ist
//
// Drei Dinge, und das dritte ist das, wofür es das Werkstück gibt:
//
//  1. Die **Spalten** lassen sich wählen, und die Antwort zeigt es.
//  2. Die **Ordnung** wirkt: nach `name` sortiert steht »Karte 1« vor
//     »Karte 10« — und nicht »Karte 10« vor »Karte 2«.
//  3. **Ein zweites Set desselben Spiels erbt die Wahl**, ohne dass jemand sie
//     dort noch einmal trifft.
func TestDieKartenlisteUeberJedeFlaecheOhneFenster(t *testing.T) {
	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := dreiNumerierteKarten(t, umgebung)

		mussKommandozeile(t, umgebung, "kartenliste", setID,
			"--spalten=name", "--sortieren=name")

		aufbau := listenaufbauAus(t,
			mussKommandozeile(t, umgebung, "kartenliste", setID, "--json"))

		if aufbau.Sortierfeld != "name" || !aufbau.Aufsteigend {
			t.Fatalf("Ordnung: %q / %v", aufbau.Sortierfeld, aufbau.Aufsteigend)
		}

		if namen := aufbau.Namen(); namen[0] != "Karte 1" ||
			namen[1] != "Karte 2" || namen[2] != "Karte 10" {
			t.Fatalf("Reihenfolge %v — »Karte 10« gehört hinter »Karte 2«", namen)
		}

		// **Und die Wahl überlebt das Set.**
		zweites := leeresSet(t, umgebung, "Zweites")

		geerbt := listenaufbauAus(t,
			mussKommandozeile(t, umgebung, "kartenliste", zweites, "--json"))
		if geerbt.Sortierfeld != "name" {
			t.Errorf("das zweite Set ordnet nach %q", geerbt.Sortierfeld)
		}

		// **Absteigend kehrt sie um.**
		mussKommandozeile(t, umgebung, "kartenliste", setID,
			"--sortieren=name", "--absteigend")

		danach := listenaufbauAus(t,
			mussKommandozeile(t, umgebung, "kartenliste", setID, "--json"))
		if namen := danach.Namen(); namen[0] != "Karte 10" {
			t.Errorf("absteigend steht %v", namen)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := dreiNumerierteKarten(t, umgebung)

		text, istFehler := mcpRuf(t, umgebung, "schreiben.kartenliste.sortierung",
			map[string]any{"setID": zahl(t, setID), "feld": "name", "aufsteigend": true})
		if istFehler {
			t.Fatalf("sortierung: %s", text)
		}

		gelesen, istFehler := mcpRuf(t, umgebung, "lesen.kartenliste",
			map[string]any{"setID": zahl(t, setID)})
		if istFehler {
			t.Fatalf("lesen.kartenliste: %s", gelesen)
		}

		// **Die Reihenfolge steht in der Antwort** — »Karte 1« vor »Karte 10«.
		if strings.Index(gelesen, "Karte 1\"") > strings.Index(gelesen, "Karte 10") {
			t.Errorf("die Ordnung wirkt nicht:\n%s", gelesen)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := dreiNumerierteKarten(t, umgebung)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPut,
			"/sets/"+setID+"/kartenliste/sortierung",
			`{"feld":"name","aufsteigend":true}`)
		if status != http.StatusOK {
			t.Fatalf("PUT sortierung: %d (%s)", status, rumpf)
		}

		status, gelesen := an.ruf(t, http.MethodGet, "/sets/"+setID+"/kartenliste", "")
		if status != http.StatusOK {
			t.Fatalf("GET kartenliste: %d (%s)", status, gelesen)
		}

		if !strings.Contains(gelesen, `"sortierfeld": "name"`) &&
			!strings.Contains(gelesen, `"sortierfeld":"name"`) {
			t.Errorf("die Antwort nennt die Ordnung nicht:\n%s", gelesen)
		}
	})
}

// **Ein Feld, das es nicht gibt, wird überall abgewiesen.**
//
// # Warum gerade dieser Fall
//
// Er ist die **Grenze** der Wahl: ein Tippfehler in einem Feldnamen legte
// sonst eine Spalte an, die niemals erscheint — und niemand sähe, warum
// (§8, §16 Regel 6).
func TestEinUnbekanntesListenfeldWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := dreiNumerierteKarten(t, umgebung)

	t.Run("MCP", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "schreiben.kartenliste.sortierung",
			map[string]any{"setID": zahl(t, setID), "feld": "gibtsnicht"})
		if !istFehler {
			t.Fatalf("das unbekannte Feld kam durch: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPut,
			"/sets/"+setID+"/kartenliste/sortierung", `{"feld":"gibtsnicht"}`)
		if status == http.StatusOK {
			t.Fatalf("das unbekannte Feld kam durch: %s", rumpf)
		}
	})
}

// listenaufbau ist die Antwort von `mse kartenliste --json`.
type listenaufbau struct {
	Spalten []struct {
		Feld     string `json:"feld"`
		Sichtbar bool   `json:"sichtbar"`
	} `json:"spalten"`

	Zeilen []struct {
		Werte map[string]string `json:"werte"`
	} `json:"zeilen"`

	Sortierfeld string `json:"sortierfeld"`
	Aufsteigend bool   `json:"aufsteigend"`
}

// Namen gibt die Kartennamen in der Reihenfolge der Liste.
func (a listenaufbau) Namen() []string {
	namen := make([]string, 0, len(a.Zeilen))
	for _, zeile := range a.Zeilen {
		namen = append(namen, zeile.Werte["name"])
	}

	return namen
}

// listenaufbauAus liest die Antwort.
func listenaufbauAus(t *testing.T, roh string) listenaufbau {
	t.Helper()

	var aufbau listenaufbau
	if err := json.Unmarshal([]byte(strings.TrimSpace(roh)), &aufbau); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	return aufbau
}

// dreiNumerierteKarten legt ein Set an, dessen Namen eine Zahl tragen.
//
// **»Karte 10« ist die Karte, um die es geht**: sie steht alphabetisch vor
// »Karte 2« und in einer richtig sortierten Liste dahinter.
func dreiNumerierteKarten(t *testing.T, umgebung []string) string {
	t.Helper()

	setID := leeresSet(t, umgebung, "Zum Ordnen")

	for _, name := range []string{"Karte 10", "Karte 2", "Karte 1"} {
		mussKommandozeile(t, umgebung, "karte", setID, "anlegen")

		karten := kartenzeilenAus(t,
			mussKommandozeile(t, umgebung, "karten", setID, "--json"))

		mussKommandozeile(t, umgebung, "feld", setID,
			karten[len(karten)-1].ID, "name", name)
	}

	return setID
}
