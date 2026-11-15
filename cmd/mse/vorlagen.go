// **Die Exportvorlagen an der Kommandozeile** —
// #1291.
//
// # Warum es ohne Fenster geht
//
// Der Speicherdialog steckt allein in `Bindungen.VorlageAusfuehren`.
// `App.VorlageAusfuehren` sagt es selbst: *„Der Pfad kommt vom Aufrufer, nicht
// aus einem Dialog."*
//
// **Und es ist die Fähigkeit, für die MSE2 seine Vorlagen mitbringt**: das
// `M15-Magic-Pack` führt zwölf `.mse-export-template` — Cockatrice, MWS,
// LackeyCCG, Planesculptors.

package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// vorlagenDTO ist eine Exportvorlage, wie die Kommandozeile sie ausgibt.
type vorlagenDTO struct {
	Paket    string       `json:"paket"`
	Name     string       `json:"name"`
	Dateityp string       `json:"dateityp"`
	Optionen []optionsDTO `json:"optionen"`
}

// optionsDTO ist ein einstellbares Feld einer Vorlage.
type optionsDTO struct {
	Name string `json:"name"`
	Art  string `json:"art"`
	Wert string `json:"wert"`
}

// vorlagenAntwort trägt die Liste samt Befunden.
type vorlagenAntwort struct {
	Vorlagen []vorlagenDTO `json:"vorlagen"`
	Befunde  []string      `json:"befunde"`
}

// vorlagenausfuhrAntwort sagt, was geschrieben wurde.
type vorlagenausfuhrAntwort struct {
	Pfad    string   `json:"pfad"`
	Befunde []string `json:"befunde"`
}

// vorlagenZeigen nennt die Exportvorlagen eines Sets.
func vorlagenZeigen(ctx context.Context, db *sql.DB, an *schreiber, setID int64) int {
	weg, err := app.VorlagenwegUeber(db, domain.UrheberKommandozeile)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	vorlagen, befunde, fehler := weg.Vorlagen(ctx, setID)
	if fehler != nil {
		an.klageUeber(fehler)

		return vorlagenCode(fehler)
	}

	antwort := vorlagenAntwort{
		Vorlagen: make([]vorlagenDTO, 0, len(vorlagen)),
		Befunde:  ohneNilTexte(befunde),
	}

	for _, vorlage := range vorlagen {
		optionen := make([]optionsDTO, 0, len(vorlage.Optionen))
		for _, option := range vorlage.Optionen {
			optionen = append(optionen, optionsDTO{
				Name: option.Name, Art: option.Art, Wert: option.Wert,
			})
		}

		antwort.Vorlagen = append(antwort.Vorlagen, vorlagenDTO{
			Paket: vorlage.Paket, Name: vorlage.Name,
			Dateityp: vorlage.Dateityp, Optionen: optionen,
		})
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	for _, vorlage := range antwort.Vorlagen {
		an.sag("%s\t%s\t%s\n", vorlage.Paket, vorlage.Name, vorlage.Dateityp)
	}

	for _, befund := range antwort.Befunde {
		an.sag("  %s\n", befund)
	}

	return an.abschluss()
}

// vorlageAusfuehren führt eine Vorlage aus und schreibt das Ergebnis.
func vorlageAusfuehren(
	ctx context.Context, db *sql.DB, an *schreiber, setID int64, paket, ziel string,
	wahl map[string]string,
) int {
	weg, err := app.VorlagenwegUeber(db, domain.UrheberKommandozeile)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	lauf, fehler := weg.Ausfuehren(ctx, setID, paket, ziel, wahl)
	if fehler != nil {
		an.klageUeber(fehler)

		// **Die Befunde reisen mit** (#209): „der Name `write_image_file` ist
		// hier nicht bekannt" ist eine Aussage über **ihre** Vorlage.
		for _, befund := range lauf.Befunde {
			an.klage("  %s", befund)
		}

		return vorlagenCode(fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, vorlagenausfuhrAntwort{
			Pfad: lauf.Pfad, Befunde: ohneNilTexte(lauf.Befunde),
		})
	}

	an.sag("Geschrieben nach %s.\n", lauf.Pfad)

	for _, befund := range lauf.Befunde {
		an.sag("  %s\n", befund)
	}

	return an.abschluss()
}

// vorlagenoptionSetzen behält den Wert einer Vorlagenoption.
func vorlagenoptionSetzen(
	ctx context.Context, db *sql.DB, an *schreiber, vorlage, feld, wert string,
) int {
	weg, err := app.VorlagenwegUeber(db, domain.UrheberKommandozeile)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	if fehler := weg.OptionSetzen(ctx, vorlage, feld, wert); fehler != nil {
		an.klageUeber(fehler)

		return vorlagenCode(fehler)
	}

	an.sag("%s.%s = %s\n", vorlage, feld, wert)

	return an.abschluss()
}

// wahlAus liest die Optionswerte aus `feld=wert`-Paaren.
//
// **Ein Gleichheitszeichen trennt, das erste**: ein Wert darf selbst eines
// enthalten, ein Feldname nicht.
func wahlAus(an *schreiber, paare []string) (map[string]string, bool) {
	wahl := make(map[string]string, len(paare))

	for _, paar := range paare {
		feld, wert, gefunden := strings.Cut(paar, "=")
		if !gefunden || feld == "" {
			an.klage("%q ist kein feld=wert.\n%s", paar, hilfe)

			return nil, false
		}

		wahl[feld] = wert
	}

	return wahl, true
}

// vorlagenCode ordnet einen Fehler einem Rückgabewert zu.
func vorlagenCode(fehler error) int {
	if errors.Is(fehler, domain.ErrValidation) {
		return CodeBedienung
	}

	return CodeFachlich
}
