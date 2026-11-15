package main

// **Die Datenpflege auf der Kommandozeile** — Werkstück O8 von Phase 5.2
// (#1432).
//
// # Zwei Befehle, und der zweite trägt die Revision
//
//   - `mse pflege-vorschau` sieht nach und **ändert nichts**. Der
//     Rückgabewert ist 1, wenn etwas dasteht — ein Skript soll den Fall
//     erkennen, ohne den Text zu lesen.
//   - `mse pflege <zeitpunkt> <art>:<revision> …` führt aus.
//
// **Die Revision steht im Argument, und das ist Absicht.** Sie ist die
// Bindung an das, was die Vorschau gezeigt hat; ein Befehl, der sie sich
// selbst holte, führte gegen einen Bestand aus, den niemand angesehen hat.

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func pflegevorschaubefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 0 {
		an.klage("pflege-vorschau nimmt keine Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.PflegeAnwendungUeber(db, bestandsort).Pflegevorschau(ctx)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			if code := alsJSONSchreiben(an, antwort); code != 0 {
				return code
			}

			return befundscode(len(antwort.Befunde))
		}

		an.sag("erstellt\t%s\n", antwort.Erstellt)
		an.sag("gueltigbis\t%s\n", antwort.Gueltigbis)

		for _, befund := range antwort.Befunde {
			an.sag("%s\t%s\t%d\t%s\n",
				befund.Art, befund.Klasse, befund.Anzahl, befund.Revision)
		}

		if code := an.abschluss(); code != 0 {
			return code
		}

		return befundscode(len(antwort.Befunde))
	})
}

// befundscode macht aus der Zahl der Befunde einen Rückgabewert.
//
// **Ein Befund ist kein Fehler und trotzdem nicht 0**: wer ein Skript baut,
// das danach weiterläuft, soll hier anhalten und nachsehen.
func befundscode(anzahl int) int {
	if anzahl > 0 {
		return CodeFachlich
	}

	return CodeGut
}

func pflegebefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 2 {
		an.klage("pflege braucht den Zeitpunkt der Vorschau und "+
			"mindestens eine Maßnahme als art:revision.\n%s", hilfe)

		return CodeBedienung
	}

	wahl := make([]app.PflegewahlDTO, 0, len(rest)-1)

	for _, roh := range rest[1:] {
		art, revision, geteilt := strings.Cut(roh, ":")
		if !geteilt || art == "" || revision == "" {
			an.klage("%q ist keine Maßnahme — erwartet wird art:revision.\n", roh)

			return CodeBedienung
		}

		wahl = append(wahl, app.PflegewahlDTO{Art: art, Revision: revision})
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.PflegeAnwendungUeber(db, bestandsort).Pflegen(ctx, rest[0], wahl)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			if code := alsJSONSchreiben(an, antwort); code != 0 {
				return code
			}

			return bilanzcode(antwort)
		}

		if antwort.Gesichert != "" {
			an.sag("gesichert\t%s\n", antwort.Gesichert)
		}

		for _, zeile := range antwort.Zeilen {
			an.sag("%s\t%s\t%d\t%s\n", zeile.Art, zeile.Ausgang, zeile.Anzahl, zeile.Grund)
		}

		an.sag("vorher\t%d\n", len(antwort.Vorher))
		an.sag("nachher\t%d\n", len(antwort.Nachher))

		if code := an.abschluss(); code != 0 {
			return code
		}

		return bilanzcode(antwort)
	})
}

// bilanzcode sagt mit dem Rückgabewert, ob alles durchging.
//
// **Ein Konflikt ist kein Erfolg.** Wer ein Skript baut, das nach der Pflege
// weiterläuft, soll bei einer Maßnahme anhalten, die nicht gelaufen ist.
func bilanzcode(antwort app.PflegebilanzAntwort) int {
	for _, zeile := range antwort.Zeilen {
		switch domain.Pflegeausgang(zeile.Ausgang) {
		case domain.PflegeKonflikt, domain.PflegeGescheitert:
			return CodeFachlich
		}
	}

	return CodeGut
}
