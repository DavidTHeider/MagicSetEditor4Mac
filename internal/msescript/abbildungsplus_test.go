package msescript_test

// **`+` auf zwei Abbildungen verkettet, es vereinigt nicht**
// (#271).
//
//	ScriptValueP ScriptConcatCollection::getMember(const String& name) const {
//	  ScriptValueP member = a->getMember(name);
//	  if (member->type() != SCRIPT_ERROR) return member;
//	  …
//	  return b->getMember(name);
//	}
//
// `src/script/value.cpp`, Zeilen 551–562, Fassung f9d9356 vom 2021-07-06.
// **Die linke Seite wird zuerst gefragt** — bei einem doppelten Schlüssel
// gewinnt sie.
//
// # Am Orakel geprüft
//
// `magicseteditor --cli White.mse-set -q --raw`
// ([0026](../../docs/entscheidungen/0026-mse2-rendert-die-referenz-selbst.md)):
//
//	to_string(([a: 1] + [a: 2]).a)      → "1"
//	to_string(length([a: 1] + [a: 2]))  → "2"
//	to_string(([1, 2] + [3]).1)         → "2"
//
// # Warum das gemessen wurde, bevor es geändert wurde
//
// Die alte Fassung — Vereinigung mit Vorrang für **rechts** — kam aus
// beobachtetem Paketverhalten (`map := map + [stylesheet: …]`), und ein
// Wechsel ohne Zahl wäre ein Rückschritt gewesen (§10). Über den Packlauf
// (190 Stile, 5.950 Feldstile) treffen **6.464** Verbindungen zweier
// Abbildungen auf mindestens einen doppelten Schlüssel.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestAbbildungsplusVerkettet(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		{"die linke Seite gewinnt", `to_string(([a: 1] + [a: 2]).a)`, "1"},
		{"beide Glieder bleiben", `to_string(length([a: 1] + [a: 2]))`, "2"},
		{"ein neuer Schlüssel kommt dazu", `to_string(([a: 1] + [b: 2]).b)`, "2"},
		{"die linke bleibt daneben", `to_string(([a: 1] + [b: 2]).a)`, "1"},

		// **Eine Liste zählt weiter über beide.**
		{"Listen hängen aneinander", `to_string(([1, 2] + [3]).1)`, "2"},
		{"und die rechte folgt", `to_string(([1, 2] + [3]).2)`, "3"},
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

// `+` vereinigt Abbildungen, und der Eintrag rechts gewinnt.
//
// Der zuvor genannte Testfall prüfte nur, dass eine Abbildung gelesen wird —
// nicht die behauptete Vereinigung. Gefunden vom Nachlauf des Zweitblicks über
// Phase 3.
func TestDiePlusVerkettungVonAbbildungen(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		// **Beide Glieder bleiben stehen**, und gelesen wird das linke —
		// am Orakel geprüft: `length([a: 1] + [a: 2])` ergibt 2, und
		// `([a: 1] + [a: 2]).a` ergibt 1
		// (#271).
		{"beide Einträge bleiben", `[a: 1, b: 2] + [b: 9]`, `[a: 1, b: 2, b: 9]`},
		{"neue Einträge kommen hinten an", `[a: 1] + [b: 2]`, `[a: 1, b: 2]`},
		{"die Reihenfolge der linken bleibt", `[b: 1, a: 2] + [c: 3]`, `[b: 1, a: 2, c: 3]`},
		{"mit einer leeren Abbildung", `[a: 1] + []`, `[a: 1]`},
		{"eine leere links", `[] + [a: 1]`, `[a: 1]`},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if wert := rechne(t, fall.quelle, nil); wert.String() != fall.erwartet {
				t.Errorf("%s, erwartet war %s", wert.String(), fall.erwartet)
			}
		})
	}
}
