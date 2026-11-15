package main

// **Die Kartenerkennung auf der Kommandozeile** — Werkstück A13.3 von Phase
// 5.3 (#1452).
//
// **Sie schreibt nichts** — weder eine Karte noch eine Sammlungsmenge. Was
// zurückkommt, sind gelesene Stellen und die Drucke, die dazu passen.

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// kierkennungbefehl liest ein Kartenfoto.
func kierkennungbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	const hilfe = "Aufruf: mse ki-erkennung <bilddatei> <anbieter> <modell>"

	if len(rest) != 3 {
		an.klage("ki-erkennung braucht die Bilddatei, den Anbieter und das "+
			"Modell.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.KIAnwendungUeber(db).KIKarteErkennen(ctx,
			app.KIErkennungsbezug{
				Pfad: rest[0], Anbieter: rest[1], Alias: rest[2],
			})
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		return erkennungszeilen(an, antwort)
	})
}

// erkennungszeilen schreibt, was gelesen wurde und was passt.
//
// **Erst das Gelesene, dann die Kandidaten** — in dieser Reihenfolge, weil
// die zweite aus der ersten folgt. Wer die Kandidaten zuerst sähe, hielte sie
// für eine Auskunft des Modells.
func erkennungszeilen(an *schreiber, antwort app.KIErkennungAntwort) int {
	for _, beobachtung := range antwort.Beobachtungen {
		zustand := "gelesen"
		if !beobachtung.Lesbar {
			zustand = "nicht lesbar"
		}

		an.sag("gelesen\t%s\t%s\t%s\n", beobachtung.Art, zustand, beobachtung.Wert)
	}

	if len(antwort.NichtGelesen) > 0 {
		an.sag("ohne\t%s\n", strings.Join(antwort.NichtGelesen, ", "))
	}

	for _, befund := range antwort.Befunde {
		an.sag("befund\t%s\t%s\n", befund.Art, befund.Meldung)
	}

	// **Die Zahl steht bei den Kandidaten und nicht darunter** — sie ist die
	// Unsicherheit, und „mindestens" ist etwas anderes als „genau".
	mindestens := ""
	if antwort.Weitere {
		mindestens = "mindestens "
	}

	an.sag("drucke\t%s%d\n", mindestens, antwort.Drucke)

	for _, kandidat := range antwort.Kandidaten {
		an.sag("kandidat\t%s\t%s\t%s\t%s\t%s\n", kandidat.Kennung,
			kandidat.Name, kandidat.Setcode, kandidat.Nummer,
			strings.Join(kandidat.Gestuetzt, ","))
	}

	return CodeGut
}
