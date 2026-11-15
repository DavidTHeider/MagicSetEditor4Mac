// Eine Gegenstelle, die eine lokale KI-Laufzeit nachstellt — Werkstück A14
// von Phase 5.3
// (#1453).
//
// # Warum sie hier steht und nicht im Blatt
//
// **Dieselbe Begründung wie bei der Scryfall-Gegenstelle**: die Flächenläufe
// starten `mse` und `mse-mcp` als **eigene Prozesse**, und ein Doppel im
// selben Programm erreicht sie nicht. Die Adresse kommt über die Einstellung
// `ki.adresse`, also über denselben Weg, den ein Mensch am Fenster geht.
//
// # Und warum sie zählt
//
// **Weil „es ging kein Ruf hinaus" nur an einem Zähler zu belegen ist.** Eine
// Fehlermeldung sagt, dass etwas abgewiesen wurde — nicht, ob es **vor** oder
// **nach** der Anfrage geschah (§10, „die Reichweite einer Messung"). Der
// Zähler unterscheidet beides.
//
// # Was sie nicht ist
//
// **Kein Ollama.** Sie antwortet auf die vier Wege, die `internal/ollama`
// geht, mit dem, was der Fall braucht — nachgestellt aus den Formen in
// `internal/ollama`, nicht abgeschrieben (§11).
package flaechen_test

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
)

// probemodell ist das Modell, das diese Gegenstelle führt.
//
// **Ohne Cloudsuffix**, denn davon hängt ab, wo es rechnet — und damit, ob
// der Auftrag Geld kostet (`internal/ollama`, `ortVon`).
const probemodell = "probemodell:8b"

// probeeinbetter ist das Modell, mit dem der Index gebaut wird.
//
// **Ein zweites und nicht dasselbe**: Ollama führt Einbetter und
// Textgeneratoren getrennt, und `Vermessen` verlangt `embedding` in den
// Fähigkeiten. Ein Modell, das beides meldete, gäbe es so nicht.
const probeeinbetter = "probeeinbetter"

// dieProbedimension ist die Länge der gelieferten Vektoren.
//
// **Acht Stellen und nicht 384.** Der Vertrag prüft **Länge und
// Normalisierung**, nicht die Zahl selbst; acht reichen dafür und halten die
// Antwort lesbar.
const dieProbedimension = 8

// laufzeitstelle ist die nachgestellte Laufzeit samt ihren Zählern.
type laufzeitstelle struct {
	// adresse ist, was in `ki.adresse` gehört.
	adresse string

	// gefragt zählt die Rufe an `/api/chat` — also die **Aufträge**.
	gefragt atomic.Int64

	// beruehrt zählt **jede** Anfrage, gleich auf welchem Weg.
	//
	// **Neben [laufzeitstelle.gefragt] und nicht darin**: das eine sagt, wie
	// oft ein Modell gerechnet hat, das andere, ob dieser Rechner die
	// Laufzeit überhaupt angefasst hat. „Es ging nichts hinaus" ist die
	// zweite Frage.
	beruehrt atomic.Int64

	// gesehen hält den letzten Rumpf, den ein Auftrag geschickt hat.
	gesehen atomic.Value

	// antwort ist, was das Modell sagt.
	antwort string

	// eingebettet zählt die Rufe an `/api/embed`.
	eingebettet atomic.Int64

	// schloss hütet [laufzeitstelle.warte].
	//
	// **Weil sie aus dem Lauf gesetzt und aus dem Bediener gelesen wird** —
	// zwei Goroutinen, und der Detektor sagt es beim ersten vollen Lauf.
	schloss sync.Mutex

	// warte macht jede Einbettung langsam.
	//
	// # Warum langsam und nicht angehalten
	//
	// **Ein angehaltener Aufbau erzeugt eine Lage, die es nicht gibt.** Er
	// hielte seine Schreibtransaktion über die ganze Zeit, und ein zweiter
	// Prozess wartete auf SQLite statt auf die Absprache über den Index —
	// gemessen am 2026-09-14, zweimal in die Zeitgrenze gelaufen.
	//
	// **Verlangsamt dagegen läuft er wirklich**: er bündelt, schreibt, gibt
	// frei und nimmt sich das nächste Bündel.
	warte time.Duration
}

// starteLaufzeit stellt eine Laufzeit hin, die genau eine Antwort gibt.
func starteLaufzeit(t *testing.T, antwort string) *laufzeitstelle {
	t.Helper()

	stelle := &laufzeitstelle{antwort: antwort}

	dienst := httptest.NewServer(http.HandlerFunc(stelle.bedienen))
	t.Cleanup(dienst.Close)

	stelle.adresse = dienst.URL

	return stelle
}

// bedienen beantwortet die vier Wege.
func (l *laufzeitstelle) bedienen(schreiber http.ResponseWriter, anfrage *http.Request) {
	l.beruehrt.Add(1)

	switch anfrage.URL.Path {
	case ollama.Fassungsweg:
		l.sende(schreiber, map[string]any{"version": "0.12.3"})
	case ollama.Modellweg:
		l.sende(schreiber, map[string]any{"models": []map[string]any{
			einModelleintrag(probemodell, "sha256:0123456789abcdef"),
			einModelleintrag(probeeinbetter, "sha256:fedcba9876543210"),
		}})
	case ollama.Schauweg:
		l.schau(schreiber, anfrage)
	case ollama.Einbettungsweg:
		l.einbetten(schreiber, anfrage)
	case ollama.Redeweg:
		l.rede(schreiber, anfrage)
	default:
		http.NotFound(schreiber, anfrage)
	}
}

// einModelleintrag baut eine Zeile für `/api/tags`.
func einModelleintrag(name, digest string) map[string]any {
	return map[string]any{
		"name": name, "digest": digest, "size": 42,
		"details": map[string]any{
			"format": "gguf", "family": "probe",
			"parameter_size": "8B", "quantization_level": "Q4_0",
		},
	}
}

// schau beantwortet `/api/show` — **je Modell verschieden**.
//
// **Die Angaben stammen aus llama.cpps Schlüsseln** (`tokenizer.ggml.model`,
// `<arch>.pooling_type`), so wie `internal/ollama` sie nachgelesen hat. Ein
// Einbetter ohne sie bekäme keinen Vertrag — und genau das prüft A15.
func (l *laufzeitstelle) schau(schreiber http.ResponseWriter, anfrage *http.Request) {
	var wunsch struct {
		Modell string `json:"model"`
	}

	//nolint:errcheck // eine Gegenstelle im Arbeitsspeicher
	_ = json.NewDecoder(anfrage.Body).Decode(&wunsch)

	if wunsch.Modell != probeeinbetter {
		// **Mit `num_ctx`**: das Kontextfenster je Modell (AI-01.AK-02, #1742).
		l.sende(schreiber, map[string]any{
			"license": "Probelizenz", "capabilities": []string{"completion"},
			"parameters": "num_ctx 4096",
		})

		return
	}

	l.sende(schreiber, map[string]any{
		"license": "Probelizenz", "capabilities": []string{"embedding"},
		"model_info": map[string]any{
			"tokenizer.ggml.model": "bert",
			"probe.pooling_type":   1,
		},
	})
}

// einbetten gibt je Text einen Vektor — und wartet, wenn ein Halt steht.
func (l *laufzeitstelle) einbetten(schreiber http.ResponseWriter, anfrage *http.Request) {
	var wunsch struct {
		Eingaben []string `json:"input"`
	}

	//nolint:errcheck // eine Gegenstelle im Arbeitsspeicher
	_ = json.NewDecoder(anfrage.Body).Decode(&wunsch)

	l.eingebettet.Add(1)

	// **Die Vertragsmessung läuft ohne Pause durch.**
	//
	// **Sonst dauerte der Aufbau schon, bevor er beginnt**: `Klient.Raum`
	// misst den Einbettungsvertrag an **einem** Text, bevor er die Sperre
	// nimmt ([ollama.Klient.Vermessen]).
	//
	// **Je Text und nicht je Ruf**: ein Aufbau bündelt, und eine Pause je
	// Bündel hinge an der Bündelgröße statt an der Arbeit.
	if warte := l.verzug(); warte > 0 && !nurDieVertragsprobe(wunsch.Eingaben) {
		time.Sleep(warte * time.Duration(len(wunsch.Eingaben)))
	}

	vektoren := make([][]float64, 0, len(wunsch.Eingaben))
	for _, text := range wunsch.Eingaben {
		vektoren = append(vektoren, vektorFuer(text))
	}

	l.sende(schreiber, map[string]any{"embeddings": vektoren})
}

// nurDieVertragsprobe sagt, ob dieser Ruf die Messung aus A6 ist.
//
// **Ein Text und dieser Text**: `internal/ollama` misst an einem festen Wort,
// und ein Korpusabschnitt trägt mehr als das.
func nurDieVertragsprobe(eingaben []string) bool {
	return len(eingaben) == 1 && eingaben[0] == "Blitzschlag"
}

// vektorFuer gibt einen **normalisierten** Vektor, der zum Text passt.
//
// # Warum er aus dem Text entsteht und nicht gewürfelt wird
//
// **Weil sonst nichts ähnlich wäre.** Ein Index aus Zufallsvektoren gäbe eine
// Trefferliste ohne Ordnung, und ein Lauf darüber prüfte, dass etwas
// zurückkommt — nicht, dass das Richtige zurückkommt (§10).
//
// **Und normalisiert, weil der Vertrag es zusagt** ([ollama.Klient.Vermessen]
// misst es nach): ohne Normalisierung stünde im Raum ein anderes
// Abstandsmaß, und der Lauf prüfte einen Vertrag, den er nicht wollte.
func vektorFuer(text string) []float64 {
	roh := make([]float64, dieProbedimension)

	for stelle, zeichen := range []byte(strings.ToLower(text)) {
		roh[stelle%dieProbedimension] += float64(zeichen%16) + 1
	}

	laenge := 0.0
	for _, wert := range roh {
		laenge += wert * wert
	}

	laenge = math.Sqrt(laenge)
	if laenge == 0 {
		roh[0], laenge = 1, 1
	}

	for stelle := range roh {
		roh[stelle] /= laenge
	}

	return roh
}

// rede beantwortet einen Auftrag und zählt ihn.
func (l *laufzeitstelle) rede(schreiber http.ResponseWriter, anfrage *http.Request) {
	l.gefragt.Add(1)

	var rumpf map[string]any
	if fehler := json.NewDecoder(anfrage.Body).Decode(&rumpf); fehler == nil {
		l.gesehen.Store(rumpf)
	}

	l.sende(schreiber, map[string]any{
		"model": probemodell,
		"message": map[string]any{
			"role": "assistant", "content": l.antwort,
		},
		"done": true, "done_reason": "stop",
	})
}

// sende schreibt eine Antwort als JSON.
func (l *laufzeitstelle) sende(schreiber http.ResponseWriter, wert any) {
	schreiber.Header().Set("Content-Type", "application/json")

	//nolint:errcheck // eine Gegenstelle im Arbeitsspeicher
	_ = json.NewEncoder(schreiber).Encode(wert)
}

// rumpftext gibt den letzten Auftragsrumpf als Text zurück.
//
// **Als Text und nicht als Verbund**: ein Lauf fragt danach, ob ein Satz
// **mitging**, und die Stelle im Baum, an der er steht, ist dabei nicht die
// Frage. [laufzeitstelle.gesehen] hält eine `map[string]any` — wer sie für
// eine Zeichenkette hält, bekommt von `Load` eine leere und merkt nichts.
func (l *laufzeitstelle) rumpftext(t *testing.T) string {
	t.Helper()

	rumpf, ist := l.gesehen.Load().(map[string]any)
	if !ist {
		t.Fatalf("diese Laufzeit hat keinen Auftrag gesehen")
	}

	als, fehler := json.Marshal(rumpf)
	if fehler != nil {
		t.Fatalf("den Rumpf schreiben: %v", fehler)
	}

	return string(als)
}

// auftraege sagt, wie viele Aufträge diese Laufzeit gesehen hat.
func (l *laufzeitstelle) auftraege() int64 { return l.gefragt.Load() }

// anfragen sagt, wie oft sie überhaupt angefasst wurde.
func (l *laufzeitstelle) anfragen() int64 { return l.beruehrt.Load() }

// einbettungen sagt, wie viele Rufe an `/api/embed` gingen.
func (l *laufzeitstelle) einbettungen() int64 { return l.eingebettet.Load() }

// verzoegern lässt jede Einbettung eine Weile dauern.
func (l *laufzeitstelle) verzoegern(dauer time.Duration) {
	l.schloss.Lock()
	defer l.schloss.Unlock()

	l.warte = dauer
}

// verzug gibt die eingestellte Dauer.
func (l *laufzeitstelle) verzug() time.Duration {
	l.schloss.Lock()
	defer l.schloss.Unlock()

	return l.warte
}
