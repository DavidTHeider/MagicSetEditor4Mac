package service

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Anfangswert ist, was ein Feld **ohne** `default:` und ohne Kartenwert trägt
// — MSEs Stufen 2 und 3.
//
//	ChoiceValue::ChoiceValue(const ChoiceFieldP& field, bool initial_first_choice)
//	  : value( !field->initial.empty() ? field->initial
//	         : initial_first_choice    ? field->choices->choiceName(0)
//	                                   : String() )
//
// `src/data/field/choice.cpp`, Zeilen 279–283, Fassung f9d9356 vom
// 2021-07-06.
//
// **Ausgeführt, weil der Editor dieselbe Frage stellt.** Ein `select` ohne
// Wert zeigt seine erste Option; `rarity` hat sieben Wahlen, beginnend mit
// `basic land`, und `initial: common`. Der Editor behauptete damit eine
// Seltenheit, die weder in der Karte noch auf ihr steht
// (#562).
//
// **Ohne `default:`**, denn das ist ein Skript über die Karte — und wer nach
// dem Anfangswert eines Feldes fragt, hat keine.
//
// **Eine Mehrfachwahl beginnt leer**: sie ruft `ChoiceValue(field, false)`
// (`src/data/field/multiple_choice.hpp`, Zeile 52, Fassung f9d9356 vom
// 2021-07-06). Bis F02 bekam sie hier ihre erste Wahl, und `magic-nouveau`
// zeichnete auf jede Karte das Symbol für Todesberührung
// (#2244).
func Anfangswert(feld domain.Field) string {
	if anfang := strings.TrimSpace(feld.Initial); anfang != "" {
		return anfang
	}

	if len(feld.Choices) > 0 && feld.Kind != domain.KindMultipleChoice {
		return feld.Choices[0]
	}

	return ""
}
