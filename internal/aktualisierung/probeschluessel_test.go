// **Ein echter Schlüssel für die Vorrichtung** —
// #892.
//
// # Warum ein eigener und nicht der des Herausgebers
//
// Weil der private Schlüssel nicht auf jedem Rechner liegt, auf dem diese
// Tests laufen sollen — und weil ein Test, der `gpg` aufruft, prüft, ob `gpg`
// da ist.
//
// **Was er nicht ist: eine nachgiebigere Prüfung.** Hier wird wirklich
// signiert und wirklich geprüft; getauscht ist nur, **wessen** Schlüssel gilt.
// Dass der **eingebettete** der richtige ist, hält `signatur_test.go` eigens —
// und das ist die Hälfte, die eine Vorrichtung nicht leisten kann (§3).
package aktualisierung_test

import (
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
)

// probeschluessel setzt einen frischen Schlüssel ein und gibt den Signierer.
//
// **Je Test einer.** Ein gemeinsamer wäre Zustand zwischen Tests, und der
// Aufwand ist eine Schlüsselerzeugung.
func probeschluessel(t *testing.T) func(text string) string {
	t.Helper()

	wer, err := openpgp.NewEntity("Probe", "für Tests", "probe@beispiel.test", nil)
	if err != nil {
		t.Fatalf("Schlüssel erzeugen: %v", err)
	}

	vorher := aktualisierung.MitFreigabebund(openpgp.EntityList{wer})
	t.Cleanup(func() { aktualisierung.MitFreigabebund(vorher) })

	return func(text string) string {
		var gepanzert strings.Builder

		if fehler := openpgp.ArmoredDetachSign(
			&gepanzert, wer, strings.NewReader(text), nil,
		); fehler != nil {
			t.Fatalf("signieren: %v", fehler)
		}

		return gepanzert.String()
	}
}
