package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Steht ein Name zweimal in derselben Abbildung, gewinnt der letzte**
// (#664).
//
//	void Context::makeObject(size_t n) {
//	  ScriptCustomCollectionP ret(new ScriptCustomCollection());
//	  …
//	    if (key != script_nil) { // valid key
//	      ret->key_value[key->toString()] = val;
//
// `src/script/context.cpp`, Zeilen 565–578, Fassung f9d9356 vom 2021-07-06.
// **Zuweisung in eine `std::map`** — der spätere Eintrag überschreibt den
// früheren, und gezählt wird die Abbildung danach.
//
// # Woran es aufgefallen ist
//
// `magic-m15-altered.mse-style` erklärt `template_prefix` mit **zwei**
// `draft:`-Einträgen — Zeile 85 zeigt auf `trims/conspiracy/`, das es im Pack
// nicht gibt, Zeile 114 auf `trims/draft/`, das es gibt. Über den Kanon waren
// das **40 Befunde** an 40 Karten mit `frames: conspiracy` oder
// `draft matters`.
func TestDoppelterNameInEinerAbbildung(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		quelle   string
		erwartet string
	}{
		{
			name:     "der letzte Wert gilt",
			quelle:   `[a: "erste", b: "x", a: "zweite"].a`,
			erwartet: "zweite",
		},
		// **Und der Platz des ersten bleibt.** MSE ordnet seine Abbildung nach
		// dem Schlüssel; hier steht die Reihenfolge der Datei, und die ist
		// Teil des Rundlaufs (`docs/FORMATE.md`). Ein Eintrag, der beim
		// Überschreiben ans Ende rutschte, änderte sie.
		{
			name:     "die Stelle des ersten bleibt",
			quelle:   `for each w in [a: "erste", b: "x", a: "zweite"] do w + ";"`,
			erwartet: "zweite;x;",
		},
		{
			name:     "gezählt wird die Abbildung danach",
			quelle:   `length([a: "erste", b: "x", a: "zweite"])`,
			erwartet: "2",
		},
		// **Unbenannte Glieder sind keine Namen.** Eine Liste ist eine
		// Abbildung ohne Schlüssel; sie darf sich nicht zusammenziehen.
		{
			name:     "eine Liste bleibt vollständig",
			quelle:   `length(["a", "a", "a"])`,
			erwartet: "3",
		},
		{
			name:     "und behält ihre Stellen",
			quelle:   `["a", "b", "c"].1`,
			erwartet: "b",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(
				t.Context(), fall.quelle, msescript.Bibliothek().Darueber())
			if err != nil {
				t.Fatalf("unerwarteter Fehler: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("Ergebnis = %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}
