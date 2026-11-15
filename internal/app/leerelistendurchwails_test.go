// **Eine leere Liste kommt als `[]` am Fenster an, nicht als `null`** —
// #1169,
// Werkstück 3.
//
// # Die Naht, vor der §3 warnt — und was sie an der zweiten Hälfte kostet
//
// Der Wächter in `internal/werkzeuge/leerelisten_test.go` hält die eine
// Hälfte: kein `omitempty` auf einer Liste. Die andere ist ein `nil`-Slice,
// das ohne jede Marke zu `null` wird — dagegen hilft nur `make([]T, 0)` an
// **jeder** Stelle, die die Antwort baut, und das sieht kein Muster.
//
// **Hier wird es gefahren.** Über einen **leeren** Bestand: dann ist jede
// Liste leer, und genau dann entscheidet sich, ob sie als `[]` oder als `null`
// ankommt. Marshalliert wird mit `encoding/json` — dasselbe, was Wails tut.
//
// # Was dieser Lauf nicht sieht
//
// **Bindungen mit Argumenten, die es hier nicht gibt.** Er fährt die
// Auskünfte, die ohne Vorbedingung antworten; eine Kartenliste braucht ein
// Set, und ein Set braucht eine Datei. Diese Wege haben ihre eigenen Läufe
// (`bestand_test.go`, `ersetzen_test.go`, `bildausfuhr_test.go`).
//
// **Und Wails selbst.** Es erzeugt seine Aufrufe zur Laufzeit; ein Lauf ohne
// Fenster erreicht sie nicht. Was hier gefahren wird, ist dieselbe
// Serialisierung — nicht dieselbe Brücke.
package app_test

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestKeineLeereListeKommtAlsNullAn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "leer.db"))
	bindungen := app.NeueBindungen(anwendung)

	// **Die Auskünfte ohne Vorbedingung.** Jede antwortet über einem leeren
	// Bestand, und jede trägt mindestens eine Liste.
	for name, hole := range map[string]func() any{
		"Sets":          func() any { return bindungen.Sets(t.Context()) },
		"Beispielsets":  func() any { return bindungen.Beispielsets(t.Context()) },
		"Pakete":        func() any { return bindungen.Pakete(t.Context()) },
		"Einstellungen": func() any { return bindungen.Einstellungen(t.Context()) },
		"Sammlung":      func() any { return bindungen.Sammlung(t.Context()) },
		"Portfolio":     func() any { return bindungen.Portfolio(t.Context()) },
	} {
		t.Run(name, func(t *testing.T) {
			antwort := hole()

			roh, err := json.Marshal(antwort)
			if err != nil {
				t.Fatalf("marshallieren: %v", err)
			}

			for _, feld := range nullListen(t, antwort, roh) {
				t.Errorf("%s kommt als null an — das Frontend läuft darüber, "+
					"ohne zu prüfen, und die Seite bleibt leer (AGENTS.md §3):\n%s",
					feld, roh)
			}
		})
	}
}

// nullListen gibt die Listenfelder, die **fehlen** oder `null` sind.
//
// # Beides, und das zweite ist der Zusatz aus Befund 5
//
// Der erste Anlauf ging über die Schlüssel der Antwort und sah damit nur, was
// **dasteht**. Ein Feld, das `omitempty` hat und leer ist, steht gar nicht da
// — und genau das ist der Fall, gegen den dieser Lauf gebaut ist. Er wäre
// blind für ihn gewesen.
//
// **Gefragt wird deshalb den Typ**, nicht die Antwort: jedes Feld, dessen Go-
// Typ eine Liste ist, muss im JSON stehen und darf nicht `null` sein.
//
// **Der Fehler ist keine Liste** und bleibt außen vor: `fehler: null` ist die
// richtige Antwort auf „es ging gut".
func nullListen(t *testing.T, wert any, roh []byte) []string {
	t.Helper()

	var gelesen map[string]any
	if err := json.Unmarshal(roh, &gelesen); err != nil {
		t.Fatalf("kein Verbund: %v\n%s", err, roh)
	}

	gestalt := reflect.TypeOf(wert)
	if gestalt.Kind() != reflect.Struct {
		t.Fatalf("die Antwort ist kein Verbund, sondern %s", gestalt.Kind())
	}

	var befunde []string

	for stelle := range gestalt.NumField() {
		feld := gestalt.Field(stelle)

		if feld.Type.Kind() != reflect.Slice && feld.Type.Kind() != reflect.Map {
			continue
		}

		name := listenjsonname(feld)
		if name == "" || name == "-" {
			continue
		}

		inhalt, steht := gelesen[name]

		switch {
		case !steht:
			befunde = append(befunde, name+" (fehlt ganz)")
		case inhalt == nil:
			befunde = append(befunde, name)
		}
	}

	return befunde
}

// listenjsonname gibt den Namen, unter dem ein Feld im JSON steht.
func listenjsonname(feld reflect.StructField) string {
	marke, hat := feld.Tag.Lookup("json")
	if !hat {
		return feld.Name
	}

	name, _, _ := strings.Cut(marke, ",")
	if name == "" {
		return feld.Name
	}

	return name
}
