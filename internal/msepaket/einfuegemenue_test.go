package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Eine Symbolschrift bringt ihr Einfügemenü mit**
// (#564).
//
//	insert symbol menu:
//		item: T
//		item:
//			type: line
//		item:
//			type: custom
//			name: generic
//		item:
//			name: half
//			item: 1/2
//
// `magic-mana-large.mse-symbol-font/symbol-font`, ab Zeile 841.
//
// # Die vier Arten
//
//	enum class Type { CODE, CUSTOM, LINE, SUBMENU };
//
// `src/data/symbol_font.hpp`, Zeilen 120–125, Fassung f9d9356 vom 2021-07-06.
//
// **`submenu` steht nirgends im Paket** — MSE setzt sie beim Lesen:
//
//	if (Handler::isReading && !items.empty()) type = Type::SUBMENU;
//
// `IMPLEMENT_REFLECTION(InsertSymbolMenu)`, Zeile 541. Ein Eintrag mit Kindern
// **ist** ein Untermenü, gleich was `type:` sagt.
//
// # Und die Kurzform
//
//	REFLECT_IF_READING_SINGLE_VALUE_AND(items.empty()) { REFLECT_NAMELESS(name); }
//
// Zeilen 532–534: `item: T` — ein Wert ohne Block — setzt `name`, und `type`
// bleibt auf seiner Vorgabe `CODE` (`InsertSymbolMenu::InsertSymbolMenu`,
// Zeilen 434–436).
func TestEineSymbolschriftNenntIhrEinfuegemenue(t *testing.T) {
	schrift, _ := oeffneTestpaket(t, "probe-symbole.mse-symbol-font").Symbolmodell()

	menue := schrift.Einfuegemenue
	if len(menue) != 5 {
		t.Fatalf("%d Einträge, erwartet 5: %+v", len(menue), menue)
	}

	faelle := []struct {
		art    msepaket.Menueart
		name   string
		kinder int
	}{
		{msepaket.MenueCode, "W", 0},
		{msepaket.MenueCode, "U", 0},
		{msepaket.MenueTrenner, "", 0},
		{msepaket.MenueEigen, "generic", 0},
		{msepaket.MenueUntermenue, "halb", 2},
	}

	for stelle, fall := range faelle {
		eintrag := menue[stelle]
		if eintrag.Art != fall.art || eintrag.Name != fall.name ||
			len(eintrag.Kinder) != fall.kinder {
			t.Errorf("Eintrag %d ist %+v", stelle, eintrag)
		}
	}

	// **Ein Untermenü hat seine Einträge, und die sind wieder Codes.**
	if kind := menue[4].Kinder[0]; kind.Art != msepaket.MenueCode || kind.Name != "1/2" {
		t.Errorf("das erste Kind ist %+v", kind)
	}
}

// **Eine Schrift ohne `insert symbol menu:` hat keins.**
//
// `nil` und keine leere Liste: MSE fragt `if (insert_symbol_menu)` vor allem
// anderen (`SymbolFont::insertSymbolCode`, Zeilen 424–430, Fassung f9d9356 vom
// 2021-07-06), und eine Palette ohne Einträge sähe aus wie eine kaputte statt
// wie eine, die es nicht gibt.
func TestEineSchriftOhneMenueHatKeins(t *testing.T) {
	schrift, _ := oeffneTestpaket(t, "probe-ohnemenue.mse-symbol-font").Symbolmodell()

	if schrift.Einfuegemenue != nil {
		t.Errorf("ohne Angabe kam %+v", schrift.Einfuegemenue)
	}
}
