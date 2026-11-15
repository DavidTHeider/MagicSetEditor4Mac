// **Was am Sitz steht, erreicht auch den Tisch** — PLAY-03
// (#1711).
//
// # Der Anlass
//
// `domain.Sitz` führt zwölf Felder. `domain.Sichtsitz` kopierte acht davon,
// und `app.SitzDTO` dieselben acht. **Fünf blieben in der Domäne liegen** —
// gemessen am 2026-09-21: `Strahlung` (CR 728.1), `Commander` (CR 903.3),
// `Commandersteuer` (CR 903.8), `Commanderschaden` (CR 903.10a) und
// `Leergezogen` (CR 704.5b).
//
// Alle fünf sind gerechnet, gespeichert und geprüft — und **keines** war je
// im Fenster zu sehen. Der Entwurf verlangt zwei davon wörtlich (Artboard
// `TischCommander`: „Steuer 2", „ab 21 ist die Partie für dich vorbei").
//
// Das ist dieselbe Gestalt wie
// #1756:
// eine Größe, die niemand verbraucht. Nur fällt sie hier nicht auf, weil
// jede Schicht für sich vollständig aussieht — die Lücke liegt **zwischen**
// ihnen (§10, Dimension „Nähte").
//
// # Was er prüft
//
// Jedes exportierte Feld von [domain.Sitz] steht in [domain.Sichtsitz], und
// jedes Feld von dort in `app.SitzDTO` — **oder** es steht in
// [sitzausnahmen] mit einem Grund.
//
// # Was er nicht prüft
//
// **Dass das Feld irgendwo gezeigt wird.** Er sieht, dass es durch die
// Nähte kommt, nicht dass eine Fläche es liest — das bliebe dieselbe halbe
// Zusage eine Ebene höher. Dafür steht der Lauf über `sitzkopf` daneben.
//
// **Und er prüft keine Bedeutung.** Ob `Commanderschaden` im DTO wirklich
// den Schaden je Commander trägt, sagt kein Namensvergleich.
package werkzeuge_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"
)

// sitzausnahmen nennt, was **nicht** eins zu eins durchgereicht wird — und
// warum.
var sitzausnahmen = map[string]string{
	// Das Deck selbst ist das Geheimnis des Gegenübers; die Sicht gibt
	// seinen Abdruck und seine Zahl.
	"Deck": "erscheint als Deckabdruck und Karten — eine Deckliste ist geheim",
}

func TestJedesSitzfeldErreichtDieSicht(t *testing.T) {
	sitz := felderVon(t, filepath.Join("..", "..", "internal", "domain", "sitz.go"), "Sitz")
	sicht := felderVon(t, filepath.Join("..", "..", "internal", "domain", "partiesicht.go"), "Sichtsitz")

	fehlen(t, sitz, sicht, "domain.Sichtsitz")
}

func TestJedesSichtfeldErreichtDasDTO(t *testing.T) {
	sicht := felderVon(t, filepath.Join("..", "..", "internal", "domain", "partiesicht.go"), "Sichtsitz")
	dto := felderVon(t, filepath.Join("..", "..", "internal", "app", "partie.go"), "SitzDTO")

	fehlen(t, sicht, dto, "app.SitzDTO")
}

// fehlen meldet jedes Feld, das drüben nicht ankommt.
func fehlen(t *testing.T, hier, dort []string, wohin string) {
	t.Helper()

	angekommen := make(map[string]bool, len(dort))
	for _, name := range dort {
		angekommen[name] = true
	}

	for _, name := range hier {
		if angekommen[name] {
			continue
		}

		if grund, erlaubt := sitzausnahmen[name]; erlaubt {
			t.Logf("%s steht nicht in %s: %s", name, wohin, grund)

			continue
		}

		t.Errorf("%s erreicht %s nicht — entweder durchreichen oder mit "+
			"Grund in sitzausnahmen eintragen", name, wohin)
	}
}

// felderVon gibt die exportierten Feldnamen einer Struktur.
func felderVon(t *testing.T, pfad, typ string) []string {
	t.Helper()

	baum, fehler := parser.ParseFile(token.NewFileSet(), pfad, nil,
		parser.SkipObjectResolution)
	if fehler != nil {
		t.Fatalf("%s zerteilen: %v", pfad, fehler)
	}

	var namen []string

	ast.Inspect(baum, func(knoten ast.Node) bool {
		erklaerung, istTyp := knoten.(*ast.TypeSpec)
		if !istTyp || erklaerung.Name.Name != typ {
			return true
		}

		verbund, istVerbund := erklaerung.Type.(*ast.StructType)
		if !istVerbund {
			return true
		}

		for _, feld := range verbund.Fields.List {
			for _, name := range feld.Names {
				if name.IsExported() {
					namen = append(namen, name.Name)
				}
			}
		}

		return false
	})

	if len(namen) == 0 {
		t.Fatalf("in %s fand sich kein Feld von %s — der Leser greift daneben",
			pfad, typ)
	}

	return namen
}
