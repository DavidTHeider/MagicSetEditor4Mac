package service

import (
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// layoutAus macht aus einem gesetzten Zug das `layout`-Objekt einer Vorlage.
//
// # Was darin steht
//
// `lines`, `paragraphs`, `blocks` und `separators`, dazu die Maße des Ganzen —
// so wie `TextLayout` sie hergibt (`src/data/field/text.hpp`, Zeilen 50–67, und
// `src/data/field/text.cpp`, Zeilen 114–133, Fassung f9d9356 vom 2021-07-06).
//
// Jeder Eintrag trägt `width`, `top`, `height`, `bottom` und `middle`;
// `bottom` und `middle` sind abgeleitet (`top + height` und `top + height/2`).
//
// **Ein ungemessenes Feld bekommt ein leeres Layout, nicht keines.** Ein
// `card_style.<feld>.layout.blocks[0].bottom or else 0` liest daraus null; ohne
// das Objekt liest es einen Befund
// (#338).
func layoutAus(mass Inhaltsmass) msescript.Wert {
	gesetzt := mass.Satz

	// **Ein Verbund, kein gewöhnliche Abbildung.** `TextLayout` ist bei MSE ein
	// reflektierter C++-Verbund, und seine Glieder werden über den **Präfix**
	// gefunden ([msescript.Wert.Verbund]). Im Pack greifen **28 Stellen** auf
	// `layout.separator` in der **Einzahl** zu und 35 auf `separators`
	// (#494).
	return msescript.Wert{Art: msescript.ArtAbbildung, Verbund: true, Eintraege: []msescript.Eintrag{
		{Name: "width", Wert: msescript.Zahl(mass.Breite)},
		{Name: "height", Wert: msescript.Zahl(mass.Hoehe)},
		{Name: "top", Wert: msescript.Zahl(0)},
		{Name: "bottom", Wert: msescript.Zahl(mass.Hoehe)},
		{Name: "middle", Wert: msescript.Zahl(mass.Hoehe / 2)},
		{Name: "lines", Wert: zeilenliste(gesetzt.Zeilen)},
		{Name: "paragraphs", Wert: gruppenliste(gesetzt.Absaetze)},
		{Name: "blocks", Wert: gruppenliste(gesetzt.Bloecke)},
		{Name: "separators", Wert: zahlenliste(gesetzt.Trenner)},
	}}
}

// zeilenliste macht aus gesetzten Zeilen eine Liste von Layouteinträgen.
func zeilenliste(zeilen []render.Zeile) msescript.Wert {
	glieder := make([]msescript.Eintrag, 0, len(zeilen))

	for stelle, zeile := range zeilen {
		glieder = append(glieder, msescript.Eintrag{
			Name: strconv.Itoa(stelle),
			Wert: layouteintrag(zeile.Breite, zeile.Oben, zeile.Hoehe),
		})
	}

	return msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: glieder}
}

// gruppenliste macht aus Absätzen oder Blöcken eine Liste von Layouteinträgen.
func gruppenliste(gruppen []render.Zeilengruppe) msescript.Wert {
	glieder := make([]msescript.Eintrag, 0, len(gruppen))

	for stelle, gruppe := range gruppen {
		glieder = append(glieder, msescript.Eintrag{
			Name: strconv.Itoa(stelle),
			Wert: layouteintrag(gruppe.Breite, gruppe.Oben, gruppe.Hoehe),
		})
	}

	return msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: glieder}
}

// zahlenliste macht aus Zahlen eine Liste.
func zahlenliste(zahlen []float64) msescript.Wert {
	glieder := make([]msescript.Eintrag, 0, len(zahlen))

	for stelle, zahl := range zahlen {
		glieder = append(glieder, msescript.Eintrag{
			Name: strconv.Itoa(stelle), Wert: msescript.Zahl(zahl),
		})
	}

	return msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: glieder}
}

// layouteintrag baut einen Eintrag des Layouts.
func layouteintrag(breite, oben, hoehe float64) msescript.Wert {
	// **Ein Verbund, kein gewöhnliche Abbildung.** `TextLayout` ist bei MSE ein
	// reflektierter C++-Verbund, und seine Glieder werden über den **Präfix**
	// gefunden ([msescript.Wert.Verbund]). Im Pack greifen **28 Stellen** auf
	// `layout.separator` in der **Einzahl** zu und 35 auf `separators`
	// (#494).
	return msescript.Wert{Art: msescript.ArtAbbildung, Verbund: true, Eintraege: []msescript.Eintrag{
		{Name: "width", Wert: msescript.Zahl(breite)},
		{Name: "top", Wert: msescript.Zahl(oben)},
		{Name: "height", Wert: msescript.Zahl(hoehe)},
		{Name: "bottom", Wert: msescript.Zahl(oben + hoehe)},
		{Name: "middle", Wert: msescript.Zahl(oben + hoehe/2)},
	}}
}
