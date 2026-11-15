package app_test

// **Die Problemprüfung an den Flächen ohne Fenster** — PERF02
// (#2352):
// der Weg prüft über dem Zeichenwerk des Bestands, nennt den Stand mit
// seinem Urteil, und eine Änderung am Set macht ihn veraltet.

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// bestandMitDreiKarten gibt einen migrierten Bestand mit einem Set aus drei Karten.
func bestandMitDreiKarten(t *testing.T) (db *sql.DB, setID int64) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())

	db, err := storage.Oeffnen(probekontext(t), "file:"+filepath.Join(t.TempDir(), "bestand.db"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Aufräumen

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	setID, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe", Style: "probe-stil",
		Cards: []*domain.Card{
			{Values: []domain.Value{{Field: "name", Text: "Eine"}}},
			{Values: []domain.Value{{Field: "name", Text: "Zwei"}}},
			{Values: []domain.Value{{Field: "name", Text: "Drei"}}},
		},
	}, nil)
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	return db, setID
}

func TestDerProblemwegPrueftUndNenntDenStand(t *testing.T) {
	db, setID := bestandMitDreiKarten(t)

	weg, err := app.ProblemwegUeber(db)
	if err != nil {
		t.Fatal(err)
	}

	vorher, err := weg.Stand(probekontext(t), setID)
	if err != nil || vorher.Urteil != string(domain.ProblemeUngeprueft) || vorher.Sicher || vorher.Karten == nil {
		t.Fatalf("vor der Prüfung %+v, %v", vorher, err)
	}

	gemeldet := 0

	stand, err := weg.Pruefen(probekontext(t), setID, func(int, int) { gemeldet++ })
	if err != nil {
		t.Fatal(err)
	}

	if stand.Urteil != string(domain.ProblemeAktuell) || !stand.Sicher || stand.Geprueft != 3 || stand.Gesamt != 3 || stand.Beendet == "" || gemeldet != 3 {
		t.Errorf("nach der Prüfung %+v, %d Meldungen", stand, gemeldet)
	}

	// **Eine Änderung am Set macht den Stand veraltet** — über die Auslöser
	// im Bestand, gleich, welcher Weg schrieb.
	if _, aenderung := db.ExecContext(probekontext(t), `UPDATE card_values SET wert = 'Vier' WHERE wert = 'Drei'`); aenderung != nil {
		t.Fatal(aenderung)
	}

	danach, err := weg.Stand(probekontext(t), setID)
	if err != nil || danach.Urteil != string(domain.ProblemeVeraltet) || danach.Sicher {
		t.Errorf("nach der Änderung %+v, %v", danach, err)
	}
}

// **Ein Abbruch nennt den Stand, den es gibt** — und einer über einem Set,
// das es nicht gibt, keinen.
func TestDerProblemwegNenntAbbruchUndUnbekanntes(t *testing.T) {
	db, setID := bestandMitDreiKarten(t)

	weg, err := app.ProblemwegUeber(db)
	if err != nil {
		t.Fatal(err)
	}

	ctx, abbrechen := context.WithCancel(probekontext(t))
	abbrechen()

	stand, err := weg.Pruefen(ctx, setID, nil)
	if !errors.Is(err, context.Canceled) || stand.Urteil == "" || stand.Sicher {
		t.Errorf("abgebrochen: %+v, %v", stand, err)
	}

	if _, err := weg.Pruefen(probekontext(t), 999, nil); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unbekanntes Set: %v", err)
	}

	if _, err := weg.Stand(probekontext(t), 999); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("der Stand eines unbekannten Sets: %v", err)
	}
}

func TestDerProblemwegBrauchtEinZuhause(t *testing.T) {
	db, _ := bestandMitDreiKarten(t)
	t.Setenv("HOME", "")

	if _, err := app.ProblemwegUeber(db); err == nil {
		t.Error("ohne Zuhause gebaut")
	}
}
