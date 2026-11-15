package app

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Die Güte kommt als Name über die Flächen** — Werkstück Ω.4 aus
// #1029,
// Vermerk [0049](../../docs/entscheidungen/0049-die-guete-ist-eine-wahl.md).
//
// Über MCP, REST und die Kommandozeile reist eine Zeichenkette; `render.Guete`
// ist ein `int`. Die Übersetzung steht hier, an **einer** Stelle — dieselbe
// Naht wie bei jeder anderen (§2).

// GueteVorgabe ist der Name, den eine Fläche einsetzt, wenn niemand wählt.
//
// **Leer heißt Vorgabe und ist kein Fehler.** Ein Aufruf ohne `guete` ist der
// Normalfall und soll die Parität bekommen; abgewiesen wird nur ein Name, den
// es nicht gibt.
const GueteVorgabe = "medium"

// GueteAusName gibt die Güte zu ihrem Namen — leer heißt Vorgabe.
func GueteAusName(name string) (render.Guete, error) {
	if strings.TrimSpace(name) == "" {
		return render.GueteMittel, nil
	}

	guete, bekannt := render.GueteNach(name)
	if !bekannt {
		return render.GueteMittel, unbekannteGuete(name)
	}

	return guete, nil
}

// unbekannteGuete baut den Fehler und nennt, was es gibt.
//
// **Mit der Aufzählung im Text.** Ein „unbekannte Güte" ohne die vier Namen
// schickt jeden in die Dokumentation; die Namen stehen in `render.Gueten`, und
// von dort kommen sie auch hierher — eine zweite Liste liefe auseinander.
func unbekannteGuete(name string) error {
	namen := make([]string, 0, len(render.Gueten))
	for _, guete := range render.Gueten {
		namen = append(namen, guete.Name())
	}

	// **Als Feldfehler und nicht als schlichter Validierungsfehler.** Ein
	// `fmt.Errorf` mit `domain.ErrValidation` kommt im Fenster als „Es ist ein
	// unerwarteter Fehler aufgetreten" an — dieselbe Auskunft, die die
	// Kommandozeile im Klartext gibt, wäre dort verloren. Genau die
	// Ungleichheit, gegen die §16 steht.
	return domain.NewFieldError("guete", domain.SchluesselValidierung,
		"%q ist keine Güte — erwartet ist eine von %s.",
		name, strings.Join(namen, ", "))
}
