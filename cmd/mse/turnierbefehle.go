// **Die schreibenden Turnierbefehle** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, damit `turniere.go` unter der Zeilengrenze aus §2
// bleibt. Der Schnitt folgt der Frage: dort wird **gezeigt**, hier
// **geändert**.

package main

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// turnierAnlegen legt ein Turnier an.
func turnierAnlegen(
	ctx context.Context, flaeche *app.Turnierflaeche, an *schreiber, rest []string,
) int {
	rest, format, gut := formatAus(rest)
	if !gut {
		an.klage("--format braucht die Kennung eines Formats.\n%s", hilfe)

		return CodeBedienung
	}

	if len(rest) < 2 {
		an.klage("turnier neu braucht einen Namen und ein Spiel.\n%s", hilfe)

		return CodeBedienung
	}

	notiz := ""
	if len(rest) > 2 {
		notiz = rest[2]
	}

	// **»einzelspiel« als viertes Wort und nicht als Kennzeichen.** Dieselbe
	// Überlegung wie bei `format setzen`: eine Stelle, die man nennen kann,
	// ist besser zu finden als ein Schalter, den man kennen muss.
	einzelspiel := len(rest) > 3 && rest[3] == "einzelspiel"

	antwort := flaeche.TurnierAnlegen(ctx, rest[0], rest[1], notiz, einzelspiel, format)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("%s\n", antwort.Turnier.ID)

	return an.abschluss()
}

// turnierLoeschen nimmt ein Turnier samt seinem Verlauf heraus.
func turnierLoeschen(
	ctx context.Context, flaeche *app.Turnierflaeche, an *schreiber, rest []string,
) int {
	if len(rest) < 1 {
		an.klage("turnier loeschen braucht eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.TurnierLoeschen(ctx, rest[0])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	return getan(an, nil, "Gelöscht.\n")
}

// turnierAnmelden trägt einen Teilnehmer ein.
func turnierAnmelden(
	ctx context.Context, flaeche *app.Turnierflaeche, an *schreiber, rest []string,
) int {
	if len(rest) < 2 {
		an.klage("turnier anmelden braucht Turnier und Namen.\n%s", hilfe)

		return CodeBedienung
	}

	deck := ""
	if len(rest) > 2 {
		deck = rest[2]
	}

	antwort := flaeche.TurnierAnmelden(ctx, rest[0], rest[1], deck)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("%s\n", antwort.Teilnehmer.ID)

	return an.abschluss()
}

// turnierAussteigen meldet einen Teilnehmer ab — oder wieder an.
func turnierAussteigen(
	ctx context.Context, flaeche *app.Turnierflaeche, an *schreiber,
	rest []string, ausgestiegen bool,
) int {
	// **Turnier und Teilnehmer.** Die Kennung eines Teilnehmers ist global;
	// ohne das Turnier daneben änderte ein Tippfehler jemanden anderswo
	// (Befund 6 des Zweitblicks über Phase 4.7).
	if len(rest) < 2 {
		an.klage("dieser Befehl braucht Turnier und Teilnehmer.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.TurnierAussteigen(ctx, rest[0], rest[1], ausgestiegen)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	return getan(an, nil, "Gesetzt.\n")
}

// turnierVorschlag zeigt die nächste Runde, **ohne sie zu legen**.
func turnierVorschlag(
	ctx context.Context, flaeche *app.Turnierflaeche, an *schreiber, rest []string,
) int {
	if len(rest) < 1 {
		an.klage("turnier vorschlag braucht eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.TurnierPaarungsvorschlag(ctx, rest[0])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	for _, paarung := range antwort.Paarungen {
		an.sag("%s\n", paarungstext(paarung))
	}

	return an.abschluss()
}

// turnierRunde paart die nächste Runde und legt sie an.
func turnierRunde(
	ctx context.Context, flaeche *app.Turnierflaeche, an *schreiber, rest []string,
) int {
	if len(rest) < 1 {
		an.klage("turnier runde braucht eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.TurnierRundeLegen(ctx, rest[0])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("Runde %d\n", antwort.Runde.Nummer)

	for _, paarung := range antwort.Runde.Paarungen {
		an.sag("  %s\t%s\n", paarung.ID, paarungstext(paarung))
	}

	return an.abschluss()
}

// turnierErgebnis trägt einen Spielstand ein — **»weg« nimmt ihn zurück**.
func turnierErgebnis(
	ctx context.Context, flaeche *app.Turnierflaeche, an *schreiber, rest []string,
) int {
	if len(rest) < 3 {
		an.klage("turnier ergebnis braucht Turnier, Paarung und einen Stand.\n%s",
			hilfe)

		return CodeBedienung
	}

	// **»weg« statt drei Nullen.** Ein Stand von 0:0:0 ist ein anderer Satz
	// als „doch nicht gespielt", und die Kommandozeile muss beide sagen
	// können.
	// **Die Fassung wird hier gelesen und mitgegeben** (§7, Regel 2). Die
	// Kommandozeile ist einschrittig — sie liest und schreibt in einem
	// Atemzug, und einen Stand von vorhin gäbe es gar nicht.
	gelesen := flaeche.Turnier(ctx, rest[0])
	if gelesen.Fehler != nil {
		return ergebnisAntwort(an, app.SchreibAntwort{Fehler: gelesen.Fehler})
	}

	if rest[2] == "weg" {
		return ergebnisAntwort(an, flaeche.TurnierErgebnisSetzen(
			ctx, rest[0], gelesen.Turnier.Fassung, rest[1], nil))
	}

	stand, gut := spielstand(an, rest[2:])
	if !gut {
		return CodeBedienung
	}

	return ergebnisAntwort(an, flaeche.TurnierErgebnisSetzen(
		ctx, rest[0], gelesen.Turnier.Fassung, rest[1], stand))
}

// spielstand liest bis zu drei Zahlen — gewonnen, verloren, unentschieden.
func spielstand(an *schreiber, rest []string) (*app.ErgebnisDTO, bool) {
	zahlen := []int{0, 0, 0}

	for stelle, roh := range rest {
		if stelle >= len(zahlen) {
			break
		}

		wert, err := strconv.Atoi(roh)
		if err != nil {
			an.klage("%q ist keine Zahl.\n%s", roh, hilfe)

			return nil, false
		}

		zahlen[stelle] = wert
	}

	return &app.ErgebnisDTO{
		SpieleA: zahlen[0], SpieleB: zahlen[1], Unentschieden: zahlen[2],
	}, true
}

// ergebnisAntwort meldet, was aus dem Setzen wurde.
func ergebnisAntwort(an *schreiber, antwort app.SchreibAntwort) int {
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	return getan(an, nil, "Gesetzt.\n")
}

// turnierRangliste rechnet den Stand.
func turnierRangliste(
	ctx context.Context, flaeche *app.Turnierflaeche, an *schreiber, rest []string,
) int {
	if len(rest) < 1 {
		an.klage("turnier rangliste braucht eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.TurnierRangliste(ctx, rest[0])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	for _, rang := range antwort.Rangliste {
		zusatz := ""
		if rang.Ausgestiegen {
			zusatz = "\tausgestiegen"
		}

		an.sag("%d.\t%s\t%d Punkte\t%d-%d-%d\tOMW %.4f\tGW %.4f\tOGW %.4f%s\n",
			rang.Platz, rang.Name, rang.Matchpunkte,
			rang.Siege+rang.Byes, rang.Niederlagen, rang.Unentschieden,
			rang.Gegnermatchquote, rang.Spielquote, rang.Gegnerspielquote, zusatz)
	}

	return an.abschluss()
}
