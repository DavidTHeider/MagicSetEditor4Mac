// **Karten aus einem Skript, an der Kommandozeile** — Werkstück B von Phase
// 4.11 (#1337).

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// kartenskriptebefehl nennt die Kartenskripte eines Sets.
func kartenskriptebefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) == 0 {
		an.klage("kartenskripte braucht ein Set.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "Set", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitKartenskripten(ctx, bestandsort, an, func(ctx context.Context, naht *app.Kartenskriptnaht) int {
		skripte, err := naht.Angebot(ctx, setID)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{"skripte": skripte})
		}

		for _, skript := range skripte {
			an.sag("%s\t%s\n", skript.Name, skript.Beschreibung)
		}

		return an.abschluss()
	})
}

// kartenskriptbefehl führt ein Kartenskript aus.
func kartenskriptbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 2 {
		an.klage("kartenskript braucht ein Set und einen Namen.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "Set", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitKartenskripten(ctx, bestandsort, an, func(ctx context.Context, naht *app.Kartenskriptnaht) int {
		lauf, err := naht.Ausfuehren(ctx, setID, rest[1])
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, lauf)
		}

		an.sag("%d Karten angelegt\n", len(lauf.Karten))

		for _, befund := range lauf.Befunde {
			an.sag("!\t%s\n", befund)
		}

		return an.abschluss()
	})
}

// mitKartenskripten baut die Naht über dem Bestand.
func mitKartenskripten(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Kartenskriptnaht) int,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.KartenskripteUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return arbeit(ctx, naht)
	})
}
