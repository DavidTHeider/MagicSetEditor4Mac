package service

// **Ein Rückweg, der scheitert, sagt es** — A05
// (#2219),
// Restbefund des Nachlaufs über Phase A.
//
// Zurückspielen, Verschlüsseln und Neuanfang drehen einen gescheiterten
// Tausch zurück. Gelingt das, steht der bisherige Stand wieder ganz da, und
// die Meldung darf das sagen; **scheitert auch der Rückweg, darf sie es
// nicht** — dann liegen Teile beiseite, und wer das liest, muss es wissen.

import (
	"errors"
	"fmt"
)

// ErrRueckwegGescheitert sagt: ein Tausch scheiterte, **und sein Rückweg
// auch** — der bisherige Stand steht nicht mehr ganz da. Was beiseite liegt,
// bleibt liegen; verloren ist es nicht.
var ErrRueckwegGescheitert = errors.New("der bisherige Stand ließ sich nicht ganz zurückholen")

// rueckweg geht einen Rückweg und kennzeichnet sein Scheitern.
func rueckweg(zurueck func() error) error {
	if fehler := zurueck(); fehler != nil {
		return fmt.Errorf("%w: %w", ErrRueckwegGescheitert, fehler)
	}

	return nil
}
