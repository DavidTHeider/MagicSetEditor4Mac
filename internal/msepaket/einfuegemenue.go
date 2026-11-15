package msepaket

// Das **Einfügemenü** einer Symbolschrift — `insert symbol menu:`
// (#564).
//
// Es steht im Paket und sagt, welche Symbole eine Oberfläche zum Einfügen
// anbietet und in welcher Ordnung:
//
//	insert symbol menu:
//		item: T
//		item:
//			type: line
//		item:
//			name: half
//			item: 1/2
//
// `magic-mana-large.mse-symbol-font/symbol-font`, ab Zeile 841.

// Menueart ist die Sorte eines Menüeintrags.
//
//	enum class Type { CODE, CUSTOM, LINE, SUBMENU };
//
// `src/data/symbol_font.hpp`, Zeilen 120–125, Fassung f9d9356 vom 2021-07-06.
type Menueart string

const (
	// MenueCode ist der Normalfall: [Menueeintrag.Name] **ist** der Code, der
	// eingefügt wird.
	MenueCode Menueart = "code"

	// MenueEigen ist `custom` — MSE öffnet einen Eingabedialog und fügt ein,
	// was jemand tippt. `magic-mana-large` benutzt das für `generic`, die
	// beliebige generische Manazahl.
	MenueEigen Menueart = "custom"

	// MenueTrenner ist `line` — eine Linie zwischen zwei Gruppen.
	MenueTrenner Menueart = "line"

	// MenueUntermenue trägt Kinder.
	//
	// **Es steht so gut wie nie im Paket.** MSE setzt die Art beim Lesen:
	//
	//	if (Handler::isReading && !items.empty()) type = Type::SUBMENU;
	//
	// `IMPLEMENT_REFLECTION(InsertSymbolMenu)`, Zeile 541. Ein Eintrag mit
	// Kindern **ist** ein Untermenü, gleich was `type:` daneben behauptet.
	MenueUntermenue Menueart = "submenu"
)

// Menueeintrag ist ein Eintrag des Einfügemenüs.
type Menueeintrag struct {
	Art Menueart

	// Name ist bei [MenueCode] der einzufügende Code, bei [MenueEigen] der
	// Schlüssel, unter dem Beschriftung und Rückfrage im `.mse-locale` stehen.
	Name string

	// Beschriftung ist `label:` — die Aufschrift.
	//
	// **Meist leer.** MSE holt sie dann aus dem Sprachpaket:
	//
	//	if (m.label.empty())
	//	  m.label.default_ = tr(*symbol_font_for_reading(), "menu_item", m.name, capitalize);
	//
	// `after_reading(InsertSymbolMenu&)`, Zeilen 545–549, Fassung f9d9356 vom
	// 2021-07-06. Solange die `.mse-locale`-Pakete ungelesen sind
	// (#561),
	// bleibt der großgeschriebene Name.
	Beschriftung string

	// Rueckfrage ist `prompt:` — der Satz über dem Eingabefeld eines
	// [MenueEigen].
	Rueckfrage string

	// Uebersetzungen ist `localized label:`, je Sprachkürzel eine Aufschrift.
	Uebersetzungen map[string]string

	Kinder []Menueeintrag
	Stelle Stelle
}

// einfuegemenue liest einen `insert symbol menu:`-Block.
//
// **`nil` ohne Block**, nicht die leere Liste: MSE fragt
// `if (insert_symbol_menu)`, bevor es irgendetwas tut
// (`SymbolFont::insertSymbolCode`, Zeilen 424–430, Fassung f9d9356 vom
// 2021-07-06) — eine Palette ohne Einträge sähe aus wie eine kaputte statt wie
// eine, die es nicht gibt.
func einfuegemenue(knoten *Knoten) []Menueeintrag {
	if knoten == nil {
		return nil
	}

	return menueeintraege(knoten)
}

// menueeintraege liest die `item:`-Kinder eines Knotens.
func menueeintraege(knoten *Knoten) []Menueeintrag {
	kinder := knoten.Alle("item")
	if len(kinder) == 0 {
		return nil
	}

	gelesen := make([]Menueeintrag, 0, len(kinder))
	for _, kind := range kinder {
		gelesen = append(gelesen, menueeintrag(kind))
	}

	return gelesen
}

// menueeintrag liest einen `item:`-Knoten.
//
// **Die Kurzform zuerst.**
//
//	REFLECT_IF_READING_SINGLE_VALUE_AND(items.empty()) { REFLECT_NAMELESS(name); }
//
// Zeilen 532–534: `item: T` — ein Wert ohne Block — setzt `name`, und die Art
// bleibt bei ihrer Vorgabe `CODE` (`InsertSymbolMenu::InsertSymbolMenu`,
// Zeilen 434–436).
func menueeintrag(knoten *Knoten) Menueeintrag {
	kinder := menueeintraege(knoten)

	eintrag := Menueeintrag{
		Art:            MenueCode,
		Name:           knoten.Wert,
		Beschriftung:   wertOderLeer(knoten.Ersten("label")),
		Rueckfrage:     wertOderLeer(knoten.Ersten("prompt")),
		Uebersetzungen: aufschriften(knoten.Ersten("localized label")),
		Kinder:         kinder,
		Stelle:         knoten.Stelle,
	}

	if name := knoten.Ersten("name"); name != nil {
		eintrag.Name = name.Wert
	}

	switch {
	// **Kinder schlagen `type:`** — genau MSEs Reihenfolge: die Art wird
	// gelesen und danach überschrieben, wenn Einträge dastehen.
	case len(kinder) > 0:
		eintrag.Art = MenueUntermenue
	case knoten.Ersten("type") != nil:
		eintrag.Art = Menueart(wertOderLeer(knoten.Ersten("type")))
	}

	return eintrag
}
