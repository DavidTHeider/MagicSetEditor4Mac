package storage

// **Ein Treiber, der seinen Zusammenbruch fängt** — BK-03a und BK-03b
// (#2127,
// #2128,
// Vermerk 0175).
//
// # Der Befund
//
// `database/sql` fängt keinen `panic` des Treibers. Bricht er in einer
// Anweisung zusammen, gibt niemand die Verbindung an den Vorrat zurück — und
// der Vorrat hat **eine** (`SetMaxOpenConns(1)`): jeder weitere Ruf wartete
// bis zu seiner Frist, auch nachdem Wails oder `domain.Auffangen` den
// Zusammenbruch gefangen hatten. Der Treiber kennt solche Stellen: er wirft
// bei erschöpftem Speicher der WASM-Laufzeit (`errutil.OOMErr`) und bei
// verletzten Zusicherungen.
//
// # Die Hülle
//
// Sie legt sich um jede Verbindung, Anweisung, Zeilenfolge und Transaktion,
// fängt den Zusammenbruch und gibt ihn als `domain.ErrZusammengebrochen`
// zurück. Die Verbindung gilt danach als **verbraucht** (`IsValid`), und
// `database/sql` verwirft sie, statt sie wieder auszugeben; die nächste
// entsteht frisch, mit Schlüssel und Pragmas.
//
// **Bewusst nicht `driver.ErrBadConn`**: darauf wiederholt `database/sql` den
// Ruf auf einer neuen Verbindung — und ein Schreiben, das schon wirkte, stünde
// zweimal im Bestand.

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"sync/atomic"

	treiber "github.com/ncruces/go-sqlite3/driver"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// fangverbinder öffnet Verbindungen über `innen` und hüllt sie ein.
type fangverbinder struct {
	innen driver.Connector

	// vorbereiten setzt, was je Verbindung gilt — Schlüssel, Pragmas,
	// Volltextsuche ([oeffner]).
	vorbereiten func(driver.Conn) error
}

// Connect öffnet eine Verbindung und bereitet sie vor — und fängt einen
// Zusammenbruch auch **dabei**.
//
// **Gerade hier zählt es**: `database/sql` zählt die Verbindung vor dem Ruf
// mit und nimmt sie nur bei einem zurückgegebenen Fehler wieder heraus. Ein
// `panic` hielte den einzigen Platz des Vorrats für immer (Zweitblick über C
// und D, Befund 2). Eine schon geöffnete Verbindung wird dabei geschlossen.
func (v fangverbinder) Connect(ctx context.Context) (_ driver.Conn, fehler error) {
	var verbindung driver.Conn

	defer func() {
		if wert := recover(); wert != nil {
			fehler = fmt.Errorf("%w: beim Öffnen einer Verbindung: %v", domain.ErrZusammengebrochen, wert)

			if verbindung != nil {
				fehler = errors.Join(fehler, gefangen(verbindung.Close))
			}
		}
	}()

	verbindung, err := v.innen.Connect(ctx)
	if err != nil {
		return nil, err
	}

	if err := v.vorbereiten(verbindung); err != nil {
		return nil, errors.Join(err, verbindung.Close())
	}

	return umhuellen(verbindung), nil
}

// gefangen ruft `tun` und macht aus einem Zusammenbruch einen Fehler.
func gefangen(tun func() error) (fehler error) {
	defer func() {
		if wert := recover(); wert != nil {
			fehler = fmt.Errorf("%w: %v", domain.ErrZusammengebrochen, wert)
		}
	}()

	return tun()
}

func (v fangverbinder) Driver() driver.Driver { return v.innen.Driver() }

// vollverbindung ist, was der Treiber an einer Verbindung kann und die Hülle
// weiterreicht.
//
// **Samt `Raw`** — wer über `sql.Conn.Raw` an die Verbindung von SQLite will
// (der Zähler der Abfragepläne etwa), bekommt sie durch die Hülle hindurch.
type vollverbindung interface {
	treiber.Conn
	driver.ExecerContext
	driver.NamedValueChecker
}

// umhuellen gibt eine Verbindung, die kann, was die Hülle weiterreicht, in
// ihrer Hülle zurück — und jede andere, wie sie ist.
func umhuellen(verbindung driver.Conn) driver.Conn {
	voll, ok := verbindung.(vollverbindung)
	if !ok {
		return verbindung
	}

	return &fangverbindung{vollverbindung: voll}
}

type fangverbindung struct {
	vollverbindung
	verbraucht atomic.Bool
}

// zusammenbruch macht aus einem gefangenen Wert den Fehler und verbraucht
// die Verbindung.
func (v *fangverbindung) zusammenbruch(wert any) error {
	v.verbraucht.Store(true)

	return fmt.Errorf("%w: im Datenbanktreiber, die Verbindung wird verworfen: %v",
		domain.ErrZusammengebrochen, wert)
}

// IsValid sagt `database/sql`, ob die Verbindung zurück in den Vorrat darf.
func (v *fangverbindung) IsValid() bool { return !v.verbraucht.Load() }

func (v *fangverbindung) ExecContext(
	ctx context.Context, anweisung string, werte []driver.NamedValue,
) (ergebnis driver.Result, fehler error) {
	defer func() {
		if wert := recover(); wert != nil {
			fehler = v.zusammenbruch(wert)
		}
	}()

	return v.vollverbindung.ExecContext(ctx, anweisung, werte)
}

func (v *fangverbindung) PrepareContext(ctx context.Context, anweisung string) (_ driver.Stmt, fehler error) {
	defer func() {
		if wert := recover(); wert != nil {
			fehler = v.zusammenbruch(wert)
		}
	}()

	innen, err := v.vollverbindung.PrepareContext(ctx, anweisung)
	if err != nil {
		return nil, err
	}

	voll, ok := innen.(vollanweisung)
	if !ok {
		return innen, nil
	}

	return &fanganweisung{vollanweisung: voll, verbindung: v}, nil
}

func (v *fangverbindung) BeginTx(ctx context.Context, wahl driver.TxOptions) (_ driver.Tx, fehler error) {
	defer func() {
		if wert := recover(); wert != nil {
			fehler = v.zusammenbruch(wert)
		}
	}()

	innen, err := v.vollverbindung.BeginTx(ctx, wahl)
	if err != nil {
		return nil, err
	}

	return &fanggeschaeft{Tx: innen, verbindung: v}, nil
}

type vollanweisung interface {
	driver.Stmt
	driver.StmtExecContext
	driver.StmtQueryContext
	driver.NamedValueChecker
}

type fanganweisung struct {
	vollanweisung
	verbindung *fangverbindung
}

func (a *fanganweisung) ExecContext(
	ctx context.Context, werte []driver.NamedValue,
) (_ driver.Result, fehler error) {
	defer func() {
		if wert := recover(); wert != nil {
			fehler = a.verbindung.zusammenbruch(wert)
		}
	}()

	return a.vollanweisung.ExecContext(ctx, werte)
}

func (a *fanganweisung) QueryContext(
	ctx context.Context, werte []driver.NamedValue,
) (_ driver.Rows, fehler error) {
	defer func() {
		if wert := recover(); wert != nil {
			fehler = a.verbindung.zusammenbruch(wert)
		}
	}()

	innen, err := a.vollanweisung.QueryContext(ctx, werte)
	if err != nil {
		return nil, err
	}

	voll, ok := innen.(vollzeilen)
	if !ok {
		return innen, nil
	}

	return &fangzeilen{vollzeilen: voll, verbindung: a.verbindung}, nil
}

// Close finalisiert die Anweisung — und nach einem Zusammenbruch **noch
// einmal**.
//
// **Weil sie sonst die Verbindung festhält**: `database/sql` verwirft die
// verbrauchte Verbindung, aber SQLite schließt keine, an der eine Anweisung
// unfinalisiert hängt — die Datei und der Speicher der Laufzeit blieben,
// während der Vorrat gesund aussähe (Zweitblick über C und D, Befund 5). Eine
// geschlossene Anweisung ein zweites Mal zu schließen, ist im Treiber
// zugesagt harmlos.
func (a *fanganweisung) Close() (fehler error) {
	defer func() {
		if wert := recover(); wert != nil {
			fehler = errors.Join(a.verbindung.zusammenbruch(wert), gefangen(a.vollanweisung.Close))
		}
	}()

	return a.vollanweisung.Close()
}

// vollzeilen ist, was die Zeilenfolge des Treibers kann — samt dem Weg, über
// den `database/sql` seit Go 1.26 liest (`RowsColumnScanner`). Eine Hülle,
// die ihn verschwiege, änderte, wie gelesen wird.
type vollzeilen interface {
	driver.Rows
	driver.RowsColumnScanner
	driver.RowsColumnTypeScanType
	driver.RowsColumnTypeDatabaseTypeName
	driver.RowsColumnTypeNullable
}

type fangzeilen struct {
	vollzeilen
	verbindung *fangverbindung
}

func (z *fangzeilen) NextRow() (fehler error) {
	defer func() {
		if wert := recover(); wert != nil {
			fehler = z.verbindung.zusammenbruch(wert)
		}
	}()

	return z.vollzeilen.NextRow()
}

func (z *fangzeilen) ScanColumn(spur driver.ScanContext, stelle int, ziel any) (fehler error) {
	defer func() {
		if wert := recover(); wert != nil {
			fehler = z.verbindung.zusammenbruch(wert)
		}
	}()

	return z.vollzeilen.ScanColumn(spur, stelle, ziel)
}

func (z *fangzeilen) Close() (fehler error) {
	defer func() {
		if wert := recover(); wert != nil {
			fehler = z.verbindung.zusammenbruch(wert)
		}
	}()

	return z.vollzeilen.Close()
}

type fanggeschaeft struct {
	driver.Tx
	verbindung *fangverbindung
}

func (g *fanggeschaeft) Commit() (fehler error) {
	defer func() {
		if wert := recover(); wert != nil {
			fehler = g.verbindung.zusammenbruch(wert)
		}
	}()

	return g.Tx.Commit()
}

func (g *fanggeschaeft) Rollback() (fehler error) {
	defer func() {
		if wert := recover(); wert != nil {
			fehler = g.verbindung.zusammenbruch(wert)
		}
	}()

	return g.Tx.Rollback()
}
