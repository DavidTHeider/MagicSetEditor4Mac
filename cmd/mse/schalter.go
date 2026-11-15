package main

// **Bereiche und Funktionen abschalten, an der Kommandozeile** — OPS-10
// (#1908).
//
// # Das Tor vor jedem Befehl
//
// **Vor der Verteilung und mit eigenem Griff in den Bestand**: ein Befehl
// öffnet ihn erst in seiner eigenen Arbeit, und jede Verteilung einzeln zu
// fragen wären dreißig Stellen statt einer. **Nur ein Befehl aus der Tafel
// öffnet ihn**; ein Befehl des Kerns kostet nichts.
//
// **Geht der Bestand nicht auf, lässt das Tor durch** — ein Schalter sichert
// nichts, und der Befehl sagt dann selbst, was mit dem Bestand ist.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// schaltertor weist einen Befehl ab, dessen Bereich abgeschaltet ist. Der
// zweite Wert sagt, ob abgewiesen wurde.
func schaltertor(ctx context.Context, befehl string, rest []string, ort string, an *schreiber) (int, bool) {
	weg, gefuehrt := app.Befehlsweg(befehl, rest)
	if !gefuehrt {
		return CodeGut, false
	}

	db, fehler := storage.Oeffnen(ctx, ort)
	if fehler != nil {
		// **Durchlassen, und der Befehl nennt den Weg zurück**: er öffnet
		// den Bestand gleich selbst über `mitBestand`, und dort steht
		// `app.MitWegZurueck` — zwei Meldungen für einen Fehler hälfen
		// niemandem.
		return CodeGut, false
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur gelesen

	imProfil, fehler := profilkontext(ctx, db)
	if fehler != nil {
		return CodeGut, false
	}

	if abgeschaltet := app.SchalterwacheUeber(db).Pruefe(imProfil, domain.FlaecheCLI, weg); abgeschaltet != nil {
		an.klageUeber(abgeschaltet)

		return CodeFachlich, true
	}

	return CodeGut, false
}

// schalterbefehl nennt die Schalter — **Kern**, damit ein abgeschalteter
// Bereich wieder zu finden ist. Geschaltet wird mit
// `mse einstellungen setzen schalter.<kennung> an|aus`.
func schalterbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 0 {
		an.klage("schalter nimmt keine Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.SchalterwacheUeber(db).Stand(ctx)
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		for _, einer := range antwort.Schalter {
			stand := domain.SchalterAus
			if einer.An {
				stand = domain.SchalterAn
			}

			an.sag("%s\t%s\t%s\t%s\n", einer.Kennung, einer.Art, einer.Reife, stand)
		}

		return an.abschluss()
	})
}
