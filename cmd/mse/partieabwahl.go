package main

// **Die Abwahl beim Anfangen** — P1a, Befund C-01 aus dem Audit zur
// Produktionsreife 0.9.0
// (#2335).
//
// # Warum `--ohne=` und nicht ein drittes Argument
//
// **Weil die Sitze beliebig viele sind.** `partie-anlegen` nimmt das Spiel
// und danach jeden Sitz als eigenes Argument; ein Platz für die Abwahl
// dahinter wäre von einem Sitz nicht zu unterscheiden. **Und `ohne` heißt
// es wie am Dienst** (`GET /regelfreigabe?ohne=`): wer die Wahl dort prüft,
// schreibt sie hier genauso.
//
// # Wo geprüft wird
//
// **Hier nur die Form.** Ob es einen Bereich gibt, entscheidet
// [domain.Regelwerk.Pruefe] — an **einer** Stelle für alle vier Flächen
// (§2, §16), und eine unbekannte Abwahl kommt deshalb als derselbe
// fachliche Fehler zurück wie am Dienst.

import "strings"

// abwahlvorsatz leitet `--ohne=<bereich>[,<bereich>…]` ein.
const abwahlvorsatz = "--ohne="

// trenneAbwahl gibt die Argumente ohne den Schalter und die Bereiche darin.
//
// **Mehrere Schalter sind eine Liste**, in der Reihenfolge, in der sie
// dastanden; ein leerer Teil zwischen zwei Kommas ist kein Bereich.
func trenneAbwahl(rest []string) (uebrig, abgewaehlt []string) {
	uebrig = make([]string, 0, len(rest))

	for _, argument := range rest {
		bereiche, gefunden := strings.CutPrefix(argument, abwahlvorsatz)
		if !gefunden {
			uebrig = append(uebrig, argument)

			continue
		}

		for bereich := range strings.SplitSeq(bereiche, ",") {
			if bereich = strings.TrimSpace(bereich); bereich != "" {
				abgewaehlt = append(abgewaehlt, bereich)
			}
		}
	}

	return uebrig, abgewaehlt
}
