// **Turniere an der Kommandozeile** — Werkstück 5 von Phase 4.7
// (#1294),
// nach Vermerk 0062.
//
// Einordnung nach §16: **erweiternd** — neue Befehle, keine geänderten.

package main

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// turnierbefehl bedient `mse turniere` und `mse turnier …`.
func turnierbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	return mitTurnieren(ctx, bestandsort, an, func(ctx context.Context, flaeche *app.Turnierflaeche) int {
		if len(rest) == 0 {
			return turnierlisteZeigen(ctx, flaeche, an, "")
		}

		switch rest[0] {
		case "neu":
			return turnierAnlegen(ctx, flaeche, an, rest[1:])

		case "loeschen":
			return turnierLoeschen(ctx, flaeche, an, rest[1:])

		case "kopieren":
			return turnierKopieren(ctx, flaeche, an, rest[1:])

		case "anmelden":
			return turnierAnmelden(ctx, flaeche, an, rest[1:])

		case "aussteigen":
			return turnierAussteigen(ctx, flaeche, an, rest[1:], true)

		case "einsteigen":
			return turnierAussteigen(ctx, flaeche, an, rest[1:], false)

		case "vorschlag":
			return turnierVorschlag(ctx, flaeche, an, rest[1:])

		case "runde":
			return turnierRunde(ctx, flaeche, an, rest[1:])

		case "ergebnis":
			return turnierErgebnis(ctx, flaeche, an, rest[1:])

		case "folgen":
			return turnierFolgen(ctx, flaeche, an, rest[1:])

		case "abschliessen":
			return turnierAbschluss(ctx, flaeche, an, rest[1:], true)

		case "wiedereroeffnen":
			return turnierAbschluss(ctx, flaeche, an, rest[1:], false)

		case "protokoll":
			return turnierProtokoll(ctx, flaeche, an, rest[1:])

		case "rangliste":
			return turnierRangliste(ctx, flaeche, an, rest[1:])

		case "ausgeben":
			return turnierdateiAus(ctx, flaeche, an, rest[1:], false)

		case "bericht":
			return turnierdateiAus(ctx, flaeche, an, rest[1:], true)

		case "einlesen":
			return turnierEinlesen(ctx, flaeche, an, rest[1:])
		}

		return turnierZeigen(ctx, flaeche, an, rest[0])
	})
}

// turnierlisteZeigen nennt die Turniere.
func turnierlisteZeigen(
	ctx context.Context, flaeche *app.Turnierflaeche, an *schreiber, spiel string,
) int {
	antwort := flaeche.Turniere(ctx, spiel)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	for _, turnier := range antwort.Turniere {
		an.sag("%s\t%s\t%s\t%s\n", turnier.ID, turnier.Spiel, turnier.Name,
			turnier.Notiz)
	}

	return an.abschluss()
}

// turnierZeigen gibt ein Turnier samt Verlauf.
func turnierZeigen(
	ctx context.Context, flaeche *app.Turnierflaeche, an *schreiber, kennung string,
) int {
	antwort := flaeche.Turnier(ctx, kennung)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	turnier := antwort.Turnier

	an.sag("%s — %s, %d Teilnehmer, %d Runden\n", turnier.Name, turnier.Spiel,
		len(turnier.Teilnehmer), len(turnier.Runden))

	// **Ein abgeschlossenes Turnier sagt es im Kopf** — DOM-04.AK-04 (#1699).
	if turnier.Abgeschlossen != "" {
		an.sag("abgeschlossen am %s\n", turnier.Abgeschlossen)
	}

	for _, teilnehmer := range turnier.Teilnehmer {
		zusatz := ""
		if teilnehmer.Ausgestiegen {
			zusatz = "\tausgestiegen"
		}

		an.sag("  %s\t%s%s\n", teilnehmer.ID, teilnehmer.Name, zusatz)
	}

	for _, runde := range turnier.Runden {
		an.sag("Runde %d\n", runde.Nummer)

		for _, paarung := range runde.Paarungen {
			an.sag("  %s\t%s\n", paarung.ID, paarungstext(paarung))
		}
	}

	return an.abschluss()
}

// paarungstext beschreibt eine Paarung in einer Zeile.
//
// **Ein Bye sagt „Bye" und nicht „gegen 0".** Der Unterschied ist der ganze
// Punkt: er ist ein Ergebnis, kein fehlender Gegner.
func paarungstext(paarung app.PaarungDTO) string {
	wer := paarung.A + " gegen " + paarung.B
	if paarung.B == "" {
		wer = paarung.A + " — Bye"
	}

	if paarung.Ergebnis == nil {
		return wer + "\toffen"
	}

	return wer + "\t" + strconv.Itoa(paarung.Ergebnis.SpieleA) + ":" +
		strconv.Itoa(paarung.Ergebnis.SpieleB) + ":" +
		strconv.Itoa(paarung.Ergebnis.Unentschieden)
}

// mitTurnieren baut die Turnierfläche über dem Bestand.
//
// **Ohne Fehlerzweig**, anders als bei den Decks: Turniere hängen an keiner
// Gegenstelle.
func mitTurnieren(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Turnierflaeche) int,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return arbeit(ctx, app.TurnierflaecheUeber(db))
	})
}
