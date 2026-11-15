package storage

// **Ein Zusammenbruch im Treiber verwirft die Verbindung und gibt den Vorrat
// frei** — BK-03a und BK-03b
// (#2127,
// #2128,
// Vermerk 0175).
//
// Je Stelle eine echte Datei, eine Verbindung im Vorrat wie im Betrieb und
// ein Zusammenbruch zwischen Treiber und Fanghülle. Danach gilt: kein
// `panic` dringt nach oben, der Vorrat ist frei, die Verbindung ist
// verworfen — der nächste Ruf öffnet eine neue — und er gelingt.

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	treiber "github.com/ncruces/go-sqlite3/driver"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gezaehlteSchicht zählt, wie oft eine Verbindung entsteht.
type gezaehlteSchicht struct {
	stolperschicht
	verbunden *atomic.Int32
}

func (s gezaehlteSchicht) Connect(ctx context.Context) (driver.Conn, error) {
	s.verbunden.Add(1)

	return s.stolperschicht.Connect(ctx)
}

// zusammenbruchsbestand legt einen Bestand mit Set 1 an und öffnet ihn über
// Stolperschicht und Fanghülle — mit `schluessel` verschlüsselt.
func zusammenbruchsbestand(t *testing.T, plan *stolperplan, schluessel string) (*sql.DB, *atomic.Int32) {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "bestand.sqlite")

	anlage, err := OeffnenMitSchluessel(t.Context(), "file:"+pfad, schluessel)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if migrierfehler := Migriere(probekontext(t), anlage, nil); migrierfehler != nil {
		t.Fatalf("migrieren: %v", migrierfehler)
	}

	if _, einfuegefehler := anlage.ExecContext(t.Context(),
		`INSERT INTO sets (id, name, spiel, geoeffnet, profil_id) VALUES (1, 'Saat', 'magic', '2026-10-07', 1)`); einfuegefehler != nil {
		t.Fatal(einfuegefehler)
	}

	if schliessfehler := anlage.Close(); schliessfehler != nil {
		t.Fatal(schliessfehler)
	}

	innen, err := (&treiber.SQLite{}).OpenConnector(mitHuelle(mitPragmas(alsQuelle("file:"+pfad)), schluessel))
	if err != nil {
		t.Fatal(err)
	}

	verbunden := &atomic.Int32{}
	vorbereiten := vorbereitenMit(schluessel)
	db := sql.OpenDB(fangverbinder{
		innen: gezaehlteSchicht{stolperschicht: stolperschicht{innen: innen, plan: plan}, verbunden: verbunden},
		vorbereiten: func(verbindung driver.Conn) error {
			plan.bricht(stolpertBeimEinrichten)

			return vorbereiten(verbindung)
		},
	})
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })

	return db, verbunden
}

// lies liest Set 1 über eine Abfrage mit Platzhalter — Vorbereiten, Fragen,
// Weiterlesen, Spaltenlesen und Schließen in einem Weg.
func lies(ctx context.Context, db *sql.DB) (fehler error) {
	zeilen, err := db.QueryContext(ctx, `SELECT name FROM sets WHERE id = ?`, 1)
	if err != nil {
		return err
	}

	// **Der Schließfehler zählt** — der Fall „Rows.Close" bricht genau dort.
	defer func() { fehler = errors.Join(fehler, zeilen.Close()) }()

	for zeilen.Next() {
		var name string
		if err := zeilen.Scan(&name); err != nil {
			return err
		}
	}

	return zeilen.Err()
}

func zusammenbruchsfaelle() map[stolperstelleTreiber]func(context.Context, *sql.DB) error {
	ausfuehren := func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, `UPDATE sets SET name = 'Ernte' WHERE id = 1`)

		return err
	}
	anweisen := func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, `UPDATE sets SET name = ? WHERE id = 1`, "Ernte")

		return err
	}
	geschaeft := func(abschliessen func(*sql.Tx) error) func(context.Context, *sql.DB) error {
		return func(ctx context.Context, db *sql.DB) error {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				return err
			}

			if _, err := tx.ExecContext(ctx, `UPDATE sets SET name = 'Ernte' WHERE id = 1`); err != nil {
				return errors.Join(err, tx.Rollback())
			}

			return abschliessen(tx)
		}
	}

	return map[stolperstelleTreiber]func(context.Context, *sql.DB) error{
		stolpertBeimAusfuehren:     ausfuehren,
		stolpertBeimVorbereiten:    lies,
		stolpertBeimFragen:         lies,
		stolpertBeimWeiterlesen:    lies,
		stolpertBeimSpaltenlesen:   lies,
		stolpertBeimZeilenende:     lies,
		stolpertBeimAnweisen:       anweisen,
		stolpertBeimAnweisungsende: anweisen,
		stolpertBeimBeginnen:       geschaeft(func(tx *sql.Tx) error { return tx.Commit() }),
		stolpertBeimBestaetigen:    geschaeft(func(tx *sql.Tx) error { return tx.Commit() }),
		stolpertBeimZuruecknehmen:  geschaeft(func(tx *sql.Tx) error { return tx.Rollback() }),
	}
}

func TestEinZusammenbruchImTreiberVerwirftDieVerbindung(t *testing.T) {
	for stelle, tun := range zusammenbruchsfaelle() {
		t.Run(string(stelle), func(t *testing.T) {
			plan := &stolperplan{stelle: stelle}
			db, verbunden := zusammenbruchsbestand(t, plan, "")

			// **Erst ein gewöhnlicher Ruf** — die Verbindung steht, und der
			// Zähler beginnt bei eins.
			if err := lies(t.Context(), db); err != nil {
				t.Fatalf("vorher: %v", err)
			}

			plan.scharf.Store(true)

			var fehler error
			if gefangen := bricht(func() { fehler = tun(t.Context(), db) }); gefangen != nil {
				t.Fatalf("der Zusammenbruch drang nach oben: %v", gefangen)
			}

			if plan.scharf.Load() {
				t.Fatalf("die Stelle %s wurde nie erreicht (%v)", stelle, fehler)
			}

			t.Logf("Endstatus: %v · belegt %d", fehler, db.Stats().InUse)

			if fehler != nil && !errors.Is(fehler, domain.ErrZusammengebrochen) {
				t.Errorf("gemeldet wurde %v statt eines Zusammenbruchs", fehler)
			}

			if belegt := db.Stats().InUse; belegt != 0 {
				t.Errorf("nach dem Zusammenbruch sind %d Verbindungen belegt", belegt)
			}

			kurz, ende := context.WithTimeout(t.Context(), time.Second)
			defer ende()

			if err := lies(kurz, db); err != nil {
				t.Errorf("der nächste Ruf: %v", err)
			}

			if neu := verbunden.Load(); neu != 2 {
				t.Errorf("es entstanden %d Verbindungen — die verbrauchte wurde nicht verworfen", neu)
			}

			// **Verworfen heißt geschlossen** — auch für SQLite selbst, nicht
			// nur für den Vorrat. Eine Anweisung, die nicht finalisiert ist,
			// hielte die alte Verbindung samt Datei offen.
			if anzahl, fehler := plan.schliessfehler(); anzahl != 1 || fehler != nil {
				t.Errorf("geschlossen %d, dabei %v", anzahl, fehler)
			}
		})
	}
}

// **Ein Zusammenbruch, bevor die Verbindung steht** — im Verbinden selbst
// oder beim Setzen von Schlüssel und Pragmas. `database/sql` zählt die
// Verbindung vor `Connect` mit und nimmt sie nur bei einem **Fehler** zurück;
// ein `panic` hielte den einzigen Platz für immer.
func TestEinZusammenbruchBeimVerbindenGibtDenPlatzZurueck(t *testing.T) {
	for stelle, geschlossen := range map[stolperstelleTreiber]int{
		stolpertBeimVerbinden:  0,
		stolpertBeimEinrichten: 1,
	} {
		t.Run(string(stelle), func(t *testing.T) {
			plan := &stolperplan{stelle: stelle}
			db, verbunden := zusammenbruchsbestand(t, plan, NeuerSchluessel())

			plan.scharf.Store(true)

			var fehler error
			if gefangen := bricht(func() { fehler = lies(t.Context(), db) }); gefangen != nil {
				t.Fatalf("der Zusammenbruch drang nach oben: %v", gefangen)
			}

			if !errors.Is(fehler, domain.ErrZusammengebrochen) {
				t.Errorf("gemeldet wurde %v statt eines Zusammenbruchs", fehler)
			}

			if offen := db.Stats().OpenConnections; offen != 0 {
				t.Errorf("nach dem Zusammenbruch zählt der Vorrat %d Verbindungen", offen)
			}

			kurz, ende := context.WithTimeout(t.Context(), time.Second)
			defer ende()

			if err := lies(kurz, db); err != nil {
				t.Errorf("der nächste Ruf: %v", err)
			}

			if neu := verbunden.Load(); neu != 2 {
				t.Errorf("es entstanden %d Verbindungen", neu)
			}

			if anzahl, schliessfehler := plan.schliessfehler(); anzahl != geschlossen || schliessfehler != nil {
				t.Errorf("geschlossen %d statt %d, dabei %v", anzahl, geschlossen, schliessfehler)
			}
		})
	}
}

// **Ein Schreiben, das schon wirkte, wird nicht wiederholt** — darum ist der
// Zusammenbruch kein `driver.ErrBadConn`.
func TestEinZusammenbruchNachDemSchreibenWiederholtEsNicht(t *testing.T) {
	plan := &stolperplan{stelle: stolpertNachDemAusfuehren}
	db, _ := zusammenbruchsbestand(t, plan, "")

	plan.scharf.Store(true)

	_, err := db.ExecContext(t.Context(),
		`INSERT INTO sets (name, spiel, geoeffnet, profil_id) VALUES ('Zweit', 'magic', '2026-10-07', 1)`)
	if !errors.Is(err, domain.ErrZusammengebrochen) {
		t.Fatalf("erwartet war ein Zusammenbruch, gekommen ist %v", err)
	}

	var anzahl int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM sets WHERE name = 'Zweit'`).Scan(&anzahl); err != nil {
		t.Fatal(err)
	}

	if anzahl != 1 {
		t.Errorf("das Set steht %d-mal da", anzahl)
	}
}

// **Die neue Verbindung trägt den Schlüssel** — ein verschlüsselter Bestand
// liest sich nach dem Zusammenbruch wie davor.
func TestNachDemZusammenbruchTraegtDieNeueVerbindungDenSchluessel(t *testing.T) {
	plan := &stolperplan{stelle: stolpertBeimFragen}
	db, verbunden := zusammenbruchsbestand(t, plan, NeuerSchluessel())

	plan.scharf.Store(true)

	if err := lies(t.Context(), db); !errors.Is(err, domain.ErrZusammengebrochen) {
		t.Fatalf("erwartet war ein Zusammenbruch, gekommen ist %v", err)
	}

	var name string
	if err := db.QueryRowContext(t.Context(), `SELECT name FROM sets WHERE id = 1`).Scan(&name); err != nil || name != "Saat" {
		t.Errorf("nach dem Zusammenbruch: %q (%v)", name, err)
	}

	if neu := verbunden.Load(); neu != 2 {
		t.Errorf("es entstanden %d Verbindungen", neu)
	}
}
