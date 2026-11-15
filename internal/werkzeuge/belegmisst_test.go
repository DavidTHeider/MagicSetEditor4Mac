package werkzeuge_test

// **Was ein Messlauf aus dem Baum liest, steht in `misst` seines Plans** —
// QA-01.AK-04 (#1733),
// angestrichen vom Zweitblick über Phase 6.0: der gezogene Stichprobenplan
// `docs/belege/kanonstichprobe.json` stand dort nicht. Ein neu gezogener Plan
// hätte einen Beleg »aktuell« gelassen, der an einer anderen Kohorte gemessen
// war — und dasselbe galt für jede Vorrichtung unter `testdaten/`.
//
// # Was er hält
//
// **Jeder Pfad in den Baum, der in einer Datei hinter dem Tag des Plans als
// Literal dasteht**, passt auf ein Muster aus `misst` — mit den Regeln von
// `git`, denn `git ls-files` rechnet daraus den Inhaltshash: `*` trifft auch
// über `/` hinweg. Gelesen werden `filepath.Join` aus lauter festen Teilen,
// die mit `..` beginnen, und Zeichenketten, die mit `docs/`, `testdaten/`
// oder `saat/` anfangen. **Ein Verzeichnis zählt mit jeder Datei darin.**
//
// **Und jede Datei, die ein `//go:embed` in den Bau nimmt** — die Wanderungen
// des Bestands etwa. Sie ist kein Go und stand deshalb nicht unter `*.go`;
// eine geänderte Wanderung hätte einen Beleg »aktuell« gelassen, der am
// alten Schema gemessen war (gefunden beim Schreiben des Auftrags an den
// Nachlauf).
//
// # Was er nicht hält
//
//   - **Einen Pfad aus einer Veränderlichen** oder aus einem Helfer in einer
//     Datei ohne den Tag. Deshalb steht `testdaten/*` im Plan als Ganzes: die
//     Helfer der gewöhnlichen Tests lesen dort, und ein gemessener Lauf darf
//     sie rufen.
//   - **Ob ein Muster zu weit ist.** Ein zu weites Muster macht einen Beleg
//     öfter »anders« als nötig — lästig, aber kein falsches »aktuell«.

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// planMisst ist, was ein Plan über sich sagt — sein Tag und was er misst.
type planMisst struct {
	Tag   string   `json:"tag"`
	Misst []string `json:"misst"`
}

// joinLiteral trifft ein `filepath.Join` aus lauter festen Teilen.
var joinLiteral = regexp.MustCompile(`filepath\.Join\(((?:"[^"]*"\s*,\s*)*"[^"]*")\)`)

// teilLiteral trifft einen Teil darin.
var teilLiteral = regexp.MustCompile(`"([^"]*)"`)

// baumLiteral trifft einen Pfad, der von der Wurzel aus dasteht.
var baumLiteral = regexp.MustCompile(`"((?:docs|testdaten|saat)/[^"\s]+)"`)

// **Jeder Plan unter docs/belege: jeder Pfad, den seine Läufe lesen, steht
// in `misst`.**
func TestWasEinMesslaufAusDemBaumLiestStehtImPlan(t *testing.T) {
	t.Parallel()

	wurzel := filepath.Join("..", "..")

	plaene, fehler := filepath.Glob(filepath.Join(wurzel, "docs", "belege", "*.plan.json"))
	if fehler != nil || len(plaene) == 0 {
		t.Fatalf("keine Pläne unter docs/belege: %v", fehler)
	}

	for _, planpfad := range plaene {
		roh, fehler := os.ReadFile(planpfad) //nolint:gosec // ein Plan aus dem Baum
		if fehler != nil {
			t.Fatal(fehler)
		}

		var plan planMisst
		if fehler := json.Unmarshal(roh, &plan); fehler != nil {
			t.Fatalf("%s: %v", planpfad, fehler)
		}

		gelesen := gelesenHinterTag(t, wurzel, plan.Tag)

		for pfad, woher := range eingebettet(t, wurzel) {
			gelesen[pfad] = woher
		}

		for pfad, woher := range gelesen {
			if !gemessen(plan.Misst, pfad) {
				t.Errorf("%s liest %s — `misst` in %s nennt es nicht, und ein Beleg "+
					"bliebe »aktuell«, wenn es sich ändert", woher, pfad, filepath.Base(planpfad))
			}
		}
	}
}

// gelesenHinterTag gibt jeden Pfad im Baum, den eine Testdatei hinter `tag`
// als Literal nennt — relativ zur Wurzel, Verzeichnisse aufgelöst in ihre
// Dateien —, mit der Datei, die ihn nennt.
func gelesenHinterTag(t *testing.T, wurzel, tag string) map[string]string {
	t.Helper()

	gelesen := map[string]string{}
	dahinter := 0

	for _, teil := range []string{"internal", "cmd"} {
		fehler := filepath.WalkDir(filepath.Join(wurzel, teil), func(pfad string, eintrag fs.DirEntry, fehler error) error {
			if fehler != nil || eintrag.IsDir() || !strings.HasSuffix(pfad, "_test.go") {
				return fehler
			}

			roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // Testdateien im Baum
			if lesefehler != nil {
				return lesefehler
			}

			quelle := string(roh)
			if !hinterTag(quelle, tag) {
				return nil
			}

			dahinter++

			woher, _ := filepath.Rel(wurzel, pfad) //nolint:errcheck // pfad liegt unter der Wurzel

			for _, genannt := range genanntIn(quelle, filepath.Dir(woher)) {
				for _, datei := range dateienUnter(t, wurzel, genannt) {
					gelesen[datei] = woher
				}
			}

			return nil
		})
		if fehler != nil {
			t.Fatalf("%s durchgehen: %v", teil, fehler)
		}
	}

	// **Keine Datei hinter dem Tag heißt: der Leser greift daneben.** Dass
	// eine Datei nichts aus dem Baum liest, ist dagegen eine Auskunft — ein
	// Driftlauf gegen einen fremden Dienst liest das Netz (F04, #2246).
	if dahinter == 0 {
		t.Fatalf("keine Testdatei steht hinter %q — der Leser greift daneben", tag)
	}

	return gelesen
}

// einbettung trifft eine Anweisung `//go:embed` mit ihren Mustern.
var einbettung = regexp.MustCompile(`(?m)^//go:embed (.+)$`)

// eingebettet gibt jede Datei, die ein `//go:embed` unter `internal` oder
// `cmd` in den Bau nimmt — relativ zur Wurzel, mit der Datei, die sie nimmt.
func eingebettet(t *testing.T, wurzel string) map[string]string {
	t.Helper()

	gefunden := map[string]string{}

	for _, teil := range []string{"internal", "cmd"} {
		fehler := filepath.WalkDir(filepath.Join(wurzel, teil), func(pfad string, eintrag fs.DirEntry, fehler error) error {
			if fehler != nil || eintrag.IsDir() || !strings.HasSuffix(pfad, ".go") {
				return fehler
			}

			roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // Quelldateien im Baum
			if lesefehler != nil {
				return lesefehler
			}

			woher, _ := filepath.Rel(wurzel, pfad) //nolint:errcheck // pfad liegt unter der Wurzel

			for _, treffer := range einbettung.FindAllStringSubmatch(string(roh), -1) {
				for _, muster := range strings.Fields(treffer[1]) {
					muster = strings.TrimPrefix(strings.Trim(muster, "\"`"), "all:")

					dateien, _ := filepath.Glob(filepath.Join(filepath.Dir(pfad), muster)) //nolint:errcheck // ein kaputtes Muster lehnt schon der Übersetzer ab
					for _, datei := range dateien {
						relativ, _ := filepath.Rel(wurzel, datei) //nolint:errcheck // datei liegt unter der Wurzel
						for _, einzeln := range dateienUnter(t, wurzel, filepath.ToSlash(relativ)) {
							gefunden[einzeln] = filepath.ToSlash(woher)
						}
					}
				}
			}

			return nil
		})
		if fehler != nil {
			t.Fatalf("%s durchgehen: %v", teil, fehler)
		}
	}

	if len(gefunden) == 0 {
		t.Fatal("kein //go:embed gefunden — der Leser greift daneben")
	}

	return gefunden
}

// genanntIn gibt die Pfade im Baum, die eine Quelle als Literal nennt —
// relativ zur Wurzel; `ort` ist das Verzeichnis der Quelle.
func genanntIn(quelle, ort string) []string {
	var pfade []string

	for _, treffer := range joinLiteral.FindAllStringSubmatch(quelle, -1) {
		var teile []string
		for _, teil := range teilLiteral.FindAllStringSubmatch(treffer[1], -1) {
			teile = append(teile, teil[1])
		}

		if teile[0] != ".." {
			continue
		}

		pfad := filepath.ToSlash(filepath.Clean(filepath.Join(append([]string{ort}, teile...)...)))
		if !strings.HasPrefix(pfad, "../") {
			pfade = append(pfade, pfad)
		}
	}

	for _, treffer := range baumLiteral.FindAllStringSubmatch(quelle, -1) {
		pfade = append(pfade, treffer[1])
	}

	return pfade
}

// dateienUnter gibt einen Pfad als Liste seiner Dateien — ein Verzeichnis
// mit allem, was darin liegt, und Verborgenes nicht.
func dateienUnter(t *testing.T, wurzel, pfad string) []string {
	t.Helper()

	info, fehler := os.Stat(filepath.Join(wurzel, pfad))
	if fehler != nil || !info.IsDir() {
		// **Auch ein Pfad, den es nicht gibt, zählt** — ein Lauf, der ihn
		// erwartet, misst ihn, sobald er daliegt.
		return []string{pfad}
	}

	var dateien []string

	_ = filepath.WalkDir(filepath.Join(wurzel, pfad), func(voll string, eintrag fs.DirEntry, fehler error) error { //nolint:errcheck // ein unlesbarer Zweig fehlt dann in der Liste und fällt dem Muster nicht zur Last
		if fehler != nil || strings.HasPrefix(eintrag.Name(), ".") && voll != filepath.Join(wurzel, pfad) {
			if eintrag != nil && eintrag.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		if !eintrag.IsDir() {
			relativ, _ := filepath.Rel(wurzel, voll) //nolint:errcheck // voll liegt unter der Wurzel
			dateien = append(dateien, filepath.ToSlash(relativ))
		}

		return nil
	})

	return dateien
}

// gemessen sagt, ob ein Muster aus `misst` den Pfad trifft — **mit den Regeln
// einer Pfadangabe von git**: ohne Platzhalter der Pfad selbst oder alles
// darunter, mit Platzhaltern `fnmatch` ohne `FNM_PATHNAME`, `*` also auch
// über `/` hinweg.
func gemessen(muster []string, pfad string) bool {
	for _, einer := range muster {
		if !strings.ContainsAny(einer, "*?[") {
			if pfad == einer || strings.HasPrefix(pfad, strings.TrimSuffix(einer, "/")+"/") {
				return true
			}

			continue
		}

		ausdruck := strings.NewReplacer(`\*`, `.*`, `\?`, `.`).Replace(regexp.QuoteMeta(einer))
		if regexp.MustCompile("^" + ausdruck + "$").MatchString(pfad) {
			return true
		}
	}

	return false
}

// **Das Muster folgt git** — sonst hielte der Wächter etwas anderes, als der
// Inhaltshash rechnet.
func TestDasMusterFolgtGit(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		muster []string
		pfad   string
		trifft bool
	}{
		"ein Stern über Verzeichnisse hinweg": {[]string{"*.go"}, "internal/app/a.go", true},
		"ein Stern trifft nicht die Endung":   {[]string{"*.go"}, "internal/app/a.json", false},
		"ein fester Pfad":                     {[]string{"docs/PARITAET.md"}, "docs/PARITAET.md", true},
		"ein Verzeichnis ohne Platzhalter":    {[]string{"testdaten"}, "testdaten/probe.mse-set/set", true},
		"kein Vorsatz eines anderen Namens":   {[]string{"docs/PARITAET"}, "docs/PARITAET.md", false},
		"testdata tief im Baum":               {[]string{"*/testdata/*"}, "internal/app/testdata/a/b.png", true},
		"ein Fragezeichen für ein Zeichen":    {[]string{"docs/?.md"}, "docs/a.md", true},
	}

	for name, einer := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if trifft := gemessen(einer.muster, einer.pfad); trifft != einer.trifft {
				t.Errorf("%v gegen %s: %v, erwartet %v", einer.muster, einer.pfad, trifft, einer.trifft)
			}
		})
	}
}
