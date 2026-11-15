package service

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// guetenbefund gibt den Satz, mit dem eine Güte ihre Abweichung ansagt.
//
// **An der Ausgabe und nicht in der Dokumentation** — so verlangt es Werkstück
// Ω.4 aus #1029.
//
// # Warum hier und nicht im Rasterer
//
// Weil hier die Güte **gewählt** wird. `render.Guete.Hinweis` gab den Satz von
// Anfang an her, und `guete.go` behauptete, der Rasterer hänge ihn an — er tat
// es nie, und niemand rief die Methode. Angestrichen vom Zweitblick über Phase
// Ω als „Code ohne Aufrufer".
//
// # Wen er erreicht
//
// **Das Fenster.** `App.BildSchreiben` reicht die Befunde durch, und die
// Bilderseite zeigt sie.
//
// **Kommandozeile, MCP und REST ebenso.** Sie gehen durch
// `app.Zeichenwerk.SchreibeBild`, und dessen `Bildschreibung` trägt die
// Befunde seit dem 2026-08-25 als Sätze
// (#1137).
//
// # Zwei verschiedene Sätze
//
// Bei SVG **wirkt** die Güte gar nicht: der Weg zweigt vor dem Rasterer ab, und
// eine SVG entsteht aus dem Zeichenplan. Die Oberfläche bietet die Wahl
// trotzdem daneben an; sie stillschweigend fallen zu lassen wäre schlimmer als
// sie zu erklären.
//
// # Die Vorgabe schweigt
//
// `medium` weicht nicht ab. Ein Befund bei jedem gewöhnlichen Bild wäre Lärm,
// und Lärm liest niemand.
func guetenbefund(guete render.Guete, format Bildformat, datei string) []domain.Befund {
	if guete == render.GueteMittel {
		return nil
	}

	meldung := guete.Hinweis()
	if format == FormatSVG {
		meldung = "die Güte " + guete.Name() + " wirkt bei SVG nicht: " +
			"eine SVG entsteht aus dem Zeichenplan und nicht aus Bildpunkten"
	}

	return []domain.Befund{{
		Ort:     datei,
		Meldung: meldung,

		// **Die Vorlage kann nichts dafür, und dieser Rechner auch nicht.** Es
		// ist eine Wahl, und die mildeste vorhandene Art ist die, die sagt:
		// hier ist etwas anders als bei MSE.
		Art: domain.ArtVorlage,
	}}
}
