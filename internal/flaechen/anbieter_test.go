// **Kontostand und Fremdverlauf über jede Fläche ohne Fenster** — Werkstück A5
// von Phase 5.3 (#1444).
//
// # Was hier gemessen wird
//
// **Die Ablehnung, und dass sie überall dieselbe ist** (§16) — samt einem
// Zähler daneben: ohne Zugangsdaten geht **keine einzige Verbindung** hinaus.
// Das ist die Abnahme aus dem Issue („Fehlende/entzogene … Capability ergibt
// erklärten Status und keinen automatischen Wechsel in kostenpflichtigen
// Pfad").
//
// **Nicht gemessen wird der geglückte Abruf.** Er verlangt Schlüssel und Team
// im Schlüsselbund; beides legt keine Fläche ohne Fenster an. Dass er trägt,
// prüfen `internal/scrydex`, `internal/service` und `internal/app` dort, wo er
// entsteht (§3).
package flaechen_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// **Ohne Zugang lehnt jede Fläche ab — und niemand fragt hinaus.**
func TestDerAnbieterstandOhneZugangUeberJedeFlaeche(t *testing.T) {
	var anfragen atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			anfragen.Add(1)
			schreiben.WriteHeader(http.StatusOK)
		}))
	t.Cleanup(server.Close)

	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	fuerAlleDreiMit(t, "der Kontostand ohne Zugang",
		[3]string{"scrydex braucht", "scrydex braucht", "scrydex braucht"},
		func() string {
			_, klage, _ := kommandozeile(t, umgebung, "anbieterstand", "scrydex")

			return klage
		},
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.anbieterstand",
				map[string]any{"dienst": "scrydex"})
			if !istFehler {
				t.Fatalf("MCP antwortete ohne Zugang mit Erfolg: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet, "/anbieter/scrydex/stand", "")
			if status < http.StatusBadRequest {
				t.Fatalf("der Dienst antwortete mit %d (%s)", status, rumpf)
			}

			return rumpf
		})

	if gezaehlt := anfragen.Load(); gezaehlt != 0 {
		t.Errorf("ohne Zugang gingen %d Anfragen hinaus, erwartet keine", gezaehlt)
	}
}

// **Und der Fremdverlauf ebenso.**
func TestDerFremdverlaufOhneZugangUeberJedeFlaeche(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	fuerAlleDreiMit(t, "der Fremdverlauf ohne Zugang",
		[3]string{"scrydex braucht", "scrydex braucht", "scrydex braucht"},
		func() string {
			_, klage, _ := kommandozeile(t, umgebung,
				"fremdverlauf", "scrydex", "11111111-1111-4111-8111-111111111111")

			return klage
		},
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "schreiben.fremdverlauf",
				map[string]any{
					"quelle":   "scrydex",
					"printing": "11111111-1111-4111-8111-111111111111",
				})
			if !istFehler {
				t.Fatalf("MCP antwortete ohne Zugang mit Erfolg: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodPost,
				"/drucke/11111111-1111-4111-8111-111111111111/fremdverlauf?quelle=scrydex", "")
			if status < http.StatusBadRequest {
				t.Fatalf("der Dienst antwortete mit %d (%s)", status, rumpf)
			}

			return rumpf
		})
}

// **Ein Dienst, den es nicht gibt, wird überall gleich abgewiesen.**
func TestEinUnbekannterAnbieterWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	if _, klage, _ := kommandozeile(t, umgebung,
		"anbieterstand", "cardmarket"); !strings.Contains(klage, "cardmarket") {
		t.Errorf("die Kommandozeile klagt: %s", klage)
	}

	aus, istFehler := mcpRuf(t, umgebung, "lesen.anbieterstand",
		map[string]any{"dienst": "cardmarket"})
	if !istFehler || !strings.Contains(aus, "cardmarket") {
		t.Errorf("MCP antwortet: %s", aus)
	}

	if status, rumpf := an.ruf(t, http.MethodGet,
		"/anbieter/cardmarket/stand", ""); status < http.StatusBadRequest {
		t.Errorf("der Dienst antwortet mit %d (%s)", status, rumpf)
	}
}
