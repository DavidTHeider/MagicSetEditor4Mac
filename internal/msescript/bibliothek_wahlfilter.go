package msescript

import (
	"context"
	"strconv"
	"strings"
)

// **Eine Wahlliste sieben** — `remove_choice` und `require_exclusive_choice`.
//
// Ein `choice`-Feld mit Mehrfachauswahl trägt seine Werte als Text: `"a, b,
// c"`. Beide Funktionen rufen dieselbe Siebung mit verschiedenen Grenzen
// (`src/script/functions/editor.cpp`, Zeilen 248–363, Fassung f9d9356 vom
// 2021-07-06):
//
//   - `remove_choice` mit `min = max = 0` — die genannten Wahlen fallen weg.
//   - `require_exclusive_choice` mit `min = max = 1` — genau eine bleibt.
//
// **Zusammen 4.750 Befunde** über die 190 Stile von `M15-Magic-Pack`, sobald
// die einzeiligen Wertskripte laufen
// (#345).
func init() {
	eintragen("remove_choice", wahlEntfernen)
	eintragen("require_exclusive_choice", wahlErzwingen)
	eintragen("exclusive_choice", wahlBegrenzen)
	eintragen("require_choice", wahlFordern)
}

// wahlFordern ist `require_choice`.
//
//	SCRIPT_RETURN(filter_choices(input, choices, 1, (int)choices.size(), last_change));
//
// `src/script/functions/editor.cpp`, Zeilen 331–337, Fassung f9d9356 vom
// 2021-07-06.
//
// **`min` ist eins, `max` die Zahl der Wahlen**: es dürfen alle stehen bleiben,
// aber nicht keine. `magic.mse-game` ruft es an der Kartenfarbe
// (`require_choice(choices: "multicolor, hybrid, land, artifact")`), und ohne
// diese Funktion fielen dort zwei Wertskripte je mehrfarbiger Karte aus
// (#270).
func wahlFordern(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("require_choice"); err != nil {
		return Nichts, err
	}

	bevorzugt := ""
	if wert, da := ruf.Benannter("last_change"); da {
		bevorzugt = wert.AlsText()
	}

	wahlen := gewuenschteWahlen(ruf)

	return Text(gesiebteWahlen(
		ruf.Stellung[0].AlsText(), wahlen, 1, len(wahlen), bevorzugt)), nil
}

// wahlEntfernen ist `remove_choice`.
func wahlEntfernen(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("remove_choice"); err != nil {
		return Nichts, err
	}

	return Text(gesiebteWahlen(
		ruf.Stellung[0].AlsText(), gewuenschteWahlen(ruf), 0, 0, "")), nil
}

// wahlErzwingen ist `require_exclusive_choice`.
func wahlErzwingen(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("require_exclusive_choice"); err != nil {
		return Nichts, err
	}

	bevorzugt := ""
	if wert, da := ruf.Benannter("last_change"); da {
		bevorzugt = wert.AlsText()
	}

	return Text(gesiebteWahlen(
		ruf.Stellung[0].AlsText(), gewuenschteWahlen(ruf), 1, 1, bevorzugt)), nil
}

// wahlBegrenzen ist `exclusive_choice`.
//
//	SCRIPT_RETURN(filter_choices(input, choices, 0, 1, last_change));
//
// `src/script/functions/editor.cpp`, Zeilen 340–346, Fassung f9d9356 vom
// 2021-07-06.
//
// **Die Grenzen sind 0…1 statt 1…1**: es darf **keine** Wahl übrig bleiben,
// aber nicht zwei — das ist der ganze Unterschied zu [wahlErzwingen]
// (#270).
func wahlBegrenzen(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("exclusive_choice"); err != nil {
		return Nichts, err
	}

	bevorzugt := ""
	if wert, da := ruf.Benannter("last_change"); da {
		bevorzugt = wert.AlsText()
	}

	return Text(gesiebteWahlen(
		ruf.Stellung[0].AlsText(), gewuenschteWahlen(ruf), 0, 1, bevorzugt)), nil
}

// gewuenschteWahlen liest die genannten Wahlen — `read_choices_param`.
//
// **Drei Schreibweisen, in dieser Reihenfolge**: `choices:` als Liste, `choice:`
// einzeln, und die nummerierten `choice1:`, `choice2:` … (Zeilen 310–328). Die
// Reihe der nummerierten endet an der ersten Lücke.
func gewuenschteWahlen(ruf Ruf) []string {
	if wert, da := ruf.Benannter("choices"); da {
		return wahlliste(wert.AlsText())
	}

	if wert, da := ruf.Benannter("choice"); da {
		return []string{wert.AlsText()}
	}

	wahlen := make([]string, 0, 2)

	for nummer := 1; ; nummer++ {
		wert, da := ruf.Benannter("choice" + strconv.Itoa(nummer))
		if !da {
			break
		}

		wahlen = append(wahlen, wert.AlsText())
	}

	return wahlen
}

// wahlliste zerlegt `"a, b, c"` in seine Wahlen.
//
// **Der Leerraum vorn fällt weg, der hinten nicht** — `next_choice`
// überspringt Leerzeichen vor einer Wahl und schneidet am Komma (Zeilen
// 216–224).
func wahlliste(text string) []string {
	teile := strings.Split(text, ",")
	wahlen := make([]string, 0, len(teile))

	for _, teil := range teile {
		wahlen = append(wahlen, strings.TrimLeft(teil, " "))
	}

	return wahlen
}

// gesiebteWahlen ist `filter_choices`.
//
// # Wie MSE siebt
//
// Die Eingabe wird durchgegangen; jede Wahl, die in `wahlen` steht und noch
// nicht gesehen wurde, wird **gemerkt** und aus dem Ergebnis genommen. Was
// nicht genannt ist, bleibt an seiner Stelle stehen. Am Ende werden die
// gemerkten wieder angehängt — je nach `min` und `max` ergänzt oder gekürzt.
//
// **Die Reihenfolge ändert sich dabei:** die unbeteiligten Wahlen stehen vorn,
// die behaltenen hinten. Das ist MSEs Verhalten (Zeilen 264–268 gegen
// 299–305) und nicht ein Nebeneffekt dieser Umsetzung.
//
// **Ohne genannte Wahlen bleibt die Eingabe** (Zeile 249).
func gesiebteWahlen(eingabe string, wahlen []string, mindestens, hoechstens int, bevorzugt string) string {
	if len(wahlen) == 0 {
		return eingabe
	}

	gesehen := make([]bool, len(wahlen))
	anzahl := 0

	var uebrig []string

	for _, teil := range wahlliste(eingabe) {
		getroffen := false

		for stelle, wahl := range wahlen {
			if !gesehen[stelle] && strings.HasPrefix(teil, wahl) {
				gesehen[stelle] = true
				anzahl++
				getroffen = true

				break
			}
		}

		if !getroffen && teil != "" {
			uebrig = append(uebrig, teil)
		}
	}

	// **Erst auffüllen, dann kürzen** — dieselbe Reihenfolge wie bei MSE. Die
	// zweite Zahl braucht niemand mehr; `gesehen` trägt das Ergebnis.
	gekuerzt(gesehen, wahlen,
		aufgefuellt(gesehen, wahlen, anzahl, mindestens, bevorzugt),
		hoechstens, bevorzugt)

	for stelle, wahl := range wahlen {
		if gesehen[stelle] {
			uebrig = append(uebrig, wahl)
		}
	}

	return strings.Join(uebrig, ", ")
}

// aufgefuellt ergänzt Wahlen, bis `mindestens` erreicht ist.
//
// **Die bevorzugte zuerst**, danach der Reihe nach (Zeilen 271–288).
func aufgefuellt(gesehen []bool, wahlen []string, anzahl, mindestens int, bevorzugt string) int {
	if anzahl >= mindestens {
		return anzahl
	}

	for stelle, wahl := range wahlen {
		if wahl == bevorzugt {
			if !gesehen[stelle] {
				gesehen[stelle] = true
				anzahl++
			}

			break
		}
	}

	for stelle := range wahlen {
		if anzahl >= mindestens {
			break
		}

		if !gesehen[stelle] {
			gesehen[stelle] = true
			anzahl++
		}
	}

	return anzahl
}

// gekuerzt nimmt Wahlen weg, bis `hoechstens` erreicht ist.
//
// **Von hinten und die bevorzugte zuletzt** (Zeilen 290–298). Das Ergebnis
// steht in `gesehen`; die Zahl gibt es zurück, weil die Schleife sie braucht.
func gekuerzt(gesehen []bool, wahlen []string, anzahl, hoechstens int, bevorzugt string) int {
	for stelle := len(wahlen) - 1; stelle >= 0; stelle-- {
		if anzahl <= hoechstens {
			break
		}

		if !gesehen[stelle] {
			continue
		}

		if hoechstens > 0 && wahlen[stelle] == bevorzugt {
			continue
		}

		gesehen[stelle] = false
		anzahl--
	}

	return anzahl
}
