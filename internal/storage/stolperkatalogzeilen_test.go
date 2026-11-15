// **Was der Stolpertreiber dem Paketkatalog antwortet.**
//
// Eine eigene Datei, weil `stolperzeilen_test.go` mit ihr über 400 Zeilen
// ginge (§2). Der Schnitt folgt der Frage: dort stehen die Zeilen der Karten
// und Sets, hier die zwei, die der Katalog seit
// #1325
// abfragt.
package storage

import (
	"database/sql/driver"
	"io"
)

// zaehlerzeile gibt den Schreibzähler der Verbindung.
//
// Die Gestalt, die `Abgleichen` seit
// #1325
// abfragt: `SELECT total_changes()`.
//
// **Er zählt hoch, so wie SQLite es täte.** Zweimal derselbe Wert ergäbe
// „nichts geschrieben", und der Katalog gäbe zurück, **bevor** er bestätigt —
// der Fehlerweg „das Commit scheitert" bliebe unbetreten und der Test grün.
type zaehlerzeile struct{ gegeben bool }

// zaehlerstand ist der Stand über alle Abfragen eines Laufs.
//
// **Als Paketvariable wie `zweiterSchreibzaehler` daneben**: der Treiber baut
// je Abfrage eine neue Zeile, und ein Zähler in ihr begänne jedes Mal von
// vorn. `paketfehler_test.go` setzt ihn vor jedem Fall zurück.
var zaehlerstand int64

// zaehleraufrufe zählt, wie oft der Schreibzähler abgefragt wurde.
//
// **Für die Stelle `beimZweitenSchreibzaehler`**: `Abgleichen` fragt ihn
// zweimal — einmal am Anfang der Transaktion, einmal an ihrem Ende.
var zaehleraufrufe int

func (r *zaehlerzeile) Columns() []string { return []string{"total_changes()"} }
func (r *zaehlerzeile) Close() error      { return nil }

func (r *zaehlerzeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true
	zaehlerstand++
	werte[0] = zaehlerstand

	return nil
}

// bedarfszeilen geben die Abhängigkeiten eines Pakets — **keine**.
//
// Die Gestalt, die `bedarfGleich` abfragt (#1325). Eine leere Antwort heißt
// „im Katalog steht nichts", und damit schreibt der Abgleich — genau das
// braucht der Fehlerweg.
type bedarfszeilen struct{}

func (r *bedarfszeilen) Columns() []string         { return []string{"braucht", "version"} }
func (r *bedarfszeilen) Close() error              { return nil }
func (r *bedarfszeilen) Next([]driver.Value) error { return io.EOF }

// stolpernbedarf gibt eine Zeile, deren Lesen scheitert.
//
// **Für den Scan-Weg in `bedarfGleich`** (#1325): eine Antwort, die kommt und
// beim Auslesen umfällt. Ohne sie bliebe der Zweig eine Behauptung (§3).
type stolpernbedarf struct{}

func (r *stolpernbedarf) Columns() []string         { return []string{"braucht", "version"} }
func (r *stolpernbedarf) Close() error              { return nil }
func (r *stolpernbedarf) Next([]driver.Value) error { return errGestolpert }

// falscherbedarf gibt eine Zeile, deren **Gestalt** nicht passt.
//
// **Für den Scan-Weg** (#1325): `stolpernbedarf` daneben scheitert in `Next`
// und landet damit in `rows.Err()`; dieser hier liefert eine Zeile, an der
// `Scan` umfällt. Zwei Wege, zwei Zeilen — sonst bliebe einer eine Behauptung
// (§3).
type falscherbedarf struct{ gegeben bool }

// **Eine Spalte, wo zwei erwartet werden.** `Scan` weist das ab, und genau das
// ist der Zweig — ein Wert vom falschen Typ hilft nicht, weil `database/sql`
// fast jeden umwandelt.
func (r *falscherbedarf) Columns() []string { return []string{"braucht"} }
func (r *falscherbedarf) Close() error      { return nil }

func (r *falscherbedarf) Next(werte []driver.Value) error {
	if r.gegeben {
		return io.EOF
	}

	r.gegeben = true

	werte[0] = "eins"

	return nil
}

// leerezeilen geben nichts zurück.
//
// Für den Fall, dass `INSERT … ON CONFLICT DO UPDATE … WHERE` nichts geändert
// hat: `RETURNING` bleibt still, und der Katalog liest die Kennung nach.
type leerezeilen struct{}

func (r *leerezeilen) Columns() []string         { return []string{"id"} }
func (r *leerezeilen) Close() error              { return nil }
func (r *leerezeilen) Next([]driver.Value) error { return io.EOF }
