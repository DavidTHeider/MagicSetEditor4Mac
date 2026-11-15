// **Wie viele Werkzeuge der gebaute MCP-Server wirklich anmeldet** — Befund 6
// des Zweitblicks über die Phasen 4.0 bis 4.9.
//
// # Warum hier und nicht in `internal/werkzeuge`
//
// Dort zählte ein Wächter die MCP-Werkzeuge, indem er in den Go-Quellen nach
// Zeichenketten der Form `"lesen.…"` suchte. Das ist **keine Messung der
// Fläche**, sondern eine des Quelltexts: eine passende Zeichenkette in einer
// Erklärung oder einem Kommentar erfüllt sie, ohne dass ein Werkzeug über
// `tools/list` erreichbar wäre.
//
// # Und warum nicht in `internal/mcp`
//
// **Weil dort die Nähte fehlen.** Gemessen am 2026-09-03:
// `mcp.Werkzeuge(mcp.Naehte{})` gibt **80** Werkzeuge, die Vorrichtung der
// Paketprüfungen **88** — ganze Gruppen hängen an einer Naht und entfallen
// ohne sie. Ein Zähler dort zählte verlässlich die falsche Zahl.
//
// **Hier läuft das gebaute Programm**, mit den Nähten, die es im Betrieb hat.
// Das ist die einzige Zahl, die etwas über die Fläche sagt.
package flaechen_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// angemeldeteWerkzeuge fragt den gebauten Server über `tools/list`.
func angemeldeteWerkzeuge(t *testing.T, umgebung []string) []string {
	t.Helper()

	const anmeldung = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":` +
		`{"protocolVersion":"2025-06-18","capabilities":{},` +
		`"clientInfo":{"name":"zaehler","version":"1"}}}` + "\n"

	eingabe := anmeldung + `{"jsonrpc":"2.0","id":2,"method":"tools/list"}` + "\n"

	// #nosec G204 -- gebautes Programm aus TestMain
	lauf := exec.CommandContext(t.Context(), gebaut.mcp)
	lauf.Env = umgebung
	lauf.Stdin = strings.NewReader(eingabe)

	var aus bytes.Buffer
	lauf.Stdout = &aus

	if err := lauf.Run(); err != nil {
		t.Fatalf("mse-mcp tools/list: %v", err)
	}

	var namen []string

	for _, zeile := range strings.Split(aus.String(), "\n") {
		if !strings.Contains(zeile, `"tools"`) {
			continue
		}

		var antwort struct {
			Ergebnis struct {
				Werkzeuge []struct {
					Name string `json:"name"`
				} `json:"tools"`
			} `json:"result"`
		}

		if err := json.Unmarshal([]byte(zeile), &antwort); err != nil {
			t.Fatalf("die Werkzeugliste lesen: %v\n%s", err, zeile)
		}

		for _, werkzeug := range antwort.Ergebnis.Werkzeuge {
			namen = append(namen, werkzeug.Name)
		}
	}

	if len(namen) == 0 {
		t.Fatalf("keine Werkzeuge in der Antwort:\n%s", aus.String())
	}

	return namen
}

// **Die Zahl in `FAEHIGKEITEN.md` ist die Zahl, die der Server anmeldet.**
func TestDieZahlDerWerkzeugeStimmtMitDerAngemeldeten(t *testing.T) {
	namen := angemeldeteWerkzeuge(t, eigenerBestand(t))

	pfad := filepath.Join("..", "..", "docs", "FAEHIGKEITEN.md")

	inhalt, err := os.ReadFile(pfad) // #nosec G304 -- fester Pfad im Repositorium
	if err != nil {
		t.Fatalf("%s nicht lesbar: %v", pfad, err)
	}

	muster := regexp.MustCompile(`(\d+)\s*\n?MCP-Werkzeuge`)

	treffer := muster.FindSubmatch(inhalt)
	if treffer == nil {
		t.Fatal("FAEHIGKEITEN.md nennt keine Zahl von MCP-Werkzeugen")
	}

	genannt, err := strconv.Atoi(string(treffer[1]))
	if err != nil {
		t.Fatalf("die genannte Zahl ist keine: %v", err)
	}

	if genannt != len(namen) {
		t.Errorf("FAEHIGKEITEN.md nennt %d MCP-Werkzeuge — der gebaute Server "+
			"meldet %d an. Gezählt wird über `tools/list`, nicht über "+
			"Zeichenketten im Quelltext", genannt, len(namen))
	}
}

// **Und jeder angemeldete Name steht im Katalog.**
//
// Die andere Richtung: ein Werkzeug, das der Server herausgibt und das in
// `FAEHIGKEITEN.md` fehlt, wäre eine Fläche ohne Zeile — und niemand hätte je
// entschieden, dass es sie gibt.
func TestJedesAngemeldeteWerkzeugStehtImKatalog(t *testing.T) {
	pfad := filepath.Join("..", "..", "docs", "FAEHIGKEITEN.md")

	inhalt, err := os.ReadFile(pfad) // #nosec G304 -- fester Pfad im Repositorium
	if err != nil {
		t.Fatalf("%s nicht lesbar: %v", pfad, err)
	}

	katalog := string(inhalt)

	for _, name := range angemeldeteWerkzeuge(t, eigenerBestand(t)) {
		if !strings.Contains(katalog, "`"+name+"`") {
			t.Errorf("das Werkzeug %q ist angemeldet und steht in keiner "+
				"Zeile von FAEHIGKEITEN.md", name)
		}
	}
}
