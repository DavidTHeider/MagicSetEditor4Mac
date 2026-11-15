package storage

// **Der Stolpertreiber für den Arbeitsverlauf** — Werkstück U11 von Phase 5.1
// (#1422): ein Fehler mitten im Lesen, den SQLite nicht hergibt.

import (
	"database/sql/driver"
	"strings"
)

// arbeitsverlaufabfrage lässt das Lesen der Schritte oder des Angehefteten
// beim ersten Schritt stolpern.
func arbeitsverlaufabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool) {
	befehl := strings.TrimSpace(abfrage)

	switch {
	case stelle == beimArbeitsverlaufende && strings.HasPrefix(befehl, "SELECT art, kennung, name, vorgang, zeitpunkt FROM arbeitsverlauf"):
		return &stolpernbedarf{}, true
	case stelle == beimAngeheftetende && strings.HasPrefix(befehl, "SELECT art, kennung, name FROM angeheftet"):
		return &stolpernbedarf{}, true
	}

	return nil, false
}
