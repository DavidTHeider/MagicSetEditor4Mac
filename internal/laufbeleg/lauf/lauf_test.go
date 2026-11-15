package main

// **Das Programm, das misst und belegt** — QA-01
// (#1733).
//
// Die Läufe hier fahren es gegen einen **eigenen Arbeitsbaum** und ein
// **Double** der Werkzeuge: `git` antwortet aus einer Tabelle, `go test`
// schreibt einen vorgegebenen Strom. Dass die echten Werkzeuge dieselbe Naht
// bedienen, prüft `kasten_test.go`.

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/laufbeleg"
)

// fest ist die Uhrzeit jedes Laufs hier.
var fest = time.Date(2026, 9, 22, 23, 0, 0, 0, time.UTC)

// commit ist der Stand, an dem das Double misst.
const commit = "abc1234def5678"

// status ist die Frage nach dem Baum — **ohne den eigenen Beleg**.
const status = "status --porcelain -- . :(exclude)docs/belege/fremdpakete.json"

// inhalt ist die Frage nach dem, was der Lauf misst — die Pfade des Plans und
// der Plan selbst.
const inhalt = "ls-files -s -- *.go go.mod docs/belege/fremdpakete.plan.json"

// offen ist die Frage nach dem Arbeitsbaum dort, wo der Lauf misst — dieselben
// Pfade wie [inhalt].
const offen = "status --porcelain -- *.go go.mod docs/belege/fremdpakete.plan.json"

// double steht für die Werkzeuge.
type double struct {
	git         map[string]string
	gitFehler   map[string]error
	strom       string
	rueckgabe   int
	startfehler error
	daneben     string
	gerufen     [][]string
	umgebung    map[string]string
	// goFassung ist, was `go env GOVERSION` sagt — leer heißt go1.27.1.
	goFassung string
}

// kasten gibt den Werkzeugkasten des Doubles.
func (d *double) kasten() werkzeugkasten {
	return werkzeugkasten{
		git: func(_ context.Context, argumente ...string) (string, error) {
			schluessel := strings.Join(argumente, " ")

			return d.git[schluessel], d.gitFehler[schluessel]
		},
		gotest: func(_ context.Context, argumente []string, aus, fehler io.Writer) (int, error) {
			d.gerufen = append(d.gerufen, argumente)
			_, _ = io.WriteString(aus, d.strom)      //nolint:errcheck // ein Puffer im Test
			_, _ = io.WriteString(fehler, d.daneben) //nolint:errcheck // ein Puffer im Test

			return d.rueckgabe, d.startfehler
		},
		version: func(context.Context) string {
			if d.goFassung != "" {
				return d.goFassung
			}

			return "go1.27.1"
		},
		system:   func(context.Context) string { return "macOS 27.0" },
		jetzt:    func() time.Time { return fest },
		umgebung: func(name string) string { return d.umgebung[name] },
	}
}

// gruenerStrom ist ein Lauf, in dem der Pflichtfall besteht und der Trend
// rot ist — und damit ein grüner Lauf.
const gruenerStrom = `{"Action":"start","Package":"m.test/internal/app"}
{"Action":"run","Package":"m.test/internal/app","Test":"TestKanon"}
{"Action":"output","Package":"m.test/internal/app","Test":"TestKanon","Output":"    k_test.go:3: kennzahl kanon.portierung=0\n"}
{"Action":"pass","Package":"m.test/internal/app","Test":"TestKanon","Elapsed":1.5}
{"Action":"run","Package":"m.test/internal/app","Test":"TestAbstand"}
{"Action":"output","Package":"m.test/internal/app","Test":"TestAbstand","Output":"    k_test.go:9: Abstand 20,056\n","OutputType":"error"}
{"Action":"fail","Package":"m.test/internal/app","Test":"TestAbstand","Elapsed":2}
{"Action":"fail","Package":"m.test/internal/app","Elapsed":3.6}
`

// einPlan ist der Plan des Arbeitsbaums.
const einPlan = `{
  "lauf": "fremdpakete",
  "tag": "fremdpakete",
  "beleg": "docs/belege/fremdpakete.json",
  "roh": ".laufbeleg/fremdpakete.jsonl",
  "frist": "720h",
  "misst": ["*.go", "go.mod"],
  "eingaben": [{"name": "MSE_KANON", "muster": "*.mse-set"}],
  "schalter": {"MSE_STICHPROBE": "bestimmt, wie viele Karten gemessen werden"},
  "trend": {"internal/app.TestAbstand": "absichtlich rot, als Verlauf zu lesen"}
}`

// schreibe legt eine Datei an.
func schreibe(t *testing.T, pfad, inhalt string) {
	t.Helper()

	if fehler := os.MkdirAll(filepath.Dir(pfad), 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.WriteFile(pfad, []byte(inhalt), 0o600); fehler != nil {
		t.Fatal(fehler)
	}
}

// arbeitsbaum legt einen Baum mit einem markierten Lauf an und wechselt
// hinein. Er gibt das Double zurück, das an einem sauberen Stand misst.
func arbeitsbaum(t *testing.T) *double {
	t.Helper()

	wurzel := t.TempDir()
	t.Chdir(wurzel)

	schreibe(t, "go.mod", "module m.test\n\ngo 1.27\n")
	schreibe(t, "internal/app/k_test.go", "//go:build fremdpakete\n\npackage app\n\n"+
		"import \"testing\"\n\nfunc TestKanon(t *testing.T) {}\n\nfunc TestAbstand(t *testing.T) {}\n")
	schreibe(t, "docs/belege/fremdpakete.plan.json", einPlan)

	zuhause := filepath.Join(wurzel, "zuhause")
	schreibe(t, filepath.Join(zuhause, "kanon", "Archenemy.mse-set"), "set")

	return &double{
		git: map[string]string{
			"rev-parse HEAD": commit, status: "",
			inhalt: "100644 1a2b3c 0\tinternal/app/k_test.go",
			// **Der ganze Baum trägt den Beleg des vorigen Laufs** — er
			// zählt nicht, sonst hieße jeder Beleg nach dem ersten
			// »anderer Stand«.
			"status --porcelain": "?? docs/belege/fremdpakete.json",
		},
		gitFehler: map[string]error{},
		strom:     gruenerStrom, rueckgabe: 1,
		umgebung: map[string]string{
			"HOME": zuhause, "MSE_KANON": filepath.Join(zuhause, "kanon"),
		},
	}
}

// fahre ruft das Programm.
func fahre(d *double, argumente ...string) (rueckgabe int, ausgabe, fehlerausgabe string) {
	var aus, fehler bytes.Buffer

	rueckgabe = lauf(argumente, &aus, &fehler, d.kasten())

	return rueckgabe, aus.String(), fehler.String()
}

// belegLesen liest den geschriebenen Beleg.
func belegLesen(t *testing.T) laufbeleg.Beleg {
	t.Helper()

	datei, fehler := os.Open("docs/belege/fremdpakete.json")
	if fehler != nil {
		t.Fatalf("kein Beleg geschrieben: %v", fehler)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	beleg, fehler := laufbeleg.Lesen(datei)
	if fehler != nil {
		t.Fatal(fehler)
	}

	return beleg
}

// **Ein grüner Lauf hinterlässt einen vollständigen Beleg** (AK-01).
func TestEinGruenerLaufHinterlaesstEinenBeleg(t *testing.T) {
	d := arbeitsbaum(t)

	rueckgabe, aus, fehler := fahre(d, "-plan", "docs/belege/fremdpakete.plan.json")
	if rueckgabe != 0 {
		t.Fatalf("Rückgabe %d\n%s\n%s", rueckgabe, aus, fehler)
	}

	beleg := belegLesen(t)

	if beleg.Status != laufbeleg.Bestanden || beleg.Commit != commit || !beleg.BaumSauber ||
		!strings.HasPrefix(beleg.Inhalt, "sha256:") ||
		beleg.Erwartet != 2 || beleg.Gemessen != 2 || beleg.Kennzahlen["kanon.portierung"] != "0" ||
		!beleg.Anfang.Equal(fest) || beleg.Rueckgabe != 1 {
		t.Errorf("Beleg %+v", beleg)
	}

	if beleg.Werkzeuge["go"] != "go1.27.1" || beleg.Werkzeuge["system"] != "macOS 27.0" ||
		beleg.Werkzeuge["plattform"] == "" {
		t.Errorf("Werkzeuge %v", beleg.Werkzeuge)
	}

	// **Ein Schalter steht auch da, wenn er nicht gesetzt war** — leer, und
	// damit vergleichbar mit einem späteren Lauf, der ihn setzt.
	if wert, steht := beleg.Schalter["MSE_STICHPROBE"]; !steht || wert != "" {
		t.Errorf("Schalter %v", beleg.Schalter)
	}

	// **Das Zuhause steht nicht im Beleg**, der eingecheckt wird.
	if eingabe := beleg.Eingaben[0]; eingabe.Pfad != "~/kanon" || eingabe.Hash == "" {
		t.Errorf("Eingabe %+v", eingabe)
	}

	if len(beleg.Artefakte) != 1 || beleg.Artefakte[0].Name != "Rohprotokoll" ||
		beleg.Artefakte[0].Hash == "" {
		t.Errorf("Artefakte %+v", beleg.Artefakte)
	}

	// **Gemessen wird ohne Zwischenspeicher und nur, was erwartet ist.**
	gerufen := d.gerufen[0]
	for _, noetig := range []string{
		"test", "-json", "-count=1", "fremdpakete",
		"^(TestAbstand|TestKanon)$", "m.test/internal/app",
	} {
		if !slices.Contains(gerufen, noetig) {
			t.Errorf("go %v ohne %q", gerufen, noetig)
		}
	}

	if !strings.Contains(aus, "✓ internal/app.TestKanon") || !strings.Contains(aus, "PASS") {
		t.Errorf("Ausgabe:\n%s", aus)
	}
}

// **Fehlt eine Pflichteingabe, wird nicht gemessen** — und der Beleg sagt
// BLOCKED, nicht PASS (AK-02). Genau das war #1656.
func TestOhneEingabeWirdNichtGemessen(t *testing.T) {
	d := arbeitsbaum(t)
	delete(d.umgebung, "MSE_KANON")

	if rueckgabe, aus, _ := fahre(d, "-plan", "docs/belege/fremdpakete.plan.json"); rueckgabe != 1 ||
		!strings.Contains(aus, "BLOCKED") {
		t.Errorf("Rückgabe %d\n%s", rueckgabe, aus)
	}

	if len(d.gerufen) != 0 {
		t.Errorf("go test wurde trotzdem gerufen: %v", d.gerufen)
	}

	if beleg := belegLesen(t); beleg.Status != laufbeleg.Blockiert ||
		beleg.Grund != "MSE_KANON: nicht gesetzt" {
		t.Errorf("Status %s, Grund %q", beleg.Status, beleg.Grund)
	}
}

// **Lässt sich `go test` nicht starten, ist das ein gescheiterter Lauf** —
// und kein leerer grüner.
func TestEinUnstartbarerLaufIstGescheitert(t *testing.T) {
	d := arbeitsbaum(t)
	d.startfehler = errors.New("go fehlt")

	if rueckgabe, _, fehler := fahre(d, "-plan", "docs/belege/fremdpakete.plan.json"); rueckgabe != 1 ||
		!strings.Contains(fehler, "go fehlt") {
		t.Errorf("Rückgabe %d, Fehler %q", rueckgabe, fehler)
	}

	if beleg := belegLesen(t); beleg.Status != laufbeleg.Gescheitert {
		t.Errorf("Status %s", beleg.Status)
	}
}

// **Ändert sich der Baum während des Laufs, war er nicht sauber** — gemessen
// wurde dann ein Stand, den es als Commit nicht gibt.
func TestEinBaumDerSichAendertIstNichtSauber(t *testing.T) {
	d := arbeitsbaum(t)
	d.git[status] = " M internal/app/k.go"

	if rueckgabe, _, _ := fahre(d, "-plan", "docs/belege/fremdpakete.plan.json"); rueckgabe != 0 {
		t.Fatalf("Rückgabe %d", rueckgabe)
	}

	if beleg := belegLesen(t); beleg.BaumSauber {
		t.Error("ein Baum mit Änderungen stand als sauber im Beleg")
	}
}

// **Ohne Commit ist der Beleg kein grüner** — `git` antwortet nicht.
func TestOhneGitGibtEsKeinenGruenenBeleg(t *testing.T) {
	d := arbeitsbaum(t)
	d.gitFehler["rev-parse HEAD"] = errors.New("kein Repositorium")
	d.gitFehler[status] = errors.New("kein Repositorium")

	if rueckgabe, _, _ := fahre(d, "-plan", "docs/belege/fremdpakete.plan.json"); rueckgabe != 1 {
		t.Errorf("Rückgabe %d", rueckgabe)
	}

	if beleg := belegLesen(t); beleg.Status != laufbeleg.Gescheitert || beleg.BaumSauber {
		t.Errorf("Status %s, sauber %v", beleg.Status, beleg.BaumSauber)
	}
}

// **Was das Programm nicht ausführen kann, sagt es — mit Rückgabe 2.**
func TestEinFalscherAufrufWirdAbgewiesen(t *testing.T) {
	faelle := map[string]struct {
		vorbereiten func(*testing.T)
		argumente   []string
		meldung     string
	}{
		"ohne Plan":                   {func(*testing.T) {}, nil, "-plan"},
		"ein unbekannter Schalter":    {func(*testing.T) {}, []string{"-gibtsnicht"}, "gibtsnicht"},
		"ein Plan, den es nicht gibt": {func(*testing.T) {}, []string{"-plan", "fehlt.json"}, "fehlt.json"},
		"ein kaputter Plan": {
			func(t *testing.T) { schreibe(t, "kaputt.json", "{") },
			[]string{"-plan", "kaputt.json"},
			"kein Plan",
		},
		"ein Baum ohne go.mod": {
			func(t *testing.T) {
				if fehler := os.Remove("go.mod"); fehler != nil {
					t.Fatal(fehler)
				}
			},
			[]string{"-plan", "docs/belege/fremdpakete.plan.json"},
			"go.mod",
		},
		// **Wohin der Beleg soll, muss es geben können.**
		"ein Beleg, der nicht zu schreiben ist": {
			func(t *testing.T) {
				schreibe(t, "unmoeglich.json", strings.Replace(einPlan,
					`"docs/belege/fremdpakete.json"`, `"go.mod/beleg.json"`, 1))
			},
			[]string{"-plan", "unmoeglich.json"},
			"go.mod/beleg.json",
		},
		"ein Rohprotokoll, das nicht zu schreiben ist": {
			func(t *testing.T) {
				schreibe(t, "unmoeglich.json", strings.Replace(einPlan,
					`".laufbeleg/fremdpakete.jsonl"`, `"go.mod/roh.jsonl"`, 1))
			},
			[]string{"-plan", "unmoeglich.json"},
			"go.mod/roh.jsonl",
		},
		"ein Rohprotokoll, dessen Ort ein Verzeichnis ist": {
			func(t *testing.T) {
				schreibe(t, "unmoeglich.json", strings.Replace(einPlan,
					`".laufbeleg/fremdpakete.jsonl"`, `"docs"`, 1))
			},
			[]string{"-plan", "unmoeglich.json"},
			"das Rohprotokoll docs",
		},
		// **Ein Verzeichnis, das niemand lesen darf**, verkürzte die Liste
		// der erwarteten Fälle still.
		"ein gesperrtes Verzeichnis im Baum": {
			func(t *testing.T) {
				if fehler := os.Mkdir("gesperrt", 0o000); fehler != nil {
					t.Fatal(fehler)
				}

				t.Cleanup(func() { _ = os.Chmod("gesperrt", 0o750) }) //nolint:errcheck,gosec // nur zum Aufräumen; ein Verzeichnis braucht das Ausführungsrecht
			},
			[]string{"-plan", "docs/belege/fremdpakete.plan.json"},
			"gesperrt",
		},
	}

	for name, einer := range faelle {
		t.Run(name, func(t *testing.T) {
			d := arbeitsbaum(t)
			einer.vorbereiten(t)

			if rueckgabe, _, fehler := fahre(d, einer.argumente...); rueckgabe != 2 ||
				!strings.Contains(fehler, einer.meldung) {
				t.Errorf("Rückgabe %d, Fehler %q — erwartet 2 und %q", rueckgabe, fehler, einer.meldung)
			}
		})
	}
}
