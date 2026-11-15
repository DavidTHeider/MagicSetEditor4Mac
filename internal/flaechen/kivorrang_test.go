package flaechen_test

import (
	"net/http"
	"testing"
)

// **Wen der Vorrang fragt, steht an allen drei Flächen** — #2054, aus AI-02
// (#1743): dieselbe Reihe, die das Formular nennt, an Kommandozeile, MCP und
// Dienst; und ein gewähltes Modell, das die Laufzeit nicht nennt, wird an
// allen dreien gemeldet und nicht still ersetzt.
func TestDieReiheDesVorrangsUeberJedeFlaeche(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())
	laufzeit := starteLaufzeit(t, `{}`)

	kiEinschalten(t, umgebung, laufzeit.adresse)

	an := starteDienst(t, umgebung)

	katalog := [3]func() string{
		func() string { return mussKommandozeile(t, umgebung, "ki-modelle") },
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.ki.modelle", map[string]any{})
			if istFehler {
				t.Fatalf("lesen.ki.modelle: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet, "/ki/modelle", "")
			if status != http.StatusOK {
				t.Fatalf("GET ki/modelle: %d (%s)", status, rumpf)
			}

			return rumpf
		},
	}

	fuerAlleDreiMit(t, "die Reihe des Vorrangs",
		[3]string{"vorrang\tollama/" + probemodell, `"ollama/` + probemodell + `"`, `"vorrang":["ollama/` + probemodell + `"]`},
		katalog[0], katalog[1], katalog[2])

	mussKommandozeile(t, umgebung, "einstellungen", "setzen", "ki.vorrangmodell", "fehlt:1b")

	fuerAlleDreiMit(t, "ein gewähltes Modell, das fehlt",
		[3]string{"Vorrangmodell \"fehlt:1b\"", `Vorrangmodell \"fehlt:1b\"`, `Vorrangmodell \"fehlt:1b\"`},
		katalog[0], katalog[1], katalog[2])
}
