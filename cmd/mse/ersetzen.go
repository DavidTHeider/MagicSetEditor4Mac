package main

// **Suchen und Ersetzen an der Kommandozeile** — Werkstück 2 von Phase 3.7
// (#1168).
//
// # Die Vorschau ist die Vorgabe
//
//	mse ersetzen 1 "Fliegend" "Rennend"          zeigt, was sich ändern würde
//	mse ersetzen 1 "Fliegend" "Rennend" --tun    tut es
//
// **Herum und nicht andersherum**: eine Ersetzung über dreihundert Karten ist
// nicht zu überblicken, und ein vertipptes Wort schriebe sie alle. Wer sie
// will, sagt es — §16 Regel 6 verlangt dasselbe für fremde Daten, und der
// Grund ist derselbe.

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ersetzbefehl beantwortet `mse ersetzen`.
func ersetzbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 3 {
		an.klage("ersetzen braucht ein Set, einen Suchtext und einen Ersatz.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	frage := domain.Ersetzungsfrage{Text: rest[1], Ersatz: rest[2]}
	tun := false

	for _, fahne := range rest[3:] {
		switch {
		case fahne == "--tun":
			tun = true
		case fahne == "--gross-klein":
			frage.GrossKleinBeachten = true
		case fahne == "--ganzes-wort":
			frage.GanzesWort = true
		case strings.HasPrefix(fahne, "--feld="):
			frage.Feld = strings.TrimPrefix(fahne, "--feld=")
		default:
			an.klage("unbekannte Angabe %q bei ersetzen.\n%s", fahne, hilfe)

			return CodeBedienung
		}
	}

	if befunde := frage.Validate(); len(befunde) > 0 {
		an.klageUeber(befunde[0])

		return CodeFachlich
	}

	return mitErsetzung(ctx, bestandsort, an, func(ctx context.Context, naht *app.Fremdersetzung) int {
		schau, fehler := arbeitAn(ctx, naht, setID, frage, tun)
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		return ersetzungAusgeben(an, schau, tun)
	})
}

// arbeitAn wählt zwischen Vorschau und Ausführung.
func arbeitAn(
	ctx context.Context, naht *app.Fremdersetzung, setID int64,
	frage domain.Ersetzungsfrage, tun bool,
) (domain.Ersetzungsvorschau, error) {
	if tun {
		return naht.Ausfuehren(ctx, setID, frage)
	}

	return naht.Vorschau(ctx, setID, frage)
}

// ersetzungAusgeben schreibt die Übersicht.
func ersetzungAusgeben(an *schreiber, schau domain.Ersetzungsvorschau, tun bool) int {
	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{
			"stellen": ersetzungsstellen(schau.Stellen),
			"karten":  schau.Karten,
			"treffer": schau.Treffer,
			"weitere": schau.Weitere,
			"getan":   tun,
		})
	}

	for _, stelle := range schau.Stellen {
		an.sag("%d\t%s\n", stelle.Nummer+1, stelle.Feld)
		an.sag("  − %s\n", stelle.Vorher)
		an.sag("  + %s\n", stelle.Nachher)
	}

	if schau.Weitere {
		an.sag("… weitere sind nicht aufgeführt.\n")
	}

	// **Die Zeile darunter sagt, ob geschrieben wurde.** Ohne sie sähe eine
	// Vorschau aus wie eine getane Ersetzung — und das ist der teure Irrtum.
	if tun {
		an.sag("Ersetzt — Treffer: %d, Karten: %d\n", schau.Treffer, schau.Karten)
	} else {
		an.sag("Treffer: %d, betroffene Karten: %d — mit --tun ersetzen.\n",
			schau.Treffer, schau.Karten)
	}

	return an.abschluss()
}

// ersetzungsstellen macht aus den Stellen eine Liste — **immer eine**.
func ersetzungsstellen(stellen []domain.Ersetzungsstelle) []map[string]any {
	liste := make([]map[string]any, 0, len(stellen))

	for _, stelle := range stellen {
		liste = append(liste, map[string]any{
			"karteID": stelle.KarteID,
			"nummer":  stelle.Nummer + 1,
			"feld":    stelle.Feld,
			"vorher":  stelle.Vorher,
			"nachher": stelle.Nachher,
			"treffer": stelle.Treffer,
		})
	}

	return liste
}

// mitErsetzung öffnet den Bestand und baut die Naht.
func mitErsetzung(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Fremdersetzung) int,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.Ersetzungsflaechennaht(db, domain.UrheberKommandozeile)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return arbeit(ctx, naht)
	})
}
