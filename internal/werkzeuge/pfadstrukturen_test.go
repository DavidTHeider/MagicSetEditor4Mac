package werkzeuge

// **Jedes Pfadfeld eines offenen Servers ist entschieden, und keines steht
// verschachtelt** — #2073.
//
// # Was er hält
//
//  1. **Ein Name, der mal einen Pfad trägt und mal nicht** — `quelle`, `ort`,
//     `bild` —, ist je Datei gezählt und entschieden. Ein neues Feld unter
//     einem dieser Namen macht ihn rot, bis jemand sagt, ob es ein Pfad ist.
//     Die Pfade darunter sind so viele, wie `wegpfadfelder` im Dienst und die
//     `Pfadfelder` der Werkzeuge nennen.
//  2. **Kein Pfadfeld steht unter einem anderen Feld.** Der Vertrag liest nur
//     die oberste Ebene eines Rumpfs (`domain.PfadeIn`), weil darunter freie
//     Daten stehen — `ziel` eines Angriffs ist ein Spieler. Ein Pfad eine
//     Ebene tiefer käme an ihm vorbei.
//
// **Angelegt, weil der erste Entwurf der Liste falsch war**: er nannte vier
// Wege, und erst die Frage nach jedem `quelle` fand den fünften — die
// Sammeldatei des Indexaufbaus, die bis `os.Open` ging.
//
// # Was er nicht hält
//
//   - **Welche Felder einer Datei die Pfade sind** — er zählt sie. Wer in
//     derselben Datei einen Pfad gegen einen anderen tauscht, bleibt grün;
//     dafür fragen `internal/rest/pfadvertragschreibweise_test.go` und sein
//     Gegenstück in `internal/mcp` jeden genannten Weg am Vertrag.
//   - **Typen aus anderen Paketen.** Ein Feld vom Typ `domain.X` sieht er
//     nicht von innen; heute nimmt kein Rumpf einen.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// mehrdeutigeNamen tragen mal einen Pfad und mal nicht.
var mehrdeutigeNamen = []string{"quelle", "ort", "bild"}

// pfadentscheidung ist, was eine Datei unter den mehrdeutigen Namen trägt.
type pfadentscheidung struct {
	pfade, andere int
	grund         string
}

const (
	nieDekodiert = "Antwort, nie dekodiert"
	bildname     = "ein Dateiname im Bildverzeichnis — `domain.Nutzer` weist einen Weg ab"
	scryfallbild = "die Adresse eines Kartenbilds bei Scryfall"

	zielEinesAngriffs = "das Ziel eines Angriffs im Stand einer Partie — ein Spieler oder eine bleibende Karte, keine Datei"
)

var mehrdeutigeFelder = map[string]pfadentscheidung{
	"internal/rest/paketplan.go":        {2, 1, "die Quelle des Plans (Adresse) und beim Ersetzen ist ein Paketverzeichnis; " + nieDekodiert},
	"internal/rest/paketpflege.go":      {2, 0, "Quelle beim Einbinden und Ort beim Übernehmen sind Verzeichnisse"},
	"internal/rest/retrievalindex.go":   {1, 0, "die Sammeldatei eines Indexaufbaus"},
	"internal/rest/ki.go":               {0, 7, "Ort ist »lokal« oder »cloud«; die Quellen nennen Anbieter und Korpus — " + nieDekodiert},
	"internal/rest/naehtepakete.go":     {0, 1, "der Ort eines Bezugs — " + nieDekodiert},
	"internal/rest/preise.go":           {0, 5, "eine Preisquelle, heute »mtgjson«"},
	"internal/rest/regelfreigabe.go":    {0, 2, "die Quelle eines Regelbereichs — " + nieDekodiert},
	"internal/rest/sammlungsausfuhr.go": {0, 1, "eine Bezeichnung, die an neuen Zeilen notiert wird"},
	"internal/rest/scryfallauskunft.go": {0, 1, nieDekodiert},
	"internal/rest/nutzer.go":           {0, 3, bildname},
	"internal/rest/scryfallsuche.go":    {0, 1, scryfallbild},

	"internal/mcp/paketplan.go":         {2, 1, "die Quelle des Plans und beim Ersetzen ist ein Paketverzeichnis; " + nieDekodiert},
	"internal/mcp/paketpflege.go":       {2, 0, "Quelle beim Installieren und Ort beim Übernehmen sind Verzeichnisse"},
	"internal/mcp/anbieter.go":          {0, 2, "ein Anbieter, heute »scrydex«"},
	"internal/mcp/bestandswerkzeuge.go": {0, 1, "der Ort eines Bezugs — " + nieDekodiert},
	"internal/mcp/kieinschaetzung.go":   {0, 2, nieDekodiert},
	"internal/mcp/ki.go":                {0, 3, "Ort ist »lokal« oder »cloud«; die Quelle eines Belegs — " + nieDekodiert},
	"internal/mcp/markt.go":             {0, 3, "ein Marktplatz, heute »cardtrader«"},
	"internal/mcp/marktkonto.go":        {0, 1, "ein Marktplatz, heute »cardtrader«"},
	"internal/mcp/preise.go":            {0, 5, "eine Preisquelle, heute »mtgjson«"},
	"internal/mcp/regelfreigabe.go":     {0, 2, nieDekodiert},
	"internal/mcp/recherche.go":         {0, 1, nieDekodiert},
	"internal/mcp/retrievalindex.go":    {0, 2, nieDekodiert + " — einen Indexaufbau hat der Werkzeugserver nicht"},
	"internal/mcp/ruecksprung.go":       {0, 1, nieDekodiert},
	"internal/mcp/sammlungsausfuhr.go":  {0, 1, "eine Bezeichnung, die an neuen Zeilen notiert wird"},
	"internal/mcp/nutzer.go":            {0, 3, bildname},
	"internal/mcp/scryfallsuche.go":     {0, 1, scryfallbild},
}

// verschachtelteOhnePfad sind Felder, unter denen ein Name aus
// `domain.IstPfadfeld` steht und trotzdem kein Pfad durch den Vertrag muss —
// mit Grund.
var verschachtelteOhnePfad = map[string]string{
	"internal/rest/partie.go angriffe.ziel": zielEinesAngriffs,
	"internal/rest/partie.go bloecke.ziel":  zielEinesAngriffs,
	"internal/mcp/partie.go angriffe.ziel":  zielEinesAngriffs,
	"internal/mcp/partie.go bloecke.ziel":   zielEinesAngriffs,

	"internal/rest/naehtepakete.go uebernommen.pfad": "der Ort eines übernommenen Pakets — " + nieDekodiert,
	"internal/mcp/paketpflege.go uebernommen.pfad":   "der Ort eines übernommenen Pakets — " + nieDekodiert,
}

// geparsteQuellen gibt die Dateien eines Pakets ohne Tests, geparst.
func geparsteQuellen(t *testing.T, paket string) map[string]*ast.File {
	t.Helper()

	namen, fehler := filepath.Glob(filepath.Join(wurzel, paket, "*.go"))
	if fehler != nil {
		t.Fatal(fehler)
	}

	menge, dateien := token.NewFileSet(), map[string]*ast.File{}

	for _, name := range namen {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}

		datei, lesefehler := parser.ParseFile(menge, name, nil, parser.SkipObjectResolution)
		if lesefehler != nil {
			t.Fatal(lesefehler)
		}

		dateien[filepath.Join(paket, filepath.Base(name))] = datei
	}

	return dateien
}

// jsonname gibt den Namen eines Felds im Rumpf — leer, wenn es keine Marke hat.
func jsonname(feld *ast.Field) string {
	if feld.Tag == nil {
		return ""
	}

	marke, _ := strconv.Unquote(feld.Tag.Value) //nolint:errcheck // eine Marke im Quelltext ist eine Zeichenkette
	name, _, _ := strings.Cut(reflect.StructTag(marke).Get("json"), ",")

	return name
}

func TestJederMehrdeutigeNameIstEntschieden(t *testing.T) {
	t.Parallel()

	gezaehlt := map[string]int{}
	genannt := map[string]int{}

	for _, paket := range []string{"internal/rest", "internal/mcp"} {
		for pfad, datei := range geparsteQuellen(t, paket) {
			ast.Inspect(datei, func(knoten ast.Node) bool {
				switch k := knoten.(type) {
				case *ast.Field:
					if slices.Contains(mehrdeutigeNamen, jsonname(k)) {
						gezaehlt[pfad]++
					}
				case *ast.CallExpr:
					if wahl, ist := k.Fun.(*ast.SelectorExpr); ist && wahl.Sel.Name == "Get" && len(k.Args) == 1 {
						if wert, lit := k.Args[0].(*ast.BasicLit); lit && wert.Kind == token.STRING {
							if name, _ := strconv.Unquote(wert.Value); slices.Contains(mehrdeutigeNamen, name) { //nolint:errcheck // eine Zeichenkette im Quelltext
								gezaehlt[pfad]++
							}
						}
					}
				case *ast.KeyValueExpr:
					genannt[paket] += pfadfeldnamen(k)
				case *ast.ValueSpec:
					if len(k.Names) == 1 && k.Names[0].Name == "wegpfadfelder" {
						genannt[paket] += wegfeldnamen(k)
					}
				}

				return true
			})
		}
	}

	pfade := map[string]int{}

	for datei, zahl := range gezaehlt {
		entscheidung, gibt := mehrdeutigeFelder[datei]
		if !gibt {
			t.Errorf("%s: %d Felder unter %v, und keines ist entschieden — trägt eines einen Pfad?", datei, zahl, mehrdeutigeNamen)

			continue
		}

		if entscheidung.pfade+entscheidung.andere != zahl {
			t.Errorf("%s: %d Felder unter %v, entschieden sind %d — ein neues trägt vielleicht einen Pfad",
				datei, zahl, mehrdeutigeNamen, entscheidung.pfade+entscheidung.andere)
		}

		pfade[filepath.Dir(datei)] += entscheidung.pfade
	}

	for datei := range mehrdeutigeFelder {
		if gezaehlt[datei] == 0 {
			t.Errorf("%s ist entschieden und trägt keines mehr", datei)
		}
	}

	for _, paket := range []string{"internal/rest", "internal/mcp"} {
		if pfade[paket] != genannt[paket] {
			t.Errorf("%s: %d Pfade unter mehrdeutigen Namen, der Vertrag kennt %d", paket, pfade[paket], genannt[paket])
		}
	}
}

// pfadfeldnamen zählt die Namen hinter `Pfadfelder:` eines Werkzeugs.
func pfadfeldnamen(paar *ast.KeyValueExpr) int {
	if schluessel, ist := paar.Key.(*ast.Ident); !ist || schluessel.Name != "Pfadfelder" {
		return 0
	}

	return zeichenketten(paar.Value)
}

// wegfeldnamen zählt die Namen in den Werten von `wegpfadfelder` — die
// Schlüssel sind Wege.
func wegfeldnamen(spez *ast.ValueSpec) int {
	zahl := 0

	for _, wert := range spez.Values {
		if liste, ist := wert.(*ast.CompositeLit); ist {
			for _, eintrag := range liste.Elts {
				if paar, paarweise := eintrag.(*ast.KeyValueExpr); paarweise {
					zahl += zeichenketten(paar.Value)
				}
			}
		}
	}

	return zahl
}

// zeichenketten zählt die Zeichenketten unter einem Knoten.
func zeichenketten(knoten ast.Node) int {
	zahl := 0

	ast.Inspect(knoten, func(innen ast.Node) bool {
		if wert, ist := innen.(*ast.BasicLit); ist && wert.Kind == token.STRING {
			zahl++
		}

		return true
	})

	return zahl
}

func TestKeinPfadfeldStehtVerschachtelt(t *testing.T) {
	t.Parallel()

	gesehen := map[string]bool{}

	for _, paket := range []string{"internal/rest", "internal/mcp"} {
		dateien := geparsteQuellen(t, paket)
		benannte := benannteStrukturen(dateien)

		for pfad, datei := range dateien {
			ast.Inspect(datei, func(knoten ast.Node) bool {
				if struktur, ist := knoten.(*ast.StructType); ist {
					for _, fund := range verschachtelt(struktur, benannte) {
						gesehen[pfad+" "+fund] = true

						if _, frei := verschachtelteOhnePfad[pfad+" "+fund]; !frei {
							t.Errorf("%s: %s — der Vertrag liest nur die oberste Ebene eines Rumpfs", pfad, fund)
						}
					}
				}

				return true
			})
		}
	}

	for eintrag := range verschachtelteOhnePfad {
		if !gesehen[eintrag] {
			t.Errorf("%s steht in der Liste und nicht mehr im Quelltext", eintrag)
		}
	}
}

// benannteStrukturen sind die Verbundtypen eines Pakets nach Namen.
func benannteStrukturen(dateien map[string]*ast.File) map[string]*ast.StructType {
	benannte := map[string]*ast.StructType{}

	for _, datei := range dateien {
		ast.Inspect(datei, func(knoten ast.Node) bool {
			if spez, ist := knoten.(*ast.TypeSpec); ist {
				if struktur, verbund := spez.Type.(*ast.StructType); verbund {
					benannte[spez.Name.Name] = struktur
				}
			}

			return true
		})
	}

	return benannte
}

// verschachtelt nennt jedes Pfadfeld, das eine Ebene unter einem Feld des
// Verbunds steht — als »feld.pfadfeld«. Ein eingebettetes Feld steht oben:
// der Dekodierer hebt seine Felder in den Rumpf.
func verschachtelt(struktur *ast.StructType, benannte map[string]*ast.StructType) []string {
	funde := []string{}

	for _, feld := range struktur.Fields.List {
		innen := innereStruktur(feld.Type, benannte)
		if len(feld.Names) == 0 || innen == nil {
			continue
		}

		for _, innenfeld := range innen.Fields.List {
			if name := jsonname(innenfeld); domain.IstPfadfeld(name) {
				funde = append(funde, jsonname(feld)+"."+name)
			}
		}
	}

	return funde
}

// innereStruktur gibt den Verbund hinter einem Feldtyp — durch Zeiger,
// Listen und die Werte einer Abbildung hindurch.
func innereStruktur(typ ast.Expr, benannte map[string]*ast.StructType) *ast.StructType {
	for {
		switch t := typ.(type) {
		case *ast.StarExpr:
			typ = t.X
		case *ast.ArrayType:
			typ = t.Elt
		case *ast.MapType:
			typ = t.Value
		case *ast.StructType:
			return t
		case *ast.Ident:
			return benannte[t.Name]
		default:
			return nil
		}
	}
}
