// **Der Bulkbezug über jede Fläche ohne Fenster** — Werkstück A3 von Phase 5.3
// (#1442).
//
// # Die Zusage
//
// Derselbe fachliche Aufruf gibt an jeder Fläche dieselbe Auskunft: wie viele
// Drucke gefragt wurden, wie viele Stände abgelegt sind, und welche Angaben
// nicht zu deuten waren (§16).
//
// **Kein Aufruf geht hinaus** — die Gegenstelle steht hier, und die Anwendung
// wird über [app.MTGJSONPreisumgebung] darauf umgelenkt.
package flaechen_test

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// dieBulkgestalt ist eine Preisdatei in der Form, die MTGJSON dokumentiert.
//
// **Die Kennung ist die, die `eigenerBestand` anlegt** — sonst fände der
// Crosswalk nichts, und der Lauf prüfte einen leeren Bestand.
const dieBulkgestalt = `{
  "meta": {"date": "2026-09-11", "version": "5.2.2"},
  "data": {
    "%s": {
      "paper": {
        "cardkingdom": {
          "currency": "USD",
          "retail": {"normal": {"2026-09-10": 4.17}}
        }
      }
    }
  }
}`

// bulkgegenstelle antwortet mit einer gepackten Preisdatei.
func bulkgegenstelle(t *testing.T, mtgjsonkennung string) *httptest.Server {
	t.Helper()

	inhalt := strings.Replace(dieBulkgestalt, "%s", mtgjsonkennung, 1)

	var puffer bytes.Buffer

	schreiber := gzip.NewWriter(&puffer)

	if _, fehler := schreiber.Write([]byte(inhalt)); fehler != nil {
		t.Fatalf("packen: %v", fehler)
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write(puffer.Bytes()) //nolint:errcheck // die Gegenstelle
		}))

	t.Cleanup(server.Close)

	return server
}

// **Ohne einen Druck mit MTGJSON-Kennung fragt keine Fläche.**
//
// # Warum das die Abnahme ist, die geht
//
// Der Crosswalk kommt aus dem Bestand, und ein frischer hat keinen Druck mit
// MTGJSON-Kennung. **Die Zusage ist damit prüfbar**: „ohne Crosswalk wird
// nicht gefragt" (§16, Satz 4) — die Datei wiegt ein Gigabyte, und wer keine
// Karte besitzt, die MTGJSON kennt, hat nichts davon.
//
// **Was das nicht prüft**, gehört daneben: den Weg **mit** Crosswalk. Ihn über
// die Flächen zu fahren verlangte einen Bestand mit Drucken, und den legt keine
// Fläche ohne Fenster an — `internal/service` und `internal/mtgjsonpreise`
// prüfen ihn dort, wo er entsteht (§3).
func TestDerBulkbezugUeberJedeFlaeche(t *testing.T) {
	server := bulkgegenstelle(t, "22222222-2222-4222-8222-222222222222")

	umgebung := append(eigenerBestand(t),
		app.MTGJSONPreisumgebung+"="+server.URL,
		"HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	// **Je Fläche ein eigener Kern**: dieselbe Auskunft in zwei Schreibweisen
	// — die Kommandozeile und MCP geben eingerücktes JSON, der Dienst enges.
	fuerAlleDreiMit(t, "der Bulkbezug",
		[3]string{`"drucke": 0`, `"drucke": 0`, `"drucke":0`},
		func() string {
			return mussKommandozeile(t, umgebung,
				"preise-bulk", "mtgjson", "heute", "--json")
		},
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "schreiben.preise.bulk",
				map[string]any{"quelle": "mtgjson", "umfang": "heute"})
			if istFehler {
				t.Fatalf("schreiben.preise.bulk: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodPost,
				"/preise/bulk/mtgjson/heute", "")
			if status != http.StatusOK {
				t.Fatalf("POST preise/bulk: %d (%s)", status, rumpf)
			}

			return rumpf
		})
}

// **Eine Quelle, die es nicht gibt, wird überall gleich abgewiesen.**
func TestEineUnbekanntePreisquelleWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	if _, klage, _ := kommandozeile(t, umgebung,
		"preise-bulk", "cardmarket", "heute"); !strings.Contains(klage, "cardmarket") {
		t.Errorf("die Kommandozeile klagt: %s", klage)
	}

	aus, istFehler := mcpRuf(t, umgebung, "schreiben.preise.bulk",
		map[string]any{"quelle": "cardmarket", "umfang": "heute"})
	if !istFehler || !strings.Contains(aus, "cardmarket") {
		t.Errorf("MCP antwortet: %s", aus)
	}

	if status, rumpf := an.ruf(t, http.MethodPost,
		"/preise/bulk/cardmarket/heute", ""); status < http.StatusBadRequest {
		t.Errorf("der Dienst antwortet mit %d (%s)", status, rumpf)
	}
}
