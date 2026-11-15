package domain

// **Farbe und Manawert eines Kostens** — Werkstück M2 der Regelmaschine
// (#1604).
//
// # Warum neben `manakosten.go` und nicht darin
//
// **Weil dort die Schreibweise steht und hier die Bedeutung.**
// `manakosten.go` rechnet zwischen zwei Schreibweisen um und zerlegt eine
// Zeichenkette; es sagt selbst: „Sie prüft nicht, ob die Symbole gültig sind."
// Hier wird gefragt, **was** sie bedeuten — CR 202.2 und CR 202.3 (Fassung
// 2026-08-19). Zwei Fragen, zwei Dateien (§2).

import "strings"

// Manakosten sind die gedeuteten Symbole eines Kostens — CR 202.1.
type Manakosten struct {
	// Roh ist die Zeichenkette, wie sie in der Karte stand.
	Roh string

	Symbole []Manasymbol
}

// ManakostenAus deutet einen Kosten — in **drei** Schreibweisen.
//
// # Warum drei
//
// **Weil es drei gibt, und alle drei stehen in echten Daten.**
//
//   - `{2}{R}` — so schreiben Scryfall und MTG JSON; der Beleg steht bei
//     [Manasymbole].
//   - `[2][R]` — so schreiben die **kanonischen Sets**. Gemessen am
//     2026-09-17 an `Red.mse-set` (3.730 Karten): `[2][R]`, `[2/R][2/R][2/R]`
//     für Flame Javelin, `[H/R]` für Gut Shot.
//   - `2R` — so legt MSEs `add_cards_scripts` Karten an; der Beleg steht bei
//     [ManakostenOhneKlammern].
//
// **Die zweite fehlte bis zur ersten Sichtprüfung mit einem echten Deck**,
// und sie fiel nicht als Fehler auf, sondern als **falsche Zahl**: ohne sie
// zerfällt `[2/R]` in fünf Stücke, und Flame Javelin kostete neun statt
// sechs. Die Farbe stimmte dabei die ganze Zeit — das ist die Gestalt, vor
// der §10 warnt.
//
// **Ohne Klammern zählt jedes Zeichen einzeln**, und mehrstellige Zahlen
// bleiben zusammen: `10R` ist zehn und rot, nicht eins, null und rot.
func ManakostenAus(roh string) Manakosten {
	kosten := Manakosten{Roh: roh}

	stuecke := Manasymbole(roh)
	if len(stuecke) == 0 {
		stuecke = eckigeSymbole(roh)
	}

	if len(stuecke) == 0 {
		stuecke = stueckeOhneKlammern(roh)
	}

	for _, stueck := range stuecke {
		kosten.Symbole = append(kosten.Symbole, SymbolAus(stueck))
	}

	return kosten
}

// eckigeSymbole zerlegt `[2][R]` in `[2]` und `[R]`.
//
// **Dieselbe Bauart wie [Manasymbole]**, nur mit dem anderen Klammerpaar —
// und mit derselben Zurückhaltung: ein halb geschriebenes Symbol wird nicht
// geraten. Aus `[2][R` kommt nur `[2]`.
func eckigeSymbole(kosten string) []string {
	var symbole []string

	rest := kosten

	for {
		auf := strings.Index(rest, "[")
		if auf < 0 {
			return symbole
		}

		zu := strings.Index(rest[auf:], "]")
		if zu < 0 {
			return symbole
		}

		symbole = append(symbole, rest[auf:auf+zu+1])
		rest = rest[auf+zu+1:]
	}
}

// stueckeOhneKlammern zerlegt `10R` in `10` und `R`.
//
// **Ziffern bleiben zusammen, alles andere steht für sich.** Ein Kosten ohne
// Klammern kann keine Hybride tragen — dafür bräuchte er den Schrägstrich, und
// den schreibt MSE dann auch in Klammern.
func stueckeOhneKlammern(roh string) []string {
	var (
		stuecke []string
		zahl    strings.Builder
	)

	for _, zeichen := range strings.TrimSpace(roh) {
		if zeichen >= '0' && zeichen <= '9' {
			zahl.WriteRune(zeichen)

			continue
		}

		if zahl.Len() > 0 {
			stuecke = append(stuecke, zahl.String())
			zahl.Reset()
		}

		if zeichen != ' ' {
			stuecke = append(stuecke, string(zeichen))
		}
	}

	if zahl.Len() > 0 {
		stuecke = append(stuecke, zahl.String())
	}

	return stuecke
}

// Farben gibt die Farben dieses Kostens — CR 202.2.
//
// **In der Reihenfolge des Regelwerks und ohne Wiederholung**: CR 202.2c
// spricht von „zwei oder mehr **verschiedenen**" farbigen Symbolen, und eine
// Karte mit zwei weißen Symbolen ist einmal weiß.
//
// **Hybride und phyrexianische zählen mit** — CR 202.2d: ein Objekt ist
// **alle** Farben solcher Symbole.
func (m Manakosten) Farben() []Farbe {
	var gesammelt []Farbe

	for _, symbol := range m.Symbole {
		gesammelt = append(gesammelt, symbol.Farben...)
	}

	return SortiereFarben(gesammelt)
}

// Farblos sagt, ob dieses Objekt keine Farbe hat — CR 202.2b, CR 105.2c.
//
// **Das ist nicht dasselbe wie `{C}`.** Ein Kosten aus lauter generischen
// Symbolen ist farblos, ohne ein einziges `{C}` zu tragen; und ein Land ohne
// jeden Kosten ist es auch (CR 202.1b).
func (m Manakosten) Farblos() bool {
	return len(m.Farben()) == 0
}

// Manawert ist die Summe des Manas in diesem Kosten — CR 202.3.
//
// **Ohne Rücksicht auf die Farbe**, sagt die Regel; und ohne Kosten ist er
// null (CR 202.3a). Was CR 202.3b–d über zweiseitige, verschmolzene und
// geteilte Karten sagen, steht hier nicht — dafür braucht es mehrere Seiten
// zugleich, und das ist `M12`
// (#1614).
func (m Manakosten) Manawert() int {
	summe := 0

	for _, symbol := range m.Symbole {
		summe += symbol.Manawert()
	}

	return summe
}
