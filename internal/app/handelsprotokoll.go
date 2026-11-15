package app

// **Das Protokoll eines Handels an der Naht** — TRADE-03.AK-03
// (#1717).

import "context"

// HandelsprotokollzeileDTO ist ein Zustandswechsel.
type HandelsprotokollzeileDTO struct {
	Revision int `json:"revision"`

	// Von ist leer, wenn die Revision neu ist.
	Von  string `json:"von"`
	Nach string `json:"nach"`

	// Zeit ist RFC 3339.
	Zeit string `json:"zeit"`
}

// HandelsprotokollAntwort ist das Protokoll samt seinem Fehler.
type HandelsprotokollAntwort struct {
	Zeilen []HandelsprotokollzeileDTO `json:"zeilen"`
	Fehler *Fehler                    `json:"fehler"`
}

// HandelProtokoll gibt, wie ein Vorgang dorthin kam, wo er steht — **rein
// lesend**, ältester Wechsel zuerst.
func (a *App) HandelProtokoll(ctx context.Context, vorgang string) HandelsprotokollAntwort {
	ctx = a.imProfil(ctx)

	protokoll, fehler := a.handelsleitung().Protokoll(ctx, vorgang)
	if fehler != nil {
		return HandelsprotokollAntwort{
			Zeilen: []HandelsprotokollzeileDTO{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	zeilen := make([]HandelsprotokollzeileDTO, 0, len(protokoll))
	for _, zeile := range protokoll {
		zeilen = append(zeilen, HandelsprotokollzeileDTO{
			Revision: zeile.Revision, Von: string(zeile.Von),
			Nach: string(zeile.Nach), Zeit: alsZeitpunkt(zeile.Zeit),
		})
	}

	return HandelsprotokollAntwort{Zeilen: zeilen}
}
