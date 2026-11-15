package main

// **Über fremde Netze erreichbar sein, ohne Fenster** — D04
// (#2235).
//
// `docs/FAEHIGKEITEN.md` führte es als **Lücke und nicht als Ablehnung**:
// „Ein `mse fernhorchen` ist gebaut nicht vorhanden, und wer ihn schreibt,
// schreibt ihn in der Gestalt von `mse horchen`." Das ist diese Gestalt —
// [0121](../../docs/entscheidungen/0121-eine-leitung-ohne-fenster.md) gilt
// unverändert: ein Befehl, der bleibt, bis ⌃C kommt, und nichts horcht von
// selbst.
//
// # Was ausgegeben wird
//
// **Die Einladung, auf die Standardausgabe** — sie ist das Ergebnis, und wer
// sie seinem Gegenüber gibt, trägt sie mit `kontakt-adresse` ein. **Sie
// stirbt mit dem Befehl**: der Serverschlüssel wird nicht aufbewahrt
// (`internal/app/peerfern.go`).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// fernhorchenbefehl fährt `mse fernhorchen` — bis ⌃C.
func fernhorchenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 0 {
		an.klage("fernhorchen nimmt kein Argument — die Einladung kommt vom Netz.\n%s",
			hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an,
		func(ctx context.Context, db *sql.DB) int {
			flaeche := app.PeerflaecheUeber(db)

			antwort := flaeche.FernHorchStarten(ctx)
			if antwort.Fehler != nil {
				an.klageUeber(antwort.Fehler.AlsFehler())

				return CodeFachlich
			}

			an.sag("%s\n", antwort.Stand.Einladung)

			<-ctx.Done()

			// **Ohne den abgebrochenen Kontext** — derselbe Grund wie bei
			// `horchen`: er ist der Anlass, hier zu stehen.
			if beendet := flaeche.FernHorchBeenden(
				context.WithoutCancel(ctx)); beendet.Fehler != nil {
				an.klageUeber(beendet.Fehler.AlsFehler())

				return CodeFachlich
			}

			return an.abschluss()
		})
}
