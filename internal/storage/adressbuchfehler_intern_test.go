package storage

// **Jeder Fehlerweg des Adressbuchs meldet sich** — ID-04 (#1703).

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// adressbuchrufe sind die vier Wege der Ablage.
func adressbuchrufe(ablage *Adressbuchablage) map[string]func(ctx context.Context) error {
	return map[string]func(ctx context.Context) error{
		"lesen": func(ctx context.Context) error {
			_, fehler := ablage.Kontakte(ctx)

			return fehler
		},
		"anlegen": func(ctx context.Context) error {
			return ablage.KontakteUebernehmen(ctx, []*domain.Kontakt{{Name: "Ada"}})
		},
		"ersetzen": func(ctx context.Context) error {
			return ablage.KontakteUebernehmen(ctx, []*domain.Kontakt{{ID: 1, Fassung: 1, Name: "Ada"}})
		},
		"löschen": func(ctx context.Context) error { return ablage.KontaktLoeschen(ctx, 1, 1) },
	}
}

func TestJederFehlerwegDesAdressbuchsMeldetSich(t *testing.T) {
	// **Ohne Profil und ohne Tafel** — jeder der vier Wege.
	for name, ruf := range adressbuchrufe(NeueAdressbuchablage(offeneDatenbank(t))) {
		if ruf(t.Context()) == nil {
			t.Errorf("%s ohne Profil kam durch", name)
		}

		if ruf(probekontext(t)) == nil {
			t.Errorf("%s ohne Tafel kam durch", name)
		}
	}

	// **Und an den Stellen, die nur ein Treiber hergibt.**
	for name, fall := range map[string]struct{ treiber, weg string }{
		"eine krumme Zeile":           {"stolper-umwandeln", "lesen"},
		"ein Fehler beim Weiterlesen": {"stolper-lesen", "lesen"},
		"das Zählen scheitert":        {"stolper-zaehlen", "ersetzen"},
		"das Nachschlagen scheitert":  {"stolper-fragen", "löschen"},
	} {
		db, err := sql.Open(fall.treiber, "")
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		t.Cleanup(func() { _ = db.Close() })

		if adressbuchrufe(NeueAdressbuchablage(db))[fall.weg](probekontext(t)) == nil {
			t.Errorf("%s kam durch", name)
		}
	}
}

// **Eine Liste, die keine Wege trägt, ist ein Fehler** — die Bedingung hält
// die Gestalt einer Liste, nicht die ihrer Einträge.
func TestKrummeWegeWerdenGemeldet(t *testing.T) {
	db := migrierteDatenbank(t)

	if _, err := db.ExecContext(probekontext(t),
		`INSERT INTO kontakte (profil_id, name, wege, uebernommen) VALUES (1, 'Ada', '[1]', 'gestern')`); err != nil {
		t.Fatal(err)
	}

	if _, fehler := NeueAdressbuchablage(db).Kontakte(probekontext(t)); fehler == nil {
		t.Error("krumme Wege kamen durch")
	}
}
