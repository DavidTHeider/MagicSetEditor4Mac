package laufbeleg_test

// **Gegen echte Ströme und nicht gegen nachgebaute** (AGENTS.md §3).
//
// Die drei Dateien unter `testdata/` sind die Ausgabe von `go test -json`
// aus Go 1.27.1 über ein Probemodul, erzeugt am 2026-09-22: ein bestandener
// Fall mit Unterfall, ein roter, ein roter mit Kennzahl, ein übersprungener,
// ein Absturz mit einem Fall dahinter, ein Übersetzungsfehler, eine
// Zeitüberschreitung und ein Treffer aus dem Zwischenspeicher. **Zwei
// Annahmen hat die Probe widerlegt** — dass ein Treffer aus dem Zwischenspeicher
// keine Zeit trage, und dass eine Zeitüberschreitung den Fall beende.

import (
	"os"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/laufbeleg"
)

// strom liest eine Datei aus testdata.
func strom(t *testing.T, name string) []laufbeleg.Ereignis {
	t.Helper()

	datei, fehler := os.Open("testdata/" + name) //nolint:gosec // fester Name im Baum
	if fehler != nil {
		t.Fatalf("%s öffnen: %v", name, fehler)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	ereignisse, fremd, fehler := laufbeleg.LiesEreignisse(datei)
	if fehler != nil || len(fremd) > 0 {
		t.Fatalf("%s lesen: %v, fremd %q", name, fehler, fremd)
	}

	return ereignisse
}

// **Jede Zeile des echten Stroms ist ein Ereignis**, auch die des Baus.
func TestEinEchterStromWirdGanzGelesen(t *testing.T) {
	t.Parallel()

	ereignisse := strom(t, "strom.jsonl")
	if len(ereignisse) != 69 {
		t.Fatalf("%d Ereignisse, erwartet 69", len(ereignisse))
	}

	bau := 0

	for _, ereignis := range ereignisse {
		if strings.HasPrefix(ereignis.Action, "build-") {
			bau++

			if ereignis.ImportPath == "" {
				t.Errorf("ein Bauereignis ohne ImportPath: %+v", ereignis)
			}
		}
	}

	if bau != 3 {
		t.Errorf("%d Bauereignisse, erwartet 3", bau)
	}
}

// **Was kein Ereignis ist, wird zurückgegeben und nicht verschluckt.**
func TestEineFremdeZeileWirdGenannt(t *testing.T) {
	t.Parallel()

	eingabe := "go: downloading example.com/x v1.0.0\n\n" +
		`{"Action":"start","Package":"p"}` + "\n" + `{"Package":"ohne Handlung"}` + "\n"

	ereignisse, fremd, fehler := laufbeleg.LiesEreignisse(strings.NewReader(eingabe))
	if fehler != nil {
		t.Fatal(fehler)
	}

	if len(ereignisse) != 1 || ereignisse[0].Package != "p" {
		t.Errorf("Ereignisse %+v, erwartet eines für p", ereignisse)
	}

	if len(fremd) != 2 || !strings.HasPrefix(fremd[0], "go: downloading") {
		t.Errorf("fremd %q, erwartet die Zeile von go und die ohne Handlung", fremd)
	}
}

// **Eine Zeile über der Grenze ist ein Lesefehler** — und kein stilles Ende.
func TestEineZuLangeZeileIstEinFehler(t *testing.T) {
	t.Parallel()

	lang := `{"Action":"output","Output":"` + strings.Repeat("x", 17<<20) + `"}`

	_, _, fehler := laufbeleg.LiesEreignisse(strings.NewReader(lang))
	if fehler == nil {
		t.Fatal("eine Zeile über sechzehn MiB ging ohne Fehler durch")
	}
}
