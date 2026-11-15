package msescript_test

import (
	"testing"
)

// **`sort_text` behält nur, was seine Ordnung nennt.**
//
//	} else { // single char
//	  size_t count = count_and_remove(it.value, input);
//	  ret.append(count, it.value);
//	}
//	…
//	return ret;
//
// `spec_sort`, `src/util/spec_sort.cpp`, Zeilen 266–352, Fassung f9d9356 vom
// 2021-07-06. Die Schleife hängt an `ret` **nur** an, was die Spezifikation
// nennt; was in `input` übrig bleibt, wird verworfen. Der Aufrufer
// (`String spec_sort(const String& spec, String input)`, Zeile 355) gibt `ret`
// zurück und wirft `input` weg.
//
// # Was das gekostet hat
//
// Ein kanonisches MSE2-Set speichert Manakosten **mit Klammern**:
// `casting cost: [3][W]`. Der Kartenwert läuft durch
// `mana_filter → sort_text@(order: "\\?XYZI[0123456789]VHSFCAI(EKPLO)(WUBRG)")`
// (`magic.mse-game/script:509`), und dort fallen die Klammern weg — sie stehen
// in der Ordnung nicht als Zeichen, sondern als Gruppenklammer.
//
// Dieses Programm behielt sie und zeichnete sie als Symbole: auf **jeder**
// gewöhnlichen Karte standen vier leere Kästen neben den Manakosten
// (#411).
func TestSortTextVerwirftWasDieOrdnungNichtNennt(t *testing.T) {
	faelle := map[string]struct{ ordnung, eingabe, erwartet string }{
		"Klammern der kanonischen Sets": {
			ordnung: "XYZI[0123456789]HSCA(WUBRG)", eingabe: "[3][W]", erwartet: "3W",
		},
		"ohne Klammern unverändert": {
			ordnung: "XYZI[0123456789]HSCA(WUBRG)", eingabe: "3W", erwartet: "3W",
		},
		"nicht genannte Buchstaben fallen weg": {
			ordnung: "(WUBRG)", eingabe: "WxU", erwartet: "WU",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			gebaut, err := ordne(t, fall.ordnung, fall.eingabe)
			if err != nil {
				t.Fatalf("ordnen: %v", err)
			}

			if gebaut != fall.erwartet {
				t.Errorf("sort_text(%q, order: %q) = %q, erwartet %q",
					fall.eingabe, fall.ordnung, gebaut, fall.erwartet)
			}
		})
	}
}
