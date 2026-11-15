// Eine Gegenstelle, die einen **Cloudanbieter** nachstellt — Werkstück A14
// von Phase 5.3
// (#1453).
//
// # Warum es sie neben der Laufzeit gibt
//
// **Weil „nur lokal" nur an einem Anbieter zu messen ist, der nicht lokal
// ist.** Die Probelaufzeit nebenan rechnet auf diesem Rechner — sie bleibt
// unter jeder Ortsregel erlaubt, und ein Lauf über sie sagt nichts über
// einen Entzug.
//
// # Und warum sie zählt
//
// **Weil der Abnahmesatz „greift im nächsten Aufruf VOR Request" lautet.**
// Eine Absage sagt nur, dass abgewiesen wurde — nicht, ob die Anfrage vorher
// schon draußen war. Der Zähler unterscheidet beides, und ohne ihn wäre der
// Lauf grün, auch wenn der Riegel **hinter** dem Ruf säße (§10).
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anthropic"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// wolkenmodell ist das Modell, das diese Gegenstelle führt.
const wolkenmodell = "probe-wolke-1"

// wolkenstelle ist der nachgestellte Cloudanbieter samt seinem Zähler.
type wolkenstelle struct {
	adresse string

	// gefragt zählt **jede** Anfrage, gleich auf welchem Weg.
	//
	// **Jede und nicht nur die Modelle**: gezählt wird, ob dieser Rechner
	// mit dem Anbieter gesprochen hat. Ein Zähler je Weg beantwortete eine
	// engere Frage, als der Abnahmesatz stellt.
	gefragt atomic.Int64

	// beauftragt zählt die Rufe an `/v1/messages` — also das, was Geld
	// kostet.
	//
	// **Neben [wolkenstelle.gefragt] und nicht darin**: „der Rechner hat
	// gesprochen" und „es ist etwas bezahlt worden" sind zwei Fragen, und
	// die zweite ist die, um die es bei einer Kostenfreigabe geht.
	beauftragt atomic.Int64
}

// starteWolke stellt einen Cloudanbieter hin, der ein Modell führt.
func starteWolke(t *testing.T) *wolkenstelle {
	t.Helper()

	stelle := &wolkenstelle{}

	dienst := httptest.NewServer(http.HandlerFunc(stelle.bedienen))
	t.Cleanup(dienst.Close)

	stelle.adresse = dienst.URL

	return stelle
}

// bedienen beantwortet die Modellliste und zählt mit.
func (w *wolkenstelle) bedienen(schreiber http.ResponseWriter, anfrage *http.Request) {
	w.gefragt.Add(1)

	if anfrage.URL.Path == anthropic.Nachrichtenweg {
		w.beauftragt.Add(1)

		schreiber.Header().Set("Content-Type", "application/json")

		//nolint:errcheck // eine Gegenstelle im Arbeitsspeicher
		_ = json.NewEncoder(schreiber).Encode(map[string]any{
			"id": "msg-1", "model": wolkenmodell, "stop_reason": "end_turn",
			"content": []map[string]any{{
				"type": "text", "text": `{"name":"Aus der Wolke"}`,
			}},
		})

		return
	}

	if anfrage.URL.Path != anthropic.Modellweg {
		http.NotFound(schreiber, anfrage)

		return
	}

	schreiber.Header().Set("Content-Type", "application/json")

	//nolint:errcheck // eine Gegenstelle im Arbeitsspeicher
	_ = json.NewEncoder(schreiber).Encode(map[string]any{
		"data": []map[string]any{{
			"id": wolkenmodell, "display_name": "Probe Wolke",
			"created_at": "2026-01-01T00:00:00Z",
			"capabilities": map[string]any{
				"image_input":        map[string]any{"supported": false},
				"structured_outputs": map[string]any{"supported": true},
				"pdf_input":          map[string]any{"supported": false},
				"batch":              map[string]any{"supported": false},
			},
			"max_input_tokens": 1000, "max_tokens": 1000,
		}},
		"has_more": false, "last_id": "",
	})
}

// anfragen sagt, wie oft dieser Rechner mit dem Anbieter gesprochen hat.
func (w *wolkenstelle) anfragen() int64 { return w.gefragt.Load() }

// auftraege sagt, wie oft etwas bezahlt worden wäre.
func (w *wolkenstelle) auftraege() int64 { return w.beauftragt.Load() }

// eineEingerichteteWolke gibt eine Umgebung mit eingeschaltetem Cloudanbieter.
//
// **Mit Schlüssel**, denn ohne einen weist das Blatt jede Anfrage ab, bevor
// sie eine wird ([anthropic.Zugang.Fehlt]) — und dann bliebe der Zähler auch
// dann bei null, wenn der Riegel gar nicht griffe. Das wäre eine Gegenprobe,
// die nichts misst (§10).
func eineEingerichteteWolke(t *testing.T) ([]string, *wolkenstelle) {
	t.Helper()

	// **Mit eigener Saat**, denn ein Vorschlag baut sein Schema aus den
	// tatsächlichen Feldern des Pakets — ohne `probe.mse-game` scheitert er,
	// bevor irgendein Anbieter gefragt wird, und der Lauf prüfte dann das
	// Paketverzeichnis statt der Freigabe (§10).
	//
	// **Der eigene Bund kommt dazu**: ohne ihn griffe der Lauf in den
	// Schlüsselbund dieses Rechners (§3).
	bundort := probebund(t)
	umgebung := append(mitEigenerSaat(t), schluesselbund.Bundort+"="+bundort)

	wolke := starteWolke(t)

	// **Der Schlüssel geht in den Bund und nicht in eine Einstellung.**
	//
	// **Und von hier aus und nicht über eine Fläche**: es gibt keine. Ein
	// Zugang wird am Fenster hinterlegt (`App.ZugangSetzen`); Kommandozeile,
	// MCP und Dienst haben dafür keinen Befehl. Das ist eine Lücke in „§16,
	// dieselbe Fähigkeit überall" — sie steht hier, damit sie jemand liest,
	// und nicht als Fußnote.
	if fehler := schluesselbund.NeuerBund(bundort).Schreibe(t.Context(),
		app.Schluesselbunddienst,
		domain.ZugangKontoFuer(domain.DienstAnthropic,
			domain.ZugangSchluessel, wolke.adresse),
		"probe-schluessel"); fehler != nil {
		t.Fatalf("den Schlüssel ablegen: %v", fehler)
	}

	mussKommandozeile(t, umgebung, "einstellungen", "setzen",
		domain.DienstAdresse(domain.DienstAnthropic), wolke.adresse)
	mussKommandozeile(t, umgebung, "einstellungen", "setzen",
		domain.DienstEinstellung(domain.DienstAnthropic), domain.DienstAn)

	return umgebung, wolke
}

// **Ein entzogener Cloudscope greift im nächsten Ruf — vor der Anfrage.**
//
// # Was der Lauf in dieser Reihenfolge belegt
//
//  1. Der Prozess **erreicht** den Anbieter wirklich — sonst sagte Schritt 3
//     nichts.
//  2. Der Entzug geschieht **von außen**, während der Prozess lebt.
//  3. Der nächste Ruf desselben Prozesses lässt den Zähler stehen.
//
// **Ohne den ersten Schritt wäre der dritte wertlos.** Ein Zähler, der bei
// null bleibt, weil nie etwas ging, sieht aus wie einer, den ein Riegel
// aufgehalten hat.
//
// # Drei Gegenproben, und erst die dritte hat es gezeigt
//
// **Es sind zwei Schranken, und jede hält eine andere Hälfte des Satzes.**
// Gemessen am 2026-09-14, jede einzeln entschärft:
//
//	Transportriegel aus (`App.pruefeCloudanbieter`)   → GRÜN
//	Wirksamkeitsschranke aus (`domain.mitNurLokal`)   → rot, aber NUR die Meldung
//	beide aus                                          → rot, und eine Anfrage geht hinaus
//
// **Die erste allein ändert nichts**, weil die zweite davor liegt: unter
// »nur lokal« ist der **wirksame** Wert jedes `dienst.`-Schlüssels `aus`, und
// `App.standVon` fragt einen abgeschalteten Anbieter gar nicht erst. Die
// zweite allein lässt den Zähler stehen — dann hält der Transportriegel —,
// aber die Antwort sagt nicht mehr, **wer** abgeschaltet hat: sie meldet
// „unerreichbar" statt der Richtlinie, und das schickte jemanden an den
// falschen Schalter (§8).
//
// **Deshalb prüft dieser Lauf beides**: den Zähler und den Befund. Eine
// Absage ohne Grund und ein Grund ohne Absage sind zwei verschiedene halbe
// Zusagen.
func TestEinEntzogenerCloudscopeGreiftImNaechstenRuf(t *testing.T) {
	t.Parallel()

	umgebung, wolke := eineEingerichteteWolke(t)
	agent := starteMCPProzess(t, umgebung)

	// **1. Der Prozess spricht mit dem Anbieter.**
	if katalog := agent.muss("lesen.ki.modelle", map[string]any{}); !strings.Contains(
		katalog, wolkenmodell) {
		t.Fatalf("das Wolkenmodell steht nicht im Katalog: %s", katalog)
	}

	vorher := wolke.anfragen()
	if vorher == 0 {
		t.Fatalf("es ging keine Anfrage hinaus — der Entzug wäre nicht zu messen")
	}

	// **2. Der Entzug — von außen, mitten in der Sitzung.**
	mussKommandozeile(t, umgebung, "einstellungen", "setzen",
		domain.RichtlinieNurLokal, domain.NurLokalAn)

	// **3. Und der nächste Ruf geht nicht mehr hinaus.**
	katalog := agent.muss("lesen.ki.modelle", map[string]any{})

	if nachher := wolke.anfragen(); nachher != vorher {
		t.Errorf("nach dem Entzug gingen %d Anfragen hinaus", nachher-vorher)
	}

	if strings.Contains(katalog, wolkenmodell) {
		t.Errorf("das Wolkenmodell steht nach dem Entzug noch im Katalog: %s", katalog)
	}

	// **Und die Antwort sagt, WER abgeschaltet hat.**
	//
	// **Nicht nur »aus«**: der Schalter des Anbieters steht weiter auf `an`,
	// und wer nur ihn läse, suchte den Fehler dort. Der Befund nennt die
	// Richtlinie — dieselbe Zusage wie in `app.vonDerRichtlinieAbgeschaltet`,
	// hier über die ganze Naht gemessen.
	if !strings.Contains(katalog, `"nurLokal": true`) {
		t.Errorf("der Katalog nennt die Regel nicht: %s", katalog)
	}

	if !strings.Contains(katalog, "»nur lokal« schaltet jeden fremden Dienst ab") {
		t.Errorf("der Katalog sagt nicht, wer abgeschaltet hat: %s", katalog)
	}
}
