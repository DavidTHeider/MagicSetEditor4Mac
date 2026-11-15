package main

// **Die Kartensuche an der Kommandozeile** — Werkstück 1 von Phase 3.6
// (#1167).

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// scryfallsuchbefehl sucht Karten bei Scryfall.
//
// # Warum er den Bestand doch braucht
//
// **Hier stand „er kommt ohne Bestand aus", weil er nichts aus ihm liest.**
// Das stimmt für die Suche und nicht mehr für die Frage davor: seit Phase 4.8
// steht in den Einstellungen, **ob Scryfall überhaupt gefragt werden darf**
// ([0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md)), und die
// Einstellungen liegen im Bestand.
//
// **Ohne ihn ginge ein abgeschalteter Dienst an dieser Fläche weiter
// hinaus** — und die Zusage „aus heißt null Anfragen" gälte nur im Fenster.
//
// # Und warum die Frage aus **allen** Argumenten entsteht
//
// Scryfalls Suchsyntax hat Leerzeichen: `t:creature c:r cmc<=3`. Wer nur das
// erste Argument nähme, suchte nach `t:creature` und schwiege über den Rest.
func scryfallsuchbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	frage := strings.TrimSpace(strings.Join(rest, " "))
	if frage == "" {
		an.klage("scryfall-suche braucht eine Frage.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return scryfallSuchen(ctx, db, an, frage)
	})
}

// scryfallSuchen führt die Suche über einem offenen Bestand.
func scryfallSuchen(
	ctx context.Context, db *sql.DB, an *schreiber, frage string,
) int {
	naht, err := app.Fremdsuchnaht(ctx, db)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	ergebnis, suchfehler := naht.Suche(ctx, frage)
	if suchfehler != nil {
		an.klageUeber(suchfehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, ergebnis)
	}

	for _, treffer := range ergebnis.Treffer {
		an.sag("%s\t%s\t%s %s\t%s\t%s\n", treffer.Kennung, treffer.Name,
			treffer.Setcode, treffer.Nummer, treffer.Seltenheit,
			treffer.Manakosten)
	}

	// **Und der Stand sagt, wie alt er ist** — Befund 3 des Nachlaufs über
	// Phase 3.6.
	//
	// §16, Regel 5: „Der letzte gute Stand bleibt, und er sagt, wie alt er
	// ist." Der JSON-Weg trug das Alter die ganze Zeit; der Textweg warf es
	// weg, und eine Antwort aus dem Zwischenspeicher sah aus wie eine frische.
	// **Nur wenn es eines gibt**: „0 Stunden alt" wäre eine Zeile, die nichts
	// sagt und jede Ausgabe länger macht.
	if ergebnis.AlterStunden > 0 {
		an.sag("Stand: %d Stunden alt\n", ergebnis.AlterStunden)
	}

	// **»Es gibt mehr« steht dabei**, denn eine Liste, die schweigend endet,
	// sieht aus wie eine vollständige.
	if ergebnis.Weitere {
		an.sag("… Scryfall hat mehr Treffer; die Frage enger fassen\n")
	}

	for _, befund := range ergebnis.Befunde {
		an.sag("Befund: %s\n", befund)
	}

	return an.abschluss()
}
