package main

// **Die Wege neben dem Hauptweg** — abgetrennt von `lauf_test.go`, als die
// Datei über die Zeilengrenze aus §2 ging.

import (
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/laufbeleg"
)

// **Was neben dem Strom kommt, wird gezeigt** — und macht den Lauf nicht grün
// oder rot: stderr von `go test`, eine Zeile, die kein Ereignis ist.
func TestWasNebenDemStromKommtWirdGezeigt(t *testing.T) {
	d := arbeitsbaum(t)
	d.daneben = "ld: warning: -bind_at_load is deprecated\n"
	d.strom = "go: downloading beispiel.test v1.0.0\n" + gruenerStrom

	rueckgabe, _, fehler := fahre(d, "-plan", "docs/belege/fremdpakete.plan.json")
	if rueckgabe != 0 || !strings.Contains(fehler, "go test schrieb daneben:\nld: warning") ||
		!strings.Contains(fehler, "keine Zeile von go test: go: downloading") {
		t.Errorf("Rückgabe %d, Fehler %q", rueckgabe, fehler)
	}
}

// **Ein Strom, der sich nicht zu Ende lesen lässt, ist ein abgebrochener
// Lauf** — und kein kürzerer grüner.
func TestEinUnlesbarerStromIstAbgebrochen(t *testing.T) {
	d := arbeitsbaum(t)
	d.strom = gruenerStrom + `{"Action":"output","Output":"` + strings.Repeat("x", 17<<20) + "\"}\n"

	rueckgabe, _, fehler := fahre(d, "-plan", "docs/belege/fremdpakete.plan.json")
	if rueckgabe != 1 || !strings.Contains(fehler, "nicht zu Ende zu lesen") {
		t.Errorf("Rückgabe %d, Fehler %q", rueckgabe, fehler)
	}

	if beleg := belegLesen(t); beleg.Status != laufbeleg.Gescheitert || beleg.Rueckgabe != -1 {
		t.Errorf("Status %s, Rückgabe %d", beleg.Status, beleg.Rueckgabe)
	}
}

// **Ein Plan ohne markierte Fälle misst nichts** — und sagt es, statt grün
// zu sein.
func TestEinPlanOhneFaelleMisstNichts(t *testing.T) {
	d := arbeitsbaum(t)
	schreibe(t, "leer.json", strings.Replace(strings.Replace(einPlan,
		`"tag": "fremdpakete"`, `"tag": "gibtsnicht"`, 1),
		`"trend": {"internal/app.TestAbstand": "absichtlich rot, als Verlauf zu lesen"}`, `"trend": {}`, 1))

	if rueckgabe, _, _ := fahre(d, "-plan", "leer.json"); rueckgabe != 1 || len(d.gerufen) != 0 {
		t.Errorf("Rückgabe %d, gerufen %v", rueckgabe, d.gerufen)
	}

	if beleg := belegLesen(t); beleg.Status != laufbeleg.NichtGelaufen {
		t.Errorf("Status %s", beleg.Status)
	}
}

// **Das Zeitlimit des Plans geht an `go test`** — ohne eines bleibt dessen
// Vorgabe von zehn Minuten, und die reicht für den Kanonlauf nicht.
func TestDasZeitlimitGehtAnGoTest(t *testing.T) {
	d := arbeitsbaum(t)
	schreibe(t, "lang.json", strings.Replace(einPlan, `"frist": "720h",`,
		`"frist": "720h", "zeitlimit": "7200s",`, 1))

	fahre(d, "-plan", "docs/belege/fremdpakete.plan.json")
	fahre(d, "-plan", "lang.json")

	if slices.Contains(d.gerufen[0], "-timeout") {
		t.Errorf("ohne Zeitlimit im Plan gab es eines: %v", d.gerufen[0])
	}

	if i := slices.Index(d.gerufen[1], "-timeout"); i < 0 || d.gerufen[1][i+1] != "2h0m0s" {
		t.Errorf("mit Zeitlimit im Plan: %v", d.gerufen[1])
	}
}
