// Die Prüfungen über die Funktionen auf Werten — Text, Zahl, Wahrheit.
//
// **Eigene Datei**, weil `bibliothek_test.go` sonst über die Zeilengrenze aus
// AGENTS.md §2 wüchse. Der Schnitt folgt der Datei, die sie prüft.
package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`to_text` macht aus jedem Wert seinen Text.**
//
// # Woher der Fall kommt
//
// Aus den Exportvorlagen: `to_text(card.casting_cost)`,
// `to_text(for each card in cards do …)` — alle 19 Vorlagen von
// `M15-Magic-Pack` benutzen sie, und ohne sie läuft keine einzige
// (#51).
//
// **Sie gehört zur Standardbibliothek** und damit zum Umfang von
// #25; sie
// steht hier, weil #51 sonst nichts vorzuweisen hätte.
func TestToTextMachtAusJedemWertSeinenText(t *testing.T) {
	faelle := map[string]string{
		`to_text("schon Text")`: "schon Text",
		`to_text(3)`:            "3",
		`to_text(true)`:         "true",
		`to_text(3 + 4)`:        "7",
	}

	for quelle, erwartet := range faelle {
		t.Run(quelle, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(
				t.Context(), quelle, msescript.Bibliothek().Darueber())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), erwartet)
			}
		})
	}

	// **Mit genau einem Wert**, wie jede andere auch: `to_text()` ohne Wert ist
	// ein Tippfehler und kein leerer Text.
	if _, err := msescript.AuswertenQuelle(
		t.Context(), "to_text()", msescript.Bibliothek().Darueber()); err == nil {
		t.Error("kein Fehler ohne Wert")
	}
}

// **`replace` und `filter_text` — die Musterfunktionen.**
//
// `replace@(match: …, replace: …)` blockierte acht der neunzehn
// Exportvorlagen von `M15-Magic-Pack` allein durch ihr Fehlen
// (#51).
func TestReplaceUndFilterText(t *testing.T) {
	faelle := map[string]string{
		`replace@(match: "a+", replace: "X")("baaad")`: "bXd",

		// **`replace:` darf fehlen und heißt „weg damit".**
		`replace@(match: "[0-9]")("a1b2")`: "ab",

		// **Wörtlich ersetzt, nicht mit Gruppen.** `$1` ist in einer Vorlage
		// ein Dollarzeichen und keine Rückreferenz.
		`replace@(match: "a", replace: "$1")("a")`: "$1",

		// `filter_text` behält, was trifft — das Gegenstück.
		`filter_text@(match: "[0-9]")("a1b2")`: "12",
	}

	for quelle, erwartet := range faelle {
		t.Run(quelle, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(
				t.Context(), quelle, msescript.Bibliothek().Darueber())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), erwartet)
			}
		})
	}
}

// **Was schiefgehen kann, meldet sich mit Namen.**
func TestDieMusterfunktionenMeldenIhreFehler(t *testing.T) {
	for name, quelle := range map[string]string{
		"replace ohne match":              `replace@(replace: "x")("a")`,
		"replace ohne Eingabe":            `replace@(match: "a")()`,
		"replace mit kaputtem Muster":     `replace@(match: "[")("a")`,
		"filter_text ohne match":          `filter_text@(replace: "x")("a")`,
		"filter_text ohne Eingabe":        `filter_text@(match: "a")()`,
		"filter_text mit kaputtem Muster": `filter_text@(match: "[")("a")`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := msescript.AuswertenQuelle(
				t.Context(), quelle, msescript.Bibliothek().Darueber()); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **Die Namen der Bibliothek sind abfragbar** — für Wächter, nicht für die
// Auswertung. `service.TestDieSpracheKannKeineDateiAnfassen` benutzt sie, um
// festzuhalten, dass keine Funktion ins Dateisystem greift (#51).
func TestDieBibliothekNenntIhreNamen(t *testing.T) {
	namen := msescript.Bibliothek().Namen()

	if len(namen) == 0 {
		t.Fatal("keine Namen")
	}

	gefunden := false

	for _, name := range namen {
		if name == "to_text" {
			gefunden = true
		}
	}

	if !gefunden {
		t.Errorf("to_text fehlt in %v", namen)
	}
}

// **Eine unbekannte Fluchtfolge behält ihren Rückstrich.**
//
// `"!promo\b"` meint das Wortende eines Musters; wer den Rückstrich
// verschluckte, machte ein anderes Muster daraus.
func TestEineUnbekannteFluchtfolgeBehaeltIhrenRueckstrich(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(
		t.Context(), `"!promo\b"`, msescript.Bibliothek().Darueber())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != `!promo\b` {
		t.Errorf("%q", wert.AlsText())
	}

	// Und ein Tabulator wird übersetzt.
	tab, err := msescript.AuswertenQuelle(
		t.Context(), `"a\tb"`, msescript.Bibliothek().Darueber())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if tab.AlsText() != "a\tb" {
		t.Errorf("%q", tab.AlsText())
	}
}
