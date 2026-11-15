package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// `number_of_items` — dieselbe Funktion wie `length`, unter anderem Namen und
// mit einem anderen Argument.
//
// # Herkunft
//
// `src/script/functions/basic.cpp:568` von twanvl/MagicSetEditor2, Stand
// 2026-08-16:
//
//	SCRIPT_FUNCTION(number_of_items) {
//	  SCRIPT_PARAM_C(ScriptValueP, in);
//	  SCRIPT_RETURN(script_length_of(ctx, in));
//	}
//
// **Sie liest `in:`, nicht `input`** — das ist der einzige Unterschied zu
// `length`, und er ist der Grund, warum ein Aufruf ohne dieses Wissen
// scheitert. In `magic.mse-game` steht sie als
// `number_of_items(in: colors)`.
//
// Die Registrierung in MSE trägt den Kommentar `// deprecated`; im Pack steht
// sie trotzdem, und ein Paket von 2007 kümmert das nicht.
//
// # Was daran hing
//
// `color_background` in `magic-blends.mse-include/new-blends` zählt damit die
// Farben einer Karte:
//
//	color_count := if colors == "c" then 0 else number_of_items(in: colors)
//
// Ohne sie fiel `card color → image` aus — das Feld, das den **Rahmen** einer
// Karte zeichnet
// (#240).
//
// Fassung f9d9356 vom 2021-07-06.
func TestNumberOfItemsZaehltUeberIn(t *testing.T) {
	faelle := []struct {
		name    string
		quelle  string
		erwartt float64
	}{
		// So steht es im Pack: eine Zeichenkette aus Farbkürzeln.
		{"eine Zeichenkette zählt Zeichen", `number_of_items(in: "wu")`, 2},
		{"eine leere Zeichenkette", `number_of_items(in: "")`, 0},

		// Eine Liste ist in dieser Sprache eine Abbildung.
		{"eine Liste zählt Einträge", `number_of_items(in: [a: 1, b: 2, c: 3])`, 3},

		// **`nil` ist der leere Text** und damit null — dieselbe Umwandlung wie
		// überall (`umwandlung.go`).
		{"nichts ist null", `number_of_items(in: nil)`, 0},

		// **In Zeichen, nicht in Bytes** — wie `length`.
		{"mehrbytige Zeichen", `number_of_items(in: "äöü")`, 3},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.Zahl != fall.erwartt {
				t.Errorf("%v, erwartet war %v", wert.Zahl, fall.erwartt)
			}
		})
	}
}

// **Ohne `in:` ist es ein Tippfehler.**
//
// Die Gegenprobe: `length` nimmt sein Argument in der Stellung, diese Funktion
// benannt. Wer beides gleich behandelte, ließe `number_of_items(colors)`
// durchgehen — eine Schreibweise, die MSE nicht kennt.
func TestNumberOfItemsOhneInWirdAbgewiesen(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(),
		`number_of_items("wu")`, msescript.Bibliothek())
	if err == nil {
		t.Fatal("kein Fehler")
	}
}
