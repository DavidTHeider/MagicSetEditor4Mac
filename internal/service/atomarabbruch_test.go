package service_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Abbruch mitten im Schreiben lässt die alte Datei stehen** —
// FMT-06.AK-03 (#1695).
//
// # Was die übrigen Läufe nicht zeigen
//
// Sie lassen das Schreiben **scheitern** — an einem Verzeichnis im Weg, an
// einem Verzeichnis ohne Schreibrecht. Ein Prozess, der mittendrin stirbt,
// scheitert nicht: er ist einfach weg, und kein Aufräumen läuft mehr. Die
// Zusage aus AK-03 lautet trotzdem: danach steht der alte gültige Stand oder
// ein vollständig gültiger neuer — nie ein halber.
//
// # Wie er entsteht
//
// **Das Testprogramm ruft sich selbst** als Kindprozess, das in einer
// Schleife 128 MiB über [service.SchreibeAtomar] schreibt. Getötet wird es
// (SIGKILL) **zu 24 verschiedenen Zeitpunkten**, im Abstand einer halben
// Millisekunde, nachdem seine Nachbardatei steht; jeder Versuch muss am Ziel
// alt oder vollständig neu finden.
//
// **Warum ein Zeitplan und nicht „sofort"**: gemessen am 2026-09-23 schreiben
// sich hier 64 MiB in rund 6 ms. Sofort getötet hatte das Kind noch kein Byte
// geschrieben — und dann bliebe das Ziel auch bei einem Schreiben alt, das es
// mittendrin kürzt. Die Gegenprobe, ein solches Schreiben, blieb so grün; mit
// dem Zeitplan wird sie **in zwanzig Läufen neunzehnmal** rot (gemessen am
// 2026-09-23). Ein Wächter mit einer Trefferrate, und die gehört gesagt: ein
// Rückfall auf ein kürzendes Schreiben fiele in einem von zwanzig Läufen
// nicht auf.
//
// # Was liegen bleiben darf, und das ist benannt
//
// **Die Nachbardatei** mit führendem Punkt (`schreibeAtomar`, „Und was es
// nicht hält"). Sie ist sichtbar und nicht ein halbes `.mse-set` unter dem
// richtigen Namen.
func TestEinAbbruchMittenImSchreibenLaesstDieAlteDateiStehen(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	ziel := filepath.Join(ordner, "set")
	alt := bytes.Repeat([]byte("alt\n"), 1024)
	neu := abbruchinhalt()

	for versuch := range 24 {
		if fehler := os.WriteFile(ziel, alt, 0o600); fehler != nil {
			t.Fatalf("vorbereiten: %v", fehler)
		}

		// #nosec G204 G702 -- das eigene Testprogramm, mit festem Lauf
		kind := exec.CommandContext(t.Context(), os.Args[0],
			"-test.run=^TestAbbruchKindSchreibtOhneEnde$")
		kind.Env = append(os.Environ(), "MSE_ABBRUCHZIEL="+ziel)

		if fehler := kind.Start(); fehler != nil {
			t.Fatalf("das Kind starten: %v", fehler)
		}

		warteAufNachbardatei(t, ordner)
		time.Sleep(time.Duration(versuch) * 500 * time.Microsecond)

		if fehler := kind.Process.Kill(); fehler != nil {
			t.Fatalf("das Kind töten: %v", fehler)
		}

		_ = kind.Wait() //nolint:errcheck // getötet — sein Ausgang ist der Grund

		inhalt, fehler := os.ReadFile(ziel) // #nosec G304 -- t.TempDir()
		if fehler != nil {
			t.Fatalf("Versuch %d: das Ziel lesen: %v", versuch, fehler)
		}

		if !bytes.Equal(inhalt, alt) && !bytes.Equal(inhalt, neu) {
			t.Fatalf("Versuch %d: das Ziel ist weder alt noch vollständig neu (%d Bytes)",
				versuch, len(inhalt))
		}

		raeumeNachbardateien(t, ordner)
	}
}

// TestAbbruchKindSchreibtOhneEnde ist das Kind — es schreibt, bis es
// getötet wird.
//
// **Übersprungen, wenn es kein Kind ist**, und ausdrücklich: ein Lauf, der
// hier grün meldete, ohne zu schreiben, stünde als bestanden da und prüfte
// nichts.
func TestAbbruchKindSchreibtOhneEnde(t *testing.T) {
	ziel := os.Getenv("MSE_ABBRUCHZIEL")
	if ziel == "" {
		t.Skip("läuft nur als Kindprozess von TestEinAbbruchMittenImSchreibenLaesstDieAlteDateiStehen")
	}

	neu := abbruchinhalt()

	for {
		if fehler := service.SchreibeAtomar(ziel, neu); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}
	}
}

// abbruchinhalt ist, was das Kind schreibt — groß genug, dass das Schreiben
// dauert, und in beiden Prozessen derselbe.
func abbruchinhalt() []byte {
	return bytes.Repeat([]byte("neu\n"), 32<<20)
}

// warteAufNachbardatei wartet, bis im Ordner eine Nachbardatei steht.
func warteAufNachbardatei(t *testing.T, ordner string) {
	t.Helper()

	frist := time.Now().Add(30 * time.Second)

	for time.Now().Before(frist) {
		if len(nachbardateien(t, ordner)) > 0 {
			return
		}

		time.Sleep(100 * time.Microsecond)
	}

	t.Fatal("das Kind hat in 30 Sekunden keine Nachbardatei angelegt")
}

// raeumeNachbardateien entfernt, was ein getötetes Kind liegen ließ.
func raeumeNachbardateien(t *testing.T, ordner string) {
	t.Helper()

	for _, name := range nachbardateien(t, ordner) {
		if fehler := os.Remove(filepath.Join(ordner, name)); fehler != nil && !os.IsNotExist(fehler) {
			t.Fatalf("räumen: %v", fehler)
		}
	}
}

// nachbardateien gibt die Namen der Nachbardateien des Ziels `set`.
func nachbardateien(t *testing.T, ordner string) []string {
	t.Helper()

	eintraege, fehler := os.ReadDir(ordner)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	var namen []string

	for _, eintrag := range eintraege {
		if strings.HasPrefix(eintrag.Name(), ".set.") {
			namen = append(namen, eintrag.Name())
		}
	}

	return namen
}
