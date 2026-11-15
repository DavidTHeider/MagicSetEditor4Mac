package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Eine Zeichenkette darf Ausdrücke enthalten — `"a{e}b"`.**
//
// # Die Regel
//
// MSEs Zerteiler zerlegt eine solche Kette in drei Marken und setzt sie mit
// `+` wieder zusammen:
//
//	// for smart strings: "x" {{ e }} "y"
//	"a{e}b"  -->  "a"  "{  e  }"  "b"
//
// `src/script/parser.cpp`, Zeilen 302–307 und 842–860, Fassung f9d9356 vom
// 2021-07-06. Aus `"a{e}b"` wird also `"a" + e + "b"`.
//
// **`magic.mse-game` baut damit Auszeichnungen zusammen** —
// `"<{tag}>{input}</{tag}>"` und `"<nospellcheck><key>{keyword}</key></nospellcheck>"`.
// Ohne die Zerlegung stand die geschweifte Klammer wörtlich auf der Karte
// (#345).
func TestEineZeichenketteDarfAusdrueckeEnthalten(t *testing.T) {
	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("tag", msescript.Text("i"))
	umgebung.Setze("input", msescript.Text("Text"))

	faelle := []struct{ name, quelle, erwartet string }{
		{"ein Name mittendrin", `"a{tag}b"`, "aib"},
		{"eine Rechnung", `"a{1 + 2}b"`, "a3b"},
		{"zwei Stellen", `"<{tag}>{input}</{tag}>"`, "<i>Text</i>"},
		{"am Anfang", `"{tag}b"`, "ib"},
		{"am Ende", `"a{tag}"`, "ai"},
		{"nur der Ausdruck", `"{tag}"`, "i"},
		{"ohne Ausdruck bleibt alles", `"ab"`, "ab"},

		// **Eine maskierte Klammer bleibt eine.**
		{"eine maskierte Klammer", `"a\{b}c"`, "a{b}c"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, umgebung)
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Ein Ausdruck mit einer eigenen Zeichenkette darin wird gelesen**
// (#468).
//
// Hier stand bis zum 2026-08-18 der Wächter über die Gegenrichtung: die
// verschachtelte Zeichenkette beendete die äußere, und der Grund stand
// daneben statt in einem stillen Fehlschlag. Er ist rot geworden, als die
// Grenze fiel — das war seine Aufgabe.
//
// **Was ihn hat fallen lassen**: MSE2 hält einen Klammerstapel, und ein `"`
// in einer Einbettung legt einen **zweiten** `BRACE_STRING` darauf
// (`src/script/parser.cpp`, Zeilen 268–272, Fassung f9d9356 vom
// 2021-07-06). Die Zusagen über die Fluchtzeichen von sed bleiben davon
// unberührt: `TestEinErsatztextKenntDieFluchtzeichenVonSed` und
// `TestWasSichNichtZerteilenLaesstIstText` prüfen sie weiter.
func TestEineZeichenketteImAusdruckWirdGelesen(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		// Die Form, an der es aufgefallen ist:
		// `magic-m15-split-fusable.mse-style/style`, Zeile 555 des
		// `init script:`.
		{
			"zwei Zweige, einer davon leer",
			`"borders/room{if 1 < 2 then "_cfuse" else ""}.png"`,
			"borders/room_cfuse.png",
		},
		{
			"der andere Zweig",
			`"borders/room{if 2 < 1 then "_cfuse" else ""}.png"`,
			"borders/room.png",
		},

		// Die Form aus `magic.mse-game/script` — eine Auszeichnung, deren
		// Name selbst aus einem Ausdruck kommt.
		{
			"eine Auszeichnung um einen Wert",
			`"<{if 1 < 2 then "i" else "b"}>"`,
			"<i>",
		},

		// **Eine geschweifte Klammer in der inneren Zeichenkette schließt die
		// Einbettung nicht.** Ohne diese Buchführung fände die Suche nach der
		// schließenden Klammer die falsche.
		{"eine Klammer in der inneren Kette", `"a{"}" + "b"}c"`, "a}bc"},

		// **Zwei Einbettungen nacheinander**, damit die Zählung nach der
		// ersten wieder bei null steht.
		{"zwei Einbettungen", `"{"a"}-{"b"}"`, "a-b"},

		// **Ein geflohenes Anführungszeichen in der inneren Zeichenkette**
		// beendet sie nicht — sonst fände die Suche nach der schließenden
		// Klammer die falsche.
		{"ein Anführungszeichen in der inneren Kette", `"a{"x\"y"}b"`, `ax"yb`},

		// **Eine geschützte Klammer in der Einbettung schließt sie nicht.**
		// `\{` ist der gewöhnliche Weg, eine geschweifte Klammer in einen
		// Text zu setzen, und er gilt auch hier.
		{"eine geschützte Klammer in der Einbettung", `"{"a\{b"}"`, "a{b"},

		// **Ein Anführungszeichen ohne Partner in der Einbettung** macht sie
		// zu Text: die schließende Klammer ist nicht mehr zu finden, und was
		// dasteht, bleibt stehen (§8).
		{"ein Anführungszeichen ohne Partner", `"a{b\"c}d"`, `a{b"c}d`},

		// **Eine geschützte Klammer zählt auch außerhalb einer inneren
		// Zeichenkette nicht mit.** `\{` steht hier zwischen den Klammern
		// der Einbettung und nicht in einer Zeichenkette darin; was übrig
		// bleibt, ist kein Ausdruck und damit Text.
		{"eine geschützte Klammer zwischen den Klammern", `"{a\{b}"`, "{a{b}"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(
				t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Was sich nicht zerteilen lässt, ist Text.**
//
// Eine geschweifte Klammer steht auch dort, wo niemand einen Ausdruck meint:
// `replace(…, replace: "\x{20AC}")` schreibt eine Ersatzzeichenfolge im
// sed-Stil, und `{20AC}` ist darin eine Angabe für den Mustervergleich. MSE
// unterscheidet die beiden gar nicht; hier fällt es auf Text zurück, und diese
// Abweichung steht im Code
// (#345).
func TestWasSichNichtZerteilenLaesstIstText(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		// `20AC` ist keine gültige Zahl und kein Name.
		{"kein Ausdruck", `"a{20AC}b"`, "a{20AC}b"},

		// **Eine Fluchtfolge davor schützt die Klammer** — auch dann, wenn ihr
		// Inhalt für sich genommen ein Ausdruck wäre.
		{"nach einer Musterflucht", `"\\x{110000}"`, `\x{110000}`},
		// **`\{` schützt die Klammer selbst** und fällt dabei weg — der
		// gewöhnliche Weg, eine geschweifte Klammer in einen Text zu setzen.
		{"eine geschützte Klammer", `"\\{1}"`, `{1}`},
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

// **Eine Klammer ohne Ende bleibt stehen.**
//
// Ein Kartentext darf `{` enthalten; die Zeichenkette endet weiter am
// Anführungszeichen, und was danach ohne Partner dasteht, ist Text.
func TestEineKlammerOhneEndeBleibtStehen(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(
		t.Context(), `"a{b"`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "a{b" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "a{b")
	}
}

// **`\}` ist eine geschweifte Klammer im Text — wie `\{`.**
//
// # Was am 2026-08-31 nachgelesen wurde (§15)
//
// `TokenIterator::readToken` im Zeichenkettenmodus (`src/script/parser.cpp`,
// Zeile 287, Fassung f9d9356 vom 2021-07-06):
//
//	else if (c == '\\' || c == '"' || c == '\'' || c == '{' || c == '}') {
//	  str += c; // don't warn about these, since they look escape-worthy
//
// **Beide Klammern, nicht nur die öffnende.** Hier stand nur `\{`, und ein
// Textskript wie `\{cmc({input})\}` — genau die Form, in der
// `magic.mse-game` seine Parameterverweise schreibt — behielt sein `\`:
// heraus kam `{cmc(param1)\}`.
//
// **Gemessen am 2026-08-31** über die Datendateien des Packs (die ohne
// Endung): 337-mal `\{` und **253-mal `\}`** in 17 Dateien.
func TestEineMaskierteSchliessendeKlammerIstText(t *testing.T) {
	t.Parallel()

	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("input", msescript.Text("param1"))

	wert, fehler := msescript.AuswertenTextskript(
		t.Context(), `\{cmc({input})\}`, umgebung)
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if wert.AlsText() != "{cmc(param1)}" {
		t.Errorf("heraus kam %q", wert.AlsText())
	}
}
