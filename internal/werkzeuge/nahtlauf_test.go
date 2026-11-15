// **Die Läufe außerhalb der Pipeline bleiben verdrahtet** —
// #1288 und
// #1290.
//
// # Warum es diesen Wächter gibt
//
// `scripts/nahtlauf.mjs` läuft **nicht** in `make ci` — er baut die Anwendung
// und öffnet ein Fenster. Was nicht in der Pipeline läuft, läuft nur, wenn
// jemand es tut; und wer es tun soll, erfährt es aus `docs/RELEASE.md`. Genau
// derselbe Satz steht beim Kanonlauf: *er gehört auf diese Liste, weil ihn
// sonst niemand fährt.*
//
// Ein Skript, das aus dieser Liste fällt, ist ein Skript, das niemand mehr
// fährt — und es fällt nicht mit einem Knall heraus, sondern beim Umschreiben
// eines Absatzes.
//
// # Was er hält
//
//   - `docs/RELEASE.md` nennt den Aufruf.
//   - Der Vertrag führt mindestens acht Antwortformen — eine Liste, die auf
//     zwei zusammenschrumpft, ist keine Vertragsmatrix mehr.
//   - Jede Prüfungsart, die der Vertrag benutzt, kennt der Läufer auch.
//
// # Was er nicht hält
//
// **Ob der Lauf grün ist.** Dafür müsste er die Anwendung bauen, und dann wäre
// er der Lauf. Was er hält, ist die Verdrahtung.
//
// **Und nicht, ob eine Form das Richtige prüft.** Dass `istNull` auf
// `fehler` die Zusage „kein Fehler" ist und nicht ein Versehen, sagt der Name
// des Falls — den liest ein Mensch.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// vertragsfall findet den Namen eines Falls: `name: "…"`.
var vertragsfall = regexp.MustCompile(`(?m)^\s+name: "([^"]+)"`)

// pruefungsart findet die Arten der mehrzeiligen Gestalt.
var pruefungsart = regexp.MustCompile(`(?m)^ {6}(\w+): `)

// einzeiligeArt findet die Arten in `pruefe: { liste: […], leer: […] }`.
var einzeiligeArt = regexp.MustCompile(`[{,] (\w+): `)

// pruefungImLaeufer findet die Arten, die der Läufer kennt.
var pruefungImLaeufer = regexp.MustCompile(`(?m)^ {2}(\w+): \(wert`)

// mindestformen ist die Untergrenze der Vertragsmatrix.
//
// **Acht und nicht elf.** Eine Form darf wegfallen, wenn sie sich als
// dieselbe wie eine andere herausstellt; was nicht passieren darf, ist ein
// stilles Zusammenschrumpfen auf die zwei bequemen.
const mindestformen = 8

func TestDerNahtlaufStehtInDerFreigabeliste(t *testing.T) {
	t.Parallel()

	freigabe := lies(t, "docs/RELEASE.md")

	for _, lauf := range laeufeInScripts(t) {
		if !strings.Contains(freigabe, "node "+lauf) {
			t.Errorf("docs/RELEASE.md nennt `node %s` nicht — ein Lauf außerhalb "+
				"der Pipeline, den keine Liste nennt, wird nicht gefahren", lauf)
		}
	}
}

// laeufeInScripts findet die Skripte, die **gefahren** werden.
//
// # Warum gesucht und nicht aufgezählt
//
// Hier stand eine Liste aus vier Namen. Sie war am 2026-08-30 schon um zwei
// zu kurz: `kuerzellauf.mjs` stand in `docs/RELEASE.md` und **nicht** in ihr,
// und `abzuege.mjs` stand in keinem von beiden — der Lauf, der damals die
// Bildschirmabzüge des README machte (seit 0.6.0 die Hüllenprobe
// `readmebilder`).
//
// Das ist dieselbe Klasse wie bei `messwegemuster` in
// `messschwellen_test.go`, und dort steht auch, warum sie teuer ist: die
// Zusage „jeder Lauf außerhalb der Pipeline steht in der Freigabeliste" hing
// daran, dass jemand an eine Aufzählung denkt (§10).
//
// # Woran ein Lauf zu erkennen ist
//
// **An `process.argv` oder `spawn(`.** Ein Läufer nimmt Argumente entgegen
// oder startet einen Prozess; eine Bibliothek daneben tut beides nicht. Am
// 2026-08-30 trennt das die zwölf Dateien sauber in acht Läufe und vier
// Helfer.
//
// **Was er nicht prüft**: ob der genannte Lauf auch gefahren *wird*. Das ist
// Handarbeit und steht so in `docs/RELEASE.md`.
func laeufeInScripts(t *testing.T) []string {
	t.Helper()

	eintraege, err := os.ReadDir(filepath.Join("..", "..", "scripts"))
	if err != nil {
		t.Fatalf("scripts/ lesen: %v", err)
	}

	laeufe := []string{}

	// **Ein Test, den die Pipeline fährt, ist kein Lauf außerhalb** — die
	// `*.test.sh` unter scripts/ laufen in `make ci` (`skriptpruefungen`).
	// Bis zum 2026-10-11 hielt dieser Wächter jede `.mjs` mit `spawn(` für
	// einen Lauf, auch `gastlauf.test.mjs`, als sein Test einen Kindprozess
	// bekam. Ein `.test.mjs`, das **kein** `.test.sh` nennt, bleibt ein Lauf.
	inDerPipeline := ""
	for _, eintrag := range eintraege {
		if strings.HasSuffix(eintrag.Name(), ".test.sh") {
			inDerPipeline += lies(t, "scripts/"+eintrag.Name())
		}
	}

	for _, eintrag := range eintraege {
		if eintrag.IsDir() || !strings.HasSuffix(eintrag.Name(), ".mjs") {
			continue
		}
		if strings.HasSuffix(eintrag.Name(), ".test.mjs") && strings.Contains(inDerPipeline, "scripts/"+eintrag.Name()) {
			continue
		}

		inhalt := lies(t, "scripts/"+eintrag.Name())
		if strings.Contains(inhalt, "process.argv") || strings.Contains(inhalt, "spawn(") {
			laeufe = append(laeufe, "scripts/"+eintrag.Name())
		}
	}

	if len(laeufe) < 6 {
		t.Fatalf("nur %d Läufer gefunden — die Erkennung greift nicht mehr", len(laeufe))
	}

	return laeufe
}

func TestDieVertragsmatrixIstVollstaendigVerdrahtet(t *testing.T) {
	t.Parallel()

	vertrag := lies(t, "scripts/nahtvertrag.mjs")
	laeufer := lies(t, "scripts/nahtlauf.mjs")

	formen := vertragsfall.FindAllStringSubmatch(vertrag, -1)
	if len(formen) < mindestformen {
		t.Errorf("die Vertragsmatrix führt %d Antwortformen — erwartet sind "+
			"mindestens %d", len(formen), mindestformen)
	}

	bekannt := map[string]bool{}
	for _, treffer := range pruefungImLaeufer.FindAllStringSubmatch(laeufer, -1) {
		bekannt[treffer[1]] = true
	}

	if len(bekannt) == 0 {
		t.Fatal("der Läufer kennt keine Prüfungsart — der Leser greift daneben")
	}

	// **Nur der Abschnitt hinter `pruefe:`**, denn `name:` und `ruf:` sind
	// keine Prüfungsarten.
	for _, art := range benutzteArten(vertrag) {
		if !bekannt[art] {
			t.Errorf("der Vertrag benutzt die Prüfungsart %q — der Läufer kennt "+
				"sie nicht, und der Fall bräche erst, wenn ihn jemand fährt", art)
		}
	}
}

// benutzteArten gibt die Prüfungsarten aus allen `pruefe:`-Blöcken.
//
// **Nur die obersten Schlüssel.** Jede Prüfungsart nimmt eine **flache**
// Angabe (`nahtlauf.mjs` sagt, warum); damit ist „oberster Schlüssel eines
// `pruefe:`-Blocks" dasselbe wie „Prüfungsart", und der Leser braucht keinen
// Parser für JavaScript.
func benutzteArten(vertrag string) []string {
	var arten []string

	for _, block := range strings.Split(vertrag, "pruefe:")[1:] {
		ende := strings.Index(block, "\n  },")
		if ende < 0 {
			ende = len(block)
		}

		// Die mehrzeilige Gestalt: ein Schlüssel je Zeile, sechs Leerzeichen tief.
		for _, treffer := range pruefungsart.FindAllStringSubmatch(block[:ende], -1) {
			arten = append(arten, treffer[1])
		}

		// Die einzeilige: `pruefe: { liste: ["sets"], leer: ["sets"] }`.
		erste := block
		if bruch := strings.Index(erste, "\n"); bruch >= 0 {
			erste = erste[:bruch]
		}

		for _, wort := range einzeiligeArt.FindAllStringSubmatch(erste, -1) {
			arten = append(arten, wort[1])
		}
	}

	return arten
}
