package main

// **Ein Set vorab ansehen** — FMT-03.AK-01
// (#1692).
//
// `mse set-vorschau <pfad>` sagt, was `mse oeffnen` mit der Datei täte, und
// schreibt nichts. Einordnung nach §16: **erweiternd** — ein Befehl kommt
// hinzu.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// setvorschaubefehl liest die Datei wie `oeffnen` und zeigt die Vorschau.
//
// **Eine Zeile je Aussage**: zuerst Name, Karten und Spiel/Stil, dann ob sie
// schon offen ist, dann je Befund und je fehlendem Paket eine Zeile — so
// lässt sich mit `grep fehlt` herausziehen, was vor dem Einfahren zu beziehen
// ist.
func setvorschaubefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 1 {
		an.klage("set-vorschau braucht genau einen Pfad.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg, err := app.PaketnahtUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		vorschau, fehler := weg.Setvorschau(ctx, db, rest[0])
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, vorschau)
		}

		an.sag("%s\t%d Karten\t%s/%s\n", vorschau.Name, vorschau.Karten, vorschau.Spiel, vorschau.Stil)

		if vorschau.SchonOffen != "" {
			an.sag("schon offen\t%s\n", vorschau.SchonOffen)
		}

		for _, befund := range vorschau.Befunde {
			an.sag("befund\t%s\n", befund)
		}

		for _, bedarf := range vorschau.Bedarf {
			an.sag("%s\t%s braucht %s\n", bedarf.Art, bedarf.Paket, bedarf.Braucht)
		}

		return an.abschluss()
	})
}
