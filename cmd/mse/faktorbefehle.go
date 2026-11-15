// **Der zweite Faktor an der Kommandozeile** — Werkstück 7 von Phase 4.12
// (#1351).
//
// # Das Geheimnis geht genau einmal über den Schirm
//
// **`faktor-neu` gibt Adresse, Geheimnis und zehn Wiederherstellungscodes
// aus** — und danach nie wieder. Wer sie wegscrollt, richtet neu ein.
//
// **Das ist der Grund, warum sie auf die Standardausgabe gehen und nicht auf
// den Fehlerstrom**: sie sind das Ergebnis des Befehls, und ein Skript soll
// sie in eine Datei lenken können.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// faktorenbefehl nennt die zweiten Faktoren eines Nutzers.
func faktorenbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 1 {
		an.klage("faktoren braucht genau eine Nutzerkennung.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).Faktoren(ctx, rest[0])
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		for _, einer := range antwort.Faktoren {
			an.sag("%s\t%s\n", einer.Art, einer.Bezeichnung)
		}

		// **Was der Rechner kann, steht dabei** — sonst wüsste niemand, warum
		// `faktor-neu biometrie` abweist.
		if !antwort.Biometrie {
			an.sag("# Biometrie: %s\n", antwort.Biometriegrund)
		}

		return an.abschluss()
	})
}

// faktorneubefehl richtet einen zweiten Faktor ein.
func faktorneubefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 2 {
		an.klage("faktor-neu braucht eine Nutzerkennung und eine Art "+
			"(totp oder biometrie).\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		flaeche := app.NutzerflaecheUeber(db)

		if rest[1] == "biometrie" {
			return schreibAntwortMelden(an, flaeche.BiometrieEinrichten(ctx, rest[0]))
		}

		return einrichtungSchreiben(an, flaeche.FaktorEinrichten(ctx, rest[0]))
	})
}

// faktorbestaetigenbefehl trägt den TOTP-Faktor ein.
func faktorbestaetigenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("faktor-bestaetigen braucht eine Nutzerkennung und einen "+
			"Code.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return schreibAntwortMelden(an,
			app.NutzerflaecheUeber(db).FaktorBestaetigen(ctx, rest[0], rest[1]))
	})
}

// faktorloeschenbefehl nimmt einen zweiten Faktor heraus.
func faktorloeschenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("faktor-loeschen braucht eine Nutzerkennung und eine Art.\n%s",
			hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return schreibAntwortMelden(an,
			app.NutzerflaecheUeber(db).FaktorEntfernen(ctx, rest[0], rest[1]))
	})
}

// einrichtungSchreiben gibt aus, was einmal zu sehen ist.
func einrichtungSchreiben(an *schreiber, antwort app.Einrichtungsantwort) int {
	if antwort.Fehler != nil {
		an.klageUeberFehler(antwort.Fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("%s\n%s\n", antwort.Adresse, antwort.Geheimnis)

	for _, code := range antwort.Codes {
		an.sag("%s\n", code)
	}

	return an.abschluss()
}

// schreibAntwortMelden gibt eine Schreibantwort aus — oder ihren Fehler.
func schreibAntwortMelden(an *schreiber, antwort app.SchreibAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeberFehler(antwort.Fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	return an.abschluss()
}
