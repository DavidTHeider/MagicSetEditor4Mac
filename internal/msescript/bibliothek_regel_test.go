package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`x_rule(…)` ist `x@(…)`**
// (#603).
//
// # Woher die Regel kommt
//
// MSE legt zu mehreren Funktionen eine zweite Fassung mit der Endung `_rule`
// an:
//
//	ctx.setVariable(_("replace"),      script_replace_text);
//	ctx.setVariable(_("replace_rule"), make_intrusive<ScriptRule>(script_replace_text));
//
// `src/script/functions/regex.cpp`, Zeilen 267–271, und
// `src/script/functions/basic.cpp`, Zeilen 768–791. Was die Endung bedeutet,
// steht in einer Zeile:
//
//	ScriptValueP ScriptRule::eval(Context& ctx, bool openScope) const {
//	  return ctx.makeClosure(fun);
//	}
//
// `src/script/value.cpp`, Zeilen 620–622. `Context::makeClosure` nimmt die
// Namen der **innersten** Ebene mit (`src/script/context.cpp`, Zeilen
// 318–333) — an einer Rufstelle sind das genau die benannten Argumente, und
// das Stellungsargument als `input`.
//
// Alles Fassung f9d9356 vom 2021-07-06.
//
// # Am Orakel bestätigt, nicht erschlossen
//
// Gegen `White.mse-set`, am 2026-08-20:
//
//	to_string(replace_rule(match:"a", replace:"b")("aaa"))   → "bbb"
//	to_string(replace@(match:"a", replace:"b")("aaa"))       → "bbb"
//	to_string(tag_remove_rule(tag:"<i")("x<i>y</i>z"))       → "xyz"
//
// # Was das kostet, wenn es fehlt
//
// `planechase.mse-game` schreibt in Zeile 202 `type_over_type :=
// replace_rule(…)` und in Zeile 203 `plane_type_filter := tag_remove_rule(…)`.
// Fehlen die Namen, scheitern die Zuweisungen — und der Packlauf meldete
// **23-mal**, einer der vier so gebundenen Namen sei unbekannt.
func TestEineRegelfassungIstEineTeilanwendung(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name   string
		regel  string
		gleich string
	}{
		{
			name:   "replace_rule",
			regel:  `replace_rule(match: "a", replace: "b")("aaa")`,
			gleich: `replace@(match: "a", replace: "b")("aaa")`,
		},
		{
			name:   "filter_rule",
			regel:  `filter_rule(match: "[0-9]")("a1b2")`,
			gleich: `filter_text@(match: "[0-9]")("a1b2")`,
		},
		{
			name:   "break_rule",
			regel:  `break_rule(match: "[0-9]")("a1b2")`,
			gleich: `break_text@(match: "[0-9]")("a1b2")`,
		},
		{
			name:   "match_rule",
			regel:  `to_string(match_rule(match: "b")("abc"))`,
			gleich: `to_string(match@(match: "b")("abc"))`,
		},
		{
			name:   "sort_rule",
			regel:  `sort_rule(order: "cba")("abc")`,
			gleich: `sort_text@(order: "cba")("abc")`,
		},
		{
			name:   "tag_contents_rule",
			regel:  `tag_contents_rule(tag: "<i", contents: "z")("a<i>y</i>b")`,
			gleich: `tag_contents@(tag: "<i", contents: "z")("a<i>y</i>b")`,
		},
		{
			name:   "tag_remove_rule",
			regel:  `tag_remove_rule(tag: "<i")("x<i>y</i>z")`,
			gleich: `remove_tag@(tag: "<i")("x<i>y</i>z")`,
		},
		{
			// **Ein Stellungsargument wird zu `input` und bleibt gebunden.**
			// MSE legt es an der Rufstelle so ab, und `makeClosure` nimmt es
			// von dort mit — der Aufruf danach braucht dann keines mehr.
			name:   "replace_rule bindet auch das Stellungsargument",
			regel:  `replace_rule("aaa", match: "a", replace: "b")()`,
			gleich: `replace(input: "aaa", match: "a", replace: "b")`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			regel := auswerten(t, fall.regel)
			gleich := auswerten(t, fall.gleich)

			if regel != gleich {
				t.Errorf("%s ergibt %q, `@` ergibt %q", fall.name, regel, gleich)
			}
		})
	}
}

// auswerten wertet eine Quelle über der Bibliothek aus.
func auswerten(t *testing.T, quelle string) string {
	t.Helper()

	wert, err := msescript.AuswertenQuelle(t.Context(), quelle, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("%s: %v", quelle, err)
	}

	return wert.AlsText()
}

// **Jede Regelfassung hat ihre Grundfunktion** — der Wächter über die Zusage,
// die `regelfassungenIn` macht.
//
// # Die Gegenprobe
//
// Ohne ihn wäre ein Tippfehler in der Tabelle stumm: `regelfassungenIn`
// bekäme `nil` aus `eingebaute`, legte einen Funktionswert ohne Rumpf und ohne
// Go-Funktion ab, und der erste Aufruf liefe in einen Fehler, dessen Wortlaut
// über den falschen Namen spräche.
//
// Am 2026-08-20 gefahren: einen Eintrag der Tabelle auf `"replace_typo"`
// gesetzt, und dieser Lauf wird rot.
func TestJedeRegelfassungHatIhreGrundfunktion(t *testing.T) {
	t.Parallel()

	bibliothek := msescript.Bibliothek()

	for regel, grund := range msescript.Regelfassungen() {
		wert, da := bibliothek.Suche(grund)
		if !da {
			t.Errorf("%s wickelt %q — den Namen gibt es nicht", regel, grund)

			continue
		}

		if wert.Art != msescript.ArtFunktion || wert.Funktion == nil ||
			wert.Funktion.Eingebaut == nil {
			t.Errorf("%s wickelt %q — das ist keine eingebaute Funktion, sondern %s",
				regel, grund, wert.Art)
		}

		if _, da := bibliothek.Suche(regel); !da {
			t.Errorf("%s steht nicht in der Bibliothek", regel)
		}
	}
}
