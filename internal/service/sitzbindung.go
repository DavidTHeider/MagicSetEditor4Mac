package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// sitzbindungsschluessel ist der Schlüssel der Sitzbindung im Kontext.
type sitzbindungsschluessel struct{}

// MitSitzbindung hängt an einen Kontext die Sitzbindung des Tokens, über das
// ein Aufruf kam — gesetzt von der Tokenwache, nirgends sonst (#1819).
//
// **Am Fenster, an der Kommandozeile und mit der Prozessmarke gibt es
// keine**: dort sitzt ein Mensch, und den Sitz einer Sicht nennt er selbst
// (G8, #1465).
func MitSitzbindung(ctx context.Context, bindung domain.Sitzbindung) context.Context {
	return context.WithValue(ctx, sitzbindungsschluessel{}, bindung)
}

// SitzbindungAus liest sie — leer, wenn kein gebundenes Token im Spiel ist.
func SitzbindungAus(ctx context.Context) domain.Sitzbindung {
	bindung, _ := ctx.Value(sitzbindungsschluessel{}).(domain.Sitzbindung) //nolint:errcheck // ohne Bindung die leere

	return bindung
}
