package main

// **Der Tauschabgleich an der Kommandozeile** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// # Warum der fremde Stand aus einer Datei kommt
//
// **Weil er von drüben kommt und nicht von hier.** Was ein Gegenüber zeigt,
// gibt es mit `kontakt-sicht --json` aus; was hier hereingeht, ist genau
// diese Datei. Damit ist der Abgleich heute schon zu fahren — über jeden Weg,
// auf dem zwei Menschen eine Datei austauschen —, und mit `T5` tritt eine
// Leitung an die Stelle des Umwegs.
//
// **Und deshalb steht kein Vorschlag darin, den niemand geschickt hat.** Eine
// erfundene Gegenseite erzeugte Handelspositionen, die es nicht gibt.

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// tauschabgleichbefehl fährt `mse tausch-abgleich`.
func tauschabgleichbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 3 {
		an.klage("tausch-abgleich nimmt drei Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	stand, fehler := fremdstandLesen(rest[1])
	if fehler != nil {
		an.klage("%v", fehler)

		return CodeBedienung
	}

	behalten, zahlfehler := strconv.Atoi(rest[2])
	if zahlfehler != nil {
		an.klage("%q ist keine Zahl.\n%s", rest[2], hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.PeerflaecheUeber(db).
			Tauschabgleich(ctx, rest[0], stand, behalten, 0)
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		schreibeAbgleich(an, antwort)

		return an.abschluss()
	})
}

// fremdstandLesen liest, was ein Gegenüber freigegeben hat.
//
// **Zwei Listen in einer Datei.** `kontakt-sicht --json` gibt eine davon; wer
// beide hat, legt sie unter `angebot` und `wuensche` zusammen. Eine Datei, die
// keines von beidem nennt, ist ein Bedienfehler und kein leerer Abgleich.
func fremdstandLesen(pfad string) (app.FremdstandDTO, error) {
	roh, fehler := os.ReadFile(pfad) // #nosec G304 -- der Pfad ist das Argument
	if fehler != nil {
		return app.FremdstandDTO{}, fmt.Errorf("den fremden Stand lesen: %w", fehler)
	}

	var stand app.FremdstandDTO
	if fehler := json.Unmarshal(roh, &stand); fehler != nil {
		return app.FremdstandDTO{}, fmt.Errorf("den fremden Stand deuten: %w", fehler)
	}

	return stand, nil
}

// schreibeAbgleich gibt den Abgleich als Text.
func schreibeAbgleich(an *schreiber, antwort app.TauschabgleichAntwort) {
	for _, zeile := range antwort.Gebe {
		an.sag("gebe\t%d\t%s\t%s\t%s\t%d\n", zeile.Menge, zeile.Angebot,
			zeile.Wunsch, zeile.Rang, zeile.Fehlend)
	}

	for _, zeile := range antwort.Bekomme {
		an.sag("bekomme\t%d\t%s\t%s\t%s\t%d\n", zeile.Menge, zeile.Angebot,
			zeile.Wunsch, zeile.Rang, zeile.Fehlend)
	}

	if antwort.HatVorschlag {
		an.sag("vorschlag\t%d\t%d\n",
			antwort.Vorschlag.Gebe.Stueck, antwort.Vorschlag.Bekomme.Stueck)
	}

	for _, befund := range antwort.Befunde {
		an.sag("befund\t%s\n", befund)
	}
}
