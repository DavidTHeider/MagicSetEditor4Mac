package domain

// **Was die Kartenliste zeigt und wonach sie ordnet** — Werkstück 3 von
// Phase 3.7
// (#1168).
//
// # Die Wahl gehört dem Spiel
//
//	map<String, ColumnSettings> cardlist_columns;
//	String                      sort_cards_by;
//	bool                        sort_cards_ascending;
//
// `src/data/settings.hpp`, Zeilen 78–80, Fassung f9d9356 vom 2021-07-06 — und
// sie stehen dort in **`GameSettings`**. Wer sich seine Spalten für Magic
// einrichtet, findet sie in jedem Magic-Set wieder.

// Spaltenwahl ist, was jemand über **eine** Spalte entschieden hat.
//
// **Ohne Breite.** MSE merkt sich zusätzlich `ColumnSettings::width`, weil
// dort eine `wxListCtrl` steht, deren Spalten man mit der Maus zieht. Hier
// legt das Fenster die Breite selbst — eine Zahl in Punkten, die aus einem
// Programm von 2007 stammt, sagt über einen heutigen Bildschirm nichts
// (§15, Regel 7).
type Spaltenwahl struct {
	// Feld ist der Feldname, wie das Spielpaket ihn schreibt: `casting cost`,
	// nicht `casting_cost`. **Nicht übersetzt** (§9).
	Feld string

	Sichtbar bool

	// Stelle ist die **Reihenfolge**, nicht die Nummer der Spalte — MSEs
	// `ColumnSettings::position`, das aus `card list column:` kommt.
	Stelle int
}

// Kartenordnung ist, wonach die Liste sortiert.
type Kartenordnung struct {
	// Feld ist der Feldname — **leer heißt: gar nicht sortieren**, wie MSEs
	// `sort_by_column = -1` und `gs.sort_cards_by = wxEmptyString`.
	Feld string

	// Aufsteigend ist **ohne Wahl wahr**.
	//
	//	: sort_cards_ascending(true)
	//
	// `src/data/settings.cpp`, Zeile 72, Fassung f9d9356 vom 2021-07-06.
	Aufsteigend bool
}

// StandardKartenordnung ist, was ohne Wahl gilt.
func StandardKartenordnung() Kartenordnung {
	return Kartenordnung{Aufsteigend: true}
}

// Listenspalte ist eine Spalte, wie die Liste sie anbietet.
type Listenspalte struct {
	// Feld ist der Feldname, wie das Spielpaket ihn schreibt.
	Feld string

	// Aufschrift ist `card list name:` — leer, wenn das Paket keine nennt;
	// dann steht der Feldname da, und das entscheidet die Oberfläche.
	Aufschrift string

	Sichtbar    bool
	Ausrichtung Listenausrichtung
}

// Listenzeile ist eine Karte, wie die Liste sie zeigt.
type Listenzeile struct {
	KarteID int64

	// Nummer ist die Stelle im **Set** und nicht in der Liste — sie bleibt,
	// wenn jemand anders sortiert.
	Nummer int

	// Werte trägt **jede** angebotene Spalte, auch die unsichtbaren: wer eine
	// Spalte einblendet, soll nicht auf eine neue Antwort warten müssen.
	//
	// **Ohne Auszeichnungen** — ein `<i>` in einer Tabellenzelle sieht wie ein
	// Fehler aus.
	Werte map[string]string

	// Marken sind die Marken dieser Karte, alphabetisch —
	// #1208.
	//
	// **In der Zeile und nicht in einem zweiten Ruf.** Sie werden **gezeigt**
	// und es wird nach ihnen **gefiltert**; beides braucht sie zu jeder Zeile,
	// und eine Abfrage je Zeile wären bei 3.788 Karten 3.788 Abfragen.
	//
	// **Nicht in [Listenzeile.Werte].** Dort stehen Feldwerte des Spiels; eine
	// Marke ist keiner, und eine Spalte „marke" kollidierte mit einem
	// Kartenfeld dieses Namens, sobald eine Vorlage eines führt.
	Marken []string

	// Schluessel sind die Ergebnisse der `sort script:` dieser Zeile —
	// #1232.
	//
	// **Nur für Felder, die eines haben**, und deshalb meist leer.
	//
	// **Getrennt von [Listenzeile.Werte]**: `getSortKey()` gibt den
	// Sortierwert, `toString()` den Text. Wer beides in einer Abbildung
	// führte, zeigte den Schlüssel in der Zelle — und ein `1` statt `common`
	// ist keine Auskunft.
	//
	// **Sie geht nicht ans Fenster.** Die Ordnung entsteht hier; die Antwort
	// trägt die fertige Reihenfolge, und ein zweiter Satz Werte dort wäre
	// Ballast, den niemand liest.
	Schluessel map[string]string
}

// Kartenlistenaufbau ist, was die Liste braucht: Spalten, Zeilen, Ordnung.
type Kartenlistenaufbau struct {
	Spalten []Listenspalte
	Zeilen  []Listenzeile

	// Ordnung ist die **wirksame** — nicht die gemerkte. Zeigt die gemerkte
	// auf ein Feld, das es nicht mehr gibt, steht hier keines.
	Ordnung Kartenordnung

	// OhneVorlage sagt, dass die Kartenfelder des Spiels nicht zu lesen waren
	// — dann kommen die Spalten aus den **Karten**.
	//
	// **Es steht in der Antwort und nicht in einem Fehler** (§8): ein Set aus
	// MTG JSON ohne sein Spielpaket bliebe sonst ohne Kartenliste, und die
	// Liste ist der Weg zu jeder Karte. Was fehlt, sind Aufschrift,
	// Reihenfolge und Verbot — und wer den Hinweis sieht, weiß, warum.
	OhneVorlage bool
}
