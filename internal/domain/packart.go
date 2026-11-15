package domain

// **Die Packarten, die ein Set selbst mitbringt**
// (#1081).
//
// # Warum ein Set eigene haben darf
//
// MSE sucht eine Packart **zuerst im Set**, dann im Spiel:
//
//	FOR_EACH_CONST(type, set->pack_types)       { … }
//	FOR_EACH_CONST(type, set->game->pack_types) { … }
//
// `src/data/pack.cpp`, Zeilen 373–392, Fassung f9d9356 vom 2021-07-06. Das
// Feld heißt dort ausdrücklich *„Additional/replacement pack types"*
// (`src/data/set.hpp`, Zeile 57) — **beides**: ergänzen und ersetzen.
//
// # Warum die Gestalt hier steht und nicht `msepaket.Pack` reicht
//
// `internal/domain` importiert nichts aus dem Projekt (§2). Die Wiederholung
// ist der Preis dafür, und sie ist derselbe wie bei [Keyword].

// PackArt ist eine Packart eines Sets — ein Booster, ein Blatt, eine Gruppe.
type PackArt struct {
	// ID ist die Kennung im Bestand — 0 für eine, die es dort noch nicht gibt.
	ID int64

	Name string

	// Select ist `select` — **schon aufgelöst**, wie in `msepaket.Pack`.
	//
	// `auto` steht in der Datei; was es heißt, entscheidet
	// `after_reading(PackType&)` (`src/data/pack.cpp`, Zeilen 43–46, Fassung
	// f9d9356 vom 2021-07-06): mit Filter `no replace`, mit Posten `all`.
	Select string

	// Filter ist ein MSE-Script-Ausdruck über `card` — oder leer.
	Filter string

	// Selectable sagt, ob die Oberfläche die Art zur Auswahl stellt.
	Selectable bool

	// Items sind die Unterpacks.
	Items []PackItem

	// Reihenfolge ist die Stelle, an der diese Art in der **Datei** stand.
	//
	// **Sie entscheidet nichts fachlich** — anders als bei [Keyword], wo die
	// Reihenfolge sagt, welches zuerst passt. Hier ist sie allein der Ort, an
	// den zurückgeschrieben wird (0011).
	Reihenfolge int
}

// PackItem ist ein Posten in einer Packart.
type PackItem struct {
	Name string

	// Amount ist `amount` — wie viele Züge aus diesem Posten kommen.
	// **Ohne Angabe eins** (`PackItem::PackItem`, `src/data/pack.cpp`,
	// Zeilen 75–79, Fassung f9d9356 vom 2021-07-06).
	Amount int

	// Weight ist `weight` — die relative Wahrscheinlichkeit. **Ohne Angabe
	// eins**, ebenda.
	Weight float64
}
