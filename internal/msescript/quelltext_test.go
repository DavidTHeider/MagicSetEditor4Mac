package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Ein Wert als Quelltext** —
// #1070.
//
// MSEs Kommandozeile zeigt das Ergebnis eines Ausdrucks mit `toCode()` und
// nicht mit `toString()` (`src/cli/cli_main.cpp`, Zeile 234, Fassung f9d9356
// vom 2021-07-06). Der Unterschied ist der Punkt: `"a"` ist eine
// Zeichenkette, `a` wäre ein Name.
func TestEinWertAlsQuelltext(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		wert     msescript.Wert
		erwartet string
	}{
		// **Zeichenketten bekommen Anführungszeichen und Fluchtzeichen** —
		// `quote_string`, `src/script/value.cpp`, Zeilen 303–317, Fassung f9d9356 vom 2021-07-06.
		"eine Zeichenkette":      {msescript.Text("a"), `"a"`},
		"mit Anführungszeichen":  {msescript.Text(`a"b`), `"a\"b"`},
		"mit Rückstrich":         {msescript.Text(`a\b`), `"a\\b"`},
		"mit Zeilenumbruch":      {msescript.Text("a\nb"), `"a\nb"`},
		"mit Wagenrücklauf":      {msescript.Text("a\rb"), `"a\rb"`},
		"mit Tabulator":          {msescript.Text("a\tb"), `"a\tb"`},
		"nichts heißt nil":       {msescript.Nichts, "nil"},
		"eine Zahl bleibt bloß":  {msescript.Zahl(3), "3"},
		"ein Wahrheitswert auch": {msescript.Wahrheit(true), "true"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if gebaut := fall.wert.AlsQuelltext(); gebaut != fall.erwartet {
				t.Errorf("AlsQuelltext = %q, erwartet %q", gebaut, fall.erwartet)
			}
		})
	}
}

// **Eine Sammlung wird eine Liste** — `ScriptCollectionBase::toCode`,
// `src/script/value.cpp`, Zeilen 467–484, Fassung f9d9356 vom 2021-07-06.
//
// **Mit Schlüssel, wo es einen gibt.** MSE prüft den Index: `-1` heißt „kein
// Zähler, also ein Name" und schreibt `schlüssel:wert`.
func TestEineSammlungAlsQuelltext(t *testing.T) {
	t.Parallel()

	liste := msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: []msescript.Eintrag{
		{Wert: msescript.Text("a")},
		{Wert: msescript.Zahl(2)},
	}}

	if gebaut := liste.AlsQuelltext(); gebaut != `["a",2]` {
		t.Errorf("Liste = %q", gebaut)
	}

	abbildung := msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: []msescript.Eintrag{
		{Name: "eins", Wert: msescript.Text("a")},
	}}

	if gebaut := abbildung.AlsQuelltext(); gebaut != `["eins":"a"]` {
		t.Errorf("Abbildung = %q", gebaut)
	}
}
