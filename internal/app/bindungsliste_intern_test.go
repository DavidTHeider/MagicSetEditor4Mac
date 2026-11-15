// Der Wächter über die **Liste** der Bindungen — Go gegen den quittierten
// Vertrag, und die Fassade liest denselben.
//
// **Eine eigene Datei**, seit `bindungen_intern_test.go` mit den beiden
// Auszählungsbindungen über die Zeilengrenze aus §2 ging
// (#1338).
// Der Schnitt folgt der Frage: dort steht, **wie** eine Bindung aussehen muss
// — kein Kontext, `main` bindet sie —, hier, **welche** es gibt.
package app

import (
	"encoding/json"
	"maps"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// vertragsnamen liest die Namen aus `frontend/src/bindungsvertrag.json`.
//
// # Die eine Liste
//
// **Hier stand `gebundeneMethoden` von Hand** — und daneben führten der
// Vertrag und fünf Dateien der Fassade dieselben 428 Namen: drei Listen für
// eine Tatsache, und jede neue Bindung musste in alle drei
// (#1825).
//
// **Die Zusage bleibt von Hand geführt**, und das war der Grund für die Liste:
// eine aus dem Typ abgeleitete prüfte sich selbst. Der Vertrag wird nicht
// abgeleitet, sondern **quittiert** — `internal/werkzeuge/bindungsvertrag_test.go`
// druckt die Vorlage, eingetragen wird sie von Hand, und wer eine Methode
// ergänzt, sieht dabei, dass er die Oberfläche vergrößert. Sie steht nur noch
// einmal.
func vertragsnamen(t *testing.T) []string {
	t.Helper()

	roh, err := os.ReadFile("../../frontend/src/bindungsvertrag.json")
	if err != nil {
		t.Fatalf("bindungsvertrag.json nicht lesbar: %v", err)
	}

	var vertrag struct {
		Methoden map[string]string `json:"methoden"`
	}

	if fehler := json.Unmarshal(roh, &vertrag); fehler != nil {
		t.Fatalf("bindungsvertrag.json nicht lesbar: %v", fehler)
	}

	// **Fail-closed**: ein leerer Vertrag sähe aus wie eine erfüllte Zusage.
	if len(vertrag.Methoden) == 0 {
		t.Fatal("der Vertrag nennt keine Methode — der Wächter griffe ins Leere")
	}

	return slices.Sorted(maps.Keys(vertrag.Methoden))
}

// **Die gebundene Oberfläche ist genau die quittierte.**
//
// Wails bindet jede exportierte Methode. Ohne diesen Wächter wandert eine neue
// Naht — ein `MitIrgendwas` — ins Fenster, sobald jemand sie schreibt.
//
// **Über den Typ zur Laufzeit** und nicht über den Quelltext wie der Wächter
// in `internal/werkzeuge`: `reflect` sieht, was Wails sieht, auch eine
// Methode, die ein eingebetteter Typ mitbringt.
func TestDieGebundeneOberflaecheIstDieAngemeldete(t *testing.T) {
	typ := reflect.TypeOf(&Bindungen{})

	gefunden := make([]string, 0, typ.NumMethod())

	for i := range typ.NumMethod() {
		gefunden = append(gefunden, typ.Method(i).Name)
	}

	slices.Sort(gefunden)

	// **Den Unterschied nennen und nicht zwei Listen zu je 428 Namen** — ein
	// roter Lauf soll sagen, welcher Name fehlt, statt ihn suchen zu lassen.
	angemeldet := vertragsnamen(t)

	for _, name := range gefunden {
		if !slices.Contains(angemeldet, name) {
			t.Errorf("%s ist gebunden, aber nicht quittiert", name)
		}
	}

	for _, name := range angemeldet {
		if !slices.Contains(gefunden, name) {
			t.Errorf("%s ist quittiert, aber nicht gebunden", name)
		}
	}
}

// **Und die Fassade im Frontend liest denselben Vertrag.**
//
// `frontend/src/fassade.ts` ist die Stelle, an der ein Platzhalter entsteht,
// wenn Wails eine Methode nicht kennt. **Dass jeder Name dort ankommt, hält
// `tsc`** — in beide Richtungen gegen `Api` (siehe dort). Hier steht, was der
// Übersetzer nicht sieht: dass die Fassade den Vertrag überhaupt liest, und
// unter welchem Pfad Wails die Bindungen ablegt.
func TestDieFassadeImFrontendLiestDenVertrag(t *testing.T) {
	quelle, err := os.ReadFile("../../frontend/src/fassade.ts")
	if err != nil {
		t.Fatalf("fassade.ts nicht lesbar: %v", err)
	}

	inhalt := ohneKommentare(string(quelle))

	if !strings.Contains(inhalt, "from './bindungsvertrag.json'") {
		t.Error("fassade.ts liest den Bindungsvertrag nicht — woher kämen ihre Namen?")
	}

	// **Und den Pfad, unter dem Wails sie ablegt.**
	//
	// Wails legt die gebundene Struktur unter ihrem **Typnamen** ab. Der erste
	// Anlauf dieser Fassade hieß `App`; als daraus `Bindungen` wurde, suchte
	// das Frontend weiter unter `go.app.App` — fand nichts und legte klaglos
	// seine Platzhalter darüber. Kein Test sah es, weil die Vorrichtung selbst
	// anlegt, was die Fassade sucht.
	pfad := "go?.app?." + reflect.TypeOf(Bindungen{}).Name()
	if !strings.Contains(inhalt, pfad) {
		t.Errorf("fassade.ts greift nicht auf %q zu — Wails legt die Bindungen dort ab", pfad)
	}
}
