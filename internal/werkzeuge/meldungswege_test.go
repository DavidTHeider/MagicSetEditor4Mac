// Der Wächter über das, was in einer Meldung steht.
//
// # Die Zusage
//
// **Kein Satz, der die Oberfläche erreicht, trägt einen Pfad aus dem Inneren,
// einen fremden Go-Fehler oder einen SQLite-Text** (AGENTS.md §8, Werkstück 1
// von Phase 4.13,
// #1352).
//
// # Warum am `fmt.Errorf` und nicht an der Oberfläche
//
// **Weil die Oberfläche vier Stück ist.** Das Fenster übersetzt über den
// Schlüssel und sieht den Go-Satz nie; Kommandozeile, MCP und REST haben
// keinen Katalog und zeigen ihn immer. Eine Prüfung an der letzten Stelle
// prüfte nur eine von vieren — und die drei ohne Katalog sind gerade die, die
// es betrifft.
//
// **Geprüft wird deshalb dort, wo der Satz entsteht**: an jedem `fmt.Errorf`,
// das eines der neun Wahrzeichen umschließt. Was eines umschließt, kommt durch
// [app.Uebersetze] als `Meldung` heraus.
//
// # Die zwei Gestalten
//
//  1. **Ein zweites `%w`** — der Text eines fremden Fehlers wandert in den
//     Satz. `fmt.Errorf("%w: %w", domain.ErrNotFound, sqlitefehler)` ist die
//     schärfste Form: der rohe Treibertext, unverändert.
//  2. **Ein voller Pfad** — ein Bezeichner wie `pfad`, `ort` oder `ziel`,
//     nicht durch `filepath.Base` gekürzt.
//
// **Der Ausweg ist [domain.MitGrund]**: das Wahrzeichen und der Satz für
// Menschen bleiben, die Ursache geht in die Kette und von dort ins Protokoll.
//
// # Was er nicht prüft
//
// **Ob ein Pfad einer ist, den jemand selbst gewählt hat.** „Die Bilddatei
// /Users/wer/bild.png ließ sich nicht öffnen" nennt die Datei, die jemand
// eben ausgesucht hat — das ist Hilfe und kein Leck. Eine Maschine
// unterscheidet das nicht; die Ausnahmen unten tun es, jede mit ihrem Grund.
//
// **Und ob der Satz gut ist.** Er sieht, dass nichts Fremdes darin steht,
// nicht dass darin steht, was zu tun ist (§10).
package werkzeuge

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// wahrzeichen sind die Fehlerarten, deren Satz die Oberfläche erreicht.
var wahrzeichen = map[string]bool{
	"ErrValidation": true, "ErrNotFound": true, "ErrConflict": true,
	"ErrBelegt": true, "ErrNurLesend": true, "ErrErlaubnis": true,
	"ErrSchluesselFehlt": true, "ErrKeinSetMitStil": true, "ErrSicherung": true,
}

// pfadnamen sind Bezeichner, hinter denen ein Dateipfad steht.
var pfadnamen = map[string]bool{
	"pfad": true, "ort": true, "ziel": true, "adresse": true,
	"datei": true, "verzeichnis": true, "quelle": true, "bestandspfad": true,
}

// meldungsausnahme ist eine Stelle, an der ein fremder Text im Satz stehen
// darf — **mit ihrem Grund**.
//
// **Erkannt an Datei und einem Stück des Satzes, nicht an der Zeilennummer.**
// Der erste Anlauf nahm die Zeile, und schon der nächste eingefügte Kommentar
// verschob sie: der Wächter wurde rot, ohne dass sich etwas an der Sache
// geändert hätte. Ein Schlüssel, der bei jeder Umformatierung bricht, wird
// nachgezogen statt gelesen — und dann steht dort irgendwann eine Ausnahme,
// die niemand mehr geprüft hat.
type meldungsausnahme struct {
	datei    string
	satzteil string
	grund    string
}

// erlaubt sucht die Ausnahme zu einem Befund.
func erlaubt(befund meldungsbefund) (string, bool) {
	for _, ausnahme := range meldungsausnahmen {
		if strings.HasPrefix(befund.stelle, ausnahme.datei+":") &&
			strings.Contains(befund.ruf, ausnahme.satzteil) {
			return ausnahme.grund, true
		}
	}

	return "", false
}

func TestKeineMeldungTraegtEinenFremdenText(t *testing.T) {
	t.Parallel()

	befunde := meldungsbefunde(t, "../")

	for _, befund := range befunde {
		if grund, egal := erlaubt(befund); egal {
			t.Logf("%s: erlaubt — %s", befund.stelle, grund)

			continue
		}

		t.Errorf("%s trägt %s in einer Meldung, die die Oberfläche erreicht:\n\t%s\n"+
			"\tDer Ausweg ist domain.MitGrund — der Satz für Menschen bleibt, "+
			"die Ursache geht ins Protokoll (§8)",
			befund.stelle, befund.art, befund.ruf)
	}
}

// meldungsbefund ist eine Fundstelle.
type meldungsbefund struct {
	stelle string
	art    string
	ruf    string
}

// meldungsbefunde sammelt jeden Ruf, der einen fremden Text in den Satz setzt.
func meldungsbefunde(t *testing.T, wurzel string) []meldungsbefund {
	t.Helper()

	var befunde []meldungsbefund

	menge := token.NewFileSet()

	if fehler := filepath.Walk(wurzel, func(pfad string, angabe os.FileInfo, fehler error) error {
		if fehler != nil {
			return fehler
		}

		if angabe.IsDir() || !strings.HasSuffix(pfad, ".go") ||
			strings.HasSuffix(pfad, "_test.go") {
			return nil
		}

		baum, lesefehler := parser.ParseFile(menge, pfad, nil, parser.SkipObjectResolution)
		if lesefehler != nil {
			return lesefehler
		}

		befunde = append(befunde, befundeInDatei(menge, baum, pfad)...)

		return nil
	}); fehler != nil {
		t.Fatalf("durchgehen: %v", fehler)
	}

	return befunde
}

// befundeInDatei geht einen Syntaxbaum durch.
func befundeInDatei(menge *token.FileSet, baum *ast.File, pfad string) []meldungsbefund {
	var befunde []meldungsbefund

	ast.Inspect(baum, func(knoten ast.Node) bool {
		ruf, ist := knoten.(*ast.CallExpr)
		if !ist || !istErrorf(ruf) || !umschliesstWahrzeichen(ruf) {
			return true
		}

		stelle := kurzeStelle(pfad) + ":" + strconv.Itoa(menge.Position(ruf.Pos()).Line)
		text := formatstring(ruf)

		if strings.Count(text, "%w") > 1 {
			befunde = append(befunde, meldungsbefund{
				stelle: stelle, art: "einen fremden Go-Fehler (ein zweites %w)",
				ruf: text,
			})
		}

		if name := pfadargument(ruf); name != "" {
			befunde = append(befunde, meldungsbefund{
				stelle: stelle, art: "einen vollen Pfad (" + name + ")", ruf: text,
			})
		}

		return true
	})

	return befunde
}

// istErrorf sagt, ob der Ruf `fmt.Errorf` ist.
func istErrorf(ruf *ast.CallExpr) bool {
	wahl, ist := ruf.Fun.(*ast.SelectorExpr)
	if !ist || wahl.Sel.Name != "Errorf" {
		return false
	}

	paket, istName := wahl.X.(*ast.Ident)

	return istName && paket.Name == "fmt"
}

// umschliesstWahrzeichen sagt, ob eines der neun Wahrzeichen unter den
// Argumenten steht.
func umschliesstWahrzeichen(ruf *ast.CallExpr) bool {
	for _, argument := range ruf.Args {
		switch benannt := argument.(type) {
		case *ast.SelectorExpr:
			if wahrzeichen[benannt.Sel.Name] {
				return true
			}

		case *ast.Ident:
			if wahrzeichen[benannt.Name] {
				return true
			}
		}
	}

	return false
}

// formatstring setzt die Teile des ersten Arguments zusammen.
func formatstring(ruf *ast.CallExpr) string {
	if len(ruf.Args) == 0 {
		return ""
	}

	var gesammelt strings.Builder

	ast.Inspect(ruf.Args[0], func(knoten ast.Node) bool {
		wert, ist := knoten.(*ast.BasicLit)
		if !ist || wert.Kind != token.STRING {
			return true
		}

		text, fehler := strconv.Unquote(wert.Value)
		if fehler == nil {
			gesammelt.WriteString(text)
		}

		return true
	})

	return gesammelt.String()
}

// pfadargument gibt den Namen des ersten Arguments, hinter dem ein Pfad steht
// — oder nichts.
//
// **`filepath.Base(pfad)` zählt nicht**: es ist die Abhilfe, nicht der Mangel.
func pfadargument(ruf *ast.CallExpr) string {
	for _, argument := range ruf.Args[1:] {
		name, ist := argument.(*ast.Ident)
		if ist && pfadnamen[strings.ToLower(name.Name)] {
			return name.Name
		}
	}

	return ""
}

// kurzeStelle gibt `<paket>/<datei>.go`.
func kurzeStelle(pfad string) string {
	sauber := filepath.ToSlash(filepath.Clean(pfad))
	teile := strings.Split(sauber, "/")

	if len(teile) < 2 {
		return sauber
	}

	return strings.Join(teile[len(teile)-2:], "/")
}
