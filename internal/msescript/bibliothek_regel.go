package msescript

import "context"

// regelfassungen sind die `_rule`-Namen und die Funktionen, die sie wickeln.
//
// # Was eine Regelfassung ist
//
// MSE legt zu mehreren Funktionen eine zweite Fassung an:
//
//	ctx.setVariable(_("replace"),      script_replace_text);
//	ctx.setVariable(_("replace_rule"), make_intrusive<ScriptRule>(script_replace_text));
//
// `src/script/functions/regex.cpp`, Zeilen 267–271; die übrigen stehen in
// `src/script/functions/basic.cpp`, Zeilen 768–791. Was die Endung bedeutet,
// ist eine Zeile:
//
//	ScriptValueP ScriptRule::eval(Context& ctx, bool openScope) const {
//	  return ctx.makeClosure(fun);
//	}
//
// `src/script/value.cpp`, Zeilen 620–622. **Wer den Namen auswertet, bekommt
// die Funktion mit den gerade gebundenen Namen** — `Context::makeClosure`
// nimmt die Ebene mit, die an einer Rufstelle gerade offen ist
// (`src/script/context.cpp`, Zeilen 318–333). Das sind die benannten
// Argumente und, wenn eines dasteht, das Stellungsargument als `input`.
//
// **`x_rule(…)` ist damit `x@(…)`** — am Orakel bestätigt und nicht
// erschlossen (§15):
//
//	to_string(replace_rule(match:"a", replace:"b")("aaa"))   → "bbb"
//	to_string(replace@(match:"a", replace:"b")("aaa"))       → "bbb"
//
// Alles Fassung f9d9356 vom 2021-07-06.
//
// # Warum sie fehlten, und was das kostete
//
// `planechase.mse-game` schreibt in Zeile 202 `type_over_type :=
// replace_rule(…)` und in Zeile 203 `plane_type_filter := tag_remove_rule(…)`.
// Ohne die Namen scheitern beide Zuweisungen, und der Packlauf meldete
// **23-mal**, einer der vier so gebundenen Namen sei unbekannt
// (#603).
//
// # `format_rule` kam mit `format`
//
// MSE führt es als neunte (`basic.cpp`, Zeile 769). Es wickelt `format`, und
// bis zum 2026-08-22 gab es `format` hier nicht — eine Regelfassung ohne ihre
// Grundfunktion wäre ein Name gewesen, der eine Funktion zurückgibt, die
// niemand rufen kann. Beide kamen zusammen
// (#843).
var regelfassungen = map[string]string{
	"replace_rule":         "replace",
	"filter_rule":          "filter_text",
	"break_rule":           "break_text",
	"match_rule":           "match",
	"sort_rule":            "sort_text",
	"tag_contents_rule":    "tag_contents",
	"tag_remove_rule":      "remove_tag",
	"expand_keywords_rule": "expand_keywords",
	"format_rule":          "format",
}

// regelfassungenIn legt die `_rule`-Namen in eine Umgebung.
//
// **Nicht über [eintragen], sondern hier** — eine Regelfassung braucht ihre
// Grundfunktion, und die steht erst fest, wenn alle Bibliotheksdateien ihre
// `init` gefahren haben. [Bibliothek] ruft danach.
//
// **Dass jede Grundfunktion dasteht, hält ein Wächter**
// (`TestJedeRegelfassungHatIhreGrundfunktion`); deshalb steht hier keine
// Prüfung, die im Betrieb nie zuschlüge.
func regelfassungenIn(umgebung *Umgebung) {
	for regel, grund := range regelfassungen {
		umgebung.Setze(regel, Wert{Art: ArtFunktion, Funktion: &Funktionsteil{
			Eingebaut: regelbinder(eingebaute[grund]),
		}})
	}
}

// regelbinder gibt die eingebaute Funktion, die eine Regelfassung ausmacht:
// sie ruft ihre Grundfunktion **nicht**, sondern bindet die Argumente an sie.
func regelbinder(grund Eingebaut) Eingebaut {
	return func(_ context.Context, ruf Ruf) (Wert, error) {
		// **Dieselbe Rechnung wie `@`**, und zwar buchstäblich dieselbe:
		// [gebundeneArgumente]. Ein zweites Mal hingeschrieben, liefen die
		// beiden auseinander.
		return Wert{Art: ArtFunktion, Funktion: &Funktionsteil{
			Eingebaut: grund,
			Gebunden:  gebundeneArgumente(ruf),
		}}, nil
	}
}

// Regelfassungen gibt die `_rule`-Namen und ihre Grundfunktionen.
//
// **Für den Wächter daneben**, nicht für die Auswertung: er hält die Zusage,
// dass jede Regelfassung eine Grundfunktion hat.
func Regelfassungen() map[string]string { return regelfassungen }
