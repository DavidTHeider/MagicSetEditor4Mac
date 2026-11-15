package main

// **Karten kopieren und einfügen an der Kommandozeile** — Werkstück 1 von
// Phase 3.7
// (#1168).
//
// # Ohne Fenster ist die Zwischenablage die Standardausgabe
//
//	mse ablage-kopieren 1 11 12 > karten.txt
//	mse ablage-einfuegen 2 < karten.txt
//
// Damit trägt der Weg auch dorthin, wo das Fenster nicht hinkommt: in eine
// Datei, durch eine Pipe, über eine Maschinengrenze. §16 verlangt dieselbe
// Fähigkeit an jeder Fläche — und hier ist sie sogar mehr wert.

import (
	"context"
	"database/sql"
	"io"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ablagekopierbefehl gibt die genannten Karten als Text.
func ablagekopierbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	return ablagehinaus(ctx, bestandsort, an, rest,
		func(ctx context.Context, naht *app.Fremdablage, setID int64, karten []int64) (string, error) {
			return naht.Kopieren(ctx, setID, karten)
		})
}

// ablageschnittbefehl gibt sie als Text und entfernt sie aus dem Set.
func ablageschnittbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	return ablagehinaus(ctx, bestandsort, an, rest,
		func(ctx context.Context, naht *app.Fremdablage, setID int64, karten []int64) (string, error) {
			return naht.Ausschneiden(ctx, setID, karten)
		})
}

// ablagehinaus ist der gemeinsame Rumpf von Kopieren und Ausschneiden.
func ablagehinaus(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
	hole func(context.Context, *app.Fremdablage, int64, []int64) (string, error),
) int {
	if len(rest) < 2 {
		an.klage("ablage-kopieren braucht ein Set und mindestens eine Karte.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	karten := make([]int64, 0, len(rest)-1)

	for _, roh := range rest[1:] {
		karteID, guteKarte := kennung(an, "Kartenkennung", roh)
		if !guteKarte {
			return CodeBedienung
		}

		karten = append(karten, karteID)
	}

	return mitAblage(ctx, bestandsort, an, func(ctx context.Context, naht *app.Fremdablage) int {
		inhalt, fehler := hole(ctx, naht, setID, karten)
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{"inhalt": inhalt})
		}

		// **Wörtlich und ohne Zeile drumherum.** Was hier herauskommt, geht
		// unverändert wieder hinein — eine Kopfzeile machte daraus etwas, das
		// `ablage-einfuegen` nicht mehr liest.
		an.sag("%s", inhalt)

		return an.abschluss()
	})
}

// ablageeinfuegebefehl legt die Karten aus der Standardeingabe im Set an.
func ablageeinfuegebefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) == 0 {
		an.klage("ablage-einfuegen braucht eine Setkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	// **Ohne Eingabe ist die Ablage leer.** `nil` heißt „nichts" — kein
	// Fehler, sondern derselbe Fall wie eine leere Zwischenablage.
	var roh []byte

	if an.eingabe != nil {
		gelesen, lesefehler := io.ReadAll(an.eingabe)
		if lesefehler != nil {
			an.klageUeber(lesefehler)

			return CodeFachlich
		}

		roh = gelesen
	}

	if strings.TrimSpace(string(roh)) == "" {
		an.klage("ablage-einfuegen liest den Inhalt von der Standardeingabe.\n%s", hilfe)

		return CodeBedienung
	}

	return mitAblage(ctx, bestandsort, an, func(ctx context.Context, naht *app.Fremdablage) int {
		karten, befunde, fehler := naht.Einfuegen(ctx, setID, string(roh))
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{
				"karten": karten, "befunde": befunde,
			})
		}

		for _, karteID := range karten {
			an.sag("%s\n", karteID)
		}

		for _, befund := range befunde {
			an.sag("Befund: %s\n", befund)
		}

		return an.abschluss()
	})
}

// mitAblage öffnet den Bestand und baut die Naht.
func mitAblage(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Fremdablage) int,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.Ablageflaechennaht(db, domain.UrheberKommandozeile)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return arbeit(ctx, naht)
	})
}
