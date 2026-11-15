package flaechen_test

// **Der Dienstkatalog über jede Fläche** — Werkstück A1 von Phase 5.3
// (#1440):
// dieselbe Auskunft an der Kommandozeile, über MCP und im Dienst — und an
// keiner geht dabei eine Anfrage hinaus.

import (
	"net/http"
	"strings"
	"testing"
)

func TestDerDienstkatalogUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	fuerAlleDreiMit(t, "der Dienstkatalog",
		[3]string{"scryfall\tsuche\tbereit", `"status": "bereit"`, `"status":"bereit"`},
		func() string { return mussKommandozeile(t, umgebung, "dienstkatalog") },
		func() string { return mussMCP(t, umgebung, "lesen.dienstkatalog", map[string]any{}) },
		func() string { return mussREST(t, an, http.MethodGet, "/dienstkatalog", "") })

	// **Die drei neuen fangen aus an** (0063) — und der Grund steht dabei.
	fuerAlleDreiMit(t, "der abgeschaltete Dienst",
		[3]string{
			"scrydex\tsuche\taus", `"grund": "der Dienst ist abgeschaltet"`,
			`"grund":"der Dienst ist abgeschaltet"`,
		},
		func() string { return mussKommandozeile(t, umgebung, "dienstkatalog") },
		func() string { return mussMCP(t, umgebung, "lesen.dienstkatalog", map[string]any{}) },
		func() string { return mussREST(t, an, http.MethodGet, "/dienstkatalog", "") })

	// **Ein Nachweis aus einer anderen Fassung gilt nicht für diese** — und
	// die JSON-Flächen sagen es Wort für Wort gleich.
	rest := mussREST(t, an, http.MethodGet, "/dienstkatalog", "")
	if !strings.Contains(rest, `"giltFuerDieseFassung"`) {
		t.Errorf("der Dienst nennt die Geltung des Nachweises nicht: %s", rest)
	}
}
