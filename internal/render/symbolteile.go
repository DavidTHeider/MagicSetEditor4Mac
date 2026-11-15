package render

// **Die Gestalt eines Symbols, wie der Rasterer sie braucht.**
//
// Sie steht hier ein zweites Mal, obwohl `msepaket` sie schon liest: nach §2
// kennt `internal/render` nichts aus dem Projekt. Was fehlt, ist alles, was
// beim Zeichnen keine Rolle spielt — Namen, Sperren, die Griffe eines Editors
// (#323).

// Symbolteilart sagt, was ein Teil ist.
type Symbolteilart int

const (
	// TeilForm ist ein geschlossener Zug aus Punkten.
	TeilForm Symbolteilart = iota

	// TeilSymmetrie vervielfältigt ihre Kinder.
	TeilSymmetrie

	// TeilGruppe fasst Kinder zusammen, ohne sie zu ändern.
	TeilGruppe
)

// Kombinationsart sagt, wie eine Form mit dem verrechnet wird, was dasteht.
//
// **Der Nullwert ist `overlap`**, wie bei MSE.
type Kombinationsart int

const (
	KombinationUeberdecken Kombinationsart = iota
	KombinationVereinen
	KombinationAbziehen
	KombinationSchneiden
	KombinationUnterschied
	KombinationUmranden
)

// Symmetrieart ist `kind:` einer Symmetrie.
type Symmetrieart int

const (
	SpiegelungDrehung Symmetrieart = iota
	SpiegelungAchse
)

// Symbolteil ist ein Teil einer Zeichnung.
type Symbolteil struct {
	Art Symbolteilart

	// Kombination und Punkte gelten für [TeilForm].
	Kombination Kombinationsart
	Punkte      []Symbolpunkt

	// Spiegelart, Kopien, Mitte und Griff gelten für [TeilSymmetrie].
	Spiegelart Symmetrieart
	Kopien     int
	Mitte      Punkt
	Griff      Punkt

	// Teile sind die Kinder einer Symmetrie oder einer Gruppe.
	Teile []Symbolteil
}

// Symbolbereich ist, was aus einem Bildpunkt geworden ist.
//
//	SymbolSet point = data[1] ? (data[0] ? SYMBOL_BORDER : SYMBOL_OUTSIDE)
//	                          : SYMBOL_INSIDE;
//
// `filter_symbol`, `src/render/symbol/filter.cpp`, Zeilen 32–36, Fassung
// f9d9356 vom 2021-07-06.
type Symbolbereich int

const (
	SymbolAussen Symbolbereich = iota
	SymbolRand
	SymbolInnen
)

// Symbolmaske ist ein gerastertes Symbol, noch ohne Farbe.
type Symbolmaske struct {
	Breite, Hoehe int

	// rot und gruen sind MSEs beide Farbkanäle des Schirms. Der Filter liest
	// genau sie, und zwei Kanäle statt eines Bereichs zu führen ist keine
	// Umständlichkeit: die Rechenarten `wxOR` und `wxAND_INVERT` arbeiten auf
	// ihnen, nicht auf einer Deutung.
	rot, gruen *Symbolpuffer
}

// Bei sagt, was aus einem Bildpunkt geworden ist.
func (m Symbolmaske) Bei(x, y int) Symbolbereich {
	if m.gruen.Bei(x, y) == 0 {
		return SymbolInnen
	}

	if m.rot.Bei(x, y) == 0 {
		return SymbolAussen
	}

	return SymbolRand
}
