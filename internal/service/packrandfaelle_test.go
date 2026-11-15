// Die Randfälle des Ziehens
// (#1072).
//
// **Von `packauswahl_test.go` getrennt**, als sie über 400 Zeilen wuchs (§2).
// Der Schnitt folgt der Frage: dort, **wie** jede Auswahlart zieht, hier, was
// geschieht, wenn eine Angabe fehlt oder ins Leere zeigt.
package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Randfälle des Ziehens.**
//
// Sie stehen zusammen, weil sie dieselbe Frage beantworten: was tut der
// Zieher, wenn eine Angabe fehlt oder ins Leere zeigt? Die Antwort ist
// überall dieselbe — er zieht weniger, statt zu scheitern (§8).
func TestDieRandfaelleDesZiehens(t *testing.T) {
	t.Parallel()

	t.Run("ein Posten, den es nicht gibt", func(t *testing.T) {
		t.Parallel()

		zieher := service.NeuerPackzieher([]msepaket.Pack{
			{Name: "wurzel", Auswahl: "all", Posten: []msepaket.Packposten{
				{Name: "gibtsnicht", Anzahl: 1, Gewicht: 1},
			}},
		}, nil)

		if _, fehler := zieher.Ziehen("wurzel", 1); fehler == nil {
			t.Error("der Posten ins Leere blieb still")
		}
	})

	t.Run("eine Packart steht zweimal — die erste gewinnt", func(t *testing.T) {
		t.Parallel()

		zieher := service.NeuerPackzieher([]msepaket.Pack{
			{Name: "doppelt", Auswahl: "all"},
			{Name: "doppelt", Auswahl: "first"},
		}, map[string][]*domain.CardRow{"doppelt": packkarten("d", 3)})

		gezogen, fehler := zieher.Ziehen("doppelt", 1)
		if fehler != nil {
			t.Fatalf("Ziehen: %v", fehler)
		}

		// **`all` gewinnt** und nimmt alle drei; `first` nähme eine.
		if len(gezogen) != 3 {
			t.Errorf("%d Karten, erwartet drei", len(gezogen))
		}
	})

	t.Run("ein leeres Blatt zieht nichts", func(t *testing.T) {
		t.Parallel()

		for _, auswahl := range []string{"no replace", "replace", "first", "equal"} {
			zieher := service.NeuerPackzieher(
				[]msepaket.Pack{{Name: "leer", Auswahl: auswahl}}, nil)

			gezogen, fehler := zieher.Ziehen("leer", 1)
			if fehler != nil {
				t.Fatalf("%s: %v", auswahl, fehler)
			}

			if len(gezogen) != 0 {
				t.Errorf("%s zog %d Karten aus dem Leeren", auswahl, len(gezogen))
			}
		}
	})

	t.Run("equal mit einem einzigen Posten", func(t *testing.T) {
		t.Parallel()

		// **Ein einziger Posten bekommt alles** — `weighted_equal_divide`
		// hat dafür einen eigenen Zweig (Zeile 195).
		zieher := service.NeuerPackzieher([]msepaket.Pack{
			{Name: "wurzel", Auswahl: "all", Posten: []msepaket.Packposten{
				{Name: "gleich", Anzahl: 4, Gewicht: 1},
			}},
			{Name: "gleich", Auswahl: "equal"},
		}, map[string][]*domain.CardRow{"gleich": packkarten("g", 3)})

		gezogen, fehler := zieher.Ziehen("wurzel", 1)
		if fehler != nil {
			t.Fatalf("Ziehen: %v", fehler)
		}

		// **Vier Züge auf drei Karten**: jede einmal, eine davon zweimal.
		if len(gezogen) != 4 {
			t.Errorf("%d Karten, erwartet vier", len(gezogen))
		}
	})

	t.Run("equal proportional und equal nonempty verteilen", func(t *testing.T) {
		t.Parallel()

		for _, auswahl := range []string{"equal proportional", "equal nonempty"} {
			zieher := service.NeuerPackzieher([]msepaket.Pack{
				{Name: "wurzel", Auswahl: "all", Posten: []msepaket.Packposten{
					{Name: "gleich", Anzahl: 6, Gewicht: 1},
				}},
				{Name: "gleich", Auswahl: auswahl, Posten: []msepaket.Packposten{
					{Name: "voll", Anzahl: 1, Gewicht: 1},
					{Name: "leer", Anzahl: 1, Gewicht: 1},
				}},
				{Name: "voll", Auswahl: "no replace"},
				{Name: "leer", Auswahl: "no replace"},
			}, map[string][]*domain.CardRow{"voll": packkarten("v", 4)})

			gezogen, fehler := zieher.Ziehen("wurzel", 2)
			if fehler != nil {
				t.Fatalf("%s: %v", auswahl, fehler)
			}

			// **Der leere Posten bekommt nichts**: bei `proportional` wiegt er
			// null, bei `nonempty` zählt er nicht mit.
			if len(gezogen) != 6 {
				t.Errorf("%s zog %d Karten, erwartet sechs", auswahl, len(gezogen))
			}
		}
	})

	t.Run("proportional wiegt mit dem Gewicht des Postens", func(t *testing.T) {
		t.Parallel()

		zieher := service.NeuerPackzieher([]msepaket.Pack{
			{Name: "wurzel", Auswahl: "proportional", Posten: []msepaket.Packposten{
				{Name: "leer", Anzahl: 1, Gewicht: 1000},
				{Name: "voll", Anzahl: 1, Gewicht: 1},
			}},
			{Name: "leer", Auswahl: "no replace"},
			{Name: "voll", Auswahl: "no replace"},
		}, map[string][]*domain.CardRow{"voll": packkarten("v", 4)})

		for same := range uint64(5) {
			gezogen, fehler := zieher.Ziehen("wurzel", same)
			if fehler != nil {
				t.Fatalf("Same %d: %v", same, fehler)
			}

			// **Der leere Posten wiegt null**, gleich wie groß sein Gewicht
			// steht: `proportional` rechnet Gewicht **mal** Kartenzahl.
			if len(gezogen) != 1 {
				t.Errorf("Same %d: %d Karten", same, len(gezogen))
			}
		}
	})
}

// **Zwei Packarten dürfen denselben Posten nennen.**
//
// Der Stand wird **einmal** gebaut; wer ihn zweimal baute, hätte zwei
// Gewichte für dasselbe und zöge zweimal daraus.
func TestZweiPackartenDuerfenDenselbenPostenNennen(t *testing.T) {
	t.Parallel()

	zieher := service.NeuerPackzieher([]msepaket.Pack{
		{Name: "wurzel", Auswahl: "all", Posten: []msepaket.Packposten{
			{Name: "links", Anzahl: 1, Gewicht: 1},
			{Name: "rechts", Anzahl: 1, Gewicht: 1},
		}},
		{Name: "links", Auswahl: "all", Posten: []msepaket.Packposten{
			{Name: "geteilt", Anzahl: 1, Gewicht: 1},
		}},
		{Name: "rechts", Auswahl: "all", Posten: []msepaket.Packposten{
			{Name: "geteilt", Anzahl: 1, Gewicht: 1},
		}},
		{Name: "geteilt", Auswahl: "no replace"},
	}, map[string][]*domain.CardRow{"geteilt": packkarten("g", 5)})

	gezogen, fehler := zieher.Ziehen("wurzel", 1)
	if fehler != nil {
		t.Fatalf("Ziehen: %v", fehler)
	}

	// **Zweimal einer verlangt, zweimal einer gezogen.**
	if len(gezogen) != 2 {
		t.Errorf("%d Karten, erwartet zwei", len(gezogen))
	}
}

// **`equal` mit lauter Nullgewichten verteilt nichts.**
//
// Der Zweig ist erreichbar: ein Pack, dessen Posten alle leer sind und das
// selbst keine Karten hat.
func TestEqualMitLauterNullgewichtenVerteiltNichts(t *testing.T) {
	t.Parallel()

	zieher := service.NeuerPackzieher([]msepaket.Pack{
		{Name: "wurzel", Auswahl: "all", Posten: []msepaket.Packposten{
			{Name: "gleich", Anzahl: 4, Gewicht: 1},
		}},
		{Name: "gleich", Auswahl: "equal nonempty", Posten: []msepaket.Packposten{
			{Name: "leer", Anzahl: 1, Gewicht: 1},
		}},
		{Name: "leer", Auswahl: "no replace"},
	}, nil)

	gezogen, fehler := zieher.Ziehen("wurzel", 1)
	if fehler != nil {
		t.Fatalf("Ziehen: %v", fehler)
	}

	if len(gezogen) != 0 {
		t.Errorf("%d Karten aus dem Leeren", len(gezogen))
	}
}

// **Der spätere Posten gewinnt, wenn er schwerer wiegt.**
//
// `CompareWeightedItems` (Zeilen 181–189) vergleicht in beide Richtungen;
// ohne diesen Lauf bliebe eine der beiden ungeprüft — und eine Verteilung,
// die nur in einer Richtung stimmt, ist keine.
func TestDerSchwerereGewinntAuchWennErSpaeterSteht(t *testing.T) {
	t.Parallel()

	zieher := service.NeuerPackzieher([]msepaket.Pack{
		{Name: "wurzel", Auswahl: "all", Posten: []msepaket.Packposten{
			{Name: "gleich", Anzahl: 8, Gewicht: 1},
		}},
		{Name: "gleich", Auswahl: "equal", Posten: []msepaket.Packposten{
			{Name: "selten", Anzahl: 1, Gewicht: 1},
			{Name: "haeufig", Anzahl: 1, Gewicht: 3},
		}},
		{Name: "selten", Auswahl: "no replace"},
		{Name: "haeufig", Auswahl: "no replace"},
	}, map[string][]*domain.CardRow{
		"haeufig": packkarten("h", 20),
		"selten":  packkarten("s", 20),
	})

	gezogen, fehler := zieher.Ziehen("wurzel", 5)
	if fehler != nil {
		t.Fatalf("Ziehen: %v", fehler)
	}

	haeufige, seltene := zaehleNachVorsilbe(gezogen)
	if haeufige != 6 || seltene != 2 {
		t.Errorf("%d häufige und %d seltene, erwartet 6 und 2", haeufige, seltene)
	}
}

// **Der Vorgabezweig von `generate_one_random`: `replace` mit Posten.**
//
// `proportional` und `nonempty` haben ihre eigenen Zeilen; `replace` und
// `equal` bei einer Kopie nehmen den Zweig darunter — das bloße Gewicht des
// Postens. Ohne diesen Lauf bliebe er ungeprüft.
func TestReplaceZiehtAusEinemPosten(t *testing.T) {
	t.Parallel()

	zieher := service.NeuerPackzieher([]msepaket.Pack{
		{Name: "wurzel", Auswahl: "replace", Posten: []msepaket.Packposten{
			{Name: "blatt", Anzahl: 2, Gewicht: 1},
		}},
		{Name: "blatt", Auswahl: "no replace"},
	}, map[string][]*domain.CardRow{"blatt": packkarten("b", 6)})

	gezogen, fehler := zieher.Ziehen("wurzel", 3)
	if fehler != nil {
		t.Fatalf("Ziehen: %v", fehler)
	}

	// **Die Wurzel hat keine eigenen Karten**, also fällt der Zufall in den
	// Posten — und der gibt zwei.
	if len(gezogen) != 2 {
		t.Errorf("%d Karten, erwartet zwei", len(gezogen))
	}
}

// **Bei Gleichstand entscheidet das Los.**
//
//	WeightedItem wi = {0,0,(int)parent.gen()};
//	…
//	return b->tiebreaker < a->tiebreaker;
//
// `src/data/pack.cpp`, Zeilen 175–188, Fassung f9d9356 vom 2021-07-06.
//
// # Warum das kein Schmuck ist
//
// Zwei Posten mit demselben Gewicht und demselben Stand sind gleich gut. Ohne
// Los bekäme **immer derselbe** den Zuschlag — bei drei Zügen auf zwei
// gleiche Posten stünde es dann jedes Mal 2:1 für denselben.
//
// Geprüft wird über die Samen: derselbe gibt dasselbe, und über mehrere
// fällt der ungerade Zug mal hierhin, mal dorthin.
func TestBeiGleichstandEntscheidetDasLos(t *testing.T) {
	t.Parallel()

	bauen := func() *service.Packzieher {
		return service.NeuerPackzieher([]msepaket.Pack{
			{Name: "wurzel", Auswahl: "all", Posten: []msepaket.Packposten{
				{Name: "gleich", Anzahl: 3, Gewicht: 1},
			}},
			{Name: "gleich", Auswahl: "equal", Posten: []msepaket.Packposten{
				{Name: "haeufig", Anzahl: 1, Gewicht: 1},
				{Name: "selten", Anzahl: 1, Gewicht: 1},
			}},
			{Name: "haeufig", Auswahl: "no replace"},
			{Name: "selten", Auswahl: "no replace"},
		}, map[string][]*domain.CardRow{
			"haeufig": packkarten("h", 9),
			"selten":  packkarten("s", 9),
		})
	}

	gesehen := map[int]bool{}

	for same := range uint64(12) {
		gezogen, fehler := bauen().Ziehen("wurzel", same)
		if fehler != nil {
			t.Fatalf("Same %d: %v", same, fehler)
		}

		haeufige, _ := zaehleNachVorsilbe(gezogen)
		gesehen[haeufige] = true
	}

	// **Beide Ausgänge müssen vorkommen.** Ohne Los stünde hier eine einzige
	// Zahl — und genau das ist die Gegenprobe.
	if len(gesehen) < 2 {
		t.Errorf("über zwölf Samen kam immer dasselbe heraus: %v", gesehen)
	}
}

// TestEinePackartMitFilterUndPostenBekommtBeides ist Befund 3 des Zweitblicks
// über Phase W.
//
//	// Filter cards
//	if (pack_type.filter) {
//	  FOR_EACH(card, parent.set->cards) {
//	    …
//	  }
//	}
//
// `src/data/pack.cpp`, Zeilen 110–117, Fassung f9d9356 vom 2021-07-06. **Der
// Filter läuft, wenn es einen gibt** — die Posten ändern daran nichts.
// `generate_all` hängt danach die eigenen Karten **und** alle Posten an
// (Zeilen 317–328).
//
// Hier wurde jede Packart mit einem Posten von der Filterauswertung
// ausgeschlossen. Die Begründung galt nur für `select: auto`: dort **wird**
// eine Postenart zu `all`, aber nur, wenn kein Filter dasteht (Zeilen 43–46).
//
// **Gemessen am 2026-08-24**: von 29 Packarten in
// `M15-Magic-Pack/data/magic.mse-game/packs` trägt **keine einzige** Filter
// und Posten zugleich. Der Befund ist echt und im echten Pack folgenlos; er
// wird behoben, weil die Abweichung sonst als gelesene Regel dasteht.
func TestEinePackartMitFilterUndPostenBekommtBeides(t *testing.T) {
	t.Parallel()

	karten := []*domain.CardRow{packkarte(1, "common"), packkarte(2, "rare")}

	bereich := packbereichMit(karten, service.Vorlage{Packs: []msepaket.Pack{
		{Name: "common", Auswahl: "no replace", Filter: `card.rarity == "common"`},
		{
			Name: "selten", Auswahl: "all", Waehlbar: true,
			Filter: `card.rarity == "rare"`,
			Posten: []msepaket.Packposten{{Name: "common", Anzahl: 1, Gewicht: 1}},
		},
	}}, true)

	gezogen, befunde, err := bereich.Ziehen(t.Context(), 1, "selten", 7)
	if err != nil {
		t.Fatalf("Ziehen: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde: %+v", befunde)
	}

	// **Beides**: die eigene seltene Karte aus dem Filter und die eine
	// gewöhnliche aus dem Posten.
	if len(gezogen) != 2 {
		t.Fatalf("%d Karten gezogen, will die eigene und die des Postens: %+v", len(gezogen), gezogen)
	}

	gesehen := map[int64]bool{}
	for _, karte := range gezogen {
		gesehen[karte.ID] = true
	}

	if !gesehen[1] || !gesehen[2] {
		t.Errorf("gezogen wurden %v — will beide", gesehen)
	}
}
