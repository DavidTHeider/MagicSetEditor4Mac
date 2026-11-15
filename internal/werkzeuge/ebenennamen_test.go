package werkzeuge_test

// **Jede Ebene hat im Fenster einen Namen** — UX-04.AK-02
// (#1676).
//
// # Der Anlass
//
// `Wahl.tsx` schreibt die Ebene über `ebenenNamen[eintrag.ebene] ??
// eintrag.ebene`: fehlt der Name, steht der Schlüssel aus Go roh da. So stand
// unter den vier Einstellungen des Arbeitsplatzes „Ebene arbeitsplatz" —
// `domain.Ebenen()` kennt vier Ebenen, die Kataloge nannten drei. Gefunden am
// 2026-09-24 beim Erklären der Ebenen an der Suche.
//
// # Was er hält
//
// Beide Richtungen, in beiden Sprachen: jede Ebene aus der Domäne hat einen
// Namen, und jeder Name meint eine Ebene, die es gibt.
//
// # Was er **nicht** hält
//
// **Ob der Name trifft.** Dass `geraet` „Gerät" heißt und nicht „Rechner",
// entscheidet ein Mensch.

import (
	"regexp"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

var (
	ebenenblock     = regexp.MustCompile(`(?s)ebenenNamen: \{(.*?)\}`)
	blockschluessel = regexp.MustCompile(`(?m)^\s*(\w+):`)
)

func TestJedeEbeneHatImFensterEinenNamen(t *testing.T) {
	t.Parallel()

	for _, sprache := range []string{"de", "en"} {
		datei := "frontend/src/sprache/" + sprache + "/seiteneinstellungen.ts"

		block := ebenenblock.FindStringSubmatch(lies(t, datei))
		if block == nil {
			t.Fatalf("%s führt keinen Block `ebenenNamen` — der Leser greift "+
				"ins Leere", datei)
		}

		namen := []string{}
		for _, treffer := range blockschluessel.FindAllStringSubmatch(block[1], -1) {
			namen = append(namen, treffer[1])
		}

		for _, ebene := range domain.Ebenen() {
			if !slices.Contains(namen, string(ebene)) {
				t.Errorf("%s: die Ebene %q hat keinen Namen — im Fenster "+
					"stünde der Schlüssel roh da", datei, ebene)
			}
		}

		for _, name := range namen {
			if !slices.Contains(domain.Ebenen(), domain.Ebene(name)) {
				t.Errorf("%s nennt die Ebene %q — die Domäne kennt sie "+
					"nicht", datei, name)
			}
		}
	}
}
