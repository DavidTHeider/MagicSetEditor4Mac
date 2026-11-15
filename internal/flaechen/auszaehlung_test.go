// **Auszählen nach einer erklärten Dimension — an jeder Fläche** — Werkstück C
// von Phase 4.11
// (#1338,
// #1349).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Was hier geprüft wird
//
// Die Saat `probe.mse-game` bringt seit #1338 einen `statistics dimension:`-
// Block mit — dieselbe Gestalt wie `magic.mse-game/statistics_en`, eigener
// Inhalt (§11). Er zählt nach dem Feld `farbe` und **erklärt die Reihenfolge**:
// „blau" vor „rot", obwohl „rot" alphabetisch früher käme. Daran hängt die
// Zusage, dass die erklärte Reihenfolge gilt und nicht die Sortierung.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// **Jede Fläche nennt dieselben Dimensionen.**
func TestDieAuszaehldimensionenStehenUeberallGleich(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Zaehlset")

		pruefeDimensionen(t, mussKommandozeile(t, umgebung,
			"auszaehlung", setID, "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Zaehlset")

		roh, istFehler := mcpRuf(t, umgebung, "lesen.auszaehldimensionen",
			map[string]any{"setID": zahl(t, setID)})
		if istFehler {
			t.Fatalf("lesen.auszaehldimensionen: %s", roh)
		}

		pruefeDimensionen(t, roh)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Zaehlset")

		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodGet, "/sets/"+setID+"/auszaehlung", "")
		if status != http.StatusOK {
			t.Fatalf("GET /sets/%s/auszaehlung: %d (%s)", setID, status, rumpf)
		}

		pruefeDimensionen(t, rumpf)
	})
}

// **Und jede Fläche zählt dieselben Gruppen — in der erklärten Reihenfolge.**
func TestDieAuszaehlungLaeuftUeberallGleich(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := zaehlsetMitKarten(t, umgebung)

		pruefeGruppen(t, mussKommandozeile(t, umgebung,
			"auszaehlung", setID, "farbe", "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := zaehlsetMitKarten(t, umgebung)

		roh, istFehler := mcpRuf(t, umgebung, "lesen.auszaehlung",
			map[string]any{"setID": zahl(t, setID), "name": "farbe"})
		if istFehler {
			t.Fatalf("lesen.auszaehlung: %s", roh)
		}

		pruefeGruppen(t, roh)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := zaehlsetMitKarten(t, umgebung)

		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodGet,
			"/sets/"+setID+"/auszaehlung/farbe", "")
		if status != http.StatusOK {
			t.Fatalf("GET /sets/%s/auszaehlung/farbe: %d (%s)", setID, status, rumpf)
		}

		pruefeGruppen(t, rumpf)
	})
}

// **Und eine Dimension, die es nicht gibt, ist überall dasselbe Nein.**
func TestEineUnbekannteDimensionScheitertUeberallGleich(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Zaehlset")

		if _, _, code := kommandozeile(t, umgebung,
			"auszaehlung", setID, "gibt es nicht"); code != 1 {
			t.Errorf("Rückgabewert %d, erwartet war 1 (fachlich)", code)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Zaehlset")

		if _, istFehler := mcpRuf(t, umgebung, "lesen.auszaehlung",
			map[string]any{"setID": zahl(t, setID), "name": "gibt es nicht"}); !istFehler {
			t.Error("MCP nahm eine unbekannte Dimension an")
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Zaehlset")

		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodGet,
			"/sets/"+setID+"/auszaehlung/gibt%20es%20nicht", "")
		if status != http.StatusNotFound {
			t.Errorf("Status %d, erwartet war 404 (%s)", status, rumpf)
		}
	})
}

// zaehlsetMitKarten legt ein Set an und lässt das Kartenskript zwei Karten
// hineinlegen — eine rote und eine blaue.
func zaehlsetMitKarten(t *testing.T, umgebung []string) string {
	t.Helper()

	setID := neuesSet(t, umgebung, "Zaehlset")

	mussKommandozeile(t, umgebung, "kartenskript", setID, "Zwei Farben")

	return setID
}

// pruefeDimensionen hält die fachliche Zusage — an jeder Fläche dieselbe.
func pruefeDimensionen(t *testing.T, roh string) {
	t.Helper()

	var stand struct {
		Dimensionen []struct {
			Name         string `json:"name"`
			Beschreibung string `json:"beschreibung"`
		} `json:"dimensionen"`
	}

	if err := json.Unmarshal([]byte(roh), &stand); err != nil {
		t.Fatalf("die Antwort ist kein JSON: %v\n%s", err, roh)
	}

	// **Erst die automatischen der Felder, dann die erklärte** (#1985,
	// `Game::validate` in MSE2). Das Feld `farbe` gibt keine zweite, weil die
	// erklärte gleichen Namens vorgeht; `kunst` ist ein Bildfeld.
	namen := make([]string, 0, len(stand.Dimensionen))
	for _, eine := range stand.Dimensionen {
		namen = append(namen, eine.Name)
	}

	if strings.Join(namen, ",") != "name,text,farbe" {
		t.Fatalf("Dimensionen %v, erwartet name, text, farbe: %s", namen, roh)
	}

	if stand.Dimensionen[2].Beschreibung == "" {
		t.Error("die Beschreibung der erklärten fehlt — sie steht in der Saat")
	}
}

// pruefeGruppen hält die Zusage über die Auszählung.
func pruefeGruppen(t *testing.T, roh string) {
	t.Helper()

	var stand struct {
		Dimension string `json:"dimension"`
		Gruppen   []struct {
			Name      string   `json:"name"`
			Karten    int      `json:"karten"`
			Farbe     string   `json:"farbe"`
			KartenIDs []string `json:"kartenIDs"`
		} `json:"gruppen"`
		Befunde []string `json:"befunde"`
	}

	if err := json.Unmarshal([]byte(roh), &stand); err != nil {
		t.Fatalf("die Antwort ist kein JSON: %v\n%s", err, roh)
	}

	if len(stand.Gruppen) != 2 {
		t.Fatalf("%d Gruppen, erwartet waren zwei: %s", len(stand.Gruppen), roh)
	}

	// **Die erklärte Reihenfolge und nicht die alphabetische.** „blau" steht
	// in der Saat vorn; wer die Erklärung fallen ließe, bekäme „rot" zuerst.
	if stand.Gruppen[0].Name != "blau" || stand.Gruppen[1].Name != "rot" {
		t.Errorf("die Reihenfolge lautet %q, %q",
			stand.Gruppen[0].Name, stand.Gruppen[1].Name)
	}

	for _, gruppe := range stand.Gruppen {
		if gruppe.Karten != 1 {
			t.Errorf("die Gruppe %q zählt %d Karten", gruppe.Name, gruppe.Karten)
		}

		if gruppe.Farbe == "" {
			t.Errorf("der Gruppe %q fehlt ihre erklärte Farbe", gruppe.Name)
		}

		// **Jede Gruppe nennt ihre Karte** (UXR-F6, #2013) — an jeder Fläche
		// dieselbe, denn die Seite filtert damit die Kartenliste, und der Rest
		// fragt, welche Karte im Balken stand.
		if len(gruppe.KartenIDs) != gruppe.Karten {
			t.Errorf("die Gruppe %q zählt %d Karten und nennt %v",
				gruppe.Name, gruppe.Karten, gruppe.KartenIDs)
		}
	}

	// **Und es sind verschiedene**: die rote und die blaue Karte sind zwei.
	if len(stand.Gruppen[0].KartenIDs) == 1 && len(stand.Gruppen[1].KartenIDs) == 1 &&
		stand.Gruppen[0].KartenIDs[0] == stand.Gruppen[1].KartenIDs[0] {
		t.Errorf("beide Gruppen nennen die Karte %s", stand.Gruppen[0].KartenIDs[0])
	}

	if stand.Befunde == nil {
		t.Error("die Befunde sind null statt einer leeren Liste")
	}
}
