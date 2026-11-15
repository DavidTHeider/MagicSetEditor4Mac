package app

// **Die zwei Rohre zum Kind und was sein Tod hinterlässt** — Werkstück N6 von
// Phase 5.4
// (#1572).
//
// **Herausgelöst aus `dienstwerk_intern_test.go`**, als die Datei über die
// Zeilengrenze aus §2 wuchs. Der Schnitt folgt der Frage: dort, **dass** ein
// Programm gefahren wird; hier, **womit** es spricht und was geschieht, wenn
// es von selbst aufhört.

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ohne Rohre wird nichts gefahren.**
//
// **Der Zweig, den der Betrieb nicht betritt** — ein frischer [exec.Cmd] hat
// weder `Stdin` noch `Stdout` gesetzt, und ohne diese Naht bliebe der Weg
// daneben ungemessen (§3).
func TestOhneRohreWirdNichtsGefahren(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	einProgramm(t, ort, "mse-rest", "echo sollte nicht laufen")

	ohneRohre := imOrt(ort)
	ohneRohre.rohre = func(*exec.Cmd) (io.WriteCloser, io.ReadCloser, error) {
		return nil, nil, errors.New("kein Rohr")
	}

	_, fehler := starteDienstprogrammUeber(context.Background(), ohneRohre,
		domain.DienstREST, domain.Bindung{Adresse: domain.Schleifenadresse})
	if fehler == nil {
		t.Fatal("ohne Rohre lief etwas")
	}
}

// **Und beide Rohre scheitern aus demselben Grund.**
//
// **Gemessen an dem Grund, den es wirklich gibt**: `exec` weist ein zweites
// Rohr ab, wenn schon eines steht. Das ist kein erfundener Fall — es ist der
// einzige, den [exec.Cmd.StdinPipe] überhaupt nennt.
func TestEinBesetztesRohrWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*exec.Cmd){
		"die Eingabe steht schon": func(befehl *exec.Cmd) {
			befehl.Stdin = strings.NewReader("")
		},
		"die Ausgabe steht schon": func(befehl *exec.Cmd) {
			befehl.Stdout = io.Discard
		},
	}

	for name, besetzen := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			befehl := exec.CommandContext(
				t.Context(), "/usr/bin/true") //#nosec G204 -- ein Literal

			besetzen(befehl)

			if _, _, fehler := rohreAufmachen(
				befehl); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("ein besetztes Rohr gab %v", fehler)
			}
		})
	}

	// **Und frei geht es durch** — sonst sagte der Lauf nur, dass etwas
	// scheitert, und nicht, dass der Normalfall trägt.
	//#nosec G204 -- ein Literal
	frisch := exec.CommandContext(t.Context(), "/usr/bin/true")

	if _, _, fehler := rohreAufmachen(frisch); fehler != nil {
		t.Errorf("frische Rohre gaben %v", fehler)
	}
}

// **Ein Dienst, der von selbst stirbt, hinterlässt einen Befund.**
//
// **Ins Protokoll und nicht ans Fenster**: der Ruf, der ihn gestartet hat, ist
// längst beantwortet. Ohne diesen Lauf stünde nur, dass der Weg dorthin
// existiert.
func TestEinSterbenderDienstHinterlaesstEinenBefund(t *testing.T) {
	// **Nicht parallel**: der Lauf tauscht das voreingestellte Protokoll.
	vorher := slog.Default()

	t.Cleanup(func() { slog.SetDefault(vorher) })

	mitschrift := &gesammelt{}
	slog.SetDefault(slog.New(slog.NewTextHandler(mitschrift, nil)))

	ort := t.TempDir()

	einProgramm(t, ort, "mse-rest", `
echo "Dienst auf http://127.0.0.1:8123"
echo "Marke geheim"
exit 3
`)

	lauf, fehler := starteDienstprogrammUeber(context.Background(), imOrt(ort),
		domain.DienstREST, domain.Bindung{Adresse: domain.Schleifenadresse})
	if fehler != nil {
		t.Fatalf("starteDienstprogrammUeber: %v", fehler)
	}

	<-lauf.fertig

	if !strings.Contains(mitschrift.text(), "Dienstposten meldet einen Befund") {
		t.Errorf("das Protokoll sagt nichts: %q", mitschrift.text())
	}
}

// gesammelt hält, was das Protokoll schreibt — gegen gleichzeitiges Lesen
// gesichert, denn geschrieben wird aus der wartenden Goroutine.
type gesammelt struct {
	schranke sync.Mutex
	inhalt   strings.Builder
}

func (g *gesammelt) Write(roh []byte) (int, error) {
	g.schranke.Lock()
	defer g.schranke.Unlock()

	return g.inhalt.Write(roh) //nolint:wrapcheck // ein Puffer scheitert nicht
}

func (g *gesammelt) text() string {
	g.schranke.Lock()
	defer g.schranke.Unlock()

	return g.inhalt.String()
}
