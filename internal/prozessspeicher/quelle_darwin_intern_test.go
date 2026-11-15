//go:build darwin

package prozessspeicher

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// **Am echten System**: der eigene Fußabdruck ist gemessen, und die Liste
// der WebKit-Prozesse ließ sich fragen — ein Testlauf hält keine Webansicht.
func TestDerEigeneSpeicherIstGemessen(t *testing.T) {
	t.Parallel()

	stand, err := Messen(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}

	if stand.Anwendung.FussabdruckBytes < 1<<20 || stand.Ungemessen != "" || len(stand.Web) != 0 {
		t.Errorf("Stand %+v", stand)
	}
}

// **Ein Prozess, den es nicht mehr gibt**, hat weder Programm noch Speicher.
func TestEinBeendeterProzessHatNichts(t *testing.T) {
	t.Parallel()

	lauf := exec.CommandContext(t.Context(), "/usr/bin/true")
	if err := lauf.Run(); err != nil {
		t.Fatal(err)
	}

	pid := lauf.Process.Pid

	if _, err := programmpfad(pid); err == nil {
		t.Error("ein Programm zu einem beendeten Prozess")
	}

	if _, err := fussabdruckVon(pid); err == nil {
		t.Error("Speicher zu einem beendeten Prozess")
	}

	// **Und er lebt nicht** — der eigene und ein fremder schon (Zweitblick
	// über Phase 4, Befund 9): launchd gehört root, und `EPERM` heißt, es
	// gibt ihn.
	if prozessLebt(pid) || !prozessLebt(os.Getpid()) || !prozessLebt(1) {
		t.Errorf("lebt: beendet %t, eigen %t, launchd %t", prozessLebt(pid), prozessLebt(os.Getpid()), prozessLebt(1))
	}
}

// **Ein Symbol, das es nicht gibt, wird genannt** — und ein Ruf darauf
// fängt der Riegel.
func TestEinFehlendesSymbolWirdNichtGemessen(t *testing.T) {
	t.Parallel()

	var fehlt func() int32

	if err := binde(&fehlt, "gibt_es_nicht_mse4mac"); err == nil {
		t.Error("ein fehlendes Symbol wurde gebunden")
	}

	if _, err := geriegelt(fehlt); err == nil {
		t.Error("der Riegel fing nichts")
	}
}

// **Der eigene Programmpfad ist der des Testprogramms** — die Messung fragt
// den Pfad seit dem Nachlauf über Phase 4 nur noch für eigene Kinder, und
// ein Testlauf hat keine.
func TestDerEigeneProgrammpfad(t *testing.T) {
	t.Parallel()

	pfad, err := programmpfad(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}

	if eigen, _ := os.Executable(); filepath.Base(pfad) != filepath.Base(eigen) { //nolint:errcheck // ohne Pfad scheitert der Vergleich
		t.Errorf("Pfad %q, erwartet %q", pfad, eigen)
	}
}
