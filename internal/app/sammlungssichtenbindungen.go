// **Die Bindungen der Sammlungssichten** — Werkstück 5 von Phase 4.5
// (#1292,
// Vermerk [0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sammlungssichtenquelle ist, was das Fenster für die vier Fragen braucht.
type Sammlungssichtenquelle interface {
	Besitz(ctx context.Context, printing string) (*BesitzDTO, error)
	Lagerorte(ctx context.Context) ([]LagerortDTO, error)
	Dubletten(ctx context.Context, behalten int) ([]DublettenDTO, error)
	Wuensche(ctx context.Context) ([]ExemplarDTO, error)
}

// MitSammlungssichten setzt die Naht.
func (a *App) MitSammlungssichten(quelle Sammlungssichtenquelle) {
	a.sammlungssichten = quelle
}

// Besitzantwort trägt den Besitzstand eines Drucks.
type Besitzantwort struct {
	Besitz *BesitzDTO `json:"besitz"`
	Fehler *Fehler    `json:"fehler"`
}

// Lagerortantwort trägt die Lagerorte.
type Lagerortantwort struct {
	Lagerorte []LagerortDTO `json:"lagerorte"`
	Fehler    *Fehler       `json:"fehler"`
}

// Dublettenantwort trägt die überzähligen Exemplare.
type Dublettenantwort struct {
	Dubletten []DublettenDTO `json:"dubletten"`
	Fehler    *Fehler        `json:"fehler"`
}

// SammlungBesitz sagt, was von einem Druck da ist.
func (a *App) SammlungBesitz(ctx context.Context, printing string) Besitzantwort {
	ctx = a.imProfil(ctx)

	if a.sammlungssichten == nil {
		return Besitzantwort{Fehler: Uebersetze(ctx, keineSichten())}
	}

	stand, err := a.sammlungssichten.Besitz(ctx, printing)
	if err != nil {
		return Besitzantwort{Fehler: Uebersetze(ctx, err)}
	}

	return Besitzantwort{Besitz: stand}
}

// SammlungLagerorte zählt die Lagerorte auf.
func (a *App) SammlungLagerorte(ctx context.Context) Lagerortantwort {
	ctx = a.imProfil(ctx)

	if a.sammlungssichten == nil {
		return Lagerortantwort{
			Lagerorte: []LagerortDTO{}, Fehler: Uebersetze(ctx, keineSichten()),
		}
	}

	orte, err := a.sammlungssichten.Lagerorte(ctx)
	if err != nil {
		return Lagerortantwort{Lagerorte: []LagerortDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	return Lagerortantwort{Lagerorte: orte}
}

// SammlungDubletten zeigt, wovon mehr da ist als gewollt.
func (a *App) SammlungDubletten(ctx context.Context, behalten int) Dublettenantwort {
	ctx = a.imProfil(ctx)

	if a.sammlungssichten == nil {
		return Dublettenantwort{
			Dubletten: []DublettenDTO{}, Fehler: Uebersetze(ctx, keineSichten()),
		}
	}

	dubletten, err := a.sammlungssichten.Dubletten(ctx, behalten)
	if err != nil {
		return Dublettenantwort{Dubletten: []DublettenDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	return Dublettenantwort{Dubletten: dubletten}
}

// SammlungWuensche gibt die Wunschliste.
//
// **Dieselbe Antwortgestalt wie die Sammlung**: eine Wunschzeile ist derselbe
// Verbund, nur mit dem anderen Vorzeichen (0057).
func (a *App) SammlungWuensche(ctx context.Context) Sammlungsantwort {
	ctx = a.imProfil(ctx)

	if a.sammlungssichten == nil {
		return Sammlungsantwort{
			Exemplare: []ExemplarDTO{}, Fehler: Uebersetze(ctx, keineSichten()),
		}
	}

	wuensche, err := a.sammlungssichten.Wuensche(ctx)
	if err != nil {
		return Sammlungsantwort{Exemplare: []ExemplarDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	return Sammlungsantwort{Exemplare: wuensche}
}

// SammlungBesitz sagt, was von einem Druck da ist.
func (b *Bindungen) SammlungBesitz(ctx context.Context, printing string) Besitzantwort {
	return b.app(ctx).SammlungBesitz(b.ruf(ctx), printing)
}

// SammlungLagerorte zählt die Lagerorte auf.
func (b *Bindungen) SammlungLagerorte(ctx context.Context) Lagerortantwort {
	return b.app(ctx).SammlungLagerorte(b.ruf(ctx))
}

// SammlungDubletten zeigt, wovon mehr da ist als gewollt.
func (b *Bindungen) SammlungDubletten(ctx context.Context, behalten int) Dublettenantwort {
	return b.app(ctx).SammlungDubletten(b.ruf(ctx), behalten)
}

// SammlungWuensche gibt die Wunschliste.
func (b *Bindungen) SammlungWuensche(ctx context.Context) Sammlungsantwort {
	return b.app(ctx).SammlungWuensche(b.ruf(ctx))
}

// keineSichten ist der Fehler ohne eingerichtete Naht.
func keineSichten() error {
	return fmt.Errorf("%w: die Sammlungssichten sind nicht eingerichtet",
		domain.ErrValidation)
}
