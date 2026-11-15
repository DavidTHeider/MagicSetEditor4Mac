package storage_test

// **Die jüngsten Stände aller Drucke in einem Zug** — Werkstück 1 von
// Phase 4.5 (#1292).
//
// # Warum es diese Abfrage gibt
//
// Das Portfolio rief `Neueste` einmal je Exemplar. Gemessen an 5.000
// Exemplaren **ohne einen einzigen Preis**: 89 ms gegen 5 ms für dieselben
// Zeilen. Die Zusage, dass es nur noch **ein** Ruf ist, hält
// `service.TestPortfolioFragtNichtJeExemplar`; was hier geprüft wird, ist die
// Abfrage selbst — gegen echtes SQLite, wie §3 es für diese Schicht verlangt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// zweiterdruck ist der Druck neben `probedruck`.
const zweiterdruck = "22222222-2222-4222-8222-222222222222"

// **Über mehrere Drucke hinweg — und je Reihe der jüngste.**
//
// Der Fall, der beides zugleich prüft: zwei Drucke, und der erste hat in
// derselben Reihe zwei Stände. Eine Abfrage, die die Auswahl vergäße, gäbe
// drei Zeilen statt zwei; eine, die nach dem ersten Druck aufhörte, eine.
func TestNeuesteAlleGibtJeReiheDenJuengsten(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerPreisspeicher(db)

	spaeter := probestand(21, 499)

	for _, stand := range []domain.Preisstand{
		probestand(20, 417),
		spaeter,
		{
			Printing: zweiterdruck, Quelle: domain.PreisquelleScryfall,
			Waehrung: "EUR", Finish: domain.FinishNonfoil,
			Wert: 800, Erhoben: amTag(20),
		},
	} {
		if err := speicher.Eintragen(probekontext(t), stand); err != nil {
			t.Fatalf("Eintragen: %v", err)
		}
	}

	alle, err := speicher.NeuesteAlle(probekontext(t))
	if err != nil {
		t.Fatalf("NeuesteAlle: %v", err)
	}

	if len(alle) != 2 {
		t.Fatalf("es sind %d Zeilen, erwartet 2: %+v", len(alle), alle)
	}

	// **Nach Druck geordnet**, wie die Abfrage zusagt.
	werte := map[string]int64{}
	for _, stand := range alle {
		werte[stand.Printing] = stand.Wert
	}

	if werte[probedruck] != 499 {
		t.Errorf("der erste Druck steht bei %d, erwartet 499 — der ältere Stand "+
			"hat gewonnen", werte[probedruck])
	}

	if werte[zweiterdruck] != 800 {
		t.Errorf("der zweite Druck steht bei %d, erwartet 800", werte[zweiterdruck])
	}
}

// **Ein leerer Bestand gibt eine leere Liste und kein `nil`.**
//
// Dieselbe Zusage wie bei der Sammlung: über die Bindungen käme sonst ein
// `null` heraus, wo die Oberfläche eine Liste erwartet.
func TestNeuesteAlleOhneEinenEinzigenPreis(t *testing.T) {
	speicher := storage.NeuerPreisspeicher(frischeDatenbank(t))

	alle, err := speicher.NeuesteAlle(probekontext(t))
	if err != nil {
		t.Fatalf("NeuesteAlle: %v", err)
	}

	if alle == nil {
		t.Fatal("die Liste ist nil statt leer")
	}

	if len(alle) != 0 {
		t.Errorf("es sind %d Zeilen, erwartet 0", len(alle))
	}
}
