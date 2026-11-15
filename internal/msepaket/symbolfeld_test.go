package msepaket_test

// **Die Feldart `symbol`** — Werkstück A von Phase 4.11
// (#1339,
// #1349).
//
// # Was am 2026-08-31 nachgelesen wurde (§15)
//
// `src/data/field/symbol.cpp` (Fassung f9d9356 vom 2021-07-06):
//
//	IMPLEMENT_FIELD_TYPE(Symbol, "symbol");
//
// Das Feld selbst trägt **keine eigenen Angaben** — im Quelltext steht
// wörtlich „no extra data". Was es ausmacht, steht am Wert
// (`src/data/field/symbol.hpp`):
//
//	ValueType filename;    ///< Filename of the symbol (in the current package)
//	Age       last_update; ///< When was the symbol last changed?
//
// **Der Wert ist ein Dateiname im Paket**, keine eingebettete Zeichnung.
//
// # Und was gemessen wurde
//
// In `full-magic-pack` (526 Pakete, 2026-08-31) steht `type: symbol`
// **neunmal in sieben Dateien** — und **jedes Mal als `set field:`**, kein
// einziges Mal als `card field:`. Die Feldart geht bei MSE durch dieselbe
// Abbildung (`artAus` liest beide Blöcke), also trifft eine Ergänzung beide:
// das Kartenfeld, das #1339 verlangt, **und** die sechs Spielpakete, deren
// Setsymbolfeld hier bisher als `KindUnknown` durchfiel.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **`type: symbol` wird als Symbolfeld gelesen — an der Karte und am Set.**
func TestDieFeldartSymbolWirdGelesen(t *testing.T) {
	t.Parallel()

	paket := spielMitBlock(t, `card field:
	type: symbol
	name: kartensymbol
set field:
	type: symbol
	name: symbol
`)

	spiel, befunde := paket.Spielmodell()
	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(spiel.Fields) != 1 || spiel.Fields[0].Kind != domain.KindSymbol {
		t.Errorf("Kartenfelder: %+v", spiel.Fields)
	}

	if len(spiel.SetFields) != 1 || spiel.SetFields[0].Kind != domain.KindSymbol {
		t.Errorf("Setfelder: %+v", spiel.SetFields)
	}
}
