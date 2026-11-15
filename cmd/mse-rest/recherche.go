package main

// **Die Recherche am Dienst** — Werkstück A16 von Phase 5.3
// (#1515).
//
// **Über dieselbe Anwendung wie jede andere Fläche** (§16).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// rechercheauskunft reicht die drei Wege an den Dienst.
type rechercheauskunft struct{ db *sql.DB }

func (r rechercheauskunft) Recherche(
	ctx context.Context, fall, frage string, grenze int,
	sprache, korpusstand, fortsetzung string, freigegeben []string,
) (any, error) {
	antwort := app.RechercheAnwendungUeber(r.db).Recherchieren(ctx, fall,
		frage, grenze,
		app.RecherchefilterDTO{
			Sprache: sprache, Korpusstand: korpusstand,
			Freigegeben: freigegeben,
		},
		fortsetzung)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}

func (r rechercheauskunft) AehnlicheKarten(
	ctx context.Context, karteID int64, stand string, grenze int,
	sprache, fortsetzung string, freigegeben []string,
) (any, error) {
	antwort := app.RechercheAnwendungUeber(r.db).AehnlicheKarten(ctx, karteID,
		stand, grenze,
		app.RecherchefilterDTO{Sprache: sprache, Freigegeben: freigegeben},
		fortsetzung)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}

func (r rechercheauskunft) Kontextpaket(
	ctx context.Context, wunsch rest.Kontextwunsch,
) (any, error) {
	antwort := app.RechercheAnwendungUeber(r.db).KontextZusammenstellen(ctx,
		wunsch.Aufgabe, wunsch.Zielrevision, wunsch.Fall, wunsch.Frage,
		wunsch.Grenze, wunsch.Anbieter, wunsch.Modell, wunsch.Kontextgrenze,
		wunsch.Ausgabereserve, wunsch.HoechstensBelege, wunsch.Freigegeben)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}
