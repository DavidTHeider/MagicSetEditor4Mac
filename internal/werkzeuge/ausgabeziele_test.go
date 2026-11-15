package werkzeuge_test

// **Die Ziele der Ausgabe stehen zweimal da, und gleich** — AP-10
// (#2367).
//
// `domain.Ausgabeziele` prüft, was in `ausgabe.ziele` stehen darf;
// `ausgabeziele` in `frontend/src/ausgabe.ts` ist, was der Assistent anbietet.
// Ein Ziel nur im Fenster wiese der Bestand beim Merken ab — die Seite sagte
// dann bei jeder Wahl, dass sie sich nichts merken kann. Und die Höchstzahl,
// bis zu der das Fenster abschneidet, ist die, die der Bestand annimmt.
//
// # Was er nicht hält
//
// **Ob das gemerkte Ziel einen Neustart übersteht.** Das zeigt der Lauf am
// nativen Fenster; hier stehen zwei Listen gegeneinander.

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

var (
	zielliste  = regexp.MustCompile(`(?s)export const ausgabeziele = \[(.*?)\] as const`)
	zielwort   = regexp.MustCompile(`'([a-z]+)'`)
	zielgrenze = regexp.MustCompile(`(?m)^const hoechstens = (\d+)$`)
)

func TestDieZieleDerAusgabeStehenInGoUndImFensterGleich(t *testing.T) {
	t.Parallel()

	quelle, fehler := os.ReadFile("../../frontend/src/ausgabe.ts")
	if fehler != nil {
		t.Fatal(fehler)
	}

	block := zielliste.FindSubmatch(quelle)
	if block == nil {
		t.Fatal("die Liste `ausgabeziele` steht nicht mehr in ausgabe.ts — der Leser greift daneben")
	}

	woerter := zielwort.FindAllSubmatch(block[1], -1)
	fenster := make([]string, 0, len(woerter))

	for _, wort := range woerter {
		fenster = append(fenster, string(wort[1]))
	}

	if strings.Join(fenster, " ") != strings.Join(domain.Ausgabeziele(), " ") {
		t.Errorf("das Fenster bietet %v an, der Bestand nimmt %v", fenster, domain.Ausgabeziele())
	}
}

func TestDasFensterSchneidetBeiDerGrenzeDesBestandsAb(t *testing.T) {
	t.Parallel()

	quelle, fehler := os.ReadFile("../../frontend/src/haken/useAusgabeziel.ts")
	if fehler != nil {
		t.Fatal(fehler)
	}

	treffer := zielgrenze.FindSubmatch(quelle)
	if treffer == nil {
		t.Fatal("`const hoechstens` steht nicht mehr in useAusgabeziel.ts")
	}

	grenze, fehler := strconv.Atoi(string(treffer[1]))
	if fehler != nil {
		t.Fatal(fehler)
	}

	if grenze != domain.HoechstensGemerkteZiele {
		t.Errorf("das Fenster schneidet bei %d ab, der Bestand nimmt %d", grenze, domain.HoechstensGemerkteZiele)
	}
}
