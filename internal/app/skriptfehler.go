// **Der Skriptfehler als übersetzter Befund** — die eine Übersetzung, die
// eine **Stelle** trägt.
//
// **Eine eigene Datei**, weil `fehler.go` mit ihr über 400 Zeilen ginge (§2).
// Der Schnitt folgt der Frage: dort steht, **welcher Fachfehler welchen Code
// bekommt**; hier steht der eine, der zusätzlich Zeile, Spalte und die
// eingesetzten Teile eines Ausdrucks mitbringt — und deshalb als einziger
// `internal/msescript` kennt.

package app

import (
	"errors"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// istSkriptfehler sagt, ob ein Zerteil- oder Auswertfehler darin steckt.
func istSkriptfehler(fehler error) bool {
	var stelle *msescript.Fehler

	return errors.As(fehler, &stelle)
}

// skriptfehler übersetzt ihn samt seinem Satz.
//
// **Der Satz steht in `Werte` und nicht nur in `Meldung`.** Das Fenster setzt
// den Katalogtext ein und füllt dessen Platzhalter aus `Werte`; wer ihn nur
// in `Meldung` legte, hätte ihn dort wieder verloren (#940).
//
// **Zeile und Spalte stehen nicht darin.** Sie kommen als eigenes Feld der
// Konsolenantwort ans Fenster, das sie zum Anstreichen braucht — ein zweites
// Mal im Satz wären sie dieselbe Angabe an zwei Orten.
func skriptfehler(fehler error) *Fehler {
	var stelle *msescript.Fehler

	// Die Wache steht in `istSkriptfehler`; hier wird nur ausgepackt.
	_ = errors.As(fehler, &stelle)

	werte := map[string]string{
		// **Der deutsche Satz bleibt der Rückfall** (#1178). Wo ein Schlüssel
		// im Katalog des Fensters fehlt — und für die 48 Sätze der
		// Standardbibliothek gibt es keinen —, steht er da.
		"meldung": stelle.Meldung,

		// **Und der Schlüssel daneben**, damit das Fenster übersetzen kann.
		"skriptschluessel": string(stelle.Schluessel),
	}

	// **Die eingesetzten Teile einzeln.** `nach %s fehlt %s` braucht zwei;
	// sie im Satz zu lassen hieße, ihn nicht übersetzen zu können.
	for i, wert := range stelle.Werte {
		werte["skriptwert"+strconv.Itoa(i)] = wert
	}

	return &Fehler{
		Code:       CodeValidierung,
		Schluessel: domain.SchluesselSkriptfehler,
		Meldung:    fehler.Error(),
		Werte:      werte,
	}
}
