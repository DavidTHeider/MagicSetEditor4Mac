//go:build unix

package app

// **Die Ansage eines Dienstes, der gleich wieder endet** — gefunden am
// 2026-09-24, als `TestEinSterbenderDienstHinterlaesstEinenBefund` in der
// Pipeline unter Last rot wurde: „der Dienst hat sich nicht gemeldet", obwohl
// das Programm beide Zeilen geschrieben hatte.
//
// **Die Ursache stand in `os/exec`**: `Cmd.Wait` schließt die Ausgabe, sobald
// das Kind endet, und es sei deshalb falsch, `Wait` vor dem letzten Lesen zu
// rufen (Go 1.27.1, `StdoutPipe`). `starteDienstprogrammUeber` rief es gleich
// nach dem Start. Endete der Dienst schneller, als gelesen wurde, war die
// Ansage fort.
//
// **Hier nur für Unix**, weil der Leser über `kill(pid, 0)` erkennt, ob das
// Kind schon abgeholt ist — und weil die Dienste als Shellskripte laufen.

import (
	"io"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieAnsageEinesSchnellEndendenDienstesGehtNichtVerloren(t *testing.T) {
	ort := t.TempDir()

	einProgramm(t, ort, "mse-rest", `
echo "Dienst auf http://127.0.0.1:8123"
echo "Marke geheim"
exit 0
`)

	// **Der Leser kommt zu spät, mit Absicht**: erst, wenn das Kind abgeholt
	// ist. Wer es gleich nach dem Start abholen ließ, hat die Ausgabe dann
	// schon geschlossen — und die Ansage ist fort.
	naehte := imOrt(ort)
	echte := naehte.rohre
	naehte.rohre = func(befehl *exec.Cmd) (io.WriteCloser, io.ReadCloser, error) {
		eingabe, ausgabe, fehler := echte(befehl)

		return eingabe, &spaeterLeser{ReadCloser: ausgabe, befehl: befehl}, fehler
	}

	lauf, fehler := starteDienstprogrammUeber(t.Context(), naehte,
		domain.DienstREST, domain.Bindung{Adresse: domain.Schleifenadresse})
	if fehler != nil {
		t.Fatalf("die Ansage ging verloren: %v", fehler)
	}

	<-lauf.fertig

	if lauf.adresse != "http://127.0.0.1:8123" || lauf.marke != "geheim" {
		t.Errorf("gelesen %q und %q", lauf.adresse, lauf.marke)
	}
}

// spaeterLeser liest zum ersten Mal, wenn das Kind abgeholt ist — oder nach
// einer Sekunde, wenn niemand es abholt.
type spaeterLeser struct {
	io.ReadCloser

	befehl *exec.Cmd
	einmal sync.Once
}

func (l *spaeterLeser) Read(puffer []byte) (int, error) {
	l.einmal.Do(func() {
		ende := time.Now().Add(time.Second)

		// **Abgeholt heißt: `kill(pid, 0)` kennt es nicht mehr.** Ein beendetes,
		// nicht abgeholtes Kind steht als Zombie noch da und antwortet.
		for time.Now().Before(ende) {
			if syscall.Kill(l.befehl.Process.Pid, 0) != nil {
				// `Wait` schließt die Ausgabe gleich nach dem Abholen.
				time.Sleep(50 * time.Millisecond)

				return
			}

			time.Sleep(5 * time.Millisecond)
		}
	})

	return l.ReadCloser.Read(puffer) //nolint:wrapcheck // durchgereicht
}
