package main

// **Die freigegebenen Sichten im Dienst** — Werkstück `T3` von Phase 5.6
// (#1468).
//
// **Eine Naht und keine Fachlogik**: ob eine Freigabe gilt, entscheidet
// `internal/service`; hier wird übersetzt — Feld für Feld und nicht über
// `Marshal`.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// freigabeweg reicht die Sichten an den Dienst.
type freigabeweg struct{ db *sql.DB }

func (f freigabeweg) Sammlungsfreigaben(
	ctx context.Context, kennung string,
) (rest.Sammlungsfreigaben, error) {
	return alsRestfreigaben(
		app.PeerflaecheUeber(f.db).Sammlungsfreigaben(ctx, kennung))
}

func (f freigabeweg) SammlungFreigeben(
	ctx context.Context, freigabe rest.Sammlungsfreigabe,
) (rest.Sammlungsfreigaben, error) {
	return alsRestfreigaben(app.PeerflaecheUeber(f.db).SammlungFreigeben(ctx,
		app.SammlungsfreigabeDTO{
			Empfaenger: freigabe.Empfaenger,
			Umfang:     freigabe.Umfang,
			Zweck:      freigabe.Zweck,
			Auswahl:    freigabe.Auswahl,
			Behalten:   freigabe.Behalten,
			Ablauf:     freigabe.Ablauf,
		}))
}

func (f freigabeweg) SammlungsfreigabeWiderrufen(
	ctx context.Context, kennung, zweck, umfang string,
) (rest.Sammlungsfreigaben, error) {
	return alsRestfreigaben(app.PeerflaecheUeber(f.db).
		SammlungsfreigabeWiderrufen(ctx, kennung, zweck, umfang))
}

func (f freigabeweg) Freigabesicht(
	ctx context.Context, kennung, zweck string,
) ([]rest.Sichtkarte, error) {
	antwort := app.PeerflaecheUeber(f.db).Freigabesicht(ctx, kennung, zweck)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	zeilen := make([]rest.Sichtkarte, 0, len(antwort.Zeilen))

	for _, karte := range antwort.Zeilen {
		zeilen = append(zeilen, rest.Sichtkarte{
			Referenz:    karte.Referenz,
			Printing:    karte.Printing,
			Menge:       karte.Menge,
			Zustand:     karte.Zustand,
			Sprache:     karte.Sprache,
			Finish:      karte.Finish,
			Bezeichnung: karte.Bezeichnung,
		})
	}

	return zeilen, nil
}

// alsRestfreigaben macht aus der Antwort der Anwendung die des Dienstes.
func alsRestfreigaben(
	antwort app.SammlungsfreigabenAntwort,
) (rest.Sammlungsfreigaben, error) {
	if antwort.Fehler != nil {
		return rest.Sammlungsfreigaben{}, antwort.Fehler.AlsFehler()
	}

	freigaben := make([]rest.Sammlungsfreigabe, 0, len(antwort.Freigaben))

	for _, freigabe := range antwort.Freigaben {
		freigaben = append(freigaben, rest.Sammlungsfreigabe{
			Empfaenger: freigabe.Empfaenger,
			Umfang:     freigabe.Umfang,
			Zweck:      freigabe.Zweck,
			Auswahl:    freigabe.Auswahl,
			Behalten:   freigabe.Behalten,
			Ablauf:     freigabe.Ablauf,
			Erteilt:    freigabe.Erteilt,
		})
	}

	return rest.Sammlungsfreigaben{
		Freigaben: freigaben, Umfaenge: antwort.Umfaenge,
		Nachladefehler: alsNachladebefund(antwort.Nachladefehler),
	}, nil
}
