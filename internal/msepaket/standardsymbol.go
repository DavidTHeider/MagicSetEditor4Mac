package msepaket

// **Ohne Datei zeichnet MSE sein Standardsymbol.**
//
//	SymbolP the_symbol;
//	if (filename.empty()) {
//	  the_symbol = default_symbol();
//	} else {
//	  the_symbol = package->readFile<SymbolP>(filename);
//	}
//
// `SymbolToImage::generate`, `src/gfx/generated_image.cpp`, Zeilen 453–462,
// Fassung f9d9356 vom 2021-07-06.
//
// **Ein Set ohne eigenes Symbol ist der Normalfall**, nicht der Ausnahmefall:
// die vier Referenzkarten sind solche, und MSE malt an ihrer Typzeile das
// Quadrat. Bis zu diesem Werkstück meldete der Rasterer „die Symbolzeichnung
// liegt nicht vor" und zeichnete nichts
// (#270).

// Standardsymbol gibt MSEs `default_symbol()` — ein Quadrat.
//
//	SymbolShapeP default_symbol_part(double d) {
//	  part->points.push_back(make_intrusive<ControlPoint>(d + .2, d + .2));
//	  part->points.push_back(make_intrusive<ControlPoint>(d + .2, d + .8));
//	  part->points.push_back(make_intrusive<ControlPoint>(d + .8, d + .8));
//	  part->points.push_back(make_intrusive<ControlPoint>(d + .8, d + .2));
//	  part->name = _("Square");
//	}
//	SymbolP default_symbol() {
//	  auto symbol = make_intrusive<Symbol>();
//	  symbol->parts.push_back(default_symbol_part(0));
//	}
//
// `src/data/symbol.cpp`, Zeilen 361–378. **`d` ist null** — den verschobenen
// Zweitteil benutzt MSE nur beim Anlegen einer neuen Symboldatei im Editor.
//
// Fassung f9d9356 vom 2021-07-06.
func Standardsymbol() Symbolzeichnung {
	ecken := []Vektor{{X: 0.2, Y: 0.2}, {X: 0.2, Y: 0.8}, {X: 0.8, Y: 0.8}, {X: 0.8, Y: 0.2}}

	punkte := make([]Kontrollpunkt, 0, len(ecken))
	for _, ecke := range ecken {
		// **Geraden**: `ControlPoint(double x, double y)` ist MSEs Konstruktor
		// „for straight lines", und er lässt beide Griffe auf null.
		punkte = append(punkte, Kontrollpunkt{Stelle: ecke})
	}

	return Symbolzeichnung{Teile: []Symbolteil{{
		Art: TeilForm, Name: "Square", Punkte: punkte,
	}}}
}
