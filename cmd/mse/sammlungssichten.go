// **Die Fragen an eine geordnete Sammlung, an der Kommandozeile** —
// Werkstück 5 von Phase 4.5
// (#1292,
// Vermerk [0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).

package main

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// besitzbefehl sagt, was von einem Druck da ist.
func besitzbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) == 0 {
		an.klage("besitz braucht eine Druckkennung.\n%s", hilfe)

		return CodeBedienung
	}

	return mitSichten(ctx, bestandsort, an, func(ctx context.Context, naht *app.Sammlungssichtennaht) int {
		stand, err := naht.Besitz(ctx, rest[0])
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, stand)
		}

		an.sag("%d besessen, %d gewünscht\n", stand.Stueck, stand.Gewuenscht)

		for _, zeile := range stand.Zeilen {
			an.sag("%d\t%s\t%d\t%s\t%s\t%s\n", zeile.ID, absichttext(zeile.Wunsch),
				zeile.Menge, zeile.Zustand, zeile.Lagerort, zeile.Bezeichnung)
		}

		return an.abschluss()
	})
}

// absichttext sagt, ob eine Zeile Besitz oder Wunsch ist.
func absichttext(wunsch bool) string {
	if wunsch {
		return "wunsch"
	}

	return "besitz"
}

// lagerortebefehl zählt die Lagerorte auf.
func lagerortebefehl(ctx context.Context, bestandsort string, an *schreiber) int {
	return mitSichten(ctx, bestandsort, an, func(ctx context.Context, naht *app.Sammlungssichtennaht) int {
		orte, err := naht.Lagerorte(ctx)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{"lagerorte": orte})
		}

		for _, ort := range orte {
			an.sag("%d\t%d\t%s\n", ort.Stueck, ort.Zeilen, ort.Lagerort)
		}

		return an.abschluss()
	})
}

// dublettenbefehl zeigt, wovon mehr da ist als gewollt.
func dublettenbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	behalten := 1

	if len(rest) > 0 {
		gelesen, fehler := strconv.Atoi(rest[0])
		if fehler != nil {
			an.klage("»%s« ist keine Zahl.\n%s", rest[0], hilfe)

			return CodeBedienung
		}

		behalten = gelesen
	}

	return mitSichten(ctx, bestandsort, an, func(ctx context.Context, naht *app.Sammlungssichtennaht) int {
		dubletten, err := naht.Dubletten(ctx, behalten)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{"dubletten": dubletten})
		}

		for _, dublette := range dubletten {
			an.sag("%d\t%d\t%s\n", dublette.Ueberzaehlig,
				dublette.Exemplar.Menge, dublette.Exemplar.Bezeichnung)
		}

		return an.abschluss()
	})
}

// wuenschebefehl zeigt die Wunschliste.
func wuenschebefehl(ctx context.Context, bestandsort string, an *schreiber) int {
	return mitSichten(ctx, bestandsort, an, func(ctx context.Context, naht *app.Sammlungssichtennaht) int {
		wuensche, err := naht.Wuensche(ctx)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{"wuensche": wuensche})
		}

		for _, wunsch := range wuensche {
			an.sag("%d\t%d\t%s\t%s\n", wunsch.ID, wunsch.Menge,
				wunsch.Zustand, wunsch.Bezeichnung)
		}

		return an.abschluss()
	})
}

// mitSichten baut die Naht über dem Bestand und führt die Arbeit aus.
func mitSichten(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Sammlungssichtennaht) int,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return arbeit(ctx, app.SichtenUeber(db))
	})
}
