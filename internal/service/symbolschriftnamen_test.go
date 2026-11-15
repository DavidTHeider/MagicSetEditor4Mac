package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Symbolschrift, die eine Stiloption benennt, wird trotzdem geladen.**
//
// # Der Befund
//
// Die Symbolschriften wurden unter dem Namen vorgeladen, wie er im Feldstil
// **dasteht**; gesucht wird unter dem **ausgerechneten**. Beides ist nur bei
// einem Literal dasselbe.
//
// `magic-m15.mse-style` schreibt am Regeltextfeld
//
//	symbol font:
//	    name: { styling.text_box_mana_symbols }
//
// und die Stiloption dazu trägt `initial: magic-mana-small.mse-symbol-font`.
//
// **Gezählt über `M15-Magic-Pack`: 326 von 445 `symbol font: name:` sind ein
// Skript**, und 160 `initial:`-Werte nennen eine Symbolschrift. Die Mehrheit
// wurde nie geladen
// (#328).
func TestEineSymbolschriftAusEinerStiloptionWirdGeladen(t *testing.T) {
	stil := msepaket.Stil{
		Felder: []msepaket.Feldstil{{
			Feld:          "text",
			Symbolschrift: msepaket.Symbolangabe{Name: "{ styling.symbole }"},
		}},
		Stiloptionen: []domain.StyleField{
			{Name: "symbole", Initial: "probe-symbole.mse-symbol-font"},
		},
	}

	namen := service.SymbolschriftnamenFuerTest(stil)

	if !enthaelt(namen, "probe-symbole") {
		t.Errorf("die Namen sind %v, erwartet auch probe-symbole", namen)
	}
}

// **Auch die Auswahlwerte einer Stiloption zählen.**
//
// Eine Option mit `choice:` kann jeden ihrer Werte annehmen; wer nur `initial:`
// lädt, hat die Schrift für jede andere Wahl nicht zur Hand.
func TestAuchDieAuswahlwerteEinerStiloptionZaehlen(t *testing.T) {
	stil := msepaket.Stil{
		Felder: []msepaket.Feldstil{{
			Feld:          "text",
			Symbolschrift: msepaket.Symbolangabe{Name: "{ styling.symbole }"},
		}},
		Stiloptionen: []domain.StyleField{{
			Name:    "symbole",
			Initial: "a.mse-symbol-font",
			Choices: []string{"a.mse-symbol-font", "b.mse-symbol-font"},
		}},
	}

	namen := service.SymbolschriftnamenFuerTest(stil)

	for _, erwartet := range []string{"a", "b"} {
		if !enthaelt(namen, erwartet) {
			t.Errorf("die Namen sind %v, erwartet auch %q", namen, erwartet)
		}
	}
}

// **Ein literaler Name bleibt, was er ist** — und ohne Endung.
func TestEinLiteralerSymbolschriftnameBleibt(t *testing.T) {
	stil := msepaket.Stil{
		Felder: []msepaket.Feldstil{{
			Feld:          "casting cost",
			Symbolschrift: msepaket.Symbolangabe{Name: "magic-mana-large"},
		}},
	}

	namen := service.SymbolschriftnamenFuerTest(stil)

	if len(namen) != 1 || !enthaelt(namen, "magic-mana-large") {
		t.Errorf("die Namen sind %v, erwartet nur magic-mana-large", namen)
	}
}

// **Was keine Symbolschrift benennt, kommt nicht mit.**
//
// Eine Stiloption `center text` mit den Werten `always`/`never` ist keine
// Schrift. Sie mitzuladen hieße, für jede Stiloption des Pakets ein Paket zu
// suchen — 190 Stile mit im Schnitt zwanzig Optionen.
func TestEineOptionOhneSymbolschriftKommtNichtMit(t *testing.T) {
	stil := msepaket.Stil{
		Felder: []msepaket.Feldstil{{
			Feld:          "text",
			Symbolschrift: msepaket.Symbolangabe{Name: "{ styling.symbole }"},
		}},
		Stiloptionen: []domain.StyleField{
			{Name: "zentriert", Initial: "never", Choices: []string{"always", "never"}},
		},
	}

	if namen := service.SymbolschriftnamenFuerTest(stil); len(namen) != 0 {
		t.Errorf("die Namen sind %v, erwartet keine", namen)
	}
}

// **Ohne Feld mit Symbolschrift wird nichts gesucht.**
func TestOhneSymbolfeldWirdNichtsGesucht(t *testing.T) {
	stil := msepaket.Stil{
		Stiloptionen: []domain.StyleField{
			{Name: "symbole", Initial: "a.mse-symbol-font"},
		},
	}

	if namen := service.SymbolschriftnamenFuerTest(stil); len(namen) != 0 {
		t.Errorf("die Namen sind %v, erwartet keine", namen)
	}
}

func enthaelt(namen []string, gesucht string) bool {
	for _, name := range namen {
		if name == gesucht {
			return true
		}
	}

	return false
}
