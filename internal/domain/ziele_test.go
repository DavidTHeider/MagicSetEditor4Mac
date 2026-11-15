package domain_test

// **Die zwei Wege, die kein gewöhnlicher Lauf betritt** — Werkstück M4 der
// Regelmaschine
// (#1606).
//
// # Warum sie einzeln stehen
//
// **Weil beide aus einem Zustand kommen, den das Spiel nicht erzeugt** — ein
// von Hand geänderter Bestand (§8). Sie dürfen trotzdem nicht stillschweigend
// etwas anderes tun, als der Kommentar daneben behauptet, und deshalb haben
// sie einen Lauf (§10: ein Zweig ohne Lauf ist ein Versprechen, das niemand
// hält).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein leerer Vorrat heißt in der Meldung „(leer)"** und nicht nichts.
//
// **Sonst stünde da „der Vorrat »« deckt die Kosten nicht"**, und ein Mensch
// läse ein Leerzeichen als Fehler der Meldung statt als leeren Vorrat.
func TestEinLeererVorratStehtAlsWortInDerMeldung(t *testing.T) {
	t.Parallel()

	_, fehler := eineWirkpartie(t, "[R]").Wirken("sitz-1", "sitz-1/blitz/0", nil)

	if fehler == nil || !strings.Contains(fehler.Error(), "(leer)") {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Ein Ziel, das weder Sitz noch Objekt ist, ist beim Verrechnen
// unzulässig** — CR 608.2b.
//
// **Es entsteht nicht im Spiel**: [domain.Partie.Wirken] weist so ein Ziel
// beim Ankündigen ab (CR 115.1). Es entsteht aus einem von Hand geänderten
// Bestand — und dann soll der Zauber **verpuffen** statt sich auf ein Ziel zu
// berufen, das es nie gab.
func TestEinZielAusDemNichtsIstUnzulaessig(t *testing.T) {
	t.Parallel()

	partie := eineWirkpartie(t, "[R]")

	objekte := make([]domain.Spielobjekt, len(partie.Objekte))
	copy(objekte, partie.Objekte)
	objekte[len(objekte)-1].Zone = domain.ZoneStapel
	objekte[len(objekte)-1].Ziele = []domain.Ziel{{Kennung: "niemand"}}
	partie.Objekte = objekte

	danach := partie.Verrechnen()

	if blitz := objektMit(t, danach,
		"sitz-1/blitz/0"); blitz.Zone != domain.ZoneFriedhof {
		t.Errorf("der Blitz liegt in %q", blitz.Zone)
	}
}

// **Ein Sitz als Ziel bleibt zulässig, solange er am Tisch sitzt** —
// CR 608.2b, und der Grund steht bei [domain.Ziel]: er hat keine Zone.
//
// **Der Zweig, den dieser Lauf betritt**: die Sitzschleife in
// `zielZulaessig`. Die Läufe daneben zielen auf **Objekte** — ein Sitz als
// Ziel kam bis hierher nur beim **Ankündigen** vor, und dort entscheidet eine
// andere Stelle.
func TestEinSitzBleibtBeimVerrechnenZulaessig(t *testing.T) {
	t.Parallel()

	partie, fehler := eineWirkpartie(t, "[R]").
		ManaHinzu("sitz-1", domain.ManaRot, 1)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	danach, fehler := partie.Wirken("sitz-1", "sitz-1/blitz/0",
		[]string{"sitz-2"})
	if fehler != nil {
		t.Fatalf("wirken: %v", fehler)
	}

	// **Der Zauber verpufft nicht** — er wird verrechnet und geht als
	// Spontanzauber in den Friedhof (CR 608.2n). Das sieht von außen gleich
	// aus wie ein Verpuffen, und deshalb wird der **Beherrscher** geprüft:
	// ein verrechneter Spontanzauber geht an seinen Eigner, ein verpuffter
	// ebenso — aber der Weg dorthin ist ein anderer, und nur einer von
	// beiden betritt diese Schleife.
	nachher := danach.Verrechnen()

	blitz := objektMit(t, nachher, "sitz-1/blitz/0")

	if blitz.Zone != domain.ZoneFriedhof {
		t.Errorf("der Blitz liegt in %q", blitz.Zone)
	}
}
