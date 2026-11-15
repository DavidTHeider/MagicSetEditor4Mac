package app

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Fehler des Datenträgers bei knappem Platz heißt an jeder Fläche
// gleich** — M02 (#2346),
// seit #2407 „knapp" und nicht mehr „voll".
//
// Gefunden von der Hüllenprobe `wanderung09` am 2026-10-14: auf einem
// vollen Abbild scheiterte die Sicherung vor der Wanderung mit einem Fehler
// des Datenträgers, und `mse` riet, den Datenträger zu prüfen. Das Fenster
// sagt in derselben Lage seit F05 „schaffe Platz" (`startplatz.go`); `mse`,
// `mse-mcp` und `mse-rest` melden ihren Start über [MitWegZurueck] und sagen
// jetzt dasselbe (§16) — seit #2407 beides: schaffe Platz und prüfe den
// Datenträger, denn ohne `ENOSPC` belegt der Fehler keine volle Platte.
func TestMitWegZurueckNenntDenKnappenPlatz(t *testing.T) {
	t.Parallel()

	ort := filepath.Join(t.TempDir(), "bestand.db")
	ea := fmt.Errorf("den Bestand vor der Migration sichern: %w",
		domain.MitGrund(domain.ErrEA, errors.New("sqlite3: disk I/O error"), "prüfe den Datenträger"))

	faelle := []struct {
		name  string
		frei  int64
		knapp bool
	}{
		{name: "kaum Platz frei", frei: 0, knapp: true},
		{name: "genug Platz frei", frei: 1 << 30, knapp: false},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			gemeldet := mitWegZurueck(t.Context(), ort, ea, func(string) (int64, error) { return fall.frei, nil })

			if knapp := errors.Is(gemeldet, domain.ErrKaumPlatz); knapp != fall.knapp || errors.Is(gemeldet, domain.ErrVoll) {
				t.Errorf("knapp: %v, erwartet %v — %v", knapp, fall.knapp, gemeldet)
			}

			if !errors.Is(gemeldet, domain.ErrEA) {
				t.Errorf("der Fehler des Datenträgers ging verloren: %v", gemeldet)
			}
		})
	}
}
