// **Was „Bindung am Fenster" heißt — und was der Zähler davon sieht**
// (#1287).
//
// # Der Befund
//
// `docs/FAEHIGKEITEN.md` nannte am 2026-08-29 „143 Bindungen am Fenster", und
// die Zahl war **gewacht**: `flaechengroesse_test.go` zählte
// `^func (b *Bindungen) ` und kam auf 143. Wails erzeugte im selben Baum
// **142** Aufrufe (`frontend/wailsjs/go/app/Bindungen.d.ts`, 142 × `export
// function`). Die Differenz sind `waehlenUndOeffnen`, `dienstStarten` und
// `dienstBeenden` — **unexportierte**
// Hilfsmethode, die Wails nicht bindet und die im erzeugten JS und TS nicht
// vorkommt. **Seit dem Nachlauf über ID-02** (#1701) kommen die drei Zugänge
// des Sitzungstors dazu: `app`, `tor` und `frei`.
//
// Der Wächter maß „Methoden auf `*Bindungen`", der Satz behauptete „Bindungen
// am Fenster". **Die Reichweite einer Messung ist die Reichweite ihrer
// Aussage** (§10) — hier reichte sie um genau eine Methode zu weit, und weil
// die Zahl grün war, sah es niemand.
//
// # Was dieser Zähler hält
//
// Er zählt die **exportierten** Methoden auf `*Bindungen` in `internal/app`.
// Das ist die Menge, aus der Wails seine Aufrufe erzeugt: die Bibliothek liest
// die Methoden des gebundenen Werts über Reflexion, und `reflect` kennt keine
// unexportierten Methoden eines fremden Pakets.
//
// # Nur was unmittelbar an `*Bindungen` steht
//
// **Und das trägt nur, solange der Verbund nichts einbettet.** Eine
// exportierte Methode eines eingebetteten Typs gehört zur Menge, die
// `reflect` sieht, hat aber keine eigene `func (b *Bindungen) …`-Zeile — der
// Zähler unterschlüge sie. Der Zweitblick über Phase 4.0 hat es
// nachgestellt: ein eingebettetes `SichtbareProbe()` ließ ihn grün, während
// der Reflexionswächter in `internal/app` rot wurde (Befund 3).
//
// Die Annahme steht deshalb nicht mehr im Kommentar, sondern als Lauf:
// `TestDerBindungsverbundBettetNichtsEin`. Solange er grün ist, sind „direkt
// erklärt" und „von Wails gesehen" dieselbe Menge.
//
// # Was er nicht hält
//
// **Dass die erzeugte Datei danebenliegt.** `frontend/wailsjs/` ist Generat
// und steht in `.gitignore` (§2); ein Wächter, der sie läse, wäre auf einem
// frischen Klon blind — und ein übersprungener Wächter ist keiner. Was die
// beiden Seiten wirklich zusammenhält, ist der Brückenlauf aus Phase 4.1
// (#1288),
// nicht diese Zählung.
//
// **Und nicht, ob eine Bindung eine Fähigkeit ist.** Das sagt der Katalog.
package werkzeuge

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// bindungsnamen gibt die Methodennamen auf `*Bindungen` aus einem Quelltext.
//
// **`nurExportierte` ist der ganze Unterschied** zwischen der Zahl im
// Einleitungssatz und der Zahl im Fähigkeitskatalog: der Satz spricht über die
// Fläche, die Wails erzeugt, der Katalog über jeden Weg des Fensters in eine
// Fähigkeit — auch den, der über eine unexportierte Hilfe führt.
func bindungsnamen(baum *ast.File, nurExportierte bool) []string {
	var namen []string

	for _, erklaerung := range baum.Decls {
		verfahren, istVerfahren := erklaerung.(*ast.FuncDecl)
		if !istVerfahren || verfahren.Recv == nil || len(verfahren.Recv.List) != 1 {
			continue
		}

		zeiger, istZeiger := verfahren.Recv.List[0].Type.(*ast.StarExpr)
		if !istZeiger {
			continue
		}

		name, istName := zeiger.X.(*ast.Ident)
		if !istName || name.Name != "Bindungen" {
			continue
		}

		if nurExportierte && !verfahren.Name.IsExported() {
			continue
		}

		namen = append(namen, verfahren.Name.Name)
	}

	return namen
}

// bindungenAusQuelltext ist der Weg für die Gegenprobe: er liest aus einer
// Zeichenkette statt aus dem Baum des Projekts.
func bindungenAusQuelltext(t *testing.T, quelltext string, nurExportierte bool) []string {
	t.Helper()

	baum, err := parser.ParseFile(token.NewFileSet(), "probe.go", quelltext,
		parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("die Probe lässt sich nicht zerteilen: %v", err)
	}

	return bindungsnamen(baum, nurExportierte)
}

// bindungenAusVerzeichnis liest die Methoden auf `*Bindungen` unter `internal/app`.
func bindungenAusVerzeichnis(t *testing.T, nurExportierte bool) []string {
	t.Helper()

	var namen []string

	menge := token.NewFileSet()

	fehler := filepath.WalkDir(filepath.Join(wurzel, "internal", "app"),
		func(pfad string, eintrag os.DirEntry, err error) error {
			if err != nil || eintrag.IsDir() ||
				!strings.HasSuffix(pfad, ".go") || strings.HasSuffix(pfad, "_test.go") {
				return err
			}

			baum, lesefehler := parser.ParseFile(menge, pfad, nil, parser.SkipObjectResolution)
			if lesefehler != nil {
				return lesefehler
			}

			namen = append(namen, bindungsnamen(baum, nurExportierte)...)

			return nil
		})
	if fehler != nil {
		t.Fatalf("internal/app durchgehen: %v", fehler)
	}

	if len(namen) == 0 {
		t.Fatal("in internal/app fand sich keine Bindung — der Leser greift daneben")
	}

	return namen
}

// **Die Gegenprobe steht als Lauf da, nicht als Erinnerung an einen Versuch.**
//
// §10 verlangt sie: *ein Wächter beweist nichts, bevor eine gezielte
// Verletzung ihn rot gemacht hat.* Eine Gegenprobe von Hand ist einmal
// gelaufen und danach eine Behauptung — diese läuft bei jedem `make ci` mit
// und beschreibt zugleich, worin der Unterschied besteht.
func TestDerBindungszaehlerZaehltNurExportierte(t *testing.T) {
	t.Parallel()

	const probe = `package app

type Bindungen struct{}

func (b *Bindungen) Sets() error { return nil }

func (b *Bindungen) waehlenUndOeffnen() error { return nil }

func (x *Bindungen) Karten() error { return nil }

func (b Bindungen) AmWert() error { return nil }

func Frei() error { return nil }
`

	t.Run("exportiert", func(t *testing.T) {
		t.Parallel()

		namen := bindungenAusQuelltext(t, probe, true)
		if will := []string{"Sets", "Karten"}; !slices.Equal(namen, will) {
			t.Errorf("gezählt %v — erwartet %v", namen, will)
		}
	})

	// **Der andere Zweig gehört dazu.** Ohne ihn bliebe offen, ob der Zähler
	// die unexportierte Methode *auslässt* oder sie überhaupt nie sieht — und
	// eine Gegenprobe, die das nicht unterscheidet, ist keine (§10).
	t.Run("alle", func(t *testing.T) {
		t.Parallel()

		namen := bindungenAusQuelltext(t, probe, false)
		if will := []string{"Sets", "waehlenUndOeffnen", "Karten"}; !slices.Equal(namen, will) {
			t.Errorf("gezählt %v — erwartet %v", namen, will)
		}
	})
}

// **Und die Differenz ist genau eine benannte Methode.**
//
// Ohne diesen Lauf hinge die Erklärung oben in der Luft: die beiden Zahlen
// gingen auseinander, und niemand wüsste, *welche* Methode dazwischen steht.
func TestDerUnterschiedDerBeidenZaehlungenIstBenannt(t *testing.T) {
	t.Parallel()

	alle := bindungenAusVerzeichnis(t, false)
	exportierte := bindungenAusVerzeichnis(t, true)

	var unexportierte []string

	for _, name := range alle {
		if !ast.IsExported(name) {
			unexportierte = append(unexportierte, name)
		}
	}

	if len(alle)-len(exportierte) != len(unexportierte) {
		t.Fatalf("%d insgesamt, %d exportiert, %d unexportiert — das geht nicht auf",
			len(alle), len(exportierte), len(unexportierte))
	}

	// **Drei seit N6** (#1572):
	// `dienstStarten` und `dienstBeenden` sind der gemeinsame Weg beider
	// schaltbaren Dienste. Wails erzeugt für sie keinen Aufruf — sie sind
	// ein Weg des Fensters in eine Fähigkeit, wie `waehlenUndOeffnen`.
	//
	// **Und drei seit dem Nachlauf über ID-02** (#1701): `app`, `tor` und
	// `frei` sind die Zugänge durch das Sitzungstor. **`ruf` seit BK-01**
	// (#2109): er gibt den Kontext des Rufs hinter dem Tor.
	if will := []string{
		"app", "tor", "frei", "ruf", "waehlenUndOeffnen", "dienstStarten", "dienstBeenden",
	}; !slices.Equal(unexportierte, will) {
		t.Errorf("unexportiert sind %v — der Kommentar oben nennt %v; "+
			"wer hier etwas ändert, ändert auch ihn", unexportierte, will)
	}
}

// **Der Verbund `Bindungen` bettet nichts ein.**
//
// Ohne diesen Lauf ist der Zähler darüber eine Behauptung: er liest
// `func (b *Bindungen) …`-Zeilen, und Wails liest die Methodenmenge über
// `reflect`. Beide sind dieselbe Menge **genau dann**, wenn kein eingebetteter
// Typ Methoden hineinreicht.
//
// **Warum nicht einfach eingebettete Methoden mitzählen.** Weil das
// Typinformation braucht und nicht den Syntaxbaum: `eingebetteterTyp` kann in
// einem anderen Paket stehen. Ein Zähler, der dafür das ganze Programm
// übersetzte, wäre teurer und stünde vor derselben Frage — welche Methoden
// sind exportiert. Die Einbettung zu verbieten ist die kleinere Zusage, und
// sie ist die, die dieses Projekt ohnehin einhält.
func TestDerBindungsverbundBettetNichtsEin(t *testing.T) {
	t.Parallel()

	menge := token.NewFileSet()

	baum, err := parser.ParseFile(menge,
		filepath.Join(wurzel, "internal", "app", "bindungen.go"), nil,
		parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("bindungen.go zerteilen: %v", err)
	}

	verbund := verbundAusBaum(t, baum, "Bindungen")

	for _, feld := range verbund.Fields.List {
		if len(feld.Names) == 0 {
			t.Errorf("`Bindungen` bettet %s ein — dessen exportierte Methoden "+
				"bindet Wails mit, und der Zähler in dieser Datei sieht sie nicht",
				typname(feld.Type))
		}
	}
}

// verbundAusBaum gibt die Feldliste eines benannten Verbunds.
func verbundAusBaum(t *testing.T, baum *ast.File, name string) *ast.StructType {
	t.Helper()

	for _, erklaerung := range baum.Decls {
		allgemein, ist := erklaerung.(*ast.GenDecl)
		if !ist || allgemein.Tok != token.TYPE {
			continue
		}

		for _, spezifisch := range allgemein.Specs {
			typ, istTyp := spezifisch.(*ast.TypeSpec)
			if !istTyp || typ.Name.Name != name {
				continue
			}

			verbund, istVerbund := typ.Type.(*ast.StructType)
			if !istVerbund {
				t.Fatalf("%s ist kein Verbund mehr — der Leser greift daneben", name)
			}

			return verbund
		}
	}

	t.Fatalf("den Verbund %s gibt es nicht mehr — der Leser greift daneben", name)

	return nil
}

// typname gibt einen lesbaren Namen für einen eingebetteten Typ.
func typname(ausdruck ast.Expr) string {
	switch genau := ausdruck.(type) {
	case *ast.Ident:
		return genau.Name
	case *ast.StarExpr:
		return "*" + typname(genau.X)
	case *ast.SelectorExpr:
		return typname(genau.X) + "." + genau.Sel.Name
	default:
		return "einen unbenannten Typ"
	}
}
