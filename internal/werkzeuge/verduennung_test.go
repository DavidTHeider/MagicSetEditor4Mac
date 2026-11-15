// Der Wächter gegen **verdünnte** Vordergrundfarben.
//
// # Warum es ihn neben `kontraste_test.go` gibt
//
// Jener hält jede Vordergrundfarbe gegen `docs/KONTRASTE.md` — und sein Muster
// endet an der Farbe. `border-kante/60` trifft `border-kante`, findet dort
// seine Zeile und ist grün; gezeichnet wird aber eine Farbe, die in keiner
// Tabelle steht.
//
// **Gemessen am 2026-09-11** bei der Sichtprüfung zu
// #1500: die
// Zeilen der Fensterliste trugen `border-kante/60`, und der Kontrast der
// Trennlinie war im hellen Erscheinungsbild **2,08** — die Schranke aus §3
// ist 3. Dieselbe Farbe voll ist 3,99.
//
// # Was er nicht prüft
//
// **Flächen nicht** (`bg-`): eine verdünnte Fläche ist ein Grund, und über
// Gründe entscheidet das Paar, in dem sie vorkommt — genau die Lücke, die
// `kontraste_test.go` bei sich selbst benennt.
//
// **Und er misst nicht.** Er weiß nur, dass eine verdünnte Farbe nicht die
// gemessene ist. Ob die verdünnte reichte, sagt allein eine Messung im
// Browser.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// farbnamen liest die Namen der Gestaltungssprache aus `stil.css`.
//
// **Aus der Datei und nicht aus einer Aufzählung hier**: eine zweite Liste
// liefe auseinander, und die kürzere wäre wieder die grüne. Sie steht in
// dieser Datei, weil beide Wächter sie brauchen — `kontraste_test.go` für
// die volle Farbe, dieser für die verdünnte.
func farbnamen(t *testing.T) []string {
	t.Helper()

	treffer := regexp.MustCompile(`--color-([a-z]+):`).
		FindAllStringSubmatch(StilblattMitEinbindungen(t), -1)
	if len(treffer) == 0 {
		t.Fatal("kein --color-Merkmal in stil.css — der Wächter liefe ins Leere")
	}

	namen := make([]string, 0, len(treffer))

	for _, eintrag := range treffer {
		namen = append(namen, eintrag[1])
	}

	return namen
}

// verduennteFarben findet eine **Farbe** mit einem Schrägstrich dahinter.
//
// **Nur eine Farbe, und das ist der Punkt**: `text-sm/6` ist Tailwinds
// Kurzform für Schriftgröße und Zeilenhöhe, keine Verdünnung. Der erste Anlauf
// nahm jedes Wort und strich drei Stellen an, von denen zwei richtig waren —
// ein Wächter, der zu viel verlangt, wird bemerkt (§10).
func verduennteFarben(t *testing.T) *regexp.Regexp {
	t.Helper()

	return regexp.MustCompile(
		`(?:dark:)?(?:text|border|divide|outline|ring)-(?:` +
			strings.Join(farbnamen(t), "|") + `|[a-z]+-\d{2,3}|white|black)/\d{1,3}`)
}

func TestKeineVerduennteVordergrundfarbe(t *testing.T) {
	t.Parallel()

	muster := verduennteFarben(t)
	wurzel := filepath.Join("..", "..", "frontend", "src")

	var funde []string

	if err := filepath.Walk(wurzel, func(pfad string, info os.FileInfo, fehler error) error {
		if fehler != nil {
			return fehler
		}

		if info.IsDir() || !strings.HasSuffix(pfad, ".tsx") {
			return nil
		}

		inhalt, lesefehler := os.ReadFile(pfad) //nolint:gosec // Pfad aus dem Baum selbst
		if lesefehler != nil {
			return lesefehler
		}

		for _, treffer := range muster.FindAllString(
			ohneQuelltextkommentar(string(inhalt)), -1) {
			funde = append(funde, pfad+": "+treffer)
		}

		return nil
	}); err != nil {
		t.Fatalf("frontend/src durchgehen: %v", err)
	}

	for _, fund := range funde {
		t.Errorf("%s — eine verdünnte Vordergrundfarbe steht in keiner Zeile von "+
			"docs/KONTRASTE.md; die Tabelle nennt die volle (AGENTS.md §3)", fund)
	}
}
