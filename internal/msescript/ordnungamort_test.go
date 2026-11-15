package msescript_test

import (
	"testing"
)

// **`in_place(…)` ordnet an Ort und Stelle — und behält den Rest.**
//
//	/// Sort things in place, keep the rest of the input
//	void in_place_sort(const String& spec, String& input, String& ret) {
//	  String result;
//	  spec_sort(spec, input, result);
//	  // restore into the same order as in 'input'
//	  size_t pos_r = 0;
//	  FOR_EACH_CONST(c, input) {
//	    if (c == REMOVED) {
//	      if (pos_r < result.size()) ret += result.GetChar(pos_r++);
//	    } else {
//	      ret += c;
//	    }
//	  }
//	  input.clear(); // we ate all the input
//	}
//
// `src/util/spec_sort.cpp`, Zeilen 244–261, gerufen in Zeile 315–318, Fassung
// f9d9356 vom 2021-07-06.
//
// Drei Dinge stehen darin, und alle drei zählen:
//
//  1. Die Zeichen, die die Unterordnung nennt, werden **untereinander**
//     sortiert.
//  2. Sie bleiben an ihren **Stellen** — die Lücken werden der Reihe nach
//     gefüllt.
//  3. Alles andere **bleibt**, anders als beim gewöhnlichen `sort_text`
//     (#411),
//     und die Eingabe ist danach ganz verbraucht.
//
// # Woran es aufgefallen ist
//
// `mana_sort_guild` schreibt `sort_text(order:"in_place((WUBRG))")`
// (`magic.mse-game/script:521`). Dieses Programm meldete das Schlüsselwort als
// ungebaut — mit dem Grund „kein Paket von M15-Magic-Pack führt es". Das
// **volle** Pack führt es, und es trifft den **Regeltext**: die fünf
// schlechtesten Karten des Vergleichs gegen MSE2 scheiterten alle daran
// (#415).
func TestInPlaceOrdnetAnOrtUndStelle(t *testing.T) {
	// **Jeder erwartete Wert ist bei MSE2 selbst erfragt**, nicht aus dem
	// C++ erschlossen: `magicseteditor --cli SET -q --raw`, gebaut nach
	// [0026](../../docs/entscheidungen/0026-mse2-rendert-die-referenz-selbst.md).
	// Der erste Anlauf erwartete für `G/W` ein `W/G` — geraten und falsch.
	faelle := map[string]struct{ ordnung, eingabe, erwartet string }{
		// **Der Zyklus dreht hier nichts**: `G` vor `W` ist der kürzere Weg.
		"Hybridpaar bleibt": {
			ordnung: "in_place((WUBRG))", eingabe: "G/W", erwartet: "G/W",
		},
		// **Und hier doch** — `U` und `W` tauschen, der Rest bleibt liegen.
		"Paar wird gedreht, Fremdes bleibt": {
			ordnung: "in_place((WUBRG))", eingabe: "U/W/x", erwartet: "W/U/x",
		},
		// **Der Rest bleibt**, anders als beim gewöhnlichen `sort_text`.
		"Unbekanntes bleibt stehen": {
			ordnung: "in_place((WUBRG))", eingabe: "x", erwartet: "x",
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

// **Ein `in_place(`, das nicht schließt, ist ein Fehler** — keine Ordnung, die
// plausibel aussieht und falsch ist.
func TestEinOffenesInPlaceWirdGemeldet(t *testing.T) {
	// **Ohne schließende Klammer am Ende.** `in_place((WUBRG)` endete auf `)`
	// und lief in die Unterordnung — der erste Anlauf traf diesen Zweig nicht.
	if _, err := ordne(t, "in_place((WUBRG)x", "WU"); err == nil {
		t.Fatal("kein Fehler")
	}
}

// **Und eine kaputte Unterordnung wird durchgereicht**, statt still zu einer
// leeren zu werden.
func TestEineKaputteUnterordnungWirdGemeldet(t *testing.T) {
	if _, err := ordne(t, "in_place(mixed(WU))", "WU"); err == nil {
		t.Fatal("kein Fehler")
	}
}
