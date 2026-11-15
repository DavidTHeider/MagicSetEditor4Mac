package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Die drei Textfunktionen, die die Exportvorlagen rufen.
func TestDieErgaenztenTextfunktionen(t *testing.T) {
	for _, fall := range []struct{ name, quelle, erwartet string }{
		{"to_string macht aus einer Zahl Text", `to_string(12)`, "12"},
		{"to_string lässt Text Text", `to_string("a" + "b")`, "ab"},
		{"to_upper schreibt groß", `to_upper("prb")`, "PRB"},
		{"trim schneidet beidseitig", `trim("  Kreatur  ")`, "Kreatur"},
		{"substring mit end", `substring("Illustrator", end: 5)`, "Illus"},
		{"substring mit begin", `substring("Illustrator", begin: 5)`, "trator"},
		{"substring mit beidem", `substring("Illustrator", begin: 1, end: 4)`, "llu"},
		{"substring ohne beides", `substring("kurz")`, "kurz"},

		// **Über die Grenze hinaus wird eingefangen.** In den Vorlagen steht
		// `end: 7`, um lange Werte zu kürzen — ein kurzer Wert ist dort kein
		// Fehler, sondern der Normalfall.
		{"substring über das Ende hinaus", `substring("ab", end: 7)`, "ab"},
		{"substring mit verdrehten Grenzen", `substring("abc", begin: 2, end: 1)`, ""},
		{"substring vor dem Anfang", `substring("abc", begin: 0 - 5)`, "abc"},

		// **In Runen gezählt.** Ein Umlaut wiegt zwei Bytes; wer in Bytes
		// schnitte, zerbräche ihn.
		{"substring zählt Runen", `substring("Öl", end: 1)`, "Ö"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Jede von ihnen verlangt genau ein Stellungsargument.**
func TestDieErgaenztenTextfunktionenBrauchenIhrArgument(t *testing.T) {
	for _, quelle := range []string{`to_string()`, `trim()`, `substring()`, `to_upper()`} {
		t.Run(quelle, func(t *testing.T) {
			if _, err := msescript.AuswertenQuelle(t.Context(), quelle,
				msescript.Bibliothek()); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}
