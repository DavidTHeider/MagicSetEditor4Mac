//go:build fremdpakete

// Der Lauf der Paritätsmatrix **gegen den echten MSE2-Quelltext**
// (#1058).
//
// # Warum das hinter einem Build-Tag steht
//
// `/Users/david/mse2-quelle` liegt **neben** dem Repositorium und nicht darin
// (§11). Ein Lauf, der ihn braucht, kann in der Pipeline nicht stehen — und
// einer, der ihn braucht und ohne ihn grün ist, wäre schlimmer als keiner.
//
//	MSE2_QUELLE=~/mse2-quelle go test -tags fremdpakete ./internal/werkzeuge/ -run Fremd -v
//
// **Ohne die Umgebungsvariable wird übersprungen**, nicht geraten: ein fester
// Pfad im Quelltext wäre eine Behauptung über den Rechner, auf dem gebaut wird.
//
// # Was er hält
//
//   - Jede Datei, die die Matrix als C++-Quelle nennt, gibt es wirklich.
//   - Jede `.cpp`-Datei in den vermessenen Verzeichnissen steht in der Matrix.
//
// # Was er nicht hält
//
// **Ob die Marke stimmt** — dieselbe Grenze wie beim Wächter nebenan.
//
// **Was unterhalb der Datei liegt.** Eine einzelne Funktion kann fehlen, ohne
// dass er es zeigt; die Körnung der Matrix ist die Datei, und das steht dort.
package werkzeuge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// vermessen sind die Verzeichnisse, über die die Matrix eine Aussage macht.
//
// **Nicht alle**: `src/util`, `src/gfx` und `src/main.cpp` sind Unterbau und
// haben keine nutzersichtbare Einheit. Wer sie aufnimmt, nimmt sie hier auf —
// und dann verlangt der Lauf eine Zeile je Datei.
var vermessen = []string{
	"src/gui/set",
	"src/data/field",
	"src/gui/value",
	"src/data",
	"src/data/format",
	"src/script",
	"src/render/card",
	"src/render/value",
	"src/render/text",
	"src/render/symbol",
}

// TestFremdDieMatrixNenntNurEchteQuellen prüft die eine Richtung.
func TestFremdDieMatrixNenntNurEchteQuellen(t *testing.T) {
	quelle := mse2quelle(t)

	for _, pfad := range quellpfadeDerMatrix(t) {
		if _, err := os.Stat(filepath.Join(quelle, pfad)); err != nil {
			t.Errorf("die Matrix nennt %q — im Quelltext gibt es das nicht", pfad)
		}
	}
}

// TestFremdJedeDateiStehtInDerMatrix prüft die andere.
//
// **Sie ist die wichtigere**: die erste Richtung findet Tippfehler, diese
// findet, was **nie angesehen** wurde — und dafür gibt es die Matrix.
func TestFremdJedeDateiStehtInDerMatrix(t *testing.T) {
	quelle := mse2quelle(t)
	genannt := strings.Join(quellpfadeDerMatrix(t), "\n")

	// **Mit den Unterordnern, seit dem 2026-08-28**
	// (#1170).
	//
	// `os.ReadDir` sah nur die oberste Ebene, und der Wächter übersprang
	// Verzeichnisse ausdrücklich. Damit blieben elf Dateien unbesehen —
	// `src/script/functions/*` (die halbe Standardbibliothek) und
	// `src/data/action/*` (die Rücknahme). Der Lauf war grün, und die Matrix
	// behauptete daneben, sie beantworte „was ist unvermessen?".
	//
	// **Was weiterhin außen bleibt, steht in [vermessen]** — und es steht
	// dort mit Grund, nicht als Auslassung.
	for _, ordner := range vermessen {
		fehler := filepath.WalkDir(filepath.Join(quelle, ordner),
			func(pfad string, eintrag os.DirEntry, err error) error {
				if err != nil || eintrag.IsDir() || !strings.HasSuffix(pfad, ".cpp") {
					return err
				}

				innen, relfehler := filepath.Rel(quelle, pfad)
				if relfehler != nil {
					return relfehler
				}

				if !strings.Contains(genannt, filepath.ToSlash(innen)) {
					t.Errorf("%s steht in keiner Zeile der Matrix", filepath.ToSlash(innen))
				}

				return nil
			})
		if fehler != nil {
			t.Fatalf("%s durchgehen: %v", ordner, fehler)
		}
	}
}

// mse2quelle gibt den Ort des Quelltexts — oder überspringt.
func mse2quelle(t *testing.T) string {
	t.Helper()

	ort := os.Getenv("MSE2_QUELLE")
	if ort == "" {
		t.Skip("ohne MSE2_QUELLE gibt es nichts zu vergleichen")
	}

	return ort
}

// quellpfadeDerMatrix sammelt jedes `src/…`-Vorkommen aus der Matrix.
func quellpfadeDerMatrix(t *testing.T) []string {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join("..", "..", "docs", "PARITAET.md"))
	if err != nil {
		t.Fatalf("PARITAET.md lesen: %v", err)
	}

	var pfade []string

	// **Nur aus Tabellenzeilen.** Der Fließtext nennt Verzeichnisse als Muster
	// (`src/gui/*.cpp`, siebzehn Dateien), und die gibt es als Datei nicht —
	// gesehen, nicht überlegt: der erste Anlauf meldete drei davon als fehlend.
	for _, zeile := range strings.Split(string(roh), "\n") {
		if !strings.HasPrefix(zeile, "|") {
			continue
		}

		for _, stueck := range strings.Split(zeile, "`") {
			if !strings.HasPrefix(stueck, "src/") || !strings.HasSuffix(stueck, ".cpp") {
				continue
			}

			pfade = append(pfade, stueck)
		}
	}

	return pfade
}
