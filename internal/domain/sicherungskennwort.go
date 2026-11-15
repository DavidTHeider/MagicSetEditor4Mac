package domain

// **Das Kennwort einer Sicherung** — OPS-06.AK-02
// (#1723).
//
// # Warum zwölf Zeichen und keine Zeichenklassen
//
// **Eine Sicherung außer Haus wird offline angegriffen**: wer die Datei hat,
// probiert, so schnell die Schlüsselableitung es zulässt, und niemand zählt
// die Versuche. Was dagegen hilft, ist Länge. Vorschriften über Ziffern und
// Sonderzeichen treiben zu kurzen, schwer zu merkenden Kennwörtern — ein Satz
// aus vier Wörtern ist länger und leichter zu behalten.
//
// # Und warum es keinen Rückweg gibt
//
// **Das Kennwort liegt nirgends**, auch nicht im Schlüsselbund: eine Sicherung
// soll auf einem anderen Rechner aufgehen. Wer es vergisst, hat die Sicherung
// verloren — das sagt die Oberfläche vor dem Sichern, und
// `docs/SICHERUNG.md` sagt es auch.

import (
	"strings"
	"unicode/utf8"
)

// SicherungskennwortMindestens ist die kleinste Zahl von Zeichen.
const SicherungskennwortMindestens = 12

// PruefeSicherungskennwort urteilt über ein Kennwort für eine Sicherung.
func PruefeSicherungskennwort(kennwort string) error {
	if strings.TrimSpace(kennwort) == "" || utf8.RuneCountInString(kennwort) < SicherungskennwortMindestens {
		return NewFieldError("kennwort", SchluesselSicherungskennwortKurz,
			"Ein Kennwort für eine Sicherung hat mindestens zwölf Zeichen — ein Satz aus vier Wörtern ist gut.")
	}

	return nil
}
