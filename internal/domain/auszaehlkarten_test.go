package domain_test

// **Welche Karten ein Balken trägt** — UXR-F6
// (#2013).
//
// # Was nachgelesen wurde (§15)
//
// `GraphDataElement` trägt neben den Gruppennummern den `original_index` der
// Karte (`src/gui/control/graph.hpp`), und `GraphData::indices` gibt nach
// einer Wahl genau diese Indizes zurück (`src/gui/control/graph.cpp`, Zeilen
// 249–258, Fassung f9d9356 vom 2021-07-06) — ein Index **einmal**, auch wenn
// mehrere Elemente derselben Karte in der Wahl liegen (`split list:`).
//
// Die Gruppen sind dieselben wie in `auszaehlung_test.go`; neu ist nur, dass
// jede ihre Karten nennt.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// dimensionswerte baut aus einem Wert und Kennungen die Werte einer Dimension:
// `dimensionswerte("Rot", 3, 9)` ist Rot an den Karten 3 und 9.
func dimensionswerte(wert string, karten ...int64) []domain.Dimensionswert {
	werte := make([]domain.Dimensionswert, 0, len(karten))

	for _, karte := range karten {
		werte = append(werte, domain.Dimensionswert{Karte: karte, Wert: wert})
	}

	return werte
}

// **Jede Gruppe nennt ihre Karten — und zählt wie vorher.**
func TestEineGruppeNenntIhreKarten(t *testing.T) {
	t.Parallel()

	gruppen := domain.AuszaehlenKarten(
		slices.Concat(dimensionswerte("Weiss", 9), dimensionswerte("Rot", 3), dimensionswerte("Weiss", 7)),
		domain.Auszaehlart{Gruppen: []string{"Weiss", "Rot"}},
	)

	if len(gruppen) != 2 {
		t.Fatalf("%d Gruppen: %+v", len(gruppen), gruppen)
	}

	// **Aufsteigend**, damit dieselbe Frage dieselbe Antwort gibt — die
	// Reihenfolge der Karten im Set ist keine Aussage über die Gruppe.
	if !slices.Equal(gruppen[0].Kartenkennungen, []int64{7, 9}) || gruppen[0].Karten != 2 {
		t.Errorf("Weiss ist %+v", gruppen[0])
	}

	if !slices.Equal(gruppen[1].Kartenkennungen, []int64{3}) || gruppen[1].Karten != 1 {
		t.Errorf("Rot ist %+v", gruppen[1])
	}
}

// **Eine erklärte Gruppe ohne Karten nennt keine — und bleibt eine Gruppe.**
func TestEineLeereGruppeNenntKeineKarten(t *testing.T) {
	t.Parallel()

	gruppen := domain.AuszaehlenKarten(
		dimensionswerte("Weiss", 1),
		domain.Auszaehlart{Gruppen: []string{"Land", "Weiss"}},
	)

	if len(gruppen) != 2 || gruppen[0].Name != "Land" ||
		len(gruppen[0].Kartenkennungen) != 0 {
		t.Errorf("die Gruppen lauten %+v", gruppen)
	}
}

// **Eine Klasse nennt die Karten aller Werte, die sie fasst.**
//
// `1` und `4` liegen in `1–5` (`to_bin`); die Karten beider stehen dort, und
// die Klasse daneben behält ihre eigene.
func TestEineKlasseNenntDieKartenAllerIhrerWerte(t *testing.T) {
	t.Parallel()

	gruppen := domain.AuszaehlenKarten(
		slices.Concat(dimensionswerte("1", 10), dimensionswerte("4", 11), dimensionswerte("7", 12)),
		domain.Auszaehlart{Numerisch: true, Klassenbreite: 5},
	)

	if len(gruppen) != 3 {
		t.Fatalf("%d Gruppen: %+v", len(gruppen), gruppen)
	}

	if gruppen[1].Name != "1–5" || !slices.Equal(gruppen[1].Kartenkennungen, []int64{10, 11}) {
		t.Errorf("die erste Klasse ist %+v", gruppen[1])
	}

	if gruppen[2].Name != "6–10" || !slices.Equal(gruppen[2].Kartenkennungen, []int64{12}) {
		t.Errorf("die zweite Klasse ist %+v", gruppen[2])
	}
}

// **Eine Karte steht in einer Gruppe einmal — auch mit zwei Werten darin.**
//
// `GraphData::indices` überspringt einen Index, der gerade schon dastand. Der
// Balken zählt beide Elemente (`Karten` ist 2), die Liste zeigt die Karte
// aber nur einmal; und das gilt auch dann, wenn zwei **verschiedene Werte**
// in dieselbe Klasse fallen und die Karte dazwischen nicht an der Reihe war.
func TestEineKarteStehtInEinerGruppeEinmal(t *testing.T) {
	t.Parallel()

	gleicher := domain.AuszaehlenKarten(
		dimensionswerte("Rot", 5, 5), domain.Auszaehlart{},
	)

	if len(gleicher) != 1 || gleicher[0].Karten != 2 ||
		!slices.Equal(gleicher[0].Kartenkennungen, []int64{5}) {
		t.Errorf("zwei gleiche Werte einer Karte: %+v", gleicher)
	}

	klasse := domain.AuszaehlenKarten(
		slices.Concat(dimensionswerte("1", 5), dimensionswerte("2", 6), dimensionswerte("4", 5)),
		domain.Auszaehlart{Numerisch: true, Klassenbreite: 5},
	)

	if len(klasse) != 2 || klasse[1].Name != "1–5" ||
		!slices.Equal(klasse[1].Kartenkennungen, []int64{5, 6}) {
		t.Errorf("zwei Werte einer Karte in einer Klasse: %+v", klasse)
	}
}

// **Die Zählung allein bleibt, wie sie war** — `Auszaehlen` kennt keine Karten
// und nennt deshalb keine.
func TestDieZaehlungOhneKartenNenntKeine(t *testing.T) {
	t.Parallel()

	gruppen := domain.Auszaehlen([]string{"Rot", "Rot"}, domain.Auszaehlart{})

	if len(gruppen) != 1 || gruppen[0].Karten != 2 || gruppen[0].Kartenkennungen != nil {
		t.Errorf("die Gruppen lauten %+v", gruppen)
	}
}

// **Ohne Werte kommt eine leere Liste und kein `nil`** — auch mit Karten.
func TestOhneDimensionswerteKommtEineLeereListe(t *testing.T) {
	t.Parallel()

	if gruppen := domain.AuszaehlenKarten(nil, domain.Auszaehlart{}); gruppen == nil {
		t.Error("die Liste ist nil")
	}
}
