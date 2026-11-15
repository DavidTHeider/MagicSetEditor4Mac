// **Ein Paket als `.mse-installer` — hinaus und wieder herein, an jeder
// Fläche** — FMT-02.AK-04
// (#1691).
//
// Je Fläche derselbe Kreis: ein eigenes Paket anlegen, weitergeben,
// entfernen, einlesen — und ein zweites Einlesen **übergeht**. Anders ist
// nur das zweite Weitergeben an denselben Ort: die Kommandozeile ersetzt wie
// `--create-installer` bei MSE2, MCP und REST überschreiben nichts (0044).
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// eingelesen ist, was `installer-einlesen --json` und der Dienst antworten.
type eingelesen struct {
	Uebernommen []struct {
		Name string `json:"name"`
	} `json:"uebernommen"`
	Uebergangen []string `json:"uebergangen"`
}

func TestEinInstallerGehtUeberJedeFlaeche(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		ablage := downloadverzeichnis(t, umgebung)
		mussKommandozeile(t, umgebung, "paket-anlegen", "mse-include", "bausteine")

		datei := strings.TrimSpace(mussKommandozeile(t, umgebung, "paket-weitergeben", ablage, "bausteine.mse-include"))
		if datei != filepath.Join(ablage, "bausteine.mse-installer") {
			t.Fatalf("geschrieben nach %q", datei)
		}

		// **Ein zweites Mal ersetzt** — wie `saveAs` bei MSE2.
		mussKommandozeile(t, umgebung, "paket-weitergeben", datei, "bausteine.mse-include")

		mussKommandozeile(t, umgebung, "paket-entfernen", "mse-include", "bausteine")
		mussKommandozeile(t, umgebung, "installer-einlesen", datei)
		imKatalog(t, umgebung, "die Kommandozeile", "bausteine")

		var nochmal eingelesen
		if err := json.Unmarshal([]byte(mussKommandozeile(t, umgebung, "installer-einlesen", datei, "--json")), &nochmal); err != nil ||
			!slices.Equal(nochmal.Uebergangen, []string{"bausteine.mse-include"}) {
			t.Errorf("ein zweites Einlesen: %+v, %v", nochmal, err)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		ablage := downloadverzeichnis(t, umgebung)
		mussKommandozeile(t, umgebung, "paket-anlegen", "mse-include", "bausteine")

		weitergabe := map[string]any{"pakete": []string{"bausteine.mse-include"}, "ziel": ablage}
		if text, istFehler := mcpRuf(t, umgebung, "schreiben.paket.weitergeben", weitergabe); istFehler ||
			!strings.Contains(text, "bausteine.mse-installer") {
			t.Fatalf("weitergeben: %s", text)
		}

		// **Nichts wird überschrieben** (0044).
		if text, istFehler := mcpRuf(t, umgebung, "schreiben.paket.weitergeben", weitergabe); !istFehler {
			t.Errorf("ein zweites Weitergeben an denselben Ort überschrieb: %s", text)
		}

		mussKommandozeile(t, umgebung, "paket-entfernen", "mse-include", "bausteine")

		datei := filepath.Join(ablage, "bausteine.mse-installer")
		if text, istFehler := mcpRuf(t, umgebung, "schreiben.paket.einlesen", map[string]any{"datei": datei}); istFehler ||
			!strings.Contains(text, `"name": "bausteine"`) {
			t.Fatalf("einlesen: %s", text)
		}

		imKatalog(t, umgebung, "MCP", "bausteine")
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		ablage := downloadverzeichnis(t, umgebung)
		an := starteDienst(t, umgebung)
		mussKommandozeile(t, umgebung, "paket-anlegen", "mse-include", "bausteine")

		rumpf := `{"pakete":["bausteine.mse-include"],"ziel":"` + ablage + `"}`
		if status, antwort := an.ruf(t, http.MethodPost, "/pakete/weitergabe", rumpf); status != http.StatusCreated ||
			!strings.Contains(antwort, "bausteine.mse-installer") {
			t.Fatalf("weitergeben: %d (%s)", status, antwort)
		}

		if status, antwort := an.ruf(t, http.MethodPost, "/pakete/weitergabe", rumpf); status != http.StatusConflict {
			t.Errorf("ein zweites Weitergeben an denselben Ort: %d (%s)", status, antwort)
		}

		mussKommandozeile(t, umgebung, "paket-entfernen", "mse-include", "bausteine")

		datei := filepath.Join(ablage, "bausteine.mse-installer")
		status, antwort := an.ruf(t, http.MethodPost, "/pakete/eingelesen", `{"datei":"`+datei+`"}`)

		var ergebnis eingelesen
		if err := json.Unmarshal([]byte(antwort), &ergebnis); status != http.StatusCreated || err != nil ||
			len(ergebnis.Uebernommen) != 1 || ergebnis.Uebernommen[0].Name != "bausteine" {
			t.Fatalf("einlesen: %d (%s)", status, antwort)
		}
	})
}
