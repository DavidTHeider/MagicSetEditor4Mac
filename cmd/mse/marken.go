package main

// **Marken an der Kommandozeile** — Werkstück 3 von Phase 3.5
// (#1208).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort. Wer dreihundert Karten
// eines eingelesenen Satzes durchsieht, markiert sie nicht von Hand im
// Fenster — er tut es hier, in einer Schleife.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// markenbefehl zeigt, setzt oder entfernt die Marken einer Karte.
//
// **Ein Befehl mit drei Gestalten und nicht drei Befehle** — dieselbe Bauform
// wie `verweis`, und aus demselben Grund: es ist eine Sache mit drei
// Zeitformen.
//
//	mse marke <set> <karte>            zeigt
//	mse marke <set> <karte> <wort>     setzt
//	mse marke <set> <karte> <wort> --weg  entfernt
func markenbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 2 {
		an.klage("marke braucht eine Set- und eine Kartenkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gutesSet := kennung(an, "Setkennung", rest[0])
	if !gutesSet {
		return CodeBedienung
	}

	karteID, guteKarte := kennung(an, "Kartenkennung", rest[1])
	if !guteKarte {
		return CodeBedienung
	}

	return mitMarken(ctx, bestandsort, an, func(ctx context.Context, pflege *app.Fremdmarken) int {
		switch {
		case len(rest) == 2:
			return markenZeigen(ctx, an, pflege, setID, karteID)

		case len(rest) > 3 && rest[3] == "--weg":
			return markeSchreiben(an, pflege.MarkeEntfernen(ctx, setID, karteID, rest[2]),
				"entfernt", rest[2])

		default:
			return markeSchreiben(an, pflege.MarkeSetzen(ctx, setID, karteID, rest[2]),
				"gesetzt", rest[2])
		}
	})
}

// setmarkenbefehl zählt die Marken eines Sets auf.
func setmarkenbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) == 0 {
		an.klage("marken braucht eine Setkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitMarken(ctx, bestandsort, an, func(ctx context.Context, pflege *app.Fremdmarken) int {
		marken, err := pflege.Setmarken(ctx, setID)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return markenAusgeben(an, marken)
	})
}

// markenZeigen schreibt die Marken einer Karte.
func markenZeigen(
	ctx context.Context, an *schreiber, pflege *app.Fremdmarken, setID, karteID int64,
) int {
	marken, err := pflege.KartenMarken(ctx, setID, karteID)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	return markenAusgeben(an, marken)
}

// markenAusgeben schreibt eine Markenliste — je Zeile eine.
//
// **Eine Zeile je Marke und kein Satz drumherum**: was hier herauskommt, geht
// in `while read`, und ein „Keine Marken." mittendrin wäre eine Marke.
func markenAusgeben(an *schreiber, marken []string) int {
	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{"marken": marken})
	}

	for _, marke := range marken {
		an.sag("%s\n", marke)
	}

	return an.abschluss()
}

// markeSchreiben meldet das Ergebnis eines Schreibvorgangs.
func markeSchreiben(an *schreiber, err error, was, marke string) int {
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	normal := domain.Markennormalform(marke)

	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{"marke": normal, was: true})
	}

	an.sag("%s\n", normal)

	return an.abschluss()
}

// mitMarken öffnet den Bestand und baut die Markenpflege.
func mitMarken(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Fremdmarken) int,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return arbeit(ctx, app.Markennaht(db))
	})
}
