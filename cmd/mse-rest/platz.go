package main

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Der Platz über den Dienst** — Werkstück 4 von Phase 4.13
// (#1352).
//
// **Über den Ort und nicht über die Verbindung**: gezählt werden Dateien.
type fremdplatz struct{ ort string }

func (f fremdplatz) Platz(ctx context.Context) (rest.Platzstand, error) {
	return alsPlatzstand(app.PlatzflaecheUeber(f.ort).Platz(ctx))
}

func (f fremdplatz) PlatzFreigeben(ctx context.Context, posten string) (
	rest.Platzstand, error,
) {
	return alsPlatzstand(app.PlatzflaecheUeber(f.ort).PlatzFreigeben(ctx, posten))
}

// alsPlatzstand übersetzt die Antwort der Fassade.
func alsPlatzstand(antwort app.PlatzAntwort) (rest.Platzstand, error) {
	if antwort.Fehler != nil {
		return rest.Platzstand{}, app.AlsFehler(antwort.Fehler)
	}

	// **Eine leere Liste und kein `nil`** (§3): `nil` wird zu `null` im JSON.
	zeilen := make([]rest.Platzzeile, 0, len(antwort.Posten))
	for _, posten := range antwort.Posten {
		zeilen = append(zeilen, rest.Platzzeile{
			Posten: posten.Posten, Dateien: posten.Dateien,
			Bytes: posten.Bytes, Raeumbar: posten.Raeumbar,
		})
	}

	return rest.Platzstand{
		Posten: zeilen, Nachladefehler: alsNachladebefund(antwort.Nachladefehler),
	}, nil
}
