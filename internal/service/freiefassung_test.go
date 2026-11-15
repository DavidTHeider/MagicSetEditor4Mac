package service_test

// **`symbol_variation` ohne `variation:` baut eine freie Fassung** aus seinen
// Argumenten — der zweite Zweig der Funktion:
//
//	// custom variation
//	SCRIPT_PARAM(double, border_radius);
//	SCRIPT_OPTIONAL_PARAM_(String, fill_type);
//	if (fill_type == _("solid") || fill_type.empty()) { … fill_color, border_color … }
//	else if (fill_type == _("linear gradient")) { … _1, _2, center_x/y, end_x/y … }
//	else if (fill_type == _("radial gradient")) { … _1, _2 … }
//	else throw ScriptError(_("Unknown fill type for symbol_variation: ") + fill_type);
//
// `src/script/functions/image.cpp`, Zeilen 171–200, Fassung f9d9356 vom
// 2021-07-06. **Zwölf Aufrufe im Pack** schreiben so, die meisten mit einem
// Verlauf; hier scheiterten sie an „kennt die Fassung nicht", und die Karte
// zeigte kein Seltenheitssymbol (gefunden am Orakel zu #2303).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func freieFassung(t *testing.T, aufruf string) (render.Bildbaum, []domain.Befund) {
	t.Helper()

	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Felder: []msepaket.Feldstil{{
				Feld: "symbol", Links: "10", Oben: "10", Breite: "20", Hoehe: "20",
				Darstellung: "image", Bild: "{ " + aufruf + " }",
			}},
		},
		Setfelder:    []domain.Field{{Name: "symbol", Kind: domain.KindText}},
		Setangaben:   []domain.Value{{Field: "symbol", Text: "/probe.mse-set/symbol.mse-symbol"}},
		Kartenfelder: []domain.Field{{Name: "symbol", Kind: domain.KindText}},
	}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(),
		&domain.CardRow{ID: 1, Values: []domain.Value{{Field: "symbol", Text: "x"}}}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for i := range plan.Ebenen {
		if plan.Ebenen[i].Art == render.ArtBild {
			return plan.Ebenen[i].Bild, befunde
		}
	}

	return render.Bildbaum{}, befunde
}

func TestEineFreieFassungTraegtIhreArgumente(t *testing.T) {
	t.Run("voll", func(t *testing.T) {
		baum, befunde := freieFassung(t,
			`symbol_variation(symbol: set.symbol, border_radius: 0.05, fill_color: rgb(10,20,30), border_color: rgb(40,50,60))`)

		if len(befunde) > 0 || baum.Knotenart != "symbol_variation" {
			t.Fatalf("Baum %q, Befunde %v", baum.Knotenart, befunde)
		}

		if baum.Zahlen["border_radius"] != 0.05 || baum.Texte["fill_type"] != "solid" ||
			baum.Farben["fill_color"] != (render.Farbe{R: 10, G: 20, B: 30, A: 255}) ||
			baum.Farben["border_color"] != (render.Farbe{R: 40, G: 50, B: 60, A: 255}) {
			t.Errorf("Zahlen %v, Texte %v, Farben %v", baum.Zahlen, baum.Texte, baum.Farben)
		}
	})

	// **Eine Farbe darf als Text kommen** (J05, #2269): MSE liest die
	// Argumente als `Color`, und aus einer Zeichenkette wird über
	// `parse_color` eine (`src/script/functions/image.cpp`, Zeile 178;
	// `src/script/value.cpp`, Zeile 354; `src/gfx/color.cpp`, Zeile 48,
	// Fassung f9d9356 vom 2021-07-06).
	t.Run("mit Farben als Text", func(t *testing.T) {
		baum, befunde := freieFassung(t,
			`symbol_variation(symbol: set.symbol, border_radius: 0.05, fill_color: "#000", border_color: "white")`)

		if len(befunde) > 0 ||
			baum.Farben["fill_color"] != (render.Farbe{A: 255}) ||
			baum.Farben["border_color"] != (render.Farbe{R: 255, G: 255, B: 255, A: 255}) {
			t.Errorf("Farben %v, Befunde %v", baum.Farben, befunde)
		}
	})

	t.Run("mit Verlauf", func(t *testing.T) {
		baum, befunde := freieFassung(t, `symbol_variation(symbol: set.symbol, border_radius: 0.15,
			fill_type: "linear gradient", fill_color_1: rgb(1,2,3), border_color_1: rgb(4,5,6),
			fill_color_2: rgb(7,8,9), border_color_2: rgb(10,11,12),
			center_x: 0.5, center_y: 0.25, end_x: 1, end_y: 0.75)`)

		if len(befunde) > 0 || baum.Texte["fill_type"] != "linear gradient" ||
			baum.Farben["fill_color_2"] != (render.Farbe{R: 7, G: 8, B: 9, A: 255}) ||
			baum.Zahlen["center_y"] != 0.25 || baum.Zahlen["end_x"] != 1 {
			t.Errorf("Zahlen %v, Texte %v, Farben %v, Befunde %v", baum.Zahlen, baum.Texte, baum.Farben, befunde)
		}
	})

	t.Run("mit Kreisverlauf", func(t *testing.T) {
		baum, befunde := freieFassung(t, `symbol_variation(symbol: set.symbol, border_radius: 0.1,
			fill_type: "radial gradient", fill_color_1: rgb(1,2,3), border_color_1: rgb(4,5,6),
			fill_color_2: rgb(7,8,9), border_color_2: rgb(10,11,12))`)

		if len(befunde) > 0 || baum.Texte["fill_type"] != "radial gradient" ||
			baum.Farben["border_color_1"] != (render.Farbe{R: 4, G: 5, B: 6, A: 255}) {
			t.Errorf("Texte %v, Farben %v, Befunde %v", baum.Texte, baum.Farben, befunde)
		}
	})
}

// **Was MSE2 verlangt, fehlt nicht still** — `SCRIPT_PARAM` wirft, und eine
// unbekannte Füllart ebenso.
func TestEineFreieFassungOhnePflichtangabeScheitert(t *testing.T) {
	for _, fall := range []struct{ name, aufruf, meldung string }{
		{name: "ohne Randbreite", aufruf: `symbol_variation(symbol: set.symbol, fill_color: rgb(0,0,0), border_color: rgb(0,0,0))`, meldung: "border_radius"},
		{name: "ohne Farbe", aufruf: `symbol_variation(symbol: set.symbol, border_radius: 0.1, fill_color: rgb(0,0,0))`, meldung: "border_color"},
		{name: "eine unbekannte Füllart", aufruf: `symbol_variation(symbol: set.symbol, border_radius: 0.1, fill_type: "karo")`, meldung: "karo"},
		{name: "ein Verlauf ohne Mitte", aufruf: `symbol_variation(symbol: set.symbol, border_radius: 0.1, fill_type: "linear gradient",
			fill_color_1: rgb(1,2,3), border_color_1: rgb(4,5,6), fill_color_2: rgb(7,8,9), border_color_2: rgb(10,11,12))`, meldung: "center_x"},
		{name: "ein Text, der keine Farbe ist", aufruf: `symbol_variation(symbol: set.symbol, border_radius: 0.1, fill_color: "keine Farbe", border_color: "#fff")`, meldung: "fill_color"},
		{name: "eine Randbreite, die keine Zahl ist", aufruf: `symbol_variation(symbol: set.symbol, border_radius: "breit")`, meldung: "keine Zahl"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			_, befunde := freieFassung(t, fall.aufruf)

			gefunden := false
			for _, befund := range befunde {
				gefunden = gefunden || strings.Contains(befund.Meldung, fall.meldung)
			}

			if !gefunden {
				t.Errorf("Befunde %v nennen %q nicht", befunde, fall.meldung)
			}
		})
	}
}
