// **Die Verdrahtung der Auszählung** —
// #1338.
//
// # Was dieser Lauf hält
//
// Dass die Wurzel den Weg **erreicht** — nicht, was er tut. Der Arbeitsablauf
// über alle Flächen steht in `internal/flaechen`; hier geht es um die
// Verdrahtung, die sonst niemand ausführt.
package main

import (
	"database/sql"
	"testing"
)

// **Ohne Zuhause kommt die Naht nicht zusammen — und sagt es.**
//
// `app.AuszaehlungUeber` braucht ein Paketverzeichnis, und das hängt an
// `HOME`. Ohne eines gibt es kein Spielpaket; die Wurzel gibt den Fehler
// weiter, statt mit einem `nil` weiterzuarbeiten.
func TestOhneZuhauseKommtDieAuszaehlungNichtZusammen(t *testing.T) {
	t.Setenv("HOME", "")

	var ohneBestand *sql.DB

	zaehlung := auszaehlung{db: ohneBestand}

	if _, fehler := zaehlung.Auszaehldimensionen(probekontext(t), 1); fehler == nil {
		t.Error("ohne Zuhause kam ein Angebot zurück")
	}

	if _, fehler := zaehlung.Auszaehlen(probekontext(t), 1, "farbe"); fehler == nil {
		t.Error("ohne Zuhause wurde gezählt")
	}
}
