package domain_test

// **Was ein Batch von Vorschlägen sein darf** — Werkstück A13.1 von Phase 5.3
// (#1452).
//
// # Was hier geprüft wird
//
// **Die Grenzen und nichts sonst.** Wie ein Batch läuft, entscheidet die
// Naht; **wie groß** er sein darf, ist eine fachliche Regel und steht hier.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Batch mit Zielen innerhalb der Grenze geht durch.**
func TestEinBatchInnerhalbDerGrenzeGehtDurch(t *testing.T) {
	t.Parallel()

	ziele := make([]int64, domain.KIBatchHoechstzahl)
	for i := range ziele {
		ziele[i] = int64(i + 1)
	}

	if fehler := domain.PruefeKIBatch(ziele); fehler != nil {
		t.Fatalf("%d Ziele: %v", len(ziele), fehler)
	}
}

// **Ein Batch ohne Ziel ist keiner.**
func TestEinBatchOhneZielIstKeiner(t *testing.T) {
	t.Parallel()

	if fehler := domain.PruefeKIBatch(nil); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("%v, erwartet ErrValidation", fehler)
	}
}

// **Ein Batch über der Grenze wird abgewiesen — und die Meldung nennt sie.**
//
// **Nicht „zu viele"** (§8): wer liest, dass es zu viele sind, weiß nicht,
// wie viele gehen. Die Zahl steht im Satz.
func TestEinBatchUeberDerGrenzeNenntSie(t *testing.T) {
	t.Parallel()

	ziele := make([]int64, domain.KIBatchHoechstzahl+1)
	for i := range ziele {
		ziele[i] = int64(i + 1)
	}

	fehler := domain.PruefeKIBatch(ziele)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("%v, erwartet ErrValidation", fehler)
	}

	if !strings.Contains(fehler.Error(), "24") {
		t.Errorf("die Meldung ist %q — sie nennt die Grenze nicht", fehler)
	}
}

// **Dieselbe Karte zweimal ist ein Fehler und kein Zufall.**
//
// # Warum das abgewiesen wird
//
// **Weil zwei Vorschläge zu **einer** Karte zwei Stände ergäben**, von denen
// der zweite den ersten nicht kennt. Wer beide übernähme, schriebe zweimal
// auf dieselbe Karte — und die zweite Übernahme scheiterte am Stand, den die
// erste gesetzt hat. Ein Batch, der das zulässt, verspricht eine Gruppe und
// liefert einen Konflikt mit sich selbst.
func TestDieselbeKarteZweimalIstEinFehler(t *testing.T) {
	t.Parallel()

	fehler := domain.PruefeKIBatch([]int64{4, 7, 4})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("%v, erwartet ErrValidation", fehler)
	}

	if !strings.Contains(fehler.Error(), "4") {
		t.Errorf("die Meldung ist %q — sie nennt die Karte nicht", fehler)
	}
}

// **Eine Kennung, die es nicht geben kann, ebenfalls.**
func TestEineUnmoeglicheKennungIstEinFehler(t *testing.T) {
	t.Parallel()

	for name, ziele := range map[string][]int64{
		"null":    {4, 0},
		"negativ": {4, -1},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := domain.PruefeKIBatch(ziele); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("%v, erwartet ErrValidation", fehler)
			}
		})
	}
}

// **Die Parallelität ist kleiner als die Höchstzahl.**
//
// **Sonst wäre sie keine** (§10): eine Schranke, die über der Menge liegt,
// die sie schrankt, hält nichts — und der Anbieter bekäme den ganzen Batz auf
// einmal.
func TestDieParallelitaetIstKleinerAlsDieHoechstzahl(t *testing.T) {
	t.Parallel()

	if domain.KIBatchParallel >= domain.KIBatchHoechstzahl {
		t.Errorf("parallel %d, höchstens %d — die Schranke schrankt nichts",
			domain.KIBatchParallel, domain.KIBatchHoechstzahl)
	}

	if domain.KIBatchParallel < 1 {
		t.Errorf("parallel %d — dann läuft gar nichts", domain.KIBatchParallel)
	}
}
