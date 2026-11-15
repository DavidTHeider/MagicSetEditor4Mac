package app

import (
	"errors"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// prozessVeraltet übersetzt einen Prozess, unter dem eine neuere Fassung
// gewandert ist — oder gibt nichts (M02, #2346). **Vor dem belegten
// Bestand**, dessen Klasse er trägt: dessen Satz riete zum Warten, und hier
// kommt nach dem Warten ein anderer Prozess.
func prozessVeraltet(fehler error) *Fehler {
	var veraltet *domain.ProzessVeraltet
	if !errors.As(fehler, &veraltet) {
		return nil
	}

	uebersetzt := mitSatz(CodeBelegt, domain.SchluesselProzessVeraltet, fehler)
	uebersetzt.Werte = map[string]string{
		"unbekannt": strings.Join(veraltet.Fassung.Unbekannt, ", "),
		"stand":     veraltet.Fassung.Stand,
	}

	return uebersetzt
}
