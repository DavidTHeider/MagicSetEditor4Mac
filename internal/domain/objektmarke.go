package domain

// **Marken an jedem Fachobjekt** — UX-12
// (#1684),
// Vermerk [0135](../../docs/entscheidungen/0135-marken-an-jedem-fachobjekt.md),
// Marke `new` wie die Kartenmarke aus 0053: MSE2 kennt keine.
//
// # Was hier steht
//
// **Welche Arten eine Marke tragen, was ein Filter aus mehreren Marken heißt
// und wie ein gespeicherter Filter aussieht.** Die Normalform einer Marke
// steht in `marke.go` und gilt hier unverändert — „Tausch" und „tausch" sind
// auch an einem Deck dieselbe Marke.

import (
	"slices"
	"strconv"
	"strings"
)

// Markenart ist die Art eines Objekts, das eine Marke trägt.
type Markenart string

// Die Markenarten.
const (
	MarkenartKarte    Markenart = "karte"
	MarkenartSet      Markenart = "set"
	MarkenartExemplar Markenart = "exemplar"
	MarkenartDeck     Markenart = "deck"
	MarkenartPaket    Markenart = "paket"
	MarkenartTurnier  Markenart = "turnier"
	MarkenartPartie   Markenart = "partie"
	MarkenartHandel   Markenart = "handel"

	// MarkenartSymbol trägt ein Symbol der Bibliothek — DOM-02 (#1697).
	MarkenartSymbol Markenart = "symbol"
)

// Markenarten sind alle Arten, die eine Marke tragen.
//
// **Das Symbol kam mit DOM-02** (#1697): Vermerk 0135, Punkt 9 ließ es
// weg, solange es ein Feld eines Sets war. Seit es eine Bibliothek gibt, ist
// es ein eigenes Objekt — und trägt seine Marken dort, nicht am Set.
func Markenarten() []Markenart {
	return []Markenart{
		MarkenartKarte, MarkenartSet, MarkenartExemplar, MarkenartDeck,
		MarkenartPaket, MarkenartTurnier, MarkenartPartie, MarkenartHandel,
		MarkenartSymbol,
	}
}

// PruefeMarkenart liest eine Art — oder sagt, dass es sie nicht gibt.
func PruefeMarkenart(roh string) (Markenart, error) {
	art := Markenart(strings.TrimSpace(roh))
	if !slices.Contains(Markenarten(), art) {
		return "", NewFieldError("art", SchluesselValidierung,
			"»%s« trägt keine Marke — es gibt Karte, Set, Exemplar, Deck, Paket, Turnier, Partie, Handel und Symbol.", roh)
	}

	return art, nil
}

// Kennung bringt die Kennung eines Objekts in die eine Form, unter der es
// seine Marken trägt.
//
// **Eine Zahlart nimmt nur eine ganze Zahl ab eins** und gibt sie ohne
// führende Null und ohne Vorzeichen zurück. SQLite vergliche `011` oder
// `11.0` mit der Zahlspalte als 11 — die Marke läge dann unter einer Form in
// `objektmarken`, die weder die Abfrage mit `11` noch der Löschauslöser
// (`CAST(OLD.id AS TEXT)`, 0084) findet. **Text bleibt Text**: eine Partie
// heißt nach ihrer Textkennung (0064), ein Handel nach seinem Vorgang, ein
// Paket nach Art und Name.
func (a Markenart) Kennung(roh string) (string, error) {
	if a == MarkenartPartie || a == MarkenartHandel || a == MarkenartPaket {
		return roh, nil
	}

	zahl, fehler := strconv.ParseInt(strings.TrimSpace(roh), 10, 64)
	if fehler != nil || zahl < 1 {
		return "", NewFieldError("kennung", SchluesselValidierung,
			"»%s« ist keine Kennung — hier zählt eine ganze Zahl ab eins.", roh)
	}

	return strconv.FormatInt(zahl, 10), nil
}

// Markenposten ist eine Marke des Profils mit der Zahl der Objekte, die sie
// tragen — über alle Arten, die Karte eingeschlossen.
type Markenposten struct {
	Marke  string `json:"marke"`
	Anzahl int    `json:"anzahl"`
}

// Verknuepfung sagt, wie mehrere Marken eines Filters zusammengehen.
type Verknuepfung string

// Die Verknüpfungen.
const (
	// VerknuepfungUnd heißt: **jede** genannte Marke.
	VerknuepfungUnd Verknuepfung = "und"

	// VerknuepfungOder heißt: **wenigstens eine**.
	VerknuepfungOder Verknuepfung = "oder"
)

// Markenfilter ist eine Frage nach Marken.
type Markenfilter struct {
	Marken       []string     `json:"marken"`
	Verknuepfung Verknuepfung `json:"verknuepfung"`
}

// NeuerMarkenfilter baut einen Filter — jede Marke in ihrer Normalform.
//
// **Ohne Verknüpfung heißt es „und"**: wer zwei Marken nennt, ohne mehr zu
// sagen, meint in jeder Suchmaske, die man kennt, beide.
func NeuerMarkenfilter(marken []string, verknuepfung string) (Markenfilter, error) {
	filter := Markenfilter{Marken: make([]string, 0, len(marken)), Verknuepfung: VerknuepfungUnd}

	switch Verknuepfung(strings.TrimSpace(verknuepfung)) {
	case "", VerknuepfungUnd:
	case VerknuepfungOder:
		filter.Verknuepfung = VerknuepfungOder
	default:
		return Markenfilter{}, NewFieldError("verknuepfung", SchluesselValidierung,
			"»%s« ist keine Verknüpfung — es gibt und und oder.", verknuepfung)
	}

	for _, roh := range marken {
		marke, fehler := PruefeMarke(roh)
		if fehler != nil {
			return Markenfilter{}, fehler
		}

		if !slices.Contains(filter.Marken, marke) {
			filter.Marken = append(filter.Marken, marke)
		}
	}

	return filter, nil
}

// Trifft sagt, ob ein Objekt mit diesen Marken durch den Filter geht.
//
// **Ohne Marke ist es kein Filter** — dann geht alles durch.
func (f Markenfilter) Trifft(getragen []string) bool {
	if len(f.Marken) == 0 {
		return true
	}

	if f.Verknuepfung == VerknuepfungOder {
		return slices.ContainsFunc(f.Marken, func(marke string) bool { return slices.Contains(getragen, marke) })
	}

	return !slices.ContainsFunc(f.Marken, func(marke string) bool { return !slices.Contains(getragen, marke) })
}
