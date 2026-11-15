package app

// **Der Abdruck einer Prüfung sieht die Dateien der Pakete** — Befund 2 des
// Zweitblicks über Phase 3 von 0.9.0 (PERF02, #2352).

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// abdruckleser führt ein Set auf `probe-schlicht` mit festem Setstand.
type abdruckleser struct{ Zeichenleser }

func (abdruckleser) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return []domain.SetHead{{ID: 7, Style: "probe-schlicht", Game: "probe"}}, nil
}

func (abdruckleser) Setwerte(context.Context, int64) ([]domain.Value, error) { return nil, nil }

func (abdruckleser) Setstand(context.Context, int64) (int64, error) { return 3, nil }

func TestDerAbdruckSiehtDieDateienDerPakete(t *testing.T) {
	wurzel := t.TempDir()
	for _, name := range []string{"probe.mse-game", "probe-schlicht.mse-style"} {
		if err := os.CopyFS(filepath.Join(wurzel, name), os.DirFS(filepath.Join("..", "..", "saat", name))); err != nil {
			t.Fatal(err)
		}
	}

	leser := abdruckleser{}
	werk := &Zeichenwerk{Leser: leser, Schriften: schriften.Neue(nil), Stile: service.NeuerStilspeicher(leser, wurzel)}

	vorher, err := werk.problemabdruck(t.Context(), 7)
	if err != nil {
		t.Fatal(err)
	}

	stil := filepath.Join(wurzel, "probe-schlicht.mse-style", "style")

	roh, err := os.ReadFile(stil) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(stil, []byte(strings.Replace(string(roh), "card width: 375", "card width: 376", 1)), 0o600); err != nil { //nolint:gosec // aus t.TempDir()
		t.Fatal(err)
	}

	if nachher, err := werk.problemabdruck(t.Context(), 7); err != nil || nachher == vorher {
		t.Errorf("Abdruck %q nach %q (%v) — die geänderte Stildatei blieb ungesehen", nachher, vorher, err)
	}
}
