// **Die Symbolbibliothek an der Kommandozeile** — DOM-02
// (#1697).
//
// Einordnung nach §16: **erweiternd** — ein Befehl mit zwölf Unterbefehlen
// kommt hinzu. **Ausgeben ohne Ziel schreibt den Text auf die Ausgabe**,
// damit er sich weiterreichen lässt — wie bei `turnier ausgeben`.

package main

import (
	"context"
	"database/sql"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// symbolbibliothekbefehl verteilt auf die Unterbefehle.
func symbolbibliothekbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		flaeche, err := app.SymbolbibliotheksflaecheUeber(db, domain.UrheberKommandozeile)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if len(rest) == 0 {
			return bibliothekZeigen(ctx, flaeche, an, "")
		}

		wort, argumente := rest[0], rest[1:]

		switch wort {
		case "liste":
			return bibliothekZeigen(ctx, flaeche, an, strings.Join(argumente, " "))

		case "loeschen":
			return bibliothekssymbolLoeschen(ctx, flaeche, an, argumente)

		case "ausgeben":
			return bibliothekssymbolAusgeben(ctx, flaeche, an, argumente)

		case "einsetzen":
			return bibliothekssymbolEinsetzen(ctx, flaeche, an, argumente)

		case "teile":
			return bibliothekssymbolTeile(ctx, flaeche, an, argumente)

		case "sichern":
			return bibliothekssymbolSichern(ctx, flaeche, an, argumente)

		case "zuruecknehmen":
			return bibliothekssymbolZuruecknehmen(ctx, flaeche, an, argumente)

		case "nachziehen":
			return bibliothekssymbolNachziehen(ctx, flaeche, an, argumente)
		}

		if weg, bekannt := bibliothekswege(flaeche)[wort]; bekannt {
			return bibliothekssymbolAngelegt(ctx, an, wort, argumente, weg)
		}

		an.klage("symbolbibliothek kennt %q nicht.\n%s", wort, hilfe)

		return CodeBedienung
	})
}

// bibliothekswege sind die Unterbefehle, die ein Symbol nennen und seine
// Kennung ausgeben — je mit zwei Angaben, außer beim Einlesen.
func bibliothekswege(
	flaeche *app.Symbolbibliotheksflaeche,
) map[string]func(ctx context.Context, erstes, zweites string) app.BibliothekssymbolAntwort {
	return map[string]func(ctx context.Context, erstes, zweites string) app.BibliothekssymbolAntwort{
		"aufnehmen":   flaeche.BibliothekssymbolAufnehmen,
		"duplizieren": flaeche.BibliothekssymbolKopieren,
		"umbenennen":  flaeche.BibliothekssymbolUmbenennen,
		"einlesen":    flaeche.BibliothekssymbolEinlesenAus,
	}
}

// bibliothekssymbolAngelegt führt einen der [bibliothekswege] aus.
func bibliothekssymbolAngelegt(
	ctx context.Context, an *schreiber, wort string, argumente []string,
	weg func(ctx context.Context, erstes, zweites string) app.BibliothekssymbolAntwort,
) int {
	// **Einlesen braucht nur die Datei** — ohne Namen heißt das Symbol wie
	// sie.
	noetig := 2
	if wort == "einlesen" {
		noetig = 1
	}

	if len(argumente) < noetig {
		an.klage("symbolbibliothek %s braucht %d Angaben.\n%s", wort, noetig, hilfe)

		return CodeBedienung
	}

	argumente = append(slices.Clone(argumente), "")

	antwort := weg(ctx, argumente[0], argumente[1])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("%s\n", antwort.ID)

	return an.abschluss()
}

// bibliothekZeigen nennt die Symbole — je Zeile Kennung, Name und Träger.
func bibliothekZeigen(ctx context.Context, flaeche *app.Symbolbibliotheksflaeche, an *schreiber, suche string) int {
	antwort := flaeche.Symbolbibliothek(ctx, suche)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort.Symbole)
	}

	switch {
	case len(antwort.Symbole) == 0 && suche == "":
		an.sag("Die Bibliothek ist leer — »symbolbibliothek aufnehmen« nimmt das Symbol eines Sets auf.\n")

	case len(antwort.Symbole) == 0:
		an.sag("Kein Symbol heißt so.\n")
	}

	for _, symbol := range antwort.Symbole {
		traeger := "kein Set"
		if len(symbol.Verwendet) > 0 {
			traeger = strings.Join(symbol.Verwendet, ", ")
		}

		an.sag("%s\t%s\t%s\n", symbol.ID, symbol.Name, traeger)
	}

	return an.abschluss()
}

// bibliothekssymbolLoeschen löscht — **ein verwendetes nur mit
// `--bestaetigt`**, und die Absage nennt die Sets.
func bibliothekssymbolLoeschen(
	ctx context.Context, flaeche *app.Symbolbibliotheksflaeche, an *schreiber, argumente []string,
) int {
	bestaetigt := slices.Contains(argumente, bestaetigtfahne)
	argumente = slices.DeleteFunc(slices.Clone(argumente), func(wort string) bool { return wort == bestaetigtfahne })

	if len(argumente) < 1 {
		an.klage("symbolbibliothek loeschen braucht die Kennung eines Symbols.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.BibliothekssymbolLoeschen(ctx, argumente[0], bestaetigt)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	if len(antwort.Sets) > 0 {
		an.sag("Gelöscht. Ihre Kopie behalten: %s\n", strings.Join(antwort.Sets, ", "))
	} else {
		an.sag("Gelöscht.\n")
	}

	return an.abschluss()
}

// bibliothekssymbolAusgeben schreibt ein Symbol in eine Datei — **ohne Ziel
// als Text auf die Ausgabe**.
func bibliothekssymbolAusgeben(
	ctx context.Context, flaeche *app.Symbolbibliotheksflaeche, an *schreiber, argumente []string,
) int {
	if len(argumente) < 1 {
		an.klage("symbolbibliothek ausgeben braucht die Kennung eines Symbols.\n%s", hilfe)

		return CodeBedienung
	}

	if len(argumente) == 1 {
		text := flaeche.BibliothekssymbolText(ctx, argumente[0])
		if text.Fehler != nil {
			return klageUeberNaht(an, text.Fehler)
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, text)
		}

		an.sag("%s", text.Inhalt)

		return an.abschluss()
	}

	antwort := flaeche.BibliothekssymbolAusgabeSchreiben(ctx, argumente[0], argumente[1])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("Geschrieben nach %s.\n", antwort.Pfad)

	return an.abschluss()
}

// bibliothekssymbolEinsetzen macht ein Symbol zu dem eines Sets.
func bibliothekssymbolEinsetzen(
	ctx context.Context, flaeche *app.Symbolbibliotheksflaeche, an *schreiber, argumente []string,
) int {
	if len(argumente) < 2 {
		an.klage("symbolbibliothek einsetzen braucht ein Symbol und ein Set.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.BibliothekssymbolEinsetzen(ctx, argumente[0], argumente[1])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("%s\n", antwort.Datei)

	return an.abschluss()
}
