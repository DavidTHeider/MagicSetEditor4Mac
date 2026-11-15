// Die Prüfungen über die **Vorbelegung** einer Umgebung.
//
// Von `auswerter_test.go` abgetrennt, weil die Datei über die Zeilengrenze aus
// AGENTS.md §2 gewachsen war. Der Schnitt folgt der Frage: dort steht, was ein
// Ausdruck ergibt, hier, was danach in der Umgebung steht.
package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`Vorbelegen` legt die Namen in der übergebenen Umgebung an, nicht darüber.**
//
// # Warum das eine eigene Funktion ist und keine Fahne
//
// [Auswerten] tut absichtlich das Gegenteil: es rechnet in einer Kindebene, und
// der Grund steht an [Umgebung] — zwei Felder, die dieselbe Vorlage benutzen,
// dürfen einander nicht beeinflussen.
//
// Ein `init script:` ist der eine Fall, in dem genau das gewollt ist: es **ist**
// die Vorbelegung. `magic-m15.mse-style` legt dort `template`, `mana_sort` und
// ein Dutzend weitere an, und jedes Feldskript des Stils benutzt sie
// (#186).
//
// Zwei Namen für zwei Absichten sind hier besser als ein Name mit einer Fahne:
// wer `Auswerten(…, true)` liest, muss nachschlagen, was `true` bedeutet.
func TestVorbelegenLegtDieNamenInDerUmgebungAn(t *testing.T) {
	umgebung := msescript.Bibliothek().Darueber()

	if err := msescript.Vorbelegen(t.Context(),
		"grundwert := 3\nverdoppelt := { grundwert * 2 }", umgebung); err != nil {
		t.Fatalf("vorbelegen: %v", err)
	}

	// **Ein späterer, eigener Lauf sieht sie.** Genau das kann `Auswerten`
	// nicht, und genau darauf kommt es an.
	wert, err := msescript.AuswertenQuelle(t.Context(), "verdoppelt()", umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.Zahl != 6 {
		t.Errorf("%v, erwartet 6", wert)
	}
}

// **Ein kaputtes Vorbelegen meldet sich und legt nichts an.**
func TestEinKaputtesVorbelegenMeldetSich(t *testing.T) {
	umgebung := msescript.Bibliothek().Darueber()

	if err := msescript.Vorbelegen(t.Context(), "a := (", umgebung); err == nil {
		t.Error("kein Fehler")
	}
}
