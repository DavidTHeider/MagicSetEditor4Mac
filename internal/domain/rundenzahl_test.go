package domain_test

// **Die empfohlene Rundenzahl** — Befund 2 des Zweitblicks über Phase 4.7
// (#1294).
//
// # Der Befund
//
// Der Vermerk sagt eine sichtbare, überschreibbare Empfehlung nach Appendix E
// zu. Im Programm gab es sie nicht — die Zahl „drei Runden bei acht
// Teilnehmern" stand allein als Schleifengrenze in einem Test. **Ein richtiger
// Quellenwert in einer Vorrichtung ist keine Anwendungsfunktion.**
//
// # Woher die Zahlen kommen
//
// **Appendix E der *Magic Tournament Rules***, Fassung vom 2026-02-27,
// abgerufen am 2026-09-01. Die Tabelle steht dort wörtlich; hier sind es
// dieselben Zahlen, als Testfälle.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Staffel aus Appendix E, Zeile für Zeile.**
func TestDieEmpfohleneRundenzahlFolgtAppendixE(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		spieler, runden int
	}{
		// **Bis acht gibt Appendix E keine Swiss-Runden.**
		//
		// Die Spalte „Swiss Rounds" lautet dort für vier Spieler „2
		// Single-Elimination Rounds (No Swiss)" und für fünf bis acht „3
		// Single-Elimination Rounds (No Swiss)" — das sind K.-o.-Runden und
		// ausdrücklich **kein** Swiss. Dieses Programm baut kein K.-o.-System
		// (Vermerk 0062, Verworfen), und eine Zahl aus ihrer Bedeutung zu
		// lösen wäre schlimmer als keine.
		//
		// **Der erste Anlauf las die 3 als Swiss-Runden** und gab sie als
		// Empfehlung aus Appendix E aus — Befund 4 des Nachlaufs über Phase
		// 4.7. Die Zahl stand in der richtigen Spalte und meinte etwas
		// anderes.
		{spieler: 1, runden: 0},
		{spieler: 4, runden: 0},
		{spieler: 5, runden: 0},
		{spieler: 8, runden: 0},

		// **Neun bis sechzehn nennt zwei Zahlen** — vier mit Draft im
		// Playoff, fünf sonst. Dieses Programm kennt kein Playoff und nimmt
		// deshalb die zweite.
		{spieler: 9, runden: 5},
		{spieler: 16, runden: 5},

		{spieler: 17, runden: 5},
		{spieler: 32, runden: 5},
		{spieler: 33, runden: 6},
		{spieler: 64, runden: 6},
		{spieler: 65, runden: 7},
		{spieler: 128, runden: 7},
		{spieler: 129, runden: 8},
		{spieler: 226, runden: 8},
		{spieler: 227, runden: 9},
		{spieler: 409, runden: 9},
		{spieler: 410, runden: 10},
		{spieler: 5000, runden: 10},
	}

	for _, fall := range faelle {
		if ist := domain.EmpfohleneRunden(fall.spieler); ist != fall.runden {
			t.Errorf("%d Spieler → %d Runden, erwartet %d",
				fall.spieler, ist, fall.runden)
		}
	}
}

// **Ein Ausgestiegener zählt nicht mit — solange die erste Runde nicht
// gelegt ist.**
//
// **Das steht so nicht in Appendix E.** Sie spricht von „Players" und kennt
// den Fall nicht; MTR §10.2 hält dagegen fest, dass die **angekündigte**
// Rundenzahl steht. Deshalb gilt: vor der ersten Runde ist die Empfehlung ein
// Vorschlag und ändert sich mit der Anmeldung — danach ist sie angekündigt
// und bleibt.
//
// **Angestrichen von Befund 4 des Nachlaufs über Phase 4.7**: der erste
// Anlauf rechnete auch mitten im Turnier neu, und die Anzeige sprang, sobald
// jemand ausstieg.
func TestEinAusgestiegenerZaehltNichtFuerDieRundenzahl(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(17)
	turnier.Teilnehmer[16].Ausgestiegen = true

	if ist := domain.RundenvorschlagFuer(turnier); ist != 5 {
		t.Errorf("%d Runden, erwartet 5 — sechzehn spielen noch", ist)
	}
}

// **Und ein Turnier ohne Teilnehmer bekommt keine Empfehlung.**
func TestOhneTeilnehmerGibtEsKeineEmpfehlung(t *testing.T) {
	t.Parallel()

	if ist := domain.RundenvorschlagFuer(domain.Turnier{}); ist != 0 {
		t.Errorf("%d Runden, erwartet keine", ist)
	}
}

// **Ist die erste Runde gelegt, steht die Empfehlung.**
//
// MTR §10.2: die angekündigte Rundenzahl wird festgehalten. Eine Anzeige, die
// mitten im Turnier springt, weil jemand aussteigt, ist keine Ankündigung.
func TestNachDemBeginnBleibtDieEmpfehlungStehen(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(17)

	runde(&turnier, spiel(1, 2, 2, 0))

	vorher := domain.RundenvorschlagFuer(turnier)

	turnier.Teilnehmer[16].Ausgestiegen = true
	turnier.Teilnehmer[15].Ausgestiegen = true

	if nachher := domain.RundenvorschlagFuer(turnier); nachher != vorher {
		t.Errorf("die Empfehlung sprang von %d auf %d — nach dem Beginn steht sie",
			vorher, nachher)
	}
}
