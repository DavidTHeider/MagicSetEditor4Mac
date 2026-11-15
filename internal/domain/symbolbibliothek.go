package domain

// **Ein Symbol der Bibliothek** — DOM-02.AK-01
// (#1697):
// „Eine Bibliothek bietet Vorschau, Suche, Tags, Kopie, Import/Export sowie
// klare Unterscheidung zwischen Setsymbol und Symbolschrift."
//
// # Was es ist, und was nicht
//
// **Eine gezeichnete `.mse-symbol`, die einem Profil gehört** und in Sets
// eingesetzt werden kann. Ein Set trägt danach eine **Kopie** — ändert sich
// das Symbol der Bibliothek, bleibt das Set, wie es war; den Zusammenhang
// hält der Abdruck (`Abdruck`), nicht ein Verweis.
//
// **Keine Symbolschrift**: die ist ein Paket (`.mse-symbol-font`), gehört
// dem Rechner und wird in der Paketwerkstatt bearbeitet. Die Bibliothek zeigt
// Schriften, aber sie führt keine.
//
// Portierungsmarke **new**: MSE2 hat keine Symbolbibliothek — gesucht in
// `src/gui/symbol/` und `src/data/symbol.cpp` (Fassung f9d9356 vom
// 2021-07-06); ein Symbol lebt dort nur im Set.

import (
	"strings"
	"time"
	"unicode/utf8"
)

// Symbolnamengrenze ist, wie lang der Name eines Symbols sein darf.
const Symbolnamengrenze = 80

// Bibliothekssymbol ist ein Symbol der Bibliothek.
type Bibliothekssymbol struct {
	ID   int64
	Name string

	// Inhalt ist die Zeichnung als `.mse-symbol`-Text — so, wie sie auch in
	// einem Set steht.
	Inhalt string

	// Abdruck ist die Prüfsumme des Inhalts; an ihr erkennt die Bibliothek
	// ein Set, das dasselbe Symbol trägt.
	Abdruck string

	// Vorher ist der Inhalt vor der letzten Änderung — der Weg zurück.
	Vorher string

	Angelegt, Geaendert time.Time
}

// PruefeSymbolname liest einen Namen — getrimmt, nicht leer, einzeilig und
// höchstens [Symbolnamengrenze] Zeichen lang.
func PruefeSymbolname(roh string) (string, error) {
	name := strings.TrimSpace(roh)

	if name == "" || utf8.RuneCountInString(name) > Symbolnamengrenze || strings.ContainsAny(name, "\r\n") {
		return "", NewFieldError("name", SchluesselValidierung,
			"Ein Symbol braucht einen einzeiligen Namen mit höchstens %d Zeichen.", Symbolnamengrenze)
	}

	return name, nil
}
