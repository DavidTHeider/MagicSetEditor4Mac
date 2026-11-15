package app

// **Die Datenpflege für die Flächen ohne Fenster** — Werkstück O8 von Phase
// 5.2 (#1432).

import (
	"database/sql"
)

// PflegeAnwendungUeber baut eine Anwendung, die pflegen kann.
//
// **Mit dem Bestandsort**, denn die Sicherung vor dem Schreiben landet neben
// dem Bestand — und nicht dort, wo der Prozess gerade steht.
func PflegeAnwendungUeber(db *sql.DB, bestandsort string) *App {
	anwendung := NeueApp(bestandsort)
	anwendung.MitPflege(PflegeUeber(func() *sql.DB { return db }, bestandsort))

	return anwendung
}
