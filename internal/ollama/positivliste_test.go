package ollama_test

// **Die Positivliste ist geschlossen** — die Abnahme aus A7
// (#1446):
// „Kein impliziter Modellpull, kein Cloudfallback".
//
// # Zwei Läufe, und sie messen Verschiedenes
//
//   - Der **Spion** sagt, welche Wege die heutigen Rufe wirklich gehen.
//   - Der **Leser** sagt, dass es im Quelltext keine anderen gibt.
//
// **Einer allein reichte nicht** (§10, „die Reichweite einer Messung"): der
// Spion sähe ein `/api/pull` nicht, das nur unter einer Bedingung gerufen
// wird, und der Leser sähe nicht, ob ein Weg der Liste überhaupt begangen
// wird.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
)

// **Jeder Ruf geht einen Weg der Liste.**
func TestJederRufGehtEinenWegDerListe(t *testing.T) {
	t.Parallel()

	spion := &ollamaspion{antwort: gegenstelle(t)}

	server := httptest.NewServer(spion)
	t.Cleanup(server.Close)

	klient := klientFuer(server)

	// **Jeder öffentliche Ruf, den es gibt** — mit gültigen Argumenten, damit
	// er die Anfrage wirklich stellt.
	if _, fehler := klient.Fassung(t.Context()); fehler != nil {
		t.Fatalf("Fassung: %v", fehler)
	}

	if _, _, fehler := klient.Modelle(t.Context()); fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if _, fehler := klient.Rede(t.Context(), ollama.Frage{
		Alias: "gemma4:8b", Auftrag: "x",
	}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	if _, fehler := klient.Einbetter(domain.Vektorraum{Modell: "einbetter:300m", Vertrag: derProbevertrag()}).
		Einbetten(t.Context(), []string{"Blitzschlag"}); fehler != nil {
		t.Fatalf("Einbetten: %v", fehler)
	}

	gesehen := spion.wege()

	erwartet := ollama.Wege()
	slices.Sort(erwartet)

	// **Gleichheit und nicht Teilmenge**: „nichts Fremdes" wäre auch dann
	// grün, wenn gar nichts angekommen wäre.
	if !slices.Equal(gesehen, erwartet) {
		t.Errorf("gegangen wurden %v, erwartet %v", gesehen, erwartet)
	}
}

// **Und im Quelltext steht kein fünfter Weg.**
//
// # Was dieser Lauf hält
//
// **Die Zusage aus dem Paketkopf**: es gibt kein `/api/pull`, kein
// `/api/create`, kein `/api/delete`. Ohne ihn wäre sie ein Satz in einem
// Kommentar — und ein Kopf, der mehr verspricht als sein Rumpf hält, ist
// schlimmer als gar keiner (§10).
func TestImQuelltextStehtKeinFuenfterWeg(t *testing.T) {
	t.Parallel()

	gefunden := wegeImQuelltext(t)

	erwartet := ollama.Wege()
	slices.Sort(erwartet)

	if !slices.Equal(gefunden, erwartet) {
		t.Errorf("im Quelltext stehen die Wege %v, die Liste führt %v",
			gefunden, erwartet)
	}
}

// wegeImQuelltext sammelt jede Zeichenkette, die wie ein Weg aussieht.
func wegeImQuelltext(t *testing.T) []string {
	t.Helper()

	dateien, fehler := filepath.Glob("*.go")
	if fehler != nil {
		t.Fatalf("die Dateien finden: %v", fehler)
	}

	// **Die Gegenprobe zum Leser selbst**: fände er keine Datei, wäre er
	// still grün und hielte gar nichts.
	if len(dateien) < 2 {
		t.Fatalf("%d Quelldateien gefunden — der Leser liest nichts", len(dateien))
	}

	satz := token.NewFileSet()

	var gefunden []string

	for _, datei := range dateien {
		if strings.HasSuffix(datei, "_test.go") {
			continue
		}

		baum, lesefehler := parser.ParseFile(satz, datei, nil, 0)
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", datei, lesefehler)
		}

		gefunden = append(gefunden, wegeAus(baum)...)
	}

	slices.Sort(gefunden)

	return slices.Compact(gefunden)
}

// wegeAus zieht die Wege aus einem Dateibaum.
func wegeAus(baum *ast.File) []string {
	var gefunden []string

	ast.Inspect(baum, func(knoten ast.Node) bool {
		wert, istText := knoten.(*ast.BasicLit)
		if !istText || wert.Kind != token.STRING {
			return true
		}

		text, fehler := strconv.Unquote(wert.Value)
		if fehler != nil || !strings.HasPrefix(text, "/api/") {
			return true
		}

		gefunden = append(gefunden, text)

		return true
	})

	return gefunden
}

// ollamaspion schreibt jeden Weg auf und reicht die Antwort durch.
type ollamaspion struct {
	antwort  *httptest.Server
	schloss  sync.Mutex
	gesehene []string
}

func (o *ollamaspion) ServeHTTP(schreiben http.ResponseWriter, gestellt *http.Request) {
	o.schloss.Lock()

	if !slices.Contains(o.gesehene, gestellt.URL.Path) {
		o.gesehene = append(o.gesehene, gestellt.URL.Path)
	}

	o.schloss.Unlock()

	o.antwort.Config.Handler.ServeHTTP(schreiben, gestellt)
}

// wege gibt die gesehenen, sortiert.
func (o *ollamaspion) wege() []string {
	o.schloss.Lock()
	defer o.schloss.Unlock()

	gesehen := slices.Clone(o.gesehene)
	slices.Sort(gesehen)

	return gesehen
}
