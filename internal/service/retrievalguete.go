package service

// **Wie gut eine Trefferliste ist** — Werkstück 5 von A15, Phase 5.3
// (#1514).
//
// # Drei Maße und nicht eines
//
// **Weil sie drei verschiedene Fragen beantworten.** Das Issue verlangt alle
// drei — „Recall@10, MRR@10, nDCG@10 je Queryklasse/Sprache" —, und die
// Begründung steckt in ihnen:
//
//   - **Recall@10**: wie viel vom Richtigen ist überhaupt dabei? Es sagt
//     nichts über die Reihenfolge.
//   - **MRR@10**: wie weit vorn steht das **erste** Richtige? Es sagt nichts
//     über den Rest.
//   - **nDCG@10**: wie gut ist die **ganze** Ordnung, abgestuft nach
//     Relevanz? Es ist das einzige, das abgestufte Urteile verwertet.
//
// Wer eines davon allein nähme, bekäme eine Zahl, die sich verbessern lässt,
// ohne dass die Suche besser wird.
//
// # Und warum sie hier stehen und nicht im Messlauf
//
// **Weil eine Rechnung, die nur im Messlauf steht, nicht geprüft wird.** Der
// Messlauf braucht eine Laufzeit und einen Korpus und läuft nicht in der
// Pipeline (§3); die Formeln sind reine Funktionen und haben dort nichts
// verloren.

import (
	"math"
	"sort"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Guetegrenze ist das `k` in Recall@k, MRR@k und nDCG@k.
//
// **Zehn**, und die Zahl steht im Issue. Sie ist die Zahl der Treffer, die
// ein Mensch ansieht, bevor er die Frage anders stellt.
const Guetegrenze = 10

// Urteil ist die Relevanz eines Dokuments zu einer Frage.
//
// **Abgestuft und nicht ja/nein**, weil nDCG es verwertet: eine Karte, die
// die Frage genau trifft, ist nicht dasselbe wie eine, die dazu passt.
type Urteil struct {
	Kennung string

	// Stufe ist 0 (nicht relevant), 1 (passend) oder 2 (genau das).
	Stufe int
}

// Guete sind die drei Maße einer Trefferliste.
type Guete struct {
	Recall float64
	MRR    float64
	NDCG   float64

	// Relevante ist, wie viele relevante Dokumente es überhaupt gibt —
	// **die Bezugsgröße**, ohne die ein Recall nichts sagt.
	Relevante int
}

// Bewerte rechnet die drei Maße für eine Trefferliste.
//
// **`treffer` ist die Reihenfolge, in der geantwortet wurde**, und wird auf
// [Guetegrenze] gekürzt. Ein Dokument, über das kein Urteil vorliegt, gilt
// als **nicht relevant** — das ist die übliche Annahme, und sie ist der Grund,
// warum die Urteilsmenge mit offengelegt wird.
func Bewerte(treffer []string, urteile []Urteil) Guete {
	stufen := map[string]int{}
	relevante := 0

	for _, urteil := range urteile {
		if urteil.Stufe > 0 {
			stufen[urteil.Kennung] = urteil.Stufe
			relevante++
		}
	}

	if len(treffer) > Guetegrenze {
		treffer = treffer[:Guetegrenze]
	}

	return Guete{
		Recall:    recall(treffer, stufen, relevante),
		MRR:       reziprokerRang(treffer, stufen),
		NDCG:      ndcg(treffer, stufen),
		Relevante: relevante,
	}
}

// recall ist der Anteil des Relevanten, der in der Liste steht.
//
// **Ohne relevante Dokumente ist er eins und nicht null.** Eine Frage, zu der
// es nichts gibt, ist vollständig beantwortet, wenn nichts kommt — das Issue
// verlangt ausdrücklich **Negativfälle**, und die wären mit einer Null
// systematisch als Fehlschlag gezählt.
func recall(treffer []string, stufen map[string]int, relevante int) float64 {
	if relevante == 0 {
		return 1
	}

	gefunden := 0

	for _, kennung := range treffer {
		if stufen[kennung] > 0 {
			gefunden++
		}
	}

	return float64(gefunden) / float64(relevante)
}

// reziprokerRang ist eins geteilt durch die Stelle des ersten Treffers.
func reziprokerRang(treffer []string, stufen map[string]int) float64 {
	for stelle, kennung := range treffer {
		if stufen[kennung] > 0 {
			return 1 / float64(stelle+1)
		}
	}

	return 0
}

// ndcg ist der abgestufte Gewinn, geteilt durch den bestmöglichen.
//
// **Mit `2^stufe − 1` als Gewinn** und `log2(stelle+1)` als Dämpfung — die
// übliche Form. Sie bestraft einen Treffer weiter hinten, und zwar immer
// schwächer: zwischen Platz 1 und 2 liegt mehr als zwischen 9 und 10.
func ndcg(treffer []string, stufen map[string]int) float64 {
	beste := bestmoeglich(stufen)
	if beste == 0 {
		// **Ohne relevante Dokumente ist die Ordnung perfekt** — dieselbe
		// Überlegung wie beim Recall.
		return 1
	}

	gewinn := 0.0

	for stelle, kennung := range treffer {
		gewinn += gewinnAn(stufen[kennung], stelle)
	}

	return gewinn / beste
}

// bestmoeglich ist der Gewinn der idealen Ordnung.
func bestmoeglich(stufen map[string]int) float64 {
	geordnet := make([]int, 0, len(stufen))
	for _, stufe := range stufen {
		geordnet = append(geordnet, stufe)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(geordnet)))

	if len(geordnet) > Guetegrenze {
		geordnet = geordnet[:Guetegrenze]
	}

	beste := 0.0
	for stelle, stufe := range geordnet {
		beste += gewinnAn(stufe, stelle)
	}

	return beste
}

// gewinnAn ist der gedämpfte Gewinn einer Stufe an einer Stelle.
func gewinnAn(stufe, stelle int) float64 {
	if stufe <= 0 {
		return 0
	}

	return (math.Pow(2, float64(stufe)) - 1) / math.Log2(float64(stelle)+2)
}

// Gueteschnitt mittelt die Maße über mehrere Fragen.
//
// **Ungewichtet**, und das ist eine Entscheidung: eine Frage mit zwanzig
// relevanten Dokumenten zählt genauso viel wie eine mit einem. Wer nach der
// Zahl der Relevanten gewichtete, ließe wenige große Fragen das Ergebnis
// bestimmen — und die Klassen des Issues sind verschieden groß.
func Gueteschnitt(einzelne []Guete) Guete {
	if len(einzelne) == 0 {
		return Guete{}
	}

	schnitt := Guete{}

	for _, guete := range einzelne {
		schnitt.Recall += guete.Recall
		schnitt.MRR += guete.MRR
		schnitt.NDCG += guete.NDCG
		schnitt.Relevante += guete.Relevante
	}

	zahl := float64(len(einzelne))
	schnitt.Recall /= zahl
	schnitt.MRR /= zahl
	schnitt.NDCG /= zahl

	return schnitt
}

// Kennungen zieht die Kennungen aus einer Trefferliste.
//
// **Damit der Messlauf die Ordnung bewerten kann**, ohne die Dokumente
// mitzuschleppen — [Bewerte] braucht nur die Reihenfolge.
func Kennungen(treffer []domain.Retrievaltreffer) []string {
	kennungen := make([]string, 0, len(treffer))
	for _, einzelner := range treffer {
		kennungen = append(kennungen, einzelner.Dokument.Kennung)
	}

	return kennungen
}
