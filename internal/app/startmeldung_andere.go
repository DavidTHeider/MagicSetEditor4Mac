//go:build !darwin

package app

import "errors"

// meldungZeigen gibt es außerhalb von macOS nicht.
//
// **Und das ist keine Lücke, sondern der Stand von §12**: gebaut wird für
// darwin/arm64. Die Übersetzung nach Windows und Linux läuft mit, damit
// plattformabhängiger Code nicht unbemerkt bricht — sie ist keine Zusage, dass
// dort etwas funktioniert.
//
// Der Aufrufer fängt den Fehler und schreibt nach stderr. Auf einer Plattform
// ohne Fenster ist das die richtige Antwort und keine halbe.
func meldungZeigen(string, string) error {
	return errors.New("ein Meldefenster gibt es auf dieser Plattform nicht")
}
