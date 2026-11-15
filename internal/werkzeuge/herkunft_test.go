// Der Wächter über die Herkunftsangaben portierter Regeln.
//
// `AGENTS.md` §11 verlangt für jede übernommene Regel „Datei und Fassung, im
// Kommentar neben dem Code". Genannt wurden Datei und **Zeile**; eine
// Zeilennummer ohne Fassung zeigt bei der nächsten Änderung an MSE2 auf etwas
// anderes ([0028](../../docs/entscheidungen/0028-die-fassung-in-einer-herkunftsangabe.md)).
//
// # Was er nicht prüft
//
// **Ob der Hash stimmt.** Er vergleicht keinen Kommentar mit dem Klon; er
// verlangt nur, dass eine Fassung dasteht. Der Klon liegt nach §11 ausdrücklich
// **neben** dem Repositorium und damit nicht in der Pipeline.
//
// **Ob die Zeilennummer noch trifft.** Aus demselben Grund.
//
// **Verweise ohne Dateipfad.** „`ScriptValue::toColor` sagt …" ohne
// `src/script/value.cpp` fällt durch — der Wächter sieht den Pfad, nicht den
// Anspruch.
//
// Ein Wächter, der die halbe Zusage hält, macht die andere unsichtbar — also
// wird sie ausgesprochen.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// cppverweis trifft einen Pfad in den Quelltext von MSE2.
var cppverweis = regexp.MustCompile(`src/[a-z_0-9]+/[a-z_0-9/]*[a-z_0-9]+\.(cpp|hpp)`)

// TestJedeHerkunftNenntIhreFassung liest jede Go-Datei unter `internal/`.
func TestJedeHerkunftNenntIhreFassung(t *testing.T) {
	var ohne []string

	if err := filepath.WalkDir("..", func(pfad string, eintrag os.DirEntry, err error) error {
		if err != nil || eintrag.IsDir() || !strings.HasSuffix(pfad, ".go") {
			return err
		}

		roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // ein Pfad aus WalkDir unter ..
		if lesefehler != nil {
			return lesefehler
		}

		for _, block := range kommentarbloecke(string(roh)) {
			if cppverweis.MatchString(block.text) && !strings.Contains(block.text, "Fassung") {
				ohne = append(ohne, pfad+":"+block.stelle+" — "+
					cppverweis.FindString(block.text))
			}
		}

		return nil
	}); err != nil {
		t.Fatalf("gehen: %v", err)
	}

	if len(ohne) > 0 {
		t.Errorf("%d Herkunftsangaben ohne Fassung (AGENTS.md §11, Vermerk 0028):\n  %s",
			len(ohne), strings.Join(ohne, "\n  "))
	}
}

// kommentarblock ist eine Folge zusammenhängender `//`-Zeilen.
type kommentarblock struct {
	text   string
	stelle string
}

// kommentarbloecke schneidet eine Datei in ihre Kommentarblöcke.
//
// **Blockweise und nicht zeilenweise**: die Fassung steht am Kopf eines
// Absatzes und gilt für jede Fundstelle darin — es gibt nur einen Klon.
func kommentarbloecke(inhalt string) []kommentarblock {
	zeilen := strings.Split(inhalt, "\n")
	bloecke := []kommentarblock{}

	for i := 0; i < len(zeilen); {
		if !strings.HasPrefix(strings.TrimSpace(zeilen[i]), "//") {
			i++

			continue
		}

		j := i
		for j < len(zeilen) && strings.HasPrefix(strings.TrimSpace(zeilen[j]), "//") {
			j++
		}

		bloecke = append(bloecke, kommentarblock{
			text:   strings.Join(zeilen[i:j], "\n"),
			stelle: zeilennummer(i + 1),
		})

		i = j
	}

	return bloecke
}

// zeilennummer macht aus einer Zahl ihren Text.
func zeilennummer(n int) string {
	ziffern := ""
	for n > 0 {
		ziffern = string(rune('0'+n%10)) + ziffern
		n /= 10
	}

	if ziffern == "" {
		return "0"
	}

	return ziffern
}

// **Die Gegenprobe** (§10): ein Wächter, der nie rot wird, prüft nichts.
func TestDerHerkunftswaechterWirdRot(t *testing.T) {
	for _, fall := range []struct {
		name   string
		block  string
		trifft bool
	}{
		{
			name:   "Verweis ohne Fassung",
			block:  "// `src/data/font.cpp`, Zeile 27.",
			trifft: true,
		},
		{
			name:   "Verweis mit Fassung im selben Block",
			block:  "// `src/data/font.cpp`, Zeile 27,\n// Fassung f9d9356 vom 2021-07-06.",
			trifft: false,
		},
		{
			name:   "kein Verweis",
			block:  "// Ein gewöhnlicher Kommentar über `font.cpp` ohne Pfad.",
			trifft: false,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			bloecke := kommentarbloecke(fall.block)
			if len(bloecke) != 1 {
				t.Fatalf("%d Blöcke, erwartet einer", len(bloecke))
			}

			rot := cppverweis.MatchString(bloecke[0].text) &&
				!strings.Contains(bloecke[0].text, "Fassung")
			if rot != fall.trifft {
				t.Errorf("rot %t, erwartet %t", rot, fall.trifft)
			}
		})
	}
}
