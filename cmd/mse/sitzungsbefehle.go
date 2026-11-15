// **Die Sitzung an der Kommandozeile** — ID-02
// (#1701).
//
// # Die Geheimnisse stehen nicht im Aufruf
//
// **Kennwort und Code kommen von der Standardeingabe**, je eine Zeile — ein
// Argument stünde in `~/.zsh_history` und in `ps aux`, wie beim Konto
// (`kontobefehle.go`).

package main

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// sitzungsbefehl nennt den festgehaltenen Stand der Sitzung.
func sitzungsbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 0 {
		an.klage("sitzung nimmt nichts an.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).Sitzungsstand(ctx)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("%s\t%s\n", antwort.Zustand, antwort.NutzerID)

		return CodeGut
	})
}

// kennwortwechselbefehl setzt ein neues Kennwort für den, der arbeitet —
// Zeile eins das bisherige (leer, wenn es keines gibt), Zeile zwei das neue.
func kennwortwechselbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 0 {
		an.klage("kennwort-wechseln nimmt nichts an; die Kennwörter kommen von der Standardeingabe.\n%s", hilfe)

		return CodeBedienung
	}

	zeilen, code := zeilenVonDerEingabe(an, "kennwort-wechseln")
	if code != CodeGut {
		return code
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return schreibantwortMelden(an,
			app.NutzerflaecheUeber(db).KennwortWechseln(ctx, zeilen[0], zeilen[1]))
	})
}

// kennwortwiederherstellbefehl setzt ein vergessenes Kennwort — Zeile eins
// der Code vom Zettel, Zeile zwei das neue Kennwort.
func kennwortwiederherstellbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 1 {
		an.klage("kennwort-wiederherstellen braucht einen Nutzer; Code und Kennwort kommen von der Standardeingabe.\n%s", hilfe)

		return CodeBedienung
	}

	zeilen, code := zeilenVonDerEingabe(an, "kennwort-wiederherstellen")
	if code != CodeGut {
		return code
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return schreibantwortMelden(an,
			app.NutzerflaecheUeber(db).KennwortWiederherstellen(ctx, rest[0], zeilen[0], zeilen[1]))
	})
}

// zeilenVonDerEingabe liest genau zwei Zeilen.
func zeilenVonDerEingabe(an *schreiber, befehl string) (zeilen []string, code int) {
	gelesen, code := kennwortVonDerEingabe(an, befehl)
	if code != CodeGut {
		return nil, code
	}

	zeilen = strings.Split(strings.ReplaceAll(gelesen, "\r\n", "\n"), "\n")
	if len(zeilen) != 2 {
		an.klage("%s liest zwei Zeilen von der Standardeingabe.\n%s", befehl, hilfe)

		return nil, CodeBedienung
	}

	return zeilen, CodeGut
}

// schreibantwortMelden sagt, ob es ging.
func schreibantwortMelden(an *schreiber, antwort app.SchreibAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeberFehler(antwort.Fehler)

		return CodeFachlich
	}

	an.sag("gesetzt\n")

	return CodeGut
}
