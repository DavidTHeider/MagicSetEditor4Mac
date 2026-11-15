package app

// Die **Symbolpalette** eines Feldes — gerastert, für die Oberfläche
// (#564).
//
// # Warum als Bild und nicht als Code
//
// Weil ein Code kein Symbol ist. `T` ist das Tap-Symbol nur für den, der es
// weiß. MSE setzt neben jeden Menüeintrag eine 16×16-Kachel
// (`InsertSymbolMenu::makeMenuItem`, `src/data/symbol_font.cpp`, Zeile 515,
// Fassung f9d9356 vom 2021-07-06), und hier ist es dasselbe Bild aus derselben
// Rasterung wie auf der Karte.

import (
	"bytes"
	"context"
	"encoding/base64"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// PalettenEintragDTO ist ein Eintrag der Symbolpalette.
type PalettenEintragDTO struct {
	// Art ist `code`, `custom`, `line` oder `submenu` — **unübersetzt** (§9):
	// die Wörter stehen so im Paket, und das Frontend schaltet darauf.
	Art string `json:"art"`

	// Code ist, was eingefügt wird. Leer bei Trenner und Untermenü.
	Code string `json:"code"`

	Beschriftung string `json:"beschriftung"`

	// Rueckfrage ist der Satz über dem Eingabefeld eines `custom`-Eintrags.
	Rueckfrage string `json:"rueckfrage"`

	// Bild ist ein PNG als Data-URI. Leer, wo es keins gibt — bei Trennern,
	// Untermenüs und jedem Code, den die Schrift gerade nicht kennt.
	Bild string `json:"bild"`

	Kinder []PalettenEintragDTO `json:"kinder"`
}

// SymbolpaletteAntwort trägt die Palette oder einen Fehler.
type SymbolpaletteAntwort struct {
	Eintraege []PalettenEintragDTO `json:"eintraege"`

	// Roh sagt, ob ein Code ohne `<sym>…</sym>` eingefügt wird —
	// `always symbol:` am Feldstil ([service.FuegtRohEin]).
	Roh bool `json:"roh"`

	Fehler *Fehler `json:"fehler"`
}

// Symbolpalette gibt, was ein Feld zum Einfügen anbietet.
//
// **Eine leere Liste und kein `nil`** bei einem Feld ohne Symbolschrift: über
// die Wails-Bindung stünde dort sonst `null`, und das Frontend prüfte auf zwei
// Dinge statt auf eines.
func (a *App) Symbolpalette(ctx context.Context, karteID, feld string) SymbolpaletteAntwort {
	ctx = a.imProfil(ctx)

	leer := []PalettenEintragDTO{}

	kennung, befund := kennungAus(karteID)
	if befund != nil {
		return SymbolpaletteAntwort{Eintraege: leer, Fehler: befund}
	}

	karte, err := a.kartenleser.KarteLesen(ctx, kennung)
	if err != nil {
		return SymbolpaletteAntwort{Eintraege: leer, Fehler: Uebersetze(ctx, err)}
	}

	eintraege, roh, _, err := a.planbauer.Kartenpalette(ctx, karte, feld)
	if err != nil {
		return SymbolpaletteAntwort{Eintraege: leer, Fehler: Uebersetze(ctx, err)}
	}

	// **Die Befunde des Planbaus bleiben hier.** Sie gehören zur Karte und
	// stehen an der Vorschau; sie ein zweites Mal an der Palette zu zeigen
	// hieße, dieselbe Meldung zweimal zu melden.
	return SymbolpaletteAntwort{Eintraege: a.palettenDTO(ctx, eintraege), Roh: roh}
}

// palettenDTO rastert jeden Eintrag und macht ihn zum DTO.
func (a *App) palettenDTO(
	ctx context.Context, eintraege []service.Paletteneintrag,
) []PalettenEintragDTO {
	umgesetzt := make([]PalettenEintragDTO, 0, len(eintraege))

	for i := range eintraege {
		umgesetzt = append(umgesetzt, PalettenEintragDTO{
			Art:          string(eintraege[i].Art),
			Code:         eintraege[i].Code,
			Beschriftung: eintraege[i].Beschriftung,
			Rueckfrage:   eintraege[i].Rueckfrage,
			Bild:         a.symbolbild(ctx, eintraege[i].Plan),
			Kinder:       a.palettenDTO(ctx, eintraege[i].Kinder),
		})
	}

	return umgesetzt
}

// symbolbild rastert einen Symbolplan zu einem Data-URI.
//
// **Ein Fehler ergibt ein leeres Bild und keinen Fehler der Antwort.** Ein
// Symbol, das sich nicht zeichnen lässt, ist ein Eintrag ohne Kachel — genau
// das, was MSE zeigt, wenn es keine findet (`makeMenuItem`, Zeilen 508–518).
// Die ganze Palette daran scheitern zu lassen, nähme sechzig brauchbare
// Schaltflächen für eine kaputte (§8).
func (a *App) symbolbild(ctx context.Context, plan *render.Plan) string {
	if plan == nil {
		return ""
	}

	// Die Palette zeigt Symbole, keine Karte — Parität genügt.
	bild, _, err := a.rasterer.Zeichne(ctx, plan, Palettenvergroesserung,
		render.GueteMittel)
	if err != nil {
		return ""
	}

	var gepackt bytes.Buffer
	if err := a.packen(&gepackt, bild); err != nil {
		return ""
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(gepackt.Bytes())
}

// Palettenvergroesserung ist die Auflösung, mit der eine Kachel entsteht.
//
// **Zwei und nicht eins**, weil ein Bildschirm mit doppelter Punktdichte der
// Normalfall ist: eine 16×16-Kachel darauf sähe weich aus, und ein Manasymbol
// lebt von seinem Rand.
const Palettenvergroesserung = 2
