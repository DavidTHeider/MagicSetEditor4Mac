// Die Fehlerpfade von SetAnlegen, die eine echte Datenbank nicht erzeugt.
//
// Ein Insert, das mitten in einer gültigen Folge scheitert, und ein Commit, das
// nach lauter geglückten Schritten fehlschlägt — beides kommt im Betrieb vor
// (volle Platte, abgebrochene Verbindung) und mit SQLite auf einer gesunden
// Platte nie. Ein Fehlerpfad, den kein Test betritt, ist eine Behauptung
// darüber, was im Ernstfall geschieht.
package storage

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

var errZaehlerStolper = errors.New("der Zähltreiber ist planmäßig gescheitert")

// zaehlerstelle sagt, ab welchem Aufruf welcher Art gescheitert wird.
var (
	stolperAb    atomic.Int32
	stolperWobei atomic.Value // string: "insert-cards", "insert-raw", "commit"
)

func init() { sql.Register("zaehler", zaehlertreiber{}) }

type zaehlertreiber struct{}

func (zaehlertreiber) Open(string) (driver.Conn, error) { return &zaehlerconn{}, nil }

type zaehlerconn struct{ gesehen int32 }

func (c *zaehlerconn) Prepare(anweisung string) (driver.Stmt, error) {
	return &zaehlerstmt{conn: c, anweisung: anweisung}, nil
}
func (c *zaehlerconn) Close() error              { return nil }
func (c *zaehlerconn) Begin() (driver.Tx, error) { return zaehlertx{}, nil }

type zaehlerstmt struct {
	conn      *zaehlerconn
	anweisung string
}

func (s *zaehlerstmt) Close() error  { return nil }
func (s *zaehlerstmt) NumInput() int { return -1 }

// stolpert sagt, ob dieser Aufruf scheitern soll.
func (s *zaehlerstmt) stolpert(marke string) bool {
	if eingestellteStelle() != marke {
		return false
	}

	s.conn.gesehen++

	return s.conn.gesehen >= stolperAb.Load()
}

func (s *zaehlerstmt) Exec([]driver.Value) (driver.Result, error) {
	if strings.Contains(s.anweisung, "raw_blocks") && s.stolpert("insert-raw") {
		return nil, errZaehlerStolper
	}

	if strings.Contains(s.anweisung, "set_quellen") && s.stolpert("insert-quelle") {
		return nil, errZaehlerStolper
	}

	return driver.RowsAffected(1), nil
}

func (s *zaehlerstmt) Query([]driver.Value) (driver.Rows, error) {
	if strings.Contains(s.anweisung, "INTO cards") && s.stolpert("insert-cards") {
		return nil, errZaehlerStolper
	}

	return &eineZeile{}, nil
}

// eineZeile liefert genau eine Kennung — das, was `RETURNING id` ergibt.
type eineZeile struct{ gelesen bool }

func (r *eineZeile) Columns() []string { return []string{"id"} }
func (r *eineZeile) Close() error      { return nil }

func (r *eineZeile) Next(ziel []driver.Value) error {
	if r.gelesen {
		return io.EOF
	}

	r.gelesen = true
	ziel[0] = int64(1)

	return nil
}

type zaehlertx struct{}

func (zaehlertx) Rollback() error { return nil }

func (zaehlertx) Commit() error {
	if eingestellteStelle() == "commit" {
		return errZaehlerStolper
	}

	return nil
}

// eingestellteStelle liest die eingestellte Stelle; ohne Einstellung ist sie leer.
func eingestellteStelle() string {
	wobei, gesetzt := stolperWobei.Load().(string)
	if !gesetzt {
		return ""
	}

	return wobei
}

// ------------------------------------------------------------------- Prüfungen

// mitStolper gibt eine Datenbank, deren erster passender Aufruf scheitert.
//
// **Ohne Zählerparameter**: er war immer `1`, und ein Parameter, den niemand
// anders belegt, ist eine Zusage über eine Freiheit, die es nicht gibt.
func mitStolper(t *testing.T, wobei string) *sql.DB {
	t.Helper()

	stolperWobei.Store(wobei)
	stolperAb.Store(1)
	t.Cleanup(func() { stolperWobei.Store("") })

	db, err := sql.Open("zaehler", "")
	if err != nil {
		t.Fatalf("Zähltreiber: %v", err)
	}
	t.Cleanup(func() {
		if fehler := db.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	return db
}

func einfachesSet() *domain.Set {
	return &domain.Set{
		Name: "Probe", Game: "probe",
		Cards: []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Erste"}}}},
	}
}

func TestSetAnlegenMeldetEineGescheiterteKarte(t *testing.T) {
	db := mitStolper(t, "insert-cards")

	if _, err := NeuerSetSpeicher(db).SetAnlegen(probekontext(t), einfachesSet(), nil); !errors.Is(err, errZaehlerStolper) {
		t.Errorf("Fehler %v, erwartet war der Stolperfehler", err)
	}
}

func TestSetAnlegenMeldetEinenGescheitertenRohblock(t *testing.T) {
	db := mitStolper(t, "insert-raw")

	_, err := NeuerSetSpeicher(db).SetAnlegen(probekontext(t), einfachesSet(),
		[]domain.RawBlock{{File: "set", Line: 3, Text: "unverstanden"}})

	if !errors.Is(err, errZaehlerStolper) {
		t.Errorf("Fehler %v, erwartet war der Stolperfehler", err)
	}
	if err != nil && !strings.Contains(err.Error(), "Zeile 3") {
		t.Errorf("die Meldung %q nennt die Zeile nicht", err.Error())
	}
}

// **Auch die Quelldatei kann scheitern** — und dann ist kein halbes Set
// entstanden: sie wird in derselben Transaktion geschrieben
// ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)).
func TestSetAnlegenMeldetEineGescheiterteQuelle(t *testing.T) {
	db := mitStolper(t, "insert-quelle")

	set := einfachesSet()
	set.Quelle = &domain.Quelle{Datei: "set", Inhalt: []byte("mse version: 2.0.0\n")}

	if _, err := NeuerSetSpeicher(db).SetAnlegen(probekontext(t), set, nil); !errors.Is(err, errZaehlerStolper) {
		t.Errorf("Fehler %v, erwartet war der Stolperfehler", err)
	}
}

func TestSetAnlegenMeldetEinGescheitertesCommit(t *testing.T) {
	db := mitStolper(t, "commit")

	if _, err := NeuerSetSpeicher(db).SetAnlegen(probekontext(t), einfachesSet(), nil); !errors.Is(err, errZaehlerStolper) {
		t.Errorf("Fehler %v, erwartet war der Stolperfehler", err)
	}
}
