package msepaket

import (
	"sort"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Stil darf eigene Felder mitbringen — `extra card field:`.**
//
//	REFLECT(extra_card_fields);
//	REFLECT_IF_READING { extra_card_style.init(extra_card_fields); }
//	REFLECT(extra_card_style);
//
// `src/data/stylesheet.cpp`, Zeilen 116–125, Fassung f9d9356 vom 2021-07-06.
//
// **Daran hängt der Kasten für Stärke und Widerstandskraft.** `magic-m15`
// zeichnet ihn nicht über ein Kartenfeld, sondern über `pt box` — ein Feld, das
// es im Spiel nicht gibt und das die Vorlage aus `card.card_color` rechnet
// (#270).

// zusatzteile liest die `extra card field:`- und `extra card style:`-Blöcke.
func zusatzteile(kopf *Baum, befunde []Befund) ([]domain.Field, []Feldstil, []Befund) {
	var (
		felder []domain.Field
		stile  []Feldstil
	)

	for _, knoten := range kopf.Alle("extra card field") {
		feld, befund := feldAus(knoten, "extra card field")
		if befund != nil {
			befunde = append(befunde, *befund)

			continue
		}

		felder = append(felder, feld)
	}

	for _, block := range kopf.Alle("extra card style") {
		for _, knoten := range block.Kinder {
			if knoten.Zeilenart != ArtEintrag {
				continue
			}

			feld, befund := feldstilAus(knoten)
			if befund != nil {
				befunde = append(befunde, *befund)
			}

			stile = append(stile, feld)
		}
	}

	// **Nach `z index`, und stabil** — derselbe Grund wie bei den Kartenfeldern.
	sort.SliceStable(stile, func(a, b int) bool { return stile[a].ZIndex < stile[b].ZIndex })

	return felder, stile, befunde
}
