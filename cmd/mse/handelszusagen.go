package main

// **Welche Besitzzeilen weggehen** —
// #1643.
//
// # Warum zwei Befehle und nicht einer mit einem Schalter
//
// **Weil das eine eine Entscheidung ist und das andere eine Auskunft.** Wer
// festhält, bindet Bestand; wer nachsieht, bindet nichts — und ein Befehl,
// der beides täte, hätte einen Modus, den man vergisst.

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// zusagensicht ist eine Zusage, wie die Kommandozeile sie schreibt.
//
// **Eine eigene Gestalt** — die Wails-DTO ist **innen**, diese Fläche
// **außen** (0033, Regel 5).
type zusagensicht struct {
	ExemplarID int64  `json:"exemplarId"`
	Menge      int    `json:"menge"`
	Fassung    int64  `json:"fassung"`
	Revision   int    `json:"revision"`
	Frist      string `json:"frist,omitempty"`
	Uebergeben bool   `json:"uebergeben"`
}

// handelFesthaltenbefehl fährt
// `mse handel-festhalten <vorgang> <exemplarID> <menge> <fassung>`.
func handelFesthaltenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 4 {
		an.klage("handel-festhalten nimmt Vorgang, Besitzzeile, Menge und "+
			"Fassung.\n%s", hilfe)

		return CodeBedienung
	}

	exemplarID, mengefehler := strconv.ParseInt(rest[1], 10, 64)
	menge, zahlfehler := strconv.Atoi(rest[2])
	fassung, fassungfehler := strconv.ParseInt(rest[3], 10, 64)

	if mengefehler != nil || zahlfehler != nil || fassungfehler != nil {
		an.klage("Besitzzeile, Menge und Fassung sind Zahlen.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an,
		func(ctx context.Context, db *sql.DB) int {
			return zusagenSchreiben(an, app.PeerflaecheUeber(db).HandelFesthalten(
				ctx, rest[0], exemplarID, menge, fassung))
		})
}

// handelszusagenbefehl fährt `mse handelszusagen <vorgang>`.
func handelszusagenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("handelszusagen nimmt ein Argument.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an,
		func(ctx context.Context, db *sql.DB) int {
			return zusagenSchreiben(an,
				app.PeerflaecheUeber(db).Handelszusagen(ctx, rest[0]))
		})
}

// zusagenSchreiben gibt aus, was festgehalten ist.
func zusagenSchreiben(an *schreiber, antwort app.ZusagenAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	// **Auch mit `--json` nur auf dem Fehlerstrom**: die Ausgabe ist eine
	// Liste, und ein Feld daneben gäbe es nur um den Preis einer anderen
	// Gestalt (0033).
	nachladenMelden(an, antwort.Nachladefehler)

	sichten := make([]zusagensicht, 0, len(antwort.Zusagen))

	for _, zusage := range antwort.Zusagen {
		sichten = append(sichten, zusagensicht{
			ExemplarID: zusage.ExemplarID, Menge: zusage.Menge,
			Fassung: zusage.Fassung, Revision: zusage.Revision,
			Frist: zusage.Frist, Uebergeben: zusage.Uebergeben,
		})
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, sichten)
	}

	for _, sicht := range sichten {
		an.sag("%d\t%d\t%d\t%v\n",
			sicht.ExemplarID, sicht.Menge, sicht.Revision, sicht.Uebergeben)
	}

	return an.abschluss()
}
