package main

// **Die atomare Gruppe auf der Kommandozeile** — Werkstück A13.1b von Phase
// 5.3 (#1452).
//
// # Warum es diesen Befehl neben `ki-vorschlag --uebernehmen` gibt
//
// **Weil eine Schleife im Skript etwas anderes tut.** Drei Aufrufe sind drei
// Transaktionen; bricht der dritte ab, stehen zwei Karten geändert da. Das
// Issue nennt es „klarer Konflikt **ohne Teilbestand**", und ein Skript kann
// das nicht selbst herstellen.

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// kigruppebefehl übernimmt mehrere Vorschläge als eine Gruppe.
func kigruppebefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	const hilfe = "Aufruf: mse ki-gruppe <kennung=feld,feld> [<kennung=feld,…> …]"

	if len(rest) == 0 {
		an.klage("ki-gruppe braucht mindestens einen Vorschlag mit seinen "+
			"Feldern.\n%s", hilfe)

		return CodeBedienung
	}

	wahl := make([]app.KIGruppenwahlDTO, 0, len(rest))

	for _, teil := range rest {
		kennung, felder, gefunden := strings.Cut(teil, "=")
		if !gefunden || kennung == "" || felder == "" {
			an.klage("%q nennt keine Kennung mit Feldern — erwartet ist "+
				"»kennung=feld,feld«.\n%s", teil, hilfe)

			return CodeBedienung
		}

		wahl = append(wahl, app.KIGruppenwahlDTO{
			Kennung: kennung, Felder: strings.Split(felder, ","),
		})
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.KIAnwendungUeber(db).KIGruppeUebernehmen(ctx, wahl)
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		return gruppenzeilen(an, antwort)
	})
}

// gruppenzeilen schreibt, was geschrieben wurde.
//
// **Je Karte eine Zeile** — eine Gruppe, die nur „fertig" sagt, verschweigt,
// welches Feld wo gelandet ist. Und weil sie ganz oder gar nicht geht, ist
// jede Zeile hier eine, die wirklich steht.
func gruppenzeilen(an *schreiber, antwort app.KIGruppenantwort) int {
	an.sag("gruppe\t%d Karten\t%d Felder\n", antwort.Karten, antwort.Felder)

	for _, zeile := range antwort.Namen {
		an.sag("%d\t%s\n", zeile.KarteID, strings.Join(zeile.Felder, ", "))
	}

	return CodeGut
}
