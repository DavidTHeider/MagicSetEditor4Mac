package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Sicherung, für die der Platz nicht reicht, heißt „voll"** — M02
// (#2346).
//
// Gefunden von der Hüllenprobe `wanderung09` am 2026-10-14: auf einem Abbild
// mit 2 MB frei scheiterte die Sicherung eines Bestands von 4 MB vor der
// Wanderung **mit einem Fehler des Datenträgers** — SQLite meldet nicht
// `FULL` —, und nach dem Räumen der halben waren wieder 2 MB frei. Die Regel
// des Starts („weniger als ein Mebibyte", [vollStattEA]) sah deshalb keinen
// vollen Datenträger. **Eine Sicherung braucht so viel, wie der Bestand
// belegt.**
//
// **Seit #2407 heißt das „knapp", nicht „voll"**: ohne `ENOSPC` in der Kette
// bleibt der Fehler des Datenträgers die Hauptsache, und der knappe Platz
// wird dazu genannt.
func TestEineSicherungOhnePlatzHeisstVoll(t *testing.T) {
	t.Parallel()

	ea := domain.MitGrund(domain.ErrEA, errors.New("sqlite3: disk I/O error"), "prüfe den Datenträger")
	ziel := filepath.Join(t.TempDir(), "bestand-vor-0.8.0.db")

	faelle := []struct {
		name     string
		gemeldet error
		frei     int64
		bedarf   int64
		voll     bool
		knapp    bool
	}{
		{name: "weniger frei, als der Bestand belegt", gemeldet: ea, frei: 2 << 20, bedarf: 4 << 20, knapp: true},
		{name: "weniger frei, und ENOSPC in der Kette", gemeldet: fmt.Errorf("%w: %w", ea, syscall.ENOSPC), frei: 2 << 20, bedarf: 4 << 20, voll: true},
		{name: "genug frei", gemeldet: ea, frei: 8 << 20, bedarf: 4 << 20},
		{name: "Bedarf unbekannt, aber kaum frei", gemeldet: ea, frei: 0, bedarf: 0, knapp: true},
		{name: "ein anderer Fehler bleibt, wie er ist", gemeldet: domain.ErrConflict, frei: 0, bedarf: 4 << 20},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			sichern := mitPlatzfrage(
				func(context.Context, string) error { return fall.gemeldet },
				func(context.Context) int64 { return fall.bedarf },
				func(ort string) (int64, error) {
					if ort != filepath.Dir(ziel) {
						t.Errorf("gefragt wurde %s, nicht neben dem Ziel", ort)
					}

					return fall.frei, nil
				})

			fehler := sichern(t.Context(), ziel)

			if voll := errors.Is(fehler, domain.ErrVoll); voll != fall.voll {
				t.Errorf("voll: %v, erwartet %v — %v", voll, fall.voll, fehler)
			}

			if knapp := errors.Is(fehler, domain.ErrKaumPlatz); knapp != fall.knapp {
				t.Errorf("knapp: %v, erwartet %v — %v", knapp, fall.knapp, fehler)
			}

			if !errors.Is(fehler, fall.gemeldet) {
				t.Errorf("der gemeldete Fehler ging verloren: %v", fehler)
			}
		})
	}
}
