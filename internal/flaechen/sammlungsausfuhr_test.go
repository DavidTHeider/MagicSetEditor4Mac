// **Die Sammlung ausgeben und zurücklesen — an jeder Fläche** — Werkstück 4
// von Phase 4.5 (#1292).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Der Rundlauf ist die Zusage
//
// Ausgeben, in einen **anderen** Bestand einlesen, und dieselbe Zeile
// bekommen. Das ist die Prüfung, die zählt: was hineingeht, muss auch wieder
// heraus — und zwar vollständig.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// mitExemplar legt ein Exemplar mit jeder Angabe an und gibt die Umgebung.
func mitExemplar(t *testing.T) []string {
	t.Helper()

	umgebung := eigenerBestand(t)

	mussKommandozeile(t, umgebung, "exemplar",
		"probe-druck", "4", "played", "de", "foil",
		"2026-08-30", "4.17", "EUR", "getauscht, 2026")

	return umgebung
}

// **Die Ausgabe trägt jede Angabe — an jeder Fläche dieselbe.**
func TestDieSammlungWirdUeberallGleichAusgegeben(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		pruefeCSV(t, mussKommandozeile(t, mitExemplar(t), "sammlung-ausgeben", "csv"))
	})

	t.Run("MCP", func(t *testing.T) {
		roh, istFehler := mcpRuf(t, mitExemplar(t), "lesen.sammlung.ausgeben",
			map[string]any{"form": "csv"})
		if istFehler {
			t.Fatalf("lesen.sammlung.ausgeben: %s", roh)
		}

		pruefeCSV(t, roh)
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, mitExemplar(t))

		status, rumpf := an.ruf(t, http.MethodGet, "/sammlung/ausgeben?form=csv", "")
		if status != http.StatusOK {
			t.Fatalf("GET /sammlung/ausgeben: %d (%s)", status, rumpf)
		}

		var antwort struct {
			Text string `json:"text"`
		}

		if err := json.Unmarshal([]byte(rumpf), &antwort); err != nil {
			t.Fatalf("die Antwort lesen: %v (%s)", err, rumpf)
		}

		pruefeCSV(t, antwort.Text)
	})
}

// pruefeCSV hält, was in der Ausgabe stehen muss.
func pruefeCSV(t *testing.T, text string) {
	t.Helper()

	for _, teil := range []string{
		"printing", "probe-druck", "played", "de", "true",
		"2026-08-30", "4.17", "EUR", "getauscht, 2026",
	} {
		if !strings.Contains(text, teil) {
			t.Errorf("die Ausgabe nennt %q nicht:\n%s", teil, text)
		}
	}
}

// **Und sie liest sich zurück — an jeder Fläche, in einen frischen Bestand.**
func TestDieSammlungLiestSichUeberallZurueck(t *testing.T) {
	t.Parallel()

	// **Einmal ausgegeben, dreimal eingelesen.** Der Text ist derselbe; was
	// sich unterscheidet, ist der Weg hinein.
	text := mussKommandozeile(t, mitExemplar(t), "sammlung-ausgeben", "csv")

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := eigenerBestand(t)

		mitEingabe(t, umgebung, text, "sammlung-einlesen")
		pruefeZurueck(t, mussKommandozeile(t, umgebung, "sammlung", "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := eigenerBestand(t)

		roh, istFehler := mcpRuf(t, umgebung, "schreiben.sammlung.einlesen",
			map[string]any{"text": text})
		if istFehler {
			t.Fatalf("schreiben.sammlung.einlesen: %s", roh)
		}

		pruefeZurueck(t, mussKommandozeile(t, umgebung, "sammlung", "--json"))
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := eigenerBestand(t)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/sammlung/einlesen",
			string(mussJSON(t, map[string]any{"text": text})))
		if status != http.StatusOK {
			t.Fatalf("POST /sammlung/einlesen: %d (%s)", status, rumpf)
		}

		pruefeZurueck(t, mussKommandozeile(t, umgebung, "sammlung", "--json"))
	})
}

// pruefeZurueck hält, was nach dem Zurücklesen im Bestand steht.
//
// **Jede Angabe, nicht nur Druck und Menge.** Ein Rundlauf, der den Preis
// verliert, sähe an Druck und Menge gemessen vollständig aus — und war es beim
// ersten Anlauf auch nicht.
func pruefeZurueck(t *testing.T, roh string) {
	t.Helper()

	var sammlung struct {
		Exemplare []struct {
			Printing        string `json:"printing"`
			Menge           int    `json:"menge"`
			Zustand         string `json:"zustand"`
			Sprache         string `json:"sprache"`
			Foil            bool   `json:"foil"`
			Finish          string `json:"finish"`
			Erwerbsdatum    string `json:"erwerbsdatum"`
			Erwerbspreis    string `json:"erwerbspreis"`
			Erwerbswaehrung string `json:"erwerbswaehrung"`
			Notiz           string `json:"notiz"`
		} `json:"exemplare"`
	}

	if err := json.Unmarshal([]byte(roh), &sammlung); err != nil {
		t.Fatalf("die Sammlung lesen: %v (%s)", err, roh)
	}

	if len(sammlung.Exemplare) != 1 {
		t.Fatalf("%d Exemplare, erwartet 1: %s", len(sammlung.Exemplare), roh)
	}

	eines := sammlung.Exemplare[0]

	for _, fall := range []struct {
		name     string
		gelesen  string
		erwartet string
	}{
		{"Druck", eines.Printing, "probe-druck"},
		{"Zustand", eines.Zustand, "played"},
		{"Sprache", eines.Sprache, "de"},
		{"Erwerbsdatum", eines.Erwerbsdatum, "2026-08-30"},
		{"Erwerbspreis", eines.Erwerbspreis, "4.17"},
		{"Währung", eines.Erwerbswaehrung, "EUR"},
		{"Notiz", eines.Notiz, "getauscht, 2026"},

		// **Die Ausführung geht durch das CSV und zurück** (T2, #1467). Der
		// Wahrheitswert daneben sagt dasselbe — hier, wo es ein Foil ist, und
		// nicht bei einer geätzten Karte.
		{"Ausführung", eines.Finish, "foil"},
	} {
		if fall.gelesen != fall.erwartet {
			t.Errorf("%s kam als %q zurück, erwartet %q",
				fall.name, fall.gelesen, fall.erwartet)
		}
	}

	if eines.Menge != 4 || !eines.Foil {
		t.Errorf("Menge %d, Foil %v — erwartet 4 und true", eines.Menge, eines.Foil)
	}
}
