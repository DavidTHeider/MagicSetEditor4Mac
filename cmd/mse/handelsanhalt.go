package main

// **Was die Seiten eines Tauschs ungefähr wert sind** — TRADE-02b
// (#1951).
//
// Beträge stehen in der kleinsten Einheit (EUR: Cent), wie bei
// `markt-konto` — eine Kommandozeile rechnet weiter, und ein Komma, das je
// Sprache ein anderes ist, stünde ihr dabei im Weg.

import (
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// anhaltSchreiben gibt Seiten, Differenzen und Hauptwährung aus.
func anhaltSchreiben(an *schreiber, antwort app.PreisanhaltAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort.Anhalt)
	}

	for _, seite := range antwort.Anhalt.Seiten {
		an.sag("seite\t%s\t%s\t%d\n", seite.Von, betraege(seite.Summen), seite.OhnePreis)
	}

	for _, differenz := range antwort.Anhalt.Differenzen {
		vollstaendig := "unvollstaendig"
		if differenz.Vollstaendig {
			vollstaendig = "vollstaendig"
		}

		an.sag("differenz\t%s\t%d\t%s\n", differenz.Waehrung, differenz.Wert, vollstaendig)
	}

	if antwort.Anhalt.Hauptwaehrung != "" {
		an.sag("hauptwaehrung\t%s\n", antwort.Anhalt.Hauptwaehrung)
	}

	return an.abschluss()
}

// betraege schreibt Summen als `EUR=2300,USD=360` — oder `-` ohne eine.
func betraege(summen []app.AnhaltspreisDTO) string {
	if len(summen) == 0 {
		return "-"
	}

	teile := make([]string, 0, len(summen))
	for _, summe := range summen {
		teile = append(teile, fmt.Sprintf("%s=%d", summe.Waehrung, summe.Wert))
	}

	return strings.Join(teile, ",")
}
