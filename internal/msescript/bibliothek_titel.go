package msescript

import (
	"context"
	"strings"
	"unicode"
)

// `to_title` und `to_color`
// (#472).
//
//	// convert a string to title case
//	SCRIPT_FUNCTION(to_title) {
//	  SCRIPT_PARAM_C(String, input);
//	  SCRIPT_RETURN(capitalize(input.Lower()));
//	}
//
// `src/script/functions/basic.cpp`, Zeilen 218–221, Fassung f9d9356 vom
// 2021-07-06.
//
// **`to_title` steht in sieben Stilen** beider Packs, `to_color` viermal im
// vollen Pack.
func init() {
	eintragen("to_title", zuTitel)
	eintragen("to_color", zuFarbwert)
}

// kleineWorte bleiben in einem Titel klein — außer am Anfang.
//
//	if (it != s.begin() &&
//	    (is_substr(it, s.end(), _("is ")) || is_substr(it, s.end(), _("the ")) ||
//	     is_substr(it, s.end(), _("in ")) || is_substr(it, s.end(), _("of "))  ||
//	     is_substr(it, s.end(), _("to ")) || is_substr(it, s.end(), _("at "))  ||
//	     is_substr(it, s.end(), _("a " )))) {
//	  // Short words are not capitalized, keep lower case
//	}
//
// `capitalize_in_place`, `src/util/string.cpp`, Fassung f9d9356 vom
// 2021-07-06.
//
// **Mit dem Leerzeichen dahinter.** `is_substr` prüft `"is "` samt Abstand —
// „Island" fängt mit `is` an und wird trotzdem groß geschrieben.
var kleineWorte = []string{"is ", "the ", "in ", "of ", "to ", "at ", "a "}

// zuTitel schreibt jedes Wort groß, außer den kleinen.
func zuTitel(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("to_title"); err != nil {
		return Nichts, err
	}

	return Text(Grossgeschrieben(strings.ToLower(ruf.Stellung[0].AlsText()))), nil
}

// Grossgeschrieben ist `capitalize_in_place` — jedes Wort groß, außer den
// kleinen mitten im Namen.
//
// **Der Schrägstrich trennt wie ein Leerzeichen** — `*it == ' ' || *it == '/'`.
// In einem Kartennamen steht er zwischen zwei Hälften.
//
// **Exportiert, weil MSE2 dieselbe Funktion an zwei weiteren Stellen nimmt**:
// für die Namen von Paketen (`PackageList::drawItem`,
// `src/gui/control/package_list.cpp`, Zeile 41) und für die Einträge eines
// Symbolschriftmenüs ohne Übersetzung (`src/data/symbol_font.cpp`, Zeilen
// 545–549). **Eine Fassung und nicht drei.**
func Grossgeschrieben(text string) string {
	runen := []rune(text)
	nachAbstand := true

	for stelle := range runen {
		if runen[stelle] == ' ' || runen[stelle] == '/' {
			nachAbstand = true

			continue
		}

		if !nachAbstand {
			continue
		}

		nachAbstand = false

		if stelle > 0 && beginntKlein(runen[stelle:]) {
			continue
		}

		runen[stelle] = unicode.ToUpper(runen[stelle])
	}

	return string(runen)
}

// beginntKlein sagt, ob hier eines der kleinen Worte anfängt.
func beginntKlein(rest []rune) bool {
	for _, wort := range kleineWorte {
		if strings.HasPrefix(string(rest), wort) {
			return true
		}
	}

	return false
}

// zuFarbwert gibt eine Farbe zurück.
//
//	SCRIPT_FUNCTION(to_color) {
//	  try {
//	    SCRIPT_PARAM_C(Color, input);
//	    SCRIPT_RETURN(input);
//	  } catch (const ScriptError& e) {
//	    return delay_error(e);
//	  }
//	}
//
// `src/script/functions/basic.cpp`, Zeilen 208–215, Fassung f9d9356 vom
// 2021-07-06.
//
// **Der Fehler wird verzögert, nicht geworfen** (`delay_error`): was sich
// nicht in eine Farbe umwandeln lässt, wird ein Fehlerwert — und der wirft
// erst, wenn ihn jemand benutzt.
func zuFarbwert(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("to_color"); err != nil {
		return Nichts, err
	}

	wert := ruf.Stellung[0]
	if _, istFarbe := wert.AlsFarbe(); !istFarbe {
		return Fehlerwert("to_color: " + wert.AlsText() + " ist keine Farbe"), nil
	}

	return wert, nil
}
