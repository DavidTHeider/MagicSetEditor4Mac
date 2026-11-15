// **Die Übersicht über die Pakete — an jeder Fläche** — DOM-01.AK-01
// (#1696).
//
// Ein Set wird geöffnet; danach nennen Kommandozeile, MCP und Dienst für sein
// Spiel und seinen Stil dieselbe Nutzung, dieselbe Herkunftsart und
// dieselben Abhängigen (§16).
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"slices"
	"testing"
)

// uebersichtszeile ist, was jede Fläche je Paket nennt.
type uebersichtszeile struct {
	Art          string   `json:"art"`
	Name         string   `json:"name"`
	Herkunftsart string   `json:"herkunftsart"`
	Geaendert    bool     `json:"geaendert"`
	GebrauchtVon []string `json:"gebrauchtVon"`
	Benutzt      int      `json:"benutzt"`
}

// pruefeUebersicht hält Spiel und Stil des Probesets.
func pruefeUebersicht(t *testing.T, flaeche, roh string) {
	t.Helper()

	var antwort struct {
		Pakete []uebersichtszeile `json:"pakete"`
	}

	if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
		t.Fatalf("%s: die Antwort ist kein gültiges JSON: %v (%s)", flaeche, err, roh)
	}

	nach := map[string]uebersichtszeile{}
	for _, zeile := range antwort.Pakete {
		nach[zeile.Name+"."+zeile.Art] = zeile
	}

	spiel, stil := nach["probe.mse-game"], nach["probe-stil.mse-style"]

	// **Von Hand ins verwaltete Verzeichnis gelegt ist eigen** — so legt
	// `mitEigenerSaat` die Saat hin.
	if spiel.Benutzt != 1 || spiel.Herkunftsart != "eigen" || spiel.Geaendert ||
		!slices.Contains(spiel.GebrauchtVon, "probe-schlicht.mse-style") {
		t.Errorf("%s: das Spiel %+v", flaeche, spiel)
	}

	if stil.Benutzt != 1 {
		t.Errorf("%s: der Stil %+v", flaeche, stil)
	}
}

func TestDieUebersichtUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json")
	an := starteDienst(t, umgebung)

	pruefeUebersicht(t, "Kommandozeile", mussKommandozeile(t, umgebung, "pakete", "--json"))
	pruefeUebersicht(t, "MCP", mussMCP(t, umgebung, "lesen.pakete", map[string]any{}))
	pruefeUebersicht(t, "REST", mussREST(t, an, http.MethodGet, "/pakete", ""))
}
