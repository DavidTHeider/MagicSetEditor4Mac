package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`to_text` nimmt die Auszeichnung heraus** —
// #2006.
//
//	SCRIPT_FUNCTION(to_text) {
//	  SCRIPT_PARAM_C(String, input);
//	  SCRIPT_RETURN(untag_hide_sep(input));
//	}
//
// `src/script/functions/export.cpp`, Zeilen 361–364, Fassung f9d9356 vom
// 2021-07-06. Gemessen am Leistungsvergleich: `magic-mws` schließt mit
// `to_text(…)`, und MSE2 schrieb den Flavor-Text ohne `<i-flavor>`, diese
// Ausfuhr mit.
func TestZuTextNimmtDieAuszeichnungHeraus(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name, quelle, erwartet string
	}{
		{"Marken fallen weg, ihr Inhalt bleibt", `to_text("<i-flavor>Glut</i-flavor>")`, "Glut"},
		{
			"weiche Trenner fallen samt Inhalt weg",
			`to_text("<word-list-type-en>Scheme</word-list-type-en><sep-soft> — </sep-soft>")`, "Scheme",
		},
		{"das geschützte Kleinerzeichen wird wieder eines", "to_text(\"a\x01b\")", "a<b"},
		{"eine offene Klammer schluckt den Rest, wie bei MSE", `to_text("a<b")`, "a"},
		{"eine Zahl bleibt ihre Textform", `to_text(3)`, "3"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%s ergab %q, erwartet %q", fall.quelle, wert.AlsText(), fall.erwartet)
			}
		})
	}
}
