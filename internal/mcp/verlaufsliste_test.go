package mcp_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Der Verlauf über MCP** — Werkstück U6 von Phase 5.1 (#1417): lesend,
// idempotent, der jüngste Schritt zuerst, mit Vorgang, Umfang, Vorher und
// Nachher — alles, was »schreiben.zuruecknehmen« mit einem Vorgang braucht.
func TestDerVerlaufNenntSeineSchritte(t *testing.T) {
	probe := probenbestand()
	werkzeug := werkzeugMitNamen(t, probe, "lesen.verlauf")

	antwort, fehler := werkzeug.Ruf(t.Context(), json.RawMessage(`{"setID": 1}`))
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	var gelesen struct {
		Schritte []struct {
			ID        string `json:"id"`
			Vorgang   string `json:"vorgang"`
			Umfang    int    `json:"umfang"`
			KarteID   string `json:"karteID"`
			Feld      string `json:"feld"`
			Vorher    string `json:"vorher"`
			Nachher   string `json:"nachher"`
			Urheber   string `json:"urheber"`
			Zeitpunkt string `json:"zeitpunkt"`
		} `json:"schritte"`
	}
	if err := json.Unmarshal([]byte(antwort), &gelesen); err != nil {
		t.Fatalf("Antwort %q: %v", antwort, err)
	}

	if len(gelesen.Schritte) != 2 {
		t.Fatalf("%d Schritte: %s", len(gelesen.Schritte), antwort)
	}

	if gelesen.Schritte[0].Vorgang != "g1" || gelesen.Schritte[0].Umfang != 2 ||
		gelesen.Schritte[0].KarteID != "8" || gelesen.Schritte[0].Nachher != "Land" {
		t.Errorf("jüngster Schritt: %+v", gelesen.Schritte[0])
	}

	if gelesen.Schritte[1].Feld != "name" || gelesen.Schritte[1].Vorher != "Alt" ||
		gelesen.Schritte[1].Zeitpunkt == "" {
		t.Errorf("älterer Schritt: %+v", gelesen.Schritte[1])
	}
}

// **Eine schlechte Kennung und ein Bestand, der nicht antwortet, kommen als
// Fehler** — nicht als leere Liste.
func TestDerVerlaufMeldetSeineFehler(t *testing.T) {
	werkzeug := werkzeugMitNamen(t, probenbestand(), "lesen.verlauf")

	if _, fehler := werkzeug.Ruf(t.Context(), json.RawMessage(`{"setID": 0}`)); fehler == nil {
		t.Error("die Kennung 0 ging durch")
	}

	// **Ein Bestand, der schweigt** — die Naht wird eigens gesetzt, denn
	// `probenaehte` legt sich seine eigene an.
	naehte := probenaehte(probenbestand())
	stumm := neueBestandsprobeWeit()
	stumm.fehler = errors.New("der Bestand ist zu")
	naehte.Bestand = stumm

	for _, kandidat := range mcp.Werkzeuge(naehte) {
		if kandidat.Name != "lesen.verlauf" {
			continue
		}

		if _, fehler := kandidat.Ruf(t.Context(), json.RawMessage(`{"setID": 1}`)); fehler == nil {
			t.Error("der Fehler des Bestands kam nicht an")
		}

		return
	}

	t.Fatal("kein Werkzeug lesen.verlauf")
}
