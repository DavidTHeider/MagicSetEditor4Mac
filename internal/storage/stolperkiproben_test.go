// **Was der Stolpertreiber den Modellproben antwortet** — Werkstück A9 von
// Phase 5.3
// (#1448).
//
// **Eine eigene Datei**, wie bei den Drucken: `stolperabfragen_test.go` ginge
// mit ihr über die Zeilengrenze aus §2.
//
// # Warum es diese zwei Stellen gibt
//
// **Eine echte Datenbank gibt sie nicht her.** Ein Fehler **mitten** im Lesen
// und eine Zeile, die nicht in ihre Felder passt, entstehen im Treiber und
// nirgends sonst — die übrigen Wege (Tabelle weg, geschlossener Bestand)
// laufen gegen SQLite selbst.
package storage

import (
	"database/sql/driver"
	"strings"
)

// probenabfrage antwortet auf die eine Abfrage der Modellproben.
func probenabfrage(stelle stolperstelle, befehl string) (driver.Rows, bool) {
	if !probenstelle(stelle) ||
		!strings.HasPrefix(befehl, "\n\t\tSELECT anbieter, alias, faehigkeit") {
		return nil, false
	}

	spalten := []string{
		"anbieter", "alias", "faehigkeit", "datum", "beleg", "laufzeit", "modell",
	}

	if stelle == beimProbenende {
		return &druckzeilen{spalten: spalten, stolpert: true}, true
	}

	// **Ein `NULL`, wo eine Zeichenkette hingehört** — daran scheitert der
	// Scan, und nur daran.
	//
	// **Eine Zahl täte es nicht**: `database/sql` wandelt sie stillschweigend
	// in eine Zeichenkette, und der Lauf bliebe grün.
	return &druckzeilen{spalten: spalten, zeilen: [][]driver.Value{
		{nil, "a", "b", "c", "d", "e", "f"},
	}}, true
}

// probenstelle sagt, ob eine Stelle zu den Modellproben gehört.
func probenstelle(stelle stolperstelle) bool {
	return stelle == beimProbenende || stelle == beimProbenscan
}
