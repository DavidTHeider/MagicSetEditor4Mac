package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Ein Argument ohne Namen heißt `input`.**
//
// # Der Befund
//
// MSE hat kein Stellungsargument: `f(x)` setzt `x` als Kontextvariable
// `input`, und jede eingebaute Funktion liest es dort
// (`SCRIPT_PARAM_C(String, input)`, `src/script/functions/*.cpp`, Fassung
// f9d9356 vom 2021-07-06). Beide Schreibweisen meinen dasselbe.
//
// **Sichtbar wird das erst bei der Verkettung.** `text_filter` in
// `magic.mse-game` reiht rund dreißig teilangewandte Funktionen mit `+`:
//
//	text_filter :=
//		remove_tag@(tag: "<sym-auto>") +
//		remove_tag@(tag: "<i-auto>") + …
//
// `ScriptCompose` gibt das Zwischenergebnis als `input` weiter
// (`src/script/context.cpp`, Zeilen 383–400). Wer nur ein Stellungsargument
// annimmt, bekommt an dieser Naht **kein** Argument
// (#345).
func TestEinArgumentOhneNamenHeisstInput(t *testing.T) {
	faelle := []struct {
		name, quelle, erwartet string
	}{
		{"ohne Namen", `to_upper("ab")`, "AB"},
		{"als input", `to_upper(input: "ab")`, "AB"},
		{"remove_tags als input", `remove_tags(input: "<i>a</i>b")`, "ab"},
		{"replace als input", `replace(input: "aa", match: "a", replace: "b")`, "bb"},

		// **Und die Verkettung reicht es durch** — die eigentliche Naht.
		{
			name: "eine Kette aus zwei Ersetzungen",
			quelle: `(replace@(match:"a", replace:"b") + ` +
				`replace@(match:"b", replace:"c"))(input: "aa")`,
			erwartet: "cc",
		},
		{
			// **`remove_tag` nimmt auch das schließende mit** — deshalb bleibt
			// von `<i>a</i>` nur das `a`, aus dem die zweite Stufe ein `b`
			// macht.
			name: "eine Kette, ohne Namen gerufen",
			quelle: `(remove_tag@(tag: "<i>") + ` +
				`replace@(match:"a", replace:"b"))("<i>a</i>")`,
			erwartet: "b",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Ein Stellungsargument schlägt ein benanntes `input`.**
//
// Beides zugleich kommt in keiner gesichteten Vorlage vor; die Regel steht
// hier, damit sie nicht von der Reihenfolge einer Schleife abhängt.
func TestEinStellungsargumentSchlaegtDasBenannte(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(
		t.Context(), `to_upper("ab", input: "cd")`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "AB" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "AB")
	}
}
