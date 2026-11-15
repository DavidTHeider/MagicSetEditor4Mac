package main

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// scryfallregelnbefehl zeigt die Regelhinweise zu einer Karte.
//
// # Warum er ohne Bestand auskommt
//
// **Weil er nichts aus ihm liest.** Eine Regelauskunft hängt an einem
// Kartennamen, nicht an einer Kartenkennung — und ihn zu öffnen, nur um ihn
// nicht zu benutzen, machte den Befehl von einem Zustand abhängig, den er
// nicht braucht.
func scryfallregelnbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 || strings.TrimSpace(rest[0]) == "" {
		an.klage("scryfall-regeln braucht genau einen Kartennamen.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return regelnZeigen(ctx, db, an, rest[0])
	})
}

// regelnZeigen führt den Befehl über einem offenen Bestand.
func regelnZeigen(
	ctx context.Context, db *sql.DB, an *schreiber, name string,
) int {
	naht, err := app.Fremdauskunftsnaht(db)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	geholt, holfehler := naht.Regelhinweise(ctx, strings.TrimSpace(name))
	if holfehler != nil {
		an.klageUeber(holfehler)

		return CodeFachlich
	}

	return zeigeRegeln(an, geholt)
}

// zeigeRegeln schreibt die Hinweise heraus.
func zeigeRegeln(an *schreiber, geholt app.Regelauskunft) int {
	an.sag("%s — %d Hinweise\n%s", geholt.Kartenname, len(geholt.Hinweise), geholt.Adresse)

	for _, befund := range geholt.Befunde {
		an.sag("Befund: %s", befund)
	}

	for _, einer := range geholt.Hinweise {
		// **Die Quelle steht vorn**, weil sie entscheidet, wie ein Hinweis zu
		// lesen ist: eine Anmerkung des Scryfall-Teams ist keine Regel.
		herkunft := "Scryfall"
		if einer.Offiziell {
			herkunft = "offiziell"
		}

		an.sag("\n[%s, %s]\n%s", herkunft, einer.Datum, einer.Text)
	}

	if geholt.AlterStunden > 0 {
		an.sag("\n(Stand: %d Stunden alt)", geholt.AlterStunden)
	}

	return CodeGut
}

// scryfallmanabefehl prüft einen Manakosten gegen die Symbologie.
func scryfallmanabefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("scryfall-mana braucht genau einen Manakosten, etwa »{2}{R}«.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return manaZeigen(ctx, db, an, rest[0])
	})
}

// manaZeigen führt den Befehl über einem offenen Bestand.
func manaZeigen(ctx context.Context, db *sql.DB, an *schreiber, kosten string) int {
	naht, err := app.Fremdauskunftsnaht(db)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	unbekannt, prueffehler := naht.Manakosten(ctx, kosten)
	if prueffehler != nil {
		an.klageUeber(prueffehler)

		return CodeFachlich
	}

	if len(unbekannt) == 0 {
		an.sag("%s ist in Ordnung.", kosten)

		return CodeGut
	}

	// **Ein unbekanntes Symbol ist kein Fehler des Programms**, sondern ein
	// Befund über die Eingabe — deshalb `CodeGut` und keine Klage: der Befehl
	// hat getan, wonach er gefragt wurde.
	an.sag("%s: unbekannt sind %s", kosten, strings.Join(unbekannt, ", "))

	return CodeGut
}

// scryfallsammelbefehl zeigt oder holt die Sammeldatei.
//
// # Warum ein Befehl mit zwei Wörtern
//
// `zeigen` liest nur, `holen` lädt bis zu 25 MB. Das gehört auseinander: ein
// Befehl, der beides täte, lüde beim bloßen Nachsehen.
func scryfallsammelbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 || (rest[0] != "zeigen" && rest[0] != "holen") {
		an.klage("scryfall-sammeldaten braucht »zeigen« oder »holen«.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return sammeldaten(ctx, db, an, rest[0])
	})
}

// sammeldaten führt den Befehl über einem offenen Bestand.
//
// **Der Bestand ist neu** (Phase 4.8): dort steht, ob Scryfall gefragt werden
// darf. Ohne ihn ginge ein abgeschalteter Dienst an dieser Fläche weiter
// hinaus.
func sammeldaten(
	ctx context.Context, db *sql.DB, an *schreiber, was string,
) int {
	naht, err := app.Sammeldatennaht(app.NeueEinstellungswahl(storage.NeueEinstellungen(db)).MitRiegel(app.AutorisierungAus(db)))
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	stand, standfehler := sammelstand(ctx, naht, was)
	if standfehler != nil {
		an.klageUeber(standfehler)

		return CodeFachlich
	}

	return zeigeSammelstand(an, stand)
}

// sammelstand führt den gewählten Schritt aus.
func sammelstand(
	ctx context.Context, naht *app.Sammeldaten, was string,
) (app.Sammeldatenstand, error) {
	if was == "holen" {
		return naht.Holen(ctx)
	}

	return naht.Stand(ctx)
}

// zeigeSammelstand schreibt den Stand heraus.
func zeigeSammelstand(an *schreiber, stand app.Sammeldatenstand) int {
	if !stand.Vorhanden {
		an.sag("%s: nicht vorhanden (%s)", stand.Art, stand.Pfad)
	} else {
		an.sag("%s: %.1f MB, geholt %s (%s)", stand.Art, float64(stand.Bytes)/1e6,
			stand.Geholt, stand.Pfad)
	}

	if stand.Angeboten > 0 {
		an.sag("Scryfall bietet %.1f MB an, Stand %s",
			float64(stand.Angeboten)/1e6, stand.Stand)
	}

	return CodeGut
}
