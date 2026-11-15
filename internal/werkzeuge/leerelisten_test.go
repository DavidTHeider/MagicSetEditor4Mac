// **Eine leere Liste bleibt eine leere Liste**
// (#1169,
// Werkstück 3).
//
// # Die Naht, vor der §3 warnt
//
// „Frontend-Tests gehen über `test/api.ts` und laufen nie durch die
// JSON-Serialisierung; Go-Tests sehen die Struct-Marken nicht. Ein
// `json:\",omitempty\"` auf einem Slice lässt eine **leere** Liste aus der
// Antwort verschwinden, und die Seite bleibt leer — bei grüner Pipeline."
//
// Das steht dort seit Phase 1 als **Anweisung an einen Menschen**: vor jeder
// Freigabe die gebaute Anwendung starten. Ein Mensch, der jede Bindung des
// Fensters durchsieht, findet es nicht.
//
// **Hier stand eine Zahl**, und sie war um drei zu klein — Fließtext, den
// niemand nachzählt. Wie viele es sind, sagt `docs/FAEHIGKEITEN.md`; wie viele
// ein Mensch durchsieht, ist für diesen Satz ohnehin gleich (#1287).
//
// # Was dieser Wächter hält
//
// Kein Feld vom Typ Slice oder Map in `internal/app` trägt `omitempty`. Das
// ist die eine Hälfte der Naht, und sie ist mechanisch entscheidbar.
//
// # Er kennt auch benannte Listen
//
// **Seit dem Zweitblick über Phase 3.8** (Befund 5). Der erste Anlauf sah nur
// `[]T` und `map[K]V` unmittelbar am Feld; ein `type Namen []string` mit
// `Felder Namen ` + "`" + `json:"felder,omitempty"` + "`" + ` war für ihn ein `*ast.Ident` und kam
// durch — und der Laufzeitwächter daneben sieht ein **fehlendes** Feld
// überhaupt nicht. Der genannte Fall passierte damit beide.
//
// Aufgelöst wird deshalb innerhalb der Schicht: jede `type X …`-Erklärung in
// `internal/app`, deren Rumpf eine Liste ist, zählt als Liste.
//
// # Was er **nicht** hält, und das ist die andere Hälfte
//
// **Ein `nil`-Slice wird zu `null`, auch ohne `omitempty`.** Dagegen hilft nur
// `make([]T, 0)` an jeder Stelle, die die Antwort baut — und ob das geschehen
// ist, sieht kein Muster.
//
// **Seit dem 2026-08-29 sieht es ein Lauf**:
// `internal/app/leerelistenlaufzeit_test.go` ruft jede Bindung und liest ihre
// Antwort gegen den Go-Typ. Was er beim ersten Mal fand, sagt, wie groß die
// Lücke war: **132 Fundstellen in 112 von 142 Bindungen**
// (#1288).
// Er deckt die Fehlerwege; die Läufe je Bindung
// (`bildausfuhr_test.go`, `bestand_test.go`, `ersetzen_test.go` und andere)
// decken die Erfolgswege, und die sind von Hand geschrieben.
//
// **Und das Fenster selbst fährt in keinem Flächenlauf.** Wails erzeugt seine
// Aufrufe zur Laufzeit; ein Lauf ohne Fenster erreicht sie nicht. Deshalb
// bleibt der Satz aus §3 stehen: vor jeder Freigabe die **gebaute** Anwendung
// starten.
package werkzeuge

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fassadenverzeichnis ist die Schicht, deren DTOs ans Fenster gehen.
const fassadenverzeichnis = "../app"

func TestKeineLeereListeVerschwindetAusEinerAntwort(t *testing.T) {
	menge := token.NewFileSet()

	benannte := benannteListen(t, menge)

	fehler := filepath.WalkDir(fassadenverzeichnis,
		func(pfad string, eintrag os.DirEntry, err error) error {
			if err != nil || eintrag.IsDir() || !strings.HasSuffix(pfad, ".go") ||
				strings.HasSuffix(pfad, "_test.go") {
				return err
			}

			baum, lesefehler := parser.ParseFile(menge, pfad, nil, parser.SkipObjectResolution)
			if lesefehler != nil {
				return lesefehler
			}

			for _, befund := range listenMitAuslassung(baum, benannte) {
				t.Errorf("%s: %s trägt `omitempty` auf einer Liste — eine leere "+
					"Liste verschwände damit aus der Antwort, und die Seite bliebe "+
					"leer bei grüner Pipeline (AGENTS.md §3)", filepath.Base(pfad), befund)
			}

			return nil
		})
	if fehler != nil {
		t.Fatalf("%s durchgehen: %v", fassadenverzeichnis, fehler)
	}
}

// benannteListen sammelt die Typnamen der Schicht, die Listen sind.
//
// **Nur diese Schicht.** Ein Typ aus einem anderen Paket steht als
// `paket.Name` da, und ihn aufzulösen hieße, den ganzen Baum zu laden. Was das
// kostet, steht im Dateikopf: ein `domain.Befunde` mit `omitempty` bliebe
// unbemerkt. Es gibt heute keines, und der Laufzeitwächter fährt die Antworten
// zusätzlich.
//
// **Ketten werden aufgelöst.** `type Namen Basis` mit `type Basis []string`
// ist eine Liste, und der erste Anlauf sah sie nicht — angestrichen vom
// Nachlauf über Phase 3.8, Befund 6
// (#1253).
//
// **Was weiterhin nicht aufgelöst wird**: eine generische Verwendung
// (`Liste[string]`) und ein Zeiger auf eine Liste. Für den Zeiger ist das
// richtig — `*[]string` mit `omitempty` verschwindet bei `nil`, und das ist
// die Absicht; für die generische Form gibt es in dieser Schicht heute keinen
// Fall.
func benannteListen(t *testing.T, menge *token.FileSet) map[string]bool {
	t.Helper()

	listen := map[string]bool{}

	// ketten hält `type Namen Basis` — aufgelöst wird danach, denn `Basis`
	// kann in einer Datei stehen, die erst später an die Reihe kommt.
	ketten := map[string]string{}

	fehler := filepath.WalkDir(fassadenverzeichnis,
		func(pfad string, eintrag os.DirEntry, err error) error {
			if err != nil || eintrag.IsDir() || !strings.HasSuffix(pfad, ".go") {
				return err
			}

			baum, lesefehler := parser.ParseFile(menge, pfad, nil, parser.SkipObjectResolution)
			if lesefehler != nil {
				return lesefehler
			}

			ast.Inspect(baum, func(knoten ast.Node) bool {
				erklaerung, istErklaerung := knoten.(*ast.TypeSpec)
				if !istErklaerung {
					return true
				}

				if istRoheListe(erklaerung.Type) {
					listen[erklaerung.Name.Name] = true

					return true
				}

				// **Und die Ketten**: `type Basis []string; type Namen Basis`.
				// Wer nur die rohe Form sammelt, sieht die zweite nicht —
				// angestrichen vom Nachlauf über Phase 3.8, Befund 6
				// (#1253).
				if hinweis, istName := erklaerung.Type.(*ast.Ident); istName {
					ketten[erklaerung.Name.Name] = hinweis.Name
				}

				return true
			})

			return nil
		})
	if fehler != nil {
		t.Fatalf("%s durchgehen: %v", fassadenverzeichnis, fehler)
	}

	// **Bis sich nichts mehr ändert.** Eine Kette kann mehrere Glieder haben,
	// und die Dateien kommen in keiner bestimmten Reihenfolge.
	for aufgeloest := true; aufgeloest; {
		aufgeloest = false

		for name, hinweis := range ketten {
			if listen[hinweis] && !listen[name] {
				listen[name] = true
				aufgeloest = true
			}
		}
	}

	return listen
}

// listenMitAuslassung gibt die Feldnamen, die beides tragen.
func listenMitAuslassung(baum *ast.File, benannte map[string]bool) []string {
	var befunde []string

	ast.Inspect(baum, func(knoten ast.Node) bool {
		verbund, istVerbund := knoten.(*ast.StructType)
		if !istVerbund || verbund.Fields == nil {
			return true
		}

		for _, feld := range verbund.Fields.List {
			if !istListe(feld.Type, benannte) || !hatAuslassung(feld.Tag) {
				continue
			}

			befunde = append(befunde, listenfeldname(feld))
		}

		return true
	})

	return befunde
}

// istListe sagt, ob ein Typ zu `null` wird, wenn er `nil` ist.
//
// **Slice und Map, nicht Zeiger.** Ein `*Fehler` mit `omitempty` ist richtig:
// „kein Fehler" soll aus der Antwort verschwinden, und das Frontend prüft auf
// `undefined`. Eine leere **Liste** dagegen ist eine Aussage.
func istListe(typ ast.Expr, benannte map[string]bool) bool {
	if istRoheListe(typ) {
		return true
	}

	// **Und ein Name, der in dieser Schicht eine Liste ist** (Befund 5 des
	// Zweitblicks über Phase 3.8).
	name, istName := typ.(*ast.Ident)

	return istName && benannte[name.Name]
}

// istRoheListe sagt, ob ein Typ unmittelbar als Liste dasteht.
func istRoheListe(typ ast.Expr) bool {
	switch gestalt := typ.(type) {
	case *ast.ArrayType:
		// **Nur Slices, keine Felder fester Länge.** Ein `[4]int` ist nie
		// `nil` und verschwindet nur, wenn es ganz aus Nullen besteht — das
		// ist eine andere Frage, und dieses Projekt hat keines.
		return gestalt.Len == nil
	case *ast.MapType:
		return true
	default:
		return false
	}
}

// hatAuslassung sagt, ob die Marke `omitempty` trägt.
func hatAuslassung(marke *ast.BasicLit) bool {
	return marke != nil && strings.Contains(marke.Value, "omitempty")
}

// listenfeldname gibt den Namen eines Feldes — oder seinen Typ, wenn es keinen hat.
func listenfeldname(feld *ast.Field) string {
	if len(feld.Names) > 0 {
		return feld.Names[0].Name
	}

	return "ein eingebettetes Feld"
}
