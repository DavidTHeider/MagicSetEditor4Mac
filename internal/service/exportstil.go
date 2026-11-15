package service

// **Der Stil in der Ausfuhr** — QA-04-F3
// (#2039),
// Vermerk [0165](../../docs/entscheidungen/0165-card-regions-aus-mse-2-5.md).
//
// # Was MSE2 dort hat
//
// `set.stylesheet` ist der Stil des Sets, `card.stylesheet` der **eigene** Stil
// einer Karte und sonst `nil` (`REFLECT(stylesheet)`, `src/data/card.cpp`,
// Zeile 77, und `src/data/set.cpp`, Fassung f9d9356 vom 2021-07-06). Beide
// sind Gegenstände mit den Gliedern aus `IMPLEMENT_REFLECTION(StyleSheet)`
// (`src/data/stylesheet.cpp`, Zeilen 94–126) und `Packaged`
// (`src/util/io/package.cpp`, Zeilen 538–547). Gefragt am Orakel über
// `Multicolor.mse-set`: `<stylesheet>`, `short_name` „M15 Mainframe",
// `card_width` 375, und kein `to_string`.
//
// # Was hier davon steht — und was nicht
//
// `short_name`, `full_name`, `card_width`, `card_height`, `card_dpi`: was die
// Vorlagen des Packs lesen. **Nicht** `game`, `icon`, `version`,
// `installer_group`, `card_background`, `init_script` und die Stilfelder —
// kein Exporttemplate des Packs liest sie; ein Zugriff ist ein verzögerter
// Fehler, den `or else` fängt, wie jedes fehlende Glied.
//
// # Und `card_regions` aus MSE 2.5
//
// **Die Liste der Regionen, die der Stil erklärt**: `vector<CardRegionP>
// card_regions` (`src/data/stylesheet.hpp`, Zeile 44; Glieder `name`, `x`,
// `y`, `width`, `height` in `src/data/card_region.cpp`, Zeilen 19–24 — in
// `G-e-n-e-v-e-n-s-i-S/MagicSetEditor2`, Fassung a4b79c9 vom 2026-09-04),
// gelesen von [msepaket.Kartenregion]. In f9d9356 gibt es das Glied nicht, und
// die Lackey-Ausfuhr bricht daran ab (gemessen: Exit 1, keine Datei). Das Pack
// ist für 2.5.8 geschrieben. Marke `adapted`.
//
// **Hier stand „kein Stil darin erklärt Regionen"**, und die Liste war immer
// leer — gesucht nach `card regions`, geschrieben steht `card region:` je
// Eintrag; acht Stile tun es (N1 des Nachlaufs über die QA-Arbeit).

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// stilgegenstand macht aus einem Stilpaket den Gegenstand, den die Ausfuhr
// zeigt — **derselbe Name, dieselbe Kennung**: zwei Karten mit demselben Stil
// tragen denselben Gegenstand, wie in MSE ein `StyleSheet` je Paket.
func stilgegenstand(stil msepaket.Stil) msescript.Wert {
	return msescript.Wert{
		Art:     msescript.ArtAbbildung,
		Kennung: "stil:" + stilname(stil),
		Eintraege: []msescript.Eintrag{
			{Name: "short_name", Wert: msescript.Text(Kurzname(stil.KurzerName, stil.Paket))},
			{Name: "full_name", Wert: msescript.Text(stil.VollerName)},
			{Name: "card_width", Wert: msescript.Zahl(float64(stil.Breite))},
			{Name: "card_height", Wert: msescript.Zahl(float64(stil.Hoehe))},
			{Name: "card_dpi", Wert: msescript.Zahl(float64(stil.DPI))},
			{Name: "card_regions", Wert: regionenwert(stil.Regionen)},
		},
	}
}

// regionenwert macht aus den Regionen eines Stils die Liste, die die Vorlagen
// über `card_regions[i]` lesen — je Region ein Gegenstand mit `name`, `x`,
// `y`, `width` und `height`, wie MSE 2.5 sie reflektiert.
func regionenwert(regionen []msepaket.Kartenregion) msescript.Wert {
	eintraege := make([]msescript.Eintrag, 0, len(regionen))

	for _, region := range regionen {
		eintraege = append(eintraege, msescript.Eintrag{Wert: msescript.Wert{
			Art: msescript.ArtAbbildung,
			Eintraege: []msescript.Eintrag{
				{Name: "name", Wert: msescript.Text(region.Name)},
				{Name: "x", Wert: msescript.Zahl(region.X)},
				{Name: "y", Wert: msescript.Zahl(region.Y)},
				{Name: "width", Wert: msescript.Zahl(float64(region.Breite))},
				{Name: "height", Wert: msescript.Zahl(float64(region.Hoehe))},
			},
		}})
	}

	return msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: eintraege}
}

// stilwert gibt den Gegenstand zu einem Namen — `nil`, wenn es keine Quelle
// gibt oder sie ihn nicht kennt; im zweiten Fall mit einem Befund.
func stilwert(stile Ausfuhrstile, paket, wessen string) (msescript.Wert, []domain.Befund) {
	if stile == nil {
		return msescript.Nichts, nil
	}

	stil, da := stile(paket)
	if !da {
		return msescript.Nichts, []domain.Befund{{
			Ort:     Kartenstilfeld,
			Meldung: fmt.Sprintf("der Stil %s ist nicht zu finden; die Vorlage sieht nil", wessen),
		}}
	}

	return stilgegenstand(stil), nil
}

// mitStilen setzt `stylesheet` an jeder Karte: den eigenen Stil als
// Gegenstand, sonst `nil`.
func mitStilen(liste msescript.Wert, stile Ausfuhrstile) (msescript.Wert, []domain.Befund) {
	var befunde []domain.Befund

	for i := range liste.Eintraege {
		karte := liste.Eintraege[i].Wert
		karte.Eintraege = append([]msescript.Eintrag(nil), karte.Eintraege...)

		wert := msescript.Nichts
		stelle := -1

		for j, eintrag := range karte.Eintraege {
			if eintrag.Name != Kartenstilfeld {
				continue
			}

			stelle = j

			if name := eintrag.Wert.AlsText(); name != "" {
				var neue []domain.Befund
				wert, neue = stilwert(stile, name, fmt.Sprintf("%q der Karte %d", name, i+1))
				befunde = append(befunde, neue...)
			}
		}

		if stelle < 0 {
			karte.Eintraege = append(karte.Eintraege, msescript.Eintrag{Name: Kartenstilfeld, Wert: wert})
		} else {
			karte.Eintraege[stelle].Wert = wert
		}

		liste.Eintraege[i].Wert = karte
	}

	return liste, befunde
}

// mitSetstil setzt `stylesheet` am Set — **unter seinen Gliedern**: die
// Einträge des Sets sind seine Karten ([setMitKarten]), und ein Glied dort
// fände `set.stylesheet` nicht.
func mitSetstil(set msescript.Wert, stile Ausfuhrstile) (msescript.Wert, []domain.Befund) {
	wert, befunde := stilwert(stile, "", "des Sets")
	set.Glieder = append(append([]msescript.Eintrag(nil), set.Glieder...),
		msescript.Eintrag{Name: Kartenstilfeld, Wert: wert})

	return set, befunde
}
