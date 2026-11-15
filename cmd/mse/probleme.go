package main

// **`mse probleme <set>`** — PERF02
// (#2352):
// prüft jede Karte eines Sets, wie die Vorschau sie zeichnet, und nennt die
// mit Problemen. **Der Fortschritt geht auf den Fehlerkanal**, je
// angefangenem Zehntel — wie beim Holen einer Fassung; auf dem Ziel steht
// am Ende nur das Ergebnis.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func problemebefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 1 {
		an.klage("probleme braucht genau eine Setkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg, err := app.ProblemwegUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		melde := meldenAuf(an)

		stand, fehler := weg.Pruefen(ctx, setID, func(geprueft, gesamt int) { melde(int64(geprueft), int64(gesamt)) })
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		return problemeZeigen(an, stand)
	})
}

// problemeZeigen schreibt den Stand — mit `--json` als Ganzes.
func problemeZeigen(an *schreiber, stand app.ProblemstandDTO) int {
	if an.alsJSON {
		return alsJSONSchreiben(an, stand)
	}

	an.sag("%d von %d Karten geprüft, %d mit Problemen%s\n", stand.Geprueft, stand.Gesamt, stand.Zahl, vorbehalt(stand.Urteil))

	for _, karte := range stand.Karten {
		for _, befund := range karte.Befunde {
			an.sag("%s\t%s\n", karte.KarteID, befund)
		}
	}

	return an.abschluss()
}

// vorbehalt ist der Schluss des Satzes — **nur ein aktueller Stand nennt
// seine Zahl ohne Vorbehalt** (Zweitblick über Phase 3, Befund 16): sonst
// stand im Text, was nur in `--json` als `veraltet` zu lesen war.
func vorbehalt(urteil string) string {
	switch domain.Problemurteil(urteil) {
	case domain.ProblemeAktuell:
		return "."
	case domain.ProblemeVeraltet:
		return " — veraltet: das Set hat sich seither geändert, prüfe neu."
	case domain.ProblemeUnvollstaendig:
		return " — unvollständig: nicht jede Karte ist geprüft, die Zahl ist eine Untergrenze."
	default:
		return " — ungeprüft."
	}
}
