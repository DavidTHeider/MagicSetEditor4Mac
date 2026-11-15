// **Karten aus einem Skript, im Fenster** — Werkstück B von Phase 4.11
// (#1337).

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Kartenskriptquelle ist, was das Fenster für die Kartenskripte braucht.
type Kartenskriptquelle interface {
	Angebot(ctx context.Context, setID int64) ([]KartenskriptDTO, error)
	Ausfuehren(ctx context.Context, setID int64, name string) (*KartenskriptlaufDTO, error)
}

// MitKartenskripten setzt die Naht.
func (a *App) MitKartenskripten(quelle Kartenskriptquelle) { a.kartenskripte = quelle }

// Kartenskriptantwort trägt die Skripte des Spielpakets.
type Kartenskriptantwort struct {
	Skripte []KartenskriptDTO `json:"skripte"`
	Fehler  *Fehler           `json:"fehler"`
}

// Kartenskriptlaufantwort trägt, was ein Lauf angelegt hat.
//
// **Flach und nicht um ein `*KartenskriptlaufDTO` herum.** Ein geschachtelter
// Zeiger käme im JSON als `null` an, sobald `Fehler` steht — und die Seite
// hätte einen Zweig zu bedienen, den sie nie erreicht: `auspacken` wirft
// vorher. Ein Zweig, den kein Lauf betritt, ist keine Vorsicht, sondern eine
// unbelegbare Zusage (§3, §10). Beide Listen sind deshalb immer da, auf jedem
// Weg.
type Kartenskriptlaufantwort struct {
	// Karten sind die Kennungen der angelegten Karten, in ihrer Reihenfolge.
	Karten []string `json:"karten"`

	// Befunde ist, was unterwegs auffiel.
	Befunde []string `json:"befunde"`

	Fehler *Fehler `json:"fehler"`
}

// keinLauf ist die Antwort, die nichts angelegt hat.
func keinLauf(befund *Fehler) Kartenskriptlaufantwort {
	return Kartenskriptlaufantwort{
		Karten: []string{}, Befunde: []string{}, Fehler: befund,
	}
}

// Kartenskripte nennt die Kartenskripte eines Sets.
func (a *App) Kartenskripte(ctx context.Context, setID string) Kartenskriptantwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return Kartenskriptantwort{Skripte: []KartenskriptDTO{}, Fehler: befund}
	}

	if a.kartenskripte == nil {
		return Kartenskriptantwort{
			Skripte: []KartenskriptDTO{}, Fehler: keineKartenskripte(),
		}
	}

	skripte, err := a.kartenskripte.Angebot(ctx, kennung)
	if err != nil {
		return Kartenskriptantwort{Skripte: []KartenskriptDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	if skripte == nil {
		skripte = []KartenskriptDTO{}
	}

	return Kartenskriptantwort{Skripte: skripte}
}

// KartenskriptAusfuehren führt ein Skript und hängt seine Karten ans Set.
func (a *App) KartenskriptAusfuehren(
	ctx context.Context, setID, name string,
) Kartenskriptlaufantwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return keinLauf(befund)
	}

	if a.kartenskripte == nil {
		return keinLauf(keineKartenskripte())
	}

	lauf, err := a.kartenskripte.Ausfuehren(ctx, kennung, name)
	if err != nil {
		return keinLauf(Uebersetze(ctx, err))
	}

	return Kartenskriptlaufantwort{Karten: lauf.Karten, Befunde: lauf.Befunde}
}

// Kartenskripte nennt die Kartenskripte eines Sets.
func (b *Bindungen) Kartenskripte(ctx context.Context, setID string) Kartenskriptantwort {
	return b.app(ctx).Kartenskripte(b.ruf(ctx), setID)
}

// KartenskriptAusfuehren führt ein Skript und hängt seine Karten ans Set.
func (b *Bindungen) KartenskriptAusfuehren(ctx context.Context, setID, name string) Kartenskriptlaufantwort {
	return b.app(ctx).KartenskriptAusfuehren(b.ruf(ctx), setID, name)
}

// keineKartenskripte sagt, dass die Naht fehlt.
//
// **Ein Fehler und keine leere Liste**: „dieses Spiel hat keine Skripte" und
// „das Fenster ist nicht verdrahtet" sind verschiedene Auskünfte, und die
// zweite gehört gesagt.
func keineKartenskripte() *Fehler {
	return Nutzerfehler(fmt.Errorf(
		"%w: die Kartenskripte sind nicht eingerichtet", domain.ErrValidation))
}
