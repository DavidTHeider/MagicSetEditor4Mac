// **Die Kennung eines Vorgangs im Kontext** — Werkstück O9 von Phase 5.2
// (#1433).
//
// Sie liegt neben dem Dienstmodus und dem Tokenumfang (`autorisierung.go`) und
// aus demselben Grund: was ein Aufruf über sich weiß, reist im `context` und
// nicht in jeder Signatur.

package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

type vorgangsschluessel struct{}

// MitVorgang hängt an einen Kontext die Kennung dieses Rufs — **gesetzt an der
// Grenze, an der er ankommt, und nirgends sonst**: am Fenster in
// `App.imProfil`, an den drei anderen Flächen in ihrer Wurzel.
func MitVorgang(ctx context.Context, kennung domain.Vorgangskennung) context.Context {
	return context.WithValue(ctx, vorgangsschluessel{}, kennung)
}

// VorgangAus liest die Kennung — leer, wenn keine Grenze eine gesetzt hat.
//
// **Leer ist ein gültiger Wert.** Ein Test, der einen nackten Kontext baut,
// hat keine Grenze durchlaufen; eine erfundene Kennung an dieser Stelle
// behauptete einen Vorgang, den es nicht gibt.
func VorgangAus(ctx context.Context) domain.Vorgangskennung {
	if kennung, da := ctx.Value(vorgangsschluessel{}).(domain.Vorgangskennung); da {
		return kennung
	}

	return ""
}
