package app

// **Der Vertrag zwischen Fenster und Go, als Abdruck** — ENG-06.AK-04
// (#1731).
//
// # Warum es ihn braucht
//
// **Die Bindungen von Wails entstehen bei jedem Bau neu** und können nicht
// veralten. Was veralten kann, ist die Seite, die sie **liest**: die Fassade
// in `frontend/src/` ist von Hand geschrieben — mit Absicht, `api.ts` sagt
// warum —, und ihre Typen gehen durch keine Übersetzung. Ändert Go die Marke
// eines Feldes, übersetzt `tsc` weiter, und das Fenster liest `undefined`.
//
// **Der Abdruck macht das laut.** Er fasst je gebundener Methode zusammen,
// was über den Draht geht — Argumente, Antwort, und jede JSON-Marke darin —,
// und `frontend/src/bindungsvertrag.json` hält den Stand fest, den das
// Fenster zuletzt quittiert hat. Ein Wächter vergleicht beides und nennt die
// Methoden, die sich geändert haben.
//
// # Was er nicht hält
//
// **Ob die Typen im Fenster stimmen.** Er erzwingt, dass jemand hinsieht,
// nicht dass er richtig hinsieht; den Weg selbst geht der Nahtlauf
// (`scripts/nahtlauf.mjs`) an der laufenden Anwendung.

import (
	"crypto/sha256"
	"encoding"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"
)

// abdrucklaenge ist die Länge eines Abdrucks in Hexziffern.
//
// **Zwölf und nicht vierundsechzig**: der Abdruck unterscheidet Stände, er
// schützt vor keinem Angreifer — und ein Mensch vergleicht ihn mit dem Auge.
const abdrucklaenge = 12

// vertrag rechnet den Vertrag einmal je Prozess.
var vertrag = sync.OnceValue(func() map[string]string {
	typ := reflect.TypeFor[*Bindungen]()
	abdruecke := make(map[string]string, typ.NumMethod())

	for stelle := range typ.NumMethod() {
		methode := typ.Method(stelle)
		abdruecke[methode.Name] = abdruckVon(bindungssignatur(methode.Type))
	}

	return abdruecke
})

// Bindungsvertrag gibt je gebundener Methode den Abdruck ihrer Gestalt.
func Bindungsvertrag() map[string]string {
	kopie := make(map[string]string, len(vertrag()))
	for name, abdruck := range vertrag() {
		kopie[name] = abdruck
	}

	return kopie
}

// Vertragsabdruck ist der Abdruck über alle Methoden zugleich.
func Vertragsabdruck() string {
	zeilen := make([]string, 0, len(vertrag()))
	for name, abdruck := range vertrag() {
		zeilen = append(zeilen, name+" "+abdruck)
	}

	slices.Sort(zeilen)

	return abdruckVon(strings.Join(zeilen, "\n"))
}

// abdruckVon kürzt einen SHA-256 auf [abdrucklaenge].
func abdruckVon(text string) string {
	summe := sha256.Sum256([]byte(text))

	return hex.EncodeToString(summe[:])[:abdrucklaenge]
}

// bindungssignatur beschreibt Argumente und Antwort einer Methode — ohne den
// Empfänger, denn der geht nicht über den Draht.
func bindungssignatur(typ reflect.Type) string {
	teile := make([]string, 0, typ.NumIn()+typ.NumOut())

	for stelle := 1; stelle < typ.NumIn(); stelle++ {
		teile = append(teile, gestaltVon(typ.In(stelle), map[reflect.Type]bool{}))
	}

	teile = append(teile, "->")

	for stelle := range typ.NumOut() {
		teile = append(teile, gestaltVon(typ.Out(stelle), map[reflect.Type]bool{}))
	}

	return strings.Join(teile, " ")
}

// schreibtSichSelbst sind Typen mit eigener JSON-Gestalt — `time.Time` etwa.
var schreibtSichSelbst = []reflect.Type{
	reflect.TypeFor[json.Marshaler](), reflect.TypeFor[encoding.TextMarshaler](),
}

// gestaltVon beschreibt einen Typ so, wie JSON ihn über den Draht schickt.
//
// **Die Marke zählt, nicht der Feldname**: das Fenster liest `json:"…"`.
// Ein Verbund, der sich selbst enthält, steht beim zweiten Mal nur mit
// seinem Namen da — sonst liefe die Beschreibung nie zu Ende.
func gestaltVon(typ reflect.Type, gesehen map[reflect.Type]bool) string {
	for _, schnittstelle := range schreibtSichSelbst {
		if typ.Implements(schnittstelle) {
			return "json(" + typ.String() + ")"
		}
	}

	switch typ.Kind() {
	case reflect.Pointer:
		return "*" + gestaltVon(typ.Elem(), gesehen)
	case reflect.Slice:
		return "[]" + gestaltVon(typ.Elem(), gesehen)
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", typ.Len(), gestaltVon(typ.Elem(), gesehen))
	case reflect.Map:
		return "map[" + gestaltVon(typ.Key(), gesehen) + "]" + gestaltVon(typ.Elem(), gesehen)
	case reflect.Interface:
		return "any"
	case reflect.Struct:
		return verbundVon(typ, gesehen)
	default:
		return typ.Kind().String()
	}
}

// verbundVon beschreibt die Felder eines Verbunds, wie JSON sie ausgibt.
func verbundVon(typ reflect.Type, gesehen map[reflect.Type]bool) string {
	if gesehen[typ] {
		return typ.Name()
	}

	gesehen[typ] = true
	defer delete(gesehen, typ)

	return "{" + strings.Join(feldgestalten(typ, gesehen), ",") + "}"
}

// feldgestalten gibt je Feld `marke:gestalt` — eingebettete flach, wie JSON.
func feldgestalten(typ reflect.Type, gesehen map[reflect.Type]bool) []string {
	felder := []string{}

	for stelle := range typ.NumField() {
		feld := typ.Field(stelle)
		marke, _, _ := strings.Cut(feld.Tag.Get("json"), ",")

		switch {
		case !feld.IsExported() && !feld.Anonymous, marke == "-":
			continue
		case feld.Anonymous && marke == "" && feld.Type.Kind() == reflect.Struct:
			felder = append(felder, feldgestalten(feld.Type, gesehen)...)

			continue
		case marke == "":
			marke = feld.Name
		}

		if strings.Contains(feld.Tag.Get("json"), ",omitempty") {
			marke += "?"
		}

		felder = append(felder, marke+":"+gestaltVon(feld.Type, gesehen))
	}

	return felder
}
