package werkzeuge_test

// **Jede Seite ist durchgegangen** — UX-15.AK-01
// (#1771):
// „Jede registrierte Seite ist durchgegangen; das Ergebnis steht je Seite als
// *geprüft, nichts zu tun* oder als Liste der geänderten Stellen. **Eine Seite
// ohne Eintrag gilt als ungeprüft.**"
//
// # Was er hält
//
// Beide Richtungen, wie beim Entwurf nebenan: jeder Seitenschlüssel aus
// `navigation.ts` steht in [`docs/FLAECHENTEXTE.md`](../../docs/FLAECHENTEXTE.md),
// und jede Zeile dort meint eine Seite, die es gibt.
//
// # Was er **nicht** hält
//
// **Ob die Durchsicht gut war.** Er zählt Zeilen; ob der sichtbare Satz der
// richtige ist und ob hinter dem Fragezeichen etwas steht, das man zum
// Entscheiden braucht, sieht nur ein Mensch (§10). Die **Länge** misst
// `frontend/src/sprache/hinweislaenge.test.ts`.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// flaechenzeile findet "| `start` | … |" in der Durchsicht.
var flaechenzeile = regexp.MustCompile("(?m)^\\| `([a-z]+)` \\|")

func TestJedeSeiteStehtInDerFlaechendurchsicht(t *testing.T) {
	t.Parallel()

	gefuehrt := navigationsseiten(t)

	durchsicht := lies(t, "docs/FLAECHENTEXTE.md")

	geprueft := make([]string, 0, len(gefuehrt))
	for _, treffer := range flaechenzeile.FindAllStringSubmatch(durchsicht, -1) {
		geprueft = append(geprueft, treffer[1])
	}

	// **Sonst wäre dieser Lauf grün, weil er nichts liest** (§10).
	if len(geprueft) < 10 {
		t.Fatalf("docs/FLAECHENTEXTE.md führt %d Zeilen — der Leser greift "+
			"ins Leere", len(geprueft))
	}

	for _, seite := range gefuehrt {
		if !enthaelt(geprueft, seite) {
			t.Errorf("die Seite %q steht nicht in docs/FLAECHENTEXTE.md — "+
				"eine Seite ohne Eintrag gilt als ungeprüft (UX-15.AK-01)",
				seite)
		}
	}

	for _, seite := range geprueft {
		if !enthaelt(gefuehrt, seite) {
			t.Errorf("docs/FLAECHENTEXTE.md führt %q — navigation.ts kennt "+
				"diese Seite nicht", seite)
		}
	}
}

// **Und jede Erläuterung geht durch dieselbe Stelle** — UX-15.AK-03.
//
// # Was er hält
//
// Ein `<p>`, das unmittelbar einen Katalogtext mit `Erklaerung` oder
// `Hinweis` im Namen trägt, ist ein Erläuterungsabsatz und gehört durch
// [Erklaerungsabsatz]. Sonst steht der ganze Text da, und die Grenze aus
// AK-03 gilt für ihn nicht.
//
// # Was er **nicht** hält
//
// **Einen Text, der anders heißt.** Wer eine Erläuterung `satz` nennt, kommt
// hier durch — der Name ist das Merkmal, und ein Wächter über den Sinn eines
// Satzes wäre keiner.
//
// **Und einen Absatz mit `id`.** Ein solcher ist über `aria-describedby` die
// **Beschreibung eines Feldes**; ihn zu teilen hieße, die halbe Beschreibung
// zu verstecken — und genau das verbietet AK-02 („nicht die einzige Quelle
// einer Information, die zum Entscheiden nötig ist"). Vier stehen so im Baum,
// und sie stehen als Ausnahme in `docs/FLAECHENTEXTE.md`.
func TestKeineErlaeuterungStehtRohImAbsatz(t *testing.T) {
	t.Parallel()

	roh := regexp.MustCompile(
		`<p(?:\s+(?:className|ref)="?[^">]*"?)*\s*>` +
			`\s*\{[\w.]*\.\w*(?:Erklaerung|erklaerung|Hinweis)\}\s*</p>`)

	gefunden := []string{}

	wurzel := filepath.Join("..", "..", "frontend", "src")

	if err := filepath.WalkDir(wurzel, func(pfad string, eintrag os.DirEntry,
		err error,
	) error {
		if err != nil {
			return err
		}

		if eintrag.IsDir() || !strings.HasSuffix(pfad, ".tsx") ||
			strings.Contains(pfad, ".test.") {
			return nil
		}

		inhalt, err := os.ReadFile(pfad) //nolint:gosec // Pfad aus dem Baum
		if err != nil {
			return err
		}

		for _, treffer := range roh.FindAllString(string(inhalt), -1) {
			gefunden = append(gefunden,
				pfad+": "+strings.Join(strings.Fields(treffer), " "))
		}

		return nil
	}); err != nil {
		t.Fatalf("frontend/src durchgehen: %v", err)
	}

	if len(gefunden) > 0 {
		t.Errorf("diese Erläuterungen stehen roh in einem Absatz und nicht "+
			"in `Erklaerungsabsatz` — damit steht der ganze Text da "+
			"(UX-15.AK-03):\n%s", strings.Join(gefunden, "\n"))
	}
}
