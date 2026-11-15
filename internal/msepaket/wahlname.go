package msepaket

// Der Name einer `choice:` — einzeilig oder als Block
// (#556).

// wahlname gibt den Namen einer Wahl.
//
// # Zwei Schreibweisen, und MSE unterscheidet sie ausdrücklich
//
//	IMPLEMENT_REFLECTION(ChoiceField::Choice) {
//	  REFLECT_IF_READING_SINGLE_VALUE {
//	    REFLECT_NAMELESS(name);
//	  } else {
//	    // complex values are groups
//	    REFLECT(name);
//	    …
//
// `src/data/field/choice.cpp`, Zeilen 145–157, Fassung f9d9356 vom
// 2021-07-06. Steht hinter `choice:` ein Wert, **ist** er der Name; steht dort
// nichts, nennt ihn ein `name:` darunter.
//
// # Wie oft die Blockform vorkommt
//
// Über `full-magic-pack` sind **691** von 8.155 Wahlen als Block geschrieben.
// In `/magic-modules.mse-include/loyalty/choices_cost` ist **jede** der
// sieben Wahlen einer — und an der Reihenfolge der Liste hängt die
// Normalisierung eines `multiple choice`-Werts.
//
// # Was hier nicht steht
//
// **Verschachtelte Gruppen.** MSE erlaubt einer Wahl eigene `choices:`, und
// `choiceName` gibt dann den Pfad (`gruppe/kind`). Im Pack kommt das in keiner
// `multiple choice` vor; wer es hinzunimmt, prüft das nach.
func wahlname(moeglich *Knoten) string {
	if moeglich.Wert != "" {
		return moeglich.Wert
	}

	if name := moeglich.Ersten("name"); name != nil {
		return name.Wert
	}

	return ""
}
