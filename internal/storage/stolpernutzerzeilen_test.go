// **Die Zeilenatrappen des Nutzerwegs** — was der Stolpertreiber zurückgibt.
//
// Aus `stolpernutzer_test.go` herausgelöst, als es über 400 Zeilen ging (§2).
// Der Schnitt folgt der Frage: dort **welche Abfrage welche Antwort bekommt**,
// hier, **wie eine Antwort aussieht**.
package storage

import (
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
)

// krummeFuenferzeile scheitert am Scan einer Kontozeile.
type krummeFuenferzeile struct{ gegeben bool }

func (r *krummeFuenferzeile) Columns() []string {
	return []string{"a", "b", "c", "d", "e"}
}

func (r *krummeFuenferzeile) Close() error { return nil }

func (r *krummeFuenferzeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true

	for i := range werte {
		werte[i] = nil
	}

	return nil
}

// abbrechendeFuenferzeile bricht **nach** der Schleife ab.
type abbrechendeFuenferzeile struct{}

func (r *abbrechendeFuenferzeile) Columns() []string {
	return []string{"a", "b", "c", "d", "e"}
}

func (r *abbrechendeFuenferzeile) Close() error { return nil }

func (r *abbrechendeFuenferzeile) Next([]driver.Value) error { return errGestolpert }

// namenszeile gibt genau einen Namen.
//
// **Und zwar den, den die Vorrichtung erwartet**: die Namensprüfung vergleicht
// ihn mit dem, was der Aufrufer nennt — ein anderer Name führte in die
// Ablehnung statt in den Weg dahinter.
type namenszeile struct {
	name    string
	gegeben bool
}

func (r *namenszeile) Columns() []string { return []string{"name"} }
func (r *namenszeile) Close() error      { return nil }

func (r *namenszeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true
	werte[0] = r.name

	return nil
}

// zahlenzeile gibt genau eine Zahl.
type zahlenzeile struct {
	zahl    int64
	gegeben bool
}

func (r *zahlenzeile) Columns() []string { return []string{"zahl"} }
func (r *zahlenzeile) Close() error      { return nil }

func (r *zahlenzeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true
	werte[0] = r.zahl

	return nil
}

// nutzerspaltennamen gibt so viele Namen, wie eine Nutzerzeile Spalten hat.
//
// **Gerechnet und nicht aufgezählt**: die Zahl wuchs von drei über acht auf
// siebzehn, und eine ausgeschriebene Liste ginge beim nächsten Feld wieder
// daneben.
func nutzerspaltennamen(wieviele int) []string {
	namen := make([]string, wieviele)
	for i := range namen {
		namen[i] = fmt.Sprintf("s%d", i)
	}

	return namen
}

// krummeNutzerzeile scheitert am Scan einer Nutzerzeile.
//
// **NULL und nicht „keine Zahl"** — `database/sql` macht aus jedem anderen
// Wert klaglos eine Zeichenkette.
type krummeNutzerzeile struct{ gegeben bool }

func (r *krummeNutzerzeile) Columns() []string { return nutzerspaltennamen(17) }

func (r *krummeNutzerzeile) Close() error { return nil }

func (r *krummeNutzerzeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true

	for i := range werte {
		werte[i] = nil
	}

	return nil
}

// abbrechendeNutzerzeile bricht **nach** der Schleife ab.
type abbrechendeNutzerzeile struct{}

func (r *abbrechendeNutzerzeile) Columns() []string { return nutzerspaltennamen(17) }

func (r *abbrechendeNutzerzeile) Close() error { return nil }

func (r *abbrechendeNutzerzeile) Next([]driver.Value) error { return errGestolpert }

// krummeViererzeile scheitert am Scan einer Profilzeile.
type krummeViererzeile struct{ gegeben bool }

func (r *krummeViererzeile) Columns() []string { return nutzerspaltennamen(4) }

func (r *krummeViererzeile) Close() error { return nil }

func (r *krummeViererzeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true

	for i := range werte {
		werte[i] = nil
	}

	return nil
}

// profilspaltenzahl ist die Zahl der Spalten, die eine Profilzeile trägt —
// **aus [profilspalten] gezählt** und nicht abgeschrieben: hier stand eine
// Acht aus Werkstück 9 (#1351), als längst elf gelesen wurden, und jede
// krumme Zeile scheiterte an der Spaltenzahl statt am Scan (Zweitblick über
// Phase B, Befund 3).
var profilspaltenzahl = len(strings.Split(profilspalten, ","))

// abbrechendeProfilzeile bricht **nach** der Schleife ab.
type abbrechendeProfilzeile struct{}

func (r *abbrechendeProfilzeile) Columns() []string { return nutzerspaltennamen(profilspaltenzahl) }

func (r *abbrechendeProfilzeile) Close() error { return nil }

func (r *abbrechendeProfilzeile) Next([]driver.Value) error { return errGestolpert }

// abbrechendeViererzeile bricht **nach** der Schleife ab.
type abbrechendeViererzeile struct{}

func (r *abbrechendeViererzeile) Columns() []string { return nutzerspaltennamen(4) }

func (r *abbrechendeViererzeile) Close() error { return nil }

func (r *abbrechendeViererzeile) Next([]driver.Value) error { return errGestolpert }
