package mcp_test

// **Was die Doppel über Gruppen sagen** — abgetrennt am 2026-09-13, weil
// `sammlungsprobe_test.go` mit ihnen über die Zeilengrenze aus §2 ging (A13.1b, #1452).
//
// # Der Schnitt folgt der Frage
//
// Dort stehen die Doppel für den **einzelnen** Weg; hier die Antwort auf die
// eine Frage, die A13.1b hinzugefügt hat: was geschieht, wenn jemand eine
// **Gruppe** schreiben will.
//
// **Die meisten scheitern laut** (§3) — ein Doppel, das still gutginge, ließe
// einen Lauf über Gruppen grün werden, der gar keine geschrieben hat. **Die
// Preisprobe unten ist die Ausnahme**: an ihr wird der geglückte Weg
// gemessen, und sie steht deshalb hier und nicht bei den anderen.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// KIGruppe gibt es an diesem Doppel **nur der Schnittstelle wegen** (A13.1b,
// #1452) — und sie scheitert laut (§3).
func (stolpernderPreis) KIGruppe(context.Context, mcp.KIGruppenbezug) (
	mcp.KIGruppensicht, error,
) {
	return mcp.KIGruppensicht{}, errGestolpertMCP
}

// **Und die Preisprobe übernimmt sie** — sie ist das Doppel, an dem der
// geglückte Weg gemessen wird.
func (*preisprobe) KIGruppe(_ context.Context, bezug mcp.KIGruppenbezug) (
	mcp.KIGruppensicht, error,
) {
	namen := make([]mcp.KIGruppenkartensicht, 0, len(bezug.Wahl))

	felder := 0

	for _, eine := range bezug.Wahl {
		felder += len(eine.Felder)

		namen = append(namen, mcp.KIGruppenkartensicht{
			KarteID: 42, Felder: eine.Felder,
		})
	}

	return mcp.KIGruppensicht{
		Karten: len(namen), Felder: felder, Namen: namen,
	}, nil
}

// **Und die Nullprobe gibt einen Namen ohne Felder** — A13.1b (#1452).
//
// **Mit einem Eintrag und `nil` darin**: der Weg, auf dem aus der inneren
// Liste eine leere wird, ist sonst nicht zu betreten.
func (nilprobe) KIGruppe(_ context.Context, bezug mcp.KIGruppenbezug) (
	mcp.KIGruppensicht, error,
) {
	if len(bezug.Wahl) == 0 {
		return mcp.KIGruppensicht{}, nil
	}

	return mcp.KIGruppensicht{
		Karten: 1, Felder: 0,
		Namen: []mcp.KIGruppenkartensicht{{KarteID: 42}},
	}, nil
}
