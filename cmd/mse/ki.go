// **Die KI an der Kommandozeile** — Werkstück A6 von Phase 5.3
// (#1445).
//
// **Eine eigene Datei**, weil `preise.go` mit ihr über die Zeilengrenze aus §2
// ginge. Der Schnitt folgt der Frage: dort steht, was etwas **kostet**, hier,
// was ein Modell **kann**.

package main

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// kimodellebefehl nennt, was über KI-Modelle feststeht.
func kimodellebefehl(ctx context.Context, bestandsort string, an *schreiber, _ []string) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.KIAnwendungUeber(db).KIModelle(ctx)
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort.KIKatalogDTO)
		}

		// **Die Regel steht vor der Liste** — sie sagt, ob überhaupt etwas
		// diesen Rechner verlassen darf.
		an.sag("nur lokal\t%t\n", antwort.NurLokal)

		// **Und die Einrichtung davor** (A7, #1446): eine leere Liste sagt
		// nicht, **warum** sie leer ist.
		an.sag("einrichtung\t%s\n", antwort.Einrichtung)

		// **Je Anbieter eine Zeile** (A8, #1447) — welcher woran hängt, ist
		// die Auskunft, die weiterhilft.
		for _, anbieter := range antwort.Anbieter {
			an.sag("anbieter\t%s\t%s\t%s\n",
				anbieter.Name, anbieter.Einrichtung, anbieter.Laufzeit)

			for _, befund := range anbieter.Befunde {
				an.sag("  befund\t%s\n", befund)
			}
		}

		an.sag("cloud\t%s\t%s\n", antwort.Cloud.Stand, antwort.Cloud.Beleg)

		for _, befund := range antwort.Befunde {
			an.sag("befund\t%s\n", befund)
		}

		// **Wen der Vorrang fragt** (#2054) — vor der Liste, wie die Regel.
		an.sag("vorrang\t%s\n", strings.Join(antwort.Vorrang, ", "))

		for _, modell := range antwort.Modelle {
			an.sag("%s/%s\t%s\t%s\n",
				modell.Anbieter, modell.Alias, modell.Ort, modell.Laufzeit)

			for _, beleg := range modell.Faehigkeiten {
				an.sag("  %s\tgeprüft %s\t%s\n",
					beleg.Faehigkeit, beleg.Datum, beleg.Beleg)
			}

			// **Das Fenster mit seiner Herkunft** (AI-01.AK-02, #1742).
			if modell.Kontextfenster > 0 {
				an.sag("  kontextfenster\t%d\t%s\n", modell.Kontextfenster, modell.KontextBeleg)
			} else {
				an.sag("  kontextfenster\tnicht angegeben\n")
			}

			if len(modell.Herkunft) == 0 {
				an.sag("  herkunft\tunbekannt\n")
			}
		}

		return an.abschluss()
	})
}

// kivorschlagbefehl lässt ein Modell Felder einer Karte vorschlagen.
//
// **Ohne `--uebernehmen` wird nichts geschrieben.** Am Fenster sieht ein
// Mensch den Vorschlag und drückt dann; hier sagt er es vorher — die Flagge
// ist die ausdrückliche Entscheidung und kein Nebeneffekt
// ([app.App.KIVorschlagInEinemZug]).
func kivorschlagbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	// **`--freigeben` vor der Stellenzählung** (A13.4, #1452) — dieselbe
	// Form wie bei der Recherche, damit der Schalter überall gleich heißt.
	rest, freigegeben := freigabeAus(rest)

	if len(rest) < 4 {
		an.klage("ki-vorschlag braucht die Karte, den Anbieter, das Modell "+
			"und den Auftrag.\n%s", hilfe)

		return CodeBedienung
	}

	uebernehmen := false

	// felder nennt, **was** übernommen wird — leer heißt: alles, was das
	// Schema durchgelassen hat (A11, #1450).
	var felder []string

	// belegfrage geht an den Suchindex — leer heißt: gar nicht fragen
	// (A13.4, #1452).
	belegfrage := ""

	// nur nennt, **wonach** gefragt wird — leer heißt: alle Felder; Name und
	// Flavortext machen eine kleine Aufgabe (AI-05, #2042).
	var nur []string

	// **Jeder Aufruf ist eine Sitzung** (AI-05.AK-04): wer das Gerätemodell
	// hier nicht will, sagt es je Aufruf.
	ohneGeraet := false

	for _, flagge := range rest[4:] {
		switch {
		case flagge == "--uebernehmen":
			uebernehmen = true

		case strings.HasPrefix(flagge, "--felder="):
			felder = strings.Split(
				strings.TrimPrefix(flagge, "--felder="), ",")

		case strings.HasPrefix(flagge, "--belegfrage="):
			belegfrage = strings.TrimPrefix(flagge, "--belegfrage=")

		case strings.HasPrefix(flagge, "--nur="):
			nur = strings.Split(strings.TrimPrefix(flagge, "--nur="), ",")

		case flagge == "--ohne-geraetemodell":
			ohneGeraet = true

		default:
			an.klage("ki-vorschlag kennt %q nicht.\n%s", flagge, hilfe)

			return CodeBedienung
		}
	}

	karte, fehler := strconv.ParseInt(rest[0], 10, 64)
	if fehler != nil {
		an.klage("die Kartenkennung %q ist keine Zahl.\n%s", rest[0], hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		anwendung := app.KIAnwendungUeber(db)
		if ohneGeraet {
			if aus := anwendung.KIGeraetemodellFuerSitzung(ctx, false); aus.Fehler != nil {
				an.klageUeber(aus.Fehler.AlsFehler())

				return CodeFachlich
			}
		}

		antwort := anwendung.KIVorschlagInEinemZug(ctx,
			app.KIVorschlagsbezug{
				KarteID: karte, Anbieter: rest[1], Alias: rest[2],
				Auftrag: rest[3], Belegfrage: belegfrage,
				Freigegeben: freigegeben, Nur: nur,
			},
			uebernehmen, felder)
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort.KIVorschlagDTO)
		}

		an.sag("anbieter\t%s\t%s\n", antwort.Anbieter, antwort.Weg)
		an.sag("modell\t%s\n", antwort.Modell)
		an.sag("kennung\t%s\n", antwort.Kennung)

		if antwort.Ausgang != "" {
			an.sag("ausgang\t%s\n", antwort.Ausgang)
		}

		for _, anbieter := range antwort.Rueckfaelle {
			an.sag("rueckfall\t%s\n", anbieter)
		}
		an.sag("uebernommen\t%d\n", antwort.Uebernommen)

		for _, feld := range antwort.Felder {
			an.sag("%s\t%s\n", feld.Feld, strings.Join(feld.Werte, ", "))
		}

		for _, befund := range antwort.Befunde {
			an.sag("befund\t%s\n", befund)
		}

		return an.abschluss()
	})
}
