package domain

// **Wie teuer die wiederholungsfreie Suche wirklich ist** — Befund 5 des
// Zweitblicks über Phase 4.7
// (#1294).
//
// **Ein interner Lauf**, weil er den Zähler braucht. Von außen ist nur zu
// sehen, **dass** eine Paarung herauskommt — nicht, wie viele Schritte sie
// gekostet hat.
//
// # Warum es ihn gibt
//
// Weil neben [Suchschritte] eine Zahl steht, und eine Zahl ohne Messung ist
// eine Behauptung (§10). Der erste Anlauf schrieb dort „gemessen", bevor
// gemessen war.

import (
	"fmt"
	"testing"
)

// **Ein Feld, in dem jeder gegen jeden gespielt hat, findet keinen Ausweg —
// und sucht ihn nicht lange.**
func TestDieSucheGibtSchnellAuf(t *testing.T) {
	t.Parallel()

	for _, wieviele := range []int{4, 8, 12, 20} {
		t.Run(fmt.Sprintf("%d Teilnehmer", wieviele), func(t *testing.T) {
			t.Parallel()

			var (
				offen    = make([]int64, 0, wieviele)
				gespielt = map[[2]int64]bool{}
			)

			for wer := 1; wer <= wieviele; wer++ {
				offen = append(offen, int64(wer))
			}

			// **Jeder gegen jeden** — danach gibt es keine
			// wiederholungsfreie Paarung mehr.
			for einer := range offen {
				for anderer := range offen {
					gespielt[[2]int64{offen[einer], offen[anderer]}] = true
				}
			}

			schritte := 0

			if _, gut := ohneWiederholung(offen, gespielt, &schritte); gut {
				t.Fatal("eine Lösung gefunden, wo es keine gibt")
			}

			t.Logf("%d Teilnehmer: %d Schritte", wieviele, schritte)

			if schritte > Suchschritte {
				t.Errorf("%d Schritte — die Grenze liegt bei %d",
					schritte, Suchschritte)
			}
		})
	}
}

// **Und ein Feld, in dem noch alles offen ist, kostet fast nichts.**
//
// Der Normalfall: die erste Wahl trägt, und die Suche ist so teuer wie die
// kurzsichtige Paarung, die sie ersetzt hat.
func TestDerNormalfallKostetFastNichts(t *testing.T) {
	t.Parallel()

	offen := make([]int64, 0, 64)
	for wer := 1; wer <= 64; wer++ {
		offen = append(offen, int64(wer))
	}

	schritte := 0

	if _, gut := ohneWiederholung(offen, map[[2]int64]bool{}, &schritte); !gut {
		t.Fatal("keine Lösung gefunden, wo jede geht")
	}

	t.Logf("64 Teilnehmer, nichts gespielt: %d Schritte", schritte)

	// **Ein Schritt je Paarung** — mehr braucht es nicht, wenn die erste
	// Wahl trägt.
	if schritte != len(offen)/2 {
		t.Errorf("%d Schritte, erwartet %d", schritte, len(offen)/2)
	}
}

// **Die Grenze greift, wenn die Suche sie erreicht.**
//
// **Über eine gesetzte Grenze und nicht über ein großes Feld**: der Zweig
// hinter [Suchschritte] ist im Betrieb nicht erreichbar — 20 Teilnehmer
// kosten 19 Schritte. Ein Lauf, der ihn über ein wirklich großes Feld
// beträte, prüfte die Laufzeit dieses Rechners und nicht die Zusage.
func TestDieGrenzeGreift(t *testing.T) {
	t.Parallel()

	offen := []int64{1, 2, 3, 4}

	// **Nur die erste Wahl ist verbaut** — ohne die Grenze fände die Suche
	// die Lösung im zweiten Schritt.
	gespielt := map[[2]int64]bool{{1, 2}: true, {2, 1}: true}

	// **Schon aufgebraucht.** Damit bricht der erste Schritt ab.
	schritte := Suchschritte

	if _, gut := ohneWiederholung(offen, gespielt, &schritte); gut {
		t.Error("die Suche lief weiter, obwohl die Grenze erreicht war")
	}
}

// **Und der Ausweg wiederholt, wenn nichts anderes bleibt.**
//
// `gegner` gibt dann den Ersten der Übrigen. Ohne diesen Lauf bliebe die
// Zeile unbetreten, und eine Runde ohne Tische wäre die stille Folge.
func TestDerAuswegWiederholt(t *testing.T) {
	t.Parallel()

	gespielt := map[[2]int64]bool{{1, 2}: true, {2, 1}: true}

	if gegen := gegner(1, []int64{2}, gespielt); gegen != 2 {
		t.Errorf("der Ausweg gab %d, erwartet 2", gegen)
	}

	// **Und er überspringt, was schon gespielt hat, solange es geht.**
	//
	// Dieser Zweig wird im Betrieb erreicht, wenn die Suche an ihrer Grenze
	// abbricht: dann gibt es sehr wohl noch freie Gegner, nur hat sie sie
	// nicht mehr gesucht.
	if gegen := gegner(1, []int64{2, 3}, gespielt); gegen != 3 {
		t.Errorf("der Ausweg gab %d, erwartet 3 — 1 gegen 2 hat gespielt", gegen)
	}
}

// **Und die Grenze greift auch in der Suche nach den wenigsten
// Wiederholungen** — aus demselben Grund und mit derselben Gegenprobe.
//
// **Was sie dann gibt, ist das Kurzsichtige.** Die Suche fängt damit an und
// verbessert es; bricht sie sofort ab, bleibt der Anfang stehen. Das ist die
// Zusage, für die es keinen zweiten Ausweg braucht: **eine Antwort gibt es
// immer.**
func TestDieGrenzeGreiftAuchBeiDenWenigsten(t *testing.T) {
	t.Parallel()

	// **Der gemessene Fall aus `paarungsplan.go`**: sechs Teilnehmer in der
	// Reihenfolge 2,1,3,4,5,6, und nur 1–3 ist noch frisch. Das Kurzsichtige
	// gibt drei Wiederholungen, zwei reichen.
	//
	// **Ein Fall, in dem das Kurzsichtige schon richtig liegt, taugt hier
	// nicht**: dann steht die Schranke sofort auf null, `suche` kehrt in der
	// ersten Zeile um, und die Grenze wird gar nicht erst gefragt. Der erste
	// Anlauf dieses Tests war genau so — grün, ohne die Zeile zu betreten.
	offen := []int64{2, 1, 3, 4, 5, 6}

	gespielt := map[[2]int64]bool{}

	for a := int64(1); a <= 6; a++ {
		for b := a + 1; b <= 6; b++ {
			if a == 1 && b == 3 {
				continue
			}

			gespielt[[2]int64{a, b}] = true
			gespielt[[2]int64{b, a}] = true
		}
	}

	// **Schon aufgebraucht.** Damit bricht der erste Schritt ab.
	schritte := Suchschritte
	plan := wenigsteWiederholungen(offen, gespielt, &schritte)

	// **Was dann bleibt, ist das Kurzsichtige** — 2–1, 3–4, 5–6. Das ist die
	// Zusage, für die es keinen zweiten Ausweg braucht: eine Antwort gibt es
	// immer.
	if len(plan) != 3 {
		t.Fatalf("die Suche gab %d Paarungen, erwartet 3", len(plan))
	}

	if gezaehlteWiederholungen(plan, gespielt) != 3 {
		t.Errorf("die Suche gab %d Wiederholungen, erwartet die drei des "+
			"Kurzsichtigen", gezaehlteWiederholungen(plan, gespielt))
	}
}
