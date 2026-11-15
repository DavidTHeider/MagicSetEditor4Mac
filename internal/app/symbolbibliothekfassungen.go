package app

// **Ein Symbol der Bibliothek in den Fassungen eines Sets** — DOM-02.AK-02
// (#1697):
// die Größen- und Farbvorschau.
//
// **Gezeichnet vom Rasterer der Karten**: der Dienst baut je Fassung einen
// Plan über denselben Bildknoten wie auf der Karte, hier wird er in drei
// Größen gerastert. **Nur am Fenster** — ein Bild in einer Textantwort liest
// niemand, dieselbe Abwägung wie bei der Symbolpalette.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Fassungsgroessen sind die Kantenlängen der Vorschau in Punkten — klein wie
// an der Typzeile, mittel und groß wie im Editor.
var Fassungsgroessen = []int{16, 32, 64}

// fassungsdichte ist die Vergrößerung je Punkt — **zwei**, für Bildschirme
// mit doppelter Dichte.
const fassungsdichte = 2

// FassungsbildDTO ist ein Symbol in einer Fassung und einer Größe.
type FassungsbildDTO struct {
	// Groesse ist die Kantenlänge in Punkten.
	Groesse int `json:"groesse"`

	// Vorschau ist ein PNG als Data-URI — leer, wo es sich nicht zeichnen
	// oder packen ließ.
	Vorschau string `json:"vorschau"`
}

// SymbolfassungDTO ist eine Fassung mit ihren Bildern.
type SymbolfassungDTO struct {
	Name   string            `json:"name"`
	Bilder []FassungsbildDTO `json:"bilder"`
}

// SymbolfassungenAntwort trägt die Fassungen oder einen Fehler.
type SymbolfassungenAntwort struct {
	Fassungen []SymbolfassungDTO `json:"fassungen"`
	Fehler    *Fehler            `json:"fehler"`
}

// BibliothekssymbolFassungen zeichnet ein Symbol in den Fassungen eines Sets.
func (a *App) BibliothekssymbolFassungen(ctx context.Context, id, setID string) SymbolfassungenAntwort {
	ctx = a.imProfil(ctx)
	leer := SymbolfassungenAntwort{Fassungen: []SymbolfassungDTO{}}

	kennung, befund := benannteKennungAus("id", id)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	setKennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	if a.symbolbibliothek == nil {
		leer.Fehler = keineSymbolbibliothek()

		return leer
	}

	if a.rasterer == nil {
		leer.Fehler = Uebersetze(ctx, fmt.Errorf("%w: hier zeichnet niemand", domain.ErrValidation))

		return leer
	}

	plaene, err := a.symbolbibliothek.Fassungsplaene(ctx, kennung, setKennung, 1)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	for _, plan := range plaene {
		leer.Fassungen = append(leer.Fassungen, SymbolfassungDTO{Name: plan.Name, Bilder: a.fassungsbilder(ctx, plan.Plan)})
	}

	return leer
}

// fassungsbilder rastert einen Plan in jeder Größe.
//
// **Ein Plan von einem Punkt, und die Größe über die Vergrößerung** — so
// entsteht jede Kante aus derselben Zeichnung.
func (a *App) fassungsbilder(ctx context.Context, plan *render.Plan) []FassungsbildDTO {
	bilder := make([]FassungsbildDTO, 0, len(Fassungsgroessen))

	for _, groesse := range Fassungsgroessen {
		vorschau := ""

		if bild, _, err := a.rasterer.Zeichne(ctx, plan, float64(groesse*fassungsdichte), render.GueteMittel); err == nil {
			vorschau, _ = a.alsDatenURI(bild) //nolint:errcheck // leer heißt: ohne Bild, wie bei der Palette
		}

		bilder = append(bilder, FassungsbildDTO{Groesse: groesse, Vorschau: vorschau})
	}

	return bilder
}
