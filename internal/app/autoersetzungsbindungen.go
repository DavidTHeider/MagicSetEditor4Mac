// **Die Autoersetzungen im Fenster** — Werkstück E von Phase 4.11
// (#1341).

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Autoersetzungspflegequelle ist, was das Fenster dafür braucht.
type Autoersetzungspflegequelle interface {
	Liste(ctx context.Context, setID int64) ([]AutoersetzungsregelDTO, bool, error)
	Setzen(ctx context.Context, setID int64, regel AutoersetzungsregelDTO) error
	Loeschen(ctx context.Context, setID int64, muster string) error
	Zuruecksetzen(ctx context.Context, setID int64) error
	SchalterSetzen(ctx context.Context, setID int64, an bool) error
}

// MitAutoersetzungspflege setzt die Naht.
func (a *App) MitAutoersetzungspflege(quelle Autoersetzungspflegequelle) {
	a.autoersetzungspflege = quelle
}

// Autoersetzungsantwort trägt die Regeln und den Schalter.
type Autoersetzungsantwort struct {
	Regeln []AutoersetzungsregelDTO `json:"regeln"`

	// An ist der globale Schalter dieses Spiels.
	An bool `json:"an"`

	Fehler *Fehler `json:"fehler"`
}

// Autoersetzungen gibt die Regeln, die für dieses Set gelten.
func (a *App) Autoersetzungen(ctx context.Context, setID string) Autoersetzungsantwort {
	ctx = a.imProfil(ctx)

	leer := Autoersetzungsantwort{Regeln: []AutoersetzungsregelDTO{}}

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	if a.autoersetzungspflege == nil {
		leer.Fehler = keineAutoersetzungspflege()

		return leer
	}

	regeln, an, err := a.autoersetzungspflege.Liste(ctx, kennung)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	if regeln == nil {
		regeln = []AutoersetzungsregelDTO{}
	}

	return Autoersetzungsantwort{Regeln: regeln, An: an}
}

// AutoersetzungSetzen legt eine Regel an oder schreibt sie fort.
func (a *App) AutoersetzungSetzen(
	ctx context.Context, setID, muster, ersatz string, ganzesWort, aktiv bool,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.autoersetzungsschritt(ctx, setID, func(kennung int64) error {
		return a.autoersetzungspflege.Setzen(ctx, kennung, AutoersetzungsregelDTO{
			Muster: muster, Ersatz: ersatz, GanzesWort: ganzesWort, Aktiv: aktiv,
		})
	})
}

// AutoersetzungLoeschen nimmt eine eigene Regel heraus.
func (a *App) AutoersetzungLoeschen(
	ctx context.Context, setID, muster string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.autoersetzungsschritt(ctx, setID, func(kennung int64) error {
		return a.autoersetzungspflege.Loeschen(ctx, kennung, muster)
	})
}

// AutoersetzungenZuruecksetzen vergisst alle Abweichungen dieses Spiels.
func (a *App) AutoersetzungenZuruecksetzen(
	ctx context.Context, setID string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.autoersetzungsschritt(ctx, setID, func(kennung int64) error {
		return a.autoersetzungspflege.Zuruecksetzen(ctx, kennung)
	})
}

// AutoersetzungSchalten legt den globalen Schalter dieses Spiels um.
func (a *App) AutoersetzungSchalten(
	ctx context.Context, setID string, an bool,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.autoersetzungsschritt(ctx, setID, func(kennung int64) error {
		return a.autoersetzungspflege.SchalterSetzen(ctx, kennung, an)
	})
}

// autoersetzungsschritt prüft die Kennung und die Naht — für alle vier Wege.
func (a *App) autoersetzungsschritt(
	ctx context.Context, setID string, tun func(int64) error,
) SchreibAntwort {
	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	if a.autoersetzungspflege == nil {
		return SchreibAntwort{Fehler: keineAutoersetzungspflege()}
	}

	if err := tun(kennung); err != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return SchreibAntwort{}
}

// Autoersetzungen gibt die Regeln, die für dieses Set gelten.
func (b *Bindungen) Autoersetzungen(ctx context.Context, setID string) Autoersetzungsantwort {
	return b.app(ctx).Autoersetzungen(b.ruf(ctx), setID)
}

// AutoersetzungSetzen legt eine Regel an oder schreibt sie fort.
func (b *Bindungen) AutoersetzungSetzen(
	ctx context.Context,
	setID, muster, ersatz string, ganzesWort, aktiv bool,
) SchreibAntwort {
	return b.app(ctx).AutoersetzungSetzen(
		b.ruf(ctx), setID, muster, ersatz, ganzesWort, aktiv)
}

// AutoersetzungLoeschen nimmt eine eigene Regel heraus.
func (b *Bindungen) AutoersetzungLoeschen(ctx context.Context, setID, muster string) SchreibAntwort {
	return b.app(ctx).AutoersetzungLoeschen(b.ruf(ctx), setID, muster)
}

// AutoersetzungenZuruecksetzen vergisst alle Abweichungen dieses Spiels.
func (b *Bindungen) AutoersetzungenZuruecksetzen(ctx context.Context, setID string) SchreibAntwort {
	return b.app(ctx).AutoersetzungenZuruecksetzen(b.ruf(ctx), setID)
}

// AutoersetzungSchalten legt den globalen Schalter dieses Spiels um.
func (b *Bindungen) AutoersetzungSchalten(ctx context.Context, setID string, an bool) SchreibAntwort {
	return b.app(ctx).AutoersetzungSchalten(b.ruf(ctx), setID, an)
}

// keineAutoersetzungspflege sagt, dass die Naht fehlt.
func keineAutoersetzungspflege() *Fehler {
	return Nutzerfehler(fmt.Errorf(
		"%w: die Autoersetzungspflege ist nicht eingerichtet", domain.ErrValidation))
}
