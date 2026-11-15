package main

// **Die Wege nach draußen an der Kommandozeile** — Werkstück `T8` von Phase
// 5.6 (#1473).
//
// # Warum auch hier geöffnet wird
//
// **Weil die Kommandozeile eine reguläre Fläche ist** (§16). Dass ein Befehl
// einen Browser öffnet, ist dabei keine Ausnahme: er tut es auf **diesen**
// Aufruf hin und endet danach — das ist genau der Unterschied zu einem
// Prozess, der bis ⌃C horcht.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// linksicht ist ein Weg, wie die Kommandozeile ihn schreibt.
//
// **Eine eigene Gestalt** — die Wails-DTO ist **innen**, diese Fläche
// **außen** (0033, Regel 5).
type linksicht struct {
	Anbieter string `json:"anbieter"`
	Art      string `json:"art"`
	Herkunft string `json:"herkunft"`
	Adresse  string `json:"adresse"`
}

// marktlinksbefehl fährt `mse markt-links <printing|--name <name>>`.
func marktlinksbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 && len(rest) != 2 {
		an.klage("markt-links nimmt eine Druckkennung und wahlfrei einen "+
			"Namen.\n%s", hilfe)

		return CodeBedienung
	}

	name := ""
	if len(rest) == 2 {
		name = rest[1]
	}

	return mitBestand(ctx, bestandsort, an,
		func(ctx context.Context, db *sql.DB) int {
			antwort := app.PeerflaecheUeber(db).Marktlinks(ctx, rest[0], name)
			if antwort.Fehler != nil {
				an.klageUeber(antwort.Fehler.AlsFehler())

				return CodeFachlich
			}

			return linkeSchreiben(an, antwort.Links)
		})
}

// linkeSchreiben gibt die Wege aus.
func linkeSchreiben(an *schreiber, links []app.MarktlinkDTO) int {
	sichten := make([]linksicht, 0, len(links))

	for _, link := range links {
		sichten = append(sichten, linksicht{
			Anbieter: link.Anbieter, Art: link.Art,
			Herkunft: link.Herkunft, Adresse: link.Adresse,
		})
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, sichten)
	}

	for _, sicht := range sichten {
		an.sag("%s\t%s\t%s\n", sicht.Anbieter, sicht.Art, sicht.Adresse)
	}

	return an.abschluss()
}

// marktoeffnenbefehl fährt `mse markt-oeffnen <anbieter> <adresse>`.
func marktoeffnenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("markt-oeffnen nimmt den Anbieter und die Adresse.\n%s",
			hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an,
		func(ctx context.Context, db *sql.DB) int {
			antwort := app.PeerflaecheUeber(db).
				MarktlinkOeffnen(ctx, rest[0], rest[1])
			if antwort.Fehler != nil {
				an.klageUeber(antwort.Fehler.AlsFehler())

				return CodeFachlich
			}

			an.sag("Geöffnet: %s\n", rest[1])

			return an.abschluss()
		})
}
