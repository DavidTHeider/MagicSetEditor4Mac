// **Ein Kartenbild von der Platte einsetzen** —
// #1291.
//
// # Warum es ohne Fenster geht
//
// Der Dateidialog steckt allein in `Bindungen.BildSetzen`. `App.BildAusDatei`
// nimmt einen Pfad entgegen — es gibt sie seit
// #246, weil
// das Fenster Dateien auch **ziehen und ablegen** lässt.
//
// **Der Zuschnitt ist der vorgeschlagene**: ohne Fenster zieht niemand einen
// Rahmen, und genommen wird der, den MSEs Regel aus Quelle und Feldfläche
// errechnet.

package main

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// bildsetzung sagt, was gesetzt wurde.
type bildsetzung struct {
	Feld         string `json:"feld"`
	QuelleBreite int    `json:"quelleBreite"`
	QuelleHoehe  int    `json:"quelleHoehe"`
	ZielBreite   int    `json:"zielBreite"`
	ZielHoehe    int    `json:"zielHoehe"`
}

// bildSetzen liest eine Bilddatei und schreibt sie in ein Kartenfeld.
func bildSetzen(
	ctx context.Context, db *sql.DB, an *schreiber, setID, karteID int64, feld, pfad string,
) int {
	naht, err := app.Bilddateinaht(db, domain.UrheberKommandozeile)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	gesetzt, fehler := naht.Einsetzen(ctx, setID, karteID, feld, pfad)
	if fehler != nil {
		an.klageUeber(fehler)

		if errors.Is(fehler, domain.ErrValidation) {
			return CodeBedienung
		}

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, bildsetzung{
			Feld:         gesetzt.Feld,
			QuelleBreite: gesetzt.QuelleBreite, QuelleHoehe: gesetzt.QuelleHoehe,
			ZielBreite: gesetzt.ZielBreite, ZielHoehe: gesetzt.ZielHoehe,
		})
	}

	an.sag("%s gesetzt — %d × %d auf %d × %d zugeschnitten.\n",
		gesetzt.Feld, gesetzt.QuelleBreite, gesetzt.QuelleHoehe,
		gesetzt.ZielBreite, gesetzt.ZielHoehe)

	return an.abschluss()
}
