package msescript_test

// **POSIX-Klassen auch auf der Rückfalllinie** — #2315 (E2315).
//
// MSE2 übersetzt mit `boost::regex`, und das kennt `[[:space:]]` in jedem
// Muster. Hier übersetzt zuerst RE2, das sie auch kennt — ein Muster mit
// Rückblick oder `(?x)` fällt aber auf `regexp2` zurück, und das liest
// `[[:space:]]` als Klasse aus `[`, `:`, `s` … gefolgt von einem `]`.
// **Gefunden am Kontext `mana_context` von `magic.mse-game`**: er beginnt
// mit `(?ix)`, fiel zurück, und „you get TK" wurde kein Symbol, wo MSE2 eines
// zeichnet.
//
// **Erwartungen vom Orakel** (`magicseteditor --cli … -q --raw`, Fassung
// f9d9356, am 2026-10-14), eigene, kleine Sätze.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestPosixKlassenGeltenAuchAufDerRueckfalllinie(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name, quelle, erwartet string
	}{
		{"Leerraum im erweiterten Modus", `match("x you get y", match: "(?ix) (^|[[:space:]])you[ ]get")`, "true"},
		{"Buchstabe im Rückblick, Ziffer dahinter", `replace("ab1 2", match: "(?<=[[:alpha:]])[[:digit:]]", replace: "X")`, "abX 2"},
		{"Satzzeichen nach einem Rückblick", `replace("a.b,c", match: "(?<!a)[[:punct:]]", replace: "_")`, "a.b_c"},
		{"die verneinte Klasse", `replace("a1_b", match: "(?<![[:^alnum:]])[[:alnum:]]", replace: "#")`, "##_b"},
		{"eine eckige Klammer als erstes Zeichen einer Klasse", `match("a]b", match: "(?x) a []] b")`, "true"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("%v", err)
			}

			if gesehen := wert.AlsText(); gesehen != fall.erwartet {
				t.Errorf("ergibt %q, MSE2 gibt %q", gesehen, fall.erwartet)
			}
		})
	}
}

// **Eine verneinte Teilklasse wird abgewiesen** — Befund 12 des Zweitblicks
// über Phase 1 von 0.9.0. RE2 kennt `[a[:^digit:]]`; .NET hat dafür keine
// Entsprechung in ASCII (`\D` meint dort Unicode), und stehen gelassen las
// `regexp2` die Schreibweise als Zeichen. **Lieber laut als still falsch** —
// im Pack steht sie nicht.
func TestEineVerneinteTeilklasseWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, err := msescript.UebersetzeMuster(`(?x)[a[:^digit:]]`); err == nil {
		t.Error("die verneinte Teilklasse wurde übersetzt")
	}

	if _, err := msescript.UebersetzeMuster(`(?x)[[:^digit:]]`); err != nil {
		t.Errorf("die ganze verneinte Klasse wurde abgewiesen: %v", err)
	}
}
