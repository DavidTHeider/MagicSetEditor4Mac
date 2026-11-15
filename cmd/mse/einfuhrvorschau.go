// **Eine Einfuhr vorab ansehen, an der Kommandozeile** — DOM-03.AK-01
// (#1698).
//
// Einordnung nach §16: **erweiternd** — zwei Befehle kommen hinzu, und
// `sammlung-einlesen` und `liste-einlesen` nehmen ein freiwilliges
// `--auslassen`. Ohne das Wort wird addiert wie bisher.

package main

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// auslassenwort ist das freiwillige Wort der zwei Einlesebefehle.
const auslassenwort = "--auslassen"

// zusammenfuehrungAus nimmt `--auslassen` aus den Argumenten und sagt, was
// mit einer vorhandenen Zeile geschieht.
func zusammenfuehrungAus(rest []string) (uebrig []string, zusammenfuehrung string) {
	if !slices.Contains(rest, auslassenwort) {
		return rest, "addieren"
	}

	return slices.DeleteFunc(slices.Clone(rest), func(wort string) bool {
		return wort == auslassenwort
	}), "auslassen"
}

// sammlungVorschaubefehl sagt, was `sammlung-einlesen` mit einem CSV von der
// Standardeingabe täte — und schreibt nichts.
func sammlungVorschaubefehl(ctx context.Context, bestandsort string, an *schreiber) int {
	text, fehler := io.ReadAll(an.eingabe)
	if fehler != nil {
		an.klageUeber(fmt.Errorf("die Sammlung lesen: %w", fehler))

		return CodeBedienung
	}

	return mitAusfuhr(ctx, bestandsort, an, func(ctx context.Context, naht *app.Sammlungsausfuhrnaht) int {
		vorschau, err := naht.CSVVorschauen(ctx, string(text))
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return sammelvorschauZeigen(an, vorschau)
	})
}

// listeVorschaubefehl sagt, was `liste-einlesen` mit einer Liste von der
// Standardeingabe täte — dieselben eindeutigen Zeilen, und nichts wird
// geschrieben.
func listeVorschaubefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	text, fehler := io.ReadAll(an.eingabe)
	if fehler != nil {
		an.klageUeber(fmt.Errorf("die Liste lesen: %w", fehler))

		return CodeBedienung
	}

	zustand := domain.ZustandNearMint
	if len(rest) > 0 {
		zustand = domain.Zustand(rest[0])
	}

	return mitEinfuhr(ctx, bestandsort, an, func(ctx context.Context, naht *app.Listeneinfuhrnaht) int {
		deutung, err := naht.Deuten(ctx, string(text))
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		zeilen, _ := eindeutigeZeilen(deutung, zustand)

		vorschau, err := naht.Vorschauen(ctx, zeilen)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return sammelvorschauZeigen(an, vorschau)
	})
}

// sammelvorschauZeigen gibt eine Vorschau aus — eine Zeile je Posten.
func sammelvorschauZeigen(an *schreiber, vorschau *app.Sammelvorschau) int {
	if an.alsJSON {
		return alsJSONSchreiben(an, vorschau)
	}

	for _, posten := range vorschau.Posten {
		an.sag("%s\tZeile %d\t%d×\t%s\t%d → %d\n", posten.Art, posten.Zeile, posten.Menge,
			posten.Bezeichnung, posten.Vorher, posten.Nachher)
	}

	for _, befund := range vorschau.Befunde {
		an.sag("!\t%s\n", befund)
	}

	return an.abschluss()
}

// ausgelassenText nennt die ausgelassenen Zeilen — **nur, wenn es welche
// gibt**: ohne `--auslassen` bleibt die Ausgabe, wie sie war.
func ausgelassenText(bericht *app.Einfuhrbericht) string {
	if bericht.Ausgelassen == 0 {
		return ""
	}

	return fmt.Sprintf(", %d ausgelassen", bericht.Ausgelassen)
}
