package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// `split_text` — 176-mal in `M15-Magic-Pack` (#214).
func TestSplitTextZerlegtAnEinemMuster(t *testing.T) {
	faelle := []struct {
		name, quelle, erwartet string
	}{
		// `split_text(match:"\n", card.rule_text)` — 84-mal.
		{"an Zeilenumbrüchen", `length(split_text(match:"\n", "eins\nzwei\ndrei"))`, "3"},

		// `split_text@(match: " *, *")` — die Trennzeichen dürfen Leerraum tragen.
		{"an einem Muster mit Leerraum", `length(split_text(match:" *, *", "a ,  b,c"))`, "3"},

		// **Der leere Text ergibt ein Element, nicht keines** — `include_empty`
		// ist wahr (`src/script/functions/regex.cpp`, Zeile 215, Fassung
		// f9d9356 vom 2021-07-06). Hier stand bis zum 2026-08-17 „eine leere
		// Liste", und die Begründung war erschlossen; `TestSplitTextBehaelt-
		// LeereTeile` führt die nachgelesene Regel.
		{"der leere Text", `length(split_text(match:"\n", ""))`, "1"},

		{"ohne Treffer bleibt ein Stück", `length(split_text(match:";", "ohne"))`, "1"},
	}

	for _, fall := range faelle {
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

// **Die Stücke sind zu erreichen, nicht nur zu zählen.**
func TestSplitTextGibtDieStueckeHer(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`(for each teil in split_text(match:",", "a,b,c") do teil + "|")`,
		msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "a|b|c|" {
		t.Errorf("%q", wert.AlsText())
	}
}

// `chosen` — 300-mal, immer über einem Mehrfachfeld (#214).
func TestChosenFragtNachEinerWahl(t *testing.T) {
	faelle := []struct {
		name, quelle string
		erwartet     bool
	}{
		// `chosen(card.card_color, choice:"white")` — 44-mal je Farbe.
		{"die Wahl steht da", `chosen("white, blue", choice:"white")`, true},
		{"eine zweite Wahl", `chosen("white, blue", choice:"blue")`, true},
		{"eine Wahl, die nicht dasteht", `chosen("white, blue", choice:"red")`, false},

		// **Kein `contains`.** `white` steckt in `whiteblue`, und die Vorlagen
		// fragen nach *einer* Wahl. Wer hier `contains` nähme, färbte jede
		// zweifarbige Karte falsch.
		{"kein Teiltreffer", `chosen("whiteblue", choice:"white")`, false},

		{"leere Auswahl", `chosen("", choice:"white")`, false},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.Wahrheit != fall.erwartet {
				t.Errorf("%v, erwartet %v", wert.Wahrheit, fall.erwartet)
			}
		})
	}
}

// **Beide weisen ein fehlendes benanntes Argument ab.**
func TestSplitTextUndChosenBrauchenIhrArgument(t *testing.T) {
	for _, quelle := range []string{`split_text("a")`, `chosen("a")`} {
		t.Run(quelle, func(t *testing.T) {
			if _, err := msescript.AuswertenQuelle(t.Context(), quelle,
				msescript.Bibliothek()); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// `to_int` — neben `to_real`, dort wo aus einer Größe eine Stelle wird (#214).
func TestToIntSchneidetAb(t *testing.T) {
	for _, fall := range []struct{ name, quelle, erwartet string }{
		{"aus einer Zahl", `to_int(10.6)`, "10"},

		// **Abschneiden, nicht runden.** Zwei Ebenen, die 10,4 und 10,6
		// rechnen, säßen beim Runden einen Punkt auseinander.
		{"auch nach unten", `to_int(10.4)`, "10"},
		{"negativ zur Null hin", `to_int(0 - 10.6)`, "-10"},
		{"aus Text", `to_int("42")`, "42"},
		{"eine ganze Zahl bleibt", `to_int(7)`, "7"},
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

// **Eine unbrauchbare Angabe ist ein Fehler, keine Null.**
// **Und der Fehler ist verzögert.**
//
//	return delay_error(ScriptErrorConversion(str, …, _TYPE_("integer")));
//
// `src/script/functions/basic.cpp`, Zeile 123, Fassung f9d9356 vom
// 2021-07-06. Die Vorlagen schreiben `to_int(x) or else 0`; wer hier sofort
// abbräche, ließe **607** Kartenfelder scheitern, deren Vorlage den Fall
// längst abgefangen hat.
func TestToIntWeistUnbrauchbaresAb(t *testing.T) {
	// Benutzt wirft er.
	if _, err := msescript.AuswertenQuelle(t.Context(), `to_int("keine Zahl") + 1`,
		msescript.Bibliothek()); err == nil {
		t.Error("kein Fehler")
	}

	// Aufgefangen nicht.
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`to_string(to_int("keine Zahl") or else 3)`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "3" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "3")
	}
}

// **`split_text` behält leere Teile — und `include_empty: false` nimmt sie weg.**
//
//	SCRIPT_PARAM_DEFAULT(bool, include_empty, true);
//	while (match->matches(results, start, input.end())) {
//	  if (include_empty || pos.first != start) {
//	    ret->value.push_back(to_script(String(start, pos.first)));
//	  }
//	  start = pos.second;
//	}
//	if (include_empty || start != input.end()) {
//	  ret->value.push_back(to_script(String(start, input.end())));
//	}
//
// `src/script/functions/regex.cpp`, Zeilen 212–232, Fassung f9d9356 vom
// 2021-07-06.
//
// **Der leere Text ergibt damit ein Element, nicht keines.** Hier stand das
// Gegenteil, und die Begründung war erschlossen: „eine leere Regelbox hätte
// danach eine Zeile". Sie ist falsch — für das Zeilenzählen schreibt
// `magic.mse-game` ausdrücklich `include_empty: false`
// (`line_count := split_text@(match:"\n+", include_empty:false) + length`,
// Zeile 47).
//
// **Was das gekostet hat:** `pull_comma_array` liest `array[cell]` aus einer
// Zerlegung. War die leer, wo MSE ein Element hat, gab der Zugriff einen
// Fehlerwert — **280 Befunde** über die 190 Stile von `M15-Magic-Pack`
// (#345).
func TestSplitTextBehaeltLeereTeile(t *testing.T) {
	faelle := []struct {
		name, quelle string
		erwartet     string
	}{
		{"der leere Text ergibt ein Element", `to_string(length(split_text("", match: ",")))`, "1"},
		{"und das Element ist leer", `to_string(split_text("", match: ",")[0])`, ""},
		{"zwei Teile bleiben zwei", `to_string(length(split_text("a,b", match: ",")))`, "2"},
		{"ein Komma am Ende zählt mit", `to_string(length(split_text("a,", match: ",")))`, "2"},
		{"zwei Kommas hintereinander", `to_string(length(split_text("a,,b", match: ",")))`, "3"},

		// **`include_empty: false` nimmt sie weg** — so zählt das Pack Zeilen.
		{
			"ohne leere Teile",
			`to_string(length(split_text("a,,b,", match: ",", include_empty: false)))`,
			"2",
		},
		{
			"der leere Text ohne leere Teile",
			`to_string(length(split_text("", match: ",", include_empty: false)))`,
			"0",
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
