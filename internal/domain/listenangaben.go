package domain

// **Was ein Spiel über die Spalte eines Kartenfeldes sagt** — Werkstück 3 von
// Phase 3.7
// (#1168).
//
// **Eine eigene Datei**, weil `karte.go` mit ihr über 400 Zeilen ginge (§2).
// Der Schnitt folgt einer Frage: dort steht, **was ein Feld ist**, hier, wie
// die Kartenliste es zeigt.

// Listenangaben sind die `card list …:`-Angaben eines Kartenfeldes.
//
// **Ein eingebetteter Verbund und keine fünf Stellen in [Field]**: sie gehören
// zusammen, sie kommen aus derselben Quelle, und sie werden zusammen gelesen.
// Eingebettet, damit `feld.Listenstelle` weiter zu schreiben ist — der
// Verbund ordnet, ohne die Aufrufer zu ändern.
type Listenangaben struct {
	//	int       card_list_column; ///< What column to use in the card list?
	//	UInt      card_list_width;  ///< Width of the card list column (pixels).
	//	bool      card_list_visible;///< Is this field shown in the card list?
	//	bool      card_list_allow;  ///< Is this field allowed to appear in the card list?
	//	Alignment card_list_align;  ///< Alignment of the card list colummn.
	//
	// `src/data/field.hpp`, Zeilen 55–60, Fassung f9d9356 vom 2021-07-06.
	//
	// **Es ist der Vorschlag des Spiels, nicht die Wahl der Nutzenden.** MSE
	// legt daraus die `ColumnSettings` an und merkt sie **je Spiel**
	// (`src/data/settings.hpp`, Zeile 78); was danach gilt, steht dort.

	// Listenstelle ist `card list column:` — die **Reihenfolge**, nicht die
	// Nummer der Spalte. Das Pack zählt in Hundertern (`0400`, `1610`), damit
	// zwischen zwei Feldern noch Platz ist; die höchste Stelle in
	// `magic.mse-game` ist **9999**. Wer sie als Index läse, legte ein Feld
	// mit zehntausend Einträgen an.
	Listenstelle int

	// Listenbreite ist `card list width:` in Punkten; 0 heißt „ungesagt".
	Listenbreite int

	// Listensichtbar ist `card list visible:` — **ohne Angabe falsch**.
	//
	// **Gemessen über `magic.mse-game`** (`internal/msepaket/listenspaltenfremd_test.go`,
	// 2026-08-27, über den Leser dieses Projekts und samt eingebundenen
	// Dateien): 243 Kartenfelder tragen eine Stelle, **elf** davon sind
	// sichtbar. Ohne diese Vorgabe stünden 243 Spalten nebeneinander.
	Listensichtbar bool

	// Listenerlaubt ist `card list allow:` — **ohne Angabe wahr**.
	//
	//	if (cs.visible && f->card_list_allow) { … }
	//
	// `src/gui/control/card_list.cpp`, Zeile 218, Fassung f9d9356 vom
	// 2021-07-06. Ein Feld mit `false` darf
	// **nie** in die Liste, auch wenn jemand es wählte — es steht deshalb gar
	// nicht erst zur Wahl. **Sechs** der 249 Kartenfelder von
	// `magic.mse-game` tragen es.
	Listenerlaubt bool

	// Listenausrichtung ist `card list alignment:` — links, rechts oder mittig.
	Listenausrichtung Listenausrichtung
}
