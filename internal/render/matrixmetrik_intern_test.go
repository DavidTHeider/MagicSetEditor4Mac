//go:build fremdpakete

// **`MATRIX.TTF` gegen die gemessenen Werte**
// (#597).
//
// # Warum hinter einem Build-Tag
//
// Die Datei gehört zu `Full-Magic-Pack` und kommt nicht ins Repositorium
// (§11). Der Lauf schließt die Lücke, ohne sie zu holen:
//
//	MSE_FREMDPAKETE="$HOME/full-magic-pack/data" \
//	  go test -tags fremdpakete ./internal/render/ -run Matrix -v
//
// # Was er hält
//
// **Die Zahl, an der die Regel gemessen wurde.** wx rechnet für diese Schrift
// mit `ceil(0,860 · Punktgröße)` als Höhe und `ceil(0,633 · Punktgröße)` als
// Grundlinie — am 2026-08-20 über zwanzig beziehungsweise fünf Punktgrößen
// gemessen. Beide Werte stehen in ihrer `OS/2`-Tabelle.
//
// **Ohne diesen Lauf wäre die Regel an einer Schrift belegt, die es hier nicht
// gibt** — und die Gegenprobe stünde nur im Kommentar.
package render

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

// matrixschrift ist der Ort der Schrift neben den Paketen.
func matrixschrift(t *testing.T) []byte {
	t.Helper()

	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	// **Die Schriften liegen neben `data`**, nicht darin.
	pfad := filepath.Join(filepath.Dir(wurzel), "Magic - Fonts", "MATRIX.TTF")

	roh, err := os.ReadFile(pfad) //nolint:gosec // der Pfad kommt aus der Umgebung des Laufs
	if err != nil {
		t.Skipf("keine Schriftdatei unter %s: %v", pfad, err)
	}

	return roh
}

// **`MATRIX.TTF` bekommt die typografische Metrik, nicht den Tintenkasten.**
func TestMatrixNimmtDieTypoMetrik(t *testing.T) {
	schrift, err := SchriftAusDatei(matrixschrift(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// `OS/2` typo: asc 633, desc −167, gap 60 bei em = 1000.
	const (
		erwarteteHoehe      = 0.860
		erwarteteGrundlinie = 0.633

		// **Der Tintenkasten gäbe diese hier** — sie stehen dabei, damit ein
		// Rückfall auf ihn nicht als Erfolg durchgeht.
		tintenhoehe      = 0.957
		tintengrundlinie = 0.779
	)

	t.Logf("Zeilenhöhe %.6f, Grundlinie %.6f", schrift.Zeilenhoehe(), schrift.Grundlinie())

	if math.Abs(schrift.Zeilenhoehe()-erwarteteHoehe) > 1e-9 {
		t.Errorf("Zeilenhöhe %g, erwartet %g — der Tintenkasten gäbe %g",
			schrift.Zeilenhoehe(), erwarteteHoehe, tintenhoehe)
	}

	if math.Abs(schrift.Grundlinie()-erwarteteGrundlinie) > 1e-9 {
		t.Errorf("Grundlinie %g, erwartet %g — der Tintenkasten gäbe %g",
			schrift.Grundlinie(), erwarteteGrundlinie, tintengrundlinie)
	}
}
