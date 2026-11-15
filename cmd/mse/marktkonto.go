package main

// **Das eigene Konto an der Kommandozeile** — Werkstück `T11` von Phase 5.6
// (#1476).
//
// **Ein Befehl und nicht drei**: derselbe Grund wie am Fenster — Konto,
// Angebote und Bestellungen sind eine Übersicht.
//
// **Und es gibt keinen zweiten**, der schriebe. Was der Anbieter dafür
// anbietet, steht in
// [0122](../../docs/entscheidungen/0122-ein-fremdes-konto-wird-gelesen.md);
// gebaut ist davon nichts.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// marktkontobefehl fährt `mse markt-konto <quelle> [seite] [menge]`.
func marktkontobefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 1 || len(rest) > 3 {
		an.klage("markt-konto nimmt den Marktplatz und wahlfrei Seite und "+
			"Menge.\n%s", hilfe)

		return CodeBedienung
	}

	seite, menge, gelesen := blaetterangaben(an, rest[1:])
	if !gelesen {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an,
		func(ctx context.Context, db *sql.DB) int {
			antwort := app.PeerflaecheUeber(db).
				Marktkonto(ctx, rest[0], seite, menge)
			if antwort.Fehler != nil {
				an.klageUeber(antwort.Fehler.AlsFehler())

				return CodeFachlich
			}

			return kontostandSchreiben(an, antwort.Stand)
		})
}

// blaetterangaben liest Seite und Menge — **null heißt: die Vorgabe des
// Anbieters**, und nicht „alles".
func blaetterangaben(an *schreiber, rest []string) (seite, menge int, gut bool) {
	for stelle, roh := range rest {
		was := "die Seite"
		if stelle == 1 {
			was = "die Menge"
		}

		zahl, gelesen := kennung(an, was, roh)
		if !gelesen {
			return 0, 0, false
		}

		if stelle == 0 {
			seite = int(zahl)
		} else {
			menge = int(zahl)
		}
	}

	return seite, menge, true
}

// kontostandSchreiben gibt die Übersicht aus.
func kontostandSchreiben(an *schreiber, stand *app.KontostandDTO) int {
	if an.alsJSON {
		return alsJSONSchreiben(an, stand)
	}

	an.sag("konto\t%s\t%d\n", stand.Name, stand.NutzerID)

	for _, befund := range stand.Befunde {
		an.sag("!\t%s\n", befund)
	}

	for _, angebot := range stand.Angebote {
		an.sag("angebot\t%d\t%s\t%d\t%d %s\t%s\n", angebot.ID, angebot.Name,
			angebot.Menge, angebot.PreisCent, angebot.Waehrung,
			angebot.Eigennotiz)
	}

	// **Die Spalte »bewegt« ist eine Auskunft und keine Buchung** — sie sagt,
	// was nach der Regel des Anbieters geschehen **wäre**.
	for _, bestellung := range stand.Bestellungen {
		an.sag("bestellung\t%d\t%s\t%s\t%s\tbewegt=%v\n", bestellung.ID,
			bestellung.Code, bestellung.Zustand, bestellung.Rolle,
			bestellung.BewegtBestand)
	}

	return an.abschluss()
}
