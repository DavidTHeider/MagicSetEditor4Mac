// **Jedes Fuzzziel steht auf der Freigabeliste** —
// #1289.
//
// # Warum es diesen Wächter gibt
//
// Ein Fuzzziel läuft in `make ci` **nur mit seinem Saatkorpus**. Das ist die
// Regressionsprobe und nicht das Fuzzen; das eigentliche Fuzzen braucht Zeit
// und findet nur statt, wenn jemand es anstößt — und wer es anstoßen soll,
// erfährt den Aufruf aus `docs/RELEASE.md`.
//
// Damit gilt hier derselbe Satz wie beim Nahtlauf und beim Kanonlauf: *was
// nicht in der Pipeline läuft, läuft nur, wenn eine Liste es nennt.* Ein
// siebtes Ziel, das niemand einträgt, ist ein Ziel, das nie gefuzzt wird — und
// es fällt nicht mit einem Knall heraus, sondern dadurch, dass jemand es
// hinzufügt und die Liste vergisst.
//
// # Was er hält
//
//   - Jedes `func FuzzX` unter `internal/` steht mit **seinem** Paket in
//     `docs/RELEASE.md`.
//   - Jeder Aufruf in `docs/RELEASE.md` nennt ein Ziel, das es gibt — im
//     genannten Paket.
//
// # Was er nicht hält
//
// **Ob das Ziel etwas taugt.** Ein `f.Fuzz(func(*testing.T, []byte) {})` ohne
// eine einzige Zusicherung stünde hier genauso grün da. Was ein Fund ist, steht
// im Kopf jedes Ziels und wird von einem Menschen gelesen.
//
// **Und nicht, dass der Lauf gefahren wurde.** Die gemessenen Ausführungen
// stehen in derselben Zeile von `docs/RELEASE.md`, mit Datum — das ist eine
// Angabe und keine Prüfung.
package werkzeuge_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// fuzzziel findet die Deklaration eines Fuzzziels.
var fuzzziel = regexp.MustCompile(`(?m)^func (Fuzz[A-Z]\w*)\(`)

// fuzzaufruf findet einen Aufruf in der Freigabeliste.
//
// `go test ./internal/msepaket -run Fuzz -fuzz FuzzZerteile -fuzztime 2m`
//
// **`[\w/]+` und nicht `\w+`**: der Baumgang unten besucht jedes
// Unterverzeichnis unter `internal/`, und ein Ziel in einem davon hätte einen
// zweistufigen Weg. Der erste Anlauf las ihn nicht und hätte den Aufruf für
// „gibt es nicht" gehalten — angestrichen vom Zweitblick über Phase 4.2.
var fuzzaufruf = regexp.MustCompile(`go test \./(internal/[\w/]+) -run Fuzz -fuzz (Fuzz[A-Z]\w*)`)

// zielort ist die Kennung eines Fuzzziels.
//
// **Paket **und** Name, nicht der Name allein.**
//
// Der erste Anlauf führte beide Seiten als `map[Name]Paket`. Zwei Pakete mit
// demselben Zielnamen — `FuzzZerteile` gibt es in `msepaket`, und nichts
// hindert `msescript` daran, seines auch so zu nennen — überschrieben einander,
// und der Wächter hätte trotz seiner Zwei-Richtungs-Zusage ein vorhandenes,
// ungelistetes Ziel übersehen. Angestrichen vom Zweitblick über Phase 4.2.
type zielort struct {
	paket string
	name  string
}

func (z zielort) String() string { return z.name + " in " + z.paket }

// TestJedesFuzzzielStehtInDerFreigabeliste hält beide Richtungen.
func TestJedesFuzzzielStehtInDerFreigabeliste(t *testing.T) {
	t.Parallel()

	imQuelltext := fuzzzieleImQuelltext(t)
	if len(imQuelltext) == 0 {
		t.Fatal("kein einziges Fuzzziel gefunden — der Leser greift daneben")
	}

	inDerListe := map[zielort]bool{}
	for _, treffer := range fuzzaufruf.FindAllStringSubmatch(lies(t, "docs/RELEASE.md"), -1) {
		inDerListe[zielort{paket: treffer[1], name: treffer[2]}] = true
	}

	for ziel := range imQuelltext {
		if !inDerListe[ziel] {
			t.Errorf("%s steht in keinem Aufruf in docs/RELEASE.md — "+
				"ein Ziel, das keine Liste nennt, wird nie gefuzzt", ziel)
		}
	}

	for ziel := range inDerListe {
		if !imQuelltext[ziel] {
			t.Errorf("docs/RELEASE.md nennt %s — dieses Ziel gibt es dort nicht", ziel)
		}
	}
}

// fuzzzieleImQuelltext gibt jedes Fuzzziel unter `internal/` mit seinem Paket.
func fuzzzieleImQuelltext(t *testing.T) map[zielort]bool {
	t.Helper()

	wurzel := filepath.Join("..", "..", "internal")
	ziele := map[zielort]bool{}

	fehler := filepath.WalkDir(wurzel, func(pfad string, eintrag os.DirEntry, err error) error {
		if err != nil || eintrag.IsDir() || !strings.HasSuffix(pfad, "_test.go") {
			return err
		}

		inhalt, lesefehler := os.ReadFile(pfad) //nolint:gosec // Weg aus dem Baumgang
		if lesefehler != nil {
			return lesefehler
		}

		// **Der Paketname kommt aus dem Weg und nicht aus `package`.** Ein
		// Testpaket heißt `msepaket_test`, der Aufruf lautet aber
		// `./internal/msepaket` — was der Aufruf nennt, ist das Verzeichnis.
		paket := filepath.ToSlash(strings.TrimPrefix(
			filepath.Dir(pfad), filepath.Join("..", "..")+string(filepath.Separator)))

		for _, treffer := range fuzzziel.FindAllStringSubmatch(string(inhalt), -1) {
			ziele[zielort{paket: paket, name: treffer[1]}] = true
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("die Fuzzziele aufzählen: %v", fehler)
	}

	return ziele
}

// TestDieFreigabelisteNenntJedesZielGenauEinmal weist die stille Dopplung ab.
//
// **Zwei Aufrufe für dasselbe Ziel sind kein Schönheitsfehler**: bei einer
// Änderung wird einer davon nachgezogen und der andere nicht, und ab dann sagt
// die Liste zweierlei über denselben Lauf.
func TestDieFreigabelisteNenntJedesZielGenauEinmal(t *testing.T) {
	t.Parallel()

	gezaehlt := map[zielort]int{}
	for _, treffer := range fuzzaufruf.FindAllStringSubmatch(lies(t, "docs/RELEASE.md"), -1) {
		gezaehlt[zielort{paket: treffer[1], name: treffer[2]}]++
	}

	var doppelt []string

	for ziel, zahl := range gezaehlt {
		if zahl > 1 {
			doppelt = append(doppelt, fmt.Sprintf("%s (%d-mal)", ziel, zahl))
		}
	}

	sort.Strings(doppelt)

	for _, eintrag := range doppelt {
		t.Errorf("docs/RELEASE.md nennt %s", eintrag)
	}
}

// registerzeile findet eine Zeile des Registers: `| `FuzzX` | `internal/…` |`.
var registerzeile = regexp.MustCompile("(?m)^\\| `(Fuzz[A-Z]\\w*)` \\| `(internal/[\\w/]+)` \\|")

// TestJedesFuzzzielStehtImRegister hält `docs/FUZZZIELE.md` gegen den
// Quelltext — QA-06 (#1737).
//
// **Beide Richtungen**, wie oben: ein Ziel ohne Zeile hat keine Fläche und
// keine Invariante, an der ein Mensch es messen könnte, und eine Zeile ohne
// Ziel ist eine Deckung, die es nicht gibt. **Was er nicht hält**: ob die
// Fläche richtig zugeordnet und die Invariante die des Ziels ist.
func TestJedesFuzzzielStehtImRegister(t *testing.T) {
	t.Parallel()

	imQuelltext := fuzzzieleImQuelltext(t)
	imRegister := map[zielort]bool{}

	for _, treffer := range registerzeile.FindAllStringSubmatch(lies(t, "docs/FUZZZIELE.md"), -1) {
		imRegister[zielort{paket: treffer[2], name: treffer[1]}] = true
	}

	for ziel := range imQuelltext {
		if !imRegister[ziel] {
			t.Errorf("%s steht nicht in docs/FUZZZIELE.md — ohne Fläche und Invariante", ziel)
		}
	}

	for ziel := range imRegister {
		if !imQuelltext[ziel] {
			t.Errorf("docs/FUZZZIELE.md nennt %s — dieses Ziel gibt es dort nicht", ziel)
		}
	}
}
