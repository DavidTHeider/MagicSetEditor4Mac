// **Profilvorgaben an der Kommandozeile** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7.
//
// **Zwei Befehle, beide erweiternd** (§16): sie kommen hinzu, keiner ändert
// einen bestehenden.
//
// # Warum die Werte von der Standardeingabe kommen
//
// **Weil zwei der fünf Arten mehrwertig sind.** Als Wörter im Aufruf wären
// Präferenzdecks eine Liste ohne Ende — und ein Wert mit Leerzeichen („Kiste
// im Keller") ließe sich von zwei Werten nicht unterscheiden. Was hereinkommt,
// ist ein JSON-Feld von Zeichenketten, wie `mse vorgaben --json` es ausgibt.

package main

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// vorgabenbefehl nennt die Vorgaben eines Profils samt ihrem Angebot.
func vorgabenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("vorgaben braucht genau eine Profilkennung.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).Profilvorgaben(ctx, rest[0])
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		for _, stand := range antwort.Vorgaben {
			an.sag("%s\t%s\n", stand.Art, strings.Join(stand.Werte, ", "))
		}

		return an.abschluss()
	})
}

// vorgabesetzenbefehl ersetzt die Werte **einer** Art.
func vorgabesetzenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("vorgabe-setzen braucht Profilkennung und Art; die Werte "+
			"kommen als JSON-Feld von der Standardeingabe.\n%s", hilfe)

		return CodeBedienung
	}

	werte, code := vorgabewerteVonDerEingabe(an)
	if code != CodeGut {
		return code
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).ProfilvorgabeSetzen(ctx,
			app.VorgabeangabenDTO{ProfilID: rest[0], Art: rest[1], Werte: werte})
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		nachladenMelden(an, antwort.Nachladefehler)

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		return an.abschluss()
	})
}

// vorgabewerteVonDerEingabe liest ein JSON-Feld von Zeichenketten.
//
// **Ein Feld und kein Verbund**, weil es genau eine Angabe gibt: Kennung und
// Art stehen im Aufruf, und ein Verbund mit einem Feld darin wäre eine Hülle
// ohne Inhalt.
func vorgabewerteVonDerEingabe(an *schreiber) (werte []string, code int) {
	return jsonVonDerEingabe(an, "vorgabe-setzen", werte)
}
