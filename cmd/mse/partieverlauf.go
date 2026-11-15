package main

// **Der Verlauf einer Partie an der Kommandozeile** — Werkstück G5 von Phase
// 5.5 (#1462).
//
// # Warum er hier steht und nicht in `partie.go`
//
// **Weil jene Datei an der Zeilengrenze steht** (§2) und weil der Schnitt der
// Frage folgt: dort geht es darum, was eine Partie **tut**, hier darum, was sie
// **war**.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// partieverlaufbefehl zeigt den Verlauf für einen Sitz.
func partieverlaufbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("partie-verlauf braucht die Partiekennung und eine "+
			"Sitzkennung.\n%s", hilfe)

		return CodeBedienung
	}

	return mitPartieflaeche(ctx, bestandsort, an,
		func(ctx context.Context, flaeche *app.Partieflaeche) int {
			return partieverlaufAusgeben(
				an, flaeche.PartieVerlauf(ctx, rest[0], rest[1]))
		})
}

// partieverlaufAusgeben schreibt den Verlauf.
func partieverlaufAusgeben(an *schreiber, antwort app.PartieverlaufAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	// **Auch die Leere bekommt einen Satz.** „Nichts" ohne Erklärung liest
	// sich wie ein Fehler; hier ist es eine Auskunft (§6, leere Zustände).
	if len(antwort.Verlauf) == 0 {
		an.sag("In dieser Partie ist noch nichts geschehen.\n")

		return CodeGut
	}

	for _, ereignis := range antwort.Verlauf {
		an.sag("%s  %s\n", ereignis.Folge, ereignis.Art)

		// **Was privat ist, steht eingerückt darunter** — und es steht nur
		// da, wenn es für diesen Sitz bestimmt war. Gefiltert hat `service`.
		if len(ereignis.Privat) > 0 {
			an.sag("      privat: %v\n", ereignis.Privat)
		}
	}

	return CodeGut
}
