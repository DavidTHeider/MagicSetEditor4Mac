// **MTG JSON an der Kommandozeile** —
// #1291.
//
// # Warum es ohne Fenster geht
//
// Der Dateidialog steckt allein in den Bindungen: `Bindungen.MTGJSONOeffnen`
// fragt nach einer Datei, `Bindungen.MTGJSONVerzeichnisOeffnen` nach einem
// Ordner, `Bindungen.MTGJSONSchreiben` nach einem Ort zum Speichern. Der Weg
// dahinter nimmt einen Pfad entgegen und kennt kein Fenster.
//
// **Und ein Weg für beides**: `mtgjson-lesen` nimmt Datei **oder** Verzeichnis
// — die zwei Bindungen gibt es nur, weil macOS zwei Dialoge kennt. Wer einen
// Pfad tippt, tippt beides gleich.

package main

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// mtgjsonEinfahrt sagt, was eingelesen wurde.
type mtgjsonEinfahrt struct {
	Angelegt   []string `json:"angelegt"`
	SchonOffen bool     `json:"schonOffen"`

	// Befunde ohne `omitempty`: eine leere Liste ist die Aussage „nichts ist
	// aufgefallen" (§3).
	Befunde []string `json:"befunde"`
}

// mtgjsonAusfuhr sagt, was geschrieben wurde.
type mtgjsonAusfuhr struct {
	Pfad    string   `json:"pfad"`
	Karten  int      `json:"karten"`
	Befunde []string `json:"befunde"`
}

// mtgjsonLesen fährt eine MTG-JSON-Quelle ein.
func mtgjsonLesen(ctx context.Context, db *sql.DB, an *schreiber, pfad, spiel string) int {
	eingefahren, fehler := app.MTGJSONLesen(ctx, db, pfad, spiel)

	// **Erst das Ergebnis, dann der Fehler** (#155). Eine AllSets-Datei führt
	// Hunderte Sets; was vor dem Abbruch durchkam, steht im Bestand, und wer
	// nur den Fehler zeigt, lässt Sets zurück, von denen niemand weiß.
	if eingefahren != nil {
		mtgjsonSagen(an, eingefahren)
	}

	if fehler != nil {
		an.klageUeber(fehler)

		return mtgjsonCode(fehler)
	}

	return an.abschluss()
}

// mtgjsonSagen gibt aus, was entstanden ist.
func mtgjsonSagen(an *schreiber, eingefahren *app.MTGJSONEinfahrt) {
	kennungen := make([]string, 0, len(eingefahren.Angelegt))
	for _, kennung := range eingefahren.Angelegt {
		kennungen = append(kennungen, strconv.FormatInt(kennung, 10))
	}

	if an.alsJSON {
		alsJSONSchreiben(an, mtgjsonEinfahrt{
			Angelegt: kennungen, SchonOffen: eingefahren.SchonOffen,
			Befunde: eingefahren.Befunde,
		})

		return
	}

	if eingefahren.SchonOffen {
		an.sag("Diese Quelle liegt schon im Bestand.\n")
	}

	for _, kennung := range kennungen {
		an.sag("%s\n", kennung)
	}

	for _, befund := range eingefahren.Befunde {
		an.sag("  %s\n", befund)
	}
}

// mtgjsonSchreiben schreibt ein Set als MTG JSON.
func mtgjsonSchreiben(
	ctx context.Context, db *sql.DB, an *schreiber, setID int64, spiel, ziel string,
) int {
	ausfuhr, fehler := app.MTGJSONSchreiben(ctx, db, setID, spiel, ziel)
	if fehler != nil {
		an.klageUeber(fehler)

		return mtgjsonCode(fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, mtgjsonAusfuhr{
			Pfad: ausfuhr.Pfad, Karten: ausfuhr.Karten, Befunde: ausfuhr.Befunde,
		})
	}

	an.sag("Geschrieben nach %s — %d Karten.\n", ausfuhr.Pfad, ausfuhr.Karten)

	for _, befund := range ausfuhr.Befunde {
		an.sag("  %s\n", befund)
	}

	return an.abschluss()
}

// mtgjsonCode ordnet einen Fehler einem Rückgabewert zu.
//
// **Ein unbekanntes Spiel ist ein Bedienfehler**, eine unlesbare Datei ein
// fachliches Nein — ein Skript soll das unterscheiden können (§8).
func mtgjsonCode(fehler error) int {
	if errors.Is(fehler, domain.ErrValidation) {
		return CodeBedienung
	}

	return CodeFachlich
}
