package main

// **Das Kontextpaket auf der Kommandozeile** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Ein Befehl, der nichts sendet
//
// **`mse kontextpaket <aufgabe> <fall> <frage> <anbieter> <modell> <fenster>
// <reserve> [--belege=n]`** stellt zusammen, was an ein Modell ginge, und
// **gibt es aus**. Gerufen wird kein **generatives** Modell und geschrieben
// wird nichts; für die Vektorhälfte der Suche rechnet das lokale
// Einbettungsmodell mit.
//
// **Fenster und Ausgabereserve sind Pflicht und keine Wahl.** Sie gehören zum
// Zielmodell, und dieses Programm kennt es nicht: eine Vorgabe hier wäre eine
// Zahl, die für kein Modell gilt — und ein Paket, das das Fenster sprengt,
// schneidet das Modell still ab.
//
// **Das ist der Punkt.** Wer ein Skript schreibt, das ein Modell füttert,
// sieht hier vorher: welche Ausschnitte, wie viele Tokens, welche
// Datenkategorien — und ob eigener Inhalt dabei ist. Nach dem Senden ist
// diese Frage nicht mehr zu stellen.

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// kontextpaketbefehl stellt ein Paket zusammen und zeigt es.
func kontextpaketbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	rest, freigegeben := freigabeAus(rest)

	rest, belege, gedeutet := ohneBelegzahl(rest)
	if !gedeutet {
		an.klage("--belege braucht eine Zahl.\n%s", hilfe)

		return CodeBedienung
	}

	if len(rest) != 7 {
		an.klage("kontextpaket braucht Aufgabe, Fall, Frage, Anbieter, "+
			"Modell, Kontextfenster und Ausgabereserve. Fälle: %s.\n%s",
			faelle(), hilfe)

		return CodeBedienung
	}

	fenster, reserve, gelesen := fensterAus(rest)
	if !gelesen {
		an.klage("Fenster und Ausgabereserve sind Zahlen.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return kontextpaketSchreiben(an, app.RechercheAnwendungUeber(db).
			KontextZusammenstellen(ctx, rest[0], "", rest[1], rest[2],
				domain2grenze, rest[3], rest[4], fenster, reserve, belege,
				freigegeben))
	})
}

// domain2grenze ist die Zahl der Treffer, aus denen ausgewählt wird.
//
// **Sie ist nicht die Zahl der Belege.** Ausgewählt wird aus den Treffern,
// und wie viele davon ins Budget passen, entscheidet das Budget — `--belege`
// begrenzt zusätzlich.
const domain2grenze = 20

// kontextpaketSchreiben gibt das Paket aus.
func kontextpaketSchreiben(an *schreiber, antwort app.KontextpaketAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeberFehler(antwort.Fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("%s\n", antwort.Freigabe.Satz)
	an.sag("Paket:    %s (%s, läuft ab %s)\n", antwort.PaketID,
		antwort.Kontextpolicy, antwort.Laeuftab)
	an.sag("Budget:   %d von %d Tokens, %d freigehalten (%s)\n",
		antwort.Verbraucht, antwort.Kontextgrenze, antwort.Ausgabereserve,
		antwort.Zaehlweise)

	for _, beleg := range antwort.Belege {
		veraltet := ""
		if beleg.Veraltet {
			veraltet = "  ⚠ nicht der aktive Stand"
		}

		an.sag("[%s] %s  [%s %s]%s\n      %s\n", beleg.Kennung, beleg.Titel,
			beleg.Quelle, beleg.Quellrevision, veraltet,
			einzeilig(beleg.Ausschnitt))
	}

	if len(antwort.Belege) == 0 {
		an.sag("Keine Belege — %s\n", antwort.Ergebnis.Grund)
	}

	return an.abschluss()
}

// ohneBelegzahl schneidet `--belege=n` heraus.
//
// **Vorn oder hinten** — es ist eine Angabe über den Lauf und keine Stelle in
// der Reihe, genau wie `--delta` beim Indexaufbau.
func ohneBelegzahl(rest []string) (uebrig []string, belege int, gedeutet bool) {
	uebrig = make([]string, 0, len(rest))

	for _, teil := range rest {
		wert, ist := strings.CutPrefix(teil, "--belege=")
		if !ist {
			uebrig = append(uebrig, teil)

			continue
		}

		zahl, fehler := strconv.Atoi(wert)
		if fehler != nil {
			return rest, 0, false
		}

		belege = zahl
	}

	return uebrig, belege, true
}

// fensterAus liest die beiden Budgetzahlen.
//
// **Beide sind Pflicht** — der Grund steht im Kopf dieser Datei.
func fensterAus(rest []string) (fenster, reserve int, gelesen bool) {
	for stelle, ziel := range []*int{&fenster, &reserve} {
		zahl, fehler := strconv.Atoi(rest[5+stelle])
		if fehler != nil {
			return 0, 0, false
		}

		*ziel = zahl
	}

	return fenster, reserve, true
}
