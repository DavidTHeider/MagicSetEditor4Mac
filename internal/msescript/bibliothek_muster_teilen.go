package msescript

import (
	"context"
	"strconv"
	"strings"
)

// Zwei Funktionen, die aus einem Text mehrere Werte machen.
//
// Beide sind ausgezählt und nicht geraten
// (#214):
// `split_text` steht 176-mal in `M15-Magic-Pack`, `chosen` 300-mal.
func init() {
	eintragen("split_text", teileText)
	eintragen("chosen", gewaehlt)
	eintragen("count_chosen", zahlDerWahlen)
}

// teileText zerlegt einen Text an einem Muster — `split_text(match:"\n", card.rule_text)`.
//
// # Was sie zurückgibt
//
// **Eine Liste**, und in dieser Sprache ist das eine Abbildung mit den Stellen
// als Namen — dieselbe Gestalt, die `cards` in einer Exportvorlage hat. Die
// Vorlagen gehen mit `for each` darüber und messen mit `length`.
//
// # Leere Teile bleiben — `include_empty` ist wahr
//
//	SCRIPT_PARAM_DEFAULT(bool, include_empty, true);
//	if (include_empty || pos.first != start) { … }
//	if (include_empty || start != input.end()) { … }
//
// `src/script/functions/regex.cpp`, Zeilen 212–232, Fassung f9d9356 vom
// 2021-07-06. **Der leere Text ergibt damit ein Element, nicht keines.**
//
// Der naheliegende Einwand — „eine leere Regelbox hätte danach eine Zeile" —
// trägt nicht: fürs Zeilenzählen schreibt `magic.mse-game` ausdrücklich
// `include_empty: false` (`line_count`, Zeile 47).
func teileText(_ context.Context, ruf Ruf) (Wert, error) {
	roh, benannt := ruf.Benannter("match")
	if !benannt {
		return Nichts, ruf.fehlerAn("split_text braucht ein Argument `match:`")
	}

	if err := ruf.eineEingabe("split_text"); err != nil {
		return Nichts, err
	}

	muster, err := UebersetzeMuster(roh.AlsText())
	if err != nil {
		return Nichts, ruf.fehlerAn("split_text bekam " + err.Error())
	}

	leereMit := true
	if wert, da := ruf.Benannter("include_empty"); da {
		leereMit, _ = wert.AlsWahrheit()
	}

	teile, err := muster.Teile(ruf.Stellung[0].AlsText())
	if err != nil {
		return Nichts, ruf.fehlerAn("split_text: " + err.Error())
	}

	eintraege := make([]Eintrag, 0, len(teile))

	for _, teil := range teile {
		if teil == "" && !leereMit {
			continue
		}

		eintraege = append(eintraege,
			Eintrag{Name: strconv.Itoa(len(eintraege)), Wert: Text(teil)})
	}

	return Wert{Art: ArtAbbildung, Eintraege: eintraege}, nil
}

// gewaehlt sagt, ob eine Mehrfachauswahl eine bestimmte Wahl trägt —
// `chosen(card.card_color, choice:"white")`.
//
// # Warum kein `contains`
//
// Ein Mehrfachfeld schreibt seine Wahlen als Text mit Trennern, und `white`
// steckt auch in `whiteblue`. `contains` fände beides; die Vorlagen fragen aber
// nach **einer** Wahl. Gesucht wird deshalb das ganze Wort zwischen den
// Trennern, die MSE benutzt: Komma und Leerzeichen.
func gewaehlt(_ context.Context, ruf Ruf) (Wert, error) {
	roh, benannt := ruf.Benannter("choice")
	if !benannt {
		return Nichts, ruf.fehlerAn("chosen braucht ein Argument `choice:`")
	}

	if err := ruf.eineEingabe("chosen"); err != nil {
		return Nichts, err
	}

	gesucht := strings.TrimSpace(roh.AlsText())

	for _, wahl := range strings.Split(ruf.Stellung[0].AlsText(), ",") {
		if strings.EqualFold(strings.TrimSpace(wahl), gesucht) {
			return Wahrheit(true), nil
		}
	}

	return Wahrheit(false), nil
}

// zahlDerWahlen zählt, wie viele der **genannten** Wahlen ein Feld trägt —
// `count_chosen@(choices: "white,blue,black,red,green,artifact")` aus
// `magic.mse-game`
// (#240).
//
// # Warum `choices:` und nicht einfach zählen
//
// So steht es in allen drei Fundstellen: die Liste der Kandidaten kommt von
// außen. Der Grund ist im Kartenspiel zu sehen — `card_color` trägt neben den
// fünf Farben auch `artifact` und `land`, und die Frage „wie **viel**farbig ist
// die Karte" meint die Farben. Wer stumpf die Einträge zählte, machte aus einem
// farblosen Artefakt eine einfarbige Karte.
//
// **Dieselbe Zerlegung wie [gewaehlt].** Zwei verschiedene Vorstellungen davon,
// was eine Wahl trennt, liefen auseinander: eine Karte gälte als zweifarbig und
// trüge trotzdem keine der beiden Farben.
func zahlDerWahlen(_ context.Context, ruf Ruf) (Wert, error) {
	roh, benannt := ruf.Benannter("choices")
	if !benannt {
		return Nichts, ruf.fehlerAn("count_chosen braucht ein Argument `choices:`")
	}

	if err := ruf.eineEingabe("count_chosen"); err != nil {
		return Nichts, err
	}

	getragen := map[string]bool{}

	for _, wahl := range strings.Split(ruf.Stellung[0].AlsText(), ",") {
		if geputzt := strings.TrimSpace(wahl); geputzt != "" {
			getragen[strings.ToLower(geputzt)] = true
		}
	}

	anzahl := 0

	for _, kandidat := range strings.Split(roh.AlsText(), ",") {
		if getragen[strings.ToLower(strings.TrimSpace(kandidat))] {
			anzahl++
		}
	}

	return Zahl(float64(anzahl)), nil
}
