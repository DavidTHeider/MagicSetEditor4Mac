// **Ein Symbol der Bibliothek bearbeiten, an der Kommandozeile** — DOM-02.AK-02
// (#1697).
//
// Einordnung nach §16: **erweiternd** — vier Unterbefehle kommen hinzu. Die
// Teile gehen wie bei `symbol` und `symbol-setzen` als JSON über die
// Standardein- und -ausgabe.

package main

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// bibliothekssymbolTeile gibt die Zeichnung eines Symbols als Teile.
func bibliothekssymbolTeile(
	ctx context.Context, flaeche *app.Symbolbibliotheksflaeche, an *schreiber, argumente []string,
) int {
	if len(argumente) < 1 {
		an.klage("symbolbibliothek teile braucht die Kennung eines Symbols.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.BibliothekssymbolTeile(ctx, argumente[0])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	return alsJSONSchreiben(an, antwort)
}

// bibliothekssymbolSichern legt eine neue Fassung aus den Teilen auf der
// Standardeingabe ab.
func bibliothekssymbolSichern(
	ctx context.Context, flaeche *app.Symbolbibliotheksflaeche, an *schreiber, argumente []string,
) int {
	if len(argumente) < 1 {
		an.klage("symbolbibliothek sichern braucht die Kennung eines Symbols.\n%s", hilfe)

		return CodeBedienung
	}

	var teile []app.SymbolteilDTO
	if err := json.Unmarshal([]byte(gelesenerText(an.eingabe)), &teile); err != nil {
		an.klageUeber(app.SymbolteileUnlesbar(err))

		return CodeBedienung
	}

	return symbolaenderungMelden(an, flaeche.BibliothekssymbolSichern(ctx, argumente[0], teile))
}

// bibliothekssymbolZuruecknehmen holt die vorige Fassung zurück.
func bibliothekssymbolZuruecknehmen(
	ctx context.Context, flaeche *app.Symbolbibliotheksflaeche, an *schreiber, argumente []string,
) int {
	if len(argumente) < 1 {
		an.klage("symbolbibliothek zuruecknehmen braucht die Kennung eines Symbols.\n%s", hilfe)

		return CodeBedienung
	}

	return symbolaenderungMelden(an, flaeche.BibliothekssymbolZuruecknehmen(ctx, argumente[0]))
}

// symbolaenderungMelden sagt, welche Sets noch die vorige Fassung tragen.
func symbolaenderungMelden(an *schreiber, antwort app.SymbolaenderungAntwort) int {
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("Gesichert.\n")

	if len(antwort.Veraltet) > 0 {
		an.sag("Die vorige Fassung tragen noch: %s — »symbolbibliothek nachziehen %s« zieht sie nach.\n",
			strings.Join(antwort.Veraltet, ", "), antwort.ID)
	}

	return an.abschluss()
}

// bibliothekssymbolNachziehen legt die jetzige Fassung in die Sets, die die
// vorige tragen.
func bibliothekssymbolNachziehen(
	ctx context.Context, flaeche *app.Symbolbibliotheksflaeche, an *schreiber, argumente []string,
) int {
	if len(argumente) < 1 {
		an.klage("symbolbibliothek nachziehen braucht die Kennung eines Symbols.\n%s", hilfe)

		return CodeBedienung
	}

	antwort := flaeche.BibliothekssymbolNachziehen(ctx, argumente[0])
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("Nachgezogen: %s\n", strings.Join(antwort.Sets, ", "))

	return an.abschluss()
}
