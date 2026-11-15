// **Die Wertspalte eines Knotens, dessen Zeile kürzer ist als seine Stelle**
// (#1324).
//
// # Warum es diesen Lauf gibt
//
// `wertspalte` schneidet die rohe Zeile ab ihrer Spalte an. Beim Lesen aus
// einer Datei kann das nicht danebengehen — die Stelle kommt aus derselben
// Zeile. **Beim Schreiben schon**: wer einen Wert ändert, streicht `Roh`, und
// [Knoten.Zeile] setzt die Zeile dann neu zusammen — kürzer, als sie war.
//
// Ohne den Wächter darin liefe das in einen Zugriff daneben. Der Zweig ist
// zwei Zeilen und sah aus wie eine Formalie; ein Zweig, den kein Lauf
// begeht, ist eine Behauptung darüber, was im Ernstfall geschieht (§3).
//
// # Warum ein interner Lauf
//
// `wertspalte` ist paketintern, und der Zustand — eine Stelle jenseits der
// Zeile — entsteht nur, wenn jemand einen Knoten von Hand setzt. Von außen
// wäre er nicht herzustellen.
package msepaket

import "testing"

func TestDieWertspalteHaeltEineZuKurzeZeileAus(t *testing.T) {
	knoten := &Knoten{
		Zeilenart:  ArtEintrag,
		Schluessel: "script",
		Wert:       "1 + 2",
		Stelle:     Stelle{Zeile: 1, Spalte: 40},
	}

	// **`Roh` ist leer**, also setzt `Zeile()` zusammen — und das ergibt eine
	// Zeile, die keine 40 Zeichen hat.
	if spalte := knoten.wertspalte(); spalte != 40 {
		t.Errorf("Spalte %d, erwartet war die Stelle selbst (40)", spalte)
	}
}
