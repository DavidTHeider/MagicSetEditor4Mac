package domain_test

// **Der sitzungsgebundene Name einer Bestandszeile** — Werkstück `T3` von
// Phase 5.6 (#1468).
//
// # Wozu es ihn gibt
//
// **Damit ein Angebot auf eine Zeile zeigen kann, ohne sie zu benennen.** Ein
// Gegenüber muss sagen können „das dritte Stück aus deiner Liste"; es darf
// dabei nicht erfahren, welche Zeile des Bestands das ist — und es darf zwei
// Sichten aus zwei Sitzungen nicht aneinanderlegen können.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Dieselbe Sitzung, dieselbe Zeile, derselbe Name.**
//
// **Ohne das trüge jede Antwort neue Namen**, und ein Angebot zeigte auf
// nichts mehr.
func TestDieselbeSitzungGibtDenselbenNamen(t *testing.T) {
	t.Parallel()

	if erste, zweite := domain.Sitzungsreferenz("s-1", 4711),
		domain.Sitzungsreferenz("s-1", 4711); erste != zweite {
		t.Errorf("%q und %q", erste, zweite)
	}
}

// **Zwei Sitzungen geben zwei Namen — und zwei Zeilen auch.**
func TestZweiSitzungenGebenZweiNamen(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string][2]string{
		"zwei Sitzungen": {
			domain.Sitzungsreferenz("s-1", 4711),
			domain.Sitzungsreferenz("s-2", 4711),
		},
		"zwei Zeilen": {
			domain.Sitzungsreferenz("s-1", 4711),
			domain.Sitzungsreferenz("s-1", 4712),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fall[0] == fall[1] {
				t.Errorf("beide heißen %q", fall[0])
			}
		})
	}
}

// **Der Name trägt die Kennung nicht — auch nicht verkleidet.**
//
// **Gemessen und nicht behauptet**: aus einer Kennung, die als Zahl oder in
// einer gewöhnlichen Kodierung darin stünde, ließe sich der Bestand zählen.
func TestDerNameTraegtDieKennungNicht(t *testing.T) {
	t.Parallel()

	name := domain.Sitzungsreferenz("s-1", 4711)

	for _, verkleidung := range []string{"4711", "MTcxMQ", "1267"} {
		if strings.Contains(name, verkleidung) {
			t.Errorf("%q steht in %q", verkleidung, name)
		}
	}

	// **Und er ist kurz genug, um lesbar zu bleiben, und lang genug, um
	// nicht geraten zu werden.**
	if len(name) < 16 || len(name) > 32 {
		t.Errorf("der Name ist %d Zeichen lang: %q", len(name), name)
	}
}

// **Aus dem Namen kommt die Zeile zurück — innerhalb der Sitzung.**
//
// **Über ein Nachrechnen und keine Tabelle.** Eine Tabelle müsste irgendwo
// liegen, irgendwann aufgeräumt werden und überlebte einen Neustart falsch;
// nachrechnen kostet nichts und kennt keinen halben Stand.
func TestAusDemNamenKommtDieZeileZurueck(t *testing.T) {
	t.Parallel()

	exemplare := []*domain.Exemplar{
		{ID: 1, Printing: "abc", Menge: 1},
		{ID: 4711, Printing: "def", Menge: 2},
	}

	gefunden, gibt := domain.ZeileZuReferenz(
		"s-1", domain.Sitzungsreferenz("s-1", 4711), exemplare)
	if !gibt || gefunden.ID != 4711 {
		t.Errorf("gefunden wurde %+v (%v)", gefunden, gibt)
	}

	// **Aus einer anderen Sitzung kommt nichts.** Sonst wäre der Name doch
	// über Sitzungen hinweg brauchbar.
	if _, gibt := domain.ZeileZuReferenz(
		"s-2", domain.Sitzungsreferenz("s-1", 4711), exemplare); gibt {
		t.Error("der Name galt in einer fremden Sitzung")
	}

	if _, gibt := domain.ZeileZuReferenz("s-1", "erfunden", exemplare); gibt {
		t.Error("ein erfundener Name fand eine Zeile")
	}
}
