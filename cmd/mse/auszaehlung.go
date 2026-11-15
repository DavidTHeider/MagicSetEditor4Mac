// **Auszählen nach einer erklärten Dimension, an der Kommandozeile** —
// Werkstück C von Phase 4.11
// (#1338).

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// auszaehlungbefehl nennt die Dimensionen — oder zählt nach einer aus.
//
// **Ein Befehl und nicht zwei.** Ohne Namen steht das Angebot da, mit Namen
// die Auszählung; dasselbe Muster wie bei `stile` und `vorlagen`.
func auszaehlungbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) == 0 {
		an.klage("auszaehlung braucht ein Set.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "Set", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitAuszaehlung(ctx, bestandsort, an, func(ctx context.Context, naht *app.Auszaehlnaht) int {
		if len(rest) == 1 {
			return dimensionenZeigen(ctx, naht, an, setID)
		}

		return auszaehlenZeigen(ctx, naht, an, setID, rest[1])
	})
}

// dimensionenZeigen nennt, wonach dieses Spiel auszählen lässt.
func dimensionenZeigen(
	ctx context.Context, naht *app.Auszaehlnaht, an *schreiber, setID int64,
) int {
	dimensionen, err := naht.Angebot(ctx, setID)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{"dimensionen": dimensionen})
	}

	for _, eine := range dimensionen {
		an.sag("%s\t%s\n", eine.Name, eine.Beschreibung)
	}

	return an.abschluss()
}

// auszaehlenZeigen zählt das Set nach der benannten Dimension aus.
func auszaehlenZeigen(
	ctx context.Context, naht *app.Auszaehlnaht, an *schreiber,
	setID int64, name string,
) int {
	gezaehlt, err := naht.Auszaehlen(ctx, setID, name)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, gezaehlt)
	}

	for _, gruppe := range gezaehlt.Gruppen {
		an.sag("%6d\t%s\n", gruppe.Karten, gruppe.Name)
	}

	for _, befund := range gezaehlt.Befunde {
		an.sag("!\t%s\n", befund)
	}

	return an.abschluss()
}

// mitAuszaehlung baut die Naht über dem Bestand.
func mitAuszaehlung(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Auszaehlnaht) int,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.AuszaehlungUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return arbeit(ctx, naht)
	})
}
