package flaechen_test

// **Was die Seiten eines Tauschs ungefähr wert sind, an jeder Fläche gleich**
// — TRADE-02b (#1951).
//
// Der Preis kommt aus der Gegenstelle in `scryfallgegenstelle_test.go`
// (1,50 EUR und 1,80 USD für einen gewöhnlichen Druck) — **kein Aufruf geht
// hinaus**.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

func TestDerPreisanhaltIstUeberDreiFlaechenDerselbe(t *testing.T) {
	umgebung := append(scryfallumgebung(t, scryfallgegenstelle(t)), schluesselbund.Bundort+"="+probebund(t))
	an := starteDienst(t, umgebung)

	mussKommandozeile(t, umgebung, "peer-anlegen")
	mussKommandozeile(t, umgebung, "preise-holen", "bolt-m10")

	eroeffnet := mitEingabe(t, umgebung, diePositionen, "handel", "eroeffnen", "", "tausch", dasGegenueberImHandel)
	kennung, _, _ := strings.Cut(eroeffnet, "\t")

	for flaeche, aus := range map[string]string{
		"Kommandozeile": mussKommandozeile(t, umgebung, "handel", "anhalt", kennung, "--json"),
		"Dienst":        mussREST(t, an, http.MethodGet, "/handel/"+kennung+"/anhalt", ""),
		"MCP":           mussMCP(t, umgebung, "lesen.handelsanhalt", map[string]any{"vorgang": kennung}),
	} {
		var anhalt struct {
			Zeilen []struct {
				Preise []struct {
					Waehrung string `json:"waehrung"`
					Wert     int64  `json:"wert"`
				} `json:"preise"`
			} `json:"zeilen"`
			Differenzen []struct {
				Waehrung     string `json:"waehrung"`
				Wert         int64  `json:"wert"`
				Vollstaendig bool   `json:"vollstaendig"`
			} `json:"differenzen"`
			Hauptwaehrung string `json:"hauptwaehrung"`
		}

		if fehler := json.Unmarshal([]byte(aus), &anhalt); fehler != nil || len(anhalt.Zeilen) != 1 ||
			len(anhalt.Differenzen) != 2 || anhalt.Hauptwaehrung != "EUR" {
			t.Errorf("%s: %+v (%v) aus %s", flaeche, anhalt, fehler, aus)

			continue
		}

		// **Das Gegenüber gibt, wir geben nichts** — die Differenz ist zu
		// unseren Gunsten und vollständig.
		eur := anhalt.Differenzen[0]
		if eur.Waehrung != "EUR" || eur.Wert != 150 || !eur.Vollstaendig {
			t.Errorf("%s: %+v", flaeche, anhalt.Differenzen)
		}
	}

	// **Und die Textausgabe nennt Seite, Differenz und Hauptwährung.**
	text := mussKommandozeile(t, umgebung, "handel", "anhalt", kennung)
	if !strings.Contains(text, "seite\t"+dasGegenueberImHandel+"\tEUR=150,USD=180\t0\n") ||
		!strings.Contains(text, "differenz\tEUR\t150\tvollstaendig\n") ||
		!strings.HasSuffix(text, "hauptwaehrung\tEUR\n") {
		t.Errorf("Text: %q", text)
	}
}
