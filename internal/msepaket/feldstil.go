package msepaket

import (
	"strconv"
)

// feldstilAus liest einen Feldstil aus seinem Knoten.
func feldstilAus(knoten *Knoten) (Feldstil, *Befund) {
	feld := Feldstil{Feld: knoten.Schluessel}

	var befund *Befund

	for _, kind := range knoten.Kinder {
		if kind.Zeilenart != ArtEintrag {
			continue
		}

		// **In kanonischer Form verglichen**, wie MSE2 jeden Schlüssel liest
		// (`canonical_name_form`, `src/util/io/reader.cpp` Zeile 217, Fassung
		// f9d9356 vom 2021-07-06): `z_index:` und `z index:` sind dasselbe (#1972).
		switch Entkanonisch(kind.Schluessel) {
		// **Ein Maß darf ein eingerückter `script:`-Block sein.**
		//
		// `magic-modules.mse-include/stamps/card_fields` schreibt alle vier so.
		// Der Wert des Knotens ist dann leer, und das Skript steht in seinen
		// Zeilen — dieselbe Form wie bei `alignment:`. Wer nur den Wert liest,
		// bekommt nichts, und die Ebene fällt auf die Maße der ganzen Karte
		// zurück: das 43 Punkte große Prägesiegel wurde über die volle Karte
		// gezeichnet
		// (#295).
		case "left":
			feld.Links = massangabeAus(kind)
		case "top":
			feld.Oben = massangabeAus(kind)
		case "width":
			feld.Breite = massangabeAus(kind)
		case "height":
			feld.Hoehe = massangabeAus(kind)
		case "right":
			feld.Rechts = massangabeAus(kind)
		case "bottom":
			feld.Unten = massangabeAus(kind)
		case "alignment":
			feld.Ausrichtung = kind.Wert
			if skript := kind.Ersten("script"); skript != nil {
				feld.AusrichtungSkript = skript.Skripttext()
			}
		// **`visible:` entscheidet, ob die Ebene gezeichnet wird.**
		//
		//	bool ValueViewer::isVisible() const { return getStyle()->visible && …; }
		//
		// `src/render/value/viewer.cpp`, Zeilen 36–42, Fassung f9d9356 vom
		// 2021-07-06. `magic-m15` schreibt am Kasten für Stärke und
		// Widerstandskraft `visible: { card.pt != "" }`; ohne die Angabe steht
		// auf jeder Verzauberung ein leerer Kasten
		// (#270).
		case "visible":
			feld.Sichtbar = massangabeAus(kind)

		case "mask":
			feld.Maske = massangabeAus(kind)
		case "render style":
			feld.Darstellung = kind.Wert
		// **Auch ein Bild darf seinen Ausdruck als `script:`-Unterblock
		// schreiben.** `magic-modules.mse-include/stamps/card_fields` tut das
		// für `card stamp`; über `Full-Magic-Pack` steht der Unterblock
		// **56-mal** an `image:` und zweimal an `mask:`, gezählt an der
		// Angabe (§10).
		//
		// Roh gelesen war die Angabe **leer**, und dann galt die Wahlliste —
		// dort steht für den Stempel „none" eine Datei, die nicht leer ist.
		// Auf jeder Karte der M15-Vorlagen lag deshalb ein dunkler Fleck
		// (#454).
		case "image":
			feld.Bild = massangabeAus(kind)
		case "default":
			feld.Standardbild = massangabeAus(kind)

		case "choice images":
			feld.Wahlbilder = wahlbilderAus(kind)

		case "z index":
			nummer, err := strconv.Atoi(kind.Wert)
			if err != nil {
				befund = &Befund{
					Stelle:  kind.Stelle,
					Meldung: "z index ist keine Zahl: " + kind.Wert,
				}

				continue
			}

			feld.ZIndex = nummer

		case "font":
			feld.Schrift = schriftAus(kind)
		case "symbol font":
			feld.Symbolschrift = symbolschriftAus(kind)
		case "always symbol":
			feld.ImmerSymbole = jaOderNein(kind)
		case "angle":
			feld.Winkel = massangabeAus(kind)
		case "padding left":
			feld.Polster.Links = kind.Wert
		case "padding right":
			feld.Polster.Rechts = kind.Wert
		case "padding top":
			feld.Polster.Oben = kind.Wert
		case "padding bottom":
			feld.Polster.Unten = kind.Wert
		case "line height soft":
			feld.Zeilenhoehen.Weich = kind.Wert
		case "line height hard":
			feld.Zeilenhoehen.Hart = kind.Wert
		case "line height line":
			feld.Zeilenhoehen.Linie = kind.Wert
		case "line height soft max":
			feld.Zeilenhoehen.WeichMax = kind.Wert
		case "line height hard max":
			feld.Zeilenhoehen.HartMax = kind.Wert
		case "paragraph height":
			feld.Zeilenhoehen.Absatz = kind.Wert

		case "line height line max":
			feld.Zeilenhoehen.LinieMax = kind.Wert

		// **Wie eine Mehrfachwahl als Reihe ihre Wahlen legt** (#2297).
		case "spacing":
			feld.Abstand = massangabeAus(kind)
		case "direction":
			feld.Richtung = kind.Wert

		// **Die Angaben eines Farbfeldstils** — Radius und die vier
		// Randbreiten (`ColorStyle`, `src/data/field/color.cpp`, Zeilen 66–74,
		// Fassung f9d9356 vom 2021-07-06). Daran hängt der Kartenrand.
		case "radius":
			feld.Ecke = massangabeAus(kind)
		case "left width":
			feld.Rand.Links = massangabeAus(kind)
		case "right width":
			feld.Rand.Rechts = massangabeAus(kind)
		case "top width":
			feld.Rand.Oben = massangabeAus(kind)
		case "bottom width":
			feld.Rand.Unten = massangabeAus(kind)
		}
	}

	return feld, befund
}

// kopfwert gibt den Wert eines Kopfeintrags; leer, wenn es ihn nicht gibt.
func kopfwert(baum *Baum, schluessel string) string {
	knoten := baum.Ersten(schluessel)
	if knoten == nil {
		return ""
	}

	return knoten.Wert
}

// zahl liest eine ganze Zahl; 0, wenn sie keine ist.
func zahl(wert string) int {
	nummer, err := strconv.Atoi(wert)
	if err != nil {
		return 0
	}

	return nummer
}

// feldstileAus liest die Einträge eines Stilblocks — `card style:`,
// `extra card style:` oder `styling style:`.
//
// **Eine Schleife für drei Blöcke**, weil es bei MSE eine ist: alle drei sind
// ein `IndexMap<FieldP, StyleP>` und werden über dieselbe Reflexion gelesen
// (`src/data/stylesheet.cpp`, Zeilen 103–125, Fassung f9d9356 vom
// 2021-07-06). Zwei Kopien liefen auseinander, sobald eine von beiden eine
// Angabe dazubekäme (§10).
func feldstileAus(
	block *Knoten, bisher []Feldstil, befunde []Befund,
) ([]Feldstil, []Befund) {
	for _, knoten := range block.Kinder {
		if knoten.Zeilenart != ArtEintrag {
			continue
		}

		feld, befund := feldstilAus(knoten)
		if befund != nil {
			befunde = append(befunde, *befund)
		}

		bisher = append(bisher, feld)
	}

	return bisher, befunde
}
