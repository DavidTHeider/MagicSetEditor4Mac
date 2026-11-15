package storage

// **Die Fehlerwege des Partiespeichers, die kein echter Treiber hergibt** —
// Werkstück G1 von Phase 5.5
// (#1458).
//
// **Ein interner Test**, weil die Stolpertreiber hier liegen. Was er prüft,
// gibt SQLite nicht her: eine Zeile, die sich nicht lesen lässt; eine
// Transaktion, die gar nicht erst beginnt; ein `RowsAffected`, das klagt.
//
// **Eine eigene Datei**, damit `stolperabfragen_test.go` unter der
// Zeilengrenze aus §2 bleibt — dieselbe Form wie beim Turnier- und beim
// Nutzerweg.

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Stellen, an denen der Partieweg stolpern kann.
const (
	beimPartiescan stolperstelle = iota + 1900
	beimPartienende
	beimEreignisscan
	beimEreignisende
	beimPartiebeginn
	beimPartiezaehlen
	beimStandschreiben

	// **Das Löschen zählt seine Zeilen** — PLAY-03.AK-05 (#1711): aus null
	// betroffenen wird „die Partie gibt es in diesem Profil nicht", und wer
	// das liest, hört auf zu suchen. Ist schon die **Zählung** gescheitert,
	// ist die Frage unbeantwortet, und das ist etwas anderes.
	beimPartieloeschzaehlen
)

func init() {
	sql.Register("stolper-partie-scan", stolpertreiber{stelle: beimPartiescan})
	sql.Register("stolper-partien-ende", stolpertreiber{stelle: beimPartienende})
	sql.Register("stolper-ereignis-scan", stolpertreiber{stelle: beimEreignisscan})
	sql.Register("stolper-ereignis-ende", stolpertreiber{stelle: beimEreignisende})
	sql.Register("stolper-partie-beginn", stolpertreiber{stelle: beimPartiebeginn})
	sql.Register("stolper-partie-zaehlen", stolpertreiber{stelle: beimPartiezaehlen})
	sql.Register("stolper-stand-schreiben", stolpertreiber{stelle: beimStandschreiben})
	sql.Register("stolper-partieloesch-zaehlen", stolpertreiber{stelle: beimPartieloeschzaehlen})
}

// partieabfrage beantwortet die Abfragen des Partiewegs — oder stolpert.
func partieabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool) {
	geglaettet := strings.TrimSpace(abfrage)

	switch {
	case strings.HasPrefix(geglaettet, "SELECT zustand FROM partien"):
		if stelle == beimPartiescan {
			// **NULL und nicht „keine Zeichenkette"**: die Spalte ist TEXT,
			// und `database/sql` macht aus jedem anderen Wert klaglos eine.
			// Nur ein `NULL` scheitert am Scan.
			return &nulleinerzeile{}, true
		}

		if stelle == beimPartienende {
			return &abbrechendeEinerzeile{}, true
		}

		// **Eine heile Zeile für alle anderen Stellen** — sonst endete die
		// Schleife bei null Partien, und die Stelle danach bliebe unbetreten
		// (§10).
		return &partiezeile{}, true

	case strings.HasPrefix(geglaettet, "SELECT folge, art, oeffentlich"):
		switch stelle {
		case beimEreignisscan:
			return &krummeEreigniszeile{}, true
		case beimEreignisende:
			return &abbrechendeEreigniszeile{}, true
		}

		return &leerezeile{spalten: 5}, true
	}

	return nil, false
}

// partiezeile gibt eine lesbare Partie.
type partiezeile struct{ gegeben bool }

func (p *partiezeile) Columns() []string { return []string{"zustand"} }
func (p *partiezeile) Close() error      { return nil }

func (p *partiezeile) Next(ziel []driver.Value) error {
	if p.gegeben {
		return io.EOF
	}

	p.gegeben = true
	ziel[0] = `{"ID":"p1","Stand":"laeuft","Fassung":1}`

	return nil
}

// nulleinerzeile gibt eine Spalte mit NULL — daran scheitert der Scan.
type nulleinerzeile struct{ gegeben bool }

func (n *nulleinerzeile) Columns() []string { return []string{"zustand"} }
func (n *nulleinerzeile) Close() error      { return nil }

func (n *nulleinerzeile) Next(ziel []driver.Value) error {
	if n.gegeben {
		return io.EOF
	}

	n.gegeben = true
	ziel[0] = nil

	return nil
}

// abbrechendeEinerzeile bricht **nach** der Schleife ab.
type abbrechendeEinerzeile struct{}

func (abbrechendeEinerzeile) Columns() []string         { return []string{"zustand"} }
func (abbrechendeEinerzeile) Close() error              { return nil }
func (abbrechendeEinerzeile) Next([]driver.Value) error { return errGestolpert }

// krummeEreigniszeile gibt ein NULL in einer Textspalte.
type krummeEreigniszeile struct{ gegeben bool }

func (k *krummeEreigniszeile) Columns() []string {
	return []string{"folge", "art", "oeffentlich", "privat", "kette"}
}

func (k *krummeEreigniszeile) Close() error { return nil }

func (k *krummeEreigniszeile) Next(ziel []driver.Value) error {
	if k.gegeben {
		return io.EOF
	}

	k.gegeben = true
	ziel[0], ziel[1], ziel[2], ziel[3], ziel[4] = int64(1), "a", nil, "[]", "k"

	return nil
}

// abbrechendeEreigniszeile bricht **nach** der Schleife ab.
type abbrechendeEreigniszeile struct{}

func (abbrechendeEreigniszeile) Columns() []string {
	return []string{"folge", "art", "oeffentlich", "privat", "kette"}
}

func (abbrechendeEreigniszeile) Close() error              { return nil }
func (abbrechendeEreigniszeile) Next([]driver.Value) error { return errGestolpert }

// partieschreiben beantwortet die Anweisungen des Partiewegs.
func partieschreiben(
	stelle stolperstelle, abfrage string,
) (driver.Result, bool, error) {
	geglaettet := strings.TrimSpace(abfrage)

	if strings.HasPrefix(geglaettet, "DELETE FROM partien") &&
		stelle == beimPartieloeschzaehlen {
		return zaehlendesErgebnis{}, true, nil
	}

	if !strings.HasPrefix(geglaettet, "UPDATE partien") {
		return nil, false, nil
	}

	switch stelle {
	case beimStandschreiben:
		return nil, true, errGestolpert

	case beimPartiezaehlen:
		// **Ein Ergebnis, dessen Zählung klagt.** SQLite gibt hier nie einen
		// Fehler; ohne diesen Treiber stünde der Zweig daneben ungeprüft da
		// und sähe aus wie eine geprüfte Zusage (§10).
		return zaehlendesErgebnis{}, true, nil
	}

	return nil, false, nil
}

// zaehlendesErgebnis klagt beim Zählen der geänderten Zeilen.
type zaehlendesErgebnis struct{}

func (zaehlendesErgebnis) LastInsertId() (int64, error) { return 0, nil }
func (zaehlendesErgebnis) RowsAffected() (int64, error) { return 0, errGestolpert }

// **Jeder Weg des Partiespeichers reicht seinen Fehler durch.**
func TestJederPartiewegReichtSeinenFehlerDurch(t *testing.T) {
	faelle := map[string]struct {
		treiber string
		fahre   func(*Partieablage) error
	}{
		"eine Partiezeile lässt sich nicht lesen": {
			"stolper-partie-scan", func(a *Partieablage) error {
				_, fehler := a.Partien(probekontext(t))

				return fehler
			},
		},
		"die Partienzeilen brechen ab": {
			"stolper-partien-ende", func(a *Partieablage) error {
				_, fehler := a.Partien(probekontext(t))

				return fehler
			},
		},
		"eine Ereigniszeile lässt sich nicht lesen": {
			"stolper-ereignis-scan", func(a *Partieablage) error {
				_, fehler := a.Ereignisse(probekontext(t), "p1", 1, 9)

				return fehler
			},
		},
		"die Ereigniszeilen brechen ab": {
			"stolper-ereignis-ende", func(a *Partieablage) error {
				_, fehler := a.Ereignisse(probekontext(t), "p1", 1, 9)

				return fehler
			},
		},
		"die Transaktion beginnt nicht": {
			"stolper-partie-beginn", func(a *Partieablage) error {
				return a.Anwenden(probekontext(t), 1, "b1",
					domain.Partie{ID: "p1"}, nil)
			},
		},
		"die geänderten Zeilen lassen sich nicht zählen": {
			"stolper-partie-zaehlen", func(a *Partieablage) error {
				return a.Anwenden(probekontext(t), 1, "b1",
					domain.Partie{ID: "p1"}, nil)
			},
		},
		"die gelöschten Zeilen lassen sich nicht zählen": {
			"stolper-partieloesch-zaehlen", func(a *Partieablage) error {
				return a.Loeschen(probekontext(t), "p1")
			},
		},
		"der Stand lässt sich nicht schreiben": {
			"stolper-stand-schreiben", func(a *Partieablage) error {
				_, fehler := a.Standsetzen(
					probekontext(t), "p1", domain.StandBeendet)

				return fehler
			},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			db, fehler := sql.Open(fall.treiber, "egal")
			if fehler != nil {
				t.Fatalf("öffnen: %v", fehler)
			}

			t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // nur zum Aufräumen

			if fehler := fall.fahre(NeuePartieablage(db)); fehler == nil {
				t.Error("der Weg ging durch")
			}
		})
	}
}
