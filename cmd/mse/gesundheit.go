package main

// **Der Betriebszustand auf der Kommandozeile** — Werkstück O9 von Phase 5.2
// (#1433).
//
// # Der Rückgabewert ist die Antwort
//
// **0, wenn er arbeiten kann, 1 sonst.** Das ist die Form, die ein Skript
// liest — dieselbe Aussage, die der Dienst als 200 oder 503 gibt. Wer den
// Grund will, liest die Ausgabe oder nimmt `--json`.
//
// **Ein Fehlschlag des Rufs ist etwas anderes** als „nicht bereit": er gibt
// `CodeFachlich`, und der Unterschied ist der zwischen „ich kann gerade
// nicht" und „ich konnte nicht einmal nachsehen".

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// CodeNichtBereit heißt: der Prozess lebt, kann aber nicht arbeiten.
const CodeNichtBereit = 1

func gesundheitsbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 0 {
		an.klage("gesundheit nimmt keine Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		anwendung := app.NeueApp("")
		anwendung.MitGesundheit(app.GesundheitsnahtUeber(db))

		antwort := anwendung.Gesundheit(ctx)

		if an.alsJSON {
			if code := alsJSONSchreiben(an, antwort); code != 0 {
				return code
			}

			return zustandscode(antwort.Bereit)
		}

		an.sag("lebt\t%t\n", antwort.Lebt)
		an.sag("bereit\t%t\n", antwort.Bereit)

		if antwort.Grund != "" {
			an.sag("grund\t%s\n", antwort.Grund)
		}

		for _, fremd := range antwort.Fremde {
			an.sag("%s\t%t\t%s\n", fremd.Dienst, fremd.Erreichbar, fremd.Grund)
		}

		if code := an.abschluss(); code != 0 {
			return code
		}

		return zustandscode(antwort.Bereit)
	})
}

// zustandscode macht aus „bereit" einen Rückgabewert.
func zustandscode(bereit bool) int {
	if bereit {
		return 0
	}

	return CodeNichtBereit
}
