package storage

// **Eine Schicht, die an einer benannten Stelle des Treibers zusammenbricht**
// — für BK-03a und BK-03b
// (#2127,
// #2128).
//
// Sie liegt **zwischen** dem echten Treiber und der Fanghülle: was darunter
// geschieht, ist SQLite; was darüber geschieht, ist `fangtreiber.go` und
// `database/sql`. Ein Zusammenbruch **im** Treiber selbst — erschöpfter
// Speicher der WASM-Laufzeit — ist damit nachgebildet, nicht herbeigeführt.

import (
	"context"
	"database/sql/driver"
	"errors"
	"sync"
	"sync/atomic"
)

type stolperstelleTreiber string

const (
	stolpertBeimAusfuehren     stolperstelleTreiber = "Exec"
	stolpertBeimVorbereiten    stolperstelleTreiber = "Prepare"
	stolpertBeimFragen         stolperstelleTreiber = "Query"
	stolpertBeimAnweisen       stolperstelleTreiber = "Stmt.Exec"
	stolpertBeimAnweisungsende stolperstelleTreiber = "Stmt.Close"
	stolpertBeimWeiterlesen    stolperstelleTreiber = "Rows.Next"
	stolpertBeimSpaltenlesen   stolperstelleTreiber = "Rows.Scan"
	stolpertBeimZeilenende     stolperstelleTreiber = "Rows.Close"
	stolpertBeimBestaetigen    stolperstelleTreiber = "Tx.Commit"
	stolpertBeimZuruecknehmen  stolperstelleTreiber = "Tx.Rollback"
	stolpertBeimBeginnen       stolperstelleTreiber = "BeginTx"
	stolpertNachDemAusfuehren  stolperstelleTreiber = "Exec danach"
	stolpertBeimVerbinden      stolperstelleTreiber = "Connect"
	stolpertBeimEinrichten     stolperstelleTreiber = "Vorbereiten"
)

type stolperplan struct {
	stelle stolperstelleTreiber
	scharf atomic.Bool

	// geschlossen hält, was das Schließen jeder Treiberverbindung ergab —
	// `database/sql` verwirft diesen Fehler, und ein Rohgriff, der offen
	// bleibt, sähe von oben gesund aus.
	mu          sync.Mutex
	geschlossen []error
}

func (p *stolperplan) schliessfehler() (anzahl int, fehler error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return len(p.geschlossen), errors.Join(p.geschlossen...)
}

// bricht wirft, wenn `stelle` gemeint und die Schicht scharf ist — einmal.
func (p *stolperplan) bricht(stelle stolperstelleTreiber) {
	if p.stelle == stelle && p.scharf.CompareAndSwap(true, false) {
		panic("planmäßiger Zusammenbruch: " + string(stelle))
	}
}

type stolperschicht struct {
	innen driver.Connector
	plan  *stolperplan
}

func (s stolperschicht) Connect(ctx context.Context) (driver.Conn, error) {
	s.plan.bricht(stolpertBeimVerbinden)

	verbindung, err := s.innen.Connect(ctx)
	if err != nil {
		return nil, err
	}

	voll, ok := verbindung.(vollverbindung)
	if !ok {
		return verbindung, nil
	}

	return &stolperverbindung{vollverbindung: voll, plan: s.plan}, nil
}

func (s stolperschicht) Driver() driver.Driver { return s.innen.Driver() }

type stolperverbindung struct {
	vollverbindung
	plan *stolperplan
}

func (v *stolperverbindung) Close() error {
	fehler := v.vollverbindung.Close()

	v.plan.mu.Lock()
	v.plan.geschlossen = append(v.plan.geschlossen, fehler)
	v.plan.mu.Unlock()

	return fehler
}

func (v *stolperverbindung) ExecContext(
	ctx context.Context, anweisung string, werte []driver.NamedValue,
) (driver.Result, error) {
	v.plan.bricht(stolpertBeimAusfuehren)

	ergebnis, err := v.vollverbindung.ExecContext(ctx, anweisung, werte)
	if err == nil {
		v.plan.bricht(stolpertNachDemAusfuehren)
	}

	return ergebnis, err
}

func (v *stolperverbindung) PrepareContext(ctx context.Context, anweisung string) (driver.Stmt, error) {
	v.plan.bricht(stolpertBeimVorbereiten)

	innen, err := v.vollverbindung.PrepareContext(ctx, anweisung)
	if err != nil {
		return nil, err
	}

	voll, ok := innen.(vollanweisung)
	if !ok {
		return innen, nil
	}

	return &stolperanweisung{vollanweisung: voll, plan: v.plan}, nil
}

func (v *stolperverbindung) BeginTx(ctx context.Context, wahl driver.TxOptions) (driver.Tx, error) {
	v.plan.bricht(stolpertBeimBeginnen)

	innen, err := v.vollverbindung.BeginTx(ctx, wahl)
	if err != nil {
		return nil, err
	}

	return stolpergeschaeft{Tx: innen, plan: v.plan}, nil
}

type stolperanweisung struct {
	vollanweisung
	plan *stolperplan
}

func (a *stolperanweisung) ExecContext(ctx context.Context, werte []driver.NamedValue) (driver.Result, error) {
	a.plan.bricht(stolpertBeimAnweisen)

	return a.vollanweisung.ExecContext(ctx, werte)
}

func (a *stolperanweisung) QueryContext(ctx context.Context, werte []driver.NamedValue) (driver.Rows, error) {
	a.plan.bricht(stolpertBeimFragen)

	innen, err := a.vollanweisung.QueryContext(ctx, werte)
	if err != nil {
		return nil, err
	}

	voll, ok := innen.(vollzeilen)
	if !ok {
		return innen, nil
	}

	return &stolperzeilen{vollzeilen: voll, plan: a.plan}, nil
}

func (a *stolperanweisung) Close() error {
	a.plan.bricht(stolpertBeimAnweisungsende)

	return a.vollanweisung.Close()
}

type stolperzeilen struct {
	vollzeilen
	plan *stolperplan
}

func (z *stolperzeilen) NextRow() error {
	z.plan.bricht(stolpertBeimWeiterlesen)

	return z.vollzeilen.NextRow()
}

func (z *stolperzeilen) ScanColumn(spur driver.ScanContext, stelle int, ziel any) error {
	z.plan.bricht(stolpertBeimSpaltenlesen)

	return z.vollzeilen.ScanColumn(spur, stelle, ziel)
}

func (z *stolperzeilen) Close() error {
	z.plan.bricht(stolpertBeimZeilenende)

	return z.vollzeilen.Close()
}

type stolpergeschaeft struct {
	driver.Tx
	plan *stolperplan
}

func (g stolpergeschaeft) Commit() error {
	g.plan.bricht(stolpertBeimBestaetigen)

	return g.Tx.Commit()
}

func (g stolpergeschaeft) Rollback() error {
	g.plan.bricht(stolpertBeimZuruecknehmen)

	return g.Tx.Rollback()
}
