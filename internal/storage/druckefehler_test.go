package storage_test

// **Jeder Fehlerweg der Druckablage ist betreten** — Werkstück A2 von
// Phase 5.3 (#1441). Was SQLite hergibt, läuft gegen SQLite: Tabelle weg,
// Auslöser, geschlossener Bestand. Was es nicht hergibt, läuft gegen den
// Stolpertreiber (`stolperdrucke_test.go`).

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func probedruckzeile() *domain.Druck {
	return &domain.Druck{Name: "Bolt", Setcode: "lea", Nummer: "161", Kennungen: []domain.Druckkennung{
		scryfallkennung("e3285e6b-3e79-4d7c-bf96-d920f973b122"),
		{System: domain.SystemCardmarket, Semantik: domain.SemantikProdukt, Wert: "1"},
	}}
}

func mussAusfuehren(t *testing.T, db *sql.DB, befehl string) {
	t.Helper()

	if _, err := db.ExecContext(context.Background(), befehl); err != nil {
		t.Fatalf("%s: %v", befehl, err)
	}
}

func TestDieDruckablageMeldetWasSQLiteHergibt(t *testing.T) {
	t.Parallel()

	t.Run("ein Auslöser sperrt das Anlegen", func(t *testing.T) {
		t.Parallel()

		db := frischeDatenbank(t)
		mussAusfuehren(t, db, `CREATE TRIGGER sperre BEFORE INSERT ON drucke BEGIN SELECT RAISE(ABORT, 'gesperrt'); END`)

		if err := storage.NeueDrucke(db).Sichern(probekontext(t), probedruckzeile()); err == nil {
			t.Error("das gesperrte Anlegen ging durch")
		}
	})

	t.Run("ein Auslöser sperrt das Fortschreiben", func(t *testing.T) {
		t.Parallel()

		db := frischeDatenbank(t)
		ablage := storage.NeueDrucke(db)
		druck := probedruckzeile()

		if err := ablage.Sichern(probekontext(t), druck); err != nil {
			t.Fatalf("anlegen: %v", err)
		}

		mussAusfuehren(t, db, `CREATE TRIGGER sperre BEFORE UPDATE ON drucke BEGIN SELECT RAISE(ABORT, 'gesperrt'); END`)

		if err := ablage.Sichern(probekontext(t), druck); err == nil {
			t.Error("das gesperrte Fortschreiben ging durch")
		}
	})

	t.Run("ein Auslöser sperrt eine Kennung", func(t *testing.T) {
		t.Parallel()

		db := frischeDatenbank(t)
		mussAusfuehren(t, db, `CREATE TRIGGER sperre BEFORE INSERT ON druckkennungen WHEN NEW.wert = 'kaputt'
			BEGIN SELECT RAISE(ABORT, 'gesperrt'); END`)

		druck := &domain.Druck{Name: "Bolt", Kennungen: []domain.Druckkennung{
			{System: domain.SystemCardmarket, Semantik: domain.SemantikProdukt, Wert: "kaputt"},
		}}

		if err := storage.NeueDrucke(db).Sichern(probekontext(t), druck); err == nil {
			t.Error("die gesperrte Kennung ging durch")
		}
	})

	t.Run("ohne Kennungstabelle scheitert das Wiedersehen und jedes Lesen", func(t *testing.T) {
		t.Parallel()

		db := frischeDatenbank(t)
		ablage := storage.NeueDrucke(db)
		druck := probedruckzeile()

		if err := ablage.Sichern(probekontext(t), druck); err != nil {
			t.Fatalf("anlegen: %v", err)
		}

		mussAusfuehren(t, db, `DROP TABLE druckkennungen`)

		if err := ablage.Sichern(probekontext(t), probedruckzeile()); err == nil {
			t.Error("das Wiedersehen ohne Kennungstabelle ging durch")
		}

		// **Mit Kennung der Zeile** geht es am Wiedersehen vorbei — und
		// scheitert beim Anhängen der Kennung.
		if err := ablage.Sichern(probekontext(t), druck); err == nil {
			t.Error("das Anhängen ohne Kennungstabelle ging durch")
		}

		if _, err := ablage.Fuer(probekontext(t), druck.ID); err == nil {
			t.Error("die Kennungen ließen sich ohne Tabelle lesen")
		}

		if _, err := ablage.NachScryfall(probekontext(t), "e3285e6b-3e79-4d7c-bf96-d920f973b122"); err == nil {
			t.Error("der Alias ließ sich ohne Tabelle lesen")
		}
	})

	t.Run("ohne Drucktabelle scheitert das Lesen", func(t *testing.T) {
		t.Parallel()

		db := frischeDatenbank(t)
		mussAusfuehren(t, db, `DROP TABLE drucke`)

		if _, err := storage.NeueDrucke(db).Alle(probekontext(t)); err == nil {
			t.Error("die Drucke ließen sich ohne Tabelle lesen")
		}
	})

	t.Run("ein geschlossener Bestand nimmt nichts an", func(t *testing.T) {
		t.Parallel()

		db := frischeDatenbank(t)
		if err := db.Close(); err != nil {
			t.Fatalf("schließen: %v", err)
		}

		if err := storage.NeueDrucke(db).Sichern(probekontext(t), probedruckzeile()); err == nil {
			t.Error("der geschlossene Bestand nahm den Druck an")
		}
	})
}

func TestDieDruckablageMeldetWasNurDerStolpertreiberHergibt(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name, treiber string
		lauf          func(ablage *storage.Drucke) error
	}{
		{"die Zeilennummer lässt sich nicht lesen", "stolper-druck-id", func(a *storage.Drucke) error {
			return a.Sichern(context.Background(), probedruckzeile())
		}},
		{"der Abschluss scheitert", "stolper-druck-commit", func(a *storage.Drucke) error {
			return a.Sichern(context.Background(), probedruckzeile())
		}},
		{"ein Fehler mitten im Lesen der Drucke", "stolper-druck-ende", func(a *storage.Drucke) error {
			_, err := a.Alle(context.Background())

			return err
		}},
		{"eine Druckzeile passt nicht in ihre Felder", "stolper-druck-scan", func(a *storage.Drucke) error {
			_, err := a.Alle(context.Background())

			return err
		}},
		{"ein Fehler mitten im Lesen der Kennungen", "stolper-kennung-ende", func(a *storage.Drucke) error {
			_, err := a.Fuer(context.Background(), 1)

			return err
		}},
		{"eine Kennungszeile passt nicht in ihre Felder", "stolper-kennung-scan", func(a *storage.Drucke) error {
			_, err := a.Fuer(context.Background(), 1)

			return err
		}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("öffnen: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			if fehler := fall.lauf(storage.NeueDrucke(db)); fehler == nil {
				t.Error("der Stolperer kam ohne Fehler durch")
			} else if errors.Is(fehler, domain.ErrConflict) {
				t.Errorf("der Stolperer wurde zum Konflikt: %v", fehler)
			}
		})
	}
}
