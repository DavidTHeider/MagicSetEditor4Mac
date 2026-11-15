package msepaket_test

// **Ein mehrzeiliger Wert behält seine tiefere Einrückung** (#2298).
//
//	previous_value += line.substr(expected_indent); // strip expected indent
//
// `Reader::getValue`, `src/util/io/reader.cpp`, Zeilen 255–298, Fassung
// f9d9356 vom 2021-07-06 — MSE nimmt je Zeile genau die Einrückung des Blocks
// weg und lässt jede tiefere stehen. `extra_data:` lebt davon: unter ihm
// stehen die Stilvorlagen, unter jeder ihre Felder. Ohne die zweite Ebene
// schrieb dieses Programm beim Sichern `keyword_icons` neben `magic-nouveau`
// statt darunter.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestEinBlockBehaeltSeineTiefereEinrueckung(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name, quelle, erwartet string
	}{
		{
			"zwei Ebenen",
			"card:\n\textra_data:\n\t\tmagic-nouveau:\n\t\t\tkeyword_icons: flying, haste\n\t\t\tcorner: day\n",
			"magic-nouveau:\n\tkeyword_icons: flying, haste\n\tcorner: day",
		},
		{
			"eine Ebene wie bisher",
			"card:\n\trule_text:\n\t\tErste Zeile\n\t\tZweite Zeile\n",
			"Erste Zeile\nZweite Zeile",
		},
		{
			"ein Sprung um zwei Tabs bleibt zwei",
			"card:\n\tnotes:\n\t\toben\n\t\t\t\tganz tief\n",
			"oben\n\t\tganz tief",
		},
		{
			// Eine Leerzeile ohne Einrückung zählt MSE als Zeilenumbruch
			// (`pending_newlines`), eine `#`-Zeile im Block ist Text.
			"Leerzeile und Raute im Block",
			"card:\n\tnotes:\n\t\toben\n\n\t\t# kein Kommentar\n\t\t\tunten\n",
			"oben\n\n# kein Kommentar\n\tunten",
		},
		{
			// **Die Grundeinrückung ist die des Schlüssels plus eins**, nicht
			// die der ersten Zeile: `Reader::handle` ruft `enterBlock`, und
			// erst danach liest `getValue` mit diesem `expected_indent`
			// (Befund 5 des Zweitblicks über E/F).
			"eine erste Zeile tiefer als der Block",
			"card:\n\tnotes:\n\t\t\tA\n\t\tB\n",
			"\tA\nB",
		},
		{
			"ein Wert hinter dem Doppelpunkt",
			"card:\n\tname: Probe\n",
			"Probe",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			baum, err := msepaket.Zerteile("set", []byte(fall.quelle))
			if err != nil {
				t.Fatalf("zerteilen: %v", err)
			}

			kind := baum.Wurzel[0].Kinder[0]
			if text := kind.Blocktext(); text != fall.erwartet {
				t.Errorf("%q, erwartet %q", text, fall.erwartet)
			}
		})
	}
}

// **Trägt der Schlüssel keine Rohzeile, gilt die erste Zeile** — ein
// Schlüssel, den der Schreiber neu zusammengesetzt hat, über gelesenen
// Kindern: ihre Einrückung ist die des Blocks, jede tiefere bleibt.
func TestOhneRohzeileDesSchluesselsGiltDieErsteZeile(t *testing.T) {
	t.Parallel()

	baum, err := msepaket.Zerteile("set", []byte("card:\n\tnotes:\n\t\toben\n\t\t\tunten\n"))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	gelesen := baum.Wurzel[0].Kinder[0]
	neu := &msepaket.Knoten{Schluessel: "notes", Kinder: gelesen.Kinder}

	if text := neu.Blocktext(); text != "oben\n\tunten" {
		t.Errorf("%q, erwartet %q", text, "oben\n\tunten")
	}
}
