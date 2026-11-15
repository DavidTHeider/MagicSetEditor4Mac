package storage_test

// **Ein Erbe ist kein zweites Programm** — #2401
// (#2401).
//
// Ein Kind, das irgendein Teil des Prozesses gerade startet, hält die
// Nachbardatei zwischen `fork` und `exec` mit. Bis dahin hieß das „wird von
// einem anderen Programm benutzt — schließe es", und es gab keins.

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinKurzerHalterVerhindertDenAlleinanspruchNicht(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "bestand.db")

	erbe, err := dateisperre.Oeffnen(pfad + storage.Sperrendung)
	if err != nil {
		t.Fatal(err)
	}

	if err := erbe.Allein(); err != nil {
		t.Fatal(err)
	}

	uhr := time.AfterFunc(20*time.Millisecond, func() { _ = erbe.Schliessen() }) //nolint:errcheck // losgelassen ist losgelassen
	t.Cleanup(func() { uhr.Stop() })

	zurueck, fehler := storage.Alleinanspruch(pfad)
	if fehler != nil {
		t.Fatalf("ein Griff, der nach 20 ms losließ, galt als zweites Programm: %v", fehler)
	}

	if zurueckfehler := zurueck(); zurueckfehler != nil {
		t.Errorf("zurückgeben: %v", zurueckfehler)
	}
}
