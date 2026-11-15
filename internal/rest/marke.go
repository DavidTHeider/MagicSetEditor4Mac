package rest

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// markenlaenge ist die Zahl der Zufallsbytes einer Marke.
//
// **32**, also 64 hexadezimale Zeichen. Der Dienst drosselt, aber eine Marke,
// die sich erraten lässt, machte die Drossel zur einzigen Hürde — und dagegen
// ist sie nicht gebaut.
const markenlaenge = 32

// vorspann ist das Wort vor der Marke in der Kopfzeile.
const vorspann = "Bearer "

// NeueMarke würfelt eine Marke.
//
// **Aus `crypto/rand` und nicht aus `math/rand`.** Der zweite ist
// vorhersagbar, sobald jemand den Startwert kennt oder erraten kann; für einen
// Zugang ist das keine Feinheit.
func NeueMarke() (string, error) {
	return MarkeAus(rand.Reader)
}

// MarkeAus würfelt aus einer gegebenen Quelle.
//
// # Warum es diese Fassung gibt
//
// Damit der Fehlerweg **betreten** wird. `crypto/rand` scheitert auf einem
// gesunden Rechner nie; ohne einsetzbare Quelle stünde hier ein `if`, das kein
// Test je durchläuft — und AGENTS.md §3 nimmt genau das nicht als Ausnahme an:
// „eine unbequeme Verzweigung ist kein Grund".
//
// **Sie ist exportiert und trotzdem keine Einladung.** Wer sie mit einer
// vorhersagbaren Quelle riefe, bekäme eine vorhersagbare Marke; deshalb ruft
// `cmd/mse-rest` [NeueMarke] und nicht diese.
func MarkeAus(quelle io.Reader) (string, error) {
	roh := make([]byte, markenlaenge)
	if _, err := io.ReadFull(quelle, roh); err != nil {
		return "", fmt.Errorf("eine Marke würfeln: %w", err)
	}

	return hex.EncodeToString(roh), nil
}

// MarkeStimmt sagt, ob eine Anfrage die erwartete Marke trägt.
//
// # Warum zeitkonstant verglichen wird
//
// Ein gewöhnlicher Zeichenkettenvergleich bricht beim ersten Unterschied ab.
// Wer viele Anfragen stellt und die Zeit misst, liest daraus die Marke Zeichen
// für Zeichen ab. Das ist kein theoretischer Angriff auf einem Rechner, auf
// dem der Angreifer ohnehin schon einen Prozess hat — und genau der ist hier
// das Gegenüber, gegen das die Marke überhaupt schützt.
//
// # Was sie nicht tut
//
// **Sie kennt keinen Ablauf.** Eine Marke gilt, solange der Dienst läuft; wer
// sie zurückziehen will, beendet ihn. Ein Ablauf setzte eine Uhr und eine
// Erneuerung voraus, und beides wäre Aufwand für einen Dienst, der auf Aufruf
// startet und mit dem Aufrufer endet.
func MarkeStimmt(erwartet string, anfrage *http.Request) bool {
	gegeben, gefunden := strings.CutPrefix(
		anfrage.Header.Get("Authorization"), vorspann)
	if !gefunden {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(gegeben), []byte(erwartet)) == 1
}
