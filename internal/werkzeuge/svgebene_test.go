// Der Wächter über den Plan im SVG-Schreiber.
//
// # Wofür
//
// Der Zweitblick über Phase P fand neun Eigenschaften an `render.Ebene`, die
// der Rasterer anwendet und `internal/svg` fallen ließ — Deckung, Winkel,
// Maske, Verlauf, Mischart, Passung, Ecke, Rand und den Bildbaum
// (#852).
// Keine davon war benannt; sie waren einfach nicht da.
//
// **Dieser Wächter macht die nächste rot.** Wer `render.Ebene` erweitert und
// den SVG-Schreiber nicht anfasst, bekommt einen roten Lauf statt einer Karte,
// der etwas fehlt.
//
// # Was er nicht prüft
//
// **Ob die Eigenschaft *richtig* geschrieben wird.** Er sieht einen
// Feldnamen im Quelltext, nicht seine Wirkung — dafür stehen die Tests in
// `internal/svg`. Ein `_ = ebene.Verlauf` genügte ihm; das ist der Preis
// jedes Wächters, der über Namen geht, und er gehört ausgesprochen.
//
// **Ob die Liste der bewusst Ausgelassenen stimmt.** Sie steht unten und ist
// eine Behauptung dieses Tests; was sie kostet, sagt der Kommentar daneben.
package werkzeuge

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// ohneEntsprechungImSVG sind die Felder von `render.Ebene`, die der
// SVG-Schreiber nicht liest — **jedes mit seinem Grund**.
var ohneEntsprechungImSVG = map[string]string{
	// **Mischart wird gemeldet, nicht angewandt.** CSS `mix-blend-mode` kennt
	// `add`, `subtract`, `stamp`, `negation`, `reflect`, `shadow` und `glow`
	// nicht, und `shadow` allein steht 667 der 698 Aufrufe im vollen Pack.
	// Fünf der zwanzig wären übernehmbar; sie zu übernehmen und die übrigen
	// nicht ergäbe eine Ausgabe, die mal stimmt und mal nicht.
	//
	// **Der Name steht trotzdem im Quelltext** — im Befundzweig. Deshalb
	// steht er hier nicht; die Liste bliebe sonst eine, die niemand liest.
}

// TestJedesFeldEinerEbeneStehtImSVGSchreiber liest beide Seiten.
func TestJedesFeldEinerEbeneStehtImSVGSchreiber(t *testing.T) {
	felder := feldnamenVon(t, "../render/plan.go", "Ebene")
	if len(felder) < 10 {
		t.Fatalf("%d Felder an render.Ebene gefunden — der Leser greift daneben", len(felder))
	}

	// **Ohne Kommentare** (#1346):
	// `internal/svg/svg.go` nennt `Mischart` in einem Kommentar. Wer die
	// produktive Behandlung entfernte, hatte trotzdem einen grünen Lauf — der
	// Nachlauf über Phase 4.4 hat es nachgestellt.
	quelltext := GoVerzeichnisOhneKommentare(t, "../svg")

	var fehlend []string

	for _, feld := range felder {
		if _, benannt := ohneEntsprechungImSVG[feld]; benannt {
			continue
		}

		if !strings.Contains(quelltext, "."+feld) {
			fehlend = append(fehlend, feld)
		}
	}

	if len(fehlend) > 0 {
		t.Errorf("internal/svg liest diese Felder von render.Ebene nicht: %s\n"+
			"Entweder schreiben — oder in ohneEntsprechungImSVG mit Grund eintragen.",
			strings.Join(fehlend, ", "))
	}
}

// TestDieAusgelasseneListeIstNichtVeraltet ist die Gegenprobe.
//
// Ein Eintrag für ein Feld, das es nicht mehr gibt, sieht wie eine
// Entscheidung aus und ist keine.
func TestDieAusgelasseneListeIstNichtVeraltet(t *testing.T) {
	felder := feldnamenVon(t, "../render/plan.go", "Ebene")

	for feld := range ohneEntsprechungImSVG {
		if !enthaelt(felder, feld) {
			t.Errorf("ohneEntsprechungImSVG führt %q — render.Ebene hat es nicht mehr", feld)
		}
	}
}

// TestBeideSeitenNehmenDieselbeWinkeltoleranz hält eine Zahl an zwei Stellen.
//
// `svg.gerade` rechnet `render.geraderWinkel` nach, weil die dort
// unexportiert ist und hier nur ein Befund davon abhängt. **Was auseinander
// laufen kann, ist die Toleranz** — und dann gälte ein Winkel im Bild als
// gerade und im SVG als schräg.
func TestBeideSeitenNehmenDieselbeWinkeltoleranz(t *testing.T) {
	const toleranz = "0.001"

	for _, pfad := range []string{"../render/drehung.go", "../svg/winkel.go"} {
		roh, err := os.ReadFile(pfad) //nolint:gosec // ein fester Pfad im Repositorium
		if err != nil {
			t.Fatalf("%s lesen: %v", pfad, err)
		}

		if !strings.Contains(string(roh), toleranz) {
			t.Errorf("%s nennt die Toleranz %s nicht", pfad, toleranz)
		}
	}
}

// feldnamenVon gibt die Feldnamen eines Verbunds aus einer Datei.
func feldnamenVon(t *testing.T, pfad, verbund string) []string {
	t.Helper()

	baum, err := parser.ParseFile(token.NewFileSet(), pfad, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("%s lesen: %v", pfad, err)
	}

	var namen []string

	ast.Inspect(baum, func(knoten ast.Node) bool {
		art, ist := knoten.(*ast.TypeSpec)
		if !ist || art.Name.Name != verbund {
			return true
		}

		verbundtyp, ist := art.Type.(*ast.StructType)
		if !ist {
			return false
		}

		for _, feld := range verbundtyp.Fields.List {
			for _, name := range feld.Names {
				namen = append(namen, name.Name)
			}
		}

		return false
	})

	return namen
}

// enthaelt sagt, ob ein Name in der Liste steht.
func enthaelt(namen []string, gesucht string) bool {
	for _, name := range namen {
		if name == gesucht {
			return true
		}
	}

	return false
}
