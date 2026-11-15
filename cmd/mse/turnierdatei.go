// **Turnierdatei und Bericht an der Kommandozeile** — DOM-04.AK-03
// (#1699),
// Vermerk 0140.
//
// Einordnung nach §16: **erweiternd** — drei Befehle kommen hinzu.

package main

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// turnierdateiAus gibt ein Turnier oder seinen Bericht aus — **ohne Ziel auf
// die Ausgabe**, damit es sich weiterreichen lässt, mit Ziel in die Datei.
func turnierdateiAus(
	ctx context.Context, flaeche *app.Turnierflaeche, an *schreiber, rest []string, bericht bool,
) int {
	befehl, alsText, alsDatei := "turnier ausgeben", flaeche.TurnierAusgabeText, flaeche.TurnierAusgabeSchreiben
	if bericht {
		befehl, alsText, alsDatei = "turnier bericht", flaeche.TurnierBerichtText, flaeche.TurnierBerichtSchreiben
	}

	if len(rest) < 1 {
		an.klage("%s braucht die Kennung eines Turniers.\n%s", befehl, hilfe)

		return CodeBedienung
	}

	var antwort app.TurnierdateiAntwort
	if len(rest) > 1 {
		antwort = alsDatei(ctx, rest[0], rest[1])
	} else {
		antwort = alsText(ctx, rest[0])
	}

	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	if antwort.Pfad != "" {
		an.sag("Geschrieben nach %s.\n", antwort.Pfad)
	} else {
		an.sag("%s", antwort.Inhalt)
	}

	return an.abschluss()
}

// turnierEinlesen legt ein Turnier aus einer Datei an — **samt Befunden**
// (§8): was die Datei mitbrachte und hier niemand kennt, wird genannt.
func turnierEinlesen(ctx context.Context, flaeche *app.Turnierflaeche, an *schreiber, rest []string) int {
	if len(rest) < 1 {
		an.klage("turnier einlesen braucht den Pfad einer Datei.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.TurnierEinlesenLesen(ctx, rest[0])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("Als Turnier %s eingelesen.\n", antwort.Kennung)

	for _, befund := range antwort.Befunde {
		an.sag("  Befund: %s\n", befund)
	}

	return an.abschluss()
}
