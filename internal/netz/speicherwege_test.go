// **Was der Zwischenspeicher aushält** — die Wege, auf denen der Ort selbst
// nicht mitmacht (#1163).
//
// Aus `speicher_test.go` herausgelöst, als es über 400 Zeilen ging (§2).
package netz_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Ein Speicherort, der die Nachbardatei nicht annimmt, macht die Antwort
// nicht kaputt.**
//
// Der Unterschied zum Lauf darüber: dort scheitert schon das **Anlegen des
// Verzeichnisses**. Hier steht es da und ist trotzdem zu — der Fall, den seit
// #1163 das
// Anlegen einer eindeutigen Nachbardatei trifft.
func TestEinZugesperrterSpeicherMachtDieAntwortNichtKaputt(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("als root greift keine Rechteprüfung")
	}

	ort := filepath.Join(t.TempDir(), "speicher")

	jetzt := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)
	speicher := netz.NeuerSpeicher(ort, time.Hour, func() time.Time { return jetzt })

	// **Ein erster Eintrag legt das Verzeichnis an.**
	if _, _, fehler := speicher.Hole(t.Context(), "karte/eins",
		func() ([]byte, error) { return []byte("gut"), nil }); fehler != nil {
		t.Fatalf("der erste Eintrag: %v", fehler)
	}

	// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht.
	if err := os.Chmod(ort, 0o500); err != nil {
		t.Fatalf("zusperren: %v", err)
	}

	t.Cleanup(func() {
		// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht.
		_ = os.Chmod(ort, 0o700) //nolint:errcheck // Aufräumen
	})

	inhalt, _, fehler := speicher.Hole(t.Context(), "karte/zwei",
		func() ([]byte, error) { return []byte("auch gut"), nil })
	if fehler != nil {
		t.Fatalf("das Holen scheiterte am Speicherort: %v", fehler)
	}

	// **Die Antwort kommt trotzdem an.** Der Speicher ist eine Beschleunigung
	// und keine Bedingung.
	if string(inhalt) != "auch gut" {
		t.Errorf("der Inhalt ist %q", inhalt)
	}
}
