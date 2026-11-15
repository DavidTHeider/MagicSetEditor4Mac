package main

// **Der belegte Regelhinweis an der Kommandozeile** — Werkstück A13.2a von
// Phase 5.3 (#1452).
//
// `mse ki-einschaetzung <karte> <art> <anbieter> <modell> <frage>` gibt die
// Einschätzung, ihre Belege, ihre Befunde und den Vorbehalt — in derselben
// Gestalt wie über MCP und den Dienst.

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// kieinschaetzungbefehl lässt ein Modell eine Karte beurteilen.
//
// **Er schreibt nichts** — auch nicht mit einer Flagge, denn es gibt keine.
// Eine Einschätzung ist eine Auskunft; das Issue führt „Keine Regel-/
// Sammlungs-/Spielzustandsänderung beim Analysieren" als eigene Abnahme.
func kieinschaetzungbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	rest, freigegeben := freigabeAus(rest)

	if len(rest) < 5 || len(rest) > 6 {
		an.klage("ki-einschaetzung braucht die Karte, die Art, den "+
			"Anbieter, das Modell und die Frage, wahlweise dazu das "+
			"Format.\n%s", hilfe)

		return CodeBedienung
	}

	// **Und `--freigeben` gibt private Korpora frei** (A13.4, #1452) — es
	// steht vor der Stellenzählung, damit es an jeder Stelle stehen darf.
	// Ohne es bleibt der eigene Bestand außen, auch bei einem Vergleich.

	// **Das Format steht hinten und ist wahlfrei** (A13.2b, #1452): ein
	// Regelhinweis hat keines, und ein Vergleich ohne Format gilt für keines
	// besonders — die Antwort sagt es dann.
	format := ""
	if len(rest) == 6 {
		format = rest[5]
	}

	karte, fehler := strconv.ParseInt(rest[0], 10, 64)
	if fehler != nil {
		an.klage("die Kartenkennung %q ist keine Zahl.\n%s", rest[0], hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.KIAnwendungUeber(db).KIEinschaetzen(ctx,
			app.KIEinschaetzungsbezug{
				Art: rest[1], KarteID: karte, Anbieter: rest[2],
				Alias: rest[3], Frage: rest[4], Format: format,
				Freigegeben: freigegeben,
			})
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		return einschaetzungSchreiben(an, antwort)
	})
}

// einschaetzungSchreiben gibt sie als Zeilen.
//
// **Der Vorbehalt zuerst und nicht zuletzt.** Wer eine Ausgabe abschneidet,
// schneidet das Ende ab — und dann stünde die Einschätzung ohne ihn da.
func einschaetzungSchreiben(an *schreiber, antwort app.KIEinschaetzungAntwort) int {
	an.sag("vorbehalt\t%s\n", antwort.Vorbehalt)
	an.sag("art\t%s\n", antwort.Art)
	an.sag("modell\t%s\n", antwort.Modell)
	an.sag("belegt\t%t\n", antwort.Belegt)
	an.sag("satz\t%s\n", antwort.Satz)

	for _, einschraenkung := range antwort.Einschraenkungen {
		an.sag("einschraenkung\t%s\n", einschraenkung)
	}

	for _, befund := range antwort.Befunde {
		an.sag("befund\t%s\t%s\t%s\n", befund.Art, befund.Stelle, befund.Meldung)
	}

	for _, beleg := range antwort.Belege {
		an.sag("beleg\t%s\t%s\t%s\t%s\n", beleg.Kennung, beleg.Quelle,
			beleg.Quelldatum, beleg.Titel)
	}

	for _, stand := range antwort.Korpusstaende {
		an.sag("korpus\t%s\t%s\t%s\n", stand.Quelle, stand.Revision, stand.Datum)
	}

	return an.abschluss()
}
