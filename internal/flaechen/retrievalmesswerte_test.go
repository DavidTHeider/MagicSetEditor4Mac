//go:build messlauf

package flaechen_test

// **Was der Messlauf zählt** — Werkstück 5 von A15, Phase 5.3
// (#1514).
//
// # Der Schnitt folgt der Frage
//
// **Dort steht, wie der Index entsteht; hier, was daran gemessen wird.**
// Zusammen liefen beide über die Zeilengrenze aus §2 — die Grenze ist der
// Anlass des Schnitts und nicht sein Grund.

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// messwert ist das Ergebnis einer Betriebsart über eine Klasse.
type messwert struct {
	Art      string  `json:"art"`
	Klasse   string  `json:"klasse"`
	Richtung string  `json:"richtung"`
	Fragen   int     `json:"fragen"`
	Recall   float64 `json:"recall"`
	MRR      float64 `json:"mrr"`
	NDCG     float64 `json:"ndcg"`
}

// messbericht ist der ganze Lauf.
type messbericht struct {
	Modell    domain.Vektorraum `json:"modell"`
	Korpus    string            `json:"korpus"`
	Frageset  string            `json:"frageset"`
	Dokumente int               `json:"dokumente"`

	// Maschine nennt, worauf gemessen wurde — das Issue verlangt es.
	Maschine string `json:"maschine"`

	// Messpunkte ist, wie viele Suchen je Betriebsart gezählt wurden.
	//
	// **Ohne die Zahl sagt ein p95 nichts**: bei zehn Messpunkten ist er der
	// zweitschlechteste Wert.
	Messpunkte int           `json:"messpunkte"`
	Bauzeit    time.Duration `json:"bauzeit"`
	Indexbytes int64         `json:"indexbytes"`

	Werte []messwert `json:"werte"`

	// Dauer ist p50 und p95 je Betriebsart, kalt und warm.
	Dauer map[string]latenz `json:"dauer"`
}

// latenz sind die zwei Perzentile.
type latenz struct {
	P50 time.Duration `json:"p50"`
	P95 time.Duration `json:"p95"`
}

// dieArten sind die drei Betriebsarten, die verglichen werden.
//
// **Drei und nicht zwei**: das Issue verlangt „FTS/BM25 gegen … Embedding-
// Modelle … und deren Hybridvarianten". Ohne die beiden Hälften einzeln wäre
// ein Hybridgewinn nicht zu beziffern.
var dieArten = []string{"lexikalisch", "semantisch", "hybrid"}

// messe fährt jede Frage in jeder Betriebsart.
func messe(
	t *testing.T, index *storage.Retrievalindex,
	einbetter *ollama.Einbetter, frageset probenfrageset,
) messbericht {
	t.Helper()

	bericht := messbericht{Dauer: map[string]latenz{}}

	for _, art := range dieArten {
		gesammelt := map[string][]service.Guete{}
		dauern := make([]time.Duration, 0, len(frageset.Fragen))

		for _, frage := range frageset.Fragen {
			begonnen := time.Now()
			treffer := sucheIn(t, index, einbetter, art, frage.Frage)
			dauern = append(dauern, time.Since(begonnen))

			schluessel := frage.Klasse + "|" + frage.Richtung
			gesammelt[schluessel] = append(gesammelt[schluessel],
				service.Bewerte(treffer, frage.Urteile))
		}

		for schluessel, einzelne := range gesammelt {
			klasse, richtung, _ := strings.Cut(schluessel, "|")
			schnitt := service.Gueteschnitt(einzelne)

			bericht.Werte = append(bericht.Werte, messwert{
				Art: art, Klasse: klasse, Richtung: richtung,
				Fragen: len(einzelne), Recall: schnitt.Recall,
				MRR: schnitt.MRR, NDCG: schnitt.NDCG,
			})
		}

		bericht.Dauer[art] = perzentile(dauern)
		bericht.Messpunkte = len(dauern)
	}

	sort.Slice(bericht.Werte, func(links, rechts int) bool {
		einer, anderer := bericht.Werte[links], bericht.Werte[rechts]
		if einer.Klasse != anderer.Klasse {
			return einer.Klasse < anderer.Klasse
		}

		if einer.Richtung != anderer.Richtung {
			return einer.Richtung < anderer.Richtung
		}

		return einer.Art < anderer.Art
	})

	return bericht
}

// dieZielsprache ist die Sprache der Dokumente, in denen gesucht wird.
//
// # Warum beide Richtungen englisch suchen
//
// **Weil „DE→EN" genau das heißt**: deutsch fragen, englische Dokumente
// durchsuchen. Der Prüfkorpus trägt beide Sprachen; ohne diese Einschränkung
// fände eine deutsche Frage die **deutschen** Dokumente — richtig, und an der
// Frage vorbei, die das Issue stellt.
//
// **Der erste Lauf hat genau das getan**: jede `de-en`-Klasse stand bei
// 0,000, und das sah wie ein Befund über das Modell aus. Es war einer über
// das Frageset.
const dieZielsprache = "en"

// sucheIn fährt eine Frage in einer Betriebsart.
//
// **Die drei Arten gehen über dieselben Nähte**: lexikalisch und semantisch
// rufen die Ablage unmittelbar, hybrid geht durch den Dienst. Ein eigener
// Nachbau der Rangfusion hier wäre eine zweite Fassung derselben Regel.
//
// **Gefragt wird weiter, als gezählt wird**, und danach auf die Zielsprache
// gefiltert — dieselbe Überlegung wie beim [service.Kandidatenfenster], nur
// von Hand, weil die beiden Hälften keinen Filter kennen.
func sucheIn(
	t *testing.T, index *storage.Retrievalindex, einbetter *ollama.Einbetter,
	art, frage string,
) []string {
	t.Helper()

	fenster := service.Guetegrenze * service.Kandidatenfenster

	switch art {
	case "lexikalisch":
		treffer, fehler := index.Lexikalisch(messkontext(t), frage, fenster)
		if fehler != nil {
			t.Fatalf("Lexikalisch: %v", fehler)
		}

		return inDerZielsprache(treffer)

	case "semantisch":
		vektoren, fehler := einbetter.Einbetten(messkontext(t),
			[]string{einbetter.Raum().Vertrag.FuerFrage(frage)})
		if fehler != nil {
			t.Fatalf("Einbetten: %v", fehler)
		}

		treffer, fehler := index.Semantisch(messkontext(t), vektoren[0], fenster)
		if fehler != nil {
			t.Fatalf("Semantisch: %v", fehler)
		}

		return inDerZielsprache(treffer)

	default:
		ergebnis, fehler := service.NeuerRetrievaldienst(index, einbetter).Suchen(
			messkontext(t), service.Suchwunsch{
				Frage: frage, Grenze: service.Guetegrenze,
				Filter: service.Retrievalfilter{
					Sprachen: []string{dieZielsprache},
				},
			})
		if fehler != nil {
			t.Fatalf("Suchen: %v", fehler)
		}

		if ergebnis.Modus != service.ModusHybrid {
			t.Fatalf("der Modus ist %q (%s)", ergebnis.Modus, ergebnis.Grund)
		}

		return service.Kennungen(ergebnis.Treffer)
	}
}

// inDerZielsprache behält, was in der gesuchten Sprache steht, und kürzt.
func inDerZielsprache(treffer []domain.Retrievaltreffer) []string {
	kennungen := make([]string, 0, service.Guetegrenze)

	for _, einzelner := range treffer {
		if einzelner.Dokument.Sprache != dieZielsprache {
			continue
		}

		kennungen = append(kennungen, einzelner.Dokument.Kennung)

		if len(kennungen) == service.Guetegrenze {
			break
		}
	}

	return kennungen
}

// perzentile gibt p50 und p95 einer Messreihe.
func perzentile(dauern []time.Duration) latenz {
	if len(dauern) == 0 {
		return latenz{}
	}

	sort.Slice(dauern, func(links, rechts int) bool {
		return dauern[links] < dauern[rechts]
	})

	return latenz{
		P50: dauern[len(dauern)*50/100],
		P95: dauern[min(len(dauern)*95/100, len(dauern)-1)],
	}
}

// indexgroesse liest, wie viel die Vektoren und der Volltext wiegen.
func indexgroesse(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	var bytes int64

	if fehler := db.QueryRowContext(messkontext(t), `
		SELECT COALESCE(SUM(LENGTH(vektor)), 0) FROM retrieval_vektoren`).
		Scan(&bytes); fehler != nil {
		t.Fatalf("die Indexgröße lesen: %v", fehler)
	}

	return bytes
}

// schreibeBericht legt das Messartefakt ab.
//
// **Als JSON neben dem Frageset und als Satz im Protokoll.** Das JSON ist das
// Artefakt, das das Issue verlangt; der Satz ist für den, der zusieht.
func schreibeBericht(t *testing.T, bericht messbericht) {
	t.Helper()

	roh, fehler := json.MarshalIndent(bericht, "", "  ")
	if fehler != nil {
		t.Fatalf("den Bericht schreiben: %v", fehler)
	}

	// **Ohne Doppelpunkt im Dateinamen** (AI-02, #1743): ein Alias mit Marke
	// — `embeddinggemma:300m` — gäbe sonst eine Datei, die ein Windows-Klon
	// nicht anlegen kann.
	ziel := filepath.Join("..", "..", "testdaten", "retrieval",
		"messung-"+strings.ReplaceAll(bericht.Modell.Modell, ":", "-")+".json")

	if fehler := os.WriteFile(ziel, append(roh, '\n'), 0o600); fehler != nil {
		t.Fatalf("%s schreiben: %v", ziel, fehler)
	}

	t.Logf("Maschine %s", bericht.Maschine)
	t.Logf("Modell %s (%s, %s), %d Dokumente, Bau %v, Vektoren %d Bytes, "+
		"%d Messpunkte je Art",
		bericht.Modell.Modell, bericht.Modell.Laufzeit,
		bericht.Modell.Modellrevision[:12], bericht.Dokumente,
		bericht.Bauzeit.Round(time.Millisecond), bericht.Indexbytes,
		bericht.Messpunkte)

	for _, wert := range bericht.Werte {
		t.Logf("%-12s %-14s %-6s n=%d  R@10=%.3f  MRR=%.3f  nDCG=%.3f",
			wert.Art, wert.Klasse, wert.Richtung, wert.Fragen,
			wert.Recall, wert.MRR, wert.NDCG)
	}

	for _, art := range dieArten {
		t.Logf("%-12s p50=%v p95=%v", art,
			bericht.Dauer[art].P50.Round(time.Microsecond),
			bericht.Dauer[art].P95.Round(time.Microsecond))
	}

	t.Logf("Bericht: %s", ziel)
}

// dieMaschine beschreibt den Rechner — für die Protokollpflicht des Issues.
func dieMaschine() string {
	return fmt.Sprintf("%s/%s, %d Kerne", runtime.GOOS, runtime.GOARCH,
		runtime.NumCPU())
}
