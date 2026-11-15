package storage_test

// **Jeder Fehlerweg des Arbeitsverlaufs ist betreten** — Werkstück U11 von
// Phase 5.1 (#1422).

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// verlaufBefehl führt einen Befehl aus, der gelingen muss.
func verlaufBefehl(t *testing.T, db *sql.DB, befehl string) {
	t.Helper()

	if _, err := db.ExecContext(context.Background(), befehl); err != nil {
		t.Fatalf("%s: %v", befehl, err)
	}
}

func TestDerArbeitsverlaufMeldetWasSQLiteHergibt(t *testing.T) {
	t.Parallel()

	objekt := domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: "1", Name: "Probe"}

	t.Run("ohne Profil geht kein Weg", func(t *testing.T) {
		t.Parallel()

		ablage := storage.NeuerArbeitsverlauf(frischeDatenbank(t))
		ctx := context.Background()

		if err := ablage.Anheften(ctx, objekt); err == nil {
			t.Error("Anheften ohne Profil ging durch")
		}

		if _, err := ablage.Loesen(ctx, domain.ArbeitsartSet, "1"); err == nil {
			t.Error("Lösen ohne Profil ging durch")
		}

		if _, err := ablage.Angeheftete(ctx); err == nil {
			t.Error("Angeheftete ohne Profil ging durch")
		}

		if err := ablage.Vergessen(ctx, domain.ArbeitsartSet, "1"); err == nil {
			t.Error("Vergessen ohne Profil ging durch")
		}
	})

	t.Run("ohne Tabellen scheitert jedes Lesen und Schreiben", func(t *testing.T) {
		t.Parallel()

		db := frischeDatenbank(t)
		ablage := storage.NeuerArbeitsverlauf(db)
		ctx := probekontext(t)

		verlaufBefehl(t, db, `DROP TABLE arbeitsverlauf`)
		verlaufBefehl(t, db, `DROP TABLE angeheftet`)

		if err := ablage.Vermerken(ctx, arbeitsschritt("1", domain.VorgangGeoeffnet, 8)); err == nil {
			t.Error("Vermerken ohne Tabelle ging durch")
		}

		if _, err := ablage.Schritte(ctx); err == nil {
			t.Error("Schritte ohne Tabelle kamen")
		}

		if err := ablage.Anheften(ctx, objekt); err == nil {
			t.Error("Anheften ohne Tabelle ging durch")
		}

		if _, err := ablage.Loesen(ctx, domain.ArbeitsartSet, "1"); err == nil {
			t.Error("Lösen ohne Tabelle ging durch")
		}

		if _, err := ablage.Angeheftete(ctx); err == nil {
			t.Error("Angeheftete ohne Tabelle kamen")
		}

		if err := ablage.Vergessen(ctx, domain.ArbeitsartSet, "1"); err == nil {
			t.Error("Vergessen ohne Tabelle ging durch")
		}
	})

	t.Run("eine Zeile, die nicht in ihre Felder passt", func(t *testing.T) {
		t.Parallel()

		db := frischeDatenbank(t)
		ablage := storage.NeuerArbeitsverlauf(db)
		ctx := probekontext(t)

		// **Ein Zeitpunkt, der keiner ist** — von Hand hineingeschrieben.
		verlaufBefehl(t, db, `INSERT INTO arbeitsverlauf (profil_id, art, kennung, name, vorgang, zeitpunkt)
			VALUES (1, 'set', '1', 'x', 'geoeffnet', 'irgendwann')`)

		if _, err := ablage.Schritte(ctx); err == nil {
			t.Error("ein unlesbarer Zeitpunkt kam durch")
		}

		// **Eine Sicht mit NULL** statt der Tabellen — Scan in eine Zeichenkette scheitert.
		verlaufBefehl(t, db, `DROP TABLE arbeitsverlauf`)
		verlaufBefehl(t, db, `CREATE VIEW arbeitsverlauf AS SELECT 1 AS profil_id, 'set' AS art, NULL AS kennung, '' AS name, 'geoeffnet' AS vorgang, '' AS zeitpunkt, 1 AS id`)

		if _, err := ablage.Schritte(ctx); err == nil {
			t.Error("eine Schrittzeile mit NULL kam durch")
		}

		verlaufBefehl(t, db, `DROP TABLE angeheftet`)
		verlaufBefehl(t, db, `CREATE VIEW angeheftet AS SELECT 1 AS profil_id, 'set' AS art, NULL AS kennung, '' AS name, '' AS zeitpunkt, 1 AS id`)

		if _, err := ablage.Angeheftete(ctx); err == nil {
			t.Error("eine angeheftete Zeile mit NULL kam durch")
		}
	})
}

func TestDerArbeitsverlaufMeldetEinenFehlerMittenImLesen(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name, treiber string
		lauf          func(a *storage.Arbeitsverlauf) error
	}{
		{"die Schritte", "stolper-arbeitsverlauf-ende", func(a *storage.Arbeitsverlauf) error {
			_, err := a.Schritte(storage.MitProfil(context.Background(), 1))

			return err
		}},
		{"das Angeheftete", "stolper-angeheftet-ende", func(a *storage.Arbeitsverlauf) error {
			_, err := a.Angeheftete(storage.MitProfil(context.Background(), 1))

			return err
		}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("öffnen: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			if fehler := fall.lauf(storage.NeuerArbeitsverlauf(db)); fehler == nil {
				t.Error("der Stolperer kam ohne Fehler durch")
			}
		})
	}
}
