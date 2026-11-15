package main

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// aenderungDTO ist eine einzelne Änderung.
type aenderungDTO struct {
	Feld string `json:"feld"`
	Alt  string `json:"alt"`
	Neu  string `json:"neu"`
}

// scryfallbefehl zeigt einen Vorschlag — oder übernimmt genannte Felder.
//
// # Warum ein Befehl und nicht zwei
//
// Weil die Wache dieselbe ist: übernommen wird nur, was im Vorschlag steht.
// Ein zweiter Befehl müsste den Vorschlag noch einmal holen, und dann wäre
// er derselbe Befehl mit einem Argument mehr — was er ist.
//
// **Ohne `--uebernehmen` schreibt er nichts** (§16, Regel 6).
func scryfallbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 3 {
		an.klage("scryfall braucht Setkennung, Kartenkennung und den Kartennamen.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	karteID, gut := kennung(an, "die Kartenkennung", rest[1])
	if !gut {
		return CodeBedienung
	}

	felder, gut := uebernahmefelder(an, rest[3:])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return scryfallAbgleichen(ctx, db, an, setID, karteID, rest[2], felder)
	})
}

// uebernahmefelder liest `--uebernehmen feld,feld` aus den Restargumenten.
//
// **Nil heißt: nur zeigen.** Eine leere Liste nach `--uebernehmen` ist ein
// Bedienfehler — wer nichts übernehmen will, lässt die Fahne weg.
func uebernahmefelder(an *schreiber, rest []string) ([]string, bool) {
	if len(rest) == 0 {
		return nil, true
	}

	if rest[0] != "--uebernehmen" || len(rest) != 2 {
		an.klage("nach dem Kartennamen steht höchstens `--uebernehmen feld,feld`.\n%s", hilfe)

		return nil, false
	}

	felder := strings.Split(rest[1], ",")
	for stelle, feld := range felder {
		felder[stelle] = strings.TrimSpace(feld)

		if felder[stelle] == "" {
			an.klage("`--uebernehmen` nennt einen leeren Feldnamen.\n%s", hilfe)

			return nil, false
		}
	}

	return felder, true
}

// scryfallAbgleichen holt den Vorschlag und übernimmt, wenn Felder genannt sind.
func scryfallAbgleichen(
	ctx context.Context, db *sql.DB, an *schreiber,
	setID, karteID int64, name string, felder []string,
) int {
	// **Dieselbe Naht wie Fenster, MCP-Server und Dienst** — ein übernommener
	// Wert geht denselben Weg wie ein getippter (#710).
	naht, err := app.Scryfallabgleichnaht(db, domain.UrheberKommandozeile)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	if len(felder) == 0 {
		vorschlag, holfehler := naht.Vorschlag(ctx, setID, karteID, name)
		if holfehler != nil {
			an.klageUeber(holfehler)

			return CodeFachlich
		}

		return zeigeVorschlag(an, vorschlag)
	}

	uebernommen, fehler := naht.Uebernehmen(ctx, setID, karteID, name, felder)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{"uebernommen": alsAenderungen(uebernommen)})
	}

	an.sag("Übernommen: %s.\n", strings.Join(felder, ", "))

	// **Der vorige Wert steht dabei** — sonst weiß ein Skript nicht, wie es
	// zurück soll.
	for _, eine := range uebernommen {
		an.sag("  %s war %q\n", eine.Feld, eine.Alt)
	}

	return an.abschluss()
}

// zeigeVorschlag schreibt einen Vorschlag — lesbar oder als JSON.
func zeigeVorschlag(an *schreiber, vorschlag app.ScryfallvorschlagDTO) int {
	if an.alsJSON {
		return alsJSONSchreiben(an, vorschlag)
	}

	if len(vorschlag.Aenderungen) == 0 {
		an.sag("Nichts zu ändern — die Karte stimmt mit Scryfall überein.\n")

		return an.abschluss()
	}

	an.sag("Aus %s", vorschlag.Herkunft)

	if stunden := vorschlag.AlterStunden; stunden > 0 {
		an.sag(" (aus dem Zwischenspeicher, %d Stunden alt)", stunden)
	}

	an.sag(":\n")

	for _, eine := range vorschlag.Aenderungen {
		an.sag("  %s\t%q → %q\n", eine.Feld, eine.Alt, eine.Neu)
	}

	for _, befund := range vorschlag.Befunde {
		an.sag("  Befund: %s\n", befund)
	}

	an.sag("\nÜbernehmen mit --uebernehmen %s\n", strings.Join(feldnamen(vorschlag), ","))

	return an.abschluss()
}

// feldnamen gibt die Feldnamen eines Vorschlags.
func feldnamen(vorschlag app.ScryfallvorschlagDTO) []string {
	namen := make([]string, 0, len(vorschlag.Aenderungen))
	for _, eine := range vorschlag.Aenderungen {
		namen = append(namen, eine.Feld)
	}

	return namen
}

// alsAenderungen übersetzt die übernommenen Felder für die JSON-Ausgabe.
func alsAenderungen(aenderungen []app.FeldvorschlagDTO) []aenderungDTO {
	gestalt := make([]aenderungDTO, 0, len(aenderungen))
	for _, eine := range aenderungen {
		gestalt = append(gestalt, aenderungDTO{Feld: eine.Feld, Alt: eine.Alt, Neu: eine.Neu})
	}

	return gestalt
}
