package app_test

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Der Vertragstest über leere Listen
// (#64).
//
// # Die Fehlerklasse, die er hält
//
// Sie liegt **zwischen** den Prüfebenen (AGENTS.md §3): Frontend-Tests gehen
// über `src/test/api.ts` und nie durch die JSON-Serialisierung, Go-Tests sehen
// die Struct-Marken nicht. Dazwischen passt eine leere Liste durch — auf zwei
// Wegen, und beide enden auf einer leeren Seite bei grüner Pipeline:
//
//   - `json:",omitempty"` auf einer Liste lässt sie **ganz verschwinden**. Im
//     Frontend kommt `undefined` an, und `antwort.sets.map(…)` bricht.
//   - Eine **nil**-Liste wird zu `null`, auch ohne `omitempty`. Im Frontend
//     kommt `null` an, und `antwort.sets.map(…)` bricht genauso.
//
// Der zweite Weg ist der leisere: er braucht keine Marke, sondern nur einen
// Rückgabepfad, der `nil` durchreicht — etwa einen Fehlerfall, der die
// vorbereitete leere Liste nicht mitnimmt.
//
// # Was er nicht prüft
//
// **Listen innerhalb von Listeneinträgen.** Er sieht die obersten Felder einer
// Antwort. Ein `[]string` in einem `SetDTO` ist erst zu sehen, wenn es ein
// `SetDTO` gibt — und dann ist die Antwort nicht mehr leer. Für den Fall, den
// dieses Issue meint (nichts da, Seite bleibt leer), liegt die Naht oben.
//
// **Andere Aufrufe als die hier genannten.** Er ist eine Liste, keine
// Reflexion über alle Bindungen: die meisten brauchen einen Bestand, und ein
// Aufruf mit Bestand prüft nicht mehr den leeren Fall.

// listenbefunde nennt jedes Listenfeld, das nicht als `[]` hinausgeht.
//
// Über die **Struktur** und nicht über einen Textvergleich auf `null`: ein
// `"fehler": null` ist richtig und soll nicht anschlagen. Gefragt ist, ob ein
// Feld, das im Frontend eine Liste sein muss, auch als Liste ankommt.
func listenbefunde(t *testing.T, antwort any) []string {
	t.Helper()

	roh, err := json.Marshal(antwort)
	if err != nil {
		t.Fatalf("serialisieren: %v", err)
	}

	var felder map[string]json.RawMessage
	if fehler := json.Unmarshal(roh, &felder); fehler != nil {
		t.Fatalf("zerlegen: %v", fehler)
	}

	var befunde []string

	art := reflect.TypeOf(antwort)

	for i := range art.NumField() {
		feld := art.Field(i)

		if feld.Type.Kind() != reflect.Slice {
			continue
		}

		name := jsonname(feld)

		wert, da := felder[name]
		if !da {
			befunde = append(befunde,
				name+" fehlt in der Antwort — steht dort ein omitempty?")

			continue
		}

		if string(wert) == "null" {
			befunde = append(befunde, name+" geht als null hinaus, nicht als []")
		}
	}

	return befunde
}

// jsonname liest den Namen aus der Marke, ohne die Zusätze dahinter.
func jsonname(feld reflect.StructField) string {
	marke := feld.Tag.Get("json")

	for i, zeichen := range marke {
		if zeichen == ',' {
			return marke[:i]
		}
	}

	if marke == "" {
		return feld.Name
	}

	return marke
}

// **Was nichts findet, gibt eine leere Liste zurück — keine fehlende.**
func TestEineLeereAntwortTraegtLeereListen(t *testing.T) {
	db := frischeAppdatenbank(t)

	leser := storage.NeuerSetLeser(db)

	anwendung := &app.App{}
	anwendung.MitBestand(leser)
	anwendung.MitSucher(leser)
	anwendung.MitPaketen(service.NeuerPaketverwalter(storage.NeuerPaketkatalog(db), t.TempDir()))
	anwendung.MitMSE2Orten(func() []string { return nil })

	faelle := []struct {
		name    string
		antwort any
	}{
		{"Sets", anwendung.Sets(t.Context())},
		{"Karten", anwendung.Karten(t.Context(), "1")},
		{"Vorbehalte", anwendung.Vorbehalte(t.Context(), "1")},
		{"Suchen", anwendung.Suchen(t.Context(), "nichts", "", "")},
		{"Suchfelder", anwendung.Suchfelder(t.Context(), "")},
		{"Pakete", anwendung.Pakete(t.Context())},
		{"Abhaengigkeiten", anwendung.Abhaengigkeiten(t.Context())},
		{"MSE2Vorschlaege", anwendung.MSE2Vorschlaege()},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			for _, befund := range listenbefunde(t, fall.antwort) {
				t.Error(befund)
			}
		})
	}
}

// Die Gegenprobe: der Vertragstest muss beide Wege sehen.
//
// Ohne sie wäre er ein Test, der grün ist, weil er nichts ansieht — genau die
// Sorte, die dieses Repository in Phase 5 viermal entfernt hat (`AGENTS.md` §10).
func TestDerListenvertragSchlaegtAn(t *testing.T) {
	fehltGanz := struct {
		Sets []string `json:"sets,omitempty"`
	}{}

	if befunde := listenbefunde(t, fehltGanz); len(befunde) != 1 {
		t.Errorf("omitempty auf einer leeren Liste geht durch: %v", befunde)
	}

	istNull := struct {
		Sets []string `json:"sets"`
	}{}

	if befunde := listenbefunde(t, istNull); len(befunde) != 1 {
		t.Errorf("eine nil-Liste geht durch: %v", befunde)
	}

	// **Und die andere Richtung.** Eine leere Liste ist richtig, und ein
	// `null` an einem Feld, das keine Liste ist, geht den Vertrag nichts an.
	inOrdnung := struct {
		Sets   []string `json:"sets"`
		Fehler *string  `json:"fehler"`
	}{Sets: []string{}}

	if befunde := listenbefunde(t, inOrdnung); len(befunde) != 0 {
		t.Errorf("eine richtige Antwort gilt als falsch: %v", befunde)
	}
}

// **Und die Klasse dazu, nicht nur die aufgezählten Aufrufe.**
//
// Der Test darüber geht eine Liste von Bindungen durch; er sieht nicht, was
// morgen dazukommt. Der zweite Weg in dieselbe Falle — `omitempty` auf einer
// Liste — lässt sich dagegen vollständig prüfen: er steht in der Marke und
// nicht im Rückgabepfad.
//
// # Was er nicht prüft
//
// **Die nil-Liste.** Die entsteht beim Bauen der Antwort und ist keiner Marke
// anzusehen — dafür steht der Vertragstest darüber.
//
// **Einen benannten Listentyp.** Er sieht `[]T` und `map[K]V`, wie sie
// dastehen. Ein `type Namen []string` mit `omitempty` am Feld ginge durch, denn
// im Syntaxbaum steht dort ein Bezeichner und kein Listentyp. Dafür bräuchte er
// `go/types` und damit einen Typprüflauf über das Paket — für einen Fall, den
// dieses Paket heute nicht kennt: jede Antwort schreibt ihre Listen aus.
// Gefunden vom Zweitblick über Phase 8; hier steht er, statt dass er
// weggerechnet wird.
func TestKeineListeInDerFassadeTraegtOmitempty(t *testing.T) {
	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("Verzeichnis lesen: %v", err)
	}

	gesehen := 0

	for _, eintrag := range eintraege {
		name := eintrag.Name()

		if eintrag.IsDir() || !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") {
			continue
		}

		gesehen++

		for _, befund := range omitemptyBefunde(t, name, nil) {
			t.Error(befund)
		}
	}

	// **Sonst wäre ein leerer Lauf grün.** Ein Verzeichnisfilter, der nichts
	// trifft, sieht aus wie ein Verzeichnis ohne Verstöße (Phase 1, Punkt 1).
	if gesehen == 0 {
		t.Fatal("keine einzige Quelldatei gelesen")
	}
}

// Die Gegenprobe.
func TestDerOmitemptyWaechterSchlaegtAn(t *testing.T) {
	faelle := []struct {
		name    string
		quelle  string
		erwartt int
	}{
		{"Liste mit omitempty", "package app\ntype A struct {\n" +
			"\tSets []string `json:\"sets,omitempty\"`\n}\n", 1},
		{"Abbildung mit omitempty", "package app\ntype A struct {\n" +
			"\tWerte map[string]int `json:\"werte,omitempty\"`\n}\n", 1},
		{"Liste ohne omitempty", "package app\ntype A struct {\n" +
			"\tSets []string `json:\"sets\"`\n}\n", 0},
		{"Zeiger mit omitempty bleibt erlaubt", "package app\ntype A struct {\n" +
			"\tFehler *Fehler `json:\"fehler,omitempty\"`\n}\n", 0},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			befunde := omitemptyBefunde(t, "probe.go", fall.quelle)

			if len(befunde) != fall.erwartt {
				t.Errorf("%d Befunde, erwartet waren %d: %v", len(befunde), fall.erwartt, befunde)
			}
		})
	}
}

// omitemptyBefunde nennt jedes Listen- oder Abbildungsfeld mit `omitempty`.
//
// Über den Syntaxbaum und nicht über ein Textmuster: `[]byte` in einem
// Kommentar, eine Marke über zwei Zeilen, ein Feld in einem eingebetteten
// Verbund — ein Muster übersieht das eine und meldet das andere.
func omitemptyBefunde(t *testing.T, name string, quelle any) []string {
	t.Helper()

	baum, err := parser.ParseFile(token.NewFileSet(), name, quelle, 0)
	if err != nil {
		t.Fatalf("%s zerteilen: %v", name, err)
	}

	var befunde []string

	ast.Inspect(baum, func(knoten ast.Node) bool {
		feld, ist := knoten.(*ast.Field)
		if !ist || feld.Tag == nil {
			return true
		}

		if !strings.Contains(feld.Tag.Value, "omitempty") {
			return true
		}

		switch feld.Type.(type) {
		case *ast.ArrayType, *ast.MapType:
			befunde = append(befunde,
				name+": "+feldname(feld)+" ist eine Liste mit omitempty — "+
					"leer verschwindet sie ganz, und im Fenster steht undefined")
		}

		return true
	})

	return befunde
}

// feldname gibt den ersten Namen eines Feldes, oder „(ohne Namen)".
func feldname(feld *ast.Field) string {
	if len(feld.Names) == 0 {
		return "(ohne Namen)"
	}

	return feld.Names[0].Name
}
