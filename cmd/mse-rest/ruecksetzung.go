package main

// **Die abgestufte Rücksetzung am Dienst** — Werkstück O4 von Phase 5.2
// (#1428).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// ruecksetzungsauskunft reicht die drei Wege an den Dienst.
//
// **Ohne den Neuanfang** (G09, #2273): der Dienst hält seinen Bestand über
// seine ganze Laufzeit, und ein Tausch ließe jede seiner Nähte an einer
// geschlossenen Verbindung zurück. Deshalb braucht sie keinen Ort.
type ruecksetzungsauskunft struct {
	db *sql.DB
}

func (r ruecksetzungsauskunft) Ruecksetzungsarten(context.Context) (any, error) {
	return app.RuecksetzungsAnwendungOhneTausch(r.db).Ruecksetzungsarten(), nil
}

func (r ruecksetzungsauskunft) Ruecksetzungsvorschau(
	ctx context.Context, art, profilID, nutzerID string,
) (any, error) {
	antwort := app.RuecksetzungsAnwendungOhneTausch(r.db).
		Ruecksetzungsvorschau(ctx, art, profilID, nutzerID)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}

func (r ruecksetzungsauskunft) Zuruecksetzen(
	ctx context.Context, art, profilID, nutzerID, name, nachweis, revision string,
) (any, error) {
	antwort := app.RuecksetzungsAnwendungOhneTausch(r.db).
		Zuruecksetzen(ctx, art, profilID, nutzerID, name, nachweis, revision)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}
