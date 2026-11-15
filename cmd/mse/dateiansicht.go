// **Eine Datei ansehen** — UX-13
// (#1685).
//
// **Ohne Bestand**: die Ansicht fragt die Platte und nicht die Datenbank. Und
// **ohne Bilddaten** — wer hier fragt, hat die Datei selbst.

package main

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// dateiansehenbefehl zeigt, was in einer Datei liegt.
func dateiansehenbefehl(an *schreiber, rest []string) int {
	if len(rest) != 1 {
		an.klage("datei-ansehen braucht genau eine Datei.\n%s", hilfe)

		return CodeBedienung
	}

	ansicht := app.DateiansichtAus(rest[0])
	if ansicht.Fehler != nil {
		an.klageUeberFehler(ansicht.Fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, ansicht)
	}

	an.sag("name\t%s\nart\t%s\nvorschau\t%s\ngroesse\t%d\ngeaendert\t%s\n",
		ansicht.Name, ansicht.Art, ansicht.Vorschau, ansicht.Groesse, ansicht.Geaendert)

	for merkmal, gesetzt := range map[string]bool{
		"gekuerzt": ansicht.Gekuerzt, "verschluesselt": ansicht.Verschluesselt,
		"zu gross": ansicht.ZuGross, "unlesbar": ansicht.Unlesbar,
	} {
		if gesetzt {
			an.sag("hinweis\t%s\n", merkmal)
		}
	}

	for _, eintrag := range ansicht.Eintraege {
		zusatz := ""
		if eintrag.Unsicher {
			zusatz = "\tunsicher"
		}

		an.sag("eintrag\t%s\t%d%s\n", eintrag.Name, eintrag.Groesse, zusatz)
	}

	if ansicht.Text != "" {
		an.sag("\n%s\n", ansicht.Text)
	}

	return an.abschluss()
}
