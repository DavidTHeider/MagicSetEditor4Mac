package msescript

// **`curly_quotes` — gerade Anführungszeichen werden runde.**

import (
	"context"
	"strings"
)

// Die Zeichen aus `src/util/string.hpp`, Zeilen 78–83, Fassung f9d9356 vom
// 2021-07-06 (Unicode-Zweig; der andere gilt für Fassungen ohne Unicode).
const (
	linkesEinfaches   = '‘'
	rechtesEinfaches  = '’'
	linkesDoppeltes   = '“'
	rechtesDoppeltes  = '”'
	halbgeviertstrich = '–'
	geviertstrich     = '—'
)

func init() { eintragen("curly_quotes", rundeAnfuehrung) }

// rundeAnfuehrung ist `curly_quotes`.
//
//	String curly_quotes(String str, bool curl) {
//	  bool open = true, in_tag = false;
//	  for (wxUniCharRef c : str) {
//	    if (c == '\'' || c == LEFT_SINGLE_QUOTE || c == RIGHT_SINGLE_QUOTE) {
//	      c = curl ? (open ? LEFT_SINGLE_QUOTE : RIGHT_SINGLE_QUOTE) : '\'';
//	    } else if (c == '"' || c == LEFT_DOUBLE_QUOTE || c == RIGHT_DOUBLE_QUOTE) {
//	      c = curl ? (open ? LEFT_DOUBLE_QUOTE : RIGHT_DOUBLE_QUOTE) : '"';
//	    } else if (c == '<') { in_tag = true;
//	    } else if (c == '>') { in_tag = false;
//	    } else if (!in_tag) {
//	      // Also allow double-nesting of quotes
//	      open = is_space_like(c);
//	    }
//	  }
//	  return str;
//	}
//
// `src/util/tagged_string.cpp`, Zeilen 763–780, Fassung f9d9356 vom
// 2021-07-06. Das Skript ruft es mit `curl = true`
// (`src/script/functions/basic.cpp`, Zeilen 373–376) — die Gegenrichtung gibt
// es nur im Editor.
//
// # Zwei Dinge, die der Kopf leicht übersieht
//
// **Ein Anführungszeichen ändert `open` nicht.** Nach `"` bleibt stehen, was
// davor galt; deshalb wird in `""` das zweite geschlossen und nicht wieder
// geöffnet.
//
// **Innerhalb einer Auszeichnung ändert sich `open` gar nicht.** `<i>` würde
// sonst als Wortzeichen zählen, und das Anführungszeichen dahinter schlösse.
func rundeAnfuehrung(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("curly_quotes"); err != nil {
		return Nichts, err
	}

	var (
		gebaut  strings.Builder
		offen   = true
		imMarke = false
	)

	for _, zeichen := range ruf.Stellung[0].AlsText() {
		switch zeichen {
		case '\'', linkesEinfaches, rechtesEinfaches:
			gebaut.WriteRune(rundesZeichen(offen, linkesEinfaches, rechtesEinfaches))

		case '"', linkesDoppeltes, rechtesDoppeltes:
			gebaut.WriteRune(rundesZeichen(offen, linkesDoppeltes, rechtesDoppeltes))

		case '<':
			imMarke = true

			gebaut.WriteRune(zeichen)

		case '>':
			imMarke = false

			gebaut.WriteRune(zeichen)

		default:
			if !imMarke {
				offen = istWieLeerraum(zeichen)
			}

			gebaut.WriteRune(zeichen)
		}
	}

	return Text(gebaut.String()), nil
}

// rundesZeichen gibt das erste Zeichen, wenn geöffnet wird, sonst das zweite.
func rundesZeichen(offen bool, links, rechts rune) rune {
	if offen {
		return links
	}

	return rechts
}

// istWieLeerraum sagt, ob nach diesem Zeichen ein Anführungszeichen **öffnet**.
//
//	bool is_space_like(wxUniChar c) {
//	  return isSpace(c) || c == '(' || c == '[' || c == '{' || c == EN_DASH || c == EM_DASH;
//	}
//
// `src/util/tagged_string.cpp`, Zeile 760, Fassung f9d9356 vom 2021-07-06.
func istWieLeerraum(zeichen rune) bool {
	switch zeichen {
	case ' ', '\t', '\n', '\r', '\f', '\v', '(', '[', '{',
		halbgeviertstrich, geviertstrich:
		return true
	}

	return false
}
