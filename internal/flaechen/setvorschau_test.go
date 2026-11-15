// **Ein Set vorab ansehen, an jeder Fläche dieselbe Antwort** — FMT-03.AK-01
// (#1692).
//
// Die Vorschau liest wie das Einfahren und schreibt nichts: vorher keine
// Kennung, danach die des eingefahrenen Sets — an allen drei Flächen gleich.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

type setvorschau struct {
	Name       string `json:"name"`
	Spiel      string `json:"spiel"`
	Karten     int    `json:"karten"`
	SchonOffen string `json:"schonOffen"`
	Bedarf     []struct {
		Braucht string `json:"braucht"`
	} `json:"bedarf"`
}

// vorschauAnJederFlaeche fragt dreimal und gibt die drei Antworten.
func vorschauAnJederFlaeche(t *testing.T, umgebung []string, an dienst, pfad string) [3]setvorschau {
	t.Helper()

	var antworten [3]setvorschau

	if err := json.Unmarshal([]byte(mussKommandozeile(t, umgebung, "--json", "set-vorschau", pfad)), &antworten[0]); err != nil {
		t.Fatalf("Kommandozeile: %v", err)
	}

	text, istFehler := mcpRuf(t, umgebung, "lesen.set.vorschau", map[string]any{"pfad": pfad})
	if istFehler || json.Unmarshal([]byte(text), &antworten[1]) != nil {
		t.Fatalf("MCP: %s", text)
	}

	status, rumpf := an.ruf(t, http.MethodPost, "/sets/vorschau", `{"pfad":`+alsJSONText(t, pfad)+`}`)

	var umschlag struct {
		Vorschau setvorschau `json:"vorschau"`
	}

	if status != http.StatusOK || json.Unmarshal([]byte(rumpf), &umschlag) != nil {
		t.Fatalf("REST: %d %s", status, rumpf)
	}

	antworten[2] = umschlag.Vorschau

	return antworten
}

func TestEineSetvorschauIstAnJederFlaecheDieselbe(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)
	pfad := probeset(t)

	vorher := vorschauAnJederFlaeche(t, umgebung, an, pfad)
	for flaeche, vorschau := range vorher {
		if vorschau.Name != "Mein Probeset" || vorschau.Karten != 3 || vorschau.SchonOffen != "" ||
			len(vorschau.Bedarf) != len(vorher[0].Bedarf) {
			t.Errorf("Fläche %d vorher: %+v", flaeche, vorschau)
		}
	}

	// **Geschrieben hat keine** — erst das Einfahren legt das Set an.
	setID := setkennungAus(t, mussKommandozeile(t, umgebung, "oeffnen", pfad, "--json"))

	for flaeche, vorschau := range vorschauAnJederFlaeche(t, umgebung, an, pfad) {
		if vorschau.SchonOffen != setID {
			t.Errorf("Fläche %d nachher: %+v, erwartet schon offen als %s", flaeche, vorschau, setID)
		}
	}
}
