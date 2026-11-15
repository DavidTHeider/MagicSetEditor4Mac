package app

// **Die Parameterhilfen der Schlüsselwörter** — Werkstück D von Phase 4.11
// (#1342).
//
// **Eine eigene Datei**, weil `schluesselwoerter.go` sonst über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort steht, **was**
// ein Schlüsselwort ist, hier, **womit** man seine Parameter schreibt.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// ParameterartDTO ist eine Parameterart, wie ein Menü sie anbietet.
type ParameterartDTO struct {
	Name         string `json:"name"`
	Beschreibung string `json:"beschreibung"`

	// Einfuegen ist der Text, der ins **Muster** kommt.
	Einfuegen string `json:"einfuegen"`
}

// ParameterverweisDTO ist eine Form, in der ein Erinnerungstext auf einen
// Parameter zeigt.
type ParameterverweisDTO struct {
	// Nummer ist die Stelle des Parameters im Muster, ab 1.
	Nummer int `json:"nummer"`

	// Art ist der Name der Parameterart.
	Art string `json:"art"`

	// Name ist der Name des Verweisskripts — leer bei der schlichten Form.
	Name         string `json:"name"`
	Beschreibung string `json:"beschreibung"`

	// Einfuegen ist der Text, der in den **Erinnerungstext** kommt.
	Einfuegen string `json:"einfuegen"`
}

// Parameterhilfenantwort trägt beide Menüs.
//
// **Flach und mit drei Listen, die nie `nil` sind** — dieselbe Zusage wie
// überall (§3).
type Parameterhilfenantwort struct {
	Arten    []ParameterartDTO     `json:"arten"`
	Verweise []ParameterverweisDTO `json:"verweise"`
	Befunde  []string              `json:"befunde"`
	Fehler   *Fehler               `json:"fehler"`
}

// Parameterhilfen gibt beide Menüs zu einem Schlüsselwort des Sets.
func (a *App) Parameterhilfen(
	ctx context.Context, setID, wortID string,
) Parameterhilfenantwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return keineHilfen(befund)
	}

	wort, wortbefund := benannteKennungAus("wortID", wortID)
	if wortbefund != nil {
		return keineHilfen(wortbefund)
	}

	if a.schluesselwoerter == nil {
		return keineHilfen(Uebersetze(ctx, fmt.Errorf(
			"%w: die Schlüsselwörter sind nicht eingerichtet", domain.ErrValidation)))
	}

	hilfen, err := a.schluesselwoerter.Parameterhilfen(ctx, kennung, wort)
	if err != nil {
		return keineHilfen(Uebersetze(ctx, err))
	}

	return Parameterhilfenantwort{
		Arten:    alsParameterarten(hilfen.Arten),
		Verweise: alsParameterverweise(hilfen.Verweise),
		Befunde:  befundtexte(hilfen.Befunde),
	}
}

// Parameterhilfen gibt beide Menüs zu einem Schlüsselwort des Sets.
func (b *Bindungen) Parameterhilfen(ctx context.Context, setID, wortID string) Parameterhilfenantwort {
	return b.app(ctx).Parameterhilfen(b.ruf(ctx), setID, wortID)
}

// keineHilfen ist die Antwort, die kein Menü hat.
func keineHilfen(befund *Fehler) Parameterhilfenantwort {
	return Parameterhilfenantwort{
		Arten:    []ParameterartDTO{},
		Verweise: []ParameterverweisDTO{},
		Befunde:  []string{},
		Fehler:   befund,
	}
}

// alsParameterarten übersetzt das Einfügeangebot.
func alsParameterarten(arten []service.Parametereinfuegung) []ParameterartDTO {
	gebaut := make([]ParameterartDTO, 0, len(arten))

	for _, art := range arten {
		gebaut = append(gebaut, ParameterartDTO{
			Name: art.Name, Beschreibung: art.Beschreibung, Einfuegen: art.Einfuegen,
		})
	}

	return gebaut
}

// alsParameterverweise übersetzt die Verweise.
func alsParameterverweise(
	verweise []service.Parameterverweis,
) []ParameterverweisDTO {
	gebaut := make([]ParameterverweisDTO, 0, len(verweise))

	for _, verweis := range verweise {
		gebaut = append(gebaut, ParameterverweisDTO{
			Nummer: verweis.Nummer, Art: verweis.Art,
			Name: verweis.Name, Beschreibung: verweis.Beschreibung,
			Einfuegen: verweis.Einfuegen,
		})
	}

	return gebaut
}
