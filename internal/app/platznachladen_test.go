//go:build darwin

// **Ein Nachladefehler ist kein gescheitertes Wegräumen** — §7, Regel 3
// (#1970).
//
// # Wie er hier entsteht
//
// **Ohne Kunstgriff, an einem echten Anwendungsordner**: das Wegräumen
// liest nur den Posten, den es räumt; die Zählung danach geht durch alle.
// Ein Verzeichnis unter den Bildern, das niemand lesen darf, lässt das
// Wegräumen der Vorsicherungen glücken und die Zählung danach scheitern. Bis
// #1970 kam das als `fehler` zurück — weggeräumt war trotzdem.
package app_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinNachladefehlerIstKeinGescheitertesWegraeumen(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	neben := filepath.Dir(bestand)

	for stelle, name := range []string{"bestand-vor-0.1.0.db", "bestand-vor-0.2.0.db"} {
		pfad := filepath.Join(neben, name)
		if err := os.WriteFile(pfad, make([]byte, 4096), 0o600); err != nil {
			t.Fatalf("vorbereiten: %v", err)
		}

		alt := time.Now().Add(time.Duration(stelle-2) * time.Hour)
		if err := os.Chtimes(pfad, alt, alt); err != nil {
			t.Fatalf("vorbereiten: %v", err)
		}
	}

	gesperrt := filepath.Join(neben, "beiwerk", "gesperrt")
	if err := os.MkdirAll(gesperrt, 0o700); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	if err := os.Chmod(gesperrt, 0o000); err != nil {
		t.Fatalf("sperren: %v", err)
	}

	// **Wieder auf, bevor das Zuhause verschwindet** — ein Verzeichnis ohne
	// Rechte räumte `t.TempDir` nicht weg.
	t.Cleanup(func() { _ = os.Chmod(gesperrt, 0o700) }) //nolint:errcheck,gosec // zurück für t.TempDir

	geraeumt := anwendung.PlatzFreigeben(t.Context(), string(domain.PostenVorsicherungen))

	switch {
	case geraeumt.Fehler != nil:
		t.Fatalf("als gescheitert gemeldet: %+v", geraeumt.Fehler)
	case geraeumt.Nachladefehler == nil:
		t.Fatalf("der Nachladefehler fehlt: %+v", geraeumt)
	case geraeumt.Posten == nil:
		t.Fatal("die Liste ist nil")
	}

	// **Weggeräumt ist weggeräumt** — die jüngste bleibt, wie immer.
	if uebrig, _ := filepath.Glob(filepath.Join(neben, "bestand-vor-*.db")); len(uebrig) != 1 { //nolint:errcheck // ein fester Ausdruck
		t.Fatalf("es blieben %v", uebrig)
	}

	// **Ein gescheiterter Schritt bleibt einer** — der Bestand ist kein
	// räumbarer Posten.
	if abgewiesen := anwendung.PlatzFreigeben(
		t.Context(), string(domain.PostenBestand)); abgewiesen.Fehler == nil ||
		abgewiesen.Nachladefehler != nil {
		t.Fatalf("der Bestand: %+v", abgewiesen)
	}
}
