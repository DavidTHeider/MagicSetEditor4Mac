package service_test

// **Das Portfolio fragt einmal, nicht fünftausendmal** — Werkstück 1 von
// Phase 4.5 (#1292).
//
// # Der Befund, der diesen Test verlangt
//
// Gemessen am 2026-08-30 an einem Bestand von 5.000 Exemplaren **ohne einen
// einzigen Preis** (`BenchmarkWegPortfolio`):
//
//	die ganze Sammlung laden      5 ms
//	die Sammlung zusammenrechnen 89 ms
//
// Das Achtzehnfache für dieselben Zeilen — und die 89 ms sind der Preis von
// 5.000 **leeren** Abfragen. `Portfolio` rief `Neueste` in der Schleife über
// die Exemplare, einmal je Exemplar.
//
// # Warum das ein Test ist und keine Messung
//
// Eine Messung sagt „89 ms auf diesem Rechner"; sie wird rot, wenn jemand ein
// langsameres Notebook nimmt, und grün, wenn jemand ein schnelleres kauft.
// **Die Zusage ist nicht die Zeit, sondern die Zahl der Abfragen** — und die
// gilt auf jeder Maschine.
//
// Die Messung steht daneben (`internal/app/messsammlung_bench_intern_test.go`)
// und sagt, was es kostet. Dieser Test sagt, dass es nicht **mit dem Bestand
// wächst**.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Bestand wächst, die Zahl der Abfragen nicht.**
//
// **Zwei Größen und nicht eine.** Eine feste Zahl („höchstens 3 Abfragen")
// ließe sich mit einer Konstante erfüllen, die zufällig passt; zwei Bestände
// mit demselben Zähler belegen, dass er **nicht** mitwächst — und genau das
// ist die Zusage.
func TestPortfolioFragtNichtJeExemplar(t *testing.T) {
	for _, groesse := range []int{10, 500} {
		ablage := &preisdoppel{}
		sammlung := &sammlungsdoppel{}

		for zaehler := range groesse {
			if _, err := sammlung.Anlegen(t.Context(), &domain.Exemplar{
				Printing: druckkennung(zaehler),
				Menge:    1,
				Zustand:  domain.ZustandNearMint,
				Sprache:  domain.Vorgabesprache,
			}); err != nil {
				t.Fatalf("Exemplar %d anlegen: %v", zaehler, err)
			}
		}

		pflege := service.NeuePreispflege(ablage, sammlung, quellendoppel{}, probeuhr())

		if _, err := pflege.Portfolio(t.Context()); err != nil {
			t.Fatalf("Portfolio bei %d Exemplaren: %v", groesse, err)
		}

		// **Höchstens eine Abfrage je Aufruf**, gleich wie groß der Bestand
		// ist. Vor der Berichtigung waren es `groesse` viele.
		if ablage.lesezaehler.Load() > 1 {
			t.Errorf("bei %d Exemplaren wurde %d mal gelesen — erwartet höchstens 1",
				groesse, ablage.lesezaehler.Load())
		}
	}
}

// druckkennung gibt eine Kennung in der Gestalt, die Scryfall benutzt.
func druckkennung(zaehler int) string {
	return "00000000-0000-4000-8000-" +
		string(rune('0'+zaehler/10000%10)) +
		string(rune('0'+zaehler/1000%10)) +
		string(rune('0'+zaehler/100%10)) +
		string(rune('0'+zaehler/10%10)) +
		string(rune('0'+zaehler%10)) +
		"0000000"
}
