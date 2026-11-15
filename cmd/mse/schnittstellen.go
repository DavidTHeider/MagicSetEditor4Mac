package main

// **`mse schnittstellen`** — OPS-09.AK-03
// (#1907):
// der Stand von Kommandozeile und Werkzeugserver, rein lesend. Eingerichtet
// wird am Fenster — die Kommandozeile richtet sich nicht selbst ein.
//
// **`mse schnittstellen widerrufen [ziel]`** — OPS03
// (#2349):
// die Außerbetriebnahme beginnt hier, auch ohne Fenster. Genommen wird nur,
// was diese Anwendung selbst eintrug; ein fremder Eintrag bleibt und wird
// genannt. Wer es ein zweites Mal ruft, ändert nichts und bekommt keinen
// Fehler.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func schnittstellenbefehl(ctx context.Context, an *schreiber, rest []string) int {
	switch {
	case len(rest) == 0:
		return schnittstellenstand(ctx, an, app.NeueApp("").Schnittstellen(ctx))
	case rest[0] == "widerrufen" && len(rest) == 1:
		return schnittstellenAlleWiderrufen(ctx, an)
	case rest[0] == "widerrufen" && len(rest) == 2:
		return schnittstellenstand(ctx, an, app.NeueApp("").SchnittstelleWiderrufen(ctx, rest[1]))
	}

	an.klage("schnittstellen nimmt nichts oder »widerrufen [ziel]« — eingerichtet wird am Fenster.\n%s", hilfe)

	return CodeBedienung
}

// schnittstellenstand schreibt den Stand der Ziele.
func schnittstellenstand(_ context.Context, an *schreiber, antwort app.SchnittstellenAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	for _, eine := range antwort.Schnittstellen {
		an.sag("%s\t%s\t%s\n", eine.Name, eine.Stand, eine.Datei)
	}

	return an.abschluss()
}

// schnittstellenAlleWiderrufen nimmt jedes eigene Ziel und nennt je Ziel,
// was ihm geschah — ein Teilfehler nimmt die übrigen nicht mit. **Mit
// `--json` steht bei einem Fehler nur seine Form da** (F03, #2342), wie an
// jedem anderen Befehl; je Ziel nennt es die Textausgabe.
func schnittstellenAlleWiderrufen(ctx context.Context, an *schreiber) int {
	antwort := app.NeueApp("").SchnittstellenAlleWiderrufen(ctx)

	if !an.alsJSON {
		for _, ausgang := range antwort.Ausgaenge {
			an.sag("%s\t%s\n", ausgang.Name, ausgang.Ausgang)
		}
	}

	if antwort.Fehler != nil {
		an.klageUeber(antwort.Fehler.AlsFehler())

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	return an.abschluss()
}
