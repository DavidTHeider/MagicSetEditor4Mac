package werkzeuge_test

// **Eine laufende Partie wechselt ihre Regeln nicht** — PLAY-02.AK-03
// (#1710):
// „Änderungen während einer laufenden Mehrspielerpartie verlangen den
// vorgesehenen gemeinsamen Zustimmungsweg."
//
// # Was dieser Wächter hält
//
// **Dass es keinen zweiten Weg gibt.** Der sicherste Zustimmungsweg ist
// einer, an dem niemand vorbeikommt — und hier kommt niemand vorbei, weil
// das Regelwerk und das Profil **nur beim Aufbau** gesetzt werden. Wer
// anders spielen will, fängt eine neue Partie an.
//
// # Was er nicht hält
//
// **Dass der Aufbau das Richtige tut** — das misst
// `internal/service/regelnimaufbau_test.go`. Und er sieht keine Zuweisung,
// die über eine Zwischenvariable geht: `alt := partie; alt.Regelwerk = …`
// fiele ihm nicht auf. Dagegen steht, dass eine solche Zeile in einem Review
// auffiele — und dass es für sie keinen Grund gäbe.

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// zuweisung trifft `…Regelwerk = …` und `…Profil = …`.
//
// **Ohne `==`**: ein Vergleich ist keine Zuweisung, und ein Wächter, der ihn
// anstriche, wäre nach dem dritten Fund abgeschaltet.
var zuweisung = regexp.MustCompile(`\.(Regelwerk|Profil)\s*(=[^=]|\+=)`)

// erlaubteStellen sind die, an denen es geschehen **soll**.
var erlaubteStellen = map[string]string{
	"internal/service/partieaufbau.go": "der Aufbau — hier entsteht die " +
		"Vereinbarung, und hier friert sie ein",
}

// **Niemand setzt Regelwerk oder Profil an einer laufenden Partie.**
func TestDieRegelnEinerPartieStehenNurBeimAufbau(t *testing.T) {
	t.Parallel()

	wurzeln := []string{"../app", "../service", "../domain"}
	befunde := make([]string, 0, len(wurzeln))

	for _, wurzel := range wurzeln {
		befunde = append(befunde, zuweisungenUnter(t, wurzel)...)
	}

	if len(befunde) > 0 {
		t.Errorf("diese Stellen setzen Regelwerk oder Profil an einer "+
			"Partie, und nur der Aufbau darf das (PLAY-02.AK-03):\n%s",
			strings.Join(befunde, "\n"))
	}
}

// zuweisungenUnter sammelt die Fundstellen eines Verzeichnisses.
func zuweisungenUnter(t *testing.T, wurzel string) []string {
	t.Helper()

	befunde := []string{}

	eintraege, fehler := os.ReadDir(wurzel)
	if fehler != nil {
		t.Fatalf("%s lesen: %v", wurzel, fehler)
	}

	for _, eintrag := range eintraege {
		name := eintrag.Name()

		if eintrag.IsDir() || !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") {
			continue
		}

		pfad := filepath.Join(wurzel, name)

		if _, erlaubt := erlaubteStellen[kurzpfad(pfad)]; erlaubt {
			continue
		}

		roh, fehler := os.ReadFile(pfad) //nolint:gosec // feste Pfade im Baum
		if fehler != nil {
			t.Fatalf("%s lesen: %v", pfad, fehler)
		}

		for nummer, zeile := range strings.Split(string(roh), "\n") {
			if zuweisung.MatchString(zeile) {
				befunde = append(befunde, kurzpfad(pfad)+":"+
					strconv.Itoa(nummer+1)+": "+strings.TrimSpace(zeile))
			}
		}
	}

	return befunde
}

// kurzpfad macht aus `../service/x.go` den Pfad vom Wurzelverzeichnis aus.
func kurzpfad(pfad string) string {
	return "internal/" + strings.TrimPrefix(filepath.ToSlash(pfad), "../")
}
