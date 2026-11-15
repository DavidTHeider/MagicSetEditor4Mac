// **Auszählen nach einer erklärten Dimension, im Fenster** — Werkstück C von
// Phase 4.11 (#1338).

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Auszaehlquelle ist, was das Fenster für die Auszählung braucht.
type Auszaehlquelle interface {
	Angebot(ctx context.Context, setID int64) ([]AuszaehldimensionDTO, error)
	Auszaehlen(ctx context.Context, setID int64, name string) (*AuszaehlungDTO, error)
}

// MitAuszaehlung setzt die Naht.
func (a *App) MitAuszaehlung(quelle Auszaehlquelle) { a.auszaehlung = quelle }

// Auszaehldimensionsantwort trägt die Dimensionen des Spielpakets.
type Auszaehldimensionsantwort struct {
	Dimensionen []AuszaehldimensionDTO `json:"dimensionen"`
	Fehler      *Fehler                `json:"fehler"`
}

// Auszaehlantwort trägt eine Auszählung.
//
// **Flach und nicht um ein DTO herum** — dieselbe Überlegung wie bei
// [Kartenskriptlaufantwort]: ein geschachtelter Zeiger käme im Fehlerfall als
// `null` an, und die Seite hätte einen Zweig zu bedienen, den sie nie
// erreicht.
type Auszaehlantwort struct {
	Dimension    string              `json:"dimension"`
	Beschreibung string              `json:"beschreibung"`
	Numerisch    bool                `json:"numerisch"`
	Gruppen      []AuszaehlgruppeDTO `json:"gruppen"`
	Befunde      []string            `json:"befunde"`
	Fehler       *Fehler             `json:"fehler"`
}

// Auszaehldimensionen nennt die Dimensionen eines Sets.
func (a *App) Auszaehldimensionen(ctx context.Context, setID string) Auszaehldimensionsantwort {
	ctx = a.imProfil(ctx)

	leer := Auszaehldimensionsantwort{Dimensionen: []AuszaehldimensionDTO{}}

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	if a.auszaehlung == nil {
		leer.Fehler = keineAuszaehlung()

		return leer
	}

	dimensionen, err := a.auszaehlung.Angebot(ctx, kennung)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	if dimensionen == nil {
		dimensionen = []AuszaehldimensionDTO{}
	}

	return Auszaehldimensionsantwort{Dimensionen: dimensionen}
}

// Auszaehlen zählt ein Set nach der benannten Dimension aus.
func (a *App) Auszaehlen(ctx context.Context, setID, name string) Auszaehlantwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return keineGruppen(befund)
	}

	if a.auszaehlung == nil {
		return keineGruppen(keineAuszaehlung())
	}

	auszaehlung, err := a.auszaehlung.Auszaehlen(ctx, kennung, name)
	if err != nil {
		return keineGruppen(Uebersetze(ctx, err))
	}

	return Auszaehlantwort{
		Dimension:    auszaehlung.Dimension,
		Beschreibung: auszaehlung.Beschreibung,
		Numerisch:    auszaehlung.Numerisch,
		Gruppen:      auszaehlung.Gruppen,
		Befunde:      auszaehlung.Befunde,
	}
}

// Auszaehldimensionen nennt die Dimensionen eines Sets.
func (b *Bindungen) Auszaehldimensionen(ctx context.Context, setID string) Auszaehldimensionsantwort {
	return b.app(ctx).Auszaehldimensionen(b.ruf(ctx), setID)
}

// Auszaehlen zählt ein Set nach der benannten Dimension aus.
func (b *Bindungen) Auszaehlen(ctx context.Context, setID, name string) Auszaehlantwort {
	return b.app(ctx).Auszaehlen(b.ruf(ctx), setID, name)
}

// keineGruppen ist die Antwort, die nichts gezählt hat.
func keineGruppen(befund *Fehler) Auszaehlantwort {
	return Auszaehlantwort{
		Gruppen: []AuszaehlgruppeDTO{}, Befunde: []string{}, Fehler: befund,
	}
}

// keineAuszaehlung sagt, dass die Naht fehlt.
func keineAuszaehlung() *Fehler {
	return Nutzerfehler(fmt.Errorf(
		"%w: die Auszählung ist nicht eingerichtet", domain.ErrValidation))
}
