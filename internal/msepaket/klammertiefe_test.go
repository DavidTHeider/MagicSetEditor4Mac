package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// TestEineOffeneKlammerEndetMitDerEinrueckung ist der Fall, der 1.980 Karten
// gekostet hat.
//
// # Der Befund
//
// `Blue.mse-set` führt die Karte `Magical Hacker`, deren Flavortext Leetspeak
// ist: `1|= y()u (4|\| r3@d 7#][5 …`. Darin steht eine runde Klammer, die nie
// geschlossen wird.
//
// Der Zerteiler zählte sie und las **alles danach** als Fortsetzung — 1.980
// der 3.693 Karten des Sets gingen still verloren, und mit ihnen 7,9 % des
// ganzen Kanons. Kein Fehler, keine Meldung, nur weniger Karten.
//
// # Was MSE2 tut
//
//	} else if (value.empty()) {
//	  // a multiline string
//	  readLine(true);
//	  while (indent >= expected_indent && !input.Eof()) {
//	    previous_value += line.substr(expected_indent);
//	    …
//	  }
//
// `src/util/io/reader.cpp`, Zeilen 258–272, Fassung f9d9356 vom 2021-07-06.
// **Ein mehrzeiliger Wert endet an der Einrückung und an nichts sonst** — MSE2
// zählt dabei keine einzige Klammer.
func TestEineOffeneKlammerEndetMitDerEinrueckung(t *testing.T) {
	t.Parallel()

	inhalt := "mse version: 2.0.0\n" +
		"card:\n" +
		"\tname: Magical Hacker\n" +
		"\tflavor_text:\n" +
		"\t\t<i-flavor>1|= y()u (4|\\| r3@d</i-flavor>\n" +
		"\tpower: 1\n" +
		"card:\n" +
		"\tname: Magnetic Flux\n"

	baum, err := msepaket.Zerteile("probe", []byte(inhalt))
	if err != nil {
		t.Fatalf("Zerteile: %v", err)
	}

	var karten int

	for _, knoten := range baum.Wurzel {
		if knoten.Zeilenart == msepaket.ArtEintrag && knoten.Schluessel == "card" {
			karten++
		}
	}

	if karten != 2 {
		t.Errorf("%d Karten, erwartet 2 — die offene Klammer hat den Rest "+
			"verschluckt", karten)
	}

	// **Und das Feld hinter dem Flavortext gehört noch zur ersten Karte.**
	erste := baum.Wurzel[1]
	if erste.Ersten("power") == nil {
		t.Error("`power` fehlt der ersten Karte — es wurde zur Fortsetzung")
	}
}

// TestEineOffeneKlammerTraegtIhrenMehrzeiligenAusdruck hält die Zusage aus
// #170 fest.
//
// **Die Klammerbilanz bleibt nötig.** Ein Skript darf über mehrere Zeilen
// laufen, und seine Zeilen tragen Doppelpunkte:
//
//	sort_text@(order: "XYZI[0123456789]HSCA(WUBRG)")
//
// Wer sie zerteilt, bekommt einen Schlüssel `sort_text@(order`. Solange die
// Fortsetzung **tiefer** eingerückt ist als die Zeile, die die Klammer
// geöffnet hat, gilt sie weiter.
func TestEineOffeneKlammerTraegtIhrenMehrzeiligenAusdruck(t *testing.T) {
	t.Parallel()

	inhalt := "mse version: 2.0.0\n" +
		"game: magic\n" +
		"init script:\n" +
		"\tmana_sort := {\n" +
		"\t\tsort_text@(\n" +
		"\t\t\torder: \"WUBRG\"\n" +
		"\t\t)\n" +
		"\t}\n" +
		"card field:\n" +
		"\tname: danach\n"

	baum, err := msepaket.Zerteile("probe", []byte(inhalt))
	if err != nil {
		t.Fatalf("Zerteile: %v", err)
	}

	var wurzeleintraege []string

	for _, knoten := range baum.Wurzel {
		if knoten.Zeilenart == msepaket.ArtEintrag {
			wurzeleintraege = append(wurzeleintraege, knoten.Schluessel)
		}
	}

	erwartet := []string{"mse version", "game", "init script", "card field"}
	if len(wurzeleintraege) != len(erwartet) {
		t.Fatalf("Wurzeleinträge %v, erwartet %v", wurzeleintraege, erwartet)
	}

	for i, name := range erwartet {
		if wurzeleintraege[i] != name {
			t.Errorf("Wurzeleintrag %d ist %q, erwartet %q", i, wurzeleintraege[i], name)
		}
	}
}

// **Ein Eintrag derselben Tiefe ist ein Geschwister, keine Fortsetzung**
// (#2293).
//
// # Der Befund
//
// `magic.mse-game/keywords_de` führt den Parameter »Präfix« mit
// `match: [^(.,”":;\n—]+` — das ASCII-Anführungszeichen in der Zeichenklasse
// öffnet eine Zeichenkette, die nie schließt, und die `]` zählt nicht. Die
// Klammer blieb offen, und `optional: false` darunter wurde zur Fortsetzung:
// der Parameter war in zehn Sprachen **optional**. Gemessen am 2026-10-11
// über den vollen Pack: 16 solcher Verluste in `magic.mse-game`.
//
// # Was MSE2 tut
//
// Es liest einen Wert bis zum Zeilenende (`src/util/io/reader.cpp`, Zeilen
// 207–223, Fassung f9d9356 vom 2021-07-06) und einen Block nur aus
// **tieferen** Zeilen (Zeilen 258–272). Eine Klammer, die ein Eintrag öffnet,
// trägt deshalb nur, was tiefer steht.
func TestEinEintragDerselbenTiefeIstEinGeschwister(t *testing.T) {
	t.Parallel()

	inhalt := "keyword parameter type:\n" +
		"\tname: Präfix\n" +
		"\tmatch: [^(.,”\":;\\n—]+\n" +
		"\toptional: false\n" +
		"keyword parameter type:\n" +
		"\tname: Name\n"

	baum, err := msepaket.Zerteile("keywords_de", []byte(inhalt))
	if err != nil {
		t.Fatalf("Zerteile: %v", err)
	}

	erster := baum.Ersten("keyword parameter type")
	if erster == nil {
		t.Fatal("kein Parameter")
	}

	if optional := erster.Ersten("optional"); optional == nil || optional.Wert != "false" {
		t.Error("`optional: false` ist zur Fortsetzung von `match:` geworden")
	}

	if match := erster.Ersten("match"); match == nil || match.Wert != "[^(.,”\":;\\n—]+" {
		t.Errorf("`match:` ist nicht mehr, was dasteht: %+v", match)
	}
}
