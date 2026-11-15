// **Ein Bildwert ist ein Dateiname, bis er geklammert ist — und ein
// Textskript ist kein Skript** — FMT-02 (#1691).
// Die Regeln und ihre Herkunft stehen an `bildschluessel` und
// `eintragsmerkmale`.
package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// artVon gibt die Farbe des Werts in einer Zeile — oder leer, wenn die Zeile
// keinen trägt.
func artVon(text string, zeile int) msepaket.Farbart {
	for _, merkmal := range msepaket.Einfaerben(text) {
		if merkmal.Zeile == zeile && merkmal.Art != msepaket.FarbeSchluessel {
			return merkmal.Art
		}
	}

	return ""
}

func TestEinBildwertIstEinDateinameBisErGeklammertIst(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		text  string
		zeile int
		art   msepaket.Farbart
	}{
		"ein Dateiname":         {"card style:\n\tname:\n\t\timage: rahmen.png\n", 3, msepaket.FarbeWert},
		"eine Maske mit Pfad":   {"card style:\n\tname:\n\t\tmask: /fremd.mse-style/maske.png\n", 3, msepaket.FarbeWert},
		"ein geklammertes Bild": {"card style:\n\tname:\n\t\timage: { \"a\" + \".png\" }\n", 3, msepaket.FarbeSkriptwert},
		"ein Wahlbild":          {"card style:\n\tfarbe:\n\t\tchoice images:\n\t\t\thell: hell.png\n", 4, msepaket.FarbeWert},
		"ein geklammertes Wahlbild": {
			"card style:\n\tfarbe:\n\t\tchoice images:\n\t\t\thell: { \"a.png\" }\n", 4, msepaket.FarbeSkriptwert,
		},
		"ein Wahlbild als script:-Unterblock": {
			"card style:\n\tfarbe:\n\t\tchoice images:\n\t\t\thell:\n\t\t\t\tscript: \"a.png\"\n", 5, msepaket.FarbeSkriptwert,
		},
		"ein Textskript unter refer script": {
			"keyword parameter type:\n\trefer script:\n\t\tname: schlicht\n\t\tscript: \\{{input}\\}\n", 4, msepaket.FarbeTextskriptwert,
		},
		// **#1934**: eine Klammer in der Mitte macht den Bildwert zur
		// Zeichenkette mit Ausdruck (`parse(*this, true)`,
		// `src/script/image.cpp` Zeile 73, Fassung f9d9356 vom 2021-07-06).
		"eine Maske mit Ausdruck in der Mitte": {
			"card style:\n\tgrund:\n\t\tmask: border-mask{pt()}.png\n", 3, msepaket.FarbeTextskriptwert,
		},
		"ein Wahlbild mit Ausdruck in der Mitte": {
			"card style:\n\tfarbe:\n\t\tchoice images:\n\t\t\thell: hell-{1}.png\n", 4, msepaket.FarbeTextskriptwert,
		},
		"ein Ausdruck vorn, aber nicht ganz geklammert": {
			"card style:\n\tname:\n\t\timage: {card.farbe}_rahmen.png\n", 3, msepaket.FarbeTextskriptwert,
		},
		"eine geschützte Klammer ist schon Zeichenkettenmodus": {
			"card style:\n\tname:\n\t\timage: a\\{b.png\n", 3, msepaket.FarbeTextskriptwert,
		},
		"ein gewöhnliches Skript bleibt eines": {"script: 1 + 2\n", 1, msepaket.FarbeSkriptwert},
	} {
		if art := artVon(fall.text, fall.zeile); art != fall.art {
			t.Errorf("%s: %q, erwartet %q", name, art, fall.art)
		}
	}
}
