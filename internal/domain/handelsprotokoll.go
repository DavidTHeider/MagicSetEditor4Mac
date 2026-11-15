package domain

// **Wie ein Handel dorthin kam, wo er steht** — TRADE-03.AK-03
// (#1717).
//
// **Eine Zeile je Zustandswechsel**, geschrieben vom Bestand selbst, auf
// jedem Weg, der eine Revision schreibt. Wer schrieb, steht hier nicht: das
// sagt die Stimme in der Revision und die Unterschrift im Beleg — eine
// dritte Angabe daneben wäre eine, die ihnen widersprechen könnte.

import "time"

// Handelsprotokollzeile ist ein Wechsel des Zustands.
type Handelsprotokollzeile struct {
	Revision int

	// Von ist leer, wenn die Revision neu ist.
	Von  Handelszustand
	Nach Handelszustand
	Zeit time.Time
}
