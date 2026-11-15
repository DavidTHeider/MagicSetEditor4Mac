package msescript

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// **`combined_editor` setzt mehrere Felder zu einem zusammen.**
//
// # Wozu es gebraucht wird
//
// Das Feld `text` von `magic.mse-game` trägt `save value: false` und entsteht
// erst daraus:
//
//	combined_editor(field1: card.rule_text, separator: "<line>\n</line>",
//	                field2: card.flavor_text)
//
// Regeltext, Blockgrenze, Flavortext — der ganze Textkasten einer Karte
// (#345).
//
// # Was hier bewusst fehlt
//
// Bei MSE ist `combined_editor` zugleich ein **Rückweg**: es schreibt beim
// Bearbeiten die Teile in ihre Felder zurück und benachrichtigt das Set
// (`src/script/functions/editor.cpp`, Zeilen 60–95). Dieses Programm zeichnet
// nur; der Rückweg gehört zur Oberfläche und steht hier nicht. Was bleibt, ist
// das Zusammensetzen — und das ist die Hälfte, die ein Bild braucht.
//
// Fassung f9d9356 vom 2021-07-06.
func init() {
	eintragen("combined_editor", verbundeneFelder)

	// **`forward_editor` ist derselbe Name für dieselbe Funktion.**
	//
	//	ctx.setVariable(_("forward_editor"), script_combined_editor); // compatability
	//
	// `src/script/functions/editor.cpp`, Zeile 390 — MSE führt ihn aus
	// Rücksicht auf ältere Vorlagen, und das Pack benutzt ihn noch
	// (#270).
	//
	// Fassung f9d9356 vom 2021-07-06.
	eintragen("forward_editor", verbundeneFelder)
}

// verbundeneFelder ist `combined_editor`.
//
// Die Regel steht in `src/script/functions/editor.cpp`, Zeilen 28–52 und
// 96–116, Fassung f9d9356 vom 2021-07-06:
//
//	new_value = value_parts.front();
//	for (i = 1 …) new_value += "<sep>" + separators[i-1] + "</sep>" + value_parts[i];
//
// **Der Trenner steht in `<sep>`-Klammern.** MSE schneidet daran beim
// Bearbeiten; gezeichnet wird sein Inhalt, und bei `magic.mse-game` ist das ein
// `<line>` — die Blockgrenze zwischen Regel- und Flavortext.
func verbundeneFelder(_ context.Context, ruf Ruf) (Wert, error) {
	felder := durchnummeriert(ruf, "field")
	if len(felder) == 0 {
		return Nichts, errors.New("combined_editor bekam kein einziges Feld")
	}

	trenner := durchnummeriert(ruf, "separator")
	if len(trenner) < len(felder)-1 {
		return Nichts, fmt.Errorf(
			"combined_editor bekam %d Felder und nur %d Trenner",
			len(felder), len(trenner))
	}

	// **Drei Fälle für den Trenner, und die Reihenfolge ist die Aussage:**
	//
	//	if (value_parts[i].second && new_value_empty && hide_when_empty) {
	//	  // no separator
	//	} else if (value_parts[i].second && soft_before_empty) {
	//	  new_value += _("<sep-soft>") + separators[i - 1] + _("</sep-soft>");
	//	} else {
	//	  new_value += _("<sep>") + separators[i - 1] + _("</sep>");
	//	}
	//
	// Zeilen 96–116. Ganz weg nur, wenn auch alles davor leer ist; sonst
	// **weich**; sonst gewöhnlich. `magic.mse-game` baut die Typzeile mit
	// beiden Angaben — bei einer Zauberei ohne Untertyp steht der
	// Gedankenstrich damit weich da, und weicher Text zählt bei der
	// Ausrichtung nicht mit
	// (#345).
	verstecken := false
	if wert, da := ruf.Benannter("hide_when_empty"); da {
		verstecken, _ = wert.AlsWahrheit()
	}

	weich := false
	if wert, da := ruf.Benannter("soft_before_empty"); da {
		weich, _ = wert.AlsWahrheit()
	}

	var bau strings.Builder

	bau.WriteString(felder[0])

	leer := istEntauszeichnetLeer(felder[0])

	for stelle := 1; stelle < len(felder); stelle++ {
		teilLeer := istEntauszeichnetLeer(felder[stelle])

		// **Nur der Trenner fällt weg, nicht der Teil.**
		//
		//	if (value_parts[i].second && new_value_empty && hide_when_empty) {
		//	  // no separator
		//	} else if …
		//	new_value += value_parts[i].first;
		//
		// Zeilen 104–115: das Anhängen steht **hinter** der Verzweigung. Diese
		// Umsetzung sprang mit `continue` darüber hinweg und verlor damit ein
		// leeres Wortlistenfeld — das die Oberfläche zum Tippen braucht.
		switch {
		case teilLeer && leer && verstecken:
			// kein Trenner, und `leer` bleibt, wie es war

		case teilLeer && weich:
			bau.WriteString("<sep-soft>" + trenner[stelle-1] + "</sep-soft>")

			leer = false

		default:
			bau.WriteString("<sep>" + trenner[stelle-1] + "</sep>")

			leer = false
		}

		bau.WriteString(felder[stelle])
	}

	return Text(bau.String()), nil
}

// istEntauszeichnetLeer sagt, ob zwischen den Marken kein Zeichen steht.
//
//	nv.second = index_to_untagged(nv.first, nv.first.size()) == 0;
//
// `src/script/functions/editor.cpp`, Zeile 93, Fassung f9d9356 vom
// 2021-07-06. **Nicht der Vergleich mit der leeren Kette**, und der
// Unterschied ist auf jeder zweiten Karte zu sehen: `magic.mse-game` gibt
// einen leeren Untertyp als leeres Wortlistenfeld zurück —
//
//	if sub_types == "" then "<word-list-{list_type_first}></word-list-{list_type_first}>"
//
// `magic.mse-game/script`, Zeile 2363. Wer damit gegen `""` verglich, nahm den
// harten Trenner, und der Gedankenstrich stand hinter `Instant`,
// `Enchantment`, `Sorcery` — 1.308 von 3.788 Karten des kanonischen Sets
// `Black` (#725).
//
// **Ein `<` ohne schließende Klammer verschluckt den Rest.** `index_to_untagged`
// zählt über `skip_tag`, und das gibt ohne `>` ein `npos`
// (`src/util/tagged_string.cpp`, Zeilen 543–556) — dieselbe Regel wie in
// `to_html`.
func istEntauszeichnetLeer(text string) bool {
	for stelle := 0; stelle < len(text); {
		if text[stelle] != '<' {
			return false
		}

		zu := strings.IndexByte(text[stelle:], '>')
		if zu < 0 {
			return true
		}

		stelle += zu + 1
	}

	return true
}

// durchnummeriert liest `name`, `name2`, `name3`, … so weit die Reihe geht.
//
// **`field` und `field2`, nicht `field1` und `field2`.** MSE hängt die Nummer
// erst ab dem zweiten an (`String name = _("field"); if (i > 0) name = name <<
// i;`) — und beginnt bei `i = 0`, also mit `field`, dann `field1`, `field2`.
// Die Vorlagen schreiben `field1` und `field2`; `field` ohne Nummer kommt
// dort nicht vor, und beide Formen werden deshalb gelesen.
//
// **Die Reihe endet an der ersten Lücke**, so wie bei MSE (`else if (i > 0)
// break`).
func durchnummeriert(ruf Ruf, name string) []string {
	var reihe []string

	for stelle := 0; ; stelle++ {
		gesucht := name
		if stelle > 0 {
			gesucht = name + strconv.Itoa(stelle)
		}

		wert, da := ruf.Benannter(gesucht)
		if !da {
			if stelle > 0 {
				break
			}

			continue
		}

		reihe = append(reihe, wert.AlsText())
	}

	return reihe
}
