package flaechen_test

// **Wo ein Handel steht, an jeder Fläche gleich** — TRADE-02.AK-02
// (#1716).
//
// **Phase, Lage, nächster und mögliche Schritte** kommen aus `domain`; hier
// steht, dass Kommandozeile, Dienst und MCP für denselben Vorgang dasselbe
// sagen — und dass der Beschluss aus TRADE-03 jetzt auch außerhalb des
// Fensters ankommt.
//
// **Und für wen das gilt** — `ich`, die Partei dieses Profils
// (#2026).
// Das Fenster setzt damit die eigene Abschlussstimme; eine Fläche, die das
// Feld nicht trüge, ließe ihren Nutzer raten, wessen Stimme fehlt.

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

func TestDieLageEinesHandelsIstUeberDreiFlaechenDieselbe(t *testing.T) {
	umgebung := append(mitEigenerSaat(t), schluesselbund.Bundort+"="+probebund(t))
	an := starteDienst(t, umgebung)

	ich := zeilenwert(t, mussKommandozeile(t, umgebung, "peer-anlegen"), "kennung")

	eroeffnet := mitEingabe(t, umgebung, diePositionen, "handel", "eroeffnen", "", "tausch", dasGegenueberImHandel)
	kennung, _, _ := strings.Cut(eroeffnet, "\t")
	mussKommandozeile(t, umgebung, "handel", "anbieten", kennung)

	for flaeche, aus := range map[string]string{
		"Kommandozeile": mussKommandozeile(t, umgebung, "handel", "stand", kennung, "--json"),
		"Dienst":        mussREST(t, an, http.MethodGet, "/handel/"+kennung, ""),
		"MCP":           mussMCP(t, umgebung, "lesen.handel", map[string]any{"vorgang": kennung}),
	} {
		var stand struct {
			Phase     string   `json:"phase"`
			Lage      string   `json:"lage"`
			Naechster string   `json:"naechster"`
			Moegliche []string `json:"moegliche"`
			Beschluss string   `json:"beschluss"`
			Ich       string   `json:"ich"`
		}

		if fehler := json.Unmarshal([]byte(aus), &stand); fehler != nil ||
			stand.Phase != "angebot" || stand.Beschluss != "zustimmen" || stand.Ich != ich ||
			!slices.Equal(stand.Moegliche, []string{"zustimmen", "abbrechen", "ablaufen"}) {
			t.Errorf("%s: %+v (%v) aus %s", flaeche, stand, fehler, aus)
		}
	}

	// **Und die Textausgabe nennt es eingerückt** — die erste Zeile bleibt,
	// wie sie war.
	text := mussKommandozeile(t, umgebung, "handel", "stand", kennung)
	if !strings.Contains(text, "\tphase\tangebot\n") || !strings.Contains(text, "\tschritte\tzustimmen,abbrechen,ablaufen\n") ||
		!strings.HasPrefix(text, kennung+"\t") {
		t.Errorf("Text: %q", text)
	}
}
