// **Freigaben an der Kommandozeile** — Werkstück 11 von Phase 4.12
// (#1351).
//
// **Drei Befehle, alle erweiternd** (§16).

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// freigabenbefehl nennt die Freigaben eines Profils.
func freigabenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("freigaben braucht genau eine Profilkennung.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return freigabenAusgeben(an,
			app.NutzerflaecheUeber(db).Freigaben(ctx, rest[0]))
	})
}

// freigabesetzenbefehl legt eine Freigabe an oder schreibt ihre Rolle fort.
func freigabesetzenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 3 {
		an.klage("freigabe-setzen braucht Profilkennung, Nutzerkennung und "+
			"Rolle — leser, mitarbeiter oder verwalter.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return freigabenAusgeben(an, app.NutzerflaecheUeber(db).FreigabeSetzen(
			ctx, app.FreigabeangabenDTO{
				ProfilID: rest[0], NutzerID: rest[1], Rolle: rest[2],
			}))
	})
}

// freigabeloeschenbefehl zieht eine Freigabe zurück.
func freigabeloeschenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("freigabe-loeschen braucht Profilkennung und "+
			"Nutzerkennung.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return freigabenAusgeben(an,
			app.NutzerflaecheUeber(db).FreigabeLoeschen(ctx, rest[0], rest[1]))
	})
}

// freigabenAusgeben schreibt eine Antwort — alle drei Wege geben denselben
// Stand zurück (§7, Regel 2).
func freigabenAusgeben(an *schreiber, antwort app.FreigabenAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeberFehler(antwort.Fehler)

		return CodeFachlich
	}

	nachladenMelden(an, antwort.Nachladefehler)

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	for _, freigabe := range antwort.Freigaben {
		an.sag("%s\t%s\t%s\n", freigabe.NutzerID, freigabe.Rolle, freigabe.Name)
	}

	return an.abschluss()
}
