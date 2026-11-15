// **Eine Farbklasse der Gestaltungssprache ist richtig zusammengesetzt** —
// Werkstück J von Phase 4.10
// (#1325).
//
// # Der Anlass
//
// Die Umstellung auf Pergament und Tinte hat aus `bg-marke text-marketext` an
// zwei Stellen `bg-marketext-marketext` gemacht — eine Klasse, die es nicht
// gibt. Tailwind erzeugt dafür **keine** Regel und **keine** Meldung; der
// Hauptknopf stand ohne Farbe da, in beiden Erscheinungsbildern, und keine der
// 2.253 Frontend-Prüfungen sah es.
//
// # Was er prüft
//
// Jede Klasse, deren Name mit einem `--color-*`-Merkmal aus `stil.css`
// **beginnt**, muss ihm **gleich** sein. `bg-marketext-marketext` beginnt mit
// `marketext` und ist es nicht — Befund.
//
// # Was er nicht prüft, und das ist die Hälfte
//
//   - **Keinen Namen, der mit keinem Merkmal beginnt.** `bg-perrgament` fällt
//     durch. Die vollständige Prüfung wäre ein Abgleich gegen das **gebaute**
//     CSS, und die braucht einen Frontend-Bau, den ein Go-Test nicht auslösen
//     soll. Die Lücke steht hier, statt dass sie niemand kennt (§10).
//   - **Nicht, ob die Farbe die richtige ist.** Er prüft die Gestalt.
//   - **Nicht Tailwinds eigene Paletten.** `bg-slate-100` malt, ob es hier
//     stehen *soll*, ist eine Frage an
//     [0065](../../docs/entscheidungen/0065-wie-diese-anwendung-aussieht.md).
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/werkzeuge"
)

// farbklassen findet jede Klasse mit einem Farbpräfix.
var farbklassen = regexp.MustCompile(
	`(?:bg|text|border|divide|outline|ring|fill|stroke|decoration)-` +
		`([a-z][a-z0-9-]*)`)

// farbmerkmale liest die Namen aus dem `@theme`-Block von `stil.css`.
var farbmerkmale = regexp.MustCompile(`--color-([a-z]+):`)

func TestJedeFarbklasseIstRichtigZusammengesetzt(t *testing.T) {
	t.Parallel()

	merkmale := farbmerkmale.FindAllStringSubmatch(werkzeuge.StilblattMitEinbindungen(t), -1)
	if len(merkmale) == 0 {
		t.Fatal("kein --color-Merkmal in stil.css — der Wächter liefe ins Leere")
	}

	namen := make([]string, 0, len(merkmale))

	for _, treffer := range merkmale {
		namen = append(namen, treffer[1])
	}

	// **Die längsten zuerst**, sonst schluckt `text` das `marketext` nicht,
	// aber `marke` schon — und `bg-marketext` sähe wie ein Befund aus.
	sort.Slice(namen, func(a, b int) bool {
		return len(namen[a]) > len(namen[b])
	})

	befunde := map[string][]string{}
	wurzel := filepath.Join("..", "..", "frontend", "src")

	if err := filepath.WalkDir(wurzel, func(pfad string, eintrag os.DirEntry,
		err error,
	) error {
		if err != nil {
			return err
		}

		if eintrag.IsDir() || !strings.HasSuffix(pfad, ".tsx") {
			return nil
		}

		inhalt, err := os.ReadFile(pfad) //nolint:gosec // Pfad aus dem Baum
		if err != nil {
			return err
		}

		for _, klasse := range schiefeFarbklassen(string(inhalt), namen) {
			befunde[klasse] = append(befunde[klasse], filepath.Base(pfad))
		}

		return nil
	}); err != nil {
		t.Fatalf("Baum nicht lesbar: %v", err)
	}

	schief := make([]string, 0, len(befunde))

	for klasse := range befunde {
		schief = append(schief, klasse)
	}

	sort.Strings(schief)

	for _, klasse := range schief {
		t.Errorf(
			"%q beginnt mit einem Merkmal aus stil.css, ist aber keines"+
				" — Tailwind erzeugt dafür keine Regel. Gefunden in: %s",
			klasse, strings.Join(befunde[klasse], ", "))
	}
}

// schiefeFarbklassen gibt die Namen, die mit einem Merkmal beginnen, ohne eines
// zu sein.
//
// **Als eigene Funktion, damit die Reichweite prüfbar ist** — dieselbe
// Begründung wie bei `vordergrundfarbenAus` daneben (§10).
func schiefeFarbklassen(inhalt string, merkmale []string) []string {
	gefunden := []string{}

	for _, treffer := range farbklassen.FindAllStringSubmatch(inhalt, -1) {
		name := treffer[1]

		for _, merkmal := range merkmale {
			if name == merkmal {
				break
			}

			if strings.HasPrefix(name, merkmal+"-") {
				gefunden = append(gefunden, name)

				break
			}
		}
	}

	return gefunden
}
