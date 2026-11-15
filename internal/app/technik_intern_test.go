// **Die Teile des technischen Bereichs, einzeln** —
// #1164.
//
// # Warum sie einzeln geprüft werden
//
// Weil die Antwort auf diesem Rechner nicht jeden Fall hergibt: das
// Testprogramm liegt in einem Pfad **ohne** Leerzeichen, und neben ihm liegt
// kein `mse`. Beides ist im Bündel anders — und genau dort zählt es.
package app

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// **Ein Pfad mit Leerzeichen kommt in Anführungszeichen, einer ohne nicht.**
//
// Unter macOS ist das der Alltag: `/Applications/…` geht ohne,
// `~/Library/Application Support/…` nicht.
//
// **Zitiert wird für eine Shell und nicht für Go.** Der erste Anlauf nahm
// `strconv.Quote`; das gibt ein Go-Literal in **doppelten** Anführungszeichen,
// und darin setzt eine Shell `$HOME` ein und führt Rückwärtszitate aus. Der
// Lauf schrieb die falsche Zusage sogar fest. Angestrichen vom Zweitblick
// durch Codex über Phase 3.3, der es mit `HOME=/tmp/EXPANDIERT` nachgemessen
// hat.
func TestEinPfadWirdNurBeiBedarfZitiert(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		pfad     string
		erwartet string
	}{
		"ohne Leerzeichen": {"/Applications/MSE.app/mse", "/Applications/MSE.app/mse"},
		"mit Leerzeichen": {
			"/Users/x/Application Support/mse",
			`'/Users/x/Application Support/mse'`,
		},
		"mit Anführungszeichen": {`/x/me"in/mse`, `'/x/me"in/mse'`},
		"mit Dollarzeichen":     {"/x/$HOME/mse", `'/x/$HOME/mse'`},
		"mit Rückstrich":        {`/x/a\b/mse`, `'/x/a\b/mse'`},
		"mit Hochkomma":         {`/x/me'in/mse`, `'/x/me'\''in/mse'`},
		"mit Tilde":             {"/x/~/mse", "'/x/~/mse'"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if gefunden := zitiert(fall.pfad); gefunden != fall.erwartet {
				t.Errorf("%s statt %s", gefunden, fall.erwartet)
			}
		})
	}
}

// **Ein Verzeichnis namens `mse` ist kein Werkzeug.**
//
// Der Ort wird abgefragt, nicht angenommen: wer ein Verzeichnis so nennt,
// bekommt keine Zeile, die ins Leere führt.
func TestEinVerzeichnisIstKeinWerkzeug(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	if err := os.Mkdir(filepath.Join(ort, "mse"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ort, "mse-rest"), []byte("x"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	gefunden := vorhandeneWerkzeuge(ort)

	if len(gefunden) != 1 || gefunden[0] != "mse-rest" {
		t.Errorf("gefunden: %v", gefunden)
	}
}

// **Findet sich das Programm nicht, kommt ein Fehler und keine halbe Antwort.**
//
// Unter macOS scheitert `os.Executable` nicht; der Zweig steht trotzdem da,
// weil die Signatur ihn verlangt — und er wird hier betreten statt in
// `.covignore` geschrieben (§3).
func TestOhneProgrammpfadGibtEsEinenFehler(t *testing.T) {
	t.Parallel()

	antwort := technikangabenNeben(t.Context(),
		func() (string, error) { return "", errors.New("kein Programm") }, "")

	if antwort.Fehler == nil {
		t.Fatal("der Fehler kam als Erfolg durch")
	}

	// **Leere Listen und kein `nil`** — auch auf dem Fehlerweg.
	if antwort.Vorhanden == nil || antwort.Zeilen == nil {
		t.Errorf("Vorhanden %v, Zeilen %v", antwort.Vorhanden, antwort.Zeilen)
	}
}

// **Und eine Shell wird gefragt, statt dass jemand die Regel behauptet.**
//
// # Warum das der eigentliche Beleg ist
//
// Die Fälle oben halten eine **Schreibweise** fest. Ob diese Schreibweise das
// bedeutet, was sie soll, weiß nur ein Zerteiler von Zeichenketten — und davon
// liegt einer auf jedem Rechner. Der Lauf setzt den zitierten Pfad hinter
// `printf %s` und liest, was ankommt.
//
// **`HOME` wird dabei umgebogen.** Trüge der Pfad ein `$HOME`, das doch
// eingesetzt wird, käme sonst zufällig etwas Ähnliches heraus.
//
// Was er **nicht** prüft: eine andere Shell als die hier vorhandenen. Geprüft
// werden `/bin/sh` und, wenn da, `/bin/zsh` — die Anmeldeshell unter macOS.
func TestEineShellLiestDenZitiertenPfadAlsDenPfad(t *testing.T) {
	t.Parallel()

	pfade := []string{
		"/Applications/MSE.app/mse",
		"/Users/x/Application Support/mse",
		`/x/me"in/mse`,
		"/x/$HOME/mse",
		`/x/a\b/mse`,
		`/x/me'in/mse`,
		"/x/~/mse",
		"/x/a`echo nein`/mse",
		"/x/a;b/mse",
		"/x/a*b/mse",
	}

	for _, shell := range []string{"/bin/sh", "/bin/zsh"} {
		if _, err := os.Stat(shell); err != nil {
			t.Logf("%s gibt es hier nicht", shell)

			continue
		}

		t.Run(filepath.Base(shell), func(t *testing.T) {
			t.Parallel()

			for _, pfad := range pfade {
				befehl := exec.CommandContext(probekontext(t), shell, "-c", // #nosec G204 -- feste Fälle
					"printf %s "+zitiert(pfad))
				befehl.Env = []string{"HOME=/tmp/EXPANDIERT", "PATH=/usr/bin:/bin"}

				gelesen, err := befehl.Output()
				if err != nil {
					t.Errorf("%q: %v", pfad, err)

					continue
				}

				if string(gelesen) != pfad {
					t.Errorf("%s macht aus %q ein %q", shell, pfad, gelesen)
				}
			}
		})
	}
}
