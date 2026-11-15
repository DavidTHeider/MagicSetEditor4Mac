// **Eine Zahl in der Gegenwartsform stimmt, oder sie trägt ihr Datum**
// (#1287).
//
// # Der Befund
//
// Vier Stellen im Repositorium nannten eine Bindungszahl in der
// Gegenwartsform, und **keine** davon stimmte:
//
//	Erhoben am 2026-08-29, gezählt waren 142
//	AGENTS.md, CLAUDE.md §16          „Es sind 64 Bindungen"
//	docs/CODEX-NAECHSTE-RUNDE.md      „Die Wails-Naht trägt 64 Bindungen"
//	internal/werkzeuge/leerelisten…   „Ein Mensch, der 139 Bindungen durchsieht"
//	docs/FAEHIGKEITEN.md              „143 Bindungen am Fenster"
//
// Nur die letzte stand unter Aufsicht — und die zählte um eine Methode zu
// weit. Die drei anderen hatte nie jemand angesehen, weil kein Wächter über
// sie hinweglief: `flaechengroesse_test.go` liest **eine** Datei.
//
// # Was dieser Wächter hält
//
// Jede Stelle in `.md` und `.go`, die eine der vier Flächen mit einer Zahl
// benennt, hält eine von zwei Bedingungen ein:
//
//  1. **Die Zahl stimmt** — dann darf sie in der Gegenwart stehen.
//  2. **Ihr Absatz nennt ein Datum** (`2026-08-25`) — dann ist sie eine
//     historische Angabe und darf abweichen.
//
// Damit ist die Regel aus §10 mechanisch: *wer eine Zahl aufschreibt, schreibt
// dazu, worüber sie erhoben wurde.* Ein Datum ist die kleinste Form davon.
//
// # Was er nicht hält
//
//   - **Zahlwörter.** „vierundsechzig Bindungen" geht durch. Ein Wächter, der
//     deutsche Zahlwörter beugte, wäre eine Übung in Grammatik.
//   - **Mehrdeutige Wörter.** Bloßes „Werkzeuge" oder „Fähigkeiten" zählt er
//     nicht mit: `internal/werkzeuge` heißt auch so, und eine Fähigkeit ist
//     keine Fläche. Was er kennt, steht in `flaechenwoerter`.
//   - **Eine eingrenzende Beifügung.** „12 Feld-Bindungen" geht durch, und das
//     ist Absicht: eine Beifügung verkleinert die Menge fast immer, und ein
//     Wächter, der sie auf 142 zwänge, hätte häufiger unrecht als recht. Was
//     er kennt, sind die **ganzen** Namen einer Fläche — `Wails-Bindungen`
//     steht deshalb ausdrücklich in der Liste, seit der Zweitblick über Phase
//     4.0 ihn damit umgangen hat (Befund 2), und `Bindings` samt
//     `Wails-Bindings` seit dem Nachlauf: die englische Schreibweise steht im
//     `README.md` und war keine Beifügung, sondern derselbe Name. Jede weitere
//     Schreibweise ist ein Eintrag und keine Regel — und dass die Liste
//     zweimal nachgezogen werden musste, ist der Preis dafür.
//   - **Ob das Datum bei der Zahl steht.** Er sieht den Absatz, nicht den
//     Satz. Ein Absatz mit einem Datum und drei Zahlen aus drei Jahren wäre
//     grün — und wäre trotzdem schlecht geschrieben.
//   - **Ob die Zahl je gestimmt hat.** Ein Datum belegt eine Absicht, keine
//     Messung.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// flaechenwort ordnet eine Schreibweise im Text einer der vier Flächen zu.
type flaechenwort struct {
	// wort steht im Text hinter der Zahl.
	wort string

	// flaeche ist das `wort` aus `dieVierFlaechen`.
	flaeche string
}

// flaechenwoerter sind die Schreibweisen, die der Wächter auflösen kann.
//
// **Die längeren zuerst**: „Endpunkte im Dienst" muss vor „Endpunkte"
// geprüft werden, sonst zerfiele derselbe Fund in zwei.
var flaechenwoerter = []flaechenwort{
	{"Wails-Bindungen", "Bindungen am Fenster"},
	{"Wails-Bindings", "Bindungen am Fenster"},
	{"Bindungen", "Bindungen am Fenster"},
	{"Bindings", "Bindungen am Fenster"},
	{"MCP-Werkzeuge", "MCP-Werkzeuge"},
	{"Befehle auf der Kommandozeile", "Befehle auf der Kommandozeile"},
	{"CLI-Befehle", "Befehle auf der Kommandozeile"},
	{"Endpunkte im Dienst", "Endpunkte im Dienst"},
	{"Endpunkte", "Endpunkte im Dienst"},
}

// datumImText trifft ein Datum in ISO-Schreibweise.
var datumImText = regexp.MustCompile(`\b20\d\d-\d\d-\d\d\b`)

// zahlVorWort baut das Muster „<Zahl> <Wort>" — auch mit Auszeichnung
// (`**7** Zwerge`) und über einen Zeilenumbruch hinweg.
//
// **Das Beispiel nennt bewusst kein Flächenwort.** Hier stand eine
// ausgezeichnete Zahl vor dem Wort »Bindungen«, und dieser Wächter hat sich
// damit am 2026-08-30 **selbst** angestrichen, als
// die Zahl von 142 auf 144 stieg: ein Beispiel für eine Schreibweise ist keine Aussage
// über dieses Repositorium, und es als eine zu zählen kostet bei jeder neuen
// Bindung eine Korrektur an einer Stelle, die nichts behauptet.
func zahlVorWort(wort string) *regexp.Regexp {
	return regexp.MustCompile(`\*{0,2}(\d+)\*{0,2}\s+` + regexp.QuoteMeta(wort) + `\b`)
}

// datierteAbschriften sind Dateien, die als Ganzes einen Stand festhalten.
//
// **Sie sind ausgenommen, und das ist eine Entscheidung und keine Bequemlichkeit.**
// Ein Audit, ein Zweitblick und die Chronik beschreiben, wie es an einem Tag
// war — unter `docs/audits/` liegen Aufträge und Ergebnisse, die einmal
// geschrieben und nie fortgeschrieben werden (J05, #2269); ihre Zahlen
// sollen **nicht** mitwachsen, sonst wäre der Verlauf weg. Sie tragen ihr
// Datum in der Überschrift des Eintrags, nicht im Absatz — und ein Wächter,
// der Überschriften über Absätze hinweg zuordnete, prüfte Gliederung statt
// Zahlen.
//
// **Der Preis steht hier**: eine falsche Zahl in einem Audit fällt niemandem
// auf. Was dagegen hilft, ist, dass ein Audit einmal geschrieben und nicht
// fortgeschrieben wird.
var datierteAbschriften = []string{
	"docs/CHRONIK.md",
	"CHANGELOG.md",
}

// istDatierteAbschrift sagt, ob ein Pfad eine solche Datei ist.
func istDatierteAbschrift(pfad string) bool {
	if strings.HasPrefix(pfad, "docs/AUDIT-") || strings.HasPrefix(pfad, "docs/audits/") {
		return true
	}

	return slices.Contains(datierteAbschriften, pfad)
}

// TestJedeFlaechenzahlStimmtOderTraegtIhrDatum ist der Wächter selbst.
func TestJedeFlaechenzahlStimmtOderTraegtIhrDatum(t *testing.T) {
	t.Parallel()

	zahlen := map[string]int{}
	for _, was := range dieVierFlaechen() {
		zahlen[was.wort] = flaechentreffer(t, was)
	}

	funde := 0

	for _, pfad := range fliesstextdateien(t) {
		if istDatierteAbschrift(pfad) {
			continue
		}

		for _, absatz := range absaetze(t, pfad) {
			funde += absatzPruefen(t, pfad, absatz, zahlen)
		}
	}

	// **Sonst wäre ein kaputter Leser grün.** Ein Wächter, der nichts findet,
	// hält keine Zusage — er behauptet nur, keine Verletzung gesehen zu haben.
	if funde < 4 {
		t.Fatalf("nur %d Flächenzahlen im ganzen Repositorium gefunden — der Leser greift daneben",
			funde)
	}
}

// absatzPruefen prüft einen Absatz und gibt die Zahl der Fundstellen.
func absatzPruefen(t *testing.T, pfad string, absatz absatz, zahlen map[string]int) int {
	t.Helper()

	datiert := datumImText.MatchString(absatz.text)
	gesehen := map[int]bool{}
	funde := 0

	for _, wort := range flaechenwoerter {
		for _, treffer := range zahlVorWort(wort.wort).FindAllStringSubmatchIndex(absatz.text, -1) {
			// **Dieselbe Stelle nicht zweimal**: eine Angabe wie „… Endpunkte
			// im Dienst" trifft auch das kürzere „Endpunkte".
			if gesehen[treffer[2]] {
				continue
			}

			gesehen[treffer[2]] = true
			funde++

			if datiert {
				continue
			}

			genannt, err := strconv.Atoi(absatz.text[treffer[2]:treffer[3]])
			if err != nil {
				t.Errorf("%s:%d: %q ist keine Zahl", pfad, absatz.zeile,
					absatz.text[treffer[2]:treffer[3]])

				continue
			}

			if erwartet := zahlen[wort.flaeche]; genannt != erwartet {
				t.Errorf("%s:%d nennt %d %s — gezählt sind %d. "+
					"Entweder die Zahl stimmt, oder ihr Absatz nennt ein Datum",
					pfad, absatz.zeile, genannt, wort.wort, erwartet)
			}
		}
	}

	return funde
}

// absatz ist ein Textblock samt der Zeile, in der er beginnt.
type absatz struct {
	text  string
	zeile int
}

// absaetze zerlegt eine Datei in Absätze.
//
// **In `.go` trennt eine nackte `//`-Zeile**, wie in einem Doc-Kommentar. Ohne
// das wäre ein sechzig Zeilen langer Kommentarkopf **ein** Absatz, und ein
// Datum irgendwo darin deckte jede Zahl darunter — die Ausnahme fräße die
// Regel.
func absaetze(t *testing.T, pfad string) []absatz {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join(wurzel, pfad)) //nolint:gosec // Pfad aus dem eigenen Baum
	if err != nil {
		t.Fatalf("%s lesen: %v", pfad, err)
	}

	// **Der Kommentarstrich fällt weg, sonst ist ein Kopf ein einziger Absatz.**
	//
	// Hier stand `strings.HasSuffix(pfad, ".go")`, und für `.mjs` blieb das
	// `//` stehen: eine Leerzeile im Kommentar heißt dort `//`, und die ist
	// nicht leer. Damit wurde ein vierzigzeiliger Kopf **ein** Absatz — und
	// ein Datum irgendwo darin nahm jede Zahl darin von der Prüfung aus.
	//
	// **Gemessen am 2026-08-31**: von den zwei veralteten Bindungszahlen in
	// `scripts/` fiel genau die eine auf, deren Kopf kein Datum trug. Die
	// andere stand vier Zeilen unter einem — und war dieselbe falsche Zahl.
	kommentarstrich := slices.ContainsFunc(sprachenMitStrich,
		func(endung string) bool { return strings.HasSuffix(pfad, endung) })

	var (
		alle    []absatz
		laufend []string
		beginn  int
	)

	schliessen := func() {
		if len(laufend) > 0 {
			alle = append(alle, absatz{text: strings.Join(laufend, " "), zeile: beginn})
			laufend = nil
		}
	}

	for nummer, zeile := range strings.Split(string(roh), "\n") {
		nackt := strings.TrimSpace(zeile)
		if kommentarstrich {
			nackt = strings.TrimSpace(strings.TrimPrefix(nackt, "//"))
		}

		if nackt == "" {
			schliessen()

			continue
		}

		if len(laufend) == 0 {
			beginn = nummer + 1
		}

		laufend = append(laufend, nackt)
	}

	schliessen()

	return alle
}

// nichtGelesen sind Verzeichnisse, die kein Fließtext dieses Projekts sind.
//
// **`.github` steht hier nicht.** Es trägt zwei verzeichnete `.md`-Dateien —
// die Vorlagen für Issue und Pull Request —, und eine Zahl darin veraltet wie
// jede andere.
var nichtGelesen = []string{
	".git", ".go", "node_modules", "dist", "wailsjs", "bin", "coverage",
}

// nichtGelesenNamentlich sind einzelne Dateien, die nicht zum Repositorium
// gehören.
//
// **`.audit-codex.md`** steht in `.gitignore` und ist der Bericht eines
// fremden Zweitblicks. Ein Wächter, der ihn läse, forderte eine Korrektur an
// einem Dokument, das niemand hier geschrieben hat — und das bei jedem
// Zweitblick neu.
var nichtGelesenNamentlich = []string{".audit-codex.md"}

// geleseneEndungen sind die Dateiarten, in denen eine Flächenzahl stehen kann.
//
// **`.mjs`, `.ts` und `.tsx` sind am 2026-08-31 dazugekommen**, und sie haben
// beim ersten Lauf sofort zwei Fundstellen gebracht: `scripts/nahtlauf.mjs` und
// `scripts/nahtvertrag.mjs` nannten je „142 Bindungen", gezählt waren 163.
//
// **Das ist derselbe Befund wie der, für den es diesen Wächter gibt** — eine
// Zahl in der Gegenwartsform, über die nie jemand hinweglief. Er stand hier
// selbst: die Aufzählung der gelesenen Endungen war zwei Bedingungen in einem
// `if`, und was nicht `.md` oder `.go` heißt, lag außerhalb, ohne dass es
// jemand gesagt hätte.
//
// **Was hier nicht steht**: `.mjs` und `.tsx` sind kein Fließtext. Sie werden
// gelesen, weil ihre **Kommentare** einer sind — und ein Kommentar veraltet
// wie ein Absatz in `README.md`.
var geleseneEndungen = []string{".md", ".go", ".mjs", ".ts", ".tsx"}

// sprachenMitStrich sind die davon, deren Kommentar mit `//` beginnt.
//
// **Alle außer `.md`** — und die Aufzählung steht trotzdem da, statt „alles,
// was nicht Markdown ist" zu heißen. Käme eine Endung dazu, deren Kommentar
// anders aussieht, träfe eine Regel über die Klasse sie still mit.
var sprachenMitStrich = []string{".go", ".mjs", ".ts", ".tsx"}

// fliesstextdateien gibt jede Datei mit einer Endung aus `geleseneEndungen`,
// relativ zur Wurzel.
func fliesstextdateien(t *testing.T) []string {
	t.Helper()

	var pfade []string

	fehler := filepath.WalkDir(wurzel, func(pfad string, eintrag os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if eintrag.IsDir() {
			if pfad != wurzel && slices.Contains(nichtGelesen, eintrag.Name()) {
				return filepath.SkipDir
			}

			return nil
		}

		// **Ein Verweis ist kein Fließtext** — in einem Worktree zeigt `.go` auf
		// den Cache des Hauptcheckouts (2026-09-09), und ein `ReadFile` darauf
		// las ein Verzeichnis.
		if eintrag.Type()&os.ModeSymlink != 0 {
			return nil
		}

		// **Eine Datei bleibt nur mit Namen liegen, nicht mit einer Klasse.**
		//
		// Hier stand „alles mit einem Punkt davor" — und damit lag `.github/`
		// mit seinen zwei verzeichneten `.md`-Dateien außerhalb, ohne dass es
		// jemand gesagt hätte. Eine Ausnahme, die eine Klasse trifft, trifft
		// beim nächsten Mal etwas, an das niemand gedacht hat (Zweitblick über
		// Phase 4.0, Befund 2).
		if slices.Contains(nichtGelesenNamentlich, eintrag.Name()) {
			return nil
		}

		if name := eintrag.Name(); slices.ContainsFunc(geleseneEndungen,
			func(endung string) bool { return strings.HasSuffix(name, endung) }) {
			kurz, schnittfehler := filepath.Rel(wurzel, pfad)
			if schnittfehler != nil {
				return schnittfehler
			}

			pfade = append(pfade, filepath.ToSlash(kurz))
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("den Baum durchgehen: %v", fehler)
	}

	if len(pfade) < 100 {
		t.Fatalf("nur %d Dateien gefunden — der Leser greift daneben", len(pfade))
	}

	return pfade
}
