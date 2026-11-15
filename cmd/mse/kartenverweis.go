package main

// **Der Verweis auf einen echten Druck an der Kommandozeile** — Werkstück 2
// von Phase 3.6
// (#1167).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort. Wer dreihundert Karten
// eines eingelesenen Satzes auf ihre Drucke bezieht, tut das nicht von Hand im
// Fenster — er tut es hier, in einer Schleife.

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// verweisbefehl zeigt, setzt oder löst den Verweis einer Karte.
//
// **Ein Befehl mit drei Gestalten und nicht drei Befehle**: `verweis <karte>`
// zeigt, `verweis <karte> <druck>` setzt, `verweis <karte> --loesen` löst.
// Dieselbe Gestalt wie `feld`, und aus demselben Grund: es ist eine Sache mit
// drei Zeitformen, nicht drei Sachen.
func verweisbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) == 0 {
		an.klage("verweis braucht eine Kartenkennung.\n%s", hilfe)

		return CodeBedienung
	}

	karteID, gut := kennung(an, "Kartenkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitVerweisen(ctx, bestandsort, an, func(ctx context.Context, naht *app.Fremdverweise) int {
		switch {
		case len(rest) == 1:
			return verweisZeigen(ctx, an, naht, karteID)

		case rest[1] == "--loesen":
			return verweisLoesen(ctx, an, naht, karteID)

		default:
			return verweisSetzen(ctx, an, naht, karteID, rest[1])
		}
	})
}

// setverweisbefehl zählt die Verweise eines Sets auf.
func setverweisbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) == 0 {
		an.klage("verweise braucht eine Setkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitVerweisen(ctx, bestandsort, an, func(ctx context.Context, naht *app.Fremdverweise) int {
		verweise, fehler := naht.JeSet(ctx, setID)
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{"verweise": verweise})
		}

		for _, verweis := range verweise {
			schreibeVerweis(an, verweis)
		}

		return an.abschluss()
	})
}

// verweisZeigen nennt den Verweis einer Karte.
//
// **Ohne Verweis ist die Antwort leer und der Rückgabewert null.** Eine Karte
// ohne Verweis ist der Normalfall; ein Fehlschlag machte aus ihm eine
// Ausnahme, und ein Skript, das über ein Set läuft, bräche daran ab.
func verweisZeigen(
	ctx context.Context, an *schreiber, naht *app.Fremdverweise, karteID int64,
) int {
	verweis, fehler := naht.Fuer(ctx, karteID)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{"verweis": verweis})
	}

	if verweis == nil {
		return an.abschluss()
	}

	schreibeVerweis(an, *verweis)

	return an.abschluss()
}

// verweisSetzen bezieht eine Karte auf einen Druck.
func verweisSetzen(
	ctx context.Context, an *schreiber, naht *app.Fremdverweise,
	karteID int64, printing string,
) int {
	if strings.TrimSpace(printing) == "" {
		an.klage("verweis braucht die Kennung eines Drucks.\n%s", hilfe)

		return CodeBedienung
	}

	verweis, fehler := naht.Setzen(ctx, karteID, printing)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{"verweis": verweis})
	}

	schreibeVerweis(an, verweis)

	return an.abschluss()
}

// verweisLoesen nimmt den Verweis einer Karte zurück.
func verweisLoesen(
	ctx context.Context, an *schreiber, naht *app.Fremdverweise, karteID int64,
) int {
	if fehler := naht.Loesen(ctx, karteID); fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{"karteID": karteID, "geloest": true})
	}

	return an.abschluss()
}

// schreibeVerweis gibt eine Zeile je Verweis.
//
// **Jedes Feld, das gesetzt werden kann, steht auch hier** — dieselbe Regel
// wie bei der Sammlung: was über eine Fläche hineingeht und nie zurückkommt,
// sieht niemand wieder und hat keinen Grund, es zu bemerken.
func schreibeVerweis(an *schreiber, verweis app.KartenverweisDTO) {
	an.sag("%s\t%s\t%s\t%s %s\t%s\t%s\n", verweis.KarteID, verweis.Printing,
		verweis.Name, verweis.Setcode, verweis.Nummer,
		verweis.Gesetzt, verweis.Adresse)
}

// mitVerweisen öffnet den Bestand und baut die Naht.
func mitVerweisen(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Fremdverweise) int,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.Verweisflaechennaht(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return arbeit(ctx, naht)
	})
}
