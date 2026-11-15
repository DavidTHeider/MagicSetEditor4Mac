// **Nutzer und Profile an der Kommandozeile** — Werkstück 1 von Phase 4.12
// (#1351).

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// nutzerbefehl listet die Nutzer samt ihren Profilen.
func nutzerbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 0 {
		an.klage("nutzer nimmt keine weiteren Wörter.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).Nutzer(ctx)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		// **Ein Nutzer je Zeile, seine Profile eingerückt darunter.** Eine
		// Zeile je Profil mit wiederholtem Nutzernamen wäre kürzer zu
		// schreiben und schwerer zu lesen.
		for _, einer := range antwort.Nutzer {
			an.sag("%s\t%s\n", einer.ID, einer.Name)

			for _, eines := range einer.Profile {
				an.sag("\t%s\t%s\n", eines.ID, eines.Name)
			}
		}

		return an.abschluss()
	})
}

// nutzerneubefehl legt einen Nutzer samt seinem ersten Profil an.
func nutzerneubefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 1 || len(rest) > 2 {
		an.klage("nutzer-neu braucht einen Namen und darf ein Bild tragen.\n%s", hilfe)

		return CodeBedienung
	}

	bild := ""
	if len(rest) > 1 {
		bild = rest[1]
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).NutzerAnlegen(ctx, rest[0], bild)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("%s\t%s\n", antwort.Nutzer.ID, antwort.Nutzer.Name)

		return an.abschluss()
	})
}

// profilneubefehl legt ein weiteres Profil für einen Nutzer an.
func profilneubefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 2 {
		an.klage("profil-neu braucht die Nutzerkennung und einen Namen.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).ProfilAnlegen(ctx, rest[0], rest[1])
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("%s\t%s\n", antwort.Profil.ID, antwort.Profil.Name)

		return an.abschluss()
	})
}

// anbieterbefehl nennt die Anbieter, über die man sich anmelden kann.
func anbieterbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 0 {
		an.klage("anbieter nimmt keine weiteren Wörter.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).Anbieter(ctx)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		for _, name := range antwort.Anbieter {
			an.sag("%s\n", name)
		}

		return an.abschluss()
	})
}

// anmeldebefehl führt den Anmeldefluss über einen Anbieter.
//
// **Der Browser geht auf, und das steht auf der Packung**: `mse anmelden
// github` öffnet ein Fenster im Standardbrowser und wartet auf den Rückruf.
// Wer abbricht, bekommt einen Fehler und keinen halben Zustand.
func anmeldebefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 1 {
		an.klage("anmelden braucht genau einen Anbieter.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).Anmelden(ctx, rest[0])
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("%s\t%s\n", antwort.Nutzer.ID, antwort.Nutzer.Name)

		return an.abschluss()
	})
}

// kontenbefehl listet die Konten eines Nutzers oder löst eines.
//
// **Ein Befehl und nicht zwei** — dieselbe Form wie `profil`: wer eines lösen
// will, sieht vorher nach, welche es gibt.
func kontenbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 1 || len(rest) > 2 {
		an.klage("konten braucht eine Nutzerkennung und darf einen Anbieter tragen.\n%s",
			hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		flaeche := app.NutzerflaecheUeber(db)

		if len(rest) == 2 {
			if geloest := flaeche.KontoLoesen(ctx, rest[0], rest[1]); geloest.Fehler != nil {
				an.klageUeberFehler(geloest.Fehler)

				return CodeFachlich
			}
		}

		antwort := flaeche.Anbieterkonten(ctx, rest[0])
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		for _, eines := range antwort.Konten {
			an.sag("%s\t%s\n", eines.Anbieter, eines.Anzeigename)
		}

		return an.abschluss()
	})
}
