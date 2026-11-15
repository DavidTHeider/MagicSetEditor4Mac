package mcp_test

// **Die Doppel für die Decklisteneinfuhr und die Marktlinks** — Werkstück
// `T8` von Phase 5.6
// (#1473).
//
// # Warum sie in der Vorrichtung stehen müssen
//
// **Sonst sehen die Wächter über der Werkzeugliste ihre Werkzeuge nicht**,
// und Idempotenzfall, Zusage und README blieben ungeprüft — genau die Gestalt
// aus §10: ein Wächter, der die halbe Zusage hält. Denselben Satz trägt schon
// die Paketpflege in `probenaehte_test.go`.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// deckeinfuhrprobe ist eine Naht, die eine feste Vorschau gibt.
type deckeinfuhrprobe struct{}

func (deckeinfuhrprobe) Vorschau(_ context.Context, deckID int64, _ string) (
	mcp.Deckvorschau, error,
) {
	return mcp.Deckvorschau{
		DeckID: deckID, Stand: "abdruck", Hauptdeck: 4, Sideboard: 0,
		Befunde: []string{},
		Zeilen: []mcp.Deckvorschauzeile{{
			Bereich: "hauptdeck", Menge: 4, Name: "Flammenball",
			Printing: "a1b2c3d4-0000-0000-0000-00000000000a",
			Treffer:  []mcp.Listentreffer{}, Drucke: 1, Uebernehmbar: true,
		}},
	}, nil
}

func (deckeinfuhrprobe) Uebernehmen(
	_ context.Context, _ int64, _ string, zeilen []mcp.Deckeinfuhrzeile,
) (mcp.Deckeinfuhrbericht, error) {
	stueck := 0
	for _, zeile := range zeilen {
		stueck += zeile.Menge
	}

	return mcp.Deckeinfuhrbericht{
		Gelegt: len(zeilen), Stueck: stueck, Stand: "abdruck-danach",
	}, nil
}
