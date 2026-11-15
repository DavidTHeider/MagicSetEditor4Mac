package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **`-pruefen` sagt, ob der Beleg zum heutigen Stand passt** (AK-04) — über
// den Inhalt dessen, was der Lauf misst, und nicht über den Commit: nach einem
// Squash-Merge liegt der gemessene Commit in keiner Geschichte mehr. **Die
// Rückgabe sagt, ob er trägt**: 0 nur für einen grünen, aktuellen Beleg.
func TestDerBelegWirdGegenDenStandGeprueft(t *testing.T) {
	faelle := map[string]struct {
		stand, heute string
		rot          bool
		rueckgabe    int
		ausgabe      string
	}{
		"derselbe Inhalt":                     {commit, "", false, 0, "fremdpakete: PASS · aktuell"},
		"ein anderer Commit, derselbe Inhalt": {"fff0000", "", false, 0, "aktuell"},
		"ein anderer Inhalt":                  {"fff0000", "100644 9f9f9f 0\tinternal/app/k.go", false, 1, "anderer Stand — seit abc1234"},
		"aktuell, aber rot":                   {commit, "", true, 1, "fremdpakete: FAIL · aktuell"},
	}

	for name, einer := range faelle {
		t.Run(name, func(t *testing.T) {
			d := arbeitsbaum(t)
			// **Rot heißt: der Pflichtfall scheitert** — ein roter Trend
			// erklärte den Rückgabewert und hielte den Lauf grün.
			if einer.rot {
				d.strom = strings.Replace(gruenerStrom,
					`"Action":"pass","Package":"m.test/internal/app","Test":"TestKanon"`,
					`"Action":"fail","Package":"m.test/internal/app","Test":"TestKanon"`, 1)
			}

			fahre(d, "-plan", "docs/belege/fremdpakete.plan.json")

			d.git["rev-parse HEAD"] = einer.stand
			if einer.heute != "" {
				d.git[inhalt] = einer.heute
			}

			rueckgabe, aus, fehler := fahre(d, "-plan", "docs/belege/fremdpakete.plan.json", "-pruefen")
			if rueckgabe != einer.rueckgabe || !strings.Contains(aus, einer.ausgabe) {
				t.Errorf("Rückgabe %d, Ausgabe %q, Fehler %q — erwartet %d und %q",
					rueckgabe, aus, fehler, einer.rueckgabe, einer.ausgabe)
			}
		})
	}
}

// **Was in die Messung eingeht, zählt auch in der Prüfung** — der
// Arbeitsbaum dort, wo der Lauf misst, und jede Pflichteingabe. Bis zum
// Zweitblick über Phase 6.0 hieß ein Beleg »aktuell«, während eine Datei im
// Baum oder das Verzeichnis der Referenzbilder schon anders war.
func TestDieEingabenUndDerArbeitsbaumZaehlenMit(t *testing.T) {
	faelle := map[string]struct {
		vorbereiten func(*testing.T, *double)
		ausgabe     string
	}{
		"ein Arbeitsbaum, der abweicht": {func(_ *testing.T, d *double) {
			d.git[offen] = "M internal/app/k_test.go\n?? internal/app/neu.go"
		}, "anderer Stand — der Arbeitsbaum weicht ab, wo der Lauf misst: internal/app/k_test.go, internal/app/neu.go"},
		"eine Eingabe mit anderem Inhalt": {func(t *testing.T, d *double) {
			schreibe(t, filepath.Join(d.umgebung["MSE_KANON"], "Archenemy.mse-set"), "anderes set")
		}, "anderer Stand — die Eingabe MSE_KANON hat heute einen anderen Inhalt"},
		"eine Eingabe, die heute nicht gesetzt ist": {func(_ *testing.T, d *double) {
			delete(d.umgebung, "MSE_KANON")
		}, "anderer Stand — die Eingabe MSE_KANON ist heute nicht zu prüfen: nicht gesetzt"},
		// **Und womit gemessen wurde** — Nachlauf zum Zweitblick über Phase 6.0.
		"eine andere Go-Fassung": {func(_ *testing.T, d *double) {
			d.goFassung = "go1.27.2"
		}, "anderer Stand — gemessen mit go go1.27.1, heute go1.27.2"},
		"ein Schalter, der heute gesetzt ist": {func(_ *testing.T, d *double) {
			d.umgebung["MSE_STICHPROBE"] = "alle"
		}, "anderer Stand — gemessen ohne MSE_STICHPROBE, heute MSE_STICHPROBE=alle"},
	}

	for name, einer := range faelle {
		t.Run(name, func(t *testing.T) {
			d := arbeitsbaum(t)
			fahre(d, "-plan", "docs/belege/fremdpakete.plan.json")

			// **Vorher trägt er** — sonst prüfte der Fall einen Beleg, der
			// ohnehin nicht passt.
			if rueckgabe, aus, _ := fahre(d, "-plan", "docs/belege/fremdpakete.plan.json",
				"-pruefen"); rueckgabe != 0 {
				t.Fatalf("vor dem Eingriff: Rückgabe %d, %q", rueckgabe, aus)
			}

			einer.vorbereiten(t, d)

			rueckgabe, aus, fehler := fahre(d, "-plan", "docs/belege/fremdpakete.plan.json", "-pruefen")
			if rueckgabe != 1 || !strings.Contains(aus, einer.ausgabe) {
				t.Errorf("Rückgabe %d, Ausgabe %q, Fehler %q — erwartet 1 und %q",
					rueckgabe, aus, fehler, einer.ausgabe)
			}
		})
	}
}

// **`-ohne-eingaben` rechnet die Eingaben nicht und sagt es** — für die
// Reichweite am Ende von `make ci`, die sie nicht kennt. **Rückgabe 1**: ein
// Beleg, dessen Eingaben niemand geprüft hat, trägt nicht.
func TestOhneEingabenWirdNichtsGerechnetUndDasGesagt(t *testing.T) {
	d := arbeitsbaum(t)
	fahre(d, "-plan", "docs/belege/fremdpakete.plan.json")

	// **Eine Eingabe, die heute fehlt, fiele sonst auf** — hier nicht.
	delete(d.umgebung, "MSE_KANON")

	rueckgabe, aus, fehler := fahre(d, "-plan", "docs/belege/fremdpakete.plan.json",
		"-pruefen", "-ohne-eingaben")
	if rueckgabe != 1 || !strings.Contains(aus, "fremdpakete: PASS · ohne Eingaben — gemessen an abc1234") {
		t.Errorf("Rückgabe %d, Ausgabe %q, Fehler %q", rueckgabe, aus, fehler)
	}
}

// **Ohne Beleg gibt es nichts zu prüfen** — und das ist ein Befund, keine
// stille Null.
func TestOhneBelegIstDiePruefungRot(t *testing.T) {
	d := arbeitsbaum(t)

	if rueckgabe, aus, _ := fahre(d, "-plan", "docs/belege/fremdpakete.plan.json", "-pruefen"); rueckgabe != 1 ||
		!strings.Contains(aus, "fremdpakete: kein Beleg unter docs/belege/fremdpakete.json") {
		t.Errorf("Rückgabe %d, Ausgabe %q", rueckgabe, aus)
	}
}

// **Was die Prüfung nicht beantworten kann, sagt sie mit Rückgabe 2.**
func TestEinePruefungOhneAntwortIstEinFehler(t *testing.T) {
	faelle := map[string]struct {
		vorbereiten func(*testing.T, *double)
		meldung     string
	}{
		"ein kaputter Beleg": {func(t *testing.T, _ *double) {
			schreibe(t, "docs/belege/fremdpakete.json", "{")
		}, "kein Beleg"},
		"ein Beleg, den niemand lesen darf": {func(t *testing.T, _ *double) {
			if fehler := os.Chmod("docs/belege/fremdpakete.json", 0o000); fehler != nil {
				t.Fatal(fehler)
			}
		}, "permission denied"},
		"git kann den Inhalt nicht lesen": {func(_ *testing.T, d *double) {
			d.gitFehler[inhalt] = errors.New("kein Repositorium")
		}, "kein Repositorium"},
		"git kann den Arbeitsbaum nicht lesen": {func(_ *testing.T, d *double) {
			d.gitFehler[offen] = errors.New("kein Arbeitsbaum")
		}, "kein Arbeitsbaum"},
	}

	for name, einer := range faelle {
		t.Run(name, func(t *testing.T) {
			d := arbeitsbaum(t)
			fahre(d, "-plan", "docs/belege/fremdpakete.plan.json")
			einer.vorbereiten(t, d)

			if rueckgabe, _, fehler := fahre(d, "-plan", "docs/belege/fremdpakete.plan.json",
				"-pruefen"); rueckgabe != 2 || !strings.Contains(fehler, einer.meldung) {
				t.Errorf("Rückgabe %d, Fehler %q — erwartet 2 und %q", rueckgabe, fehler, einer.meldung)
			}
		})
	}
}

// kaputt ist eine Ausgabe, die nichts annimmt.
type kaputt struct{}

// Write weist jedes Stück ab.
func (kaputt) Write([]byte) (int, error) { return 0, errors.New("die Leitung ist zu") }

// **Eine Ausgabe, die nicht ankommt, ist ein Fehler des Laufs** — auch wenn
// der Beleg geschrieben wurde: wer auf den Bildschirm sah, sah ihn nicht.
func TestEineAusgabeDieNichtAnkommtIstEinFehler(t *testing.T) {
	d := arbeitsbaum(t)

	var fehler strings.Builder
	if rueckgabe := lauf([]string{"-plan", "docs/belege/fremdpakete.plan.json"}, kaputt{},
		&fehler, d.kasten()); rueckgabe != 2 || !strings.Contains(fehler.String(), "die Leitung ist zu") {
		t.Errorf("Rückgabe %d, Fehler %q", rueckgabe, fehler.String())
	}
}
