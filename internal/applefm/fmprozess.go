package applefm

// **Die zwei Griffe an `/usr/bin/fm`** — herausgelöst, weil sie im Testlauf
// nie laufen dürfen (§3, [gesperrt]) und deshalb in `.covignore` stehen. Was
// darunter entscheidet — [exitcodeVon], [starte], der Riegel —, steht in
// `dienst.go` und wird gemessen; den echten Weg fährt der Echtlauf hinter
// `-tags fmlauf`.

import (
	"context"
	"os/exec"
)

// Exitcode ist der echte [Ausfuehrer] — mit `NO_COLOR`, ohne Eingabe.
//
// **Der Pfad steht als Literal da**, dieselbe Zeichenkette wie [Programm]:
// `internal/schichten/unterprozess_test.go` verlangt, dass ein gestartetes
// Programm im Quelltext steht und nicht aus einer Veränderlichen kommt.
func Exitcode(ctx context.Context, argumente ...string) (int, error) {
	if gesperrt() {
		return -1, ErrImLauf
	}

	// #nosec G204 -- fester Pfad; die Argumente setzt dieses Paket
	return exitcodeVon(exec.CommandContext(ctx, "/usr/bin/fm", argumente...)) //nolint:gosec // s. o.
}

// Prozessstarter ist der echte [Starter] — **in einem Testlauf gesperrt**.
//
// **Ohne den Kontext des Rufs**: der Prozess lebt so lange wie der Dienst,
// nicht wie die Frage, die ihn gestartet hat.
func Prozessstarter(_ context.Context, argumente ...string) (Prozess, error) {
	if gesperrt() {
		return nil, ErrImLauf
	}

	// #nosec G204 -- fester Pfad; die Argumente setzt dieses Paket
	return starte(exec.Command("/usr/bin/fm", argumente...)) //nolint:noctx,gosec // lebt länger als der Ruf
}
