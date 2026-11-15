// **Die Übersicht über die Sammlung** — Werkstück 3 von Phase 3.4
// (#1165).
package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// stand baut einen Preisstand für die Übersicht.
func stand(printing, waehrung string, finish domain.Finish, wert int64, tag int,
) domain.Preisstand {
	return domain.Preisstand{
		Printing: printing,
		Quelle:   domain.PreisquelleScryfall,

		// **Markt, Art und Lage gehören seit A3 dazu** (#1442): das Portfolio
		// summiert **Papier** und **Verkaufspreise**, und ein Stand ohne diese
		// Angaben sagt nicht, ob er dazugehört.
		Markt:    domain.MarktPapier,
		Art:      domain.PreisartVerkauf,
		Lage:     domain.PreisGenannt,
		Waehrung: waehrung,
		Finish:   finish,
		Wert:     wert,
		Erhoben:  time.Date(2026, time.August, tag, 12, 0, 0, 0, time.UTC),
	}
}

// stueck baut ein Exemplar.
func stueck(printing string, menge int, foil bool) *domain.Exemplar {
	return &domain.Exemplar{
		Printing: printing, Menge: menge, Finish: domain.FinishAus(foil),
		Zustand: domain.ZustandNearMint, Sprache: "en",
	}
}

// geaetzt baut ein Exemplar in der dritten Ausführung.
//
// **Eigens und nicht über [stueck]**: dessen Wahrheitswert kann sie nicht
// sagen — das ist der ganze Anlass von `T2`.
func geaetzt(printing string, menge int) *domain.Exemplar {
	return &domain.Exemplar{
		Printing: printing, Menge: menge, Finish: domain.FinishEtched,
		Zustand: domain.ZustandNearMint, Sprache: "en",
	}
}

// uebersichtUeber rechnet eine Übersicht über gegebenen Ständen und Exemplaren.
//
// **Die Exemplare gehen in eine Sammlungsablage** und nicht als Argument in den
// Aufruf: „was ist meine Sammlung wert" holt sie sich selbst, damit der
// Anwendungsfall in **einer** Schicht steht.
func uebersichtUeber(
	t *testing.T, ablage *preisdoppel, exemplare []*domain.Exemplar,
) (service.Portfolio, error) {
	t.Helper()

	return service.NeuePreispflege(ablage, &sammlungsdoppel{zeilen: exemplare},
		quellendoppel{}, probeuhr()).Portfolio(t.Context())
}

// **Die Menge zählt: drei Stück sind der dreifache Wert.**
func TestDieMengeZaehlt(t *testing.T) {
	t.Parallel()

	ablage := &preisdoppel{staende: []domain.Preisstand{
		stand("abc", "EUR", domain.FinishNonfoil, 400, 27),
	}}

	uebersicht, err := uebersichtUeber(t, ablage, []*domain.Exemplar{stueck("abc", 3, false)})
	if err != nil {
		t.Fatalf("Portfolio: %v", err)
	}

	if len(uebersicht.Zeilen) != 1 {
		t.Fatalf("es sind %d Zeilen: %+v", len(uebersicht.Zeilen), uebersicht.Zeilen)
	}

	if uebersicht.Zeilen[0].Wert != 1200 {
		t.Errorf("die Summe ist %d statt 1200", uebersicht.Zeilen[0].Wert)
	}

	if uebersicht.Zeilen[0].Exemplare != 3 {
		t.Errorf("gezählt wurden %d Exemplare", uebersicht.Zeilen[0].Exemplare)
	}
}

// **Währungen werden nicht vermischt.**
//
// Euro und Dollar zu addieren ergäbe eine Zahl ohne Einheit; sie umzurechnen
// hieße, einen Kurs zu erfinden.
func TestWaehrungenWerdenNichtVermischt(t *testing.T) {
	t.Parallel()

	ablage := &preisdoppel{staende: []domain.Preisstand{
		stand("abc", "EUR", domain.FinishNonfoil, 400, 27),
		stand("abc", "USD", domain.FinishNonfoil, 500, 27),
	}}

	uebersicht, err := uebersichtUeber(t, ablage, []*domain.Exemplar{stueck("abc", 1, false)})
	if err != nil {
		t.Fatalf("Portfolio: %v", err)
	}

	if len(uebersicht.Zeilen) != 2 {
		t.Fatalf("es sind %d Zeilen", len(uebersicht.Zeilen))
	}

	// **Nach Währung sortiert** — eine Map hat keine Reihenfolge, und eine
	// Übersicht, die jedes Mal anders sortiert, lässt niemanden zweimal
	// dasselbe lesen.
	if uebersicht.Zeilen[0].Waehrung != "EUR" || uebersicht.Zeilen[1].Waehrung != "USD" {
		t.Errorf("die Reihenfolge ist %q, %q",
			uebersicht.Zeilen[0].Waehrung, uebersicht.Zeilen[1].Waehrung)
	}
}

// **Ein Foil wird nicht mit dem Nonfoil-Preis gerechnet.**
//
// Der Unterschied ist oft das Zehnfache; wer ihn übergeht, rechnet eine
// Sammlung systematisch falsch.
func TestEinFoilWirdMitDemFoilpreisGerechnet(t *testing.T) {
	t.Parallel()

	ablage := &preisdoppel{staende: []domain.Preisstand{
		stand("abc", "EUR", domain.FinishNonfoil, 400, 27),
		stand("abc", "EUR", domain.FinishFoil, 4000, 27),
	}}

	for name, fall := range map[string]struct {
		foil     bool
		erwartet int64
	}{
		"nonfoil": {false, 400},
		"foil":    {true, 4000},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			uebersicht, err := uebersichtUeber(t, ablage,
				[]*domain.Exemplar{stueck("abc", 1, fall.foil)})
			if err != nil {
				t.Fatalf("Portfolio: %v", err)
			}

			if uebersicht.Zeilen[0].Wert != fall.erwartet {
				t.Errorf("die Summe ist %d statt %d",
					uebersicht.Zeilen[0].Wert, fall.erwartet)
			}
		})
	}
}

// **`etched` gilt als veredelt** — es ist eine Ausführung, die ein Exemplar
// mit `foil: true` meint.
func TestEtchedGiltAlsVeredelt(t *testing.T) {
	t.Parallel()

	ablage := &preisdoppel{staende: []domain.Preisstand{
		stand("abc", "EUR", domain.FinishEtched, 9000, 27),
	}}

	uebersicht, err := uebersichtUeber(t, ablage, []*domain.Exemplar{stueck("abc", 1, true)})
	if err != nil {
		t.Fatalf("Portfolio: %v", err)
	}

	if len(uebersicht.Zeilen) != 1 || uebersicht.Zeilen[0].Wert != 9000 {
		t.Errorf("die Übersicht ist %+v", uebersicht)
	}
}

// **Ein Exemplar ohne Preis fällt nicht unter den Tisch — es wird gezählt.**
//
// # Warum das eigens dasteht
//
// Eine Sammlung von 300 Karten, von denen 280 keinen Preis haben, hat eine
// Summe, die nichts aussagt. Ohne diese Zahl sähe sie aus wie eine über alle —
// und wer sie liest, hielte seine Sammlung für 20 Karten wert.
func TestExemplareOhnePreisWerdenGezaehlt(t *testing.T) {
	t.Parallel()

	ablage := &preisdoppel{staende: []domain.Preisstand{
		stand("abc", "EUR", domain.FinishNonfoil, 400, 27),
	}}

	uebersicht, err := uebersichtUeber(t, ablage, []*domain.Exemplar{
		stueck("abc", 1, false),
		stueck("ohnepreis", 5, false),
	})
	if err != nil {
		t.Fatalf("Portfolio: %v", err)
	}

	if uebersicht.OhnePreis != 5 {
		t.Errorf("ohne Preis sind %d", uebersicht.OhnePreis)
	}

	if uebersicht.Zeilen[0].Exemplare != 1 {
		t.Errorf("gezählt wurden %d", uebersicht.Zeilen[0].Exemplare)
	}
}

// **Die Summe ist so frisch wie ihr ältester Bestandteil.**
//
// Wer den jüngsten nennte, gäbe sie frischer aus, als sie ist — und §16,
// Satz 5 verlangt das Gegenteil.
func TestDieSummeIstSoFrischWieIhrAeltesterTeil(t *testing.T) {
	t.Parallel()

	ablage := &preisdoppel{staende: []domain.Preisstand{
		stand("abc", "EUR", domain.FinishNonfoil, 400, 27),
		stand("alt", "EUR", domain.FinishNonfoil, 100, 20),
	}}

	uebersicht, err := uebersichtUeber(t, ablage, []*domain.Exemplar{
		stueck("abc", 1, false), stueck("alt", 1, false),
	})
	if err != nil {
		t.Fatalf("Portfolio: %v", err)
	}

	alt := time.Date(2026, time.August, 20, 12, 0, 0, 0, time.UTC)

	if !uebersicht.Zeilen[0].AeltesterPreis.Equal(alt) {
		t.Errorf("der älteste Preis ist %v statt %v",
			uebersicht.Zeilen[0].AeltesterPreis, alt)
	}
}

// **Und die Reihenfolge der Stände ändert daran nichts.**
//
// Der erste Anlauf setzte den Zeitpunkt beim Anlegen der Zeile und verglich
// danach nur noch — richtig war er dann, wenn der älteste zufällig zuerst kam.
func TestDerAeltesteGewinntInJederReihenfolge(t *testing.T) {
	t.Parallel()

	alt := time.Date(2026, time.August, 20, 12, 0, 0, 0, time.UTC)

	for name, reihe := range map[string][]*domain.Exemplar{
		"alt zuerst": {stueck("alt", 1, false), stueck("abc", 1, false)},
		"neu zuerst": {stueck("abc", 1, false), stueck("alt", 1, false)},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ablage := &preisdoppel{staende: []domain.Preisstand{
				stand("abc", "EUR", domain.FinishNonfoil, 400, 27),
				stand("alt", "EUR", domain.FinishNonfoil, 100, 20),
			}}

			uebersicht, err := uebersichtUeber(t, ablage, reihe)
			if err != nil {
				t.Fatalf("Portfolio: %v", err)
			}

			if !uebersicht.Zeilen[0].AeltesterPreis.Equal(alt) {
				t.Errorf("der älteste Preis ist %v", uebersicht.Zeilen[0].AeltesterPreis)
			}
		})
	}
}

// **Eine leere Sammlung gibt eine leere Übersicht und kein `nil`** (§3).
func TestEineLeereSammlungGibtEineLeereUebersicht(t *testing.T) {
	t.Parallel()

	uebersicht, err := uebersichtUeber(t, &preisdoppel{}, []*domain.Exemplar{})
	if err != nil {
		t.Fatalf("Portfolio: %v", err)
	}

	if uebersicht.Zeilen == nil {
		t.Error("die Zeilen sind nil")
	}
}

// **Ein Fehler der Ablage kommt durch.**
func TestEinFehlerBeimPortfolioKommtDurch(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Ablage stolpert")

	_, err := uebersichtUeber(t, &preisdoppel{fehler: gestolpert},
		[]*domain.Exemplar{stueck("abc", 1, false)})

	if !errors.Is(err, gestolpert) {
		t.Errorf("der Fehler ist %v", err)
	}
}

// **Und ein Fehler beim Lesen der Sammlung ebenso.**
//
// Eine Summe über eine Sammlung, die nicht zu lesen war, wäre eine über nichts.
func TestEinFehlerBeimLesenDerSammlungHaeltDieUebersichtAuf(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Sammlung stolpert")

	_, err := service.NeuePreispflege(&preisdoppel{},
		&sammlungsdoppel{fehler: gestolpert}, quellendoppel{}, probeuhr()).
		Portfolio(t.Context())

	if !errors.Is(err, gestolpert) {
		t.Errorf("der Fehler ist %v", err)
	}
}
