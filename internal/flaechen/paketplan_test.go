// **Plan, Ersetzen und Zurück — an jeder Fläche** — DOM-01.AK-02 und AK-03
// (#1696).
//
// Derselbe Ablauf über Kommandozeile, MCP und Dienst: eine Fassung ist
// eingebunden, der Plan nennt den Tausch, das Ersetzen legt die neue ein,
// das Zurück holt die alte — und ein Zurück ohne vorige Fassung scheitert
// überall (§16).
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fassung legt ein Stilpaket in einer bestimmten Fassung an.
func fassung(t *testing.T, name, version string) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\nshort name: Probe\nversion: " + version + "\n"
	if err := os.WriteFile(filepath.Join(ort, "style"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return ort
}

// flaechenablauf sind die drei Schritte einer Fläche.
type flaechenablauf struct {
	planen     func(t *testing.T, quelle string) string
	ersetzen   func(t *testing.T, quelle string) string
	zurueck    func(t *testing.T, name string) string
	ohneFehler func(t *testing.T) bool
}

// pruefeTausch fährt den Ablauf und hält, was jede Fläche sagen muss.
func pruefeTausch(t *testing.T, flaeche string, ablauf flaechenablauf) {
	t.Helper()

	neu := fassung(t, "probe-tausch.mse-style", "2026-09-01")

	var geplant struct {
		Plan struct {
			Vorgang, Von, Zu string
			Moeglich         bool
		} `json:"plan"`
	}

	if roh := ablauf.planen(t, neu); json.Unmarshal([]byte(roh), &geplant) != nil ||
		geplant.Plan.Vorgang != "ersetzen" || geplant.Plan.Von != "2026-08-30" ||
		geplant.Plan.Zu != "2026-09-01" || !geplant.Plan.Moeglich {
		t.Errorf("%s: der Plan %s", flaeche, roh)
	}

	if roh := ablauf.ersetzen(t, neu); !strings.Contains(roh, "2026-09-01") {
		t.Errorf("%s: ersetzen %s", flaeche, roh)
	}

	if roh := ablauf.zurueck(t, "probe-tausch"); !strings.Contains(roh, "2026-08-30") {
		t.Errorf("%s: zurück %s", flaeche, roh)
	}

	if ablauf.ohneFehler(t) {
		t.Errorf("%s: ein Zurück ohne vorige Fassung ging durch", flaeche)
	}
}

func TestPlanErsetzenUndZurueckAnJederFlaeche(t *testing.T) {
	// eingebunden gibt eine Umgebung, in der Fassung 1 eingebunden ist.
	eingebunden := func(t *testing.T) []string {
		t.Helper()

		umgebung := mitEigenerSaat(t)
		mussKommandozeile(t, umgebung, "paket-installieren", fassung(t, "probe-tausch.mse-style", "2026-08-30"))

		return umgebung
	}

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := eingebunden(t)

		pruefeTausch(t, "Kommandozeile", flaechenablauf{
			planen: func(t *testing.T, quelle string) string {
				return mussKommandozeile(t, umgebung, "paket-plan", "einbinden", quelle, "--json")
			},
			ersetzen: func(t *testing.T, quelle string) string {
				return mussKommandozeile(t, umgebung, "paket-ersetzen", quelle, "--json")
			},
			zurueck: func(t *testing.T, name string) string {
				return mussKommandozeile(t, umgebung, "paket-zurueck", "mse-style", name, "--json")
			},
			ohneFehler: func(t *testing.T) bool {
				_, _, code := kommandozeile(t, umgebung, "paket-zurueck", "mse-game", "probe")

				return code == 0
			},
		})
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := eingebunden(t)

		pruefeTausch(t, "MCP", flaechenablauf{
			planen: func(t *testing.T, quelle string) string {
				return mussMCP(t, umgebung, "lesen.paket.plan", map[string]any{"vorgang": "einbinden", "quelle": quelle})
			},
			ersetzen: func(t *testing.T, quelle string) string {
				return mussMCP(t, umgebung, "schreiben.paket.ersetzen", map[string]any{"quelle": quelle})
			},
			zurueck: func(t *testing.T, name string) string {
				return mussMCP(t, umgebung, "schreiben.paket.zurueck", map[string]any{"art": "mse-style", "name": name})
			},
			ohneFehler: func(t *testing.T) bool {
				_, istFehler := mcpRuf(t, umgebung, "schreiben.paket.zurueck", map[string]any{"art": "mse-game", "name": "probe"})

				return !istFehler
			},
		})
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, eingebunden(t))

		pruefeTausch(t, "REST", flaechenablauf{
			planen: func(t *testing.T, quelle string) string {
				return mussREST(t, an, http.MethodGet, "/pakete/plan?vorgang=einbinden&quelle="+url.QueryEscape(quelle), "")
			},
			ersetzen: func(t *testing.T, quelle string) string {
				return mussREST(t, an, http.MethodPut, "/pakete/ersetzt", `{"quelle":`+alsJSONText(t, quelle)+`}`)
			},
			zurueck: func(t *testing.T, name string) string {
				return mussREST(t, an, http.MethodPost, "/pakete/mse-style/"+name+"/zurueck", "")
			},
			ohneFehler: func(t *testing.T) bool {
				status, _ := an.ruf(t, http.MethodPost, "/pakete/mse-game/probe/zurueck", "")

				return status < 400
			},
		})
	})
}

// **Was einem Set an Paketen fehlt — an jeder Fläche** (DOM-01.AK-04): der
// Stil des Probesets wird gelöscht, und alle drei nennen ihn.
func TestDerPaketbedarfAnJederFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := setkennungAus(t, mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))
	mussKommandozeile(t, umgebung, "pakete") // frischt den Katalog auf
	mussKommandozeile(t, umgebung, "paket-entfernen", "mse-style", "probe-stil")
	an := starteDienst(t, umgebung)

	for flaeche, roh := range map[string]string{
		"Kommandozeile": mussKommandozeile(t, umgebung, "paketbedarf", setID, "--json"),
		"MCP":           mussMCP(t, umgebung, "lesen.set.paketbedarf", map[string]any{"setID": setID}),
		"REST":          mussREST(t, an, http.MethodGet, "/sets/"+setID+"/paketbedarf", ""),
	} {
		var antwort struct {
			Befunde []struct{ Braucht, Art string } `json:"befunde"`
		}

		if err := json.Unmarshal([]byte(roh), &antwort); err != nil || len(antwort.Befunde) != 1 ||
			antwort.Befunde[0].Braucht != "probe-stil.mse-style" || antwort.Befunde[0].Art != "fehlt" {
			t.Errorf("%s: %s (%v)", flaeche, roh, err)
		}
	}
}
