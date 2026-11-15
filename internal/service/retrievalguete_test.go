package service_test

// **Die drei Gütemaße** — Werkstück 5 von A15, Phase 5.3
// (#1514).
//
// # Warum sie gegen Zahlen von Hand geprüft werden
//
// **Weil eine Formel, die gegen sich selbst geprüft wird, immer stimmt.** Die
// erwarteten Werte hier sind ausgerechnet und stehen als Bruch da — wer sie
// nachrechnen will, kann es.

import (
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// nahe sagt, ob zwei Zahlen bis auf ein Millionstel gleich sind.
func nahDran(links, rechts float64) bool { return math.Abs(links-rechts) < 1e-6 }

// **Eine perfekte Ordnung gibt überall eins.**
func TestEinePerfekteOrdnungGibtUeberallEins(t *testing.T) {
	t.Parallel()

	guete := service.Bewerte([]string{"a", "b"}, []service.Urteil{
		{Kennung: "a", Stufe: 2}, {Kennung: "b", Stufe: 2},
	})

	if !nahDran(guete.Recall, 1) || !nahDran(guete.MRR, 1) || !nahDran(guete.NDCG, 1) {
		t.Fatalf("die Güte ist %+v", guete)
	}

	if guete.Relevante != 2 {
		t.Errorf("%d relevante", guete.Relevante)
	}
}

// **Die Maße messen drei verschiedene Dinge.**
//
// # Der Fall, für den es diesen Lauf gibt
//
// **Dieselben Treffer, andere Reihenfolge.** Der Recall bleibt gleich — es
// ist dasselbe Richtige dabei —, MRR und nDCG fallen. Wer nur den Recall
// nähme, sähe zwischen diesen beiden Antworten keinen Unterschied.
func TestDieMasseMessenDreiVerschiedeneDinge(t *testing.T) {
	t.Parallel()

	urteile := []service.Urteil{{Kennung: "gut", Stufe: 2}}

	vorn := service.Bewerte([]string{"gut", "x", "y"}, urteile)
	hinten := service.Bewerte([]string{"x", "y", "gut"}, urteile)

	if !nahDran(vorn.Recall, hinten.Recall) {
		t.Errorf("der Recall unterscheidet sich: %v gegen %v",
			vorn.Recall, hinten.Recall)
	}

	// **MRR: 1/1 gegen 1/3.**
	if !nahDran(vorn.MRR, 1) || !nahDran(hinten.MRR, 1.0/3.0) {
		t.Errorf("MRR %v gegen %v", vorn.MRR, hinten.MRR)
	}

	// **nDCG: 3/log2(2) = 3 durch 3 = 1 gegen 3/log2(4) = 1,5 durch 3 = 0,5.**
	if !nahDran(vorn.NDCG, 1) || !nahDran(hinten.NDCG, 0.5) {
		t.Errorf("nDCG %v gegen %v", vorn.NDCG, hinten.NDCG)
	}
}

// **Eine Frage ohne richtige Antwort ist beantwortet, wenn nichts kommt.**
//
// **Das Issue verlangt Negativfälle**, und mit einer Null wären sie
// systematisch Fehlschläge — eine Suche, die auf eine unbeantwortbare Frage
// nichts findet, hat recht.
func TestEineFrageOhneRichtigeAntwortIstBeantwortet(t *testing.T) {
	t.Parallel()

	for name, treffer := range map[string][]string{
		"gar nichts gefunden": {},
		"etwas Falsches":      {"x", "y"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			guete := service.Bewerte(treffer, nil)

			if !nahDran(guete.Recall, 1) || !nahDran(guete.NDCG, 1) {
				t.Fatalf("die Güte ist %+v", guete)
			}

			// **Der MRR bleibt null** — es gibt keinen ersten Treffer, und
			// eine Eins dort behauptete einen.
			if !nahDran(guete.MRR, 0) {
				t.Errorf("MRR %v", guete.MRR)
			}
		})
	}
}

// **Abgestufte Urteile ordnen anders als binäre.**
//
// **Das ist der Grund für die Stufen**: eine Liste, die das Passende vor das
// Genaue setzt, hat denselben Recall und ein schlechteres nDCG.
func TestAbgestufteUrteileOrdnenAnders(t *testing.T) {
	t.Parallel()

	urteile := []service.Urteil{
		{Kennung: "genau", Stufe: 2}, {Kennung: "passend", Stufe: 1},
	}

	richtig := service.Bewerte([]string{"genau", "passend"}, urteile)
	verdreht := service.Bewerte([]string{"passend", "genau"}, urteile)

	if !nahDran(richtig.NDCG, 1) {
		t.Errorf("die richtige Ordnung gibt %v", richtig.NDCG)
	}

	if verdreht.NDCG >= richtig.NDCG {
		t.Errorf("verdreht %v, richtig %v", verdreht.NDCG, richtig.NDCG)
	}

	// **Und der Recall unterscheidet sie nicht.**
	if !nahDran(richtig.Recall, verdreht.Recall) {
		t.Errorf("Recall %v gegen %v", richtig.Recall, verdreht.Recall)
	}
}

// **Was hinter der Grenze steht, zählt nicht.**
func TestWasHinterDerGrenzeStehtZaehltNicht(t *testing.T) {
	t.Parallel()

	treffer := make([]string, 0, service.Guetegrenze+1)
	for range service.Guetegrenze {
		treffer = append(treffer, "x")
	}

	treffer = append(treffer, "gut")

	guete := service.Bewerte(treffer, []service.Urteil{{Kennung: "gut", Stufe: 2}})

	if !nahDran(guete.Recall, 0) || !nahDran(guete.MRR, 0) || !nahDran(guete.NDCG, 0) {
		t.Fatalf("die Güte ist %+v", guete)
	}
}

// **Mehr relevante Dokumente als Plätze drücken den bestmöglichen Gewinn nicht.**
//
// **Der ideale Vergleichswert läuft über höchstens zehn** — sonst wäre eine
// Frage mit zwanzig richtigen Antworten selbst bei perfekter Ordnung nie bei
// eins, und die Zahlen zweier Fragen ließen sich nicht vergleichen.
func TestMehrRelevanteAlsPlaetzeDrueckenDasIdealNicht(t *testing.T) {
	t.Parallel()

	urteile := make([]service.Urteil, 0, 20)
	treffer := make([]string, 0, 20)

	for i := range 20 {
		kennung := string(rune('a' + i))
		urteile = append(urteile, service.Urteil{Kennung: kennung, Stufe: 2})
		treffer = append(treffer, kennung)
	}

	guete := service.Bewerte(treffer, urteile)

	if !nahDran(guete.NDCG, 1) {
		t.Errorf("nDCG %v", guete.NDCG)
	}

	// **Der Recall ist es nicht** — zehn von zwanzig, und das ist die
	// ehrliche Zahl.
	if !nahDran(guete.Recall, 0.5) {
		t.Errorf("Recall %v", guete.Recall)
	}
}

// **Ein Urteil mit Stufe null zählt nicht als relevant.**
func TestEinUrteilMitStufeNullZaehltNicht(t *testing.T) {
	t.Parallel()

	guete := service.Bewerte([]string{"x"}, []service.Urteil{
		{Kennung: "x", Stufe: 0},
	})

	if guete.Relevante != 0 || !nahDran(guete.MRR, 0) {
		t.Fatalf("die Güte ist %+v", guete)
	}
}

// **Der Schnitt mittelt ungewichtet.**
func TestDerSchnittMitteltUngewichtet(t *testing.T) {
	t.Parallel()

	if leer := service.Gueteschnitt(nil); leer != (service.Guete{}) {
		t.Fatalf("der leere Schnitt ist %+v", leer)
	}

	schnitt := service.Gueteschnitt([]service.Guete{
		{Recall: 1, MRR: 1, NDCG: 1, Relevante: 1},
		{Recall: 0, MRR: 0, NDCG: 0, Relevante: 19},
	})

	// **Eine halbe, obwohl die zweite Frage neunzehnmal so viele relevante
	// Dokumente hat** — genau das ist die Entscheidung.
	if !nahDran(schnitt.Recall, 0.5) || !nahDran(schnitt.MRR, 0.5) ||
		!nahDran(schnitt.NDCG, 0.5) {
		t.Fatalf("der Schnitt ist %+v", schnitt)
	}

	if schnitt.Relevante != 20 {
		t.Errorf("%d relevante", schnitt.Relevante)
	}
}

// **Kennungen zieht die Reihenfolge heraus.**
func TestKennungenZiehenDieReihenfolgeHeraus(t *testing.T) {
	t.Parallel()

	gezogen := service.Kennungen([]domain.Retrievaltreffer{
		{Dokument: domain.Retrievaldokument{Kennung: "erster"}},
		{Dokument: domain.Retrievaldokument{Kennung: "zweiter"}},
	})

	if len(gezogen) != 2 || gezogen[0] != "erster" || gezogen[1] != "zweiter" {
		t.Fatalf("gezogen wurde %v", gezogen)
	}
}
