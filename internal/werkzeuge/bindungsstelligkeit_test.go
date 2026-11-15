// **Die Stelligkeit einer Bindung, auf beiden Seiten dieselbe** — Befund
// **F03** der nativen Prüfung
// (#1747).
//
// # Der Befund
//
// Die ausgelieferte Anwendung 0.5.0 antwortete auf „CSV einlesen" mit
//
//	error parsing arguments: received 1 arguments to
//	method app.Bindungen.SammlungCSVEinlesen, expected 2
//
// `internal/app` erklärt `SammlungCSVEinlesen(text, quelle string)`, der
// Vertrag in `frontend/src/apisammlung.ts` erklärte `(text: string)`, und der
// Aufrufer hielt sich an den Vertrag. **`tsc` konnte nichts merken**: Vertrag
// und Aufrufer stimmten überein, gelogen hat der Vertrag über Go.
//
// # Warum es das vorher nicht gab
//
// Zwei Wächter halten die Bindungen zusammen, und beide halten den **Namen**:
// `bindungszaehler_test.go` zählt sie, `bindungsaufrufer_test.go` verlangt für
// jede einen Aufrufer. Die **Stelligkeit** hielt keiner. Das ist §10s Satz vom
// Wächter, der die halbe Zusage hält und die andere unsichtbar macht: solange
// der Name stimmt, sieht die Naht heil aus.
//
// # Die Reichweite dieser Messung
//
// **Gemessen am 2026-09-20** über alle 422 exportierten Bindungen: genau
// **eine** Abweichung, und das ist F03. Der Wächter hält also ab dem zweiten
// Lauf einen Zustand, der stimmt — das ist der günstigste Zeitpunkt, ihn
// aufzuschreiben (dieselbe Überlegung wie in `bindungsaufrufer_test.go`).
//
// Dabei fiel die **Gegenrichtung** mit auf: `SammlungFuer` stand im Vertrag,
// ohne dass es die Bindung gab — keine `func (b *Bindungen) SammlungFuer`,
// kein Eintrag in `fassade.ts`, kein Aufrufer, kein Eintrag in
// `FAEHIGKEITEN.md`. Eine Zusage, die niemand einlöst; sie ist entfernt.
//
// # Was er hält
//
//   - Jede exportierte Bindung steht im Vertrag, und zwar mit **derselben
//     Zahl an Parametern**.
//   - Jeder Vertragseintrag hat eine Bindung.
//
// # Was er **nicht** hält
//
//   - **Nicht die Typen und nicht ihre Reihenfolge.** `(text, quelle string)`
//     gegen `(quelle: string, text: string)` käme hier durch. Was das hält,
//     ist der Brückenlauf über die laufende Anwendung — und am Ende die
//     Sichtprüfung.
//   - **Nicht, dass der Aufrufer richtig ruft.** Das tut `tsc`, sobald der
//     Vertrag die Wahrheit sagt — und genau diese Voraussetzung stellt dieser
//     Wächter her.
//   - **Nicht die unexportierten Helfer.** `waehlenUndOeffnen` und die zwei
//     Dienstmethoden sind Wege des Fensters, keine Bindungen
//     (#1287).
package werkzeuge

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// vertragszeile findet den Beginn eines Vertragseintrags.
//
// **Zwei Leerzeichen und ein großer Anfangsbuchstabe.** Die Einträge stehen
// als Glieder einer `interface`-Erklärung, also genau eine Einrückung tief;
// eine Bindung heißt in Go exportiert und damit groß. Beides zusammen trennt
// sie von allem anderen, was in diesen Dateien `name: (` schreibt.
var vertragszeile = regexp.MustCompile(`(?m)^ {2}([A-Z]\w*): \(`)

// TestJedeBindungHatImFensterDieselbeStelligkeit hält F03s Klasse fern.
func TestJedeBindungHatImFensterDieselbeStelligkeit(t *testing.T) {
	t.Parallel()

	inGo := bindungsstelligkeiten(t)
	if len(inGo) < 100 {
		t.Fatalf("nur %d Bindungen gefunden — der Leser greift daneben", len(inGo))
	}

	imVertrag := vertragsstelligkeiten(t)
	if len(imVertrag) < 100 {
		t.Fatalf("nur %d Vertragseinträge gefunden — der Leser greift daneben",
			len(imVertrag))
	}

	for name, stellen := range inGo {
		erklaert, steht := imVertrag[name]
		if !steht {
			t.Errorf("die Bindung %q steht in keiner `api…`-Datei — das Fenster "+
				"kennt sie nicht, und niemand merkt es", name)

			continue
		}

		if erklaert != stellen {
			t.Errorf("%s: Go nimmt %d Parameter, der Vertrag erklärt %d — "+
				"das ist F03s Gestalt, und `tsc` sieht sie nicht, weil Vertrag "+
				"und Aufrufer zueinander passen", name, stellen, erklaert)
		}
	}

	for name := range imVertrag {
		if _, gibtEs := inGo[name]; !gibtEs {
			t.Errorf("der Vertrag erklärt %q, es gibt aber keine "+
				"`func (b *Bindungen) %s` — eine Zusage ohne Einlösung (§2)",
				name, name)
		}
	}
}

// bindungsstelligkeiten liest Name und Parameterzahl aus `internal/app`.
//
// **Über den Syntaxbaum und nicht über ein Muster** — aus demselben Grund wie
// in `bindungsaufrufer_test.go`: 94 der Signaturen gehen über mehrere Zeilen,
// und ein Muster über `func (b *Bindungen) …(…)` liest davon keine einzige.
func bindungsstelligkeiten(t *testing.T) map[string]int {
	t.Helper()

	gefunden := map[string]int{}
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

			sammleStelligkeiten(baum, gefunden)

			return nil
		})
	if fehler != nil {
		t.Fatalf("internal/app durchgehen: %v", fehler)
	}

	return gefunden
}

// sammleStelligkeiten trägt jede exportierte Methode auf `*Bindungen` ein.
func sammleStelligkeiten(baum *ast.File, hinein map[string]int) {
	for _, erklaerung := range baum.Decls {
		funktion, ist := erklaerung.(*ast.FuncDecl)
		if !ist || funktion.Recv == nil || len(funktion.Recv.List) != 1 ||
			!funktion.Name.IsExported() {
			continue
		}

		stern, amZeiger := funktion.Recv.List[0].Type.(*ast.StarExpr)
		if !amZeiger {
			continue
		}

		name, benannt := stern.X.(*ast.Ident)
		if !benannt || name.Name != "Bindungen" {
			continue
		}

		// **Den Rufkontext reicht Wails selbst** (BK-01, #2109): er steht
		// in Go vorn und im Vertrag nicht.
		hinein[funktion.Name.Name] = parameterzahl(funktion.Type.Params) -
			rufstelle(funktion.Type.Params)
	}
}

// rufstelle ist 1, wenn der erste Parameter ein `context.Context` ist — den
// füllt Wails (`needsContext`, `bindings.go` in v3.0.0-beta.23), und das
// Fenster zählt ihn nicht mit.
func rufstelle(liste *ast.FieldList) int {
	if liste == nil || len(liste.List) == 0 {
		return 0
	}

	auswahl, ist := liste.List[0].Type.(*ast.SelectorExpr)
	if !ist || auswahl.Sel.Name != "Context" {
		return 0
	}

	if paket, benannt := auswahl.X.(*ast.Ident); !benannt || paket.Name != "context" {
		return 0
	}

	return 1
}

// parameterzahl zählt die Parameter einer Signatur.
//
// **Eine Gruppe ist nicht ein Parameter.** `(text, quelle string)` ist **ein**
// `ast.Field` mit zwei Namen und zwei Parametern — wer Felder zählte, käme auf
// eins und übersähe genau den Fall, der F03 war.
func parameterzahl(liste *ast.FieldList) int {
	if liste == nil {
		return 0
	}

	gezaehlt := 0

	for _, feld := range liste.List {
		if len(feld.Names) == 0 {
			// Ein unbenannter Parameter ist trotzdem einer.
			gezaehlt++

			continue
		}

		gezaehlt += len(feld.Names)
	}

	return gezaehlt
}

// vertragsstelligkeiten liest Name und Parameterzahl aus `frontend/src/api*.ts`.
func vertragsstelligkeiten(t *testing.T) map[string]int {
	t.Helper()

	gefunden := map[string]int{}
	ordner := filepath.Join(wurzel, "frontend", "src")

	eintraege, fehler := os.ReadDir(ordner)
	if fehler != nil {
		t.Fatalf("%s lesen: %v", ordner, fehler)
	}

	for _, eintrag := range eintraege {
		name := eintrag.Name()
		if eintrag.IsDir() || !strings.HasPrefix(name, "api") ||
			!strings.HasSuffix(name, ".ts") || strings.Contains(name, ".test.") {
			continue
		}

		roh, lesefehler := os.ReadFile(filepath.Join(ordner, name)) // #nosec G304 -- fester Ordner
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", name, lesefehler)
		}

		leseVertrag(ohneZeilenkommentare(string(roh)), gefunden)
	}

	return gefunden
}

// leseVertrag trägt jeden Eintrag der Form `  Name: (…) => …` ein.
func leseVertrag(text string, hinein map[string]int) {
	for _, stelle := range vertragszeile.FindAllStringSubmatchIndex(text, -1) {
		name := text[stelle[2]:stelle[3]]

		// Die öffnende Klammer ist das letzte Zeichen des Treffers.
		auf := stelle[1] - 1

		zu := passendeKlammer(text, auf)
		if zu < 0 {
			continue
		}

		hinein[name] = teileZaehlen(text[auf+1 : zu])
	}
}

// passendeKlammer gibt die Stelle der schließenden Klammer zu `auf`.
func passendeKlammer(text string, auf int) int {
	tiefe := 0

	for stelle := auf; stelle < len(text); stelle++ {
		switch text[stelle] {
		case '(':
			tiefe++
		case ')':
			tiefe--

			if tiefe == 0 {
				return stelle
			}
		}
	}

	return -1
}

// teileZaehlen zählt die Parameter zwischen zwei Klammern.
//
// # Drei Dinge, an denen ein naiver Zähler scheitert
//
//   - **Das nachgestellte Komma.** Prettier schreibt es bei mehrzeiligen
//     Listen; wer auf Kommas teilt, zählt einen Parameter zu viel. Das war der
//     erste Anlauf, und er meldete 157 Abweichungen statt einer.
//   - **Generika mit Komma.** `Record<string, string>` ist **ein** Parameter.
//     Drei Einträge haben die Gestalt — ohne `<`/`>` als Tiefe wären sie drei
//     falsche Befunde.
//   - **Der Pfeil.** `=>` trägt ein `>`, das keine Klammer schließt. Ohne diese
//     Unterscheidung nähme die Tiefe ins Negative ab, und jedes folgende Komma
//     zählte mit.
func teileZaehlen(inneres string) int {
	tiefe, gezaehlt, offen := 0, 0, false

	var vorher byte

	for stelle := range len(inneres) {
		zeichen := inneres[stelle]

		switch {
		case zeichen == '(' || zeichen == '[' || zeichen == '{' || zeichen == '<':
			tiefe++
		case zeichen == ')' || zeichen == ']' || zeichen == '}':
			tiefe--
		case zeichen == '>' && vorher != '=':
			tiefe--
		case zeichen == ',' && tiefe == 0:
			if offen {
				gezaehlt++
			}

			offen = false
		}

		if zeichen != ',' && !istLeerraum(zeichen) {
			offen = true
		}

		if !istLeerraum(zeichen) {
			vorher = zeichen
		}
	}

	if offen {
		gezaehlt++
	}

	return gezaehlt
}

// ohneZeilenkommentare nimmt die `//`-Zeilen aus einer TypeScript-Datei.
//
// **Sonst deckt eine Erwähnung den Vertrag.** Die `api…`-Dateien erklären in
// Kommentaren, was eine Bindung tut, und nennen dabei ihren Namen samt
// Klammern — `SammlungCSVEinlesen(text)` in einem Satz Prosa wäre für den
// Leser oben ein Eintrag.
//
// **Nur ganze Zeilen.** Blockkommentare schreibt dieses Repositorium nicht,
// und ein `//` mitten in einer Zeile steht hier nur in einer Adresse innerhalb
// eines Kommentars — also in einer Zeile, die ohnehin fällt.
func ohneZeilenkommentare(quelltext string) string {
	zeilen := strings.Split(quelltext, "\n")
	behalten := make([]string, 0, len(zeilen))

	for _, zeile := range zeilen {
		if strings.HasPrefix(strings.TrimSpace(zeile), "//") {
			continue
		}

		behalten = append(behalten, zeile)
	}

	return strings.Join(behalten, "\n")
}

// istLeerraum sagt, ob ein Zeichen zwischen den Wörtern steht.
func istLeerraum(zeichen byte) bool {
	return zeichen == ' ' || zeichen == '\t' || zeichen == '\n' || zeichen == '\r'
}
