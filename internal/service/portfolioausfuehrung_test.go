// **Welche Ausführung für welches Exemplar zählt** — die Befunde 1 des
// Zweitblicks über Phase 3.4
// (#1165).
//
// # Warum das eine eigene Datei ist
//
// `portfolio_test.go` stand an der Zeilengrenze aus §2, und der Schnitt folgt
// der Frage: dort **was gerechnet wird**, hier **welcher Preis überhaupt
// zählt**. Das ist die Frage, an der der erste Anlauf ein Foil doppelt zählte.
package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Foil mit `foil` und `etched` wird einmal gezählt, nicht zweimal.**
//
// # Warum das ein eigener Lauf ist
//
// Der erste Anlauf nahm jeden passenden Stand — und ein Foil passte auf beide.
// Hat ein Druck beide Preise, wurden Wert **und** Stückzahl verdoppelt. Das ist
// kein erfundener Fall: „Abundant Harvest" führt bei Scryfall beides.
// Angestrichen vom Zweitblick über Phase 3.4.
func TestEinFoilMitZweiVeredelungenZaehltEinmal(t *testing.T) {
	t.Parallel()

	ablage := &preisdoppel{staende: []domain.Preisstand{
		stand("abc", "USD", domain.FinishFoil, 66, 27),
		stand("abc", "USD", domain.FinishEtched, 70, 27),
	}}

	uebersicht, err := uebersichtUeber(t, ablage,
		[]*domain.Exemplar{stueck("abc", 1, true)})
	if err != nil {
		t.Fatalf("Portfolio: %v", err)
	}

	if len(uebersicht.Zeilen) != 1 {
		t.Fatalf("es sind %d Zeilen: %+v", len(uebersicht.Zeilen), uebersicht.Zeilen)
	}

	// **`foil` gewinnt gegen `etched`.** Wer eine Veredelung nicht ausdrücklich
	// führt, meint die gewöhnliche.
	if uebersicht.Zeilen[0].Wert != 66 {
		t.Errorf("die Summe ist %d statt 66", uebersicht.Zeilen[0].Wert)
	}

	if uebersicht.Zeilen[0].Exemplare != 1 {
		t.Errorf("gezählt wurden %d Exemplare", uebersicht.Zeilen[0].Exemplare)
	}
}

// **Und ein Foil ohne `foil`-Preis nimmt den `etched`.**
//
// Sonst stünde ein Exemplar ohne Preis da, obwohl einer vorliegt.
func TestEinFoilOhneFoilpreisNimmtDenEtched(t *testing.T) {
	t.Parallel()

	ablage := &preisdoppel{staende: []domain.Preisstand{
		stand("abc", "USD", domain.FinishEtched, 70, 27),
	}}

	uebersicht, err := uebersichtUeber(t, ablage,
		[]*domain.Exemplar{stueck("abc", 1, true)})
	if err != nil {
		t.Fatalf("Portfolio: %v", err)
	}

	if len(uebersicht.Zeilen) != 1 || uebersicht.Zeilen[0].Wert != 70 {
		t.Errorf("die Übersicht ist %+v", uebersicht)
	}
}

// **`tix` wird nicht summiert.**
//
// # Warum ein Preis abgelegt und gezeigt, aber nicht gerechnet wird
//
// `tix` ist der Preis einer Karte auf **Magic Online** — eine Eintrittskarte
// für einen digitalen Markt. Ein Exemplar ist ein Stück Papier; wer beides
// aufrechnet, führt eine Sammlung, die es nicht gibt.
func TestTixWirdNichtSummiert(t *testing.T) {
	t.Parallel()

	ablage := &preisdoppel{staende: []domain.Preisstand{
		stand("abc", "TIX", domain.FinishNonfoil, 3, 27),
	}}

	uebersicht, err := uebersichtUeber(t, ablage,
		[]*domain.Exemplar{stueck("abc", 1, false)})
	if err != nil {
		t.Fatalf("Portfolio: %v", err)
	}

	if len(uebersicht.Zeilen) != 0 {
		t.Errorf("TIX steht in der Übersicht: %+v", uebersicht.Zeilen)
	}

	// **Und das Exemplar gilt als ohne Preis** — es ist eines, für die
	// Übersicht.
	if uebersicht.OhnePreis != 1 {
		t.Errorf("ohne Preis sind %d", uebersicht.OhnePreis)
	}
}

// **Ein Nonfoil mit zwei Ständen nimmt den einen, der zählt.**
//
// Der Rang ist für ein Nonfoil immer derselbe — es gibt nur eine Ausführung,
// die für es zählt. Ohne diesen Lauf bliebe der Zweig ungeprüft, und die
// Auswahl hinge an der Reihenfolge, in der die Stände kommen.
func TestEinNonfoilNimmtDenNonfoilpreis(t *testing.T) {
	t.Parallel()

	ablage := &preisdoppel{staende: []domain.Preisstand{
		stand("abc", "USD", domain.FinishFoil, 900, 27),
		stand("abc", "USD", domain.FinishNonfoil, 100, 27),
	}}

	uebersicht, err := uebersichtUeber(t, ablage,
		[]*domain.Exemplar{stueck("abc", 1, false)})
	if err != nil {
		t.Fatalf("Portfolio: %v", err)
	}

	if len(uebersicht.Zeilen) != 1 || uebersicht.Zeilen[0].Wert != 100 {
		t.Errorf("die Übersicht ist %+v", uebersicht.Zeilen)
	}
}

// **Eine geätzte Karte nimmt ihren eigenen Preis** — und den `foil`-Preis nur,
// wenn es keinen eigenen gibt.
//
// # Warum das seit `T2` überhaupt eine Frage ist
//
// **Weil es die Ausführung bis dahin nicht gab** (#1467).
// Ein Exemplar war `foil` oder nicht; eine geätzte Karte trug `foil: true` und
// bekam damit den Foilpreis — bei „Abundant Harvest" ein anderer Betrag.
func TestEineGeaetzteKarteNimmtIhrenEigenenPreis(t *testing.T) {
	t.Parallel()

	fuerBeide := []domain.Preisstand{
		stand("abc", "USD", domain.FinishFoil, 66, 27),
		stand("abc", "USD", domain.FinishEtched, 70, 27),
	}

	for name, fall := range map[string]struct {
		staende  []domain.Preisstand
		erwartet int64
	}{
		"beide da":   {fuerBeide, 70},
		"nur foil":   {fuerBeide[:1], 66},
		"nur etched": {fuerBeide[1:], 70},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			uebersicht, err := uebersichtUeber(t,
				&preisdoppel{staende: fall.staende},
				[]*domain.Exemplar{geaetzt("abc", 1)})
			if err != nil {
				t.Fatalf("Portfolio: %v", err)
			}

			if len(uebersicht.Zeilen) != 1 || uebersicht.Zeilen[0].Wert != fall.erwartet {
				t.Errorf("die Übersicht ist %+v", uebersicht)
			}
		})
	}
}

// **Und eine gewöhnliche Karte nimmt keinen veredelten Preis.**
//
// **Die Gegenprobe zur Zeile darüber**: „veredelt nimmt veredelt" darf nicht
// heißen „jeder nimmt jeden". Ohne diesen Lauf hielte auch eine Fassung, die
// gar nicht unterscheidet.
func TestEineGewoehnlicheKarteNimmtKeinenVeredeltenPreis(t *testing.T) {
	t.Parallel()

	uebersicht, err := uebersichtUeber(t, &preisdoppel{staende: []domain.Preisstand{
		stand("abc", "USD", domain.FinishEtched, 70, 27),
	}}, []*domain.Exemplar{stueck("abc", 1, false)})
	if err != nil {
		t.Fatalf("Portfolio: %v", err)
	}

	if len(uebersicht.Zeilen) != 0 || uebersicht.OhnePreis != 1 {
		t.Errorf("die Übersicht ist %+v", uebersicht)
	}
}
