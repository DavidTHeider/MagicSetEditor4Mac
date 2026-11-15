// **Zehn Starts, eine Marke** — SEC02
// (#2355),
// Abnahme „je Start Ansagedatei ersetzen mit 0600 im 0700-Ordner; keine alten
// Marken nach zehn Neustarts".
//
// # Was dieser Lauf belegt — und was nicht
//
// **Was das Programm tut**: dasselbe gebaute `mse-rest` mit `-ansage`, zehnmal
// gestartet und beendet, so wie launchd es nach einem Fehler neu anwirft. Bis
// SEC02 schrieb der Dienst seine Ansage nur auf die Standardausgabe, und
// launchd hängte sie an `dienst.txt` an — nach zehn Starts standen dort zehn
// Marken.
//
// **Was launchd tut**, belegt dieser Lauf nicht: dass `StandardOutPath` ins
// Leere geht und die Fehlerausgabe in `dienst.log`, zeigt der Gastlauf
// `scripts/sec02gast.mjs` mit echten Neustarts unter launchd.
package flaechen_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestZehnStartsLassenEineMarke(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	ordner := filepath.Join(t.TempDir(), "schnittstellen")
	datei := filepath.Join(ordner, "dienst.txt")

	gesehen := make([]string, 0, 10)

	for start := range 10 {
		marke := einStartMitAnsage(t, umgebung, datei)

		roh, fehler := os.ReadFile(datei) // #nosec G304 -- Pfad unter t.TempDir
		if fehler != nil {
			t.Fatalf("Start %d: die Ansage lesen: %v", start+1, fehler)
		}

		inhalt := string(roh)

		if strings.Count(inhalt, "Marke ") != 1 || !strings.Contains(inhalt, "Marke "+marke+"\n") {
			t.Errorf("Start %d: in der Datei steht nicht genau die angesagte Marke:\n%s", start+1, inhalt)
		}

		for stelle, alte := range gesehen {
			if alte == marke {
				t.Errorf("Start %d: dieselbe Marke wie Start %d", start+1, stelle+1)
			} else if strings.Contains(inhalt, alte) {
				t.Errorf("Start %d: die Marke von Start %d steht noch darin", start+1, stelle+1)
			}
		}

		gesehen = append(gesehen, marke)
	}

	// **Kein Fremdleser**: die Datei nur für ihren Eigentümer, der Ordner
	// auch — sonst sähe ein zweiter Mensch am Rechner die Marke.
	for pfad, rechte := range map[string]os.FileMode{datei: 0o600, ordner: 0o700} {
		stand, fehler := os.Stat(pfad)
		if fehler != nil {
			t.Fatalf("%s: %v", pfad, fehler)
		}

		if stand.Mode().Perm() != rechte {
			t.Errorf("%s hat %v, erwartet %v", pfad, stand.Mode().Perm(), rechte)
		}
	}
}

// einStartMitAnsage startet `mse-rest -ansage`, liest die Marke von der
// Standardausgabe und beendet den Prozess wieder.
//
// **Gelesen wird die Datei erst nach der Ansage**: das Programm schreibt sie
// vorher, und wer die Marke auf der Ausgabe sieht, findet sie in der Datei.
func einStartMitAnsage(t *testing.T, umgebung []string, datei string) string {
	t.Helper()

	ctx, beenden := context.WithCancel(t.Context())

	// #nosec G204 -- gebautes Programm aus TestMain
	lauf := exec.CommandContext(ctx, gebaut.rest, "-port", "0", "-ansage", datei)
	lauf.Env = umgebung

	angesagt := &mitschrift{}
	lauf.Stdout, lauf.Stderr = angesagt, angesagt

	if fehler := lauf.Start(); fehler != nil {
		beenden()
		t.Fatalf("den Dienst starten: %v", fehler)
	}

	geendet := make(chan struct{})

	go func() {
		_ = lauf.Wait() //nolint:errcheck,gosec // beendet wird über den Kontext
		close(geendet)
	}()

	defer func() {
		beenden()
		<-geendet
	}()

	_, marke, _ := ansageLesen(t, angesagt, geendet)

	return marke
}
