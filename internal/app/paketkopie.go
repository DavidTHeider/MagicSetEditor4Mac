// **Eine eigene Kopie eines Pakets — an allen vier Flächen** — FMT-02.AK-04
// und S10.NAT-01 (#1691).
//
// Mitgelieferte und fremdverwaltete Pakete werden über eigene Kopien
// angepasst; was die Kopie heißt und was sie trägt, entscheidet
// `service.Paketverwalter.Kopieren`.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// PaketKopieren legt eine eigene Kopie eines Pakets an.
func (a *App) PaketKopieren(ctx context.Context, art, name, neu string) PaketAntwort {
	ctx = a.imProfil(ctx)

	paket, err := a.pakete.Kopieren(ctx, art, name, neu)
	if err != nil {
		return PaketAntwort{Fehler: Uebersetze(ctx, err)}
	}

	gelesen := a.mitUebersicht(ctx, paket)
	a.arbeitVermerken(ctx, domain.ArbeitsartPaket, gelesen.Name, domain.VorgangEingefuehrt)

	return PaketAntwort{Paket: &gelesen}
}
