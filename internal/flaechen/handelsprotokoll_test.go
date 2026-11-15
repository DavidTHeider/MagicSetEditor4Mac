package flaechen_test

// **Wie ein Handel dorthin kam, an jeder Fläche gleich** — TRADE-03.AK-03
// (#1717).
//
// Das Protokoll schreibt der Bestand selbst; hier steht, dass Kommandozeile,
// Dienst und MCP für denselben Vorgang dieselben Wechsel nennen.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

func TestDasProtokollEinesHandelsIstUeberDreiFlaechenDasselbe(t *testing.T) {
	umgebung := append(mitEigenerSaat(t), schluesselbund.Bundort+"="+probebund(t))
	an := starteDienst(t, umgebung)

	mussKommandozeile(t, umgebung, "peer-anlegen")

	eroeffnet := mitEingabe(t, umgebung, diePositionen, "handel", "eroeffnen", "", "tausch", dasGegenueberImHandel)
	kennung, _, _ := strings.Cut(eroeffnet, "\t")
	mussKommandozeile(t, umgebung, "handel", "anbieten", kennung)

	for flaeche, aus := range map[string]string{
		"Kommandozeile": mussKommandozeile(t, umgebung, "handel", "protokoll", kennung, "--json"),
		"Dienst":        mussREST(t, an, http.MethodGet, "/handel/"+kennung+"/protokoll", ""),
		"MCP":           mussMCP(t, umgebung, "lesen.handelsprotokoll", map[string]any{"vorgang": kennung}),
	} {
		var protokoll struct {
			Zeilen []struct {
				Revision int    `json:"revision"`
				Von      string `json:"von"`
				Nach     string `json:"nach"`
				Zeit     string `json:"zeit"`
			} `json:"zeilen"`
		}

		if fehler := json.Unmarshal([]byte(aus), &protokoll); fehler != nil || len(protokoll.Zeilen) != 2 {
			t.Errorf("%s: %+v (%v) aus %s", flaeche, protokoll, fehler, aus)

			continue
		}

		neu, wechsel := protokoll.Zeilen[0], protokoll.Zeilen[1]
		if neu.Von != "" || neu.Nach != "entwurf" || wechsel.Von != "entwurf" ||
			wechsel.Nach != "angeboten" || wechsel.Revision != 1 || wechsel.Zeit == "" {
			t.Errorf("%s: %+v", flaeche, protokoll.Zeilen)
		}
	}

	// **Und die Textausgabe nennt eine neue Revision mit Strich.**
	text := mussKommandozeile(t, umgebung, "handel", "protokoll", kennung)
	if !strings.HasPrefix(text, "1\t-\tentwurf\t") || !strings.Contains(text, "\n1\tentwurf\tangeboten\t") {
		t.Errorf("Text: %q", text)
	}
}
