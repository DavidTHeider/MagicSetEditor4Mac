package service

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"

// mitSchriftschaltern legt Gewicht und Lage der **Schrift** auf jeden
// Abschnitt — MSEs `FONT_BOLD` und `FONT_ITALIC` aus `weight: bold` und
// `style: italic`.
//
//	flags = (flags & ~FONT_BOLD & ~FONT_ITALIC)
//	      | (weight() == _("bold")   ? FONT_BOLD   : FONT_NORMAL)
//	      | (style()  == _("italic") ? FONT_ITALIC : FONT_NORMAL);
//
// `Font::update`, `src/data/font.cpp`, Zeilen 26–41, Fassung f9d9356 vom
// 2021-07-06. **Eine Auszeichnung legt ihren Schalter dazu**
// (`f->flags |= add_flags`, ebd. Zeilen 54–56): fett ist, was die Schrift
// **oder** ein `<b>` fett macht.
//
// **Bis F02 fehlte das ganz**
// (#2244):
// 768 `weight:` im Pack, und der Name jeder Archenemy-, Planechase- und
// Vanguard-Karte stand aufrecht da.
func mitSchriftschaltern(abschnitte []render.Abschnitt, fett, kursiv bool) []render.Abschnitt {
	for i := range abschnitte {
		abschnitte[i].Fett = abschnitte[i].Fett || fett
		abschnitte[i].Kursiv = abschnitte[i].Kursiv || kursiv
	}

	return abschnitte
}
