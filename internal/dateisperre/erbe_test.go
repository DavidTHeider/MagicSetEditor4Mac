// **Ein Erbe ist kein Halter** — #2401
// (#2401).
//
// `flock` hängt an der offenen Dateibeschreibung, und ein Kind, das
// irgendein Teil des Prozesses gerade startet, hält sie zwischen `fork` und
// `exec` mit. Nachgestellt wird das Erbe hier mit einem zweiten Griff, der
// die Sperre kurz hält und dann loslässt — dieselbe Gestalt, ohne auf einen
// Zufall zu warten.
package dateisperre_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
)

// haltenFuer nimmt die Sperre allein und gibt sie nach `dauer` frei.
func haltenFuer(t *testing.T, pfad string, dauer time.Duration) {
	t.Helper()

	halter, err := dateisperre.Oeffnen(pfad)
	if err != nil {
		t.Fatal(err)
	}

	if err := halter.Allein(); err != nil {
		t.Fatal(err)
	}

	uhr := time.AfterFunc(dauer, func() { _ = halter.Schliessen() }) //nolint:errcheck // losgelassen ist losgelassen

	t.Cleanup(func() {
		uhr.Stop()
		_ = halter.Schliessen() //nolint:errcheck // Aufräumen
	})
}

func TestEinKurzerHalterIstKeinHalter(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "probe.sperre")
	haltenFuer(t, pfad, 20*time.Millisecond)

	frager, err := dateisperre.Oeffnen(pfad)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = frager.Schliessen() }) //nolint:errcheck // Aufräumen

	if err := frager.AlleinOhneErbe(t.Context()); err != nil {
		t.Errorf("ein Griff, der nach 20 ms losließ, galt als Halter: %v", err)
	}
}

func TestEinLangerHalterBleibtEinHalter(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "probe.sperre")
	haltenFuer(t, pfad, time.Minute)

	frager, err := dateisperre.Oeffnen(pfad)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = frager.Schliessen() }) //nolint:errcheck // Aufräumen

	beginn := time.Now()
	if err := frager.AlleinOhneErbe(t.Context()); err == nil {
		t.Error("die Sperre kam durch, obwohl ein anderer sie hält")
	}

	if gewartet := time.Since(beginn); gewartet < dateisperre.Erbfrist {
		t.Errorf("nach %v aufgegeben — ein Erbe hält bis zu %v", gewartet, dateisperre.Erbfrist)
	}
}
