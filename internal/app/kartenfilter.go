// **Der Schnellfilter am Fenster und an den Flächen** — Werkstück U5 von
// Phase 5.1 (#1416).
//
// **Kennungen, keine Zeilen.** Die Liste hat das Fenster schon
// (`Kartenlistenaufbau`); was der Filter beiträgt, ist die Menge derer, die
// bleiben. So bleibt die Wahl stehen, wenn jemand tippt, und die Liste
// wechselt nicht den Stand, während sie gefiltert wird (§7, Regel 1).

package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KartenfilterAntwort trägt die Kennungen der Karten, die durch den Filter
// kommen — als Zeichenketten, wie jede Kennung an dieser Naht.
type KartenfilterAntwort struct {
	KartenIDs []string `json:"kartenIDs"`
	Fehler    *Fehler  `json:"fehler"`
}

// KartenFiltern legt den Schnellfilter über die Karten eines Sets.
func (a *App) KartenFiltern(ctx context.Context, setID, anfrage string) KartenfilterAntwort {
	ctx = a.imProfil(ctx)

	leer := KartenfilterAntwort{KartenIDs: []string{}}

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	if a.listenbau == nil {
		leer.Fehler = Uebersetze(ctx, fmt.Errorf(
			"%w: die Kartenliste ist nicht eingerichtet", domain.ErrValidation))

		return leer
	}

	kennungen, err := a.listenbau.Filtern(ctx, kennung, anfrage)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	ids := make([]string, 0, len(kennungen))
	for _, id := range kennungen {
		ids = append(ids, strconv.FormatInt(id, 10))
	}

	return KartenfilterAntwort{KartenIDs: ids}
}

// KartenFiltern filtert die Karten eines Sets vom Fenster aus.
func (b *Bindungen) KartenFiltern(ctx context.Context, setID, anfrage string) KartenfilterAntwort {
	return b.app(ctx).KartenFiltern(b.ruf(ctx), setID, anfrage)
}

// Filtern filtert auch für die Flächen ohne Fenster.
func (f *Fremdliste) Filtern(ctx context.Context, setID int64, anfrage string) ([]int64, error) {
	return f.quelle.Filtern(ctx, setID, anfrage)
}
