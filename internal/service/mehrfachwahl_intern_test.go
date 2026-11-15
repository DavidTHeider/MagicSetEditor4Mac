package service

import "testing"

// Die Randfälle der Normalform — geprüft an der Funktion, nicht am Plan
// (#556).
//
// **Über den Plan wären sie nicht zu sehen**: ein leerer Wert ergibt gar keine
// Textebene, und eine Wahlliste ohne Namen kommt aus einem Paket, das man
// eigens bauen müsste.
func TestDieNormalformInIhrenRandfaellen(t *testing.T) {
	t.Parallel()

	wahlen := []string{"none", "anti ult", "+", "0", "-", "ult", "shadow"}

	faelle := []struct {
		name     string
		wahlen   []string
		wert     string
		erwartet string
	}{
		{"ein leerer Wert bleibt leer", wahlen, "", ""},
		{"nur Leerzeichen", wahlen, "   ", ""},
		{"nur Kommata", wahlen, ",,,", ""},
		{"ohne Wahlliste bleibt nichts", nil, "0, shadow", ""},

		// **`anti ult` trägt ein Leerzeichen** — der Trenner ist das Komma,
		// nicht der Zwischenraum.
		{"eine Wahl mit Leerzeichen", wahlen, "shadow, anti ult", "anti ult, shadow"},

		// **Doppelt genannt zählt einmal**: `seen` ist ein Wahrheitswert je
		// Wahl, kein Zähler.
		{"zweimal dieselbe Wahl", wahlen, "0, 0", "0"},

		// **Die erste passende Wahl gewinnt, und verglichen wird als Präfix.**
		// MSEs `is_substr` ist `starts_with`; mit `u` vor `ult` trifft `ult`
		// die erste. Nachgebildet und nicht behoben.
		{"ein Präfix schlägt die längere Wahl", []string{"u", "ult"}, "ult", "u"},
		{"und umgekehrt trifft es die richtige", []string{"ult", "u"}, "ult", "ult"},

		// **Eine leere Wahl trifft nie.** `strings.HasPrefix(x, "")` wäre
		// sonst immer wahr, und die erste leere Wahl fräße jeden Wert.
		{"eine leere Wahl in der Liste", []string{"", "0"}, "0", "0"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if ist := normalform(fall.wahlen, fall.wert, ""); ist != fall.erwartet {
				t.Errorf("normalform(%q, %q) = %q, erwartet %q",
					fall.wahlen, fall.wert, ist, fall.erwartet)
			}
		})
	}
}
