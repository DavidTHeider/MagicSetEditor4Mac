//go:build darwin || linux

package app

// **Ein echter Prozess, der unordentlich endet** — F04
// (#2343).
//
// Die Abnahme verlangt den Kindprozess: er beginnt eine Sitzung **mit dem
// echten** `debug.SetCrashOutput` und endet dann mit einer Panik, durch
// SIGKILL oder sauber. Der nächste Start sieht den Abbruch — mit Spur nach
// der Panik, ohne nach dem Abschuss — und nach dem sauberen Schluss keinen.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// kindweise sagt dem Kind, wie es endet.
const kindweise = "MSE_STARTMARKE_KIND"

// TestStartmarkeKind ist das Kind. **Ohne Umgebung tut es nichts.**
func TestStartmarkeKind(t *testing.T) {
	weise, ordner := os.Getenv(kindweise), os.Getenv("MSE_STARTMARKE_ORDNER")
	if weise == "" {
		return
	}

	beginn, err := sitzungBeginnen(ordner, os.Getpid(), time.Now(), prozessLebt, setzeAbsturzausgabe)
	if err != nil {
		t.Fatal(err)
	}

	switch weise {
	case "panik":
		panic("Probepanik im Kind")
	case "warten":
		_ = os.WriteFile(filepath.Join(ordner, "bereit"), nil, 0o600) //nolint:errcheck,gosec // der Elternprozess wartet darauf

		time.Sleep(time.Minute)
	default:
		if err := beginn.ende(); err != nil {
			t.Fatal(err)
		}
	}
}

// kind startet das Kind und gibt seinen Prozess.
func kind(t *testing.T, ordner, weise string) *exec.Cmd {
	t.Helper()

	lauf := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestStartmarkeKind$") //nolint:gosec // das eigene Testprogramm
	lauf.Env = append(os.Environ(), kindweise+"="+weise, "MSE_STARTMARKE_ORDNER="+ordner)

	return lauf
}

// naechsterStart beginnt im selben Ordner eine Sitzung, wie es das nächste
// Fenster täte — mit der echten Prüfung, ob ein Prozess lebt.
func naechsterStart(t *testing.T, ordner string) []Abbruch {
	t.Helper()

	beginn, err := sitzungBeginnen(ordner, os.Getpid(), time.Now(), prozessLebt, func(*os.File) error { return nil })
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = beginn.ende() }) //nolint:errcheck // ein Testordner

	return beginn.abbrueche
}

func TestEinePanikImKindHinterlaesstMarkeUndSpur(t *testing.T) {
	ordner := t.TempDir()

	if err := kind(t, ordner, "panik").Run(); err == nil {
		t.Fatal("das Kind endete ohne Fehler — die Panik kam nicht")
	}

	abbrueche := naechsterStart(t, ordner)
	if len(abbrueche) != 1 || !abbrueche[0].MitSpur {
		t.Fatalf("Abbrüche %+v, erwartet einer mit Spur", abbrueche)
	}

	spuren, _ := filepath.Glob(filepath.Join(ordner, "absturz-*.txt")) //nolint:errcheck // ein festes Muster
	inhalt := ""
	for _, spur := range spuren {
		roh, _ := os.ReadFile(spur) //nolint:errcheck,gosec // ein Pfad aus Glob
		inhalt += string(roh)
	}

	if !strings.Contains(inhalt, "Probepanik im Kind") {
		t.Errorf("die Spur nennt die Panik nicht:\n%s", inhalt)
	}
}

func TestEinAbschussImKindHinterlaesstDieMarkeOhneSpur(t *testing.T) {
	ordner := t.TempDir()
	lauf := kind(t, ordner, "warten")

	if err := lauf.Start(); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 200; i++ {
		if _, err := os.Stat(filepath.Join(ordner, "bereit")); err == nil {
			break
		}

		time.Sleep(25 * time.Millisecond)
	}

	if err := lauf.Process.Signal(syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}

	_ = lauf.Wait() //nolint:errcheck // es ist abgeschossen

	abbrueche := naechsterStart(t, ordner)
	if len(abbrueche) != 1 || abbrueche[0].MitSpur {
		t.Errorf("Abbrüche %+v, erwartet einer ohne Spur", abbrueche)
	}
}

func TestEinSaubererSchlussImKindIstKeinAbbruch(t *testing.T) {
	ordner := t.TempDir()

	if out, err := kind(t, ordner, "sauber").CombinedOutput(); err != nil {
		t.Fatalf("das Kind endete nicht sauber: %v\n%s", err, out)
	}

	if abbrueche := naechsterStart(t, ordner); len(abbrueche) != 0 {
		t.Errorf("nach einem sauberen Schluss: %+v", abbrueche)
	}
}
