// **Die Sammlung ausgeben und zurücklesen** — Werkstück 4 von Phase 4.5
// (#1292).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort.
//
//	mse sammlung-ausgeben csv > sammlung.csv
//	mse sammlung-einlesen < sammlung.csv
//	mse sammlung-einlesen sammlung.csv < sammlung.csv
//
// **Zwei Formen**: `liste` ist einfügbar und verlustbehaftet, `csv` trägt
// jede Spalte. Wer sichert, nimmt `csv`.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// sammlungAusgebenbefehl schreibt die Sammlung auf die Standardausgabe.
func sammlungAusgebenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	form := "csv"
	if len(rest) > 0 {
		form = rest[0]
	}

	return mitAusfuhr(ctx, bestandsort, an, func(ctx context.Context, naht *app.Sammlungsausfuhrnaht) int {
		text, err := naht.Ausgeben(ctx, form)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{"form": form, "text": text})
		}

		// **Ohne Zeilenumbruch am Ende**: der Text bringt seinen eigenen mit,
		// und ein zweiter erzeugte beim Zurücklesen eine leere Zeile.
		an.sag("%s", text)

		return an.abschluss()
	})
}

// sammlungEinlesenbefehl liest ein CSV von der Standardeingabe.
//
// **Ein freiwilliges Wort nennt die Quelle** — Werkstück 12 (#1351):
// `mse sammlung-einlesen sammlung.csv < sammlung.csv` schreibt an jede **neu**
// entstehende Zeile, woher sie kam. Ohne das Wort steht dort nichts, und das
// ist der Fall von vorher — **erweiternd** im Sinne von 0033.
//
// **Gelesen wird trotzdem von der Standardeingabe**, nicht aus der Datei: das
// Wort ist eine Bezeichnung und kein Pfad, den dieser Befehl öffnet. Wer eine
// Datei nennt, die er nicht hineinleitet, bekommt seine Bezeichnung an den
// Zeilen aus der Eingabe — und das ist seine Angabe, nicht die des Bestands.
func sammlungEinlesenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	rest, zusammenfuehrung := zusammenfuehrungAus(rest)

	if len(rest) > 1 {
		an.klage("sammlung-einlesen nimmt höchstens eine Bezeichnung.\n%s", hilfe)

		return CodeBedienung
	}

	var quelle string
	if len(rest) == 1 {
		quelle = rest[0]
	}

	text, fehler := io.ReadAll(an.eingabe)
	if fehler != nil {
		an.klageUeber(fmt.Errorf("die Sammlung lesen: %w", fehler))

		return CodeBedienung
	}

	return mitAusfuhr(ctx, bestandsort, an, func(ctx context.Context, naht *app.Sammlungsausfuhrnaht) int {
		bericht, err := naht.CSVEinlesen(ctx, string(text), quelle, zusammenfuehrung)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, bericht)
		}

		an.sag("%d angelegt, %d erhöht, %d Stück%s\n",
			len(bericht.Angelegt), bericht.Erhoeht, bericht.Stueck, ausgelassenText(bericht))

		for _, befund := range bericht.Befunde {
			an.sag("!\t%s\n", befund)
		}

		return an.abschluss()
	})
}

// mitAusfuhr baut die Naht über dem Bestand und führt die Arbeit aus.
//
// **Ohne Fehlerweg**: `AusfuhrUeber` braucht kein Zuhause und kann deshalb
// nicht scheitern. Das ist der Unterschied zu `mitEinfuhr` — und er ist der
// Grund, warum es zwei Nähte gibt (§16, Regel 5).
func mitAusfuhr(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Sammlungsausfuhrnaht) int,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return arbeit(ctx, app.AusfuhrUeber(db))
	})
}
