package main

// **Die Recherche auf der Kommandozeile** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Drei Befehle
//
//   - `mse recherche <fall> <frage> [grenze]` fragt den Wissensbestand.
//   - `mse aehnlich <karte> [grenze]` sucht Ähnliches zu einer eigenen Karte.
//   - `mse kontextpaket <aufgabe> <fall> <frage> …` stellt zusammen, **was
//     an ein Modell ginge** — und sendet es nicht.
//
// # Der dritte ist der, der hier am meisten trägt
//
// **Weil er nichts tut.** Er zeigt, welche Ausschnitte hinausgingen, wie
// viele Tokens sie kosten und ob eigener Inhalt dabei ist — vor dem ersten
// bezahlten Ruf. Wer ein Skript schreibt, das ein Modell füttert, sieht hier,
// womit.

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// recherchebefehl beantwortet eine Frage gegen den Wissensbestand.
func recherchebefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	rest, freigegeben := freigabeAus(rest)

	if len(rest) < 2 || len(rest) > 3 {
		an.klage("recherche braucht den Fall und die Frage, dazu wahlweise "+
			"die Trefferzahl. Fälle: %s.\n%s", faelle(), hilfe)

		return CodeBedienung
	}

	grenze, gedeutet := grenzeAus(rest[1:])
	if !gedeutet {
		an.klage("die Trefferzahl ist keine Zahl.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return rechercheSchreiben(an, app.RechercheAnwendungUeber(db).
			Recherchieren(ctx, rest[0], rest[1], grenze,
				app.RecherchefilterDTO{Freigegeben: freigegeben}, ""))
	})
}

// aehnlichbefehl sucht Ähnliches zu einer eigenen Karte.
func aehnlichbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	rest, freigegeben := freigabeAus(rest)

	if len(rest) < 1 || len(rest) > 2 {
		an.klage("aehnlich braucht die Karte, dazu wahlweise die "+
			"Trefferzahl.\n%s", hilfe)

		return CodeBedienung
	}

	karteID, fehler := strconv.ParseInt(rest[0], 10, 64)
	if fehler != nil {
		an.klage("die Karte ist keine Zahl.\n%s", hilfe)

		return CodeBedienung
	}

	grenze, gedeutet := grenzeAus(rest)
	if !gedeutet {
		an.klage("die Trefferzahl ist keine Zahl.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		// **Ohne Stand**: wer hier fragt, hat die Karte nicht vor sich — der
		// Stand schützt eine Anzeige, die es auf der Kommandozeile nicht gibt.
		return rechercheSchreiben(an, app.RechercheAnwendungUeber(db).
			AehnlicheKarten(ctx, karteID, "", grenze,
				app.RecherchefilterDTO{Freigegeben: freigegeben}, ""))
	})
}

// freigabeAus zieht `--freigeben <korpus>` aus den Argumenten — A13.4 (#1452).
//
// **Mehrfach erlaubt**: jede Nennung gibt einen weiteren privaten Korpus
// frei. Ohne sie bleibt der eigene Bestand außen, und die Antwort sagt es
// unter »einschraenkungen«.
//
// **Vor der Stellenzählung**, damit der Schalter an jeder Stelle stehen darf
// — dieselbe Form wie `--revision` beim Zurücksetzen.
func freigabeAus(rest []string) (ohne, freigegeben []string) {
	ohne = make([]string, 0, len(rest))
	freigegeben = make([]string, 0, 2)

	for stelle := 0; stelle < len(rest); stelle++ {
		if rest[stelle] != "--freigeben" {
			ohne = append(ohne, rest[stelle])

			continue
		}

		if stelle+1 < len(rest) {
			freigegeben = append(freigegeben, rest[stelle+1])
			stelle++
		}
	}

	return ohne, freigegeben
}

// rechercheSchreiben gibt eine Antwort aus.
func rechercheSchreiben(an *schreiber, antwort app.RechercheAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeberFehler(antwort.Fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("Modus: %s", antwort.Modus)

	if antwort.Leere != "" {
		an.sag(" (%s: %s)", antwort.Leere, antwort.Grund)
	}

	an.sag("\n")

	for _, treffer := range antwort.Treffer {
		an.sag("%2d  %s  [%s %s]\n      %s\n", treffer.Rang, treffer.Titel,
			treffer.Quelle, treffer.Quellrevision,
			einzeilig(treffer.Ausschnitt))
	}

	for _, satz := range antwort.Einschraenkungen {
		an.sag("Hinweis: %s\n", satz)
	}

	return an.abschluss()
}

// einzeilig macht aus einem Ausschnitt eine Zeile.
//
// **Weil ein Kartentext Zeilenumbrüche hat** — Typzeile, Regeltext,
// Aromatext —, und der zweite davon stünde auf der Kommandozeile linksbündig
// unter dem Rang statt eingerückt unter seinem Treffer. Gesehen am
// 2026-09-12 an einem gebauten Index.
func einzeilig(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// faelle nennt die vier Fälle in ihrer Reihenfolge.
//
// **Aus `domain` und nicht als Text hier**: eine Aufzählung neben der Quelle
// läuft ihr davon, und die Fehlermeldung nennt dann einen Fall, den es nicht
// gibt (oder verschweigt einen, den es gibt).
func faelle() string {
	namen := make([]string, 0, len(domain.Recherchefaelle()))
	for _, fall := range domain.Recherchefaelle() {
		namen = append(namen, string(fall))
	}

	return strings.Join(namen, ", ")
}
