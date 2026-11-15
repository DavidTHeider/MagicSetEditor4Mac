// **Wie viel Hilfe das Programm gibt** — PLAY-02.AK-01
// (#1710):
// „‚Anfänger' bezeichnet einen Bedien- und Unterstützungsmodus, nicht
// vermeintlich vollständige CR-Konformität."
//
// # Warum das kein Regelprofil ist
//
// **Weil ein Profil sagt, welche Regeln gelten, und dieser Modus, wie viel
// davon das Programm übernimmt.** Zwei Runden im selben Format können
// verschieden viel Hilfe wollen — und beide spielen nach CR 100.
//
// **Ein „Anfängerprofil" wäre deshalb eine Falschaussage**: es sähe aus, als
// gälten für Anfänger andere Spielregeln. Es gelten dieselben; was sich
// ändert, ist, wie viel das Fenster erklärt und wie viel die Runde selbst
// macht.
//
// # Und warum der Anfängermodus keine Abwahl zulässt
//
// **Weil er der Modus ist, in dem das Programm so viel trägt, wie es kann.**
// Wer etwas selbst machen will, ist damit nicht mehr in diesem Modus — und
// das ist keine Sperre, sondern ein Satz: der Weg hinaus steht in der
// Ablehnung.

package domain

import "fmt"

// Bedienmodus ist, wie viel Hilfe das Fenster gibt.
type Bedienmodus = string

// Die zwei Modi.
const (
	// ModusAnfaenger — das Programm trägt, was es tragen kann.
	//
	// **Und sagt dabei, was es nicht trägt**: die Grenzen der Matrix stehen
	// auch hier, denn ein Modus, der die Lücken verschwiege, wäre die
	// Vollständigkeitszusage, die AK-01 ausschließt.
	ModusAnfaenger Bedienmodus = "anfaenger"

	// ModusFrei — die Runde entscheidet je Bereich.
	ModusFrei Bedienmodus = "frei"
)

// Bedienmodi gibt beide.
func Bedienmodi() []Bedienmodus {
	return []Bedienmodus{ModusAnfaenger, ModusFrei}
}

// IstBedienmodus sagt, ob ein Wert einer davon ist.
func IstBedienmodus(modus Bedienmodus) bool {
	return modus == ModusAnfaenger || modus == ModusFrei
}

// PasstZumModus sagt, ob diese Wahl in diesem Modus zulässig ist.
//
// **Die Ablehnung nennt den Ausweg** (§6): „geht nicht" ohne den nächsten
// Schritt ließe jemanden vor einem Kästchen sitzen, das sich nicht anklicken
// lässt und nicht sagt, warum.
func (w Regelwahl) PasstZumModus(modus Bedienmodus) error {
	if !IstBedienmodus(modus) {
		return fmt.Errorf(
			"%w: %q ist kein Bedienmodus", ErrValidation, modus)
	}

	if modus == ModusAnfaenger && len(w.Abgewaehlt) > 0 {
		return fmt.Errorf(
			"%w: im Anfängermodus trägt das Programm, was es tragen kann — "+
				"wer %d Bereiche selbst übernehmen will, stellt auf den "+
				"freien Modus um",
			ErrValidation, len(w.Abgewaehlt))
	}

	return nil
}
