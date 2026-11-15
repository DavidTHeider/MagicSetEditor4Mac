// **Die Autoersetzungen verwalten, an der Kommandozeile** — Werkstück E von
// Phase 4.11
// (#1341).

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// autoersetzungbefehl zeigt die Regeln — oder ändert eine.
//
// **Ein Befehl und nicht fünf.** Ohne Handlung steht die Liste da; dasselbe
// Muster wie bei `stile` und `vorlagen`.
func autoersetzungbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) == 0 {
		an.klage("autoersetzung braucht ein Set.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "Set", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitAutoersetzung(ctx, bestandsort, an, func(ctx context.Context, weg *app.Autoersetzungsweg) int {
		if len(rest) == 1 {
			return autoersetzungenZeigen(ctx, weg, an, setID)
		}

		return autoersetzungshandlung(ctx, weg, an, setID, rest[1:])
	})
}

// autoersetzungenZeigen nennt die Regeln, die für dieses Set gelten.
func autoersetzungenZeigen(
	ctx context.Context, weg *app.Autoersetzungsweg, an *schreiber, setID int64,
) int {
	regeln, geschaltet, err := weg.Liste(ctx, setID)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an,
			map[string]any{"regeln": regeln, "an": geschaltet})
	}

	if !geschaltet {
		an.sag("aus\n")
	}

	for _, regel := range regeln {
		art := "paket"
		if regel.Eigen {
			art = "eigen"
		}

		zustand := "an"
		if !regel.Aktiv {
			zustand = "aus"
		}

		an.sag("%s\t%s\t%s\t%s\n", art, zustand, regel.Muster, regel.Ersatz)
	}

	return an.abschluss()
}

// autoersetzungshandlung führt setzen, loeschen, zuruecksetzen oder schalten.
func autoersetzungshandlung(
	ctx context.Context, weg *app.Autoersetzungsweg, an *schreiber,
	setID int64, rest []string,
) int {
	switch rest[0] {
	case "setzen":
		return autoersetzungSetzen(ctx, weg, an, setID, rest[1:])

	case "loeschen":
		return autoersetzungLoeschen(ctx, weg, an, setID, rest[1:])

	case "zuruecksetzen":
		return getanOderKlage(an, weg.Zuruecksetzen(ctx, setID))

	case "an":
		return getanOderKlage(an, weg.SchalterSetzen(ctx, setID, true))

	case "aus":
		return getanOderKlage(an, weg.SchalterSetzen(ctx, setID, false))
	}

	an.klage("»%s« ist keine Handlung: setzen, loeschen, zuruecksetzen, an oder aus.",
		rest[0])

	return CodeBedienung
}

// autoersetzungSetzen legt eine Regel an oder schreibt sie fort.
func autoersetzungSetzen(
	ctx context.Context, weg *app.Autoersetzungsweg, an *schreiber,
	setID int64, rest []string,
) int {
	if len(rest) < 2 {
		an.klage("autoersetzung <set> setzen braucht Muster und Ersatz.")

		return CodeBedienung
	}

	// **Ohne Angabe gilt „ganzes Wort" und „an"** — MSEs Vorgaben
	// (`AutoReplace::AutoReplace`, `src/data/word_list.cpp`, Zeilen 38–42,
	// Fassung f9d9356 vom 2021-07-06).
	ganzesWort, aktiv := true, true

	for _, wort := range rest[2:] {
		switch wort {
		case "teilwort":
			ganzesWort = false

		case "aus":
			aktiv = false
		}
	}

	return getanOderKlage(an, weg.Setzen(ctx, setID, app.AutoersetzungsregelDTO{
		Muster: rest[0], Ersatz: rest[1], GanzesWort: ganzesWort, Aktiv: aktiv,
	}))
}

// autoersetzungLoeschen nimmt eine eigene Regel heraus.
func autoersetzungLoeschen(
	ctx context.Context, weg *app.Autoersetzungsweg, an *schreiber,
	setID int64, rest []string,
) int {
	if len(rest) == 0 {
		an.klage("autoersetzung <set> loeschen braucht ein Muster.")

		return CodeBedienung
	}

	return getanOderKlage(an, weg.Loeschen(ctx, setID, rest[0]))
}

// getanOderKlage meldet den Erfolg oder den Fehler.
func getanOderKlage(an *schreiber, fehler error) int {
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	return an.abschluss()
}

// mitAutoersetzung baut den Weg über dem Bestand.
func mitAutoersetzung(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Autoersetzungsweg) int,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg, err := app.AutoersetzungspflegeUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return arbeit(ctx, weg)
	})
}
