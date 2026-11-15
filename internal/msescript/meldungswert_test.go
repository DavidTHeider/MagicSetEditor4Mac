package msescript_test

// **Jeder Fehler des Auswerters trägt einen Schlüssel aus der Liste — und
// seinen Wert** — QA-03 (#1656).
//
// Sieben Stellen gaben einen **ganzen Satz** als Schlüssel. `satzZu` kannte
// ihn nicht und gab ihn roh zurück — mit `%s` darin und **ohne den Wert**, der
// die Ursache nennt. Im Kanonlauf ohne Schriftwurzel stand deshalb 253-mal
// »hier steht %s«, und welcher Wert dort stand, sagte keine Zeile. Und der
// Satz lag außerhalb der Kataloge des Fensters: im englischen Fenster stand er
// deutsch, und kein Vertragstest sah es.

import (
	"errors"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestJederAuswerterfehlerNenntSeinenWert(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		quelle     string
		schluessel msescript.Meldungsschluessel
		werte      int
	}{
		"eine Bedingung aus Text":    {`if "a" then 1 else 2`, msescript.SchluesselBedingungWahrheit, 1},
		"for each über Text":         {`for each x in "a" do x`, msescript.SchluesselFuerJedenListe, 1},
		"xor mit Text links":         {`"a" xor true`, msescript.SchluesselWahrheitLinks, 2},
		"xor mit Text rechts":        {`true xor "a"`, msescript.SchluesselWahrheitRechts, 2},
		"and mit Text links":         {`"a" and true`, msescript.SchluesselWahrheitLinks, 2},
		"or mit Text rechts":         {`false or "a"`, msescript.SchluesselWahrheitRechts, 2},
		"Argumente an Text gebunden": {`"a"@(x: 1)`, msescript.SchluesselBindenFunktion, 1},
	}

	for name, einer := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := msescript.AuswertenQuelle(t.Context(), einer.quelle, msescript.Bibliothek())

			var skriptfehler *msescript.Fehler
			if !errors.As(fehler, &skriptfehler) {
				t.Fatalf("kein Skriptfehler: %v", fehler)
			}

			if skriptfehler.Schluessel != einer.schluessel || len(skriptfehler.Werte) != einer.werte ||
				strings.Contains(skriptfehler.Meldung, "%s") {
				t.Errorf("Schlüssel %q, Werte %v, Meldung %q — erwartet %q mit %d Werten",
					skriptfehler.Schluessel, skriptfehler.Werte, skriptfehler.Meldung, einer.schluessel, einer.werte)
			}

			if !slices.Contains(msescript.Meldungsschluesselliste(), skriptfehler.Schluessel) {
				t.Errorf("%q steht in keiner Liste — das Fenster fände keinen Satz dafür", skriptfehler.Schluessel)
			}
		})
	}
}

// satzAlsSchluessel trifft einen Fehler, dessen Schlüssel ein Literal ist —
// `a.fehler(knoten, "…")` oder `NeuerFehler(z, s, "…")`.
var satzAlsSchluessel = regexp.MustCompile(
	`(?:\.fehler\(\s*[^,()]+(?:\([^()]*\))?[^,()]*|NeuerFehler\(\s*[^,]+,\s*[^,]+),\s*"`)

// **Und kein Satz steht mehr als Schlüssel da.** Der Typ hilft nicht: ein
// ungetypter Text wird still zum `Meldungsschluessel`. Gelesen wird der
// Quelltext dieses Pakets; die Gegenprobe ist der Stand vor QA-03, an dem er
// sieben Stellen fand.
func TestKeinSatzStehtAlsSchluessel(t *testing.T) {
	t.Parallel()

	dateien, fehler := os.ReadDir(".")
	if fehler != nil {
		t.Fatal(fehler)
	}

	gelesen := 0

	for _, datei := range dateien {
		name := datei.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		roh, lesefehler := os.ReadFile(name) //nolint:gosec // Quelldateien dieses Pakets
		if lesefehler != nil {
			t.Fatal(lesefehler)
		}

		gelesen++

		for _, treffer := range satzAlsSchluessel.FindAllStringIndex(string(roh), -1) {
			zeile := strings.Count(string(roh[:treffer[0]]), "\n") + 1
			t.Errorf("%s:%d übergibt einen Satz als Schlüssel", name, zeile)
		}
	}

	if gelesen < 10 {
		t.Fatalf("nur %d Quelldateien gelesen — der Leser greift daneben", gelesen)
	}
}
