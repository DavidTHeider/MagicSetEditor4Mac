package render_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Die Vorsilbenbreite rastet auf ¼ Punkt — einmal je Abschnitt, nicht je
// Wort** — #598.
//
// # Die Regel
//
//	double prev_width = 0;
//	size_t line_start = start;
//	for (size_t i = start ; i < end ; ++i) {
//	  …
//	  RealSize s = dc.GetTextExtent(content.substr(line_start - this->start, i - line_start + 1));
//	  out.push_back(CharInfo(RealSize(s.width - prev_width, s.height), …));
//	  prev_width = s.width;
//	}
//
// `FontTextElement::getCharInfo`, `src/render/text/font.cpp`, Zeilen 27–49, und
// `RotatedDC::GetTextExtent`, `src/util/rotation.cpp`, Zeilen 344–351, Fassung
// f9d9356 vom 2021-07-06: die Breite kommt als `w / (zoomX · text_scaling)`
// zurück, mit `text_scaling = 4` und ganzzahligem `w`.
//
// **Jeder Vorschub ist damit `raster(vorsilbe_i) − raster(vorsilbe_{i−1})`, und
// die Summe über einen Lauf teleskopiert zu `raster(vorsilbe_n)`** — eine
// einzige Rasterung.
//
// # Warum das nicht je Wort geht
//
// Der erste Anlauf hat `breiteVon` je Stück gerastert. Gemessen am Kanon:
// **20,679 → 24,728**, um 4,05 schlechter. `ceil` rundet immer auf, und bei
// zehn Wörtern kommt zehnmal bis zu ¼ Punkt zu viel heraus.
//
// # Aufgerundet, nicht gerundet
//
// Neun Zeichenketten bei zwei Punktgrößen gegen wx 3.3.3: `ceil` trifft 18 von
// 18, `round` verfehlt drei. Festgehalten als `TestDieBreiteTrifftWx` hinter
// `fremdpakete`.
func TestDieVorsilbenbreiteRastetJeAbschnitt(t *testing.T) {
	// **Eine Schrift mit krummer Breite**, damit die Rasterung sichtbar wird:
	// 0,17 Punkt je Zeichen bei Grad 10 sind 1,7 Punkt — kein Vielfaches von ¼.
	schrift := gleichbreit{breite: 0.17}

	gesetzt := render.SetzeAbschnitte(
		render.Satz{Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben},
		render.Rechteck{Breite: 500, Hoehe: 40},
		[]render.Abschnitt{{Text: "aa bb cc"}},
		render.Schriftsatz{Aufrecht: schrift}, nil)

	if len(gesetzt.Zeilen) != 1 {
		t.Fatalf("%d Zeilen", len(gesetzt.Zeilen))
	}

	// **Die Summe der Stückbreiten ist die gerasterte Gesamtbreite**, nicht
	// die Summe gerasterter Stücke.
	//
	// Roh: acht Zeichen × 1,7 = 13,6. Gerastert: ⌈13,6 · 4⌉ / 4 = 13,75.
	//
	// Je Stück gerastert käme etwas anderes heraus: „aa" 3,4 → 3,5; „ " 1,7 →
	// 1,75; „bb" 3,4 → 3,5; „ " → 1,75; „cc" → 3,5 — zusammen 14,00.
	summe := 0.0
	for _, stueck := range gesetzt.Zeilen[0].Stuecke {
		summe += stueck.Breite
	}

	if summe != 13.75 {
		t.Errorf("die Zeile ist %g breit, erwartet 13.75 — je Stück gerastert "+
			"wären es 14.00, ungerastert 13.6", summe)
	}
}

// **Und die Rasterung verschiebt den Umbruch.**
//
// Das ist die Folge, die im Issue an erster Stelle steht: „Ein Umbruch, der
// eine Zeile früher oder später fällt, ist kein Viertelpunkt Unterschied,
// sondern eine ganze Zeile."
//
// Hier passt der Text ungerastert (13,6) in einen Kasten von 13,7 — gerastert
// (13,75) nicht mehr.
func TestDieRasterungVerschiebtDenUmbruch(t *testing.T) {
	gesetzt := render.SetzeAbschnitte(
		render.Satz{Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben},
		render.Rechteck{Breite: 13.7, Hoehe: 60},
		[]render.Abschnitt{{Text: "aa bb cc"}},
		render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.17}}, nil)

	if len(gesetzt.Zeilen) < 2 {
		t.Errorf("%d Zeile(n) — ungerastert passte alles in eine, gerastert nicht",
			len(gesetzt.Zeilen))
	}
}

// **Ein Abschnittswechsel setzt zurück.**
//
// Bei MSE ist jeder Schriftlauf ein eigenes `FontTextElement`, und `prev_width`
// beginnt darin bei null (`font.cpp`, Zeile 31). Kursiv neben aufrecht sind
// zwei Elemente — also zwei Rasterungen.
//
// Roh je Abschnitt: vier Zeichen × 1,7 = 6,8. Gerastert: ⌈27,2⌉/4 = 6,8 —
// trifft zufällig. Deshalb hier fünf Zeichen: 8,5 → ⌈34⌉/4 = 8,5. Auch das
// trifft; genommen sind darum drei: 5,1 → ⌈20,4⌉/4 = 5,25, und zweimal 5,25
// sind 10,5 gegen 10,25 für einen Lauf aus sechs Zeichen.
func TestEinAbschnittswechselSetztDieVorsilbeZurueck(t *testing.T) {
	gesetzt := render.SetzeAbschnitte(
		render.Satz{Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben},
		render.Rechteck{Breite: 500, Hoehe: 40},
		[]render.Abschnitt{{Text: "aaa"}, {Text: "bbb", Kursiv: true}},
		render.Schriftsatz{
			Aufrecht: gleichbreit{breite: 0.17},
			Kursiv:   gleichbreit{breite: 0.17},
		}, nil)

	if len(gesetzt.Zeilen) != 1 {
		t.Fatalf("%d Zeilen", len(gesetzt.Zeilen))
	}

	summe := 0.0
	for _, stueck := range gesetzt.Zeilen[0].Stuecke {
		summe += stueck.Breite
	}

	if summe != 10.5 {
		t.Errorf("zwei Abschnitte sind zusammen %g breit, erwartet 10.5 — "+
			"als ein Lauf gerechnet wären es 10.25", summe)
	}
}

// **Ein gestauchter Symbolzug wird über seine Bilder gemessen, nicht über eine
// Vorsilbe.**
//
// Symbole haben keine Glyphenbreiten; bei MSE sind sie ein
// `SymbolTextElement`, das seine Maße von der Symbolschrift bekommt
// (`src/render/text/symbol.cpp`, Zeilen 22–26, Fassung f9d9356 vom
// 2021-07-06). Der Weg des gestauchten Zugs muss das ebenso trennen wie der
// des umbrechenden — sonst bekäme ein Manakostenfeld mit `shrink` die Breite
// null.
func TestEinGestauchterSymbolzugWirdUeberSeineBilderGemessen(t *testing.T) {
	gesetzt := render.SetzeAbschnitte(
		render.Satz{
			Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben,
			Stauchen: true, Einzeilig: true, NurBeiUeberlauf: true,
		},
		render.Rechteck{Breite: 100, Hoehe: 40},
		[]render.Abschnitt{{
			Text:    "W",
			Symbole: []render.Symbol{{Bild: render.Blatt("w.png")}},
			Symbolmasse: render.Symbolmasse{
				Groesse: 20, Bildgroesse: 20, AbstandWaagerecht: 2,
			},
		}},
		render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}},
		bilder{"w.png": vollflaeche(20, 20)})

	if len(gesetzt.Zeilen) != 1 || len(gesetzt.Zeilen[0].Stuecke) != 1 {
		t.Fatalf("gesetzt wurde %+v", gesetzt.Zeilen)
	}

	// 20 Punkte Bild plus 2 Punkte Abstand.
	if breite := gesetzt.Zeilen[0].Stuecke[0].Breite; breite != 22 {
		t.Errorf("das Symbolstück ist %g breit, erwartet 22", breite)
	}
}
