package schichten

// **Der Transport prüft, wohin er verbindet** — SEC-P4
// (#2000).
//
// # Die zweite Frage an die Netzgrenze
//
// [TestNurBezugKenntDasNetz] fragt, **wer** ins Netz darf, und lässt `netz`
// als den einen Transport durch. Bis #2000 stand dort „wohin, sagt der
// Aufrufer" — und der Transport wählte jede Adresse, die man ihm nannte,
// auch `127.0.0.1`, `169.254.169.254` und das eigene Netz, und folgte jeder
// Weiterleitung. Seitdem prüft er das Ziel in `internal/netz/ziel.go`.
//
// # Was dieser Wächter hält
//
// **Jeder `net.Dialer` in `internal/netz` setzt `Control`, und jeder
// `http.Client` setzt `CheckRedirect`.** Ein zweiter Wähler ohne die Prüfung
// nach der Auflösung, oder ein Klient mit Gos Weiterleitungsvorgabe, wäre ein
// Weg an der Zielprüfung vorbei — und er sähe aus wie jeder andere.
//
// # Was er nicht hält
//
// **Ob die Prüfung das Richtige sperrt** — das hält die Tafel in
// `internal/netz/ziel_intern_test.go`, und ob sie an der Fläche greift, die
// Läufe `internal/flaechen/zielsperre*_test.go`. **Und keinen Wähler
// außerhalb von `netz`**: `internal/leitung` wählt ein Gegenüber ohne
// Bereichssperre, weil das eigene Netz dort der Zweck ist; die Annahme
// einer Einladung steht dafür hinter `externSenden`.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// pflichtfelder nennt je Typ das Feld, ohne das er die Zielprüfung umginge.
var pflichtfelder = map[string]string{
	"net.Dialer":  "Control",
	"http.Client": "CheckRedirect",
}

func TestJederWegDesTransportsPrueftDasZiel(t *testing.T) {
	t.Parallel()

	ordner := filepath.Join(internalVerzeichnis, "netz")

	dateien, fehler := filepath.Glob(filepath.Join(ordner, "*.go"))
	if fehler != nil {
		t.Fatal(fehler)
	}

	gesehen := 0

	for _, datei := range dateien {
		if strings.HasSuffix(datei, "_test.go") {
			continue
		}

		roh, lesefehler := os.ReadFile(datei) //nolint:gosec // Quelltext dieses Repositoriums
		if lesefehler != nil {
			t.Fatal(lesefehler)
		}

		befunde, gezaehlt := ohneZielpruefung(t, filepath.Base(datei), string(roh))
		gesehen += gezaehlt

		for _, befund := range befunde {
			t.Error(befund)
		}
	}

	// **Ohne einen einzigen Fund prüfte der Wächter nichts** — ein Umbau,
	// der die Literale verschöbe, ließe ihn grün und taub (§10).
	if gesehen == 0 {
		t.Fatal("in internal/netz steht kein net.Dialer und kein http.Client — der Wächter greift daneben")
	}
}

// **Die Gegenprobe als Testfall** (§10): ein Wähler ohne `Control` und ein
// Klient ohne `CheckRedirect` werden erkannt, einer mit beiden nicht.
func TestDerZielwaechterWirdRot(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		quelle   string
		befunde  int
		gezaehlt int
	}{
		"ohne beides": {
			`package netz
func a() { _ = &net.Dialer{Timeout: 1}; _ = &http.Client{Transport: nil} }`, 2, 2,
		},
		"mit beidem": {
			`package netz
func a() { _ = &net.Dialer{Control: k}; _ = &http.Client{CheckRedirect: w} }`, 0, 2,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			befunde, gezaehlt := ohneZielpruefung(t, "probe.go", fall.quelle)
			if len(befunde) != fall.befunde || gezaehlt != fall.gezaehlt {
				t.Errorf("%d Befunde bei %d Literalen, erwartet %d bei %d: %v",
					len(befunde), gezaehlt, fall.befunde, fall.gezaehlt, befunde)
			}
		})
	}
}

// ohneZielpruefung findet die Literale, denen ihr Pflichtfeld fehlt — und
// zählt, wie viele es überhaupt gab.
func ohneZielpruefung(t *testing.T, name, quelle string) (befunde []string, gezaehlt int) {
	t.Helper()

	satz := token.NewFileSet()

	baum, fehler := parser.ParseFile(satz, name, quelle, 0)
	if fehler != nil {
		t.Fatalf("%s: %v", name, fehler)
	}

	ast.Inspect(baum, func(knoten ast.Node) bool {
		literal, ist := knoten.(*ast.CompositeLit)
		if !ist {
			return true
		}

		typ, ist := literal.Type.(*ast.SelectorExpr)
		if !ist {
			return true
		}

		paket, ist := typ.X.(*ast.Ident)
		if !ist {
			return true
		}

		pflicht, gefuehrt := pflichtfelder[paket.Name+"."+typ.Sel.Name]
		if !gefuehrt {
			return true
		}

		gezaehlt++

		if !hatFeld(literal, pflicht) {
			befunde = append(befunde, satz.Position(literal.Pos()).String()+": "+
				paket.Name+"."+typ.Sel.Name+" ohne "+pflicht+" — ein Weg an der Zielprüfung vorbei (#2000)")
		}

		return true
	})

	return befunde, gezaehlt
}

// hatFeld sagt, ob ein Literal ein Feld mit diesem Namen setzt.
func hatFeld(literal *ast.CompositeLit, feld string) bool {
	for _, element := range literal.Elts {
		paar, ist := element.(*ast.KeyValueExpr)
		if !ist {
			continue
		}

		if schluessel, ist := paar.Key.(*ast.Ident); ist && schluessel.Name == feld {
			return true
		}
	}

	return false
}
