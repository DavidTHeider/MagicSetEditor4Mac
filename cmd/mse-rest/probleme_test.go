// **Die Verdrahtung der Problemprüfung** — PERF02
// (#2352).
//
// Dass die Wurzel den Weg erreicht und den Stand übersetzt; was er fachlich
// tut, halten `internal/service` und `internal/app`.
package main

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestOhneZuhauseKommtDieProblemprufungNichtZusammen(t *testing.T) {
	t.Setenv("HOME", "")

	var ohneBestand *sql.DB

	naht := probleme{db: ohneBestand}

	if _, fehler := naht.Problemstand(probekontext(t), 1); fehler == nil {
		t.Error("ohne Zuhause kam ein Stand zurück")
	}

	if _, fehler := naht.ProblemePruefen(probekontext(t), 1); fehler == nil {
		t.Error("ohne Zuhause wurde geprüft")
	}
}

func TestDieProblemprufungUebersetztDenStand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db, err := storage.Oeffnen(probekontext(t), "file:"+filepath.Join(t.TempDir(), "bestand.db"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Aufräumen

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatal(fehler)
	}

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe", Style: "probe-stil",
		Cards: []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Eine"}}}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	naht := probleme{db: db}

	if vorher, fehler := naht.Problemstand(probekontext(t), setID); fehler != nil || vorher.Urteil != "ungeprueft" || vorher.Karten == nil {
		t.Errorf("vorher %+v, %v", vorher, fehler)
	}

	if nachher, fehler := naht.ProblemePruefen(probekontext(t), setID); fehler != nil || nachher.Urteil != "aktuell" || nachher.Gesamt != 1 {
		t.Errorf("nachher %+v, %v", nachher, fehler)
	}
}
