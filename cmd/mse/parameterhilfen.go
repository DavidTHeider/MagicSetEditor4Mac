// **Die Parameterhilfen der Schlüsselwörter, an der Kommandozeile** —
// Werkstück D von Phase 4.11
// (#1342).

package main

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// schluesselwortParameter zeigt beide Menüs zu einem Schlüsselwort.
//
// **Zeigen und nicht einfügen.** Wer einfügen will, nimmt den Text und ruft
// `schluesselwort <set> setzen <id> muster <wert>` — dieselbe Trennung wie im
// Fenster, wo ein Menü einen Text in ein Feld schreibt, das schon offen ist.
func schluesselwortParameter(
	ctx context.Context, an *schreiber, pflege *app.Schluesselwortweg,
	setID int64, rest []string,
) int {
	if len(rest) == 0 {
		an.klage("schluesselwort <set> parameter braucht eine Wortkennung.")

		return CodeBedienung
	}

	wortID, gut := kennung(an, "die Wortkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	hilfen, fehler := pflege.Parameterhilfen(ctx, setID, wortID)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, hilfen)
	}

	for _, art := range hilfen.Arten {
		an.sag("einfuegen\t%s\t%s\t%s\n", art.Name, art.Einfuegen, art.Beschreibung)
	}

	for _, verweis := range hilfen.Verweise {
		an.sag("verweis\t%d\t%s\t%s\t%s\n",
			verweis.Nummer, verweis.Art, verweis.Einfuegen, verweis.Name)
	}

	for _, befund := range hilfen.Befunde {
		an.sag("!\t%s\n", befund.Meldung)
	}

	return an.abschluss()
}
