// **Liste und Entwürfe des Handels im Dienst** — TRADE-01
// (#1715).

package main

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// HandelUebersicht gibt die Vorgänge dieses Profils.
func (l leitungsweg) HandelUebersicht(
	ctx context.Context, filter rest.Handelsfilter,
) ([]rest.Handelsuebersichtszeile, error) {
	antwort := app.PeerflaecheUeber(l.db).HandelUebersicht(ctx, app.HandelsfilterDTO{
		Partner: filter.Partner, Lagen: filter.Lagen, Suche: filter.Suche, Marke: filter.Marke,
		Seit: filter.Seit, Bis: filter.Bis, Ordnung: filter.Ordnung, Archiv: filter.Archiv,
	})
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	zeilen := make([]rest.Handelsuebersichtszeile, 0, len(antwort.Zeilen))

	for _, zeile := range antwort.Zeilen {
		partner := make([]rest.Handelspartner, 0, len(zeile.Gegenueber))
		for _, gegenueber := range zeile.Gegenueber {
			partner = append(partner, rest.Handelspartner{Kennung: gegenueber.Kennung, Name: gegenueber.Name})
		}

		zeilen = append(zeilen, rest.Handelsuebersichtszeile{
			Vorgang: restvorgangAus(zeile.Vorgang), Lage: zeile.Lage, Naechster: zeile.Naechster,
			Gegenueber: partner, Karten: zeile.Karten, Marken: zeile.Marken,
			Angelegt: zeile.Angelegt, Geaendert: zeile.Geaendert,
			Archiviert: zeile.Archiviert, Unbestaetigt: zeile.Unbestaetigt,
		})
	}

	return zeilen, nil
}

// HandelEntwurfAendern schreibt einen Entwurf neu.
func (l leitungsweg) HandelEntwurfAendern(
	ctx context.Context, wunsch rest.Handelswunsch,
) (rest.Handelsvorgang, error) {
	return alsRestvorgang(app.PeerflaecheUeber(l.db).HandelEntwurfAendern(ctx, app.Handelswunsch{
		Vorgang: wunsch.Vorgang, Art: wunsch.Art, Gegenueber: wunsch.Gegenueber,
		Positionen: alsAppPositionen(wunsch.Positionen), BewertungGeteilt: wunsch.BewertungGeteilt,
	}))
}

// HandelKopieren legt aus einem Vorgang einen neuen Entwurf an.
func (l leitungsweg) HandelKopieren(ctx context.Context, vorgang string) (rest.Handelsvorgang, error) {
	return alsRestvorgang(app.PeerflaecheUeber(l.db).HandelKopieren(ctx, vorgang))
}

// HandelEntwurfLoeschen löscht einen Entwurf.
func (l leitungsweg) HandelEntwurfLoeschen(ctx context.Context, vorgang string) error {
	return app.PeerflaecheUeber(l.db).HandelEntwurfLoeschen(ctx, vorgang).Fehler.AlsFehler()
}

// HandelArchivieren legt einen beendeten Vorgang ins Archiv oder holt ihn
// heraus.
func (l leitungsweg) HandelArchivieren(ctx context.Context, vorgang string, an bool) error {
	return app.PeerflaecheUeber(l.db).HandelArchivieren(ctx, vorgang, an).Fehler.AlsFehler()
}
