package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Palette kommt gerastert im Frontend an**
// (#564).
//
// # Warum als Bild und nicht als Code
//
// Weil ein Code kein Symbol ist. `T` ist das Tap-Symbol nur für den, der es
// weiß; die Palette zeigt das Bild, das eingefügt wird — dasselbe, das MSE als
// 16×16-Kachel neben den Menüeintrag setzt (`InsertSymbolMenu::makeMenuItem`,
// `src/data/symbol_font.cpp`, Zeile 515, Fassung f9d9356 vom 2021-07-06).
func TestDiePaletteKommtGerastertAn(t *testing.T) {
	anwendung := mitPalette(&palettenbauer{eintraege: []service.Paletteneintrag{
		{Art: msepaket.MenueCode, Code: "W", Beschriftung: "W", Plan: einsymbolplan()},
		{Art: msepaket.MenueTrenner},
		{Art: msepaket.MenueUntermenue, Beschriftung: "Halb", Kinder: []service.Paletteneintrag{
			{Art: msepaket.MenueCode, Code: "1/2", Beschriftung: "1/2", Plan: einsymbolplan()},
		}},
	}, roh: true})

	antwort := anwendung.Symbolpalette(t.Context(), "1", "kosten")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if !antwort.Roh {
		t.Error("`always symbol:` kam nicht durch")
	}

	if len(antwort.Eintraege) != 3 {
		t.Fatalf("%d Einträge", len(antwort.Eintraege))
	}

	erster := antwort.Eintraege[0]
	if erster.Art != "code" || erster.Code != "W" || erster.Beschriftung != "W" {
		t.Errorf("%+v", erster)
	}

	if erster.Bild == "" {
		t.Error("der erste Eintrag hat kein Bild")
	}

	// **Ein Trenner hat keins**, und ein Untermenü auch nicht — sie zeigen
	// nichts, sie ordnen.
	if antwort.Eintraege[1].Bild != "" {
		t.Errorf("ein Trenner trägt ein Bild: %+v", antwort.Eintraege[1])
	}

	kinder := antwort.Eintraege[2].Kinder
	if len(kinder) != 1 || kinder[0].Bild == "" {
		t.Errorf("das Untermenü ist %+v", antwort.Eintraege[2])
	}
}

// **Ein Feld ohne Symbolschrift bekommt eine leere Liste, keinen Fehler.**
//
// Und eine **leere Liste** statt `nil`: über die Wails-Bindung stünde dort
// sonst `null`, und das Frontend prüfte auf zwei Dinge statt auf eines.
func TestEinFeldOhnePaletteMeldetKeinenFehler(t *testing.T) {
	antwort := mitPalette(&palettenbauer{}).Symbolpalette(probekontext(t), "1", "name")

	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Eintraege == nil || len(antwort.Eintraege) != 0 {
		t.Errorf("%+v", antwort.Eintraege)
	}
}

// **Eine unlesbare Kennung wird abgewiesen** — dieselbe Wache wie überall.
func TestDiePaletteWeistEineUnlesbareKennungAb(t *testing.T) {
	if antwort := mitPalette(&palettenbauer{}).
		Symbolpalette(probekontext(t), "keine zahl", "kosten"); antwort.Fehler == nil {
		t.Error("eine unlesbare Kennung kam durch")
	}
}

// palettenbauer gibt eine vorgegebene Palette.
type palettenbauer struct {
	ohneWortlisten
	ohneWahlbilder

	eintraege []service.Paletteneintrag
	roh       bool
	fehler    error
}

func (p *palettenbauer) Kartenplan(context.Context, *domain.CardRow) (
	*render.Plan, []domain.Befund, error,
) {
	return &render.Plan{Breite: 10, Hoehe: 10}, nil, nil
}

func (p *palettenbauer) Kartenpalette(context.Context, *domain.CardRow, string) (
	[]service.Paletteneintrag, bool, []domain.Befund, error,
) {
	return p.eintraege, p.roh, nil, p.fehler
}

func mitPalette(bauer *palettenbauer) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitEditor(&schreibprobe{}, karteprobe{karte: &domain.CardRow{ID: 1}},
		bauer, render.NeuerRasterer(ohneBilder{}))

	return anwendung
}

// einsymbolplan ist ein Plan, der eine Fläche malt — er steht für das Bild
// eines Symbols, ohne eine Bildquelle zu brauchen.
func einsymbolplan() *render.Plan {
	return &render.Plan{
		Breite: service.Palettengroesse, Hoehe: service.Palettengroesse,
		Ebenen: []render.Ebene{{
			Art:    render.ArtFlaeche,
			Rahmen: render.Rechteck{Breite: service.Palettengroesse, Hoehe: service.Palettengroesse},
			Farbe:  render.Farbe{R: 200, G: 40, B: 40, A: 255}, Deckung: 1,
		}},
	}
}

// **Ein Fehler beim Auflösen des Stils wird gemeldet, nicht verschluckt** (§8).
func TestDiePaletteMeldetEinenFehlerBeimAufloesen(t *testing.T) {
	anwendung := mitPalette(&palettenbauer{fehler: errors.New("die Platte ist voll")})

	if antwort := anwendung.Symbolpalette(t.Context(), "1", "kosten"); antwort.Fehler == nil {
		t.Error("der Fehler kam nicht durch")
	}
}

// **Eine Karte, die es nicht gibt, ist ein Befund.**
func TestDiePaletteMeldetEineFehlendeKarte(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitEditor(&schreibprobe{},
		karteprobe{fehler: domain.ErrNotFound}, &palettenbauer{},
		render.NeuerRasterer(ohneBilder{}))

	if antwort := anwendung.Symbolpalette(t.Context(), "1", "kosten"); antwort.Fehler == nil {
		t.Error("die fehlende Karte kam nicht durch")
	}
}

// **Ein Symbol, das sich nicht zeichnen lässt, ist ein Eintrag ohne Kachel.**
//
// Genau das zeigt MSE, wenn es keine findet (`InsertSymbolMenu::makeMenuItem`,
// `src/data/symbol_font.cpp`, Zeilen 508–518, Fassung f9d9356 vom
// 2021-07-06). Die ganze Palette daran scheitern zu lassen nähme sechzig
// brauchbare Schaltflächen für eine kaputte (§8).
func TestEinUnzeichenbaresSymbolLaesstDiePaletteStehen(t *testing.T) {
	// **Ein Plan ohne Größe.** Der Rasterer weist ihn ab — dieselbe Klasse wie
	// ein Symbol, dessen Bilddatei fehlt, und ohne eine Bildquelle zu brauchen.
	anwendung := mitPalette(&palettenbauer{eintraege: []service.Paletteneintrag{
		{Art: msepaket.MenueCode, Code: "W", Beschriftung: "W", Plan: &render.Plan{}},
		{Art: msepaket.MenueCode, Code: "U", Beschriftung: "U", Plan: einsymbolplan()},
	}})

	antwort := anwendung.Symbolpalette(t.Context(), "1", "kosten")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Eintraege) != 2 {
		t.Fatalf("%d Einträge", len(antwort.Eintraege))
	}

	if antwort.Eintraege[0].Bild != "" {
		t.Error("ein unzeichenbares Symbol trägt ein Bild")
	}

	if antwort.Eintraege[1].Bild == "" {
		t.Error("das brauchbare Symbol ist mit ausgefallen")
	}
}
