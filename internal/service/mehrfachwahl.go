package service

// Die Normalform eines `multiple choice`-Wertes
// (#556).

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Wahltrenner steht zwischen zwei Wahlen einer Mehrfachwahl.
//
//	if (!val.empty()) val += _(", ");
//
// `src/data/field/multiple_choice.cpp`, Zeile 106, Fassung f9d9356 vom
// 2021-07-06. **Mit Leerzeichen**, gleich wie die Datei es geschrieben hat.
const Wahltrenner = ", "

// normalform bringt einen Mehrfachwahlwert in die Reihenfolge der Erklärung.
//
// # Die Regel, Zeile für Zeile
//
//	void MultipleChoiceValue::normalForm() {
//	  String& val = value.mutateDontChangeDefault();
//	  // which choices are active?
//	  vector<bool> seen(field().choices->lastId());
//	  for (size_t pos = 0 ; pos < val.size() ; ) {
//	    if (val.GetChar(pos) == _(' ')) {
//	      ++pos; // ingore whitespace
//	    } else {
//	      size_t end = val.find_first_of(_(','), pos);
//	      if (end == String::npos) end = val.size();
//	      for (size_t i = 0 ; i < seen.size() ; ++i) {
//	        if (is_substr(val, pos, field().choices->choiceName((int)i))) {
//	          seen[i] = true;
//	          break;
//	        }
//	      }
//	      pos = end + 1;
//	    }
//	  }
//	  // now put them back in the right order
//	  val.clear();
//	  for (size_t i = 0 ; i < seen.size() ; ++i) {
//	    if (seen[i]) {
//	      if (!val.empty()) val += _(", ");
//	      val += field().choices->choiceName((int)i);
//	    }
//	  }
//	}
//
// `src/data/field/multiple_choice.cpp`, Zeilen 82–110, gerufen aus
// `MultipleChoiceValue::update`, Zeile 59. Fassung f9d9356 vom 2021-07-06.
//
// # Drei Eigenheiten, die man ihr ansieht
//
//   - **Die erste passende Wahl gewinnt**, und verglichen wird als Präfix
//     (`is_substr` ist `starts_with`, `src/util/string.hpp`, Zeilen 289–291).
//     Hießen zwei Wahlen `u` und `ult`, träfe `ult` die erste. Nachgebildet
//     und nicht behoben.
//   - **Was in keiner Liste steht, fällt weg.** Ein Tippfehler im Wert
//     verschwindet, statt durchgereicht zu werden.
//   - **Die Reihenfolge im Wert zählt nicht**, nur die der Erklärung.
//
// # Wozu es gebraucht wird
//
// `magic.mse-game/script`, Zeile 3605, baut aus dem Wert einen Dateinamen:
//
//	shape := remove_comma(loyalty_cost_box_field(input + (instance-1)*8))
//	name  := "loyalty cost " + shape + ".png"
//
// `remove_comma` streicht nur Kommata. Ohne Normalform suchte dieses Programm
// `loyalty cost shadow 0.png`, wo `loyalty cost 0 shadow.png` liegt.
func normalform(wahlen []string, wert, leerwahl string) string {
	gesehen := make([]bool, len(wahlen))

	for stelle := 0; stelle < len(wert); {
		if wert[stelle] == ' ' {
			stelle++

			continue
		}

		ende := strings.IndexByte(wert[stelle:], ',')
		if ende < 0 {
			ende = len(wert)
		} else {
			ende += stelle
		}

		for i, wahl := range wahlen {
			if wahl != "" && strings.HasPrefix(wert[stelle:], wahl) {
				gesehen[i] = true

				break
			}
		}

		stelle = ende + 1
	}

	teile := make([]string, 0, len(wahlen))

	for i, wahl := range wahlen {
		if gesehen[i] {
			teile = append(teile, wahl)
		}
	}

	if len(teile) == 0 {
		return leerwahl
	}

	return strings.Join(teile, Wahltrenner)
}

// normalisiereMehrfachwahlen bringt jeden Mehrfachwahlwert in Normalform.
//
// **Nach jedem Durchgang und nicht einmal am Anfang.** MSE ruft `normalForm`
// in `MultipleChoiceValue::update` — also nach dem Vorgabe- **und** nach dem
// Wertskript (`multiple_choice.cpp`, Zeilen 54–60, Fassung f9d9356 vom
// 2021-07-06). Ein Skript, das `"shadow, 0"` zurückgibt, wird damit ebenso
// geordnet wie ein Wert aus der Datei.
func normalisiereMehrfachwahlen(
	felder []msescript.Eintrag, kartenfelder []domain.Field,
) []msescript.Eintrag {
	return mehrfachwahlenOrdnen(felder, kartenfelder, true)
}

// ordneMehrfachwahlen bringt jede Mehrfachwahl in Normalform, **ohne** einen
// Leernamen zu schreiben — für den Durchgang vor den Vorgabeskripten.
//
// # Warum
//
// **Weil eine leere Wahl dort noch Vorgabe ist.** MSE2 schreibt den Leernamen
// über `mutateDontChangeDefault`: der Wert bleibt Vorgabe, und das
// Vorgabeskript rechnet ihn beim nächsten Durchgang neu
// (`MultipleChoiceValue::normalForm`, `src/data/field/multiple_choice.cpp`,
// Zeilen 110–113, Fassung f9d9356 vom 2021-07-06). Hier ist eine Vorgabe
// daran zu erkennen, dass ihr Wert leer ist ([mitVorgabeskripten]) — stand
// davor schon `colorless`, rechnete niemand `card color` aus, und jede
// Karte des Kanons war farblos (Kanonlauf vom 2026-10-04, mittlerer Abstand
// zu MSE2 von 9 auf 31).
func ordneMehrfachwahlen(
	felder []msescript.Eintrag, kartenfelder []domain.Field,
) []msescript.Eintrag {
	return mehrfachwahlenOrdnen(felder, kartenfelder, false)
}

// mehrfachwahlenOrdnen ist beides — mit oder ohne Leernamen.
func mehrfachwahlenOrdnen(
	felder []msescript.Eintrag, kartenfelder []domain.Field, mitLeerwahl bool,
) []msescript.Eintrag {
	for i := range kartenfelder {
		feld := &kartenfelder[i]
		if feld.Kind != domain.KindMultipleChoice || len(feld.Choices) == 0 {
			continue
		}

		name := strings.ReplaceAll(feld.Name, " ", "_")

		for j := range felder {
			if felder[j].Name != name || felder[j].Wert.Art != msescript.ArtTextwert {
				continue
			}

			// **Auch ein Kartenfeld hat einen Leernamen** (#2006): `normalForm`
			// schreibt `empty_choice`, wo nichts gewählt ist, und
			// `magic.mse-game` erklärt `colorless` an sieben Feldern. Hier
			// stand, `empty choice:` gebe es nur an einem `styling field:`.
			leerwahl := ""
			if mitLeerwahl {
				leerwahl = feld.Leerwahl
			}

			felder[j].Wert = msescript.Text(
				normalform(feld.Choices, felder[j].Wert.Text, leerwahl))
		}
	}

	return felder
}
