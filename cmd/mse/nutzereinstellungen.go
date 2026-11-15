// **Die Einstellungen eines Menschen und seine Auskunft an der
// Kommandozeile** — Werkstück 8 von Phase 4.12
// (#1351),
// Punkte 9, 10 und 11.
//
// # Ein Befehl mit zwei Gestalten
//
// `mse nutzereinstellung <id>` nennt sie, `mse nutzereinstellung <id>
// <schluessel> <wert>` setzt eine — dieselbe Bauform wie `mse einstellung`
// daneben, und aus demselben Grund: es ist **eine** Frage.

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// nutzereinstellungsbefehl nennt die Einstellungen oder setzt eine.
func nutzereinstellungsbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 && len(rest) != 3 {
		an.klage("nutzereinstellung braucht eine Kennung — zum Setzen dazu "+
			"Schlüssel und Wert.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		flaeche := app.NutzerflaecheUeber(db)

		antwort := flaeche.Nutzereinstellungen(ctx, rest[0])
		if len(rest) == 3 {
			antwort = flaeche.NutzereinstellungSetzen(ctx, rest[0], rest[1], rest[2])
		}

		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		for _, eine := range antwort.Einstellungen {
			an.sag("%s\t%s\n", eine.Schluessel, eine.Wert)
		}

		return an.abschluss()
	})
}

// personendatenbefehl gibt heraus, was über einen Menschen gespeichert ist.
//
// **Immer als JSON, auch ohne `--json`.** Eine Auskunft über personenbezogene
// Daten ist ein **Dokument**; sie in Spalten zu zerlegen hieße, sie
// unvollständig zu machen — und wer sie mitnehmen will, schreibt sie mit `>`
// in eine Datei.
func personendatenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("personendaten braucht genau eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).Personendaten(ctx, rest[0])
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		roh, fehler := json.MarshalIndent(antwort.Daten, "", "  ")
		if fehler != nil {
			an.klageUeber(fmt.Errorf("die Auskunft packen: %w", fehler))

			return CodeFachlich
		}

		an.sag("%s\n", roh)

		return an.abschluss()
	})
}
