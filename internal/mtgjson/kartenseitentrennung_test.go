package mtgjson_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// **Eine einseitige Karte wird nicht getrennt** — Befund 1 des Zweitblicks
// über Phase 3.6
// (#1167).
//
// # Der Befund
//
// `TrenneKartenwerte` deutete **jedes** Feld mit dem Zusatz ` 2` als
// Rückseite. `magic.mse-game` führt aber 40 Felder mit einer Zahl, die kein
// Grundfeld haben — `loyalty cost 1` bis `loyalty cost 16` sind die Zeilen
// eines Planeswalkers, `additional credit 2` ist die zweite Zeile einer
// Danksagung. Ein einseitiger Planeswalker wurde damit beim Export in **zwei**
// MTG-JSON-Karten zerlegt, verbunden über `otherFaceIds` — still.
//
// Gemessen am 2026-08-27: `TrenneKartenwerte` gab für
// `{name, loyalty cost 2, additional credit 2}` ein `hinten` mit zwei Werten.
//
// # Die Regel, die jetzt gilt
//
// **Getrennt wird, wenn `name 2` einen Wert trägt.** Das ist die Marke, die
// [fuehreSeitenZusammen] selbst setzt: es legt beim Lesen für jede
// zusammengelegte Karte einen Namen an die zweite Stelle. Und MTG JSON
// verlangt für jedes Kartenobjekt einen Namen — eine Rückseite ohne ihn ließe
// sich ohnehin nicht schreiben.
func TestEineEinseitigeKarteWirdNichtGetrennt(t *testing.T) {
	t.Parallel()

	for name, werte := range map[string][]domain.Value{
		"ein Planeswalker mit zweiter Fähigkeit": {
			{Field: "name", Text: "Jace"},
			{Field: "loyalty cost 1", Text: "+1"},
			{Field: "loyalty cost 2", Text: "-2"},
		},
		"eine zweite Danksagungszeile": {
			{Field: "name", Text: "Probe"},
			{Field: "additional credit 2", Text: "Danke"},
		},
		"ein leerer zweiter Name": {
			{Field: "name", Text: "Probe"},
			{Field: "name 2", Text: "   "},
			{Field: "casting cost 2", Text: "{R}"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			vorn, hinten := mtgjson.TrenneKartenwerte(werte)

			if len(hinten) != 0 {
				t.Errorf("hinten kamen %v", hinten)
			}

			if len(vorn) != len(werte) {
				t.Errorf("vorn kamen %v — erwartet sind alle %d Werte", vorn, len(werte))
			}
		})
	}
}

// **Eine zweiseitige Karte wird getrennt** — die Gegenprobe.
//
// Ohne sie wäre auch ein `TrenneKartenwerte` grün, das **nie** trennt.
func TestEineZweiseitigeKarteWirdGetrennt(t *testing.T) {
	t.Parallel()

	vorn, hinten := mtgjson.TrenneKartenwerte([]domain.Value{
		{Field: "name", Text: "Vorn"},
		{Field: "casting cost", Text: "{1}{R}"},
		{Field: "name 2", Text: "Hinten"},
		{Field: "casting cost 2", Text: "{R}"},
	})

	if len(vorn) != 2 || vorn[0].Text != "Vorn" {
		t.Errorf("vorn kamen %v", vorn)
	}

	if len(hinten) != 2 {
		t.Fatalf("hinten kamen %v", hinten)
	}

	// **Der Zusatz fällt weg**: die Rückseite ist eine eigene Karte, und dort
	// heißt das Feld wieder `name`.
	nach := map[string]string{}
	for _, wert := range hinten {
		nach[wert.Field] = wert.Text
	}

	if nach["name"] != "Hinten" || nach["casting cost"] != "{R}" {
		t.Errorf("hinten kamen %v", nach)
	}
}

// **Ein zusammengesetzter Name ohne Vorderseite** — die dritte Gestalt.
//
// # Warum sie hier steht und nicht bei den Rundläufen
//
// Weil `TrenneKartenwerte` sie jetzt entscheidet: eine Karte **ohne** `name`,
// aber **mit** `name 2`, hat eine zweite Seite — und der zusammengesetzte Name
// ist dann der der Rückseite allein, ohne Trenner.
func TestEineKarteOhneVorderseitennamenWirdGetrennt(t *testing.T) {
	t.Parallel()

	vorn, hinten := mtgjson.TrenneKartenwerte([]domain.Value{
		{Field: "casting cost", Text: "R"},
		{Field: "name 2", Text: "Tear"},
	})

	if len(vorn) != 1 || vorn[0].Field != "casting cost" {
		t.Errorf("vorn kamen %v", vorn)
	}

	if len(hinten) != 1 || hinten[0].Text != "Tear" {
		t.Errorf("hinten kamen %v", hinten)
	}
}
