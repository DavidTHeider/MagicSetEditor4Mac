// **Die Bindungen der Massenvorgänge** — Werkstück 6 von Phase 4.5
// (#1292).

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Massenquelle ist, was das Fenster für Vorgänge über eine Auswahl braucht.
type Massenquelle interface {
	Aendern(ctx context.Context, dto MassenaenderungDTO) (*MassenberichtDTO, error)
	Loeschen(ctx context.Context, zeilen []MassenzeileDTO) (*MassenberichtDTO, error)
	PreiseInMasse(ctx context.Context, hoechstensStunden int) (*PreisschubDTO, error)
}

// MitMasse setzt die Naht.
func (a *App) MitMasse(quelle Massenquelle) { a.masse = quelle }

// Massenantwort trägt den Bericht eines Massenvorgangs.
type Massenantwort struct {
	Bericht *MassenberichtDTO `json:"bericht"`
	Fehler  *Fehler           `json:"fehler"`
}

// Preisschubantwort trägt den Schub oder einen Fehler.
type Preisschubantwort struct {
	Schub  *PreisschubDTO `json:"schub"`
	Fehler *Fehler        `json:"fehler"`
}

// MasseAendern ändert viele Exemplare auf einmal.
func (a *App) MasseAendern(ctx context.Context, dto MassenaenderungDTO) Massenantwort {
	ctx = a.imProfil(ctx)

	if a.masse == nil {
		return Massenantwort{Fehler: Uebersetze(ctx, keineMasse())}
	}

	bericht, err := a.masse.Aendern(ctx, dto)

	// **Bericht und Fehler zugleich** (§7, Regel 3, Befund 3): was geschrieben
	// ist, ist geschrieben.
	return Massenantwort{Bericht: bericht, Fehler: Uebersetze(ctx, err)}
}

// MasseLoeschen nimmt viele Exemplare heraus.
func (a *App) MasseLoeschen(
	ctx context.Context, zeilen []MassenzeileDTO,
) Massenantwort {
	ctx = a.imProfil(ctx)

	if a.masse == nil {
		return Massenantwort{Fehler: Uebersetze(ctx, keineMasse())}
	}

	bericht, err := a.masse.Loeschen(ctx, zeilen)

	// **Bericht und Fehler zugleich** (§7, Regel 3, Befund 3): was geschrieben
	// ist, ist geschrieben.
	return Massenantwort{Bericht: bericht, Fehler: Uebersetze(ctx, err)}
}

// MassePreise holt einen Schub Preise.
//
// **Einen Schub je Ruf** — der Grund ist die Drossel (§16, Regel 4). Wer den
// Fortschritt zeigen will, ruft wieder, bis `fertig` wahr ist.
func (a *App) MassePreise(ctx context.Context, hoechstensStunden int) Preisschubantwort {
	ctx = a.imProfil(ctx)

	if a.masse == nil {
		return Preisschubantwort{Fehler: Uebersetze(ctx, keineMasse())}
	}

	beenden := a.auftragBeginnen("preise", "")

	schub, err := a.masse.PreiseInMasse(ctx, hoechstensStunden)
	beenden(err)

	if err != nil {
		return Preisschubantwort{Fehler: Uebersetze(ctx, err)}
	}

	return Preisschubantwort{Schub: schub}
}

// MasseAendern ändert viele Exemplare auf einmal.
func (b *Bindungen) MasseAendern(ctx context.Context, dto MassenaenderungDTO) Massenantwort {
	return b.app(ctx).MasseAendern(b.ruf(ctx), dto)
}

// MasseLoeschen nimmt viele Exemplare heraus.
func (b *Bindungen) MasseLoeschen(ctx context.Context, zeilen []MassenzeileDTO) Massenantwort {
	return b.app(ctx).MasseLoeschen(b.ruf(ctx), zeilen)
}

// MassePreise holt einen Schub Preise.
func (b *Bindungen) MassePreise(ctx context.Context, hoechstensStunden int) Preisschubantwort {
	return b.app(ctx).MassePreise(b.ruf(ctx), hoechstensStunden)
}

// keineMasse ist der Fehler ohne eingerichtete Naht.
func keineMasse() error {
	return fmt.Errorf("%w: die Massenvorgänge sind nicht eingerichtet",
		domain.ErrValidation)
}
