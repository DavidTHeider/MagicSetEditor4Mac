package storage

// **Die Fehlerwege der Profilvorgaben** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7.
//
// **Ein interner Test**, weil die Stolpertreiber hier liegen. Was er prüft,
// gibt kein echter Treiber her: eine Abfrage, die scheitert; eine Zeile, die
// sich nicht lesen lässt; ein Abbruch **nach** der Schleife; und die drei
// Stellen einer Transaktion — Beginn, Räumen, Festschreiben.

import (
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Stellen, an denen der Vorgabenweg stolpern kann.
const (
	beimVorgabenscan stolperstelle = iota + 1500
	beimVorgabenende
	beimVorgabenraeumen
	beimVorgabenbeginn
	beimVorgabennachschlagen
)

func init() {
	sql.Register("stolper-vorgaben-scan", stolpertreiber{stelle: beimVorgabenscan})
	sql.Register("stolper-vorgaben-ende", stolpertreiber{stelle: beimVorgabenende})
	sql.Register("stolper-vorgaben-raeumen",
		stolpertreiber{stelle: beimVorgabenraeumen})
	sql.Register("stolper-vorgaben-beginn",
		stolpertreiber{stelle: beimVorgabenbeginn})
	sql.Register("stolper-vorgaben-nachschlagen",
		stolpertreiber{stelle: beimVorgabennachschlagen})
}

// vorgabenabfrage beantwortet die Abfragen des Vorgabenwegs — oder stolpert.
func vorgabenabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool, error) {
	geglaettet := strings.TrimSpace(abfrage)

	// **Die Frage, ob es das Profil gibt, beantwortet `nutzerabfrage`** — sie
	// steht in der Kette davor und kennt dieselbe Abfrage schon. Ein zweiter
	// Zweig hier käme nie zum Zug; der erste Anlauf hatte einen, und er war
	// grün, ohne etwas zu prüfen (§10).

	if !strings.HasPrefix(geglaettet, "SELECT art, wert FROM profilvorgaben") {
		return nil, false, nil
	}

	switch stelle {
	case beimVorgabenscan:
		return &krummeVorgabenzeile{}, true, nil
	case beimVorgabenende:
		return &abbrechendeVorgabenzeile{}, true, nil
	case beimFragen:
		return nil, true, errGestolpert
	}

	return &leerezeile{spalten: 2}, true, nil
}

// vorgabenExec beantwortet ein `Exec` auf dem Vorgabenweg.
//
// **Das Räumen und das Eintragen sind zwei Stellen**, und beide gehören
// betreten: die erste steht **vor** der Schleife, die zweite darin.
func (s stolperstmt) vorgabenExec() (ergebnis driver.Result, meiner bool, fehler error) {
	geglaettet := strings.TrimSpace(s.abfrage)

	if strings.HasPrefix(geglaettet, "DELETE FROM profilvorgaben") {
		if s.stelle == beimVorgabenraeumen {
			return nil, true, errGestolpert
		}

		return driver.RowsAffected(1), true, nil
	}

	if strings.HasPrefix(geglaettet, "INSERT INTO profilvorgaben") &&
		s.stelle != beimSchreiben {
		return driver.RowsAffected(1), true, nil
	}

	return nil, false, nil
}

// krummeVorgabenzeile scheitert am Scan einer Vorgabezeile.
type krummeVorgabenzeile struct{ gegeben bool }

func (r *krummeVorgabenzeile) Columns() []string { return nutzerspaltennamen(2) }

func (r *krummeVorgabenzeile) Close() error { return nil }

func (r *krummeVorgabenzeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return errGestolpert
	}

	r.gegeben = true

	for i := range werte {
		werte[i] = nil
	}

	return nil
}

// abbrechendeVorgabenzeile bricht **nach** der Schleife ab.
type abbrechendeVorgabenzeile struct{}

func (r *abbrechendeVorgabenzeile) Columns() []string { return nutzerspaltennamen(2) }

func (r *abbrechendeVorgabenzeile) Close() error { return nil }

func (r *abbrechendeVorgabenzeile) Next([]driver.Value) error { return errGestolpert }

// **Jeder Vorgabenweg reicht seinen Fehler durch.**
func TestJederVorgabenwegReichtSeinenFehlerDurch(t *testing.T) {
	setzen := func(a *Vorgabenablage) error {
		return a.VorgabeSetzen(probekontext(t), 1, domain.Profilvorgabe{
			Art: domain.VorgabeDeck, Werte: []string{"7"},
		})
	}

	lesen := func(a *Vorgabenablage) error {
		_, err := a.Vorgaben(probekontext(t), 1)

		return err
	}

	faelle := map[string]struct {
		treiber string
		lauf    func(*Vorgabenablage) error
	}{
		"das Nachschlagen des Profils scheitert": {
			"stolper-vorgaben-nachschlagen", lesen,
		},
		"die Abfrage scheitert":             {"stolper-fragen", lesen},
		"eine Zeile lässt sich nicht lesen": {"stolper-vorgaben-scan", lesen},
		"der Durchgang bricht ab":           {"stolper-vorgaben-ende", lesen},
		"die Transaktion beginnt nicht":     {"stolper-vorgaben-beginn", setzen},
		"das Räumen scheitert":              {"stolper-vorgaben-raeumen", setzen},
		"das Eintragen scheitert":           {"stolper-schreiben", setzen},
		"das Festschreiben scheitert":       {"stolper-commit", setzen},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "egal")
			if err != nil {
				t.Fatalf("öffnen: %v", err)
			}

			defer func() { _ = db.Close() }() //nolint:errcheck // Probe

			if fehler := fall.lauf(NeueVorgabenablage(db)); fehler == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}
}
