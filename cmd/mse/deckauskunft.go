// **Prüfung und Abgleich an der Kommandozeile** — Werkstück 7 von Phase 4.6
// (#1293).

package main

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// deckPruefen beurteilt ein Deck gegen sein Format.
//
// **Der Rückgabewert sagt das Urteil**: ein unzulässiges Deck ist ein
// **fachliches Nein** und kein Fehler. So lässt es sich in einem Skript
// benutzen, ohne die Ausgabe zu lesen.
func deckPruefen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("deck pruefen braucht genau eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.DeckPruefen(ctx, rest[0])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		if code := alsJSONSchreiben(an, antwort); code != CodeGut {
			return code
		}

		return urteilscode(antwort.Urteil.Zulaessig)
	}

	an.sag("%s — %d im Hauptdeck, %d im Sideboard\n",
		antwort.Urteil.Formatname, antwort.Urteil.Hauptdeck, antwort.Urteil.Sideboard)

	for _, befund := range antwort.Urteil.Groessen {
		an.sag("! %s\n", befund.Meldung)
	}

	for _, urteil := range antwort.Urteil.Urteile {
		// **Nur, was nicht schlicht zulässig ist.** Sechzig Zeilen `legal`
		// verdecken die eine, auf die es ankommt.
		if urteil.Zulaessigkeit == "legal" {
			continue
		}

		an.sag("%s\t%d×\t%s\t%s\t%s\n", urteil.Zulaessigkeit, urteil.Menge,
			urteil.Bereich, urteil.Bezeichnung, urteil.Quelle)
	}

	for _, befund := range antwort.Befunde {
		an.sag("? %s\n", befund)
	}

	if code := an.abschluss(); code != CodeGut {
		return code
	}

	return urteilscode(antwort.Urteil.Zulaessig)
}

// deckAbgleichen sagt, was von einem Deck in der Sammlung liegt.
//
// **Auch hier sagt der Rückgabewert das Ergebnis**: fehlt etwas, ist es ein
// fachliches Nein.
func deckAbgleichen(
	ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("deck bestand braucht genau eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.DeckAbgleichen(ctx, rest[0])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		if code := alsJSONSchreiben(an, antwort); code != CodeGut {
			return code
		}

		return urteilscode(antwort.Fehlend == 0)
	}

	for _, zeile := range antwort.Zeilen {
		if zeile.Printing == "" {
			an.sag("?\t%d×\t%s\n", zeile.Gebraucht, zeile.Grund)

			continue
		}

		an.sag("%d/%d\t%s\t%s\n",
			zeile.Vorhanden, zeile.Gebraucht, zeile.Printing, zeile.Bezeichnung)
	}

	an.sag("fehlend %d, unbekannt %d\n", antwort.Fehlend, antwort.Unbekannt)

	if code := an.abschluss(); code != CodeGut {
		return code
	}

	return urteilscode(antwort.Fehlend == 0)
}

// urteilscode macht aus einem Urteil einen Rückgabewert.
//
// **`1` und nicht `2`**: die Angaben waren richtig, die Antwort ist nein.
func urteilscode(gut bool) int {
	if gut {
		return CodeGut
	}

	return CodeFachlich
}
