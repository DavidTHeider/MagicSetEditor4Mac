// **Ein Freigabelauf, der aus `HOME` liest, sagt es** —
// #1170.
//
// # Der Befund
//
// `internal/app` gibt seinem ganzen Testlauf seit dem 2026-08-26 ein eigenes
// Zuhause unter `/tmp` (`echteszuhause_test.go`) — damit kein Test den Stand
// dieses Rechners sieht. Zwei Wege bauten ihren Pfad weiter aus `HOME`:
// `kanonverzeichnis` und `mse2ausgabe`.
//
// **Damit sprangen der Kanonlauf und der Abstandslauf ab.** Beide stehen auf
// der Freigabeliste in `docs/RELEASE.md`, und beide sind der einzige Vergleich
// dieses Projekts gegen einen **fremden** Maßstab. Zwei Freigaben lang wurden
// sie nicht gefahren, und niemand sah es: ein `t.Skip` endet mit `PASS`.
//
// # Was dieser Wächter hält
//
// Jeder Weg unter `internal/app`, der einen Ort **außerhalb** des Testzuhauses
// meint, fragt zuerst seine Umgebungsvariable. Geprüft wird die Gestalt: wer
// `os.Getenv("HOME")` mit `"MagicSetEditor4Mac"` verbindet, muss vorher eine
// Ausweichvariable lesen.
//
// # Was er nicht hält
//
//   - **Ob die Variable gesetzt wird.** Das steht in `docs/RELEASE.md`, und
//     eine Anweisung an einen Menschen prüft kein Test.
//   - **Ob der Lauf danach etwas findet.** Ein falscher Pfad springt weiter
//     ab. Was ihn hält, ist die Ausgabe: der Lauf nennt beim Sprung den
//     Befehl, mit dem die Ausgabe entsteht.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// zuhauseweg findet `os.Getenv("HOME")` in einem Pfadbau.
var zuhauseweg = regexp.MustCompile(`os\.Getenv\("HOME"\)`)

// TestJederKanonwegKenntSeineAusweichvariable ist die Zusage.
func TestJederKanonwegKenntSeineAusweichvariable(t *testing.T) {
	gefunden := 0

	fehler := filepath.WalkDir("../app", func(pfad string, eintrag os.DirEntry, err error) error {
		if err != nil || eintrag.IsDir() || !strings.HasSuffix(pfad, ".go") {
			return err
		}

		roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // Quelldatei im Repositorium
		if lesefehler != nil {
			return lesefehler
		}

		inhalt := string(roh)

		// **Nur die Wege in das Zuhause der Anwendung.** Ein `HOME` für einen
		// Downloadordner oder eine Prüfung des echten Zuhauses meint etwas
		// anderes, und die Variable dafür wäre eine Erfindung.
		for _, block := range funktionenMit(inhalt, "MagicSetEditor4Mac") {
			if !zuhauseweg.MatchString(block) {
				continue
			}

			gefunden++

			if !strings.Contains(block, "os.Getenv(\"MSE_") {
				t.Errorf("%s: ein Weg baut seinen Pfad aus HOME, ohne vorher eine "+
					"MSE_-Variable zu lesen — unter dem Testzuhause aus TestMain "+
					"findet er nichts, und der Lauf springt ab statt rot zu werden "+
					"(docs/RELEASE.md)", filepath.Base(pfad))
			}
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("internal/app durchgehen: %v", fehler)
	}

	if gefunden == 0 {
		t.Fatal("kein Weg ins Zuhause der Anwendung gefunden — der Leser greift daneben")
	}
}

// funktionenMit gibt die Funktionsrümpfe, die einen Text enthalten.
//
// **Über den Text und nicht über den Syntaxbaum.** Gesucht ist ein Muster in
// zwei Zeilen, die zusammengehören; ein Baum gäbe hier dieselbe Antwort und
// dreißig Zeilen Weg dorthin.
func funktionenMit(inhalt, gesucht string) []string {
	var rumpfe []string

	for _, stueck := range strings.Split(inhalt, "\nfunc ")[1:] {
		ende := strings.Index(stueck, "\n}\n")
		if ende < 0 {
			ende = len(stueck)
		}

		if rumpf := stueck[:ende]; strings.Contains(rumpf, gesucht) {
			rumpfe = append(rumpfe, rumpf)
		}
	}

	return rumpfe
}
