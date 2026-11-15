package app

// **Ein voller Datenträger ist voll, auch wenn SQLite „E/A" sagt** — F05
// (#2344).
//
// Gemessen am nativen Fenster (`startlagen09`, 2026-10-14): auf einem
// randvollen Abbild scheitert schon `PRAGMA journal_mode = WAL` mit
// `SQLITE_IOERR_SHMSIZE` (4874) — die Nebendatei `-shm` lässt sich nicht
// vergrößern, und `ENOSPC` steht nicht in der Kette. Das Fenster riet
// „prüfe den Datenträger", wo „schaffe Platz" der Rat ist.
//
// **Seit #2407 nennt es beides**: dass ein voller Datenträger einen Fehler
// des Datenträgers erzeugt, beweist nicht die Umkehrung — ein echter Fehler
// bei zugleich wenig Platz bekam sonst nur „schaffe Platz". **Voll** heißt
// es nur noch mit Beleg in der Kette (`ENOSPC`, oder schon `ErrVoll` aus
// `SQLITE_FULL`).

import (
	"errors"
	"fmt"
	"syscall"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinDatentraegerfehlerBeiVollemVolumenIstVoll(t *testing.T) {
	t.Parallel()

	ea := domain.MitGrund(domain.ErrEA, errors.New("sqlite3: disk I/O error"), "prüfen")
	frei := func(bytes int64, err error) func(string) (int64, error) {
		return func(string) (int64, error) { return bytes, err }
	}

	for _, fall := range []struct {
		name   string
		fehler error
		frei   func(string) (int64, error)
		klasse domain.Schluessel
	}{
		{"randvoll, ohne Beleg in der Kette", ea, frei(12<<10, nil), domain.SchluesselKaumPlatz},
		{"randvoll, mit ENOSPC", fmt.Errorf("%w: %w", ea, syscall.ENOSPC), frei(12<<10, nil), domain.SchluesselVoll},
		{"schon voll aus SQLITE_FULL", fmt.Errorf("%w: %w", ea, domain.ErrVoll), frei(12<<10, nil), domain.SchluesselVoll},
		// **Ganz voll meldet SQLite `CANTOPEN`** (gemessen am 2026-10-15 mit 0
		// KB frei): die Nebendatei lässt sich nicht anlegen, und der Bestand
		// hieß „gesperrt" oder „nur lesbar" — mit dem Rat, Schreibrecht zu geben.
		{"gesperrt bei kaum Platz", domain.MitGrund(domain.ErrGesperrt, errors.New("sqlite3: unable to open database file"), "freigeben"), frei(0, nil), domain.SchluesselKaumPlatz},
		{"nur lesbar bei kaum Platz", domain.MitGrund(domain.ErrNurLesend, domain.MitGrund(domain.ErrGesperrt, errors.New("cantopen"), "freigeben"), "Schreibrecht"), frei(0, nil), domain.SchluesselKaumPlatz},
		{"nur lesbar bei genug Platz", domain.MitGrund(domain.ErrNurLesend, errors.New("readonly"), "Schreibrecht"), frei(10<<20, nil), domain.SchluesselBestandNurLesbar},
		{"genug Platz", ea, frei(10<<20, nil), domain.SchluesselBestandEA},
		{"Platz unbekannt", ea, frei(0, errors.New("statfs")), domain.SchluesselBestandEA},
		{"kein Datenträgerfehler", domain.MitGrund(domain.ErrBelegt, errors.New("busy"), "gleich"), frei(0, nil), domain.SchluesselBelegt},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if klasse := startklasse(vollStattEA(fall.fehler, "/probe", fall.frei)); klasse != fall.klasse {
				t.Errorf("Klasse %q, erwartet %q", klasse, fall.klasse)
			}
		})
	}
}
