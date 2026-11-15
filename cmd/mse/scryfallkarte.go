// **Eine Karte aus einem Namen** —
// #1032.
//
// Eine eigene Datei, weil es ein eigener Vorgang ist: `scryfall.go` gleicht
// eine **vorhandene** Karte ab — hier entsteht eine.
package main

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// scryfallkartebefehl legt eine Karte aus einem Scryfall-Namen an.
//
// **Das Bild ist wahlweise dabei**: `--bild <feld> <art>`. Ohne die Fahne wird
// keines geholt — ein Bild ist ein eigener Abruf mit einer eigenen Bedingung.
func scryfallkartebefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 2 {
		an.klage("scryfall-karte braucht die Setkennung und den Kartennamen.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	bildfeld, art, gut := bildwahl(an, rest[2:])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return scryfallKarteAnlegen(ctx, db, an, setID, rest[1], bildfeld, art)
	})
}

// bildwahl liest `--bild <feld> <art>` aus den Restargumenten.
//
// **Leer heißt: kein Bild.** Wer eines will, nennt beides — welches Feld es
// aufnimmt, weiß nur die Vorlage, und es zu raten hieße, bei jeder fremden
// Vorlage falsch zu raten.
func bildwahl(an *schreiber, rest []string) (feld, art string, gut bool) {
	if len(rest) == 0 {
		return "", "", true
	}

	if rest[0] != "--bild" || len(rest) != 3 {
		an.klage("nach dem Kartennamen steht höchstens `--bild <feld> <art>`.\n%s", hilfe)

		return "", "", false
	}

	if strings.TrimSpace(rest[1]) == "" || strings.TrimSpace(rest[2]) == "" {
		an.klage("`--bild` nennt Feld und Art.\n%s", hilfe)

		return "", "", false
	}

	return rest[1], rest[2], true
}

// scryfallKarteAnlegen legt an und meldet, was hineinkam.
func scryfallKarteAnlegen(
	ctx context.Context, db *sql.DB, an *schreiber,
	setID int64, name, bildfeld, art string,
) int {
	naht, err := app.Scryfallabgleichnaht(db, domain.UrheberKommandozeile)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	ergebnis, anlegefehler := naht.Anlegen(ctx, setID, name)

	// **Die angelegte Karte wird auch im Fehlerfall genannt** (#155): sie ist
	// da, und wer sie verschweigt, lässt sie im Bestand zurück.
	if anlegefehler != nil {
		if ergebnis.KarteID != 0 {
			an.sag("Angelegt wurde die Karte %d — das Füllen ist gescheitert.\n",
				ergebnis.KarteID)
		}

		an.klageUeber(anlegefehler)

		return CodeFachlich
	}

	if art != "" {
		if code := bildEinsetzen(ctx, db, an, setID, ergebnis.KarteID,
			bildfeld, name, art); code != CodeGut {
			return code
		}
	}

	return zeigeAnlage(an, ergebnis, bildfeld, art)
}

// bildEinsetzen hängt das Bild an die eben angelegte Karte.
//
// **Über denselben Weg wie `scryfall-bild`** — eine zweite Fassung liefe bei
// der ersten Änderung an der Nennungspflicht auseinander.
func bildEinsetzen(
	ctx context.Context, db *sql.DB, an *schreiber,
	setID, karteID int64, feld, name, art string,
) int {
	bildnaht, err := app.Fremdbildnaht(db, domain.UrheberKommandozeile)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	if _, fehler := bildnaht.Einsetzen(ctx, setID, karteID, feld, name, art); fehler != nil {
		an.sag("Angelegt wurde die Karte %d — das Bild ist gescheitert.\n", karteID)
		an.klageUeber(fehler)

		return CodeFachlich
	}

	return CodeGut
}

// zeigeAnlage schreibt das Ergebnis — lesbar oder als JSON.
func zeigeAnlage(an *schreiber, ergebnis app.AnlageDTO, bildfeld, art string) int {
	felder := ergebnis.Felder
	if art != "" {
		felder = append(felder, bildfeld)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{
			"karteID":  ergebnis.KarteID,
			"felder":   felder,
			"befunde":  ergebnis.Befunde,
			"herkunft": ergebnis.Herkunft,
		})
	}

	an.sag("Karte %d angelegt aus %s.\n", ergebnis.KarteID, ergebnis.Herkunft)
	an.sag("  Gefüllt: %s\n", strings.Join(felder, ", "))

	// **Die Befunde stehen dabei** (§8): eine Karte mit einer unbekannten
	// Seltenheit ist brauchbar — die Auffälligkeit gehört daneben.
	for _, befund := range ergebnis.Befunde {
		an.sag("  Befund: %s\n", befund)
	}

	return an.abschluss()
}
