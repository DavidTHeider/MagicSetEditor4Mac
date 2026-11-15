package main

// **Die echten Werkzeuge an der Naht** — `git`, `go test`, `go env` und
// `sw_vers`, jeweils einmal wirklich gestartet.
//
// **Das Probemodul unter `testdata/probe`** hat einen grünen und einen roten
// Fall. Es liegt unter `testdata`, damit `go` es aus diesem Modul heraus nie
// als Paket sieht, und `-C` fährt es als eigenes Modul.

import (
	"bytes"
	"context"
	"runtime"
	"strings"
	"testing"
)

// **Jedes Werkzeug antwortet, wie der Lauf es erwartet.**
func TestDieEchtenWerkzeugeAntworten(t *testing.T) {
	t.Parallel()

	kasten := echterKasten()
	ctx := context.Background()

	if drin, fehler := kasten.git(ctx, "rev-parse", "--is-inside-work-tree"); drin != "true" || fehler != nil {
		t.Errorf("git gab %q, %v", drin, fehler)
	}

	if _, fehler := kasten.git(ctx, "gibtsnicht"); fehler == nil {
		t.Error("ein unbekannter Befehl von git gab keinen Fehler")
	}

	if fassung := kasten.version(ctx); !strings.HasPrefix(fassung, "go1.") {
		t.Errorf("go env GOVERSION gab %q", fassung)
	}

	if system := kasten.system(ctx); runtime.GOOS == "darwin" && !strings.HasPrefix(system, "macOS ") {
		t.Errorf("sw_vers gab %q", system)
	}

	if kasten.jetzt().IsZero() || kasten.umgebung("PATH") == "" {
		t.Error("Uhr oder Umgebung antworten nicht")
	}
}

// **`go test` meldet ein rotes Ergebnis als Rückgabe und nicht als Fehler** —
// und den Strom im Ganzen.
func TestGoTestLaeuftWirklich(t *testing.T) {
	t.Parallel()

	var aus, fehler bytes.Buffer

	rueckgabe, startfehler := echterKasten().gotest(context.Background(),
		[]string{"-C", "testdata/probe", "test", "-json", "-count=1", "./..."}, &aus, &fehler)
	if startfehler != nil || rueckgabe != 1 {
		t.Fatalf("Rückgabe %d, Fehler %v\n%s", rueckgabe, startfehler, fehler.String())
	}

	for _, erwartet := range []string{
		`"Test":"TestGut"`, `"Action":"fail","Package":"probe.test/rot"`,
		"kennzahl probe.zahl=7",
	} {
		if !strings.Contains(aus.String(), erwartet) {
			t.Errorf("im Strom fehlt %s", erwartet)
		}
	}

	// **Und ein grüner Lauf endet mit null.**
	if rueckgabe, startfehler := echterKasten().gotest(context.Background(),
		[]string{"-C", "testdata/probe", "test", "-json", "-count=1", "./gut"}, &bytes.Buffer{},
		&bytes.Buffer{}); startfehler != nil || rueckgabe != 0 {
		t.Errorf("Rückgabe %d, Fehler %v", rueckgabe, startfehler)
	}
}

// **Was nicht startet, ist ein Fehler** — und wer nicht antwortet, bekommt
// die Ersatzangabe statt einer leeren.
func TestWasNichtStartetIstEinFehler(t *testing.T) {
	t.Parallel()

	abgebrochen, abbrechen := context.WithCancel(context.Background())
	abbrechen()

	kasten := echterKasten()

	if rueckgabe, fehler := kasten.gotest(abgebrochen, []string{"version"}, &bytes.Buffer{},
		&bytes.Buffer{}); fehler == nil || rueckgabe != -1 {
		t.Errorf("Rückgabe %d, Fehler %v", rueckgabe, fehler)
	}

	if fassung := kasten.version(abgebrochen); fassung != "unbekannt" {
		t.Errorf("go env gab %q", fassung)
	}

	if system := kasten.system(abgebrochen); system != runtime.GOOS {
		t.Errorf("sw_vers gab %q", system)
	}
}
