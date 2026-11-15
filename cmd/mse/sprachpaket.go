// **Die eigenen Übersetzungen an der Kommandozeile** — LOC-02
// (#1688).
//
// Einordnung nach §16: **erweiternd** — ein Befehl mit sieben Unterbefehlen
// kommt hinzu. **Ausgeben ohne Ziel schreibt den Text auf die Ausgabe**, wie
// bei `symbolbibliothek ausgeben`. Welche Übersetzung gilt, setzt
// `einstellungen setzen sprache.eigene <kennung>`.

package main

import (
	"context"
	"database/sql"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// ersetzenfahne erlaubt dem Einlesen, eine vorhandene Übersetzung zu
// ersetzen.
const ersetzenfahne = "--ersetzen"

// sprachpaketbefehl verteilt auf die Unterbefehle.
func sprachpaketbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		flaeche := app.SprachflaecheUeber(db)

		wort, argumente := "liste", []string{}
		if len(rest) > 0 {
			wort, argumente = rest[0], rest[1:]
		}

		weg, bekannt := sprachwege[wort]
		if !bekannt {
			an.klage("sprachpaket kennt %q nicht.\n%s", wort, hilfe)

			return CodeBedienung
		}

		if len(argumente) < weg.noetig {
			an.klage("sprachpaket %s braucht %d Angaben.\n%s", wort, weg.noetig, hilfe)

			return CodeBedienung
		}

		return weg.fuehre(ctx, flaeche, an, argumente)
	})
}

// sprachweg ist ein Unterbefehl und die Zahl der Angaben, die er braucht.
type sprachweg struct {
	noetig int
	fuehre func(ctx context.Context, flaeche *app.Sprachflaeche, an *schreiber, argumente []string) int
}

// sprachwege sind die sieben Unterbefehle.
var sprachwege = map[string]sprachweg{
	"liste": {0, sprachpaketeZeigen},
	"zeige": {1, func(ctx context.Context, flaeche *app.Sprachflaeche, an *schreiber, argumente []string) int {
		return sprachantwort(an, flaeche.Sprachpaket(ctx, argumente[0]), true)
	}},
	"anlegen": {3, func(ctx context.Context, flaeche *app.Sprachflaeche, an *schreiber, argumente []string) int {
		return sprachantwort(an, flaeche.SprachpaketAnlegen(ctx, argumente[0], argumente[1], argumente[2]), false)
	}},
	"setze": {1, sprachpaketSetzen},
	"loeschen": {1, func(ctx context.Context, flaeche *app.Sprachflaeche, an *schreiber, argumente []string) int {
		if antwort := flaeche.SprachpaketLoeschen(ctx, argumente[0]); antwort.Fehler != nil {
			return klageUeberNaht(an, antwort.Fehler)
		}

		an.sag("Gelöscht.\n")

		return an.abschluss()
	}},
	"ausgeben": {1, sprachpaketAusgeben},
	"einlesen": {1, func(ctx context.Context, flaeche *app.Sprachflaeche, an *schreiber, argumente []string) int {
		ersetzen := slices.Contains(argumente, ersetzenfahne)
		argumente = slices.DeleteFunc(slices.Clone(argumente), func(wort string) bool { return wort == ersetzenfahne })

		if len(argumente) == 0 {
			an.klage("sprachpaket einlesen braucht eine Datei.\n%s", hilfe)

			return CodeBedienung
		}

		return sprachantwort(an, flaeche.SprachpaketEinlesenLesen(ctx, argumente[0], ersetzen), false)
	}},
}

// sprachpaketeZeigen nennt die Übersetzungen — je Zeile Kennung, Basis,
// Stand, übersetzt von gesamt, fehlend, veraltet und Name.
func sprachpaketeZeigen(ctx context.Context, flaeche *app.Sprachflaeche, an *schreiber, _ []string) int {
	antwort := flaeche.Sprachpakete(ctx)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort.Pakete)
	}

	if len(antwort.Pakete) == 0 {
		an.sag("Noch keine eigene Übersetzung — »sprachpaket anlegen <kennung> <name> <basis>« legt eine an.\n")
	}

	for _, paket := range antwort.Pakete {
		an.sag("%s\t%s\t%s\t%d/%d\t%d fehlen\t%d veraltet\t%s\n", paket.Kennung, paket.Basis, paket.Stand,
			paket.Uebersetzt, paket.Gesamt, paket.Fehlend, paket.Veraltet, paket.Name)
	}

	return an.abschluss()
}

// sprachpaketSetzen nimmt Angaben als `--name=…` und Sätze als
// `pfad=text`.
func sprachpaketSetzen(ctx context.Context, flaeche *app.Sprachflaeche, an *schreiber, argumente []string) int {
	angaben := app.SprachpaketangabenDTO{}
	felder := map[string]**string{
		"--name": &angaben.Name, "--stand": &angaben.Stand, "--lizenz": &angaben.Lizenz,
		"--herkunft": &angaben.Herkunft,
	}
	texte := map[string]string{}

	for _, angabe := range argumente[1:] {
		schluessel, wert, gefunden := strings.Cut(angabe, "=")
		ziel, istAngabe := felder[schluessel]

		switch {
		case !gefunden || (strings.HasPrefix(schluessel, "--") && !istAngabe):
			an.klage("sprachpaket setze versteht %q nicht — erwartet wird --name=…, --stand=…, "+
				"--lizenz=…, --herkunft=… oder pfad=text.\n", angabe)

			return CodeBedienung
		case istAngabe:
			*ziel = &wert
		default:
			texte[schluessel] = wert
		}
	}

	return sprachantwort(an, flaeche.SprachpaketAendern(ctx, argumente[0], angaben, texte), false)
}

// sprachpaketAusgeben schreibt eine Übersetzung in eine Datei — **ohne Ziel
// als Text auf die Ausgabe**.
func sprachpaketAusgeben(ctx context.Context, flaeche *app.Sprachflaeche, an *schreiber, argumente []string) int {
	if len(argumente) == 1 {
		text := flaeche.SprachpaketAusgabeText(ctx, argumente[0])
		if text.Fehler != nil {
			return klageUeberNaht(an, text.Fehler)
		}

		an.sag("%s", text.Inhalt)

		return an.abschluss()
	}

	antwort := flaeche.SprachpaketAusgabeSchreiben(ctx, argumente[0], argumente[1])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	an.sag("Geschrieben nach %s.\n", antwort.Pfad)

	return an.abschluss()
}

// sprachantwort gibt eine Übersetzung aus — ganz als JSON, wenn es
// verlangt ist oder `ganz` es will, sonst als eine Zeile.
func sprachantwort(an *schreiber, antwort app.SprachpaketAntwort, ganz bool) int {
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON || ganz {
		return alsJSONSchreiben(an, antwort)
	}

	paket := antwort.Paket
	an.sag("%s\t%d/%d übersetzt, %d veraltet\n", paket.Kennung, paket.Uebersetzt, paket.Gesamt, paket.Veraltet)

	return an.abschluss()
}
