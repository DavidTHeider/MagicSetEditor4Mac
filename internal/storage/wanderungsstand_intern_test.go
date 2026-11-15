package storage

import (
	"errors"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerWanderungsstandErkenntEineNeuereFassung(t *testing.T) {
	t.Parallel()

	db, err := Oeffnen(probekontext(t), filepath.Join(t.TempDir(), "bestand.db"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // nur zum Aufräumen

	if fehler := Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := PruefeWanderungsstand(probekontext(t), db); fehler != nil {
		t.Fatalf("der eigene Stand gilt als fremd: %v", fehler)
	}

	weitere, name := kuenftigeWanderung(t, "kuenftig", "CREATE TABLE kuenftig (wert INTEGER) STRICT;")
	if fehler := MigriereMit(probekontext(t), db, weitere, nil); fehler != nil {
		t.Fatal(fehler)
	}

	var neuer *domain.NeuereFassung
	if fehler := PruefeWanderungsstand(probekontext(t), db); !errors.As(fehler, &neuer) || neuer.Unbekannt[0] != name {
		t.Errorf("nach der Wanderung einer neueren Fassung: %v", fehler)
	}
}

func TestDerWanderungsstandSagtWasErNichtLesenKann(t *testing.T) {
	t.Parallel()

	db, err := Oeffnen(probekontext(t), filepath.Join(t.TempDir(), "bestand.db"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // nur zum Aufräumen

	kaputt := fstest.MapFS{"migrationen": {Mode: 0o644, Data: []byte("keine Liste")}}

	for name, fall := range map[string]func() error{
		"die Wanderungen": func() error { return pruefeWanderungsstandGegen(probekontext(t), db, kaputt, grundlinie) },
		"die Grundlinie":  func() error { return pruefeWanderungsstandGegen(probekontext(t), db, migrationen, kaputt) },
		"die Buchführung": func() error { _ = db.Close(); return PruefeWanderungsstand(probekontext(t), db) }, //nolint:errcheck // geschlossen gefragt
	} {
		if fehler := fall(); fehler == nil {
			t.Errorf("%s: kein Fehler", name)
		}
	}
}
