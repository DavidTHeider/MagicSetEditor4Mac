package app

// **Der Wächter über die Fristentabelle** — Werkstück A17 von Phase 5.3
// (#1532).
//
// # Was er hält und was nicht
//
// **Er hält, dass `docs/NETZEIGENSCHAFTEN.md` dieselben Zahlen nennt wie der
// Quelltext.** Die Abnahme des Werkstücks verlangt, dass die Fristen je
// Anbieter „mit ihren Zahlen in der Doku" stehen — und eine Zahl von Hand
// veraltet still, wie §9 es für jede andere sagt.
//
// **Was er nicht hält, ist die Wirkung.** Dass eine Kopffrist überhaupt
// greift, misst `internal/netz/protokollundfristen_intern_test.go` an einer
// schweigenden Gegenstelle — samt Gegenprobe, dass eine langsame Antwort
// **innerhalb** der Frist durchkommt. Und ob zwei bis fünf Minuten für die
// Modelle reichen, die es wirklich gibt, sagt erst ein Lauf gegen einen
// echten Dienst; das steht als Einschränkung neben der Tabelle.

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// fristendoku ist die Datei mit der Tabelle.
const fristendoku = "../../docs/NETZEIGENSCHAFTEN.md"

// **Jede Zeile der Fristentabelle nennt die Zahlen aus dem Quelltext.**
func TestDieFristentabelleNenntDieZahlenAusDemQuelltext(t *testing.T) {
	t.Parallel()

	roh, fehler := os.ReadFile(fristendoku) //nolint:gosec // ein fester Pfad im Repositorium
	if fehler != nil {
		t.Fatalf("%s lesen: %v", fristendoku, fehler)
	}

	for name, zeile := range map[string][3]time.Duration{
		"Ollama":    {kiverbindungsfrist, kikopffrist, kistillstandsfrist},
		"Anthropic": {kiverbindungsfristAnthropic, kikopffristAnthropic, kistillstandsfristAnthropic},
		"OpenAI":    {kiverbindungsfristOpenAI, kikopffristOpenAI, kistillstandsfristOpenAI},
		"xAI":       {kiverbindungsfristXAI, kikopffristXAI, kistillstandsfristXAI},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pruefeFristzeile(t, string(roh), name, zeile)
		})
	}
}

// pruefeFristzeile hält eine Zeile der Tabelle gegen drei Fristen.
//
// **Verglichen werden Dauern und keine Zeichenketten.** „60 s" und „1 min"
// sind dieselbe Frist, und ein Wächter, der das nicht wüsste, verlangte eine
// Schreibweise statt einer Zahl.
func pruefeFristzeile(
	t *testing.T, inhalt, name string, fristen [3]time.Duration,
) {
	t.Helper()

	spalten := tabellenzeileMit(inhalt, name)
	if len(spalten) != len(fristen) {
		t.Fatalf("%s nennt für %q keine Zeile mit drei Fristen: %v",
			fristendoku, name, spalten)
	}

	for i, frist := range fristen {
		gelesen, fehler := alsFrist(spalten[i])
		if fehler != nil {
			t.Errorf("%q in der Zeile %q: %v", spalten[i], name, fehler)

			continue
		}

		if gelesen != frist {
			t.Errorf("%s nennt für %q in Spalte %d %s — im Quelltext steht %s",
				fristendoku, name, i+1, gelesen, frist)
		}
	}
}

// tabellenzeileMit gibt die drei Fristspalten der Zeile zu diesem Namen.
func tabellenzeileMit(inhalt, name string) []string {
	for _, zeile := range strings.Split(inhalt, "\n") {
		if !strings.HasPrefix(zeile, "| ") {
			continue
		}

		spalten := strings.Split(strings.Trim(zeile, "|"), "|")
		if len(spalten) < 4 || !strings.Contains(spalten[0], name) {
			continue
		}

		return []string{spalten[1], spalten[2], spalten[3]}
	}

	return nil
}

// alsFrist liest eine Zelle wie `**5 min**` oder `60 s`.
func alsFrist(zelle string) (time.Duration, error) {
	gekuerzt := strings.TrimSpace(strings.ReplaceAll(zelle, "*", ""))

	var (
		zahl    int
		einheit string
	)

	if _, fehler := fmt.Sscanf(gekuerzt, "%d %s", &zahl, &einheit); fehler != nil {
		return 0, fmt.Errorf("keine Frist: %w", fehler)
	}

	switch einheit {
	case "s":
		return time.Duration(zahl) * time.Second, nil
	case "min":
		return time.Duration(zahl) * time.Minute, nil
	}

	return 0, fmt.Errorf("die Einheit %q kennt dieser Wächter nicht", einheit)
}
