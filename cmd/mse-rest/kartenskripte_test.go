// **Die Verdrahtung der Kartenskripte** —
// #1337.
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
// `app.KartenskripteUeber` braucht ein Paketverzeichnis, und das hängt an
// `HOME`. Ohne eines gibt es kein Spielpaket; die Wurzel gibt den Fehler
// weiter, statt mit einem `nil` weiterzuarbeiten.
func TestOhneZuhauseKommenDieKartenskripteNichtZusammen(t *testing.T) {
	t.Setenv("HOME", "")

	var ohneBestand *sql.DB

	skripte := kartenskripte{db: ohneBestand}

	if _, fehler := skripte.Kartenskripte(probekontext(t), 1); fehler == nil {
		t.Error("ohne Zuhause kam ein Angebot zurück")
	}

	if _, fehler := skripte.KartenskriptAusfuehren(probekontext(t), 1, "Zwei Farben"); fehler == nil {
		t.Error("ohne Zuhause lief ein Skript")
	}
}
