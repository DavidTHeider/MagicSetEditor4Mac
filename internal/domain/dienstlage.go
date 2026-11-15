package domain

import "fmt"

// **Die Bereitschaft einer Fähigkeit ist gerechnet** — Werkstück A1 von
// Phase 5.3 (#1440).
//
// # Warum die Regel hier steht und nicht in `app`
//
// Die Tatsachen sammelt `internal/app` — es kennt die Einstellungen, den
// Schlüsselbund und den Riegel. **Welche Tatsache gewinnt**, wenn mehrere
// gegen eine Fähigkeit stehen, ist eine fachliche Entscheidung und gehört
// hierher (§2): ein abgeschalteter Dienst braucht keinen Schlüssel, ein
// fehlender Schlüssel keinen Riegel. Wer den Status liest, soll das **erste**
// Hindernis sehen — das, das er zuerst beheben kann.

// Die Zustände einer Fähigkeit, in der Reihenfolge, in der sie gewinnen.
const (
	// StatusAus — der Schalter des Dienstes steht auf »aus«.
	StatusAus = "aus"
	// StatusZugangFehlt — der Dienst ist an, aber eine Zugangsangabe liegt nicht im Schlüsselbund.
	StatusZugangFehlt = "zugangFehlt"
	// StatusGestoert — die Bereitschaft ließ sich nicht feststellen; der Grund sagt, warum.
	StatusGestoert = "gestoert"
	// StatusGesperrt — der wirksame Akteur darf nicht nach draußen ([0078]).
	StatusGesperrt = "gesperrt"
	// StatusNichtGewaehlt — die Frage hat eine Wahl, und sie fiel auf einen anderen Dienst.
	StatusNichtGewaehlt = "nichtGewaehlt"
	// StatusBereit — nichts steht dagegen.
	StatusBereit = "bereit"
)

// Dienstlage sind die Tatsachen über eine Fähigkeit, wie `app` sie gesammelt hat.
type Dienstlage struct {
	// An sagt, ob der Schalter des Dienstes auf »an« steht.
	An bool
	// FehlendeAngabe ist die Zugangsangabe, die im Schlüsselbund fehlt — oder leer.
	FehlendeAngabe string
	// Stoerung ist ein Fehler beim Feststellen — etwa ein Schlüsselbund, der nicht antwortet.
	Stoerung error
	// Sperre ist das Nein des Riegels — oder nil.
	Sperre error
	// HatWahl sagt, ob die Frage mehr als ein Dienst beantworten kann.
	HatWahl bool
	// Gewaehlt sagt, ob die Wahl auf diesen Dienst fiel — nur bei HatWahl von Belang.
	Gewaehlt bool
}

// Status gibt den Zustand — das erste Hindernis, oder »bereit«.
func (l Dienstlage) Status() string {
	switch {
	case !l.An:
		return StatusAus
	case l.FehlendeAngabe != "":
		return StatusZugangFehlt
	case l.Stoerung != nil:
		return StatusGestoert
	case l.Sperre != nil:
		return StatusGesperrt
	case l.HatWahl && !l.Gewaehlt:
		return StatusNichtGewaehlt
	default:
		return StatusBereit
	}
}

// Grund sagt in einem Satz, was dem Status zugrunde liegt — leer bei »bereit«.
func (l Dienstlage) Grund() string {
	switch l.Status() {
	case StatusAus:
		return "der Dienst ist abgeschaltet"
	case StatusZugangFehlt:
		return fmt.Sprintf("die Angabe %q fehlt im Schlüsselbund", l.FehlendeAngabe)
	case StatusGestoert:
		return l.Stoerung.Error()
	case StatusGesperrt:
		return l.Sperre.Error()
	case StatusNichtGewaehlt:
		return "für diese Frage ist ein anderer Dienst gewählt"
	default:
		return ""
	}
}
