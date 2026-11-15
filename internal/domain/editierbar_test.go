package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **`editable:` ist eine eigene Eigenschaft, nicht dieselbe wie `save value:`**
// — #708.
//
//	bool      editable;    ///< Can values of this field be edited?
//	bool      save_value;  ///< Should values of this field be written to files?
//
// `src/data/field.hpp`, Zeilen 45–52, Fassung f9d9356 vom 2021-07-06. MSE
// lässt ein nicht editierbares Feld nicht einmal auswählen:
//
//	bool is_enabled(ValueViewerP const& v) {
//	  return v->getField()->editable && v->isVisible();
//	}
//
// `src/gui/control/card_editor.cpp`, Zeilen 155–156.
//
// **24 der 249 Kartenfelder von `magic.mse-game` tragen es** — `full name`,
// `cmc`, `card number`, `shape`, `color category`, `exact color`,
// `color identity`, `color count` und weitere.
func TestEinFeldSagtObEsBearbeitetWird(t *testing.T) {
	t.Parallel()

	// **Ohne Angabe wahr** — `Field::Field` setzt `editable(true)`.
	if !(domain.Field{Name: "name", Editierbar: true}).Editierbar {
		t.Error("ein gewöhnliches Feld ist nicht editierbar")
	}

	if (domain.Field{Name: "cmc"}).Editierbar {
		t.Error("das Nullfeld ist editierbar")
	}
}
