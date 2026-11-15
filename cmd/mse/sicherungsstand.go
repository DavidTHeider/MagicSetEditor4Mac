package main

// **`mse bestand-sicherungsstand`** — OPS01
// (#2347):
// wann der Bestand zuletzt gesichert wurde, über welche Fläche, wie lange das
// her ist, und der freiwillige Plan. Dieselbe Antwort wie am Fenster, über MCP
// und am Dienst (§16).

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func sicherungsstandBefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 0 {
		an.klage("bestand-sicherungsstand nimmt keine Angabe.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestandsgriff(ctx, bestandsort, an, func(ctx context.Context, db **sql.DB) int {
		stand, fehler := app.BestandsnahtUeber(db, bestandsort).Sicherungsstand(ctx)
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, stand)
		}

		an.sag("%s\n", sicherungssatz(stand))

		if stand.Plan.Ordner != "" {
			an.sag("Plan: wöchentlich nach %s, sieben behalten — %s.\n", stand.Plan.Ordner, plansatz(stand.Plan))
		}

		return an.abschluss()
	})
}

// sicherungssatz sagt den Stand in einem Satz — **unbekannt heißt
// unbekannt**, und nie „heute", ohne dass eine Datei zurückgelesen wurde.
func sicherungssatz(stand app.SicherungsstandDTO) string {
	switch domain.Sicherungsstufe(stand.Stufe) {
	case domain.SicherungsstufeUnbekannt:
		return "Über diese Fassung wurde der Bestand noch nie gesichert — wann zuletzt, ist unbekannt."
	case domain.SicherungsstufeUnstimmig:
		return "Zuletzt gesichert am " + stand.Gesichert + " — der Zeitpunkt liegt nach der Uhr dieses Rechners, ein Alter lässt sich nicht sagen."
	case domain.SicherungsstufeAlt:
		return "Zuletzt gesichert am " + stand.Gesichert + " über " + stand.Flaeche + " (" + stand.Datei + "), vor " +
			tage(stand.Tage) + " — älter als 14 Tage, sichere bald."
	default:
		return "Zuletzt gesichert am " + stand.Gesichert + " über " + stand.Flaeche + " (" + stand.Datei + "), vor " +
			tage(stand.Tage) + ", zurückgelesen und geprüft."
	}
}

// plansatz sagt, wie es um den Plan steht.
func plansatz(plan app.SicherungsplanDTO) string {
	letzter := "noch nie nach Plan gesichert"
	if plan.Letzter != "" {
		letzter = "zuletzt am " + plan.Letzter
	}

	if plan.Faellig {
		return letzter + ", fällig: das Fenster sichert, sobald es offen ist"
	}

	return letzter
}

// tage gibt eine Zahl von Tagen als Wort.
func tage(zahl int) string {
	switch zahl {
	case 0:
		return "weniger als einem Tag"
	case 1:
		return "einem Tag"
	default:
		return strconv.Itoa(zahl) + " Tagen"
	}
}
