package domain

// **Was ein Batch von KI-Vorschlägen sein darf** — Werkstück A13.1 von
// Phase 5.3 (#1452).
//
// # Warum die Zahlen hier stehen und nicht in der Naht
//
// **Weil sie fachliche Regeln sind und keine Einstellung.** Wie viele
// Anfragen auf einmal hinausgehen dürfen, ist eine Aussage über Höflichkeit
// gegenüber einem fremden Dienst (§16, Regel 4) und über das, was ein Mensch
// in einer Freigabe noch überblickt ([0105](../../docs/entscheidungen/0105-eine-freigabe-fuer-einen-batch.md)).
// Eine Naht, die sie selbst wählte, hätte sie je Fläche anders.
//
// # Und warum sie **neben** der Kostengrenze stehen
//
// **Weil Geld nichts über Anfragen sagt.** Ein Anbieter drosselt nach
// Anfragen; ein Batch über zweihundert Karten wäre auch dann unhöflich, wenn
// er billig bliebe. Die Freigabe begrenzt den Betrag, diese Zahl die Menge —
// und beide gelten.

import (
	"fmt"
	"slices"
)

// KIBatchHoechstzahl ist, wie viele Karten ein Batch trägt.
//
// **Vierundzwanzig**, und die Zahl ist gewählt und nicht gemessen: sie ist
// ungefähr eine Bildschirmseite Karten. Wer mehr will, macht zwei Batches —
// und gibt zweimal frei, was bei zweihundert Zielen die ehrlichere Frage ist.
const KIBatchHoechstzahl = 24

// KIBatchParallel ist, wie viele Anfragen gleichzeitig hinausgehen.
//
// **Vier**, und der Grund ist die Gegenstelle und nicht dieser Rechner: §16,
// Regel 4 verlangt, **unter** der genannten Grenze zu bleiben, und keiner der
// Anbieter nennt eine, die vier gleichzeitige Anfragen ausschlösse.
//
// **Kleiner als [KIBatchHoechstzahl]**, sonst schrankt sie nichts — das hält
// ein Lauf.
const KIBatchParallel = 4

// PruefeKIBatch sagt, ob diese Ziele einen Batch ergeben.
func PruefeKIBatch(zielkarten []int64) error {
	if len(zielkarten) == 0 {
		return fmt.Errorf(
			"%w: ein Batch ohne Zielkarte ist keiner", ErrValidation)
	}

	if len(zielkarten) > KIBatchHoechstzahl {
		return fmt.Errorf(
			"%w: ein Batch trägt höchstens %d Karten, hier stehen %d — "+
				"mehr sind zwei Batches und zwei Freigaben",
			ErrValidation, KIBatchHoechstzahl, len(zielkarten))
	}

	gesehen := make([]int64, 0, len(zielkarten))

	for _, karte := range zielkarten {
		if karte <= 0 {
			return fmt.Errorf("%w: %d ist keine Kartenkennung",
				ErrValidation, karte)
		}

		// **Dieselbe Karte zweimal ist ein Fehler und kein Zufall.** Zwei
		// Vorschläge zu einer Karte ergäben zwei Stände, von denen der zweite
		// den ersten nicht kennt — wer beide übernähme, bekäme beim zweiten
		// Mal den Konflikt, den der erste erzeugt hat.
		if slices.Contains(gesehen, karte) {
			return fmt.Errorf(
				"%w: die Karte %d steht zweimal im Batch — zwei Vorschläge "+
					"zu einer Karte kennen einander nicht",
				ErrValidation, karte)
		}

		gesehen = append(gesehen, karte)
	}

	return nil
}
