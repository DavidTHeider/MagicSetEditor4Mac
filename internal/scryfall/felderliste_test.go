package scryfall_test

import (
	"reflect"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// **`GeleseneFelder` bleibt am Typ, nicht hinter ihm zurück**
// (#1005).
//
// # Warum die Liste überhaupt von Hand geführt ist
//
// Weil `Typzeile` in **zwei** Kartenfelder geht — `super type` und `sub type`.
// Eine Liste, die aus dem Typ entstünde, könnte das nicht ausdrücken, und die
// Zuordnung führt dafür zwei Zeilen.
//
// # Was dieser Wächter deshalb tut
//
// Er hält beide Richtungen: jedes Feld des DTO steht in der Liste, und jeder
// Name der Liste ist ein Feld. Zusammen mit
// `TestJedesGeleseneFeldStehtInDerZuordnung` heißt das: **ein neues Feld im
// DTO erzwingt eine Entscheidung** — Ziel oder Grund. Ohne diese Kette wäre
// ein gelesenes Feld unsichtbar.
func TestDieListeDerGelesenenFelderPasstZumTyp(t *testing.T) {
	t.Parallel()

	gemeldet := scryfall.GeleseneFelder()
	amTyp := feldnamen(reflect.TypeOf(scryfall.Karte{}))

	for _, name := range amTyp {
		if !slices.Contains(gemeldet, name) {
			t.Errorf("die Karte hat das Feld %q, und GeleseneFelder nennt es nicht — "+
				"dann erzwingt es auch keine Entscheidung in der Zuordnung", name)
		}
	}

	for _, name := range gemeldet {
		if !slices.Contains(amTyp, name) {
			t.Errorf("GeleseneFelder nennt %q — das Feld gibt es nicht mehr", name)
		}
	}
}

// feldnamen gibt die Namen der Felder eines Structs.
func feldnamen(art reflect.Type) []string {
	namen := make([]string, 0, art.NumField())
	for stelle := range art.NumField() {
		namen = append(namen, art.Field(stelle).Name)
	}

	return namen
}
