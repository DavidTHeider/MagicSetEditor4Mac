// **Ein abgelöster Vermerk wird nicht als geltend zitiert** — Sichtprüfung
// über die ganze Anwendung, 2026-09-07.
//
// # Der Befund
//
// `README.md` sagte in der Betriebstabelle:
//
//	| **Verschlüsselung** | keine, bewusst ([0042](…)) |
//
// **Und 0042 trägt seit dem 2026-09-06 in seiner ersten Zeile „Abgelöst … durch
// 0072". Der Bestand wird verschlüsselt**, seit Werkstück 14 von Phase 4.12 —
// die Nutzerseite im Fenster verwies schon darauf, während das `README.md`
// das Gegenteil sagte.
//
// **Und `docs/ADMINISTRATION.md` sagte es ein zweites Mal**, hundertsiebzig
// Zeilen unter ihrem eigenen Abschnitt „Verschlüsselung".
//
// # Was er hält
//
// **Ein vollständig abgelöster Vermerk darf nur dort zitiert werden, wo sein
// Nachfolger in derselben Zeile steht.** Das ist die Form, in der ein Verweis
// auf die **Geschichte** aussieht („hier stand …, mit dem abgelösten Vermerk
// X — abgelöst durch Y"); ein Verweis auf die **Begründung von heute** hat den
// Nachfolger nicht daneben.
//
// # Was er nicht hält
//
//   - **Ob der Nachfolger daneben auch der richtige ist.** Ein beliebiger
//     zweiter Verweis genügt ihm.
//   - **Teilweise abgelöste Vermerke.** „In einem Punkt abgelöst" (0070) und
//     „in seinem Lizenzteil überholt" (0001) gelten weiter — für den Rest. Sie
//     zu fordern hieße, jeden Verweis auf sie zu verbieten, und der wäre
//     meistens richtig.
//   - **Die Chronik.** Sie ist Geschichte, von der ersten bis zur letzten
//     Zeile; ein Verweis dort ist nie eine Aussage über heute.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// abloesung findet die Kennzeichnung im Kopf eines Vermerks.
//
// **`Abgelöst` am Zeilenanfang und ohne Einschränkung davor** — „In einem
// Punkt abgelöst" trifft nicht, und das ist Absicht.
var abloesung = regexp.MustCompile(`(?m)^>?\s*\*\*Abgelöst am `)

// vermerkverweis findet `[0042](…/0042-…md)` in einem Fließtext.
var vermerkverweis = regexp.MustCompile(`\(([^)]*?/)?(\d{4})-[a-z0-9-]+\.md\)`)

// gelesenOhneGeschichte sind die Dateien, die eine Aussage über **heute**
// machen. Die Chronik steht bewusst nicht dabei.
var gelesenOhneGeschichte = []string{
	"README.md",
	filepath.Join("docs", "ADMINISTRATION.md"),
	filepath.Join("docs", "ARCHITEKTUR.md"),
	filepath.Join("docs", "HANDBUCH.md"),
	filepath.Join("docs", "RELEASE.md"),
	"AGENTS.md",
}

func TestKeinAbgeloesterVermerkGiltAlsBegruendung(t *testing.T) {
	t.Parallel()

	abgeloest := abgeloesteVermerke(t)

	// **Die Gegenprobe zur Messung**: findet sie keinen, ist sie grün und sagt
	// nichts (§10). Am 2026-09-07 war es genau einer — 0042.
	if len(abgeloest) == 0 {
		t.Fatal("kein Vermerk trägt die Kennzeichnung `**Abgelöst am ` — " +
			"heißt sie noch so?")
	}

	var befunde []string

	for _, datei := range gelesenOhneGeschichte {
		// **Gelesen wird je Absatz und nicht je Zeile.**
		//
		// Markdown bricht einen Satz um; ein Verweis und sein Nachfolger
		// stehen dann in zwei Zeilen, und derselbe Gedanke sähe für einen
		// zeilenweisen Wächter aus wie zwei. Der erste Anlauf las Zeilen und
		// meldete prompt einen Absatz an, der seinen Nachfolger eine Zeile
		// tiefer nennt — **ein Fehlalarm, den er selbst erzeugt hat.**
		//
		// **Und eine Tabellenzeile ist ein Absatz für sich.** Das stand hier
		// als Behauptung über „zwischen zwei Leerzeilen" — und war falsch:
		// eine Markdown-Tabelle ist ein Block **ohne** Leerzeilen, also ein
		// einziger Absatz mit vielen Verweisen, und `len(genannt) < 2` traf
		// nie zu. Die Gegenprobe blieb grün, und genau dafür gibt es sie
		// (§10): ein Eingriff, der nichts ändert, sagt nichts über den Test.
		//
		// Jede Zeile, die mit `|` beginnt, steht deshalb für sich.
		for nr, absatz := range absaetzeMitZeile(lies(t, datei)) {
			genannt := map[string]bool{}
			for _, treffer := range vermerkverweis.FindAllStringSubmatch(
				absatz.text, -1) {
				genannt[treffer[2]] = true
			}

			_ = nr

			for nummer := range genannt {
				// **Der Nachfolger im selben Absatz macht es zu einem
				// Verweis auf die Geschichte.**
				if abgeloest[nummer] && len(genannt) < 2 {
					befunde = append(befunde, filepath.ToSlash(datei)+":"+
						itoa(absatz.zeile)+" nennt "+nummer)
				}
			}
		}
	}

	if len(befunde) > 0 {
		sort.Strings(befunde)
		t.Errorf("diese Stellen zitieren einen abgelösten Vermerk als "+
			"Begründung von heute:\n  %s\n\nWer die Geschichte meint, nennt "+
			"den Nachfolger in derselben Zeile", strings.Join(befunde, "\n  "))
	}
}

// absatz ist ein Textblock mit der Zeile, in der er beginnt.
type absatz struct {
	text  string
	zeile int
}

// absaetzeMitZeile zerlegt einen Fließtext an Leerzeilen.
func absaetzeMitZeile(inhalt string) []absatz {
	var (
		gefunden []absatz
		gerade   []string
		beginn   int
	)

	schliessen := func() {
		if len(gerade) > 0 {
			gefunden = append(gefunden,
				absatz{text: strings.Join(gerade, " "), zeile: beginn})
			gerade = nil
		}
	}

	for nr, zeile := range strings.Split(inhalt, "\n") {
		blank := strings.TrimSpace(zeile)

		if blank == "" {
			schliessen()

			continue
		}

		// **Eine Tabellenzeile ist eine Aussage für sich** — der Grund steht
		// beim Aufrufer.
		if strings.HasPrefix(blank, "|") {
			schliessen()

			gefunden = append(gefunden, absatz{text: zeile, zeile: nr + 1})

			continue
		}

		if len(gerade) == 0 {
			beginn = nr + 1
		}

		gerade = append(gerade, zeile)
	}

	schliessen()

	return gefunden
}

// abgeloesteVermerke gibt die Nummern der vollständig abgelösten Vermerke.
func abgeloesteVermerke(t *testing.T) map[string]bool {
	t.Helper()

	ort := filepath.Join("..", "..", "docs", "entscheidungen")

	eintraege, err := os.ReadDir(ort)
	if err != nil {
		t.Fatalf("%s lesen: %v", ort, err)
	}

	gefunden := map[string]bool{}

	for _, eintrag := range eintraege {
		name := eintrag.Name()
		if len(name) < 5 || !strings.HasSuffix(name, ".md") {
			continue
		}

		roh, lesefehler := os.ReadFile(filepath.Join(ort, name)) // #nosec G304 -- fester Ordner
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", name, lesefehler)
		}

		// **Nur der Kopf**: eine Ablösung steht oben, und weiter unten kann
		// derselbe Satz über einen **anderen** Vermerk stehen.
		kopf := string(roh)
		if len(kopf) > 1200 {
			kopf = kopf[:1200]
		}

		if abloesung.MatchString(kopf) {
			gefunden[name[:4]] = true
		}
	}

	return gefunden
}
