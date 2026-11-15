// **Pakete als `.mse-installer` weitergeben und einlesen, an der
// Kommandozeile** — FMT-02.AK-04
// (#1691).
//
// Einordnung nach §16: **erweiternd** — zwei Befehle kommen hinzu.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// weitergabeAntwort ist die Ausgabe von `paket-weitergeben --json`.
type weitergabeAntwort struct {
	Datei string `json:"datei"`
}

// paketweitergabebefehl schreibt Pakete als Installer — erst das Ziel, dann
// die Pakete mit Endung, wie `magicseteditor --create-installer`.
func paketweitergabebefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 2 {
		an.klage("paket-weitergeben braucht das Ziel und wenigstens ein Paket.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg, err := app.PaketnahtUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		datei, fehler := weg.Weitergeben(ctx, rest[1:], rest[0])
		if fehler != nil {
			an.klageUeber(fehler)

			return paketCode(fehler)
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, weitergabeAntwort{Datei: datei})
		}

		an.sag("%s\n", datei)

		return an.abschluss()
	})
}

// installereinlesebefehl legt die Pakete eines Installers ins Verzeichnis.
func installereinlesebefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 1 {
		an.klage("installer-einlesen braucht die Datei.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg, err := app.PaketnahtUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		uebernahme, fehler := weg.Einlesen(ctx, rest[0])

		// **Erst das Ergebnis, dann der Fehler** — wie bei der Übernahme.
		if uebernahme != nil {
			uebernahmeZeigen(an, uebernahme)
		}

		if fehler != nil {
			an.klageUeber(fehler)

			return paketCode(fehler)
		}

		return an.abschluss()
	})
}
