package render

import "strings"

// Symbolschrift bildet ein Zeichen aus dem Kartentext auf ein Bild ab.
//
// Sie kommt aus einem `.mse-symbol-font`-Paket (`docs/FORMATE.md`): dort steht
// je Symbol ein Zeichen und der Paketpfad seines Bildes. `render` liest das
// Paket nicht selbst — es bekommt die fertige Abbildung.
type Symbolschrift interface {
	// Bildpfad zu einem Symbolzeichen; `false`, wenn die Schrift es nicht kennt.
	Bildpfad(zeichen string) (string, bool)
}

// Textstueck ist ein Abschnitt einer Zeile: entweder Text oder ein Symbol.
type Textstueck struct {
	// Text ist der Wortlaut. Bei einem Symbol steht hier das Zeichen samt
	// Klammern, so wie es im Kartentext stand — `{T}`, nicht `T`.
	Text string

	// Bild ist der Paketpfad des Symbols; leer bei gewöhnlichem Text.
	Bild string

	// Unbekannt ist gesetzt, wenn hier eine Symbolklammer stand, die Schrift sie
	// aber nicht kennt.
	Unbekannt bool
}

// ZerlegeSymbole teilt einen Text in Stücke aus Text und Symbolen.
//
// # Ein unbekanntes Symbol bleibt sichtbar
//
// `{Q}` in einer Schrift ohne `Q` wird als `{Q}` **gezeichnet**, nicht
// verschluckt. Das ist die Zusage aus #32, und der Grund steht in §8: fremde
// Daten sind der Normalfall. Ein Kartentext, aus dem stillschweigend etwas
// verschwindet, sieht vollständig aus — und wer ihn gegen das Original hält,
// findet die Lücke erst beim Zählen.
//
// # Eine Klammer ohne Ende ist Text
//
// `{T` am Zeilenende ist keine halbe Symbolklammer, sondern eine geschweifte
// Klammer, die jemand getippt hat. Sie bleibt stehen, so wie eine offene spitze
// Klammer in `remove_tags` stehen bleibt.
func ZerlegeSymbole(text string, schrift Symbolschrift) []Textstueck {
	var stuecke []Textstueck

	rest := text
	gesammelt := strings.Builder{}

	schreibeText := func() {
		if gesammelt.Len() > 0 {
			stuecke = append(stuecke, Textstueck{Text: gesammelt.String()})
			gesammelt.Reset()
		}
	}

	for rest != "" {
		auf := strings.IndexByte(rest, '{')
		if auf < 0 {
			break
		}

		zu := strings.IndexByte(rest[auf:], '}')
		if zu < 0 {
			break
		}

		gesammelt.WriteString(rest[:auf])
		zeichen := rest[auf+1 : auf+zu]
		ganz := rest[auf : auf+zu+1]
		rest = rest[auf+zu+1:]

		schreibeText()

		if pfad, bekannt := schrift.Bildpfad(zeichen); bekannt {
			stuecke = append(stuecke, Textstueck{Text: ganz, Bild: pfad})

			continue
		}

		stuecke = append(stuecke, Textstueck{Text: ganz, Unbekannt: true})
	}

	gesammelt.WriteString(rest)
	schreibeText()

	return stuecke
}
