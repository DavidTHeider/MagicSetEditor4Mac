package render

import "strings"

// Textbreite ist die Breite eines Textzugs, **ohne ihn zu setzen**.
//
// # Wofür sie gebraucht wird
//
// `card_style.<feld>.content_width` fragt, wie breit der gesetzte Inhalt eines
// Feldes ist, und die Feldangaben eines Stils rechnen damit — **bevor** der
// Rahmen feststeht. [Setze] kann das nicht beantworten: es braucht einen
// Rahmen, um umzubrechen, und der Rahmen ist hier gerade das Gesuchte.
//
// Gemessen wird deshalb der Zug, wie er **ohne Umbruch** dastünde. Das ist
// genau die Frage, die `content_width` stellt.
//
// # Die längste Zeile, nicht die Summe
//
// Ein Text mit hartem Umbruch besteht aus mehreren Zeilen; seine Breite ist die
// der längsten. Die Summe wäre die Breite, die er hätte, wenn man ihn wieder zu
// einer Zeile zusammenzöge — das tut niemand.
//
// # Ohne Schrift null, und das ist eine Auskunft
//
// Null heißt hier **nicht gemessen**, nicht „null breit". Der Aufrufer
// unterscheidet das und setzt dann seine deklarierte Angabe ein; eine
// geschätzte Zahl sähe aus wie eine Messung und wäre keine.
//
// **Was sie nicht kann:** Auszeichnungen. `<i>` und `<sym>` zählen hier als
// gewöhnliche Zeichen und machen den Zug zu breit. Das ist Phase D, und dort
// gehört es hin — hier steht es, damit niemand die Zahl für genauer hält, als
// sie ist.
func Textbreite(text string, groesse float64, schrift Schrift) float64 {
	if schrift == nil || text == "" || groesse <= 0 {
		return 0
	}

	breiteste := 0.0

	for _, zeile := range strings.Split(text, "\n") {
		breite := 0.0
		vorher := rune(0)

		// **Bei Größe 1**, deshalb der leere Satz: diese Breite ist ein
		// Anteil und wird am Ende mit `groesse` genommen (#752).
		einheit := Satz{Groesse: 1}

		for _, zeichen := range gesetzt(zeile, einheit, schrift) {
			breite += kernungZwischen(vorher, zeichen, einheit, schrift)
			breite += schrift.Breite(zeichen)
			vorher = zeichen
		}

		if breite > breiteste {
			breiteste = breite
		}
	}

	return breiteste * groesse
}
