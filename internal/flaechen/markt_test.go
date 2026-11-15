// **Der Marktplatz über jede Fläche ohne Fenster** — Werkstück A4 von
// Phase 5.3 (#1443).
//
// # Die Zusage
//
// Derselbe fachliche Aufruf gibt an jeder Fläche dieselbe Auskunft — und
// **derselbe fachliche Fehler** dieselbe Ablehnung (§16).
//
// # Was hier gemessen wird und was nicht
//
// **Gemessen wird die Ablehnung**, und zwar mit einem Spion daneben: ohne
// Zugangsmarke geht **keine einzige Verbindung** hinaus. Das ist die Abnahme
// aus dem Issue („Token fehlt … erzeugt null Netzverbindungen"), und sie ist
// hier prüfbar, weil ein frischer Bestand keinen Zugang hat.
//
// **Nicht gemessen wird der geglückte Abruf.** Er verlangt einen Druck mit
// Scryfall-Alias und eine Marke im Schlüsselbund; beides legt keine Fläche
// ohne Fenster an. Dass er trägt, prüfen `internal/service` und
// `internal/cardtrader` dort, wo er entsteht (§3) — und dass dabei nichts
// geschrieben wird, weist der Spion in `internal/cardtrader` nach.
package flaechen_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ohne Zugangsmarke lehnt jede Fläche ab — und niemand fragt hinaus.**
func TestDerMarktplatzOhneZugangUeberJedeFlaeche(t *testing.T) {
	var anfragen atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			anfragen.Add(1)
			schreiben.WriteHeader(http.StatusOK)
		}))
	t.Cleanup(server.Close)

	// **Ein eigenes Zuhause, und die Adresse umgelenkt** — beides einzeln
	// reicht nicht (§3): ohne das Zuhause sähe der Lauf den Schlüsselbund
	// dieses Rechners, ohne die Adresse ginge er nach api.cardtrader.com.
	umgebung := append(eigenerBestand(t),
		app.CardTraderUmgebung+"="+server.URL,
		"HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	fuerAlleDreiMit(t, "der Marktabruf ohne Zugang",
		[3]string{
			"cardtrader braucht marke", "cardtrader braucht marke",
			"cardtrader braucht marke",
		},
		func() string {
			_, klage, _ := kommandozeile(t, umgebung,
				"markt", "cardtrader", "11111111-1111-4111-8111-111111111111")

			return klage
		},
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.marktangebote",
				map[string]any{
					"quelle":   "cardtrader",
					"printing": "11111111-1111-4111-8111-111111111111",
				})
			if !istFehler {
				t.Fatalf("MCP antwortete ohne Zugang mit Erfolg: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet,
				"/drucke/11111111-1111-4111-8111-111111111111/angebote?quelle=cardtrader", "")
			if status < http.StatusBadRequest {
				t.Fatalf("der Dienst antwortete mit %d (%s)", status, rumpf)
			}

			return rumpf
		})

	if gezaehlt := anfragen.Load(); gezaehlt != 0 {
		t.Errorf("ohne Zugangsmarke gingen %d Anfragen hinaus, erwartet keine",
			gezaehlt)
	}
}

// **Ein Marktplatz, den es nicht gibt, wird überall gleich abgewiesen.**
func TestEinUnbekannterMarktplatzWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	if _, klage, _ := kommandozeile(t, umgebung,
		"markt", "cardmarket", "abc"); !strings.Contains(klage, "cardmarket") {
		t.Errorf("die Kommandozeile klagt: %s", klage)
	}

	aus, istFehler := mcpRuf(t, umgebung, "lesen.marktangebote",
		map[string]any{"quelle": "cardmarket", "printing": "abc"})
	if !istFehler || !strings.Contains(aus, "cardmarket") {
		t.Errorf("MCP antwortet: %s", aus)
	}

	if status, rumpf := an.ruf(t, http.MethodGet,
		"/drucke/abc/angebote?quelle=cardmarket", ""); status < http.StatusBadRequest {
		t.Errorf("der Dienst antwortet mit %d (%s)", status, rumpf)
	}
}

// **Und eine Ausführung, die es nicht gibt, ebenso.**
//
// **Die Gegenprobe zum Lauf darüber**: dort wird die **Quelle** geprüft, hier
// eine Angabe dahinter — ohne diesen Lauf bliebe offen, ob der Wunsch
// überhaupt angesehen wird, bevor jemand nach draußen greift.
func TestEineUnbekannteAusfuehrungWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	if _, klage, _ := kommandozeile(t, umgebung, "markt", "cardtrader", "abc",
		"--finish=glaenzend"); !strings.Contains(klage, "glaenzend") {
		t.Errorf("die Kommandozeile klagt: %s", klage)
	}

	aus, istFehler := mcpRuf(t, umgebung, "lesen.marktangebote",
		map[string]any{"quelle": "cardtrader", "printing": "abc", "finish": "glaenzend"})
	if !istFehler || !strings.Contains(aus, "glaenzend") {
		t.Errorf("MCP antwortet: %s", aus)
	}

	status, rumpf := an.ruf(t, http.MethodGet,
		"/drucke/abc/angebote?quelle=cardtrader&finish=glaenzend", "")
	if status < http.StatusBadRequest || !strings.Contains(rumpf, "glaenzend") {
		t.Errorf("der Dienst antwortet mit %d (%s)", status, rumpf)
	}
}
