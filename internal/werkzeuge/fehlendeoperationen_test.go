package werkzeuge

// **Eine Operation, die als fehlend geführt ist, hat nirgends einen Weg** —
// FMT-01.AK-02 (#1690),
// angestrichen vom Zweitblick über Phase 6.0: die Zellen »fehlend« nannten
// keinen Test, und eine später eingebaute Kopie hieße in der Matrix weiter
// »fehlend«. Dieser Lauf ist ihr ausführbarer Test — **er wird rot, sobald
// sich ihr Zustand ändert**, und dann ist die Matrix nachzutragen.
//
// # Was er hält
//
// Je Zelle eine Frage an die Namen **aller vier Flächen**: die Bindungen aus
// dem Syntaxbaum von `internal/app`, die Werkzeuge aus `internal/mcp`, die
// Befehlswörter aus den `case`-Zweigen von `cmd/mse` und die Wege aus
// `internal/rest`. Ein Name wird in Wörter zerlegt, und er trifft, wenn er das
// **Objekt** und das **Verb** der Zelle trägt und kein Unterobjekt — eine Karte
// aus einem Set zu kopieren ist keine Kopie des Sets. **Die Nutzung der
// `.mse-locale`** fragt anders: sie fehlt, solange niemand außerhalb von
// `internal/msepaket` ihr Sprachmodell liest.
//
// # Was er nicht hält
//
// **Eine Operation unter einem Namen, der weder Objekt noch Verb trägt** —
// ein `Vervielfaeltigen` ginge durch. Die Wörterlisten stehen deshalb breit,
// mit Deutsch und Englisch, und `TestDieFrageTrifftWasSieSoll` hält, dass sie
// die naheliegenden Namen treffen und die heutigen Nachbarn nicht.

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// fehlendeOperation ist eine Zelle »fehlend« und die Frage, die sie stellt.
type fehlendeOperation struct {
	objekt, verb, ausser []string
}

// kopierverben sind die Wörter für eine Kopie.
var kopierverben = []string{"kopieren", "kopie", "kopiere", "duplizieren", "duplikat", "klonen", "klon", "copy", "clone"}

// unterobjekte sind, was man aus einem Set oder Paket kopiert, ohne es selbst
// zu kopieren.
var unterobjekte = []string{
	"karte", "karten", "ablage", "partie", "partien", "symbol", "datei", "dateien",
	"paketdatei", "feld", "felder", "bild",
}

// vorlagenobjekte sind die Wörter, unter denen ein Paket heißen kann — **auch
// »Vorlage«**: `VorlageKopieren` stünde sonst unbemerkt neben der Zeile, die
// es fehlend nennt (Nachlauf zum Zweitblick über Phase 6.0).
var vorlagenobjekte = []string{"paket", "pakete", "vorlage", "vorlagen"}

// fehlendeOperationen sind die Zellen, die die Matrix als »fehlend« führt —
// benannt wie dort, damit ein roter Lauf die Zeile nennt.
var fehlendeOperationen = map[string]fehlendeOperation{
	"mse-set × Kopieren": {
		objekt: []string{"set", "sets"}, verb: kopierverben, ausser: unterobjekte,
	},
	// **»alle sechs Vorlagenarten × Kopieren« ist seit FMT-02 gebaut**
	// (#1691) — `PaketKopieren` und seine Wege an den drei anderen Flächen.
	"alle sechs Vorlagenarten × Speichern unter": {
		objekt: vorlagenobjekte,
		verb:   []string{"speichernunter", "sichernunter", "saveas", "unter"},
		ausser: unterobjekte,
	},
	// **»alle sechs Vorlagenarten × Export« ist seit FMT-02 gebaut** (#1691)
	// — ein `.mse-installer`, wie MSE2 ihn schreibt, an allen vier Flächen.
}

// **Keine Zelle »fehlend« hat an einer Fläche einen Weg.**
func TestFehlendeOperationenHabenNirgendsEinenWeg(t *testing.T) {
	namen := flaechennamen(t)

	for zelle, frage := range fehlendeOperationen {
		for _, name := range namen {
			if frage.trifft(name) {
				t.Errorf("%s ist in FAEHIGKEITEN.md »fehlend«, aber %q sieht aus wie ein Weg dorthin — "+
					"die Zeile gehört nachgetragen", zelle, name)
			}
		}
	}

	// **Die Nutzung der `.mse-locale`**: sie fehlt, solange niemand außerhalb
	// des Lesers das Sprachmodell fragt (`service/feldebene.go` sagt, dass es
	// nicht gefragt wird).
	for _, aufruf := range aufrufeAusserhalb(t, regexp.MustCompile(`(\w+\.Sprachmodell)\(`),
		regexp.MustCompile(`func \(\w+ \*Paket\) Sprachmodell\(`), "internal/msepaket") {
		t.Errorf("mse-locale × Nutzung ist in FAEHIGKEITEN.md »fehlend«, aber %s liest das Sprachmodell — "+
			"die Zeile gehört nachgetragen", aufruf)
	}
}

// trifft sagt, ob ein Name Objekt und Verb der Frage trägt und kein Unterobjekt.
func (f fehlendeOperation) trifft(name string) bool {
	woerter := woerterAus(name)
	hat := func(liste []string) bool {
		return slices.ContainsFunc(woerter, func(wort string) bool { return slices.Contains(liste, wort) })
	}

	return hat(f.objekt) && hat(f.verb) && !hat(f.ausser)
}

// wortgrenze trennt einen Namen an Trennzeichen.
var wortgrenze = regexp.MustCompile(`[^A-Za-zÄÖÜäöüß]+`)

// grossanfang trennt `SetKopieren` in `Set Kopieren`.
var grossanfang = regexp.MustCompile(`([a-zäöüß])([A-ZÄÖÜ])`)

// woerterAus zerlegt einen Namen in kleingeschriebene Wörter — `SetKopieren`,
// `schreiben.set.kopieren`, `set-kopieren` und `POST /sets/{setID}/kopie`
// geben dieselben zwei, `SetSpeichernUnter` gibt `speichernunter` zusätzlich
// zu seinen Teilen.
func woerterAus(name string) []string {
	var woerter []string

	for _, teil := range wortgrenze.Split(name, -1) {
		if teil == "" {
			continue
		}

		getrennt := strings.Fields(grossanfang.ReplaceAllString(teil, "$1 $2"))
		for _, wort := range getrennt {
			woerter = append(woerter, strings.ToLower(wort))
		}

		// **Auch zusammengezogen**: `SpeichernUnter` ist ein Verb, kein
		// Speichern und ein Unter.
		for i := 0; i+1 < len(getrennt); i++ {
			woerter = append(woerter, strings.ToLower(getrennt[i]+getrennt[i+1]))
		}
	}

	return woerter
}

// flaechennamen sind die Namen aller vier Flächen.
func flaechennamen(t *testing.T) []string {
	t.Helper()

	namen := bindungenAusApp(t)
	namen = append(namen, ausQuelltext(t, "internal/mcp", mcpwerkzeug)...)
	namen = append(namen, ausQuelltext(t, "cmd/mse", regexp.MustCompile(`case "([a-z][a-z0-9-]*)"`))...)
	namen = append(namen, ausQuelltext(t, "internal/rest",
		regexp.MustCompile(`weg\.HandleFunc\(\s*"([A-Z]+ [^"]+)"`))...)

	return namen
}

// aufrufeAusserhalb sammelt die Treffer eines Musters in den Quelldateien
// unter `internal` und `cmd`, außer unter `ausser` — und **darf leer
// bleiben**: eine Frage nach etwas, das fehlen soll, findet im guten Fall
// nichts. **Dass es das Gefragte noch gibt**, zeigt `anker` unter `ausser`:
// hieße die Methode anders, bliebe der Lauf sonst still grün.
func aufrufeAusserhalb(t *testing.T, muster, anker *regexp.Regexp, ausser string) []string {
	t.Helper()

	var treffer []string

	verankert := false

	for _, teil := range []string{"internal", "cmd"} {
		fehler := filepath.WalkDir(filepath.Join(wurzel, teil), func(pfad string, eintrag os.DirEntry, fehler error) error {
			if fehler != nil || eintrag.IsDir() || !strings.HasSuffix(pfad, ".go") ||
				strings.HasSuffix(pfad, "_test.go") {
				return fehler
			}

			inhalt, lesefehler := os.ReadFile(pfad) //nolint:gosec // Quelldatei im Repositorium
			if lesefehler != nil {
				return lesefehler
			}

			relativ, _ := filepath.Rel(wurzel, pfad) //nolint:errcheck // pfad liegt unter der Wurzel
			relativ = filepath.ToSlash(relativ)

			if strings.HasPrefix(relativ, ausser+"/") {
				verankert = verankert || anker.Match(inhalt)

				return nil
			}

			for _, einer := range muster.FindAllStringSubmatch(string(inhalt), -1) {
				treffer = append(treffer, relativ+": "+einer[1])
			}

			return nil
		})
		if fehler != nil {
			t.Fatalf("%s durchgehen: %v", teil, fehler)
		}
	}

	if !verankert {
		t.Fatalf("%s steht nicht mehr unter %s — die Frage greift daneben", anker, ausser)
	}

	return treffer
}

// **Die Frage trifft, was sie soll** — die naheliegenden Namen einer
// Operation, die es geben könnte, und nicht die Nachbarn, die es gibt.
// Ohne diesen Lauf wäre ein grünes Ergebnis darüber auch ein Muster, das nie
// etwas trifft.
func TestDieFrageTrifftWasSieSoll(t *testing.T) {
	faelle := map[string]struct {
		zelle  string
		name   string
		trifft bool
	}{
		"eine Bindung":                     {"mse-set × Kopieren", "SetKopieren", true},
		"ein Werkzeug":                     {"mse-set × Kopieren", "schreiben.set.kopieren", true},
		"ein Befehl":                       {"mse-set × Kopieren", "set-duplizieren", true},
		"ein Weg":                          {"mse-set × Kopieren", "POST /sets/{setID}/kopie", true},
		"eine Karte aus einem Set":         {"mse-set × Kopieren", "POST /sets/{setID}/karten/{karteID}/kopie", false},
		"die Ablage eines Sets":            {"mse-set × Kopieren", "POST /sets/{setID}/ablage/kopieren", false},
		"eine Partie":                      {"mse-set × Kopieren", "PartieKopieren", false},
		"das Speichern unter eines Sets":   {"alle sechs Vorlagenarten × Speichern unter", "SetSpeichernUnter", false},
		"das Speichern unter eines Pakets": {"alle sechs Vorlagenarten × Speichern unter", "PaketSpeichernUnter", true},
		"eine Vorlage speichern unter":     {"alle sechs Vorlagenarten × Speichern unter", "VorlageSpeichernUnter", true},
	}

	for name, einer := range faelle {
		t.Run(name, func(t *testing.T) {
			frage, gibt := fehlendeOperationen[einer.zelle]
			if !gibt {
				t.Fatalf("keine Frage für %s", einer.zelle)
			}

			if trifft := frage.trifft(einer.name); trifft != einer.trifft {
				t.Errorf("%s gegen %q: %v, erwartet %v", einer.zelle, einer.name, trifft, einer.trifft)
			}
		})
	}
}
