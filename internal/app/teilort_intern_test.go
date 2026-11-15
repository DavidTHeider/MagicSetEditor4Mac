package app

// **Wo die Teilungen liegen, und wann aufgeräumt wird** — NET-04.AK-04
// (#1708).
//
// **Nicht parallel**: die Läufe setzen `HOME`.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDieTeilungenLiegenUnterCaches(t *testing.T) {
	heim := t.TempDir()
	t.Setenv("HOME", heim)

	ort, fehler := Teilverzeichnis()
	if fehler != nil || ort != filepath.Join(heim, "Library", "Caches", "MagicSetEditor4Mac", "teilen") {
		t.Fatalf("der Ort: %q, %v", ort, fehler)
	}

	// **Der Ort entsteht erst beim ersten Teilen** — wer nie teilt, hat
	// keinen Ordner dafür.
	anwendung := NeueApp("")
	anwendung.teilenEinrichten(true)

	if _, fehler := os.Stat(ort); !os.IsNotExist(fehler) {
		t.Errorf("der Teilordner entstand beim Einrichten: %v", fehler)
	}

	if gut := anwendung.TeilenVorbereiten(context.Background(), "einladung", "code").Teilgut; gut == nil ||
		!strings.HasPrefix(filepath.Join(ort, gut.Kennung), ort) {
		t.Fatalf("das Teilen nach dem Einrichten: %+v", gut)
	}
}

// **Beim Start wird weggeräumt, was ein Tag alt ist** — aber nie aus einem
// Testlauf, der es nicht ausdrücklich erlaubt.
func TestDerStartRaeumtAlteTeilungen(t *testing.T) {
	heim := t.TempDir()
	t.Setenv("HOME", heim)

	ort, _ := Teilverzeichnis() //nolint:errcheck // dasselbe Zuhause wie eben
	alt := filepath.Join(ort, strings.Repeat("a", 32))

	if fehler := os.MkdirAll(alt, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	vorher := time.Now().Add(-48 * time.Hour)
	if fehler := os.Chtimes(alt, vorher, vorher); fehler != nil {
		t.Fatal(fehler)
	}

	// **Ohne Erlaubnis bleibt sie liegen** — die Sperre für Testläufe.
	NeueApp("").teilenEinrichten(false)

	if _, fehler := os.Stat(alt); fehler != nil {
		t.Fatalf("ohne Erlaubnis wurde geräumt: %v", fehler)
	}

	NeueApp("").teilenEinrichten(true)

	if _, fehler := os.Stat(alt); !os.IsNotExist(fehler) {
		t.Errorf("die alte Teilung liegt noch: %v", fehler)
	}
}

// **Ohne Zuhause gibt es keinen Ort** — und die Anwendung teilt dann nicht.
func TestOhneZuhauseKeinTeilort(t *testing.T) {
	t.Setenv("HOME", "")

	if _, fehler := Teilverzeichnis(); fehler == nil {
		t.Error("ohne Zuhause gab es einen Ort")
	}

	anwendung := NeueApp("")
	anwendung.teilenEinrichten(true)

	if fehler := anwendung.TeilenVorbereiten(context.Background(), "einladung", "code").Fehler; fehler == nil {
		t.Error("ohne Ort kam eine Teilung zustande")
	}
}
