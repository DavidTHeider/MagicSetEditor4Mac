package storage

// **Die Fehlerwege der Freigaben** — Werkstück 11 von Phase 4.12
// (#1351).
//
// **Ein interner Test**, weil die Stolpertreiber hier liegen. Was er prüft,
// gibt kein echter Treiber her: eine Abfrage, die scheitert; eine Zeile, die
// sich nicht lesen lässt; ein Abbruch **nach** der Schleife; und eine
// Löschung, deren Zeilen sich nicht zählen lassen.

import (
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Stellen, an denen der Freigabenweg stolpern kann.
const (
	beimFreigabenscan stolperstelle = iota + 1600
	beimFreigabenende
	beimFreigabenzaehlen
)

func init() {
	sql.Register("stolper-freigaben-scan",
		stolpertreiber{stelle: beimFreigabenscan})
	sql.Register("stolper-freigaben-ende",
		stolpertreiber{stelle: beimFreigabenende})
	sql.Register("stolper-freigaben-zaehlen",
		stolpertreiber{stelle: beimFreigabenzaehlen})
}

// freigabenabfrage beantwortet die Abfragen des Freigabenwegs.
func freigabenabfrage(stelle stolperstelle, abfrage string) (
	driver.Rows, bool, error,
) {
	if !strings.HasPrefix(strings.TrimSpace(abfrage),
		"SELECT profil_id, nutzer_id, rolle, angelegt") {
		return nil, false, nil
	}

	switch stelle {
	case beimFreigabenscan:
		return &krummeFreigabenzeile{}, true, nil
	case beimFreigabenende:
		return &abbrechendeFreigabenzeile{}, true, nil
	case beimFragen:
		return nil, true, errGestolpert
	}

	return &leerezeile{spalten: 4}, true, nil
}

// freigabenExec beantwortet ein `Exec` auf einem Freigabenweg.
func (s stolperstmt) freigabenExec() (
	ergebnis driver.Result, meiner bool, fehler error,
) {
	geglaettet := strings.TrimSpace(s.abfrage)

	if !strings.HasPrefix(geglaettet, "INSERT INTO profilfreigaben") &&
		!strings.HasPrefix(geglaettet, "DELETE FROM profilfreigaben") {
		return nil, false, nil
	}

	if s.stelle == beimSchreiben {
		return nil, true, errGestolpert
	}

	// **Eine Löschung, deren Zeilen sich nicht zählen lassen** — den Zweig
	// gibt kein echter Treiber her, und ohne ihn bliebe die Unterscheidung
	// „gab es nicht" gegen „ging schief" unbetreten.
	if s.stelle == beimFreigabenzaehlen {
		return zaehlfehler{}, true, nil
	}

	return driver.RowsAffected(1), true, nil
}

// krummeFreigabenzeile scheitert am Scan einer Freigabezeile.
type krummeFreigabenzeile struct{ gegeben bool }

func (r *krummeFreigabenzeile) Columns() []string { return nutzerspaltennamen(4) }

func (r *krummeFreigabenzeile) Close() error { return nil }

func (r *krummeFreigabenzeile) Next(werte []driver.Value) error {
	if r.gegeben {
		return errGestolpert
	}

	r.gegeben = true

	for i := range werte {
		werte[i] = nil
	}

	return nil
}

// abbrechendeFreigabenzeile bricht **nach** der Schleife ab.
type abbrechendeFreigabenzeile struct{}

func (r *abbrechendeFreigabenzeile) Columns() []string {
	return nutzerspaltennamen(4)
}

func (r *abbrechendeFreigabenzeile) Close() error { return nil }

func (r *abbrechendeFreigabenzeile) Next([]driver.Value) error {
	return errGestolpert
}

// **Jeder Freigabenweg reicht seinen Fehler durch.**
func TestJederFreigabenwegReichtSeinenFehlerDurch(t *testing.T) {
	amProfil := func(a *Freigabenablage) error {
		_, err := a.Freigaben(probekontext(t), 1)

		return err
	}

	beimNutzer := func(a *Freigabenablage) error {
		_, err := a.FreigabenFuer(probekontext(t), 1)

		return err
	}

	setzen := func(a *Freigabenablage) error {
		return a.FreigabeSetzen(probekontext(t), domain.Profilfreigabe{
			ProfilID: 1, NutzerID: 2, Rolle: domain.RolleLeser,
		})
	}

	loeschen := func(a *Freigabenablage) error {
		return a.FreigabeLoeschen(probekontext(t), 1, 2)
	}

	faelle := map[string]struct {
		treiber string
		lauf    func(*Freigabenablage) error
	}{
		"die Abfrage am Profil scheitert":   {"stolper-fragen", amProfil},
		"die Abfrage beim Nutzer scheitert": {"stolper-fragen", beimNutzer},
		"eine Zeile lässt sich nicht lesen": {"stolper-freigaben-scan", amProfil},
		"der Durchgang bricht ab":           {"stolper-freigaben-ende", amProfil},
		"das Setzen scheitert":              {"stolper-schreiben", setzen},
		"das Löschen scheitert":             {"stolper-schreiben", loeschen},
		"die Zeilen lassen sich nicht zählen": {
			"stolper-freigaben-zaehlen", loeschen,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "egal")
			if err != nil {
				t.Fatalf("öffnen: %v", err)
			}

			defer func() { _ = db.Close() }() //nolint:errcheck // Probe

			if fehler := fall.lauf(NeueFreigabenablage(db)); fehler == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}
}
