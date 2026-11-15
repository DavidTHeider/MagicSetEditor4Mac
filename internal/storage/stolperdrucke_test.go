package storage

// **Der Stolpertreiber für die Druckablage** — Werkstück A2 von Phase 5.3
// (#1441).
//
// Vier Wege gibt eine echte Datenbank nicht her: eine Zeilennummer, die
// sich nicht lesen lässt; ein Abschluss, der scheitert, obwohl alles davor
// glückte; ein Fehler **mitten** im Lesen der Drucke oder ihrer Kennungen;
// und eine Zeile, die sich nicht in ihre Felder legen lässt. Hier stehen sie.

import (
	"database/sql/driver"
	"io"
	"strings"
)

// druckschreiben beantwortet die Schreibbefehle der Druckablage.
//
// **Die Kennung antwortet an jeder Stelle**, nicht nur an den Druckstellen:
// seit A2 legt jede der fünf Nähte vor ihrem eigenen Schreiben den Druck an
// (`druckZeile`), und ein Stolpertest über einen Deckplatz oder einen Abzug
// käme sonst nie bis zu der Stelle, um die es ihm geht.
func druckschreiben(stelle stolperstelle, abfrage string) (driver.Result, bool) {
	befehl := strings.TrimSpace(abfrage)

	if strings.HasPrefix(befehl, "INSERT INTO druckkennungen") {
		return druckergebnis{}, true
	}

	if !druckstelle(stelle) {
		return nil, false
	}

	switch {
	case strings.HasPrefix(befehl, "INSERT INTO drucke"):
		if stelle == beimDruckID {
			return zaehlfehler{}, true
		}

		return druckergebnis{}, true
	case strings.HasPrefix(befehl, "UPDATE drucke"),
		// **Die Sammlung ändert im Zug** (A2): ihr Commit ist der Weg, den nur
		// dieser Treiber hergibt — davor muss das Ändern eine Zeile zählen.
		strings.HasPrefix(befehl, "UPDATE sammlung"):
		return druckergebnis{}, true
	}

	return nil, false
}

// druckabfrage beantwortet die Abfragen der Druckablage.
func druckabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool) {
	befehl := strings.TrimSpace(abfrage)

	// **Die zwei Fragen von `druckZeile` antworten an jeder Stelle** — siehe
	// `druckschreiben`: kein Druck da, und der neue bekommt die Nummer 1.
	switch {
	case strings.HasPrefix(befehl, "SELECT druck_id FROM druckkennungen"):
		return &druckzeilen{spalten: []string{"druck_id"}}, true
	case strings.HasPrefix(befehl, "INSERT INTO drucke"):
		return &druckzeilen{spalten: []string{"id"}, zeilen: [][]driver.Value{{int64(1)}}}, true
	}

	if !druckstelle(stelle) {
		return nil, false
	}

	switch {
	case strings.HasPrefix(befehl, "SELECT d.id, d.name, d.setcode, d.nummer, d.sprache FROM drucke"):
		spalten := []string{"id", "name", "setcode", "nummer", "sprache"}

		switch stelle {
		case beimDruckende:
			return &druckzeilen{spalten: spalten, stolpert: true}, true
		case beimDruckscan:
			return &druckzeilen{spalten: spalten, zeilen: [][]driver.Value{{"x", "Bolt", "", "", ""}}}, true
		default:
			return &druckzeilen{spalten: spalten, zeilen: [][]driver.Value{{int64(1), "Bolt", "", "", ""}}}, true
		}
	case strings.HasPrefix(befehl, "SELECT system, semantik, wert FROM druckkennungen"):
		spalten := []string{"system", "semantik", "wert"}

		switch stelle {
		case beimKennungende:
			return &druckzeilen{spalten: spalten, stolpert: true}, true
		case beimKennungscan:
			return &druckzeilen{spalten: spalten, zeilen: [][]driver.Value{{"scryfall", "druck", nil}}}, true
		default:
			return &druckzeilen{spalten: spalten}, true
		}
	}

	return nil, false
}

func druckstelle(stelle stolperstelle) bool {
	switch stelle {
	case beimDruckID, beimDruckcommit, beimDruckende, beimDruckscan, beimKennungende, beimKennungscan:
		return true
	default:
		return false
	}
}

// druckergebnis ist ein geglücktes Schreiben mit der Zeilennummer 1.
type druckergebnis struct{}

func (druckergebnis) LastInsertId() (int64, error) { return 1, nil }
func (druckergebnis) RowsAffected() (int64, error) { return 1, nil }

// druckzeilen gibt vorbereitete Zeilen — oder stolpert beim ersten Schritt.
type druckzeilen struct {
	spalten  []string
	zeilen   [][]driver.Value
	stolpert bool
	gelesen  int
}

func (r *druckzeilen) Columns() []string { return r.spalten }
func (r *druckzeilen) Close() error      { return nil }

func (r *druckzeilen) Next(ziel []driver.Value) error {
	if r.stolpert {
		return errGestolpert
	}

	if r.gelesen >= len(r.zeilen) {
		return io.EOF
	}

	copy(ziel, r.zeilen[r.gelesen])
	r.gelesen++

	return nil
}
