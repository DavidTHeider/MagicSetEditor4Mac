package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// **Jede Bindung nennt ihr Profil** — Werkstück 2 von Phase 4.12
// (#1351).
//
// # Was er hält
//
// Jede exportierte Methode auf `*App`, deren erster Parameter ein
// `context.Context` ist, beginnt mit `ctx = a.imProfil(ctx)`. Ohne die Zeile
// erreicht der Kontext den Speicher ohne Profil, und `storage.ProfilAus` weist
// ihn ab — die Bindung antwortet dann mit „nennt kein Profil", was wie ein
// Bestandsfehler aussieht und keiner ist.
//
// # Was er **nicht** hält
//
// **Er sieht nicht, ob eine Bindung den Speicher überhaupt anfasst.** Eine, die
// nur rechnet, trägt die Zeile ohne Nutzen. Das ist die billigere Seite des
// Handels: eine Aufzählung der Ausnahmen wäre eine zweite Liste, die
// veralteten kann, und der Preis für eine Zeile zu viel ist null.
//
// **Und er sieht nicht, ob das Profil das richtige ist.** Dass `a.profil` den
// Zeiger aus dem Bestand trägt, hält `TestDerStartLiestDenProfilzeiger` in
// `internal/app`.
func TestJedeBindungNenntIhrProfil(t *testing.T) {
	// **Beide Gestalten**, und die zweite ist der Grund für diesen Absatz: die
	// erste Fassung las nur einzeilige Signaturen und war grün — am
	// 2026-09-04 hatten **47 mehrzeilige** die Zeile nicht. Gefunden hat es
	// nicht der Wächter, sondern ein Test, dem der Kontext ausging.
	koepfe := []*regexp.Regexp{
		regexp.MustCompile(
			`func \(a \*App\) ([A-Z]\w*)\(ctx context\.Context[^\n]*\{\n(\t[^\n]*)`),
		// **`(?m)` macht `^` zum Zeilenanfang** — gocritic liest das nicht mit
		// und hält das zweite für überflüssig; ohne es fände der Ausdruck jede
		// schließende Klammer im Rumpf.
		//nolint:gocritic // badRegexp übersieht das (?m) am Anfang
		regexp.MustCompile(
			`(?m)^func \(a \*App\) ([A-Z]\w*)\(\n\tctx context\.Context[\s\S]*?^\) [^\n]*\{\n(\t[^\n]*)`),
	}

	verzeichnis := filepath.Join("..", "app")

	eintraege, err := os.ReadDir(verzeichnis)
	if err != nil {
		t.Fatalf("das Verzeichnis lesen: %v", err)
	}

	var ohne []string

	gezaehlt := 0

	for _, eintrag := range eintraege {
		name := eintrag.Name()
		if eintrag.IsDir() || !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") {
			continue
		}

		roh, lesefehler := os.ReadFile(filepath.Join(verzeichnis, name)) //nolint:gosec // fester Ort
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", name, lesefehler)
		}

		for _, kopf := range koepfe {
			for _, treffer := range kopf.FindAllStringSubmatch(string(roh), -1) {
				gezaehlt++

				if strings.TrimSpace(treffer[2]) != "ctx = a.imProfil(ctx)" {
					ohne = append(ohne, name+": "+treffer[1])
				}
			}
		}
	}

	if len(ohne) > 0 {
		t.Errorf("diese Bindungen nennen kein Profil: %s\n"+
			"Erste Zeile im Rumpf: ctx = a.imProfil(ctx)", strings.Join(ohne, ", "))
	}

	// **Die Reichweite der Messung** (§10): ein Ausdruck, der nichts findet,
	// wäre grün. Die Zahl darf wachsen und schrumpfen — sie darf nicht auf
	// eine Handvoll fallen, denn dann liest der Ausdruck die Datei nicht mehr.
	if gezaehlt < 160 {
		t.Errorf("nur %d Bindungen gesehen — der Ausdruck greift nicht mehr", gezaehlt)
	}
}
