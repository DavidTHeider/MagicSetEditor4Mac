//go:build fremdpakete

// Der Wächter über die **Vollständigkeit** des Symbolmodells
// (#890,
// Werkstück 2 von Phase X).
//
// # Warum hinter einem Build-Tag
//
// `/Users/david/mse2-quelle` liegt **neben** dem Repositorium und nicht darin
// (§11). Ein Lauf, der ihn braucht, kann in der Pipeline nicht stehen — und
// einer, der ihn braucht und ohne ihn grün ist, wäre schlimmer als keiner.
//
//	MSE2_QUELLE=~/mse2-quelle go test -tags fremdpakete ./internal/msepaket/ -run Symbolarten -v
//
// # Was er hält
//
// Jeder Aufzählungswert, den MSE in eine `.mse-symbol` schreibt, hat hier eine
// Entsprechung — und **umgekehrt**: keine Konstante hier, die MSE nicht kennt.
//
// # Was er nicht hält
//
// **Was ein Wert bedeutet.** Dass `border` einen Rand zieht und `subtract`
// abzieht, ist eine Aussage über die Rasterung; hier steht nur, dass der Name
// gelesen und geschrieben werden kann.
//
// **Die Felder eines Blocks.** Sie stehen in `symbolschreiber.go` einzeln mit
// ihrer Fundstelle; ein Wächter darüber müsste C++ zerteilen und prüfte am
// Ende seinen eigenen Zerteiler.
package msepaket_test

import (
	"io/fs"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// TestSymbolartenSindVollstaendig hält beide Richtungen.
func TestSymbolartenSindVollstaendig(t *testing.T) {
	t.Parallel()

	quelle := os.Getenv("MSE2_QUELLE")
	if quelle == "" {
		t.Skip("MSE2_QUELLE nicht gesetzt — der Quelltext liegt neben dem Repositorium (§11)")
	}

	// **Über eine Wurzel und nicht über einen zusammengesetzten Pfad.**
	// `os.Root` bindet jeden Zugriff an das Verzeichnis aus `MSE2_QUELLE`; ein
	// `filepath.Join` daneben wäre für den Prüfer ein Pfad aus einer fremden
	// Quelle — und der Prüfer hätte recht.
	wurzel, wurzelfehler := os.OpenRoot(quelle)
	if wurzelfehler != nil {
		t.Skipf("MSE2_QUELLE %s: %v", quelle, wurzelfehler)
	}

	t.Cleanup(func() {
		if err := wurzel.Close(); err != nil {
			t.Errorf("die Wurzel schließen: %v", err)
		}
	})

	roh, err := fs.ReadFile(wurzel.FS(), "src/data/symbol.cpp")
	if err != nil {
		t.Skipf("symbol.cpp: %v", err)
	}

	for _, fall := range []struct {
		name        string
		aufzaehlung string
		hier        []string
	}{
		{
			name:        "Kombinationsarten",
			aufzaehlung: "SymbolShapeCombine",
			hier: []string{
				string(msepaket.KombinationUeberdecken),
				string(msepaket.KombinationVereinen),
				string(msepaket.KombinationAbziehen),
				string(msepaket.KombinationSchneiden),
				string(msepaket.KombinationUnterschied),
				string(msepaket.KombinationUmranden),
			},
		},
		{
			name:        "Symmetriearten",
			aufzaehlung: "SymbolSymmetryType",
			hier: []string{
				string(msepaket.SpiegelungDrehung),
				string(msepaket.SpiegelungAchse),
			},
		},
		{
			name:        "Sperren",
			aufzaehlung: "LockMode",
			hier:        []string{"free", "direction", "size"},
		},
		{
			name:        "Streckenarten",
			aufzaehlung: "SegmentMode",
			hier:        []string{"line", "curve"},
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			dort := werteAus(t, string(roh), fall.aufzaehlung)
			if len(dort) == 0 {
				t.Fatalf("in symbol.cpp steht keine Aufzählung %q — der Wächter greift ins Leere",
					fall.aufzaehlung)
			}

			for _, wert := range dort {
				if !enthaelt(fall.hier, wert) {
					t.Errorf("MSE kennt %q, dieses Programm nicht", wert)
				}
			}

			for _, wert := range fall.hier {
				if !enthaelt(dort, wert) {
					t.Errorf("dieses Programm kennt %q, MSE nicht", wert)
				}
			}
		})
	}
}

// werteAus liest die Namen einer `IMPLEMENT_REFLECTION_ENUM`.
//
//	IMPLEMENT_REFLECTION_ENUM(SymbolShapeCombine) {
//	  VALUE_N("overlap",    SYMBOL_COMBINE_OVERLAP);
//	  …
//	}
func werteAus(t *testing.T, quelltext, aufzaehlung string) []string {
	t.Helper()

	anfang := strings.Index(quelltext, "IMPLEMENT_REFLECTION_ENUM("+aufzaehlung+")")
	if anfang < 0 {
		return nil
	}

	ende := strings.Index(quelltext[anfang:], "\n}")
	if ende < 0 {
		return nil
	}

	muster := regexp.MustCompile(`VALUE_N\("([^"]+)"`)

	var gefunden []string
	for _, treffer := range muster.FindAllStringSubmatch(quelltext[anfang:anfang+ende], -1) {
		gefunden = append(gefunden, treffer[1])
	}

	return gefunden
}

func enthaelt(werte []string, gesucht string) bool {
	for _, wert := range werte {
		if wert == gesucht {
			return true
		}
	}

	return false
}
