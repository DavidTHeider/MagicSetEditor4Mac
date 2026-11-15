package schichten

// **Die Hüllen um den Asset-Server öffnen nichts** — B01/B03
// (#2220,
// #2222).
//
// `fenster/rufbuch.go` und `fenster/rufstoerung.go` stehen in
// `mitNetzDatei`, weil sie `net/http` für die **Typen** eines `http.Handler`
// brauchen. Ohne Schalter gebaut sind sie, damit ihre Tests in der Teststufe
// laufen — wie `abnahme.go`; benutzt werden sie nur von `abnahme_darwin.go`.
// **Dieser Wächter hält die andere Hälfte des Eintrags**: aus `net/http`
// nehmen sie nur, was einen Ruf entgegennimmt und beantwortet — keinen
// Klienten, keinen Lauscher, keinen Server.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// nurHuelle ist, was eine Hülle aus `net/http` nehmen darf.
var nurHuelle = map[string]bool{
	"Handler": true, "HandlerFunc": true, "Request": true,
	"ResponseWriter": true, "MethodPost": true,
}

// fremdeHTTPTeile nennt, was eine Datei aus dem Netz nimmt und eine Hülle
// nicht braucht — **über den Syntaxbaum und den Namen, unter dem `net/http`
// eingeführt ist**: ein Muster über `http.X` sah `web "net/http"` und einen
// Punktimport nicht (Zweitblick über Phase B, Befund 10). `Status…` ist eine
// Zahl und darf; `net` selbst, ein Punktimport und jedes andere
// Unterpaket von `net/http` dürfen nicht.
func fremdeHTTPTeile(quelle string) ([]string, error) {
	datei, fehler := parser.ParseFile(token.NewFileSet(), "", quelle, parser.ImportsOnly|parser.SkipObjectResolution)
	if fehler != nil {
		return nil, fmt.Errorf("lesen: %w", fehler)
	}

	var fremd []string

	name := ""

	for _, einfuhr := range datei.Imports {
		pfad, _ := strconv.Unquote(einfuhr.Path.Value) //nolint:errcheck // der Parser hat den Pfad schon als Zeichenkette gelesen

		switch {
		case pfad == "net/http" && einfuhr.Name != nil && einfuhr.Name.Name == ".":
			fremd = append(fremd, ". \"net/http\"")
		case pfad == "net/http" && einfuhr.Name != nil:
			name = einfuhr.Name.Name
		case pfad == "net/http":
			name = "http"
		case pfad == "net" || strings.HasPrefix(pfad, "net/http/"):
			fremd = append(fremd, pfad)
		}
	}

	if name == "" {
		return fremd, nil
	}

	ganz, fehler := parser.ParseFile(token.NewFileSet(), "", quelle, parser.SkipObjectResolution)
	if fehler != nil {
		return nil, fmt.Errorf("lesen: %w", fehler)
	}

	ast.Inspect(ganz, func(knoten ast.Node) bool {
		auswahl, ist := knoten.(*ast.SelectorExpr)
		if !ist {
			return true
		}

		if wer, ident := auswahl.X.(*ast.Ident); ident && wer.Name == name &&
			!nurHuelle[auswahl.Sel.Name] && !strings.HasPrefix(auswahl.Sel.Name, "Status") {
			fremd = append(fremd, auswahl.Sel.Name)
		}

		return true
	})

	return fremd, nil
}

func TestDieHuellenImFensterOeffnenNichts(t *testing.T) {
	t.Parallel()

	// **Jede zugelassene Datei außer der Abnahmeleitung** — sie horcht als
	// einzige wirklich. Bis F02 (#2341) las der Wächter nur `fenster/ruf…`,
	// und eine neue Hülle unter anderem Namen wäre an ihm vorbeigegangen.
	for datei := range mitNetzDatei {
		if datei == "fenster/abnahme_darwin.go" {
			continue
		}

		quelle, fehler := os.ReadFile(filepath.Join("..", datei)) //nolint:gosec // Quelltext dieses Repositoriums
		if fehler != nil {
			t.Fatal(fehler)
		}

		fremd, fehler := fremdeHTTPTeile(string(quelle))
		if fehler != nil {
			t.Fatal(fehler)
		}

		if len(fremd) > 0 {
			t.Errorf("internal/%s nimmt aus dem Netz %v — eine Hülle um den Asset-Server braucht nur Handler, Request und ResponseWriter", datei, fremd)
		}
	}
}

// Die Gegenprobe — als Testfall (§10). **Mit ganzen Dateien**: ein Import
// unter anderem Namen (`web "net/http"`) oder als Punkt entging einem Muster,
// das nur `http.X` las (Zweitblick über Phase B, Befund 10).
func TestDerHuellenwaechterWirdRot(t *testing.T) {
	t.Parallel()

	datei := func(einfuhr, rumpf string) string {
		return "package fenster\n\nimport " + einfuhr + "\n\nfunc f() {\n" + rumpf + "\n}\n"
	}

	faelle := map[string]int{
		datei(`"net/http"`, `var _ http.Handler; _ = http.StatusOK`): 0,
		datei(`"net/http"`, `_, _ = http.Get("x")`):                  1,
		datei(`web "net/http"`, `_, _ = web.Get("x")`):               1,
		datei(`web "net/http"`, `var _ web.ResponseWriter`):          0,
		datei(`. "net/http"`, `_ = StatusOK`):                        1,
		datei(`"net"`, `_, _ = net.Listen("tcp", ":0")`):             1,
		datei(`"net/http"`, `_ = &http.Server{}; _ = http.Client{}`): 2,
	}

	for quelle, erwartet := range faelle {
		fremd, fehler := fremdeHTTPTeile(quelle)
		if fehler != nil {
			t.Fatalf("%q: %v", quelle, fehler)
		}

		if len(fremd) != erwartet {
			t.Errorf("%q: %d fremde Teile %v, erwartet %d", quelle, len(fremd), fremd, erwartet)
		}
	}
}
