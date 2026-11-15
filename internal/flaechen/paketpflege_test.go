// **Die Paketpflege an jeder Fläche** —
// #1291.
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// paketquelle legt ein winziges, selbst geschriebenes Paket an (§11).
func paketquelle(t *testing.T, name string) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\nshort name: Probe\nfull name: Probeteil\nversion: 2026-08-30\n"
	if err := os.WriteFile(filepath.Join(ort, "style"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return ort
}

// **Wer über die Kommandozeile, MCP oder REST ein Paket einbindet, sieht es
// danach im Katalog.**
//
// # Der Befund
//
// `docs/FAEHIGKEITEN.md` führte die Zeile so:
//
//	| Ein Paket einbinden oder entfernen | `PaketInstallieren`, `PaketEntfernen` | — | — | — | ✋ |
//
// Die erhobene Hand meint einen Verzeichnisdialog. **Wer einen Rechner aus
// einem Skript einrichtet**, hat keinen — und `paket-beziehen` allein reicht
// nicht: ein Paket, das schon auf der Platte liegt, will eingebunden und nicht
// heruntergeladen werden.
func TestEinPaketWirdAnJederFlaecheEingebunden(t *testing.T) {
	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		quelle := paketquelle(t, "probe-cli.mse-style")

		if _, fehler, code := kommandozeile(t, umgebung,
			"paket-installieren", quelle); code != 0 {
			t.Fatalf("einbinden: %d (%s)", code, fehler)
		}

		imKatalog(t, umgebung, "die Kommandozeile", "probe-cli")
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		quelle := paketquelle(t, "probe-mcp.mse-style")

		if text, istFehler := mcpRuf(t, umgebung, "schreiben.paket.installieren",
			map[string]any{"quelle": quelle}); istFehler {
			t.Fatalf("einbinden: %s", text)
		}

		imKatalog(t, umgebung, "MCP", "probe-mcp")
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		quelle := paketquelle(t, "probe-rest.mse-style")
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPut, "/pakete/eingebunden",
			`{"quelle":`+alsJSONText(t, quelle)+`}`)
		if status != http.StatusCreated {
			t.Fatalf("PUT: %d (%s)", status, rumpf)
		}

		imKatalog(t, umgebung, "REST", "probe-rest")
	})
}

// **Und dieselbe fachliche Abweisung: eine Quelle ohne Paketendung** (§16).
func TestEineQuelleOhnePaketendungWirdAnJederFlaecheAbgewiesen(t *testing.T) {
	ohne := t.TempDir()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)

		if _, _, code := kommandozeile(t, umgebung, "paket-installieren", ohne); code == 0 {
			t.Error("die Kommandozeile hat es hingenommen")
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)

		if text, istFehler := mcpRuf(t, umgebung, "schreiben.paket.installieren",
			map[string]any{"quelle": ohne}); !istFehler {
			t.Errorf("MCP hat es hingenommen: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPut, "/pakete/eingebunden",
			`{"quelle":`+alsJSONText(t, ohne)+`}`)
		if status < 400 {
			t.Errorf("REST hat es hingenommen: %d (%s)", status, rumpf)
		}
	})
}

// **Die üblichen MSE2-Orte kommen an jeder Fläche an.**
//
// **Eine leere Liste ist eine gültige Antwort** — auf einem Rechner ohne MSE2
// gibt es keinen Vorschlag. Was nicht geht, ist `null` (§3).
func TestDieMSE2OrteStehenAnJederFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	t.Run("Kommandozeile", func(t *testing.T) {
		if _, fehler, code := kommandozeile(t, umgebung, "mse2-orte", "--json"); code != 0 {
			t.Fatalf("orte: %d (%s)", code, fehler)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "lesen.mse2orte", map[string]any{})
		if istFehler {
			t.Fatalf("orte: %s", text)
		}

		if strings.Contains(text, `"orte": null`) {
			t.Errorf("die Orte sind null: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodGet, "/mse2/orte", "")
		if status != http.StatusOK {
			t.Fatalf("GET: %d (%s)", status, rumpf)
		}

		if strings.Contains(rumpf, `"orte":null`) {
			t.Errorf("die Orte sind null: %s", rumpf)
		}
	})
}

// **Und das Entfernen verlangt an den beiden unbeaufsichtigten Flächen eine
// Bestätigung — an der Kommandozeile nicht.**
//
// Das ist der Unterschied aus
// [0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md):
// die **Aufsicht**, nicht die Fähigkeit. Wer tippt, sieht seinen Aufruf; ein
// Sprachmodell und ein Dienst rufen, während niemand hinsieht.
func TestDasEntfernenIstAnDenUnbeaufsichtigtenFlaechenBestaetigt(t *testing.T) {
	t.Run("MCP weist einen falschen Pfad ab", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		quelle := paketquelle(t, "probe-mcpweg.mse-style")

		if text, istFehler := mcpRuf(t, umgebung, "schreiben.paket.installieren",
			map[string]any{"quelle": quelle}); istFehler {
			t.Fatalf("einbinden: %s", text)
		}

		if text, istFehler := mcpRuf(t, umgebung, "schreiben.paket.entfernen",
			map[string]any{
				"art": "mse-style", "name": "probe-mcpweg", "pfad": "/woanders",
			}); !istFehler {
			t.Errorf("MCP hat es hingenommen: %s", text)
		}

		// **Und es liegt danach noch da** — eine Abweisung, die trotzdem
		// löscht, wäre schlimmer als gar keine.
		imKatalog(t, umgebung, "MCP nach der Abweisung", "probe-mcpweg")
	})

	t.Run("die Kommandozeile entfernt ohne Bestätigung", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		quelle := paketquelle(t, "probe-cliweg.mse-style")

		if _, fehler, code := kommandozeile(t, umgebung,
			"paket-installieren", quelle); code != 0 {
			t.Fatalf("einbinden: %d (%s)", code, fehler)
		}

		if _, fehler, code := kommandozeile(t, umgebung,
			"paket-entfernen", "mse-style", "probe-cliweg"); code != 0 {
			t.Fatalf("entfernen: %d (%s)", code, fehler)
		}

		if liste := mussKommandozeile(t, umgebung, "pakete"); strings.Contains(liste, "probe-cliweg") {
			t.Errorf("das Paket steht noch im Katalog: %s", liste)
		}
	})
}

// imKatalog sieht nach, dass ein Paket im Katalog steht.
//
// **Über die Kommandozeile und nicht über die Antwort**: zwei grüne Seiten
// einer Grenze sind kein Nachweis für die Grenze.
func imKatalog(t *testing.T, umgebung []string, wer, name string) {
	t.Helper()

	if liste := mussKommandozeile(t, umgebung, "pakete"); !strings.Contains(liste, name) {
		t.Errorf("nach %s fehlt %q im Katalog: %s", wer, name, liste)
	}
}
