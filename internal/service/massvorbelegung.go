package service

// Womit MSE eine Maßangabe vorbelegt — und was ein Skript dieser Angabe als
// `value` sieht
// (#555).
//
// # Warum es diese Datei gibt
//
// MSE rechnet jede Angabe eines Stils über `Scriptable<T>::update`:
//
//	inline bool update(Context& ctx) {
//	  return script.invokeOn(ctx, value);
//	}
//
//	template <typename T> bool invokeOn(Context& ctx, T& value) const {
//	  if (script) {
//	    T new_value;
//	    ctx.setVariable(SCRIPT_VAR_value, to_script(value));
//
// `src/script/scriptable.hpp`, Zeilen 134–136 und 58–66, Fassung f9d9356 vom
// 2021-07-06. **`value` ist der bisherige Wert dieser Angabe** — nicht der Wert
// des Feldes, obwohl der Name das nahelegt und die eine Fundstelle im Pack
// genau das gemeint zu haben scheint (siehe [Kantenvorgabe]).
//
// Der bisherige Wert ist beim ersten Durchlauf der aus dem Konstruktor. Dieses
// Programm rechnet einen Plan **einmal**; damit ist der Konstruktorwert der
// einzige, den es gibt, und diese Datei führt ihn je Angabe auf.
//
// # Was das für spätere Durchläufe heißt
//
// MSE ruft `Style::update` erneut, sobald sich eine Abhängigkeit ändert, und
// dann steht in `value` das Ergebnis des vorigen Laufs. Ein Skript, das aus
// seinem eigenen Vorwert einen neuen rechnet, liefe hier deshalb anders. **Im
// Pack gibt es keines**: über alle 376 Stilpakete steht `value` als freier Name
// in genau einer Maßangabe, und die vergleicht ihn gegen ein Textmuster, das
// keine Zahl trifft — in MSE wie hier.

// Die Vorbelegungen, jede mit ihrer Fundstelle.
const (
	// Kantenvorgabe gilt für `left`, `top`, `right` und `bottom`.
	//
	//	left (1000000), top   (1000000)
	//	right(1000000), bottom(1000000)
	//
	// `src/data/field.cpp`, Zeilen 101–103. **Die Million ist zugleich MSEs
	// Zeichen für „nicht gesetzt"** — `is_set` fragt `x < 100000` (Zeile
	// 132). Null wäre nicht bloß ein anderer Startwert, sondern eine andere
	// Aussage.
	//
	// **Sie ist die einzige, die im Pack je gelesen wird.**
	// `magic-m15-mainframe-dfc.mse-style/style`, Zeile 4059:
	//
	//	top: {if match(value, match:"^<sym>[^<]+</sym>$") then 472 else 471}
	//
	// `match(1000000, …)` trifft nicht, also 471 — und weil MSE die Angabe
	// danach auf 471 setzt, trifft es auch im zweiten Durchlauf nicht. Dort
	// steht **immer** 471.
	//
	// Fassung f9d9356 vom 2021-07-06.
	Kantenvorgabe = 1000000

	// Laengenvorgabe gilt für `width` und `height` — `src/data/field.cpp`,
	// Zeile 102.
	//
	// Fassung f9d9356 vom 2021-07-06.
	Laengenvorgabe = 0

	// Winkelvorgabe gilt für `angle` — `src/data/field.cpp`, Zeile 104.
	//
	// Fassung f9d9356 vom 2021-07-06.
	Winkelvorgabe = 0

	// Polstervorgabe gilt für `padding left/right/top/bottom` —
	// `src/data/field/text.cpp`, Zeilen 58–61.
	//
	// Fassung f9d9356 vom 2021-07-06.
	Polstervorgabe = 0

	// Zeilenhoehenvorgabe gilt für `line height soft`, `hard` und `line` —
	// `src/data/field/text.cpp`, Zeilen 62–64. **Eins und nicht null**: es
	// ist ein Faktor.
	//
	// Fassung f9d9356 vom 2021-07-06.
	Zeilenhoehenvorgabe = 1.0

	// Hoechstzeilenhoehenvorgabe gilt für die drei `… max` —
	// `src/data/field/text.cpp`, Zeilen 65–67.
	//
	// Fassung f9d9356 vom 2021-07-06.
	Hoechstzeilenhoehenvorgabe = 0.0

	// Absatzvorgabe gilt für `paragraph height` —
	// `src/data/field/text.cpp`, Zeile 68. **Minus eins** heißt „nicht
	// gesetzt"; null wäre eine Absatzhöhe von null.
	//
	// Fassung f9d9356 vom 2021-07-06.
	Absatzvorgabe = -1

	// Schriftgroessenvorgabe gilt für `font: size` — `src/data/font.cpp`,
	// Zeile 16.
	//
	// Fassung f9d9356 vom 2021-07-06.
	Schriftgroessenvorgabe = 1

	// Kleinstgroessenvorgabe gilt für `font: scale down to` —
	// `src/data/font.cpp`, Zeile 18.
	//
	// Fassung f9d9356 vom 2021-07-06.
	Kleinstgroessenvorgabe = 100000

	// Eckenvorgabe gilt für `radius` eines Farbfelds —
	// `src/data/field/color.cpp`, Zeile 60.
	//
	// Fassung f9d9356 vom 2021-07-06.
	Eckenvorgabe = 0

	// Randvorgabe gilt für die vier `… width` eines Farbfelds —
	// `src/data/field/color.cpp`, Zeilen 61–62. **Dieselbe Zahl, die
	// [randbreiten] als „kein Rand" führt**; sie steht dort nicht zufällig.
	//
	// Fassung f9d9356 vom 2021-07-06.
	Randvorgabe = 100000

	// Symbolgroessenvorgabe gilt für `symbol font: size` —
	// `src/data/symbol_font.cpp`, Zeile 554.
	//
	// Fassung f9d9356 vom 2021-07-06.
	Symbolgroessenvorgabe = 12
)
