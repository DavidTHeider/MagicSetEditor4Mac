// **Symbole im Regeltext** —
// gemeldet an der laufenden Anwendung: `Birds of Paradise` zeigte
// `{T}: Add one mana of any color.` statt des Tap-Symbols.
package scryfall_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// **Was MSE wirklich schreibt** — nachgelesen am 2026-08-24 in
// `Green.mse-set` (MSE 2.1.2, kanonischer Satz, `stylesheet: m15-altered`):
//
//	rule_text:
//		Other Elves you control get +1/+1.
//		<sym>T</sym>: Add <sym>G</sym><sym>G</sym><sym>G</sym>.
//
// Ein `<sym>` **je Symbol**, und der Erinnerungstext steht in `<i-auto>…</i-auto>`.
// Scryfall schreibt `{T}` und `{G}` und den Erinnerungstext in Klammern.
func TestDerRegeltextBekommtSeineSymbole(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct{ roh, erwartet string }{
		"das Tap-Symbol": {
			"{T}: Add one mana of any color.",
			"<sym>T</sym>: Add one mana of any color.",
		},
		"mehrere Symbole": {
			"{T}: Add {G}{G}{G}.",
			"<sym>T</sym>: Add <sym>G</sym><sym>G</sym><sym>G</sym>.",
		},
		"ein Hybridsymbol bleibt zusammen": {
			"{W/U}: Etwas.",
			"<sym>W/U</sym>: Etwas.",
		},
		"eine Zahl ist ein Symbol": {
			"{2}: Etwas.",
			"<sym>2</sym>: Etwas.",
		},
		"ohne Symbole bleibt alles stehen": {
			"Flying", "Flying",
		},
		"eine offene Klammer bleibt stehen": {
			"Kostet {2 mehr.", "Kostet {2 mehr.",
		},
		"leer bleibt leer": {"", ""},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if gefunden := scryfall.MitSymbolen(fall.roh); gefunden != fall.erwartet {
				t.Errorf("MitSymbolen(%q) = %q, erwartet %q",
					fall.roh, gefunden, fall.erwartet)
			}
		})
	}
}

// **Der Erinnerungstext wird kursiv gesetzt** — `<i-auto>`, wie MSE es
// schreibt. Der Zusatz `auto` sagt: automatisch erzeugt, nicht von Hand.
//
//	Crew 3 <i-auto>(Tap any number of creatures you control …)</i-auto>
//
// Nachgelesen in derselben Datei.
func TestDerErinnerungstextWirdKursiv(t *testing.T) {
	t.Parallel()

	roh := "Crew 3 (Tap any number of creatures you control.)"
	erwartet := "Crew 3 <i-auto>(Tap any number of creatures you control.)</i-auto>"

	if gefunden := scryfall.MitSymbolen(roh); gefunden != erwartet {
		t.Errorf("MitSymbolen(%q) = %q, erwartet %q", roh, gefunden, erwartet)
	}
}

// **Und ein Regeltext geht über die Zuordnung durch dieselbe Wandlung.**
//
// Ohne das stünde die Umwandlung an einer Stelle und würde an der anderen
// vergessen — genau die Klasse, die #1032 beim Bildfeld getroffen hat.
func TestDieZuordnungWandeltDenRegeltext(t *testing.T) {
	t.Parallel()

	zuordnung, _, bekannt := scryfall.ZuordnungFuer("magic")
	if !bekannt {
		t.Fatal("für magic gibt es keine Zuordnung")
	}

	werte := zuordnung.Werte(&scryfall.Karte{
		Name:      "Birds of Paradise",
		Regeltext: "Flying\n{T}: Add one mana of any color.",
		Aromatext: "Etwas {G} im Aroma.",
	})

	if werte["rule text"] != "Flying\n<sym>T</sym>: Add one mana of any color." {
		t.Errorf("rule text = %q", werte["rule text"])
	}

	// **Auch der Aromatext** — er trägt gelegentlich Symbole.
	if werte["flavor text"] != "Etwas <sym>G</sym> im Aroma." {
		t.Errorf("flavor text = %q", werte["flavor text"])
	}
}

// **Eine Klammer, die nicht schließt, bleibt eine Klammer.**
//
// MSEs Erinnerungstexte stehen immer geschlossen; eine offene Klammer über
// einen Zeilenumbruch hinweg wäre eine Vermutung, und Vermutungen erfindet
// dieses Paket nicht.
func TestEineOffeneKlammerBleibtStehen(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct{ roh, erwartet string }{
		"gar keine schließende": {
			"Kostet (2 mehr", "Kostet (2 mehr",
		},
		"erst nach einem Zeilenumbruch": {
			"Erste Zeile (offen\nZweite Zeile)",
			"Erste Zeile (offen\nZweite Zeile)",
		},
		"eine offene und danach ein Paar": {
			"Offen (hier\nUnd (geschlossen) da",
			"Offen (hier\nUnd <i-auto>(geschlossen)</i-auto> da",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if gefunden := scryfall.MitSymbolen(fall.roh); gefunden != fall.erwartet {
				t.Errorf("MitSymbolen(%q) = %q, erwartet %q",
					fall.roh, gefunden, fall.erwartet)
			}
		})
	}
}
