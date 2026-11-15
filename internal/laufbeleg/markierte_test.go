package laufbeleg_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/laufbeleg"
)

// **Erwartet ist, was erst der Tag in den Lauf bringt** — berechnet und
// nicht eingetragen: eine Liste, die jemand pflegt, vergäße den nächsten Lauf,
// und der fehlte dann still (AGENTS.md §10).
func TestMarkierteFaelleWerdenBerechnet(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()

	schreibe(t, wurzel, "go.mod", "module beispiel.test\n\ngo 1.27\n")
	schreibe(t, wurzel, "app/kanon_test.go", `//go:build fremdpakete

package app

import "testing"

func TestKanon(t *testing.T) {}

func TestMain(m *testing.M) {}

func TestZweiSachen(t *testing.T, anderes int) {}

func BenchmarkKanon(b *testing.B) {}

func Testhilfe(t *testing.T) {}

func hilfe(t *testing.T) {}

type vorrichtung struct{}

func (vorrichtung) TestAlsMethode(t *testing.T) {}

type T struct{}

func TestWert(t testing.T) {}

func TestLokal(t *T) {}

func TestErgebnis(t *testing.T) int { return 0 }

func TestZweiNamen(a, b *testing.T) {}

func Test(t *testing.T) {}
`)
	schreibe(t, wurzel, "app/oder_test.go", `//go:build fremdpakete || netzlauf

package app_test

import tst "testing"

func TestOder(t *tst.T) {}
`)
	schreibe(t, wurzel, "app/immer_test.go", "//go:build darwin || linux || windows\n\npackage app\n\nimport \"testing\"\n\nfunc TestImmer(t *testing.T) {}\n")
	schreibe(t, wurzel, "app/nie_test.go", "//go:build !fremdpakete\n\npackage app\n\nimport \"testing\"\n\nfunc TestNie(t *testing.T) {}\n")
	schreibe(t, wurzel, "app/ohne_test.go", "package app\n\nimport \"testing\"\n\nfunc TestOhne(t *testing.T) {}\n")
	schreibe(t, wurzel, "app/fremd_plan9_test.go", "//go:build fremdpakete\n\npackage app\n\nimport \"testing\"\n\nfunc TestPlan9(t *testing.T) {}\n")
	schreibe(t, wurzel, "app/tief/unten_test.go", "//go:build fremdpakete\n\npackage tief\n\nimport \"testing\"\n\nfunc TestUnten(t *testing.T) {}\n")

	// **Was nicht zum Modul gehört, zählt nicht**: ein eigenes Modul, Testdaten,
	// Verborgenes und fremde Abhängigkeiten.
	for _, fremd := range []string{
		"tools/x_test.go", "app/testdata/x_test.go",
		".go/x_test.go", "frontend/node_modules/x_test.go",
	} {
		schreibe(t, wurzel, fremd, "//go:build fremdpakete\n\npackage x\n\nimport \"testing\"\n\nfunc TestFremd(t *testing.T) {}\n")
	}

	schreibe(t, wurzel, "tools/go.mod", "module werkzeug.test\n")

	faelle, fehler := laufbeleg.Markierte(wurzel, "fremdpakete")
	if fehler != nil {
		t.Fatal(fehler)
	}

	erwartet := []laufbeleg.Fall{
		{Paket: "beispiel.test/app", Test: "Test"},
		{Paket: "beispiel.test/app", Test: "TestKanon"},
		{Paket: "beispiel.test/app", Test: "TestOder"},
		{Paket: "beispiel.test/app/tief", Test: "TestUnten"},
	}

	if !reflect.DeepEqual(faelle, erwartet) {
		t.Errorf("markiert %+v\nerwartet %+v", faelle, erwartet)
	}
}

// **Was nicht zu lesen ist, ist ein Fehler** — und keine kürzere Liste.
func TestEineKaputteDateiIstEinFehler(t *testing.T) {
	t.Parallel()

	faelle := map[string]map[string]string{
		"ohne go.mod": {"a/x_test.go": "//go:build fremdpakete\n\npackage a\n"},
		"ein go.mod ohne Modul": {
			"go.mod": "go 1.27\n", "a/x_test.go": "//go:build fremdpakete\n\npackage a\n",
		},
		"eine Datei, die kein Go ist": {
			"go.mod": "module b.test\n", "a/x_test.go": "//go:build fremdpakete\n\npackage a\n\nfunc {\n",
		},
	}

	for name, dateien := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wurzel := t.TempDir()
			for datei, inhalt := range dateien {
				schreibe(t, wurzel, datei, inhalt)
			}

			if _, fehler := laufbeleg.Markierte(wurzel, "fremdpakete"); fehler == nil {
				t.Error("kein Fehler")
			}
		})
	}

	if _, fehler := laufbeleg.Markierte(filepath.Join(t.TempDir(), "gibtsnicht"), "fremdpakete"); fehler == nil ||
		!strings.Contains(fehler.Error(), "gibtsnicht") {
		t.Errorf("eine fehlende Wurzel gab %v", fehler)
	}
}

// **Was sich nicht lesen lässt, ist ein Fehler** — ein gesperrtes
// Verzeichnis verkürzte die Liste sonst still um alles, was darin liegt.
func TestEinGesperrterOrtIstEinFehler(t *testing.T) {
	t.Parallel()

	for _, gesperrt := range []string{"a", "a/x_test.go"} {
		t.Run(gesperrt, func(t *testing.T) {
			t.Parallel()

			wurzel := t.TempDir()
			schreibe(t, wurzel, "go.mod", "module c.test\n")
			schreibe(t, wurzel, "a/x_test.go", "//go:build fremdpakete\n\npackage a\n")

			ziel := filepath.Join(wurzel, gesperrt)
			if fehler := os.Chmod(ziel, 0o000); fehler != nil {
				t.Fatal(fehler)
			}

			t.Cleanup(func() { _ = os.Chmod(ziel, 0o750) }) //nolint:errcheck,gosec // nur zum Aufräumen; ein Verzeichnis braucht das Ausführungsrecht

			if _, fehler := laufbeleg.Markierte(wurzel, "fremdpakete"); fehler == nil {
				t.Error("ein gesperrter Ort ging ohne Fehler durch")
			}
		})
	}
}

// **Der Modulpfad kommt aus dem go.mod** — mit und ohne Anführungszeichen.
func TestDerModulpfadKommtAusDemGoMod(t *testing.T) {
	t.Parallel()

	for _, zeile := range []string{"module a.test/b\n", "// Kopf\nmodule \"a.test/b\"\n\ngo 1.27\n"} {
		wurzel := t.TempDir()
		schreibe(t, wurzel, "go.mod", zeile)

		if modul, fehler := laufbeleg.Modul(wurzel); modul != "a.test/b" || fehler != nil {
			t.Errorf("%q gab %q, %v", zeile, modul, fehler)
		}
	}
}
