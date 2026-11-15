package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **`symbol_variation` rastert das Setsymbol in einer benannten Fassung.**
//
//	FOR_EACH(v, style->variations) {
//	  if (v->name == variation) {
//	    return make_intrusive<SymbolToImage>(value, filename, value->last_update, v);
//	  }
//	}
//	throw ScriptError(_("Variation of symbol not found ('") + variation + _("')"));
//
// `src/script/functions/image.cpp`, Zeilen 141–170, Fassung f9d9356 vom
// 2021-07-06.
//
// **Daran hängt das Seltenheitssymbol jeder Karte** — 180 Befunde über die 190
// Stile von `M15-Magic-Pack`
// (#323).
func TestSymbolVariationRastertDasSetsymbol(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Symbolfassungen: []msepaket.Symbolfassung{{
				Feld: "symbol",
				Name: "common", Randbreite: 0.1, Fuellart: "solid",
				Fuellfarbe: "rgb(0,0,0)", Randfarbe: "rgb(255,255,255)",
			}},
			Felder: []msepaket.Feldstil{{
				Feld: "symbol", Links: "10", Oben: "10", Breite: "20", Hoehe: "20",
				Darstellung: "image",
				Bild:        `{ symbol_variation(symbol: set.symbol, variation: "common") }`,
			}},
		},
		Setfelder: []domain.Field{{Name: "symbol", Kind: domain.KindText}},
		Setangaben: []domain.Value{
			{Field: "symbol", Text: "/probe.mse-set/symbol.mse-symbol"},
		},
		Kartenfelder: []domain.Field{{Name: "symbol", Kind: domain.KindText}},
	}

	karte := &domain.CardRow{ID: 1, Values: []domain.Value{{Field: "symbol", Text: "x"}}}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "symbol_variation") {
			t.Errorf("`symbol_variation` ist unbekannt: %s", befund.Meldung)
		}
	}

	if len(plan.Ebenen) < 2 {
		t.Fatalf("%d Ebenen", len(plan.Ebenen))
	}

	baum := plan.Ebenen[1].Bild
	if baum.Knotenart != "symbol_variation" {
		t.Fatalf("die Knotenart ist %q", baum.Knotenart)
	}

	if baum.Texte["symbol"] != "/probe.mse-set/symbol.mse-symbol" {
		t.Errorf("der Symbolpfad ist %q", baum.Texte["symbol"])
	}

	if baum.Zahlen["border_radius"] != 0.1 {
		t.Errorf("die Randbreite ist %g", baum.Zahlen["border_radius"])
	}

	if baum.Farben["fill_color"].A == 0 {
		t.Errorf("die Füllfarbe fehlt: %+v", baum.Farben)
	}
}

// **Eine Fassung, die es nicht gibt, ist ein Befund.**
//
//	throw ScriptError(_("Variation of symbol not found ('") + variation + _("')"));
//
// Zeile 169. Eine erfundene Fassung zeichnete ein Symbol in falschen Farben,
// und niemand sähe warum (§8).
func TestEineUnbekannteFassungIstEinBefund(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Felder: []msepaket.Feldstil{{
				Feld: "symbol", Links: "10", Oben: "10", Breite: "20", Hoehe: "20",
				Darstellung: "image",
				Bild:        `{ symbol_variation(symbol: set.symbol, variation: "gibtsnicht") }`,
			}},
		},
		Setfelder:  []domain.Field{{Name: "symbol", Kind: domain.KindText}},
		Setangaben: []domain.Value{{Field: "symbol", Text: "/probe.mse-set/symbol.mse-symbol"}},
	}

	_, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "gibtsnicht") {
		t.Errorf("Befunde: %v", befunde)
	}
}

// **Die Gestalt der `.mse-symbol` reist mit dem Plan.**
//
// Der Rasterer öffnet keine Pakete (§2, Vermerk 0022); wer die Datei liest, ist
// diese Schicht. Was ankommt, ist die Zeichnung — Formen, Punkte, Symmetrien
// (#323).
func TestDieGestaltDerSymboldateiReistMit(t *testing.T) {
	wurzel := symbolwurzel(t)

	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Symbolfassungen: []msepaket.Symbolfassung{{
				Feld: "symbol",
				Name: "common", Randbreite: 0.1, Fuellart: "solid",
				Fuellfarbe: "rgb(0,0,0)", Randfarbe: "rgb(255,255,255)",
			}},
			Felder: []msepaket.Feldstil{{
				Feld: "symbol", Links: "10", Oben: "10", Breite: "20", Hoehe: "20",
				Darstellung: "image",
				Bild:        `{ symbol_variation(symbol: set.symbol, variation: "common") }`,
			}},
		},
		Setfelder:  []domain.Field{{Name: "symbol", Kind: domain.KindText}},
		Setangaben: []domain.Value{{Field: "symbol", Text: "/probe.mse-set/zeichen.mse-symbol"}},
	}

	plan, befunde, err := service.NeuerPlanbauer().
		MitBildern(msepaket.NeueBildquelle(wurzel)).
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(plan.Ebenen) < 2 {
		t.Fatalf("%d Ebenen", len(plan.Ebenen))
	}

	baum := plan.Ebenen[1].Bild
	if len(baum.Symbol) != 1 || len(baum.Symbol[0].Punkte) != 4 {
		t.Fatalf("die Gestalt fehlt: %+v", baum.Symbol)
	}
}

// **Eine Symboldatei, die es nicht gibt, wird beim Rastern benannt.**
//
// **Nicht beim Planbauen**, und das ist die Arbeitsteilung: der Plan sagt, was
// gezeichnet werden soll, der Rasterer, was davon ging. Die Ebene steht
// deshalb da — mit leerer Gestalt —, und der Befund nennt die Datei (§8).
func TestEineFehlendeSymboldateiWirdBeimRasternBenannt(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Symbolfassungen: []msepaket.Symbolfassung{{Feld: "symbol", Name: "common", Fuellart: "solid"}},
			Felder: []msepaket.Feldstil{{
				Feld: "symbol", Links: "10", Oben: "10", Breite: "20", Hoehe: "20",
				Darstellung: "image",
				Bild:        `{ symbol_variation(symbol: set.symbol, variation: "common") }`,
			}},
		},
		Setfelder:  []domain.Field{{Name: "symbol", Kind: domain.KindText}},
		Setangaben: []domain.Value{{Field: "symbol", Text: "/probe.mse-set/fehlt.mse-symbol"}},
	}

	plan, befunde, err := service.NeuerPlanbauer().
		MitBildern(msepaket.NeueBildquelle(t.TempDir())).
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde beim Planbauen: %v", befunde)
	}

	if len(plan.Ebenen) < 2 || len(plan.Ebenen[1].Bild.Symbol) != 0 {
		t.Fatalf("die Ebene fehlt oder trägt eine Gestalt: %+v", plan.Ebenen)
	}

	_, rasterbefunde, err := render.NeuerRasterer(msepaket.NeueBildquelle(t.TempDir())).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(rasterbefunde) != 1 || !strings.Contains(rasterbefunde[0].Meldung, "fehlt.mse-symbol") {
		t.Errorf("Rasterbefunde: %v", rasterbefunde)
	}
}

// symbolwurzel legt ein Setpaket mit einer Symboldatei an.
func symbolwurzel(t *testing.T) string {
	t.Helper()

	wurzel := t.TempDir()
	paket := filepath.Join(wurzel, "probe.mse-set")

	if err := os.MkdirAll(paket, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := `mse version: 2.0.0
part:
	type: shape
	name: viereck
	point:
		position: (0.2,0.2)
	point:
		position: (0.8,0.2)
	point:
		position: (0.8,0.8)
	point:
		position: (0.2,0.8)
`

	if err := os.WriteFile(filepath.Join(paket, "zeichen.mse-symbol"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return wurzel
}

// **Der Symbolpfad darf auch ohne Namen dastehen.**
//
// MSE liest `SCRIPT_PARAM(ScriptValueP, symbol)` — benannt. Die Vorlagen
// schreiben es so; ein Stellungsargument ist die Form, die eine eigene Vorlage
// nehmen könnte, und sie kostet nichts.
func TestDerSymbolpfadDarfOhneNamenDastehen(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Symbolfassungen: []msepaket.Symbolfassung{{Feld: "symbol", Name: "common", Fuellart: "solid"}},
			Felder: []msepaket.Feldstil{{
				Feld: "symbol", Links: "10", Oben: "10", Breite: "20", Hoehe: "20",
				Darstellung: "image",
				Bild:        `{ symbol_variation("/probe.mse-set/z.mse-symbol", variation: "common") }`,
			}},
		},
	}

	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(plan.Ebenen) < 2 ||
		plan.Ebenen[1].Bild.Texte["symbol"] != "/probe.mse-set/z.mse-symbol" {
		t.Errorf("der Symbolpfad fehlt: %+v", plan.Ebenen)
	}
}
