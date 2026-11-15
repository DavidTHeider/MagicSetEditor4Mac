// **Der Treiber selbst** — Verbindung, Anweisung, Ergebnis.
//
// **Eine eigene Datei**, weil `stolpertreiber_test.go` über die Zeilengrenze
// wuchs, als der Vermerk „zuletzt benutzt" eine Stolperstelle mehr brauchte
// (§2). Der Schnitt folgt der Frage: dort **welche Stellen es gibt** und wie
// sie heißen, hier **wie gestolpert wird**.
package storage

import (
	"database/sql/driver"
)

type stolpertreiber struct{ stelle stolperstelle }

func (t stolpertreiber) Open(string) (driver.Conn, error) { return stolperconn(t), nil }

type stolperconn stolpertreiber

func (c stolperconn) Prepare(abfrage string) (driver.Stmt, error) {
	aufzeichnen(c.stelle, abfrage)

	return stolperstmt{stelle: c.stelle, abfrage: abfrage}, nil
}
func (c stolperconn) Close() error { return nil }
func (c stolperconn) Begin() (driver.Tx, error) {
	aufzeichnen(c.stelle, "BEGIN")

	// **Ein Zug, der gar nicht erst beginnt.** Ein echter Treiber tut das
	// nur, wenn die Datei weg ist — und dann steht die Anwendung ohnehin.
	if c.stelle == beimRundenbeginn ||
		c.stelle == beimEintragenbeginnen ||
		c.stelle == beimAbmeldenbeginnen ||
		c.stelle == beimNutzerbeginn ||
		c.stelle == beimVorgabenbeginn ||
		c.stelle == beimPartiebeginn ||
		c.stelle == beimSichtenbeginn ||
		c.stelle == beimAbschlussbeginn ||
		c.stelle == beimTurnieranlagebeginn {
		return nil, errGestolpert
	}

	return stolpertx(c), nil
}
