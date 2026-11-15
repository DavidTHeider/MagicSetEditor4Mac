// **Abschluss und Protokoll eines Turniers** — DOM-04.AK-04
// (#1699),
// Vermerk 0139.

package main

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// turnierAbschluss schließt ab oder eröffnet wieder — **die Fassung liest die
// Kommandozeile selbst** (§7, Regel 2): sie ist einschrittig.
func turnierAbschluss(
	ctx context.Context, flaeche *app.Turnierflaeche, an *schreiber, rest []string, abschliessen bool,
) int {
	befehl := "turnier wiedereroeffnen"
	weg := flaeche.TurnierWiedereroeffnen

	if abschliessen {
		befehl = "turnier abschliessen"
		weg = flaeche.TurnierAbschliessen
	}

	if len(rest) < 1 {
		an.klage("%s braucht die Kennung eines Turniers.\n%s", befehl, hilfe)

		return CodeBedienung
	}

	gelesen := flaeche.Turnier(ctx, rest[0])
	if gelesen.Fehler != nil {
		return klageUeberNaht(an, gelesen.Fehler)
	}

	antwort := weg(ctx, rest[0], gelesen.Turnier.Fassung)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if abschliessen {
		return getan(an, nil, "Abgeschlossen.\n")
	}

	return getan(an, nil, "Wiedereröffnet.\n")
}

// turnierProtokoll nennt, was an einem Turnier geschah — eine Zeile je
// Vorgang: Zeit, Vorgang, Angaben, Paarung.
func turnierProtokoll(
	ctx context.Context, flaeche *app.Turnierflaeche, an *schreiber, rest []string,
) int {
	if len(rest) < 1 {
		an.klage("turnier protokoll braucht die Kennung eines Turniers.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.TurnierProtokoll(ctx, rest[0])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	for _, eintrag := range antwort.Eintraege {
		an.sag("%s\t%s\t%s\t%s\n", eintrag.Zeit, eintrag.Vorgang, eintrag.Angaben, eintrag.PaarungID)
	}

	return an.abschluss()
}
