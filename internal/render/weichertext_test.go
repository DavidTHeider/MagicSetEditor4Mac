package render_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Weicher Text wird gemessen, aber nicht gezeichnet.**
//
// # Der Befund
//
// Der Gedankenstrich der Typzeile stand bei jeder Karte da. Bei MSE2 steht er
// nur, wenn ein Untertyp folgt: `Creature — Goblin`, aber `Instant`.
//
// **Gemeldet aus der laufenden Anwendung am 2026-08-21** —
// #725.
//
// # Woran es lag
//
// Nicht am Zusammensetzen — das war richtig. `combined_editor` schreibt bei
// leerem zweiten Teil einen **weichen** Trenner:
//
//	} else if (value_parts[i].second && soft_before_empty) {
//	  new_value += _("<sep-soft>") + separators[i - 1] + _("</sep-soft>");
//
// `src/script/functions/editor.cpp`, Zeilen 104–114, Fassung f9d9356 vom
// 2021-07-06. Und `magic.mse-game` setzt beide Schalter (`card_fields`, Zeilen
// 985–991).
//
// Es lag am **Zeichnen**. Ein weiches Textstück bekommt bei MSE
// `draw_as = DRAW_ACTIVE`:
//
//	DrawWhat what = soft > 0 ? DRAW_ACTIVE : DRAW_NORMAL;
//
// `src/render/text/element.cpp`, Zeile 234 — und gezeichnet wird es nur, wenn
// der Aufrufer danach fragt:
//
//	if ((what & draw_as) != draw_as) return; // don't draw
//
// `FontTextElement::draw`, `src/render/text/font.cpp`, Zeilen 15–16.
// `DRAW_ACTIVE` heißt „draw active editor stuff, such as hidden separators"
// (`src/data/draw_what.hpp`, Zeile 23) und wird allein vom **Editor** gesetzt.
// Auf einer Karte steht es nie.
//
// # Was die Messung sagt
//
// Über fünf kanonische Sets, alle Karten, Stil `magic-m15`:
//
//	| Set       | Karten | Strich sichtbar am Ende |
//	| Black     |  3.788 |                   1.308 |
//	| Blue      |  3.693 |                   1.381 |
//	| Green     |  3.689 |                   1.133 |
//	| Red       |  3.730 |                   1.332 |
//	| Vanguard  |    107 |                     107 |
//
// **Rund jede dritte Karte** trug einen Gedankenstrich, hinter dem nichts
// stand. Die zweite Ursache steht in `bibliothek_verbinden.go`: „leer" hieß
// dort „die leere Zeichenkette", und ein leerer Untertyp kommt aus
// `magic.mse-game` als leeres Wortlistenfeld.
//
// # Und was **doch** gezeichnet wird
//
// Ein **Symbolstück** in weicher Klammer. `SymbolTextElement::draw` fragt nach
// `DRAW_NORMAL` und kennt kein `draw_as` (`src/render/text/symbol.cpp`, Zeilen
// 15–20); die Weichheit erreicht es gar nicht, weil der Zweig für Symbole vor
// der Zeile mit `DRAW_ACTIVE` steht.
func TestWeicherTextWirdNichtGezeichnet(t *testing.T) {
	schrift := echteSchrift(t)

	// gezeichnet gibt die gedeckte Fläche einer Typzeile.
	gezeichnet := func(t *testing.T, abschnitte []render.Abschnitt) int {
		t.Helper()

		plan := &render.Plan{Breite: 200, Hoehe: 40, Ebenen: []render.Ebene{{
			Art:        render.ArtText,
			Rahmen:     render.Rechteck{Links: 4, Oben: 4, Breite: 192, Hoehe: 32},
			Abschnitte: abschnitte,
			Farbe:      render.Farbe{A: 255},
			Satz:       render.Satz{Schrift: "Andale", Groesse: 12, Ausrichtung: render.LinksOben},
			Deckung:    1,
		}}}

		bild, befunde, err := render.NeuerRasterer(bilder{}).
			MitSchriften(schriften{"Andale": schrift}).
			Zeichne(t.Context(), plan, 1, render.GueteMittel)
		if err != nil {
			t.Fatalf("abgewiesen: %v", err)
		}

		if len(befunde) != 0 {
			t.Fatalf("%v", befunde)
		}

		return gedeckteFlaeche(bild)
	}

	// **Zwei Zeilen, ein Unterschied.** Beide tragen denselben Text; in der
	// zweiten steht der Gedankenstrich weich.
	hart := gezeichnet(t, []render.Abschnitt{
		{Text: "Vanguard"},
		{Text: " — "},
	})

	weich := gezeichnet(t, []render.Abschnitt{
		{Text: "Vanguard"},
		{Text: " — ", Weich: true},
	})

	nurName := gezeichnet(t, []render.Abschnitt{{Text: "Vanguard"}})

	if weich != nurName {
		t.Errorf("mit weichem Trenner sind %d Bildpunkte gedeckt, ohne Trenner %d — "+
			"der weiche Trenner wurde gezeichnet", weich, nurName)
	}

	if hart <= nurName {
		t.Errorf("der harte Trenner deckt %d Bildpunkte, der Name allein %d — "+
			"dann prüft dieser Test nichts", hart, nurName)
	}
}

// **Ein weiches Symbolstück wird gezeichnet.**
//
// Die Gegenrichtung zum Test darüber, und sie steht hier, weil eine Regel „was
// weich ist, wird nicht gemalt" eine Zusage zu viel gäbe. MSE entscheidet die
// Weichheit im Textzweig; der Symbolzweig steht davor und kennt sie nicht
// (`src/render/text/element.cpp`, Zeilen 229–234, Fassung f9d9356 vom
// 2021-07-06).
func TestEinWeichesSymbolstueckWirdGezeichnet(t *testing.T) {
	schrift := echteSchrift(t)

	plan := &render.Plan{Breite: 60, Hoehe: 40, Ebenen: []render.Ebene{{
		Art:    render.ArtText,
		Rahmen: render.Rechteck{Links: 0, Oben: 0, Breite: 60, Hoehe: 40},
		Abschnitte: []render.Abschnitt{{
			Text:  "W",
			Weich: true,
			Symbole: []render.Symbol{
				{Bild: render.Blatt("mana.png"), Text: ""},
			},
			Symbolmasse: render.Symbolmasse{Groesse: 20, Bildgroesse: 20},
		}},
		Farbe:   render.Farbe{A: 255},
		Satz:    render.Satz{Schrift: "Andale", Groesse: 12, Ausrichtung: render.LinksOben},
		Deckung: 1,
	}}}

	bild, befunde, err := render.NeuerRasterer(bilder{"mana.png": vollflaeche(20, 20)}).
		MitSchriften(schriften{"Andale": schrift}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	if len(befunde) != 0 {
		t.Fatalf("%v", befunde)
	}

	if gedeckteFlaeche(bild) == 0 {
		t.Error("das weiche Symbolstück wurde nicht gezeichnet")
	}
}
