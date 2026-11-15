// **Die Folgen einer Berichtigung** — DOM-04.AK-02
// (#1699).

package main

import (
	"context"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// turnierFolgen sagt, was eine Berichtigung verschöbe — **dieselben Wörter
// wie `turnier ergebnis`**, und nichts wird geschrieben.
func turnierFolgen(
	ctx context.Context, flaeche *app.Turnierflaeche, an *schreiber, rest []string,
) int {
	if len(rest) < 3 {
		an.klage("turnier folgen braucht Turnier, Paarung und einen Stand.\n%s", hilfe)

		return CodeBedienung
	}

	var stand *app.ErgebnisDTO

	if rest[2] != "weg" {
		gelesen, gut := spielstand(an, rest[2:])
		if !gut {
			return CodeBedienung
		}

		stand = gelesen
	}

	antwort := flaeche.TurnierKorrekturfolgen(ctx, rest[0], rest[1], stand)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	folgen := antwort.Folgen

	if len(folgen.Verschoben) == 0 {
		an.sag("Nichts verschiebt sich.\n")
	}

	for _, wechsel := range folgen.Verschoben {
		an.sag("%s\tPlatz %d → %d\t%d → %d Punkte\n", wechsel.Name,
			wechsel.PlatzVorher, wechsel.PlatzNachher, wechsel.PunkteVorher, wechsel.PunkteNachher)
	}

	// **Die späteren Runden bleiben**, und das wird gesagt.
	if len(folgen.SpaetereRunden) > 0 {
		nummern := make([]string, 0, len(folgen.SpaetereRunden))
		for _, nummer := range folgen.SpaetereRunden {
			nummern = append(nummern, strconv.Itoa(nummer))
		}

		an.sag("Nach dem alten Stand gepaart und unverändert: Runde %s\n",
			strings.Join(nummern, ", "))
	}

	return an.abschluss()
}
