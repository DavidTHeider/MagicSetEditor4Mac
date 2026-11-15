package openai_test

// **Die Positivliste ist geschlossen** — Werkstück A9 von Phase 5.3
// (#1448).
//
// # Warum sie hier besonders eng ist
//
// **Weil die Responses-API mehr kann, als dieses Programm will.** Sie legt
// Antworten beim Anbieter ab (`/v1/responses/{id}`), erzeugt Bilder, nimmt
// Dateien entgegen und spricht mit entfernten Werkzeugservern. Das Issue
// schließt das letzte ausdrücklich aus — „keine automatische Kopplung an
// Remote-MCP/Tools des Providers" —, und die Bilder stehen in A12.
//
// # Zwei Läufe, und sie messen Verschiedenes
//
//   - Der **Spion** sagt, welche Wege die heutigen Rufe wirklich gehen.
//   - Der **Leser** sagt, dass es im Quelltext keine anderen gibt.
//
// **Einer allein reichte nicht** (§10, „die Reichweite einer Messung").

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

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/openai"
)

// **Jeder Ruf geht einen Weg der Liste.**
func TestJederRufGehtEinenWegDerListe(t *testing.T) {
	t.Parallel()

	spion := &wegespion{}

	server := httptest.NewServer(spion)
	t.Cleanup(server.Close)

	klient := openai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		openai.Zugang{Schluessel: "geheim"})

	if _, _, fehler := klient.Modelle(t.Context()); fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if _, fehler := klient.Rede(t.Context(), openai.Frage{
		Alias: "m", Auftrag: "x",
	}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	erlaubt := openai.Wege()

	gesehen := spion.wege()
	if len(gesehen) == 0 {
		t.Fatal("der Spion hat nichts gesehen — dieser Lauf prüft nichts")
	}

	for _, weg := range gesehen {
		if !slices.Contains(erlaubt, weg) {
			t.Errorf("gerufen wurde %q — die Liste führt %v", weg, erlaubt)
		}
	}
}

// wegespion hält jeden gerufenen Weg fest.
type wegespion struct {
	schloss   sync.Mutex
	gerufenen []string
}

func (s *wegespion) ServeHTTP(schreiben http.ResponseWriter, gestellt *http.Request) {
	s.schloss.Lock()
	s.gerufenen = append(s.gerufenen, gestellt.URL.Path)
	s.schloss.Unlock()

	if strings.HasSuffix(gestellt.URL.Path, "/models") {
		_, _ = schreiben.Write([]byte(`{"data":[]}`)) //nolint:errcheck // die Gegenstelle

		return
	}

	_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
		`{"status":"completed","model":"m","output":[]}`))
}

func (s *wegespion) wege() []string {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	return slices.Clone(s.gerufenen)
}

// **Im Quelltext steht kein dritter Weg.**
func TestImQuelltextStehtKeinDritterWeg(t *testing.T) {
	t.Parallel()

	gefunden := wegeImQuelltext(t)

	erwartet := openai.Wege()
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
		if fehler != nil || !strings.HasPrefix(text, "/v1/") {
			return true
		}

		gefunden = append(gefunden, text)

		return true
	})

	return gefunden
}
