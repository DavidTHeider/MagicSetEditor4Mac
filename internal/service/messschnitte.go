package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Womit [textmass] setzt** — die Abschnitte und die Schnitte, in denen
// gezeichnet wird. Eine eigene Datei, weil `inhaltsmass.go` damit über die
// Zeilengrenze aus §2 ging; der Schnitt folgt der Frage: dort, **was**
// gemessen wird, hier, **in welcher Schrift**.

// messabschnitte macht aus rohen Abschnitten die des Satzes — mit Gewicht und
// Lage der Schrift auf jedem, wie [textebene] sie legt.
//
// **Mit Symbolen, über denselben Weg wie die Zeichnung** ([mitSymbolen]).
// Hier stand „ohne Symbole: ihre Stellen zählen als Zeichen" — und MSE2 setzt
// `content_*` und `layout` mit demselben `TextViewer`, der zeichnet
// (`TextViewer::prepareLines`, `src/render/text/viewer.cpp`, Zeilen 392–426,
// Fassung f9d9356 vom 2021-07-06). `{3}{U}{U}` maß als drei schmale
// Buchstaben und zeichnete als drei runde Symbole; „Inspired Idea" fehlte
// dadurch in der Messung eine Zeile (F02, #2244). Die Befunde der Symbole
// meldet die Zeichnung, nicht die Messung.
func messabschnitte(
	ctx context.Context, feld *msepaket.Feldstil, roh []rohabschnitt, rahmen render.Rechteck,
	vorlage Vorlage, umgebung *msescript.Umgebung,
) []render.Abschnitt {
	abschnitte, _ := mitSymbolen(ctx, feld, rahmen, roh, umgebung, vorlage)

	gewicht, _ := textangabe(ctx, feld.Feld, "font weight", feld.Schrift.Gewicht, umgebung)
	lage, _ := textangabe(ctx, feld.Feld, "font style", feld.Schrift.Lage, umgebung)

	return mitSchriftschaltern(abschnitte, gewicht == "bold", lage == "italic")
}

// messschnitte sucht die Schnitte, in denen gezeichnet wird — fett, kursiv und
// beides, so wie der Rasterer sie sucht. **Was fehlt, bleibt leer**, und der
// Satz fällt auf den aufrechten zurück, wie beim Zeichnen.
//
// **Bis F02 wurde nur aufrecht gemessen**
// (#2244) —
// ein fettes oder kursives Feld meldete ein `content_width`, das schmaler war
// als das, was dastand.
func messschnitte(
	ctx context.Context, feld *msepaket.Feldstil, aufrecht render.Schrift, name string,
	schriften render.Schriftquelle, umgebung *msescript.Umgebung,
) render.Schriftsatz {
	satz := render.Schriftsatz{Aufrecht: aufrecht}

	if fett, err := schriften.Suche(name, true); err == nil {
		satz.Fett = fett
	}

	kursivname := schriftnameAus(ctx, feld.Feld, "italic name", feld.Schrift.KursivName, umgebung)
	satz.Kursiv, satz.KursivFett = render.Kursivschnitte(schriften, name, kursivname)

	return satz
}
