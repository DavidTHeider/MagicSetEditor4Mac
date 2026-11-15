package domain_test

// **Die Ausführung eines Exemplars** — Werkstück `T2` von Phase 5.6
// (#1467).
//
// # Warum aus einem Wahrheitswert eine Angabe wird
//
// **Weil es drei Ausführungen gibt und nicht zwei.** `Foil bool` konnte
// `etched` nicht sagen — und der Preisstand kennt es seit Phase 5.3
// ([domain.Finish]). Ein Bestand, der eine Karte als „foil" führt, weil sie
// nicht „nonfoil" ist, führt sie falsch.
//
// # Und warum `etched` nie geraten wird
//
// **Weil aus `true` nicht hervorgeht, welche der beiden es war.** Die alte
// Wahrheit wird deshalb genau auf `nonfoil` und `foil` abgebildet; wer
// `etched` will, sagt es.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ohne Angabe ist ein Exemplar nonfoil.**
//
// **Der Nullwert ist die alte Vorgabe**: `Foil bool` war `false`, und das
// hieß nonfoil. Wer daraus einen Fehler machte, erklärte jeden Bestand von
// gestern für ungültig.
func TestOhneAngabeIstEinExemplarNonfoil(t *testing.T) {
	t.Parallel()

	if wie := (domain.Exemplar{}).Ausfuehrung(); wie != domain.FinishNonfoil {
		t.Errorf("ohne Angabe ist es %q", wie)
	}

	gesetzt := domain.Exemplar{Finish: domain.FinishEtched}
	if wie := gesetzt.Ausfuehrung(); wie != domain.FinishEtched {
		t.Errorf("gesetzt ist es %q", wie)
	}
}

// **Die alte Wahrheit wird nonfoil und foil — und nie etched.**
func TestDieAlteWahrheitWirdNonfoilUndFoil(t *testing.T) {
	t.Parallel()

	if wie := domain.FinishAus(false); wie != domain.FinishNonfoil {
		t.Errorf("aus false wurde %q", wie)
	}

	if wie := domain.FinishAus(true); wie != domain.FinishFoil {
		t.Errorf("aus true wurde %q", wie)
	}
}

// **Eine Ausführung, die es nicht gibt, ist ein Befund.**
func TestEineUnbekannteAusfuehrungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	exemplar := domain.Exemplar{
		Printing: "abc-123", Menge: 1, Sprache: "de",
		Finish: "glitzernd",
	}

	befunde := exemplar.Validate()
	if len(befunde) != 1 || befunde[0].Field != "finish" {
		t.Fatalf("die Befunde sind %+v", befunde)
	}

	// **Und eine, die es gibt, kommt durch** — sonst prüfte der Lauf nur,
	// dass überhaupt etwas abgewiesen wird (§10).
	for _, wie := range []domain.Finish{
		"", domain.FinishNonfoil, domain.FinishFoil, domain.FinishEtched,
	} {
		gut := exemplar
		gut.Finish = wie

		if befunde := gut.Validate(); len(befunde) != 0 {
			t.Errorf("%q wird abgewiesen: %+v", wie, befunde)
		}
	}
}
