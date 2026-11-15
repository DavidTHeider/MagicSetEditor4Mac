// **Format und Vorlage eines Turniers** — DOM-04.AK-01 und AK-03
// (#1699).

package main

import (
	"context"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// Die zwei freiwilligen Wörter dieser Befehle.
const (
	formatwort           = "--format"
	mitTeilnehmendenwort = "--mit-teilnehmenden"
)

// formatAus nimmt `--format <id>` aus den Argumenten — **an jeder Stelle**,
// damit die Wörter davor und danach ihre Stelle behalten.
//
// **Ein `--format` ohne Wert ist ein Bedienfehler** und keine leere Angabe:
// wer das Wort schreibt, meint ein Format.
func formatAus(rest []string) (uebrig []string, format string, gut bool) {
	stelle := slices.Index(rest, formatwort)
	if stelle < 0 {
		return rest, "", true
	}

	if stelle+1 >= len(rest) {
		return nil, "", false
	}

	uebrig = append(slices.Clone(rest[:stelle]), rest[stelle+2:]...)

	return uebrig, rest[stelle+1], true
}

// turnierKopieren legt ein Turnier nach einer Vorlage an.
func turnierKopieren(
	ctx context.Context, flaeche *app.Turnierflaeche, an *schreiber, rest []string,
) int {
	mitTeilnehmenden := slices.Contains(rest, mitTeilnehmendenwort)
	rest = slices.DeleteFunc(slices.Clone(rest), func(wort string) bool {
		return wort == mitTeilnehmendenwort
	})

	if len(rest) < 2 {
		an.klage("turnier kopieren braucht die Kennung der Vorlage und einen Namen.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.TurnierKopieren(ctx, rest[0], rest[1], mitTeilnehmenden)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("%s\n", antwort.Turnier.ID)

	return an.abschluss()
}
