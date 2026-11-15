package werkzeuge_test

// **Das Regelregister deckt sich mit dem Quelltext** — PLAY-01.AK-02
// (#1709):
// „Vorhandenes Datenfeld, implementierte Funktion, produktive Verdrahtung und
// durchgesetzte Regel werden getrennt geprüft."
//
// # Vier Fragen, einzeln
//
//  1. **Gibt es die Funktion?** Stelle und Aufrufer stehen als Funktion oder
//     Methode in `internal/domain`.
//  2. **Steht die Regel dort?** Die **Stelle** nennt die Nummer — an sich
//     selbst, in ihrem Kommentar oder Rumpf.
//  3. **Ist sie produktiv verdrahtet?** Der Aufrufer erreicht die Stelle, und
//     `internal/service` oder `internal/app` erreichen den Aufrufer. Ein
//     Datenfeld, das niemand liest, und eine Funktion, die niemand ruft,
//     fallen hier auf — genau das war Befund 3 in `docs/REGELABDECKUNG.md`
//     (`Startleben`).
//  4. **Gibt es die Läufe?** Positiv und Negativ stehen als `func Test…(`.
//
// **Hier stand eine Suchtiefe von drei**, und sie hat beim ersten Lauf die
// falsche Funktion durchgelassen: für den Kampfschaden stand `Kaempfen` im
// Register — die **Schlüsselhandlung** „kämpfen", nicht der Schaden des
// Kampfschrittes. Seit Stelle und Aufrufer getrennt stehen, gibt es keine
// Tiefe mehr, die jemand wählen muss.
//
// # Was er nicht hält
//
// **Ob die Läufe die Regel prüfen, und ob die Beschreibung trifft.** Er sieht,
// dass es sie gibt. **Und der Aufrufgraph geht nach Namen**, nicht nach
// Typen: zwei Methoden gleichen Namens gelten als dieselbe. Das macht die
// Erreichbarkeit großzügig — die Regelnummer prüft er dagegen genau, an der
// einen Stelle.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// funktion ist eine Funktion der Domäne mit dem, was sie ruft.
type funktion struct {
	text  string
	rufe  map[string]bool
	namen string
}

// **Jeder Eintrag des Registers hat Funktion, Verdrahtung, Regel und Läufe.**
func TestDasRegelregisterDecktSichMitDemQuelltext(t *testing.T) {
	t.Parallel()

	domaene := domaenenfunktionen(t)
	erreichbar := produktivErreichbar(t, domaene)
	tests := alleTestnamen(t)

	for _, eintrag := range domain.Regelregister() {
		name := "CR " + eintrag.Nummer

		if eintrag.Aufrufer != "" {
			pruefeAufrufer(t, name, eintrag, domaene, erreichbar)
		}

		for _, test := range []string{eintrag.Positiv, eintrag.Negativ} {
			if test != "" && !tests[test] {
				t.Errorf("%s nennt %s — den Test gibt es nicht", name, test)
			}
		}
	}
}

// pruefeAufrufer stellt die ersten drei Fragen.
func pruefeAufrufer(t *testing.T, name string, eintrag domain.Regeleintrag,
	domaene map[string]funktion, erreichbar map[string]bool,
) {
	t.Helper()

	aufrufer := strings.TrimPrefix(eintrag.Aufrufer, "domain.")

	if _, gibt := domaene[aufrufer]; !gibt {
		t.Errorf("%s: den Aufrufer %s gibt es nicht", name, eintrag.Aufrufer)

		return
	}

	if !erreichbar[aufrufer] {
		t.Errorf("%s: %s wird von keiner Stelle in service oder app erreicht",
			name, eintrag.Aufrufer)
	}

	// **Wer nur von Hand gilt, hat keine Stelle** — es gibt keine, an der
	// die Regel durchgesetzt würde.
	if eintrag.Stelle == "" {
		return
	}

	stelle := strings.TrimPrefix(eintrag.Stelle, "domain.")

	eine, gibt := domaene[stelle]
	if !gibt {
		t.Errorf("%s: die Stelle %s gibt es nicht", name, eintrag.Stelle)

		return
	}

	nummer := regexp.MustCompile(`(^|[^0-9.])` + regexp.QuoteMeta(eintrag.Nummer) +
		`([^0-9a-z]|$)`)
	if !nummer.MatchString(eine.text) {
		t.Errorf("%s: %s nennt die Regel nicht", name, eintrag.Stelle)
	}

	if !erreichtVon(domaene, aufrufer, stelle) {
		t.Errorf("%s: %s erreicht %s nicht", name, eintrag.Aufrufer, eintrag.Stelle)
	}
}

// erreichtVon sagt, ob eine Funktion eine andere erreicht.
func erreichtVon(domaene map[string]funktion, start, ziel string) bool {
	gesehen := map[string]bool{start: true}
	rand := []string{start}

	for len(rand) > 0 {
		if slices.Contains(rand, ziel) {
			return true
		}

		weiter := make([]string, 0, len(rand))

		for _, schluessel := range rand {
			weiter = append(weiter, nachfolger(domaene, schluessel, gesehen)...)
		}

		rand = weiter
	}

	return false
}

// nachfolger gibt die noch nicht gesehenen Funktionen, die eine ruft.
func nachfolger(domaene map[string]funktion, schluessel string,
	gesehen map[string]bool,
) []string {
	var neue []string

	for anderer, ziel := range domaene {
		if !gesehen[anderer] && domaene[schluessel].rufe[ziel.namen] {
			gesehen[anderer] = true
			neue = append(neue, anderer)
		}
	}

	return neue
}

// produktivErreichbar gibt jede Funktion der Domäne, die service oder app
// erreichen — über beliebig viele Schritte.
func produktivErreichbar(t *testing.T, domaene map[string]funktion) map[string]bool {
	t.Helper()

	gerufen := map[string]bool{}

	for _, wurzel := range []string{"../service", "../app"} {
		for _, datei := range quelldateien(t, wurzel) {
			for name := range aufrufeIn(t, datei) {
				gerufen[name] = true
			}
		}
	}

	erreichbar := map[string]bool{}

	var rand []string

	for schluessel, eine := range domaene {
		if gerufen[eine.namen] {
			erreichbar[schluessel] = true
			rand = append(rand, schluessel)
		}
	}

	for len(rand) > 0 {
		weiter := make([]string, 0, len(rand))

		for _, schluessel := range rand {
			weiter = append(weiter, nachfolger(domaene, schluessel, erreichbar)...)
		}

		rand = weiter
	}

	return erreichbar
}

// domaenenfunktionen liest jede Funktion aus internal/domain.
func domaenenfunktionen(t *testing.T) map[string]funktion {
	t.Helper()

	funktionen := map[string]funktion{}

	for _, datei := range quelldateien(t, "../domain") {
		menge := token.NewFileSet()

		roh, fehler := os.ReadFile(datei) //nolint:gosec // feste Wurzel im Baum
		if fehler != nil {
			t.Fatalf("%s lesen: %v", datei, fehler)
		}

		gelesen, fehler := parser.ParseFile(menge, datei, roh, parser.ParseComments)
		if fehler != nil {
			t.Fatalf("%s zerlegen: %v", datei, fehler)
		}

		for _, deklaration := range gelesen.Decls {
			eine, ist := deklaration.(*ast.FuncDecl)
			if !ist || eine.Body == nil {
				continue
			}

			anfang := menge.Position(eine.Pos()).Offset
			if eine.Doc != nil {
				anfang = menge.Position(eine.Doc.Pos()).Offset
			}

			funktionen[funktionsschluessel(eine)] = funktion{
				text:  string(roh[anfang:menge.Position(eine.End()).Offset]),
				rufe:  aufrufeDes(eine.Body),
				namen: eine.Name.Name,
			}
		}
	}

	return funktionen
}

// funktionsschluessel gibt »Typ.Name« oder »Name«.
func funktionsschluessel(eine *ast.FuncDecl) string {
	if eine.Recv == nil || len(eine.Recv.List) == 0 {
		return eine.Name.Name
	}

	typ := eine.Recv.List[0].Type
	if zeiger, ist := typ.(*ast.StarExpr); ist {
		typ = zeiger.X
	}

	if name, ist := typ.(*ast.Ident); ist {
		return name.Name + "." + eine.Name.Name
	}

	return eine.Name.Name
}

// aufrufeIn gibt die Namen, die eine Datei ruft.
func aufrufeIn(t *testing.T, datei string) map[string]bool {
	t.Helper()

	gelesen, fehler := parser.ParseFile(token.NewFileSet(), datei, nil, 0)
	if fehler != nil {
		t.Fatalf("%s zerlegen: %v", filepath.Base(datei), fehler)
	}

	return aufrufeDes(gelesen)
}

// aufrufeDes sammelt die gerufenen Namen unter einem Knoten.
func aufrufeDes(knoten ast.Node) map[string]bool {
	namen := map[string]bool{}

	ast.Inspect(knoten, func(unten ast.Node) bool {
		ruf, ist := unten.(*ast.CallExpr)
		if !ist {
			return true
		}

		switch ziel := ruf.Fun.(type) {
		case *ast.SelectorExpr:
			namen[ziel.Sel.Name] = true
		case *ast.Ident:
			namen[ziel.Name] = true
		}

		return true
	})

	return namen
}

// quelldateien gibt jede Go-Datei ohne Test unter einer Wurzel.
func quelldateien(t *testing.T, wurzel string) []string {
	t.Helper()

	var dateien []string

	fehler := filepath.WalkDir(wurzel,
		func(pfad string, eintrag os.DirEntry, fehler error) error {
			if fehler == nil && !eintrag.IsDir() && strings.HasSuffix(pfad, ".go") &&
				!strings.HasSuffix(pfad, "_test.go") {
				dateien = append(dateien, pfad)
			}

			return fehler
		})
	if fehler != nil {
		t.Fatalf("%s durchgehen: %v", wurzel, fehler)
	}

	return dateien
}
