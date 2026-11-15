package domain_test

// **Die Rangliste eines Turniers** — Werkstück 2 von Phase 4.7
// (#1294),
// nach Vermerk [0062](../../docs/entscheidungen/0062-was-ein-turnier-hier-ist.md).
//
// # Woher die Zahlen kommen
//
// **Aus Appendix C der *Magic Tournament Rules***, Fassung vom 2026-02-27,
// abgerufen am 2026-09-01. Die Beispiele dort sind hier Testfälle — mit
// **ihren** Zahlen, nicht mit erfundenen.
//
// Das ist die Abnahme aus dem Vermerk: eine Tiebreaker-Rechnung, die **fast**
// stimmt, sieht wie eine richtige aus. Bei einer Rangliste ist das dasselbe
// wie ein falsches Ja bei der Formatprüfung.

import (
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// nahe vergleicht zwei Quoten.
//
// **Verglichen wird ungerundet mit einer Schranke**, nicht auf Gleichheit:
// 16/24 ist in Fließkomma nicht 0,667, und der Vermerk sagt ausdrücklich, dass
// **ungerundet gerechnet** wird.
func nahe(t *testing.T, was string, ist, soll float64) {
	t.Helper()

	if math.Abs(ist-soll) > 0.0005 {
		t.Errorf("%s = %.6f, erwartet %.4f", was, ist, soll)
	}
}

// **Die drei Spieler aus Appendix C, Abschnitt *Match-win percentage*.**
//
// Wörtlich dort: „These three players competed in an 8-round tournament,
// although only the first player completed all rounds."
func TestDieMatchquoteFolgtDenTurnierregeln(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name                         string
		siege, nieder, unentschieden int
		byes                         int
		matchpunkte                  int
		runden                       int
		quote                        float64
	}{
		{
			name:  "5-2-1 über acht Runden",
			siege: 5, nieder: 2, unentschieden: 1,
			matchpunkte: 16, runden: 8, quote: 0.667,
		},
		{
			// **Die Untergrenze.** 3/(4·3) = 0,25 — gerechnet wird mit 0,33.
			name:  "1-3-0, dann Ausstieg",
			siege: 1, nieder: 3, unentschieden: 0,
			matchpunkte: 3, runden: 4, quote: 0.33,
		},
		{
			// **Ein Bye zählt als Sieg** und als gespielte Runde.
			name:  "3-2-0 mit Erstrundenbye, dann Ausstieg",
			siege: 2, nieder: 2, unentschieden: 0, byes: 1,
			matchpunkte: 9, runden: 5, quote: 0.60,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			bilanz := domain.Bilanz{
				Siege: fall.siege, Niederlagen: fall.nieder,
				Unentschieden: fall.unentschieden, Byes: fall.byes,
			}

			if ist := bilanz.Matchpunkte(); ist != fall.matchpunkte {
				t.Errorf("Matchpunkte = %d, erwartet %d", ist, fall.matchpunkte)
			}

			if ist := bilanz.Runden(); ist != fall.runden {
				t.Errorf("Runden = %d, erwartet %d", ist, fall.runden)
			}

			nahe(t, "Matchquote", bilanz.Matchquote(), fall.quote)
		})
	}
}

// **Die zwei Spieler aus Appendix C, Abschnitt *Game-win percentage*.**
//
// Wörtlich dort: „These two players competed in a four-round tournament."
func TestDieSpielquoteFolgtDenTurnierregeln(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name string
		// je Runde: gewonnene, verlorene, unentschiedene Spiele
		runden      [][3]int
		spielpunkte int
		spiele      int
		quote       float64
	}{
		{
			name:        "21 Spielpunkte aus zehn Spielen",
			runden:      [][3]int{{2, 0, 0}, {2, 1, 0}, {1, 2, 0}, {2, 0, 0}},
			spielpunkte: 21, spiele: 10, quote: 0.70,
		},
		{
			// **Die Untergrenze wieder.** 9/(3·11) = 0,27 — gerechnet wird
			// mit 0,33.
			name:        "9 Spielpunkte aus elf Spielen",
			runden:      [][3]int{{1, 2, 0}, {1, 2, 0}, {0, 2, 0}, {1, 2, 0}},
			spielpunkte: 9, spiele: 11, quote: 0.33,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			var bilanz domain.Bilanz

			for _, runde := range fall.runden {
				bilanz.Spiele(runde[0], runde[1], runde[2])
			}

			if ist := bilanz.Spielpunkte(); ist != fall.spielpunkte {
				t.Errorf("Spielpunkte = %d, erwartet %d", ist, fall.spielpunkte)
			}

			if ist := bilanz.Spielzahl(); ist != fall.spiele {
				t.Errorf("Spielzahl = %d, erwartet %d", ist, fall.spiele)
			}

			nahe(t, "Spielquote", bilanz.Spielquote(), fall.quote)
		})
	}
}

// **Ein Bye zählt als 2–0** — drei Matchpunkte und sechs Spielpunkte.
//
// Wörtlich aus Appendix C: „When a player is assigned a bye for a round, they
// are considered to have won the match 2–0. Thus, that player earns 3 match
// points and 6 game points."
func TestEinByeZaehltAlsZweiNull(t *testing.T) {
	t.Parallel()

	var bilanz domain.Bilanz

	bilanz.Bye()

	if bilanz.Matchpunkte() != 3 || bilanz.Spielpunkte() != 6 {
		t.Errorf("ein Bye gibt %d Matchpunkte und %d Spielpunkte, erwartet 3 und 6",
			bilanz.Matchpunkte(), bilanz.Spielpunkte())
	}

	if bilanz.Runden() != 1 || bilanz.Spielzahl() != 2 {
		t.Errorf("ein Bye zählt als %d Runde und %d Spiele, erwartet 1 und 2",
			bilanz.Runden(), bilanz.Spielzahl())
	}
}

// **Wer nichts gespielt hat, hat keine Quote — und nicht 0,33.**
//
// Appendix C teilt durch die Zahl der gespielten Runden. Bei null Runden gibt
// es nichts zu teilen; die Untergrenze auf ein leeres Feld anzuwenden hieße,
// eine Aussage über jemanden zu machen, über den es keine gibt.
//
// **Das steht so nicht in der MTR** — sie kennt den Fall nicht, weil dort
// jeder spielt. Die Entscheidung ist deshalb eine eigene, und sie steht hier,
// damit niemand sie für nachgelesen hält.
func TestOhneRundeGibtEsKeineQuote(t *testing.T) {
	t.Parallel()

	var leer domain.Bilanz

	if leer.Matchquote() != 0 || leer.Spielquote() != 0 {
		t.Errorf("ohne Runde: Matchquote %v, Spielquote %v — erwartet je 0",
			leer.Matchquote(), leer.Spielquote())
	}
}
