// **Die Verdrahtung der Parameterhilfen** —
// #1342.
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

// **Ohne Zuhause kommt die Naht nicht zusammen — und sagt es.**
func TestOhneZuhauseKommenDieParameterhilfenNichtZusammen(t *testing.T) {
	t.Setenv("HOME", "")

	var ohneBestand *sql.DB

	if _, fehler := (parameterhilfen{db: ohneBestand}).Parameterhilfen(
		probekontext(t), 1, 2); fehler == nil {
		t.Error("ohne Zuhause kamen Hilfen zurück")
	}
}
