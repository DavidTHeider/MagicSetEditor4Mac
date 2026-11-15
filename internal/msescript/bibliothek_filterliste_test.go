package msescript_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`filter_list` siebt eine Liste nach einer Bedingung** —
// #730.
//
//	ScriptValueP it = input->makeIterator();
//	while (ScriptValueP v = it->next()) {
//	  ctx.setVariable(SCRIPT_VAR_input, v);
//	  if (filter->eval(ctx)->toBool()) { ret->value.push_back(v); }
//	}
//	// TODO : somehow preserve keys
//
// `src/script/functions/basic.cpp`, Zeilen 574–588, Fassung f9d9356 vom
// 2021-07-06.
func TestFilterList(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, quelle, erwartet string }{
		{
			name:     "was die Bedingung hält, bleibt",
			quelle:   `filter_list([1, 2, 3, 4], filter: {input > 2})[0]`,
			erwartet: "3",
		},
		{
			name:     "und die Länge stimmt",
			quelle:   `length(filter_list([1, 2, 3, 4], filter: {input > 2}))`,
			erwartet: "2",
		},

		// **Die Namen werden neu vergeben.** Der Kommentar im Original sagt
		// es selbst: `ScriptCustomCollection` bekommt die Werte über
		// `push_back`, die Schlüssel bleiben liegen. Dieselbe Regel wie bei
		// `sort_list`.
		{
			name:     "die Nummern werden neu vergeben",
			quelle:   `filter_list([1, 2, 3, 4], filter: {input > 3})[0]`,
			erwartet: "4",
		},
		{
			name:     "eine Bedingung, die nie hält, gibt nichts",
			quelle:   `length(filter_list([1, 2], filter: {false}))`,
			erwartet: "0",
		},
		{
			name:     "und eine, die immer hält, gibt alles",
			quelle:   `length(filter_list([1, 2], filter: {true}))`,
			erwartet: "2",
		},

		// **Eine leere Liste bleibt leer** — `magic.mse-game` ruft es so:
		// `if input == [] then false else length(filter_list(…)) > 0`.
		{
			name:     "eine leere Liste",
			quelle:   `length(filter_list([], filter: {true}))`,
			erwartet: "0",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

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

// **Der Rumpf läuft über der Umgebung des Aufrufers.**
//
// `magic.mse-game/script` Zeile 259 siebt gegen ein `element`, das der Aufrufer
// erklärt hat:
//
//	if input == [] then false else length(filter_list(input, filter:{input == element})) > 0
//
// Über einer frischen Umgebung wäre der Name fort — dieselbe Klasse wie bei
// `sort_list` (#651).
func TestDerFilterSiehtDieUmgebungDesAufrufers(t *testing.T) {
	t.Parallel()

	wert, err := msescript.AuswertenQuelle(t.Context(),
		"element := 2\nlength(filter_list([1, 2, 3], filter: {input == element}))",
		msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "1" {
		t.Errorf("%q, erwartet 1", wert.AlsText())
	}
}

// **`card` wird nicht gesetzt** — anders als bei `sort_list`.
//
//	ctx.setVariable(SCRIPT_VAR_input, v);
//
// `basic.cpp`, Zeile 580: **nur `input`.** `sort_list` hat die Verzweigung
// `set ? _("card") : _("input")` (Zeile 512), `filter_list` nicht. Beide
// Aufrufe im Pack schreiben `input` aus — `filter:{included_rarity(card:input)}`
// in `magic-cockatrice-v2` gibt es als **Argument** weiter.
func TestDerFilterSetztKeinCard(t *testing.T) {
	t.Parallel()

	_, err := msescript.AuswertenQuelle(t.Context(),
		`length(filter_list([1, 2], filter: {card == 1}))`, msescript.Bibliothek())
	if err == nil || !strings.Contains(err.Error(), "card") {
		t.Fatalf("Fehler = %v", err)
	}
}

// **Was keine Liste ist, wird benannt** — dieselbe Meldung wie bei `sort_list`.
func TestFilterListBrauchtEineListe(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, quelle string }{
		{"eine Zahl", `filter_list(12, filter: {true})`},
		{"ohne Eingabe", `filter_list(filter: {true})`},
		{"ohne filter", `filter_list([1, 2])`},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err == nil || !strings.Contains(err.Error(), "filter_list") {
				t.Fatalf("Fehler = %v", err)
			}
		})
	}
}

// **Was sich nicht in eine Wahrheit umrechnen lässt, ist ein Fehler** — Befund
// des Zweitblicks durch Codex vom 2026-08-21.
//
//	bool toBool() const override {
//	  if (value == _("yes") || value == _("true")) { return true; }
//	  else if (value == _("no") || value == _("false") || value.empty()) { return false; }
//	  else { throw ScriptErrorConversion(value, typeName(), _TYPE_("boolean")); }
//	}
//
// `ScriptString::toBool`, `src/script/value.cpp`, Zeilen 345–352, Fassung
// f9d9356 vom 2021-07-06. `filter_list` ruft `filter->eval(ctx)->toBool()`
// (`basic.cpp`, Zeile 581) — und die Ausnahme geht durch bis zum Aufrufer.
//
// **Diese Umsetzung hat den zweiten Rückgabewert von [Wert.AlsWahrheit]
// weggeworfen** und `"vielleicht"` als `false` behandelt. Die Vorlage bekäme
// eine still verkürzte Liste statt einer Meldung.
func TestEineUnbrauchbareBedingungIstEinFehler(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, quelle string }{
		{"eine Zeichenkette, die keine Wahrheit ist", `filter_list([1], filter: {"vielleicht"})`},
		{"eine Zahl", `filter_list([1], filter: {7})`},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err == nil {
				t.Fatal("die Bedingung ging durch")
			}

			if !strings.Contains(err.Error(), "Wahrheitswert") {
				t.Errorf("Meldung: %v", err)
			}
		})
	}

	// **Was MSE als Wahrheit liest, geht durch** — `yes`, `true`, `no`,
	// `false` und die leere Zeichenkette.
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`length(filter_list([1, 2], filter: {"yes"}))`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "2" {
		t.Errorf("%q, erwartet 2", wert.AlsText())
	}
}
