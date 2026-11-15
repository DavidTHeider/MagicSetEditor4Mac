package main

// **Der Tauschabgleich im Dienst** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// **Eine Naht und keine Fachlogik**: was ein Treffer ist, entscheidet
// `internal/domain`; hier wird übersetzt — Feld für Feld und nicht über
// `Marshal`.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// tauschweg reicht den Abgleich an den Dienst.
type tauschweg struct{ db *sql.DB }

func (t tauschweg) Tauschabgleich(
	ctx context.Context, kennung string, stand rest.Fremdstand, behalten, hoechstzahl int,
) (rest.Tauschabgleich, error) {
	antwort := app.PeerflaecheUeber(t.db).Tauschabgleich(ctx, kennung,
		app.FremdstandDTO{
			Angebot:  alsSichtkartenDTO(stand.Angebot),
			Wuensche: alsSichtkartenDTO(stand.Wuensche),
		}, behalten, hoechstzahl)
	if antwort.Fehler != nil {
		return rest.Tauschabgleich{}, antwort.Fehler.AlsFehler()
	}

	return rest.Tauschabgleich{
		Kennung: antwort.Kennung,
		Gebe:    alsRestkandidaten(antwort.Gebe),
		Bekomme: alsRestkandidaten(antwort.Bekomme),
		Vorschlag: rest.Tauschvorschlag{
			Gebe:    alsRestseite(antwort.Vorschlag.Gebe),
			Bekomme: alsRestseite(antwort.Vorschlag.Bekomme),
		},
		HatVorschlag: antwort.HatVorschlag,
		Befunde:      antwort.Befunde,
	}, nil
}

// alsSichtkartenDTO macht aus den Zeilen des Dienstes die der Naht.
func alsSichtkartenDTO(karten []rest.Sichtkarte) []app.SichtkarteDTO {
	dtos := make([]app.SichtkarteDTO, 0, len(karten))

	for _, karte := range karten {
		dtos = append(dtos, app.SichtkarteDTO{
			Referenz: karte.Referenz, Printing: karte.Printing,
			Menge: karte.Menge, Zustand: karte.Zustand, Sprache: karte.Sprache,
			Finish: karte.Finish, Bezeichnung: karte.Bezeichnung,
		})
	}

	return dtos
}

// alsRestkandidaten macht aus den Treffern die des Dienstes.
func alsRestkandidaten(treffer []app.TauschkandidatDTO) []rest.Tauschkandidat {
	zeilen := make([]rest.Tauschkandidat, 0, len(treffer))

	for _, kandidat := range treffer {
		zeilen = append(zeilen, rest.Tauschkandidat{
			Wunsch: kandidat.Wunsch, Angebot: kandidat.Angebot,
			Menge: kandidat.Menge, Fehlend: kandidat.Fehlend,
			Rang: kandidat.Rang, Art: kandidat.Art,
		})
	}

	return zeilen
}

// alsRestseite macht aus einer Seite die des Dienstes.
func alsRestseite(seite app.TauschseiteDTO) rest.Tauschseite {
	return rest.Tauschseite{
		Zeilen: alsRestkandidaten(seite.Zeilen), Stueck: seite.Stueck,
		Gedeckt: seite.Gedeckt, Offen: seite.Offen, Wert: seite.Wert,
		Waehrung: seite.Waehrung, Bepreist: seite.Bepreist,
	}
}
