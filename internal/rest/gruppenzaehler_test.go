package rest_test

// **Was die Doppel über Gruppen sagen** — abgetrennt am 2026-09-13, weil
// `zaehlerdoppel_test.go` mit ihnen über die Zeilengrenze aus §2 ging (A13.1b, #1452).
//
// # Der Schnitt folgt der Frage
//
// Dort stehen die Doppel für den **einzelnen** Weg; hier die Antwort auf die
// eine Frage, die A13.1b hinzugefügt hat: was geschieht, wenn jemand eine
// **Gruppe** schreiben will. **Sie scheitern laut** (§3) — ein Doppel, das
// hier still gutginge, ließe einen Lauf über Gruppen grün werden, der gar
// keine geschrieben hat.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// KIGruppe übernimmt eine Gruppe und zählt mit — A13.1b (#1452).
func (z *zaehlender) KIGruppe(ctx context.Context, bezug rest.KIGruppenbezug) (
	rest.KIGruppensicht, error,
) {
	z.rufe++

	return z.williger.KIGruppe(ctx, bezug)
}
