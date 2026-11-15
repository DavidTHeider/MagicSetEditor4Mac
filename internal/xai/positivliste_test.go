package xai_test

// **Die Positivliste ist geschlossen** — Werkstück A10 von Phase 5.3
// (#1449).
//
// # Warum sie hier besonders eng ist
//
// **Weil xAI mehr anbietet, als dieses Programm will.** Es gibt
// `/v1/responses` (das Antworten **30 Tage** beim Anbieter hält),
// `/v1/image-generation-models`, eine Stapel-API und einen Weg für entfernte
// Werkzeugserver. Das Issue schließt den letzten ausdrücklich aus — „keine
// automatische Kopplung an Remote-MCP/Tools des Providers" —, die Bilder
// stehen in A12 und die Stapel in A13.
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
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/xai"
)

// **Die Positivliste nennt zwei Wege.**
func TestDiePositivlisteNenntZweiWege(t *testing.T) {
	t.Parallel()

	wege := xai.Wege()
	if len(wege) != 2 {
		t.Fatalf("die Positivliste führt %v", wege)
	}

	for _, weg := range wege {
		if !strings.HasPrefix(weg, "/v1/") {
			t.Errorf("%q sieht nicht wie ein Weg dieser Schnittstelle aus", weg)
		}
	}

	// **Und keiner davon legt beim Anbieter etwas ab.**
	if slices.Contains(wege, "/v1/responses") {
		t.Error("/v1/responses hält Antworten 30 Tage — siehe 0101")
	}
}

// **Jeder Ruf geht einen Weg der Liste.**
func TestJederRufGehtEinenWegDerListe(t *testing.T) {
	t.Parallel()

	spion := &wegespion{}

	server := httptest.NewServer(spion)
	t.Cleanup(server.Close)

	klient := xai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		xai.Zugang{Schluessel: "geheim"})

	if _, _, fehler := klient.Modelle(t.Context()); fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if _, fehler := klient.Rede(t.Context(), xai.Frage{
		Alias: "m", Auftrag: "x",
	}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	erlaubt := xai.Wege()

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

	if strings.HasSuffix(gestellt.URL.Path, "language-models") {
		_, _ = schreiben.Write([]byte(`{"models":[]}`)) //nolint:errcheck // die Gegenstelle

		return
	}

	_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
		`{"model":"m","choices":[{"message":{"content":"x"},` +
			`"finish_reason":"stop"}]}`))
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

	erwartet := xai.Wege()
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
