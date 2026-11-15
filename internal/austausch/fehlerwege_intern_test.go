package austausch

// **Die Wege, auf denen ein Umschlag nicht zustande kommt** — Werkstück N4
// von Phase 5.4 (#1457).
//
// **Innen**, weil einer davon nur über die Uhrnaht zu erreichen ist.

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// **Ein Manifest, das sich nicht schreiben lässt.**
//
// `time.Time.MarshalJSON` scheitert bei einem Jahr außerhalb von [0, 9999] —
// dieselbe dokumentierte Eigenschaft, an der auch `internal/sicherung` seinen
// Schreibfehler festmacht.
func TestEinUnschreibbaresManifestWirdGemeldet(t *testing.T) {
	vorher := jetzt
	jetzt = func() time.Time { return time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) }

	t.Cleanup(func() { jetzt = vorher })

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	if _, schreibfehler := Schreibe(
		filepath.Join(t.TempDir(), "x"+Endung),
		[]Inhalt{{Art: ArtKarte, Name: "a.json", Bytes: []byte("x")}},
		ich,
	); schreibfehler == nil {
		t.Error("ein unschreibbares Manifest kam durch")
	}
}

// **Ein Ort, an dem sich nicht schreiben lässt.**
func TestEinUnbeschreibbarerOrtWirdGemeldet(t *testing.T) {
	t.Parallel()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	// **Ein Verzeichnis, das es nicht gibt** — und kein Versuch, es
	// anzulegen: ein Umschlag wird dorthin geschrieben, wo jemand hinzeigt.
	if _, schreibfehler := Schreibe(
		filepath.Join(t.TempDir(), "gibtesnicht", "x"+Endung),
		[]Inhalt{{Art: ArtKarte, Name: "a.json", Bytes: []byte("x")}},
		ich,
	); schreibfehler == nil {
		t.Error("ein unbeschreibbarer Ort kam durch")
	}
}
