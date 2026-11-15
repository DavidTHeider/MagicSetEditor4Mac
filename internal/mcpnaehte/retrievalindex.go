package mcpnaehte

// **Der lokale Suchindex am MCP-Server** — Werkstück A15 von Phase 5.3
// (#1514).
//
// **Über dieselbe Anwendung wie jede andere Fläche** (§16): was das Fenster
// findet, findet ein Modell auch — und mit derselben Herkunftsangabe.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// indexauskunft reicht die zwei lesenden Wege an den Server.
type indexauskunft struct{ db *sql.DB }

func (i indexauskunft) Indexstand(ctx context.Context) (mcp.Indexstand, error) {
	antwort := app.RetrievalAnwendungUeber(i.db).Indexstand(ctx)
	if antwort.Fehler != nil {
		return mcp.Indexstand{}, antwort.Fehler.AlsFehler()
	}

	staende := make([]mcp.Korpusstand, 0, len(antwort.Manifest.Korpusstaende))
	for _, stand := range antwort.Manifest.Korpusstaende {
		staende = append(staende, mcp.Korpusstand{
			Quelle: stand.Quelle, Revision: stand.Revision, Datum: stand.Datum,
		})
	}

	return mcp.Indexstand{
		Aktiv:    antwort.Aktiv,
		Modell:   antwort.Manifest.Modell,
		Laufzeit: antwort.Manifest.Laufzeit,
		Staende:  staende,
		Suchbar:  antwort.Suchbar, Grund: antwort.Grund,
		Aufbau: mcp.Indexaufbaulage{
			Laeuft:      antwort.Aufbau.Laeuft,
			Geschrieben: antwort.Aufbau.Geschrieben,
			Fehler:      antwort.Aufbau.Fehler,
		},
	}, nil
}

func (i indexauskunft) Indexsuche(
	ctx context.Context, frage string, grenze int, filter mcp.Indexfilter,
) (mcp.Indexsuche, error) {
	antwort := app.RetrievalAnwendungUeber(i.db).Indexsuche(ctx, frage, grenze,
		app.IndexsuchfilterDTO{
			Quellen: filter.Quellen, Sprachen: filter.Sprachen,
			Sets: filter.Sets, Farben: filter.Farben,
			Farbidentitaet: filter.Farbidentitaet, Typen: filter.Typen,
			ManawertVon: filter.ManawertVon, ManawertBis: filter.ManawertBis,
			Format: filter.Format, Formatlage: filter.Formatlage,
			Formatstand: filter.Formatstand,
		})
	if antwort.Fehler != nil {
		return mcp.Indexsuche{}, antwort.Fehler.AlsFehler()
	}

	treffer := make([]mcp.Indextreffer, 0, len(antwort.Treffer))
	for _, einzelner := range antwort.Treffer {
		treffer = append(treffer, mcp.Indextreffer{
			Rang: einzelner.Rang, Kennung: einzelner.Kennung,
			Quelle: einzelner.Quelle, Adresse: einzelner.Bezugsadresse,
			Titel: einzelner.Titel, Text: einzelner.Text,
			Ausschnitt: einzelner.Ausschnitt, Sprache: einzelner.Sprache,
			Regelnummer: einzelner.Regelnummer, Urheber: einzelner.Urheber,
			Quellrevision: einzelner.Quellrevision,
			Quelldatum:    einzelner.Quelldatum,
			Abgerufen:     einzelner.Abgerufen,
		})
	}

	return mcp.Indexsuche{
		Treffer: treffer, Modus: antwort.Modus, Grund: antwort.Grund,
	}, nil
}
