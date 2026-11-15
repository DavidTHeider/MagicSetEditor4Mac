package app

// **Der Rücksprung für die Flächen ohne Fenster** — Werkstück O7 von Phase
// 5.2 (#1431).
//
// # Keine dritte Wurzel
//
// Der erste Anlauf baute hier einen eigenen `service.Bestandsweg` — und damit
// eine **dritte** Stelle, an der einer entsteht. `nahtverdrahtung_test.go` hat
// es angestrichen: drei Wurzeln, die Verschiedenes verdrahten, sind drei
// Wahrheiten über dasselbe Objekt. Was hier steht, ist deshalb ein Griff auf
// den Weg, den `BestandsnahtUeber` ohnehin baut.

import (
	"database/sql"
)

// RuecksprungUeber baut eine Anwendung, die über die Rückkehren Auskunft gibt.
//
// **Der Bestandsort gehört dazu.** `Ruecksprunglage` packt die Sicherung neben
// den Bestand aus; ohne den Ort landete der Auszug irgendwo.
func RuecksprungUeber(db *sql.DB, bestandsort string) *App {
	anwendung := NeueApp(bestandsort)
	naht := BestandsnahtUeber(&db, bestandsort)

	anwendung.MitRuecksprung(ruecksprungnaht{weg: naht.Weg})

	return anwendung
}
