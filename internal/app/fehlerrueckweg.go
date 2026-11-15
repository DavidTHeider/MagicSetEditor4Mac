package app

import (
	"errors"
)

// AlsFehler macht aus einem übersetzten Befund wieder einen `error`.
//
// # Warum nicht `Error()` auf [Fehler]
//
// Weil `*Fehler` dann `error` erfüllte — und ein `nil *Fehler` in einer
// `error`-Schnittstelle ist **nicht** nil. Genau die Falle, die jeder Go-Code
// einmal tritt: `if err != nil` wäre wahr, obwohl nichts schiefging.
//
// **Der Code reist mit**, damit `errors.Is` weiter greift: eine Fläche ohne
// Fenster soll `ErrNotFound` von `ErrValidation` unterscheiden können, ohne
// die Meldung zu lesen (§8).
func AlsFehler(befund *Fehler) error {
	if befund == nil {
		return nil
	}

	// **Der Schlüssel geht mit, wenn die Meldung nichts sagt.** `Meldung` ist
	// der Rückfall für einen Schlüssel, den der Katalog nicht kennt — und
	// dieser Rückfall lautet „Es ist ein unerwarteter Fehler aufgetreten".
	// Ohne den Schlüssel daneben steht ein Mensch an der Kommandozeile vor
	// einem Satz, der nichts benennt.
	satz := befund.Meldung
	if befund.Schluessel != "" {
		satz += " (" + string(befund.Schluessel) + ")"
	}

	// **Dieselbe Tabelle wie [Fehler.AlsFehler]** und derselbe Fehlertyp
	// ([nahtfehler]).
	if klasse, bekannt := arten[befund.Code]; bekannt {
		return nahtfehler{satz: satz, klasse: klasse}
	}

	return errors.New(satz)
}
