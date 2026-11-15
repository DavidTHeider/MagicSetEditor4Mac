package storage

// **Die Fehlerwege des Setteils einer Standprüfung** — #2533. Wie bei der
// Karte (`standfehler_intern_test.go`): ein verschluckter Lesefehler gäbe hier
// keine unbekannte, sondern eine **bestandene** Bedingung. Die Karte steht
// dabei auf dem erwarteten Stand, damit der Lauf bis zu den Setangaben kommt.

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Stellen, an denen der Setteil stolpern kann.
const (
	beimSetstandFragen stolperstelle = iota + 2300
	beimSetstandScan
	beimSetstandEnde
)

func init() {
	sql.Register("stolper-setstand-fragen", stolpertreiber{stelle: beimSetstandFragen})
	sql.Register("stolper-setstand-scan", stolpertreiber{stelle: beimSetstandScan})
	sql.Register("stolper-setstand-ende", stolpertreiber{stelle: beimSetstandEnde})
}

// setstandabfrage beantwortet Karte und Setangaben — oder stolpert an den
// Setangaben.
func setstandabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool, error) {
	if stelle < beimSetstandFragen || stelle > beimSetstandEnde {
		return nil, false, nil
	}

	switch {
	case strings.Contains(abfrage, "FROM set_values"):
		switch stelle {
		case beimSetstandFragen:
			return nil, true, errGestolpert
		case beimSetstandScan:
			return &nullzweierzeile{}, true, nil
		}

		return abbrechendeZweierzeile{}, true, nil

	case strings.Contains(abfrage, "SELECT card_values.feld, card_values.wert"):
		return &einekartenzeile{}, true, nil
	}

	return nil, false, nil
}

// einekartenzeile gibt eine Karte mit einem Feld.
type einekartenzeile struct{ gegeben bool }

func (*einekartenzeile) Columns() []string { return []string{"feld", "wert"} }
func (*einekartenzeile) Close() error      { return nil }

func (e *einekartenzeile) Next(werte []driver.Value) error {
	if e.gegeben {
		return io.EOF
	}

	e.gegeben = true
	werte[0], werte[1] = "name", "x"

	return nil
}

func TestJederWegDesSetteilsReichtSeinenFehlerDurch(t *testing.T) {
	karte := domain.CardRow{ID: 7, Values: []domain.Value{{Field: "name", Text: "x"}}}
	erwartet := domain.Rechenstand(karte.Stand(), nil)

	for name, treiber := range map[string]string{
		"die Abfrage scheitert":             "stolper-setstand-fragen",
		"eine Zeile lässt sich nicht lesen": "stolper-setstand-scan",
		"die Zeilen brechen ab":             "stolper-setstand-ende",
	} {
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

			// **Kein Konflikt** — die Karte stimmt; gescheitert ist das Lesen.
			gelesen := standHalten(probekontext(t), geschaeft, 7, erwartet)
			if gelesen == nil || errors.Is(gelesen, domain.ErrConflict) {
				t.Errorf("Fehler %v, erwartet einer beim Lesen der Setangaben", gelesen)
			}
		})
	}
}
