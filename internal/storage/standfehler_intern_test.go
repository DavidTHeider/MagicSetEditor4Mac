package storage

// **Die Fehlerwege der Standprüfung** — Befund 2 des Zweitblicks über Phase
// 5.3.
//
// **Ein interner Test**, weil die Stolpertreiber hier liegen. Was er prüft,
// gibt kein echter Treiber her: eine Abfrage, die scheitert; eine Zeile, die
// sich nicht lesen lässt; ein Abbruch **nach** der Schleife.
//
// **Warum diese Wege überhaupt zählen**: die Standprüfung entscheidet, ob
// eine Übernahme durchgeht. Ein verschluckter Lesefehler gäbe hier nicht
// „unbekannt", sondern eine **bestandene** Bedingung — und damit genau die
// stille Überschreibung, gegen die es sie gibt.

import (
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"
)

// Die Stellen, an denen die Standprüfung stolpern kann.
const (
	beimStandscan stolperstelle = iota + 2200
	beimStandende
)

func init() {
	sql.Register("stolper-stand-scan", stolpertreiber{stelle: beimStandscan})
	sql.Register("stolper-stand-ende", stolpertreiber{stelle: beimStandende})
}

// standabfrage beantwortet die Abfrage der Standprüfung — oder stolpert.
func standabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool, error) {
	if zeilen, erkannt, fehler := setstandabfrage(stelle, abfrage); erkannt {
		return zeilen, true, fehler
	}

	if !strings.Contains(abfrage, "SELECT card_values.feld, card_values.wert") {
		return nil, false, nil
	}

	switch stelle {
	case beimStandscan:
		return &einspaltigezeile{}, true, nil
	case beimStandende:
		return &abbrechendeZweierzeile{}, true, nil
	case beimFragen:
		return nil, true, errGestolpert
	}

	// **Ohne Stolperstelle keine Zeile** — und das ist der vierte Fall: eine
	// Karte, die es nicht gibt, hat keinen Stand.
	return &leerezeile{spalten: 2}, true, nil
}

// einspaltigezeile gibt **eine** Spalte, wo zwei erwartet werden.
//
// **Der einzige Weg, diesen Scan scheitern zu lassen.** Gelesen wird in zwei
// `sql.NullString`; die nehmen jeden Wert und auch `NULL`. Nur die Zahl der
// Spalten weist `database/sql` ab — dieselbe Überlegung wie bei
// [falscherbedarf].
type einspaltigezeile struct{ gegeben bool }

func (einspaltigezeile) Columns() []string { return []string{"feld"} }
func (einspaltigezeile) Close() error      { return nil }

func (e *einspaltigezeile) Next(werte []driver.Value) error {
	if e.gegeben {
		return errGestolpert
	}

	e.gegeben = true
	werte[0] = "name"

	return nil
}

// **Jeder Weg der Standprüfung reicht seinen Fehler durch.**
func TestJederWegDerStandpruefungReichtSeinenFehlerDurch(t *testing.T) {
	faelle := map[string]string{
		"die Abfrage scheitert":             "stolper-fragen",
		"eine Zeile lässt sich nicht lesen": "stolper-stand-scan",
		"die Zeilen brechen ab":             "stolper-stand-ende",

		// **Und eine Karte, die es nicht gibt, hat keinen Stand.** Ein
		// leeres Ergebnis als „Stand stimmt" zu lesen, hieße gegen eine
		// gelöschte Karte zu schreiben.
		"die Karte gibt es nicht": "stolper-commit",
	}

	for name, treiber := range faelle {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open(treiber, "egal")
			if err != nil {
				t.Fatalf("öffnen: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // nur zum Aufräumen

			geschaeft, fehler := db.BeginTx(probekontext(t), nil)
			if fehler != nil {
				t.Fatalf("das Geschäft beginnen: %v", fehler)
			}

			defer func() { _ = geschaeft.Rollback() }() //nolint:errcheck // nur gelesen

			if fehler := standHalten(
				probekontext(t), geschaeft, 7, "ein-stand"); fehler == nil {
				t.Error("die Bedingung kam ohne Stand durch")
			}
		})
	}
}
