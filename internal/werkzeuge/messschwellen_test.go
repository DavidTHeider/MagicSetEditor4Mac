// **Jeder gemessene Weg hat eine Schwelle** —
// #891,
// Werkstück 6 von Phase Y.
//
// # Warum es diesen Wächter braucht
//
// Ein Benchmark, der nur berichtet, fängt keinen Rückschritt: seine Ausgabe
// liest niemand zweimal, und die Zahl von vorhin steht nirgends. Erst eine
// Schwelle macht aus der Messung einen Wächter.
//
// **Der nächste Weg kommt ohne sie.** Wer einen vierten Benchmark schreibt,
// hat die Schwelle nicht im Kopf — und ein Benchmark ohne sie sieht aus wie
// die drei daneben. Dieser Wächter macht den Unterschied sichtbar.
//
// # Warum hier und nicht dort
//
// Die Benchmarks stehen hinter `fremdpakete` (§3) und laufen nie in `make ci`.
// Ein Wächter neben ihnen liefe damit auch nie. Dieser liest ihren
// **Quelltext** und läuft in jedem Lauf.
//
// # Was er nicht prüft
//
//   - **Ob die Schwelle die richtige ist.** Er sieht, dass eine dasteht, nicht
//     ob sie einen Rückschritt fängt. Dafür gibt es die Gegenprobe, und sie
//     steht bei `schwelleListe`: `Karten` statt `Kartenzeilen` braucht 35 ms
//     und muss rot werden.
//   - **Ob der Weg spürbar ist.** Welche drei Wege jemand wartet, ist eine
//     fachliche Aussage und steht im Kopf der Datei.
//   - **Ob jemand die Benchmarks fährt.** Das steht in `docs/RELEASE.md` und
//     ist Handarbeit — wie jeder Lauf hinter `fremdpakete`.
package werkzeuge_test

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// schwellenruf findet den Aufruf, der eine Schwelle hält.
//
// **Zwei Namen, weil es zwei Arten von Messung gibt**: `haelt` teilt
// `b.Elapsed()` durch die Zahl der Durchläufe, `haeltZeit` bekommt eine selbst
// gestoppte Dauer — für einen Weg, den es nur **einmal** gibt (den Kaltstart).
var schwellenruf = regexp.MustCompile(`\bhaelt(Zeit)?\(b,`)

// benchmarkkopf findet `func BenchmarkWeg…(b *testing.B) {`.
var benchmarkkopf = regexp.MustCompile(`(?m)^func (BenchmarkWeg\w*)\(b \*testing\.B\) \{`)

// messwegemuster findet die Dateien mit den gemessenen Wegen.
//
// **Gesucht und nicht aufgezählt.** Hier stand erst **eine** Datei, dann eine
// Liste von dreien — und beides hatte dasselbe Loch: ein Benchmark in einer
// neuen `*_bench_intern_test.go` ginge vorbei, und die Zusage „jeder gemessene
// Weg hat eine Schwelle" hinge daran, dass jemand an diese Liste denkt. Genau
// das ist am 2026-08-28 **zweimal** passiert, beide Male in
// #1270:
// einmal mit der Zerlegung, einmal mit dem Kaltstart.
//
// Beim zweiten Mal war die Aufzählung schon die Antwort auf das erste Mal. Das
// ist der Fall, den §10 meint: die Korrektur behandelte den **genannten** Fall
// und ließ die Klasse stehen.
// **Gelaufen und nicht geglobbt**: `filepath.Glob` bräuchte eine feste Tiefe,
// und `internal/a/b/` fiele durch.
const messwegeendung = "_bench_intern_test.go"

func TestJederGemesseneWegHatEineSchwelle(t *testing.T) {
	dateien := messwegedateien(t)

	// **Und es gibt überhaupt welche.** Eine Suche, die nichts findet, machte
	// diesen Wächter still grün — dieselbe Gestalt wie die Aufzählung, die
	// nichts nennt.
	if len(dateien) == 0 {
		t.Fatalf("keine einzige Datei auf %q — sucht der Wächter noch richtig?",
			messwegeendung)
	}

	gezaehlt := 0

	for _, datei := range dateien {
		gezaehlt += schwellenIn(t, datei)
	}

	t.Logf("%d gemessene Wege in %d Dateien, jeder mit Schwelle",
		gezaehlt, len(dateien))

	// **Und `docs/RELEASE.md` nennt sie**, sonst fährt sie niemand: sie laufen
	// nicht in `make ci` (§4), und was nur auf dieser Liste steht, geschieht.
	if !strings.Contains(lies(t, "docs/RELEASE.md"), "-bench Weg") {
		t.Error("docs/RELEASE.md nennt die Messwege nicht — sie laufen nicht " +
			"in der Pipeline, also fährt sie sonst niemand")
	}
}

// messwegedateien sammelt jede Messdatei unter `internal/`.
//
// **Die Namen kommen relativ zum Projektstamm zurück**, weil [lies] sie so
// erwartet.
func messwegedateien(t *testing.T) []string {
	t.Helper()

	stamm := filepath.Join("..", "..")

	var gefunden []string

	if err := filepath.WalkDir(filepath.Join(stamm, "internal"),
		func(pfad string, eintrag fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if eintrag.IsDir() || !strings.HasSuffix(eintrag.Name(), messwegeendung) {
				return nil
			}

			kurz, fehler := filepath.Rel(stamm, pfad)
			if fehler != nil {
				return fehler
			}

			gefunden = append(gefunden, kurz)

			return nil
		}); err != nil {
		t.Fatalf("Messwegedateien suchen: %v", err)
	}

	return gefunden
}

// schwellenIn prüft eine Datei und gibt die Zahl ihrer Wege.
func schwellenIn(t *testing.T, datei string) int {
	t.Helper()

	quelle := lies(t, datei)

	wege := benchmarkkopf.FindAllStringSubmatch(quelle, -1)
	if len(wege) == 0 {
		t.Errorf("%s nennt keinen einzigen Benchmark — heißt er nicht mehr "+
			"`BenchmarkWeg…`, oder ist die Datei weg?", datei)

		return 0
	}

	// **Jeder Weg ruft `haelt`.** Der Rumpf reicht von seinem Kopf bis zum
	// nächsten — oder bis zum Ende, beim letzten.
	stellen := benchmarkkopf.FindAllStringIndex(quelle, -1)

	for i, weg := range wege {
		ende := len(quelle)
		if i+1 < len(stellen) {
			ende = stellen[i+1][0]
		}

		if !schwellenruf.MatchString(quelle[stellen[i][0]:ende]) {
			t.Errorf("%s misst, hält aber keine Schwelle — ein Benchmark ohne "+
				"sie berichtet nur, und einen Rückschritt fängt er nicht",
				weg[1])
		}
	}

	return len(wege)
}
