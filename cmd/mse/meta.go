package main

// **Wer antwortet hier, auf der Kommandozeile** — Werkstück O9 von Phase 5.2
// (#1433).
//
// **Ohne Bestand**, und das ist der Punkt: die Frage „welche Fassung bist du"
// muss auch dann eine Antwort haben, wenn der Bestand nicht aufgeht.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func metabefehl(ctx context.Context, an *schreiber, rest []string) int {
	if len(rest) != 0 {
		an.klage("meta nimmt keine Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	angaben := app.NeueApp("").Meta(app.MitDienstmodus(ctx, domain.ModusKommandozeile))

	if an.alsJSON {
		return alsJSONSchreiben(an, angaben)
	}

	an.sag("fassung\t%s\n", angaben.Fassung)
	an.sag("plattform\t%s\n", angaben.Plattform)
	an.sag("go\t%s\n", angaben.Go)
	an.sag("flaeche\t%s\n", angaben.Flaeche)
	an.sag("stand\t%s\n", angaben.Bau.Stand)
	an.sag("gebaut\t%s\n", angaben.Bau.Gebaut)
	an.sag("veraendert\t%t\n", angaben.Bau.Veraendert)
	an.sag("bindungsvertrag\t%s\n", angaben.Bindungsvertrag)

	return an.abschluss()
}
