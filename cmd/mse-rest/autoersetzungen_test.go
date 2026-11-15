// **Die Verdrahtung der Autoersetzungspflege** —
// #1341.
//
// # Was dieser Lauf hält
//
// Dass die Wurzel den Weg **erreicht** — nicht, was er tut. Der Arbeitsablauf
// über alle Flächen steht in `internal/flaechen`.
package main

import (
	"database/sql"
	"testing"
)

// **Ohne Zuhause kommt keiner der fünf Wege zusammen — und alle sagen es.**
func TestOhneZuhauseKommtDieAutoersetzungspflegeNichtZusammen(t *testing.T) {
	t.Setenv("HOME", "")

	var ohneBestand *sql.DB

	regeln := autoersetzungen{db: ohneBestand}

	if _, fehler := regeln.Autoersetzungen(probekontext(t), 1); fehler == nil {
		t.Error("ohne Zuhause kamen Regeln zurück")
	}

	if fehler := regeln.AutoersetzungSetzen(
		probekontext(t), 1, "(W)", "weiss", true, true); fehler == nil {
		t.Error("ohne Zuhause wurde gesetzt")
	}

	if fehler := regeln.AutoersetzungLoeschen(probekontext(t), 1, "(W)"); fehler == nil {
		t.Error("ohne Zuhause wurde gelöscht")
	}

	if fehler := regeln.AutoersetzungenZuruecksetzen(probekontext(t), 1); fehler == nil {
		t.Error("ohne Zuhause wurde zurückgesetzt")
	}

	if fehler := regeln.AutoersetzungSchalten(probekontext(t), 1, false); fehler == nil {
		t.Error("ohne Zuhause wurde geschaltet")
	}
}
