package app

// **Hält der Dienst den Bestand, sagt das Zurückspielen es** — M02
// (#2346),
// Befund A-11: die Meldung lautete „von einem anderen Programm geöffnet —
// schließe es", und den Dienst sieht niemand als Programm.

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestHaeltDerDienstDenBestandSagtDasZurueckspielenEs(t *testing.T) { //nolint:paralleltest // die Frage nach dem Dienst ist eine Paketvariable
	pfad := filepath.Join(t.TempDir(), "bestand.db")

	halter, err := storage.Oeffnen(t.Context(), pfad)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = halter.Close() }) //nolint:errcheck // Aufräumen

	if fehler := storage.Migriere(t.Context(), halter, nil); fehler != nil {
		t.Fatal(fehler)
	}

	vorher := dienstprozesse

	t.Cleanup(func() { dienstprozesse = vorher })

	// **Ein Programm, das läuft und den Bestand nicht hält** — Zweitblick über
	// Phase 2, Befund 8: bis dahin genügte, dass der Dienst lief.
	fremd := exec.CommandContext(t.Context(), "/bin/sleep", "30")
	if err := fremd.Start(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = fremd.Process.Kill() //nolint:errcheck // Aufräumen
		_ = fremd.Wait()         //nolint:errcheck // beendet ist beendet
	})

	for name, fall := range map[string]struct {
		pid    int
		dienst bool
	}{
		"der Dienst hält ihn":             {pid: os.Getpid(), dienst: true},
		"der Dienst läuft und hält nicht": {pid: fremd.Process.Pid, dienst: false},
		"kein Dienst":                     {pid: 0, dienst: false},
	} {
		dienstprozesse = func(context.Context) ([]dienstprozess, error) { return nurDerNeue(fall.pid), nil }

		_, fehler := alleinnaht(t.Context(), pfad)
		if !errors.Is(fehler, domain.ErrConflict) {
			t.Fatalf("%s: %v", name, fehler)
		}

		var dienst *domain.DienstHaeltBestand
		if errors.As(fehler, &dienst) != fall.dienst {
			t.Errorf("%s: den Dienst genannt %v, erwartet %v — %v", name, !fall.dienst, fall.dienst, fehler)
		}

		if uebersetzt := Uebersetze(t.Context(), fehler); fall.dienst && uebersetzt.Schluessel != domain.SchluesselBestandVomDienstGehalten {
			t.Errorf("%s: übersetzt %+v", name, uebersetzt)
		}
	}

	// **Und ein Fragen, das scheitert, nennt niemanden** — auch nicht den
	// Dienst, den es vor dem Fehler fand; dann bleibt die allgemeine Meldung.
	dienstprozesse = func(context.Context) ([]dienstprozess, error) {
		return nurDerNeue(os.Getpid()), errors.New("launchctl kaputt")
	}

	_, fehler := alleinnaht(t.Context(), pfad)

	var dienst *domain.DienstHaeltBestand
	if !errors.Is(fehler, domain.ErrConflict) || errors.As(fehler, &dienst) {
		t.Errorf("ohne Antwort von launchctl: %v", fehler)
	}
}
