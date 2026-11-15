package storage

// **Die Ränder der Fanghülle** — BK-03b
// (#2128).
//
// Der echte Treiber kann alles, was die Hülle weiterreicht; ein Treiber, der
// es nicht kann, bekommt seine Verbindungen, Anweisungen und Zeilen
// **unverhüllt** zurück, statt eine Fähigkeit vorgespielt zu bekommen. Diese
// Wege betritt nur ein fremder Treiber — deshalb hier mit kargen Doppeln.

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	sqlite3 "github.com/ncruces/go-sqlite3"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

var errKarg = errors.New("planmäßig: das karge Doppel kann das nicht")

// karg ist eine Verbindung, eine Anweisung und eine Zeilenfolge, die nur das
// Nötigste kann.
type karg struct{}

func (karg) Prepare(string) (driver.Stmt, error) { return nil, errKarg }
func (karg) Close() error                        { return nil }
func (karg) Begin() (driver.Tx, error)           { return nil, errKarg }
func (karg) NumInput() int                       { return 0 }
func (karg) Exec([]driver.Value) (driver.Result, error) {
	return nil, errKarg
}
func (karg) Query([]driver.Value) (driver.Rows, error) { return nil, errKarg }
func (karg) Columns() []string                         { return nil }
func (karg) Next([]driver.Value) error                 { return errKarg }

type kargerVerbinder struct{ fehler error }

func (v kargerVerbinder) Connect(context.Context) (driver.Conn, error) {
	if v.fehler != nil {
		return nil, v.fehler
	}

	return karg{}, nil
}

func (kargerVerbinder) Driver() driver.Driver { return nil }

// halbvoll kann, was eine Verbindung können muss, gibt aber karge
// Anweisungen.
type halbvoll struct{ karg }

func (halbvoll) Raw() *sqlite3.Conn { return nil }
func (halbvoll) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return nil, errKarg
}
func (halbvoll) PrepareContext(context.Context, string) (driver.Stmt, error) { return karg{}, nil }
func (halbvoll) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return nil, errKarg
}
func (halbvoll) CheckNamedValue(*driver.NamedValue) error { return nil }

// volleAnweisung kann, was eine Anweisung können muss, und antwortet auf
// eine Abfrage mit `fehler` oder kargen Zeilen.
type volleAnweisung struct {
	karg
	fehler error
}

func (volleAnweisung) ExecContext(context.Context, []driver.NamedValue) (driver.Result, error) {
	return nil, errKarg
}

func (a volleAnweisung) QueryContext(context.Context, []driver.NamedValue) (driver.Rows, error) {
	if a.fehler != nil {
		return nil, a.fehler
	}

	return karg{}, nil
}

func (volleAnweisung) CheckNamedValue(*driver.NamedValue) error { return nil }

func TestDieFanghuelleReichtWeiterWasSieNichtKennt(t *testing.T) {
	ohne := func(driver.Conn) error { return nil }

	t.Run("eine Verbindung, die nicht entsteht", func(t *testing.T) {
		if _, err := (fangverbinder{innen: kargerVerbinder{fehler: errKarg}, vorbereiten: ohne}).Connect(t.Context()); !errors.Is(err, errKarg) {
			t.Errorf("gekommen ist %v", err)
		}
	})

	t.Run("eine Verbindung, die nicht von SQLite ist", func(t *testing.T) {
		_, err := (fangverbinder{innen: kargerVerbinder{}, vorbereiten: vorbereitenMit("")}).Connect(t.Context())
		if err == nil || !strings.Contains(err.Error(), "keine von SQLite") {
			t.Errorf("gekommen ist %v", err)
		}
	})

	t.Run("eine Verbindung, die zu wenig kann, bleibt unverhüllt", func(t *testing.T) {
		verbindung, err := (fangverbinder{innen: kargerVerbinder{}, vorbereiten: ohne}).Connect(t.Context())
		if _, verhuellt := verbindung.(*fangverbindung); err != nil || verhuellt {
			t.Errorf("verhüllt %v, Fehler %v", verhuellt, err)
		}
	})

	t.Run("der Treiber dahinter", func(t *testing.T) {
		if treiber := (fangverbinder{innen: kargerVerbinder{}}).Driver(); treiber != nil {
			t.Errorf("gekommen ist %v", treiber)
		}
	})

	t.Run("eine Anweisung, die zu wenig kann, bleibt unverhüllt", func(t *testing.T) {
		anweisung, err := (&fangverbindung{vollverbindung: halbvoll{}}).PrepareContext(t.Context(), "SELECT 1")
		if _, verhuellt := anweisung.(*fanganweisung); err != nil || verhuellt {
			t.Errorf("verhüllt %v, Fehler %v", verhuellt, err)
		}
	})

	t.Run("eine Abfrage, die scheitert", func(t *testing.T) {
		a := &fanganweisung{vollanweisung: volleAnweisung{fehler: errKarg}, verbindung: &fangverbindung{}}
		zeilen, err := a.QueryContext(t.Context(), nil)
		if zeilen != nil || !errors.Is(err, errKarg) {
			t.Errorf("gekommen ist %v, %v", zeilen, err)
		}
	})

	t.Run("Zeilen, die zu wenig können, bleiben unverhüllt", func(t *testing.T) {
		a := &fanganweisung{vollanweisung: volleAnweisung{}, verbindung: &fangverbindung{}}
		zeilen, err := a.QueryContext(t.Context(), nil)
		if _, verhuellt := zeilen.(*fangzeilen); err != nil || verhuellt {
			t.Errorf("verhüllt %v, Fehler %v", verhuellt, err)
		}
	})

	t.Run("ein Schließen, das selbst zusammenbricht", func(t *testing.T) {
		if err := gefangen(func() error { panic("planmäßig") }); !errors.Is(err, domain.ErrZusammengebrochen) {
			t.Errorf("gekommen ist %v", err)
		}
	})

	t.Run("eine Datenquelle, die der Treiber abweist", func(t *testing.T) {
		if _, err := oeffner("file:x.db?_txlock=quatsch", ""); err == nil {
			t.Error("kein Fehler")
		}
	})
}
