package render_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Eine leere Schlusszeile zählt nicht zur Blockhöhe**
// (#528).
//
// # Woher die leere Zeile kommt
//
// `magic.mse-game/card_fields`, Zeilen 1337–1354: das Kartenfeld `text` ist
//
//	combined_editor(field1: card.rule_text, separator: "<line>\n</line>", field2: card.flavor_text)
//
// Ist der Flavortext leer, endet der Wert auf `<line>\n</line>` und **nichts**
// dahinter. Über die zehn kanonischen Sets trifft das **9.380 von 25.148**
// Karten — 37,3 %.
//
// # Was MSE2 damit tut
//
//	// Find height of the text, don't count the last lines if they are empty
//	double height = 0;
//	for (size_t li = end_line - 1 ; li + 1 > start_line ; --li) {
//	  Line& l = lines[li];
//	  height = l.top + l.line_height;
//	  if (l.line_height) break; // not an empty line
//	}
//	height -= lines[start_line].top;
//
// `src/render/text/viewer.cpp`, Zeilen 771–778, und die Zeile, die die Höhe
// erst auf null setzt:
//
//	if (line.break_after == LineBreak::LINE) line.line_height = 0;
//
// ebenda Zeile 684. Fassung f9d9356 vom 2021-07-06.
//
// # Was es kostet, wenn man es nicht tut
//
// Der Block sitzt um **die halbe `<line>`-Sprunghöhe** zu hoch. Gemessen am
// 2026-08-19: 13 Bildpunkte bei `A-Cloister Gargoyle`, 12 bei
// `Choose Your Demise`.
func TestEineLeereSchlusszeileZaehltNichtZurBlockhoehe(t *testing.T) {
	t.Parallel()

	satz := render.Satz{
		Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksMitte,
		Zeilenfaktoren: render.Zeilenfaktoren{Weich: 1, Hart: 1, Linie: 1.5},
	}
	rahmen := render.Rechteck{Breite: 100, Hoehe: 100}
	schriften := render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}

	// **Ohne die leere Schlusszeile** — nur der Regeltext.
	ohne := render.SetzeAbschnitte(satz, rahmen,
		[]render.Abschnitt{{Text: "aaaa"}}, schriften, nil)

	// **Mit ihr**: `aaaa<line>\n</line>` und dahinter nichts — genau das, was
	// `combined_editor` bei leerem Flavortext baut. Der Abschnitt mit dem
	// Umbruch trägt das Blockende und keinen sichtbaren Text.
	mit := render.SetzeAbschnitte(satz, rahmen,
		[]render.Abschnitt{{Text: "aaaa"}, {Text: "\n", Blockende: true}},
		schriften, nil)

	if len(mit.Zeilen) < 2 {
		t.Fatalf("%d Zeilen, erwartet mindestens 2", len(mit.Zeilen))
	}

	if mit.Zeilen[0].Oben != ohne.Zeilen[0].Oben {
		t.Errorf("die erste Zeile sitzt bei %g statt %g — die leere "+
			"Schlusszeile zählt zur Blockhöhe",
			mit.Zeilen[0].Oben, ohne.Zeilen[0].Oben)
	}
}

// TestEineLeereZeileMittendrinZaehltMit ist die Gegenprobe zur Zusage oben.
//
// **MSE klammert nur die *letzten* leeren Zeilen aus** — der Kommentar sagt es
// wörtlich: „don't count the **last** lines if they are empty". Eine leere
// Zeile zwischen zwei Absätzen trägt ihre Höhe wie jede andere.
func TestEineLeereZeileMittendrinZaehltMit(t *testing.T) {
	t.Parallel()

	satz := render.Satz{
		Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksMitte,
		Zeilenfaktoren: render.Zeilenfaktoren{Weich: 1, Hart: 1, Linie: 1.5},
	}
	rahmen := render.Rechteck{Breite: 100, Hoehe: 100}
	schriften := render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}

	kurz := render.SetzeAbschnitte(satz, rahmen,
		[]render.Abschnitt{{Text: "aaaa"}, {Text: "bbbb"}}, schriften, nil)

	mitLuecke := render.SetzeAbschnitte(satz, rahmen,
		[]render.Abschnitt{
			{Text: "aaaa"},
			{Text: "\n", Blockende: true},
			{Text: "\nbbbb"},
		}, schriften, nil)

	if mitLuecke.Zeilen[0].Oben >= kurz.Zeilen[0].Oben {
		t.Errorf("mit Lücke beginnt der Block bei %g, ohne bei %g — die leere "+
			"Zeile mittendrin wurde übergangen",
			mitLuecke.Zeilen[0].Oben, kurz.Zeilen[0].Oben)
	}
}

// TestEinGanzLeererBlockHatKeineHoehe ist der Fall, den der Zweitblick über
// Phase J angestrichen hat.
//
//	Line() : start(0), end_or_soft(0), top(0), line_height(0) …
//
//	if (line_size.height < 0.01 && !lines.empty()) {
//	  // if a line has 0 height, use the height of the line above it, but at most once
//	} else {
//	  line.line_height = line_size.height;
//	}
//
// `src/render/text/viewer.cpp`, Zeilen 27–30 und 668–673 (dazu 703–716 für die
// letzte Zeile), Fassung f9d9356 vom 2021-07-06.
//
// **Die Vererbung greift nur, wenn es eine Zeile davor gibt.** Die **erste**
// Zeile ohne Inhalt bleibt bei null.
//
// # Und ein ganz leerer Block bekommt doch eine Höhe
//
// **Hier stand bis zum 2026-10-11, er habe die Höhe null.** Das war die halbe
// Regel: nach dem Setzen gibt MSE der **einzigen** Zeile ohne Breite eine
// Ersatzhöhe, die Höhe eines Zeichens der Schrift — „no text, find a dummy
// height" (`TextViewer::prepareLines`, `src/render/text/viewer.cpp`, Zeilen
// 397–405, Fassung f9d9356 vom 2021-07-06). Gefragt am 2026-10-11 mit einem
// Orakelstil: ein leerer `illustrator` meldet dieselbe `content_height` wie
// einer mit Text (#2298, `internal/service/leerinhalt_test.go`).
//
// Die Höhe geht als `content_height` in die Skripte der Vorlagen ein; wie
// hoch ein leeres Feld sich meldet, verschiebt, was daran hängt.
func TestEinGanzLeererBlockIstEineZeileHoch(t *testing.T) {
	t.Parallel()

	satz := render.Satz{Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben}
	rahmen := render.Rechteck{Breite: 100, Hoehe: 100}
	schriften := render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}

	leer := render.SetzeAbschnitte(satz, rahmen, []render.Abschnitt{{Text: ""}}, schriften, nil)
	gesetzt := render.SetzeAbschnitte(satz, rahmen, []render.Abschnitt{{Text: "aaaa"}}, schriften, nil)

	if len(leer.Zeilen) != 1 || leer.Hoehe != gesetzt.Hoehe || leer.Hoehe <= 0 {
		t.Errorf("%d Zeilen, Höhe %g — erwartet eine Zeile der Höhe %g",
			len(leer.Zeilen), leer.Hoehe, gesetzt.Hoehe)
	}
}

// TestEineLeereErsteZeileVorTextZaehltNicht prüft dieselbe Regel im Verbund.
//
// **Die erste Zeile bleibt bei null, die zweite trägt ihren Inhalt.** Der Block
// ist damit so hoch wie **eine** Zeile, nicht wie zwei.
func TestEineLeereErsteZeileVorTextZaehltNicht(t *testing.T) {
	t.Parallel()

	satz := render.Satz{Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben}
	rahmen := render.Rechteck{Breite: 100, Hoehe: 1000}
	schriften := render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}

	ohne := render.SetzeAbschnitte(satz, rahmen,
		[]render.Abschnitt{{Text: "aaaa"}}, schriften, nil)

	mit := render.SetzeAbschnitte(satz, rahmen,
		[]render.Abschnitt{{Text: "\naaaa"}}, schriften, nil)

	if mit.Hoehe != ohne.Hoehe {
		t.Errorf("mit leerer erster Zeile %g, ohne %g — sie zählt mit",
			mit.Hoehe, ohne.Hoehe)
	}
}
