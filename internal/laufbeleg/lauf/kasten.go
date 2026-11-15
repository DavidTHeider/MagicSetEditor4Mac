package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// werkzeugkasten sind die Programme, die ein Lauf startet — **eingesetzt**,
// damit ein Test sie ersetzen kann (AGENTS.md §16). Die echten stehen in
// [echterKasten], und `kasten_test.go` fährt sie gegen ein Probemodul.
type werkzeugkasten struct {
	git      func(ctx context.Context, argumente ...string) (string, error)
	gotest   func(ctx context.Context, argumente []string, aus, fehler io.Writer) (int, error)
	version  func(ctx context.Context) string
	system   func(ctx context.Context) string
	jetzt    func() time.Time
	umgebung func(name string) string
}

// echterKasten gibt die echten Werkzeuge.
//
// **Jedes Programm steht als Literal da** — `internal/schichten` weist einen
// Unterprozess ab, dessen Programm aus einer Veränderlichen kommt.
func echterKasten() werkzeugkasten {
	return werkzeugkasten{
		git: func(ctx context.Context, argumente ...string) (string, error) {
			ausgabe, fehler := exec.CommandContext(ctx, "git", argumente...).Output() // #nosec G204 -- das Programm ist ein Literal, die Argumente baut der Lauf

			return strings.TrimSpace(string(ausgabe)), fehler
		},
		gotest: func(ctx context.Context, argumente []string, aus, fehler io.Writer) (int, error) {
			befehl := exec.CommandContext(ctx, "go", argumente...) // #nosec G204 -- das Programm ist ein Literal, die Argumente baut der Lauf aus dem Plan
			befehl.Stdout, befehl.Stderr = aus, fehler

			// **Ein Rückgabewert ist ein Ergebnis und kein Fehler**: `go test`
			// endet mit 1, wenn ein Fall rot ist. Ein Fehler ist nur, was den
			// Prozess gar nicht erst starten lässt.
			var beendet *exec.ExitError

			switch gelaufen := befehl.Run(); {
			case errors.As(gelaufen, &beendet):
				return beendet.ExitCode(), nil
			case gelaufen != nil:
				return -1, gelaufen
			default:
				return 0, nil
			}
		},
		version: func(ctx context.Context) string {
			return antwortOder(exec.CommandContext(ctx, "go", "env", "GOVERSION"), "unbekannt")
		},
		system: func(ctx context.Context) string {
			fassung := antwortOder(exec.CommandContext(ctx, "sw_vers", "-productVersion"), "")
			if fassung == "" {
				return runtime.GOOS
			}

			return "macOS " + fassung
		},
		jetzt:    time.Now,
		umgebung: os.Getenv,
	}
}

// antwortOder gibt die Ausgabe eines Befehls — oder, wenn er nicht antwortet,
// was sonst dastehen soll.
func antwortOder(befehl *exec.Cmd, sonst string) string {
	ausgabe, fehler := befehl.Output()
	if fehler != nil {
		return sonst
	}

	return strings.TrimSpace(string(ausgabe))
}
