package mcp_test

// **Das Doppel des Batches** — Werkstück A13.1 von Phase 5.3
// (#1452).
//
// **Eine eigene Datei**, weil `sammlungsprobe_test.go` mit ihm über die
// Zeilengrenze aus §2 ginge.
import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// KIBatch antwortet für jedes Ziel — A13.1 (#1452).
func (p *preisprobe) KIBatch(_ context.Context, bezug mcp.KIBatchbezug) (
	mcp.KIBatchsicht, error,
) {
	ziele := make([]mcp.KIBatchlagensicht, 0, len(bezug.KarteIDs))

	for i, karte := range bezug.KarteIDs {
		// **Das letzte Ziel scheitert** — und trägt deshalb **keine** Felder
		// und keine Befunde. Ein Doppel, in dem jedes Ziel gelingt, ließe
		// den Weg ungeprüft, auf dem aus `null` eine leere Liste wird.
		if i == len(bezug.KarteIDs)-1 && len(bezug.KarteIDs) > 1 {
			ziele = append(ziele, mcp.KIBatchlagensicht{
				KarteID: karte, Fehler: "die Karte gibt es nicht",
			})

			continue
		}

		ziele = append(ziele, mcp.KIBatchlagensicht{
			KarteID: karte,
			Kennung: fmt.Sprintf("ki-%d-%d", karte, i),
			Modell:  bezug.Alias,
			Felder: []mcp.KIFeldwertsicht{
				{Feld: "name", Werte: []string{"Probe"}},
			},
			Ablauf: "2026-09-13T12:00:00Z",
		})
	}

	return mcp.KIBatchsicht{
		Kennung: "kibatch-1", Anbieter: bezug.Anbieter, Ziele: ziele,
	}, nil
}
