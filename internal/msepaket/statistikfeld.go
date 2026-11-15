package msepaket

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

// statistikAus liest, was ein Feld der Statistik mitgibt —
// #1985.
//
// **`show statistics:` ohne Angabe ja**, bei einem Bildfeld nein: MSE setzt es
// im Konstruktor von `Field` auf `true` und in dem von `ImageField` auf
// `false` (`src/data/field.cpp`, Zeile 28; `src/data/field/image.hpp`,
// Zeile 27; Fassung f9d9356 vom 2021-07-06).
//
// **`choice colors:` nur, wo es dasteht** — eine leere Tafel und keine sind
// für die Statistik dasselbe, für einen Vergleich zweier Felder nicht.
func statistikAus(knoten *Knoten, art domain.FieldKind) (zeigen bool, farben map[string]string) {
	zeigen = art != domain.KindImage
	if angabe := knoten.Ersten("show statistics"); angabe != nil {
		zeigen = jaOderNein(angabe)
	}

	tafel := knoten.Ersten("choice colors")
	if tafel == nil {
		return zeigen, nil
	}

	return zeigen, farbtafelAus(tafel)
}
