package main

// **Der Poststand an der Kommandozeile** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// **Zahlen und keine Nachrichten**: die Diagnose sagt, ob etwas klemmt, und
// nicht, worum es geht.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// poststandsicht ist der Stand, wie die Kommandozeile ihn schreibt.
//
// **Eine eigene Gestalt** — die Wails-DTO ist **innen**, diese Fläche
// **außen** (0033, Regel 5). Dieselbe Trennung wie bei `exemplarsicht`.
type poststandsicht struct {
	Kennung          string `json:"kennung"`
	Offen            int    `json:"offen"`
	HoechsteVersuche int    `json:"hoechsteVersuche"`
}

// handelspostbefehl fährt `mse handelspost`.
func handelspostbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("handelspost nimmt ein Argument.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.PeerflaecheUeber(db).Handelspoststand(ctx, rest[0])
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		sicht := poststandsicht{
			Kennung: antwort.Stand.Kennung, Offen: antwort.Stand.Offen,
			HoechsteVersuche: antwort.Stand.HoechsteVersuche,
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, sicht)
		}

		an.sag("%s\t%d\t%d\n", sicht.Kennung, sicht.Offen, sicht.HoechsteVersuche)

		return an.abschluss()
	})
}
