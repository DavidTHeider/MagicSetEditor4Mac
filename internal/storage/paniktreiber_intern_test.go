package storage

// **Ein Treiber, der planmäßig zusammenbricht** — BK-04
// (#2129).
//
// Er legt sich um den **echten** SQLite-Treiber und eine echte Datei: anders
// als der Stolpertreiber gibt er keine erfundenen Antworten, sondern lässt
// jede Anweisung durch — bis auf die eine, deren Text das Muster enthält.
// Die bricht mit einem `panic` ab, **bevor** sie den Treiber erreicht. Was
// danach mit Verbindung und Sperre geschieht, ist dann das echte Verhalten
// von `database/sql` und SQLite.
//
// **Was er nicht nachbildet**: einen Zusammenbruch **im** Treiber, mitten in
// einer Anweisung. Das ist BK-03a; hier geht es darum, wer die Transaktion
// zurückgibt, wenn ihr Besitzer zusammenbricht.

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	sqlite3 "github.com/ncruces/go-sqlite3"
	treiber "github.com/ncruces/go-sqlite3/driver"
	"github.com/ncruces/go-sqlite3/ext/fts5"
)

// panikausloeser sagt, bei welcher Anweisung zusammengebrochen wird.
type panikausloeser struct {
	muster string
	scharf atomic.Bool
}

func (a *panikausloeser) pruefe(anweisung string) {
	if a.scharf.Load() && strings.Contains(anweisung, a.muster) {
		a.scharf.Store(false)
		panic("planmäßiger Zusammenbruch vor: " + a.muster)
	}
}

type panikverbinder struct {
	innen     driver.Connector
	ausloeser *panikausloeser
}

// Connect öffnet eine echte Verbindung und meldet die Volltextsuche an —
// dasselbe, was `oeffner` je Verbindung tut; ohne sie scheiterten die
// Auslöser der Suchtafeln.
func (v panikverbinder) Connect(ctx context.Context) (driver.Conn, error) {
	verbindung, err := v.innen.Connect(ctx)
	if err != nil {
		return nil, err
	}

	roh, ok := verbindung.(interface{ Raw() *sqlite3.Conn })
	if !ok {
		return nil, errors.Join(errors.New("der Treiber gibt keine rohe Verbindung"), verbindung.Close())
	}

	if err := fts5.Register(roh.Raw()); err != nil {
		return nil, errors.Join(err, verbindung.Close())
	}

	beginn, ok1 := verbindung.(driver.ConnBeginTx)
	ausfuehrer, ok2 := verbindung.(driver.ExecerContext)
	vorbereiter, ok3 := verbindung.(driver.ConnPrepareContext)
	pruefer, ok4 := verbindung.(driver.NamedValueChecker)

	if !ok1 || !ok2 || !ok3 || !ok4 {
		return nil, errors.Join(errors.New("der Treiber kann nicht, was der Paniktreiber weiterreicht"), verbindung.Close())
	}

	return &panikverbindung{
		Conn: verbindung, beginn: beginn, ausfuehrer: ausfuehrer,
		vorbereiter: vorbereiter, pruefer: pruefer, ausloeser: v.ausloeser,
	}, nil
}

func (v panikverbinder) Driver() driver.Driver { return v.innen.Driver() }

type panikverbindung struct {
	driver.Conn
	beginn      driver.ConnBeginTx
	ausfuehrer  driver.ExecerContext
	vorbereiter driver.ConnPrepareContext
	pruefer     driver.NamedValueChecker
	ausloeser   *panikausloeser
}

func (p *panikverbindung) BeginTx(ctx context.Context, wahl driver.TxOptions) (driver.Tx, error) {
	return p.beginn.BeginTx(ctx, wahl)
}

func (p *panikverbindung) ExecContext(
	ctx context.Context, anweisung string, werte []driver.NamedValue,
) (driver.Result, error) {
	p.ausloeser.pruefe(anweisung)

	return p.ausfuehrer.ExecContext(ctx, anweisung, werte)
}

func (p *panikverbindung) PrepareContext(ctx context.Context, anweisung string) (driver.Stmt, error) {
	p.ausloeser.pruefe(anweisung)

	return p.vorbereiter.PrepareContext(ctx, anweisung)
}

func (p *panikverbindung) CheckNamedValue(wert *driver.NamedValue) error {
	return p.pruefer.CheckNamedValue(wert)
}

// panikbestand öffnet die Datei `pfad` über den Paniktreiber — mit den
// Pragmas des Betriebs und **einer** Verbindung, wie `oeffnenMit` sie setzt.
func panikbestand(t *testing.T, pfad string, ausloeser *panikausloeser) *sql.DB {
	t.Helper()

	innen, err := (&treiber.SQLite{}).OpenConnector(mitPragmas(alsQuelle("file:" + pfad)))
	if err != nil {
		t.Fatalf("der Verbinder ließ sich nicht bauen: %v", err)
	}

	db := sql.OpenDB(panikverbinder{innen: innen, ausloeser: ausloeser})
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })

	return db
}
