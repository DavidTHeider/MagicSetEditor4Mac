// **Der Schnellfilter an der Zeile** — Werkstück U5 von Phase 5.1
// (#1416).
//
// `mse kartenfilter <set> <anfrage…>` gibt die Kennungen der Karten, die
// durch den Filter kommen — in der Sprache von MSE2 (`src/data/filter.cpp`,
// Fassung f9d9356): Leerraum trennt, `"…"` hält zusammen, `feld:wert` bindet,
// ein führendes `-` verneint. Mehrere Wörter auf der Zeile sind **eine**
// Anfrage; wer die Anführungszeichen der Shell braucht, schreibt sie doppelt.

package main

import (
	"context"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// kartenfilterbefehl beantwortet `mse kartenfilter`.
func kartenfilterbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 2 {
		an.klage("kartenfilter braucht eine Setkennung und eine Anfrage.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	anfrage := strings.Join(rest[1:], " ")

	return mitListe(ctx, bestandsort, an, func(ctx context.Context, naht *app.Fremdliste) int {
		kennungen, fehler := naht.Filtern(ctx, setID, anfrage)
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		ids := make([]string, 0, len(kennungen))
		for _, id := range kennungen {
			ids = append(ids, strconv.FormatInt(id, 10))
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{"kartenIDs": ids})
		}

		for _, id := range ids {
			an.sag("%s\n", id)
		}

		return an.abschluss()
	})
}
