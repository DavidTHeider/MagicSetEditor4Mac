package domain

// **Was ein Manasymbol bedeutet** — Werkstück M2 der Regelmaschine
// (#1604).
//
// # Was hier neu ist gegenüber `manakosten.go`
//
// **Dort wird getrennt, hier wird gedeutet.** [Manasymbole] zerlegt `{2}{R}`
// in `{2}` und `{R}`; bis zum 2026-09-17 hat niemand gefragt, **was** das
// heißt — `docs/REGELABDECKUNG.md` führte CR 107 mit dem Satz „gedeutet wird
// keines".
//
// # Die Quellen
//
// CR 107.3 und CR 107.4a–h (Fassung 2026-08-19): fünf farbige Symbole, Zahlen
// und `{X}` als generisch, `{C}` als farblos, Hybride — auch der einfarbige
// `{2/B}` —, phyrexianische samt den zehn hybrid-phyrexianischen, und `{S}`
// für Schnee, das **weder Farbe noch Manaart** ist.
//
// # Was es ausdrücklich nicht tut
//
// **Es weist nichts ab.** `{Zauberstab}` ist kein Fehler, sondern ein
// **unbekanntes** Symbol — §8: ein MSE-Paket von 2007 darf schreiben, was es
// will, und die Symbolschrift der Vorlage entscheidet, ob es sich zeichnen
// lässt. Was ein unbekanntes Symbol zur Farbe beiträgt, ist nichts.

import (
	"strconv"
	"strings"
)

// Manasymbol ist ein gedeutetes Symbol aus einem Manakosten.
//
// **Ein Verbund und keine Aufzählung.** `{W/P}` ist zugleich farbig und
// phyrexianisch, `{2/B}` zugleich generisch und schwarz; eine Aufzählung
// bräuchte für jede Kombination einen eigenen Wert.
type Manasymbol struct {
	// Roh ist das Symbol, wie es dastand.
	Roh string

	// Zahl sagt, ob eine Ziffer darin stand — CR 107.4b.
	//
	// **Getrennt von [Manasymbol.Generisch]**, weil `{0}` eine Zahl ist und
	// null beiträgt, ein `{W}` aber keine Zahl ist und **eins**. Ohne dieses
	// Feld wären beide nicht zu unterscheiden.
	Zahl bool

	// Generisch ist die Zahl generischen Manas — CR 107.4b.
	//
	// **Bei `{2/B}` ist sie zwei** (CR 107.4e): das Symbol lässt sich mit
	// zwei beliebigen **oder** einem schwarzen zahlen, und CR 202.3 zählt den
	// höheren Weg.
	Generisch int

	// Variabel sagt, ob `{X}`, `{Y}` oder `{Z}` dastand — CR 107.3.
	//
	// **Für den Manawert zählt es null** (CR 107.3b), solange der Zauber
	// nicht auf dem Stapel liegt. Den gibt es hier nicht (`M1`).
	Variabel bool

	// Farben sind die Farben dieses Symbols — CR 107.4a, 107.4e, 107.4f.
	//
	// **Leer heißt: keine.** Generisch, farblos und Schnee sind keine Farben
	// (CR 107.4b, 107.4c, 107.4h).
	Farben []Farbe

	// Farblos sagt, ob `{C}` dastand — CR 107.4c.
	Farblos bool

	// Phyrexianisch sagt, ob es auch mit zwei Leben zu zahlen ist —
	// CR 107.4f.
	Phyrexianisch bool

	// Schnee sagt, ob `{S}` dastand — CR 107.4h.
	//
	// **Weder Farbe noch Manaart**, sagt die Regel ausdrücklich.
	Schnee bool

	// Unbekannt sagt, dass hier etwas steht, das CR 107 nicht kennt — §8.
	Unbekannt bool
}

// SymbolAus deutet ein einzelnes Symbol — **mit oder ohne Klammern**.
//
// **Beides, weil beide Schreibweisen vorkommen**: Scryfall und MTG JSON
// schreiben `{W/U}`, MSEs `add_cards_scripts` schreibt `W/U`. Der Grund steht
// ausführlich bei [ManakostenOhneKlammern].
func SymbolAus(roh string) Manasymbol {
	kern := strings.TrimSpace(roh)
	kern = strings.TrimSuffix(strings.TrimPrefix(kern, "{"), "}")
	kern = strings.TrimSuffix(strings.TrimPrefix(kern, "["), "]")

	symbol := Manasymbol{Roh: roh}

	// **Der Schrägstrich trennt die Hälften** — CR 107.4e und CR 107.4f. Ein
	// Symbol ohne ihn hat genau eine.
	for _, teil := range strings.Split(kern, "/") {
		deuteHaelfte(&symbol, teil)
	}

	return symbol
}

// deuteHaelfte trägt eine Hälfte in das Symbol ein.
func deuteHaelfte(symbol *Manasymbol, teil string) {
	gross := strings.ToUpper(strings.TrimSpace(teil))

	if farbe, passt := FarbeAus(gross); passt {
		symbol.Farben = append(symbol.Farben, farbe)

		return
	}

	switch gross {
	// **`P` und `H` meinen dasselbe** — CR 107.4f schreibt das
	// phyrexianische Symbol als `{W/P}`, CR 107.4g nennt `{H}` als das
	// Symbol ohne Farbhintergrund. **Die kanonischen Sets schreiben `H`, und
	// zwar vorn**: `[H/R]` für Gut Shot, gemessen am 2026-09-17.
	case "P", "H":
		symbol.Phyrexianisch = true

	case "C":
		symbol.Farblos = true

	case "S":
		symbol.Schnee = true

	// **CR 107.3**: der Wert wird beim Wirken angesagt. `{Y}` und `{Z}`
	// stehen auf denselben Karten daneben und meinen dasselbe.
	case "X", "Y", "Z":
		symbol.Variabel = true

	default:
		zahl, fehler := strconv.Atoi(gross)
		if fehler != nil {
			symbol.Unbekannt = true

			return
		}

		symbol.Zahl = true
		symbol.Generisch = zahl
	}
}

// Manawert ist, wie viel Mana dieses Symbol beiträgt — CR 202.3.
//
// **Eins für alles, was keine Zahl ist** — ein farbiges, ein farbloses und
// ein Schneesymbol sind je ein Mana. Null für `{X}` (CR 107.3b) und null für
// das Unbekannte: eine Zahl zu erfinden wäre schlimmer als keine (§10).
func (s Manasymbol) Manawert() int {
	switch {
	case s.Zahl:
		return s.Generisch

	case s.Variabel, s.Unbekannt:
		return 0
	}

	return 1
}
