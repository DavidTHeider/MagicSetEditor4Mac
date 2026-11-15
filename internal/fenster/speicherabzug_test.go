package fenster

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// **Der Speicherabzug für den Dauerbetrieb** — E06
// (#2278).
// Er sagt, ob der Go-Prozess etwas hält (HeapAlloc) oder nur nicht
// zurückgibt (HeapIdle − HeapReleased), und legt ein Heap-Profil für
// `go tool pprof -base` ab.
func TestDerSpeicherabzugSchreibtEinHeapProfil(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "heap.pprof")

	stand, fehler := speicherabzug(pfad)
	if fehler != nil {
		t.Fatalf("Abzug: %v", fehler)
	}

	if stand.HeapAlloc == 0 || stand.HeapSys < stand.HeapAlloc || stand.Bereinigungen == 0 {
		t.Errorf("Stand %+v", stand)
	}

	inhalt, fehler := os.ReadFile(pfad) // #nosec G304 -- aus t.TempDir()
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	// Ein pprof-Profil ist ein gzip-Strom.
	if !bytes.HasPrefix(inhalt, []byte{0x1f, 0x8b}) {
		t.Errorf("kein gzip am Anfang: % x", inhalt[:min(4, len(inhalt))])
	}
}

// **Ohne Pfad nur die Zahlen** — und ein Ort, an den sich nicht schreiben
// lässt, ist ein Fehler, keine leere Antwort.
func TestDerSpeicherabzugOhnePfadUndMitFalschemOrt(t *testing.T) {
	if stand, fehler := speicherabzug(""); fehler != nil || stand.HeapAlloc == 0 {
		t.Errorf("ohne Pfad: %+v, %v", stand, fehler)
	}

	if _, fehler := speicherabzug(filepath.Join(t.TempDir(), "fehlt", "heap.pprof")); fehler == nil {
		t.Error("ein fehlender Ordner gab keinen Fehler")
	}
}

// zielMitFehler ist ein Ziel, das beim Schreiben oder beim Schließen scheitert.
type zielMitFehler struct{ schreiben, schliessen error }

func (z zielMitFehler) Write(p []byte) (int, error) {
	if z.schreiben != nil {
		return 0, z.schreiben
	}

	return len(p), nil
}

func (z zielMitFehler) Close() error { return z.schliessen }

// **Ein Profil, das sich nicht schreiben oder schließen lässt, ist ein
// Fehler** — sonst läge eine halbe Datei da, und `pprof -base` verglich
// gegen sie.
func TestEinProfilDasNichtAnkommtIstEinFehler(t *testing.T) {
	kaputt := errors.New("voll")

	for name, ziel := range map[string]zielMitFehler{
		"beim Schreiben": {schreiben: kaputt},
		"beim Schließen": {schliessen: kaputt},
	} {
		t.Run(name, func(t *testing.T) {
			if fehler := profilAblegen(ziel); !errors.Is(fehler, kaputt) {
				t.Errorf("Fehler %v", fehler)
			}
		})
	}
}
