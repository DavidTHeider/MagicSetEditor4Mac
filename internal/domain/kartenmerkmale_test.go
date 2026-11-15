package domain_test

// **Was ein Symbol, eine Typzeile und eine Zahl bedeuten** — Werkstück M2 der
// Regelmaschine
// (#1604).
//
// # Wogegen hier geprüft wird
//
//   - CR 105.1 — fünf Farben; CR 202.2 — die Farbe folgt den Manasymbolen.
//   - CR 107.4a–h — farbig, generisch, farblos, hybrid, phyrexianisch, Schnee.
//   - CR 202.3 — der Manawert zählt alles, gleich welcher Farbe.
//   - CR 205.2a/205.3a/205.4a — fünfzehn Kartentypen, fünf Supertypen,
//     offene Untertypen.
//   - CR 208.2/208.3 — der Stern, und dass eine Nichtkreatur **keine**
//     Stärke hat statt null.
//
// **Die rohe Form steht im Test** (§3): geprüft wird das, was in einer
// MSE-Datei steht, nicht ein schon zerlegter Verbund.
//
// **Eine eigene Datei, und der Schnitt folgt der Frage** (§2): hier steht,
// was ein einzelnes Merkmal **bedeutet**; nebenan, wie aus Feldern eine Karte
// wird, und in `objekteigenschaften_test.go`, was am Tisch daraus gilt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die erste Abnahme des Issues**: eine Karte mit `{2}{U}{U}` ist blau,
// ohne dass jemand es einträgt.
func TestEineKarteMitBlauenSymbolenIstBlau(t *testing.T) {
	t.Parallel()

	kosten := domain.ManakostenAus("{2}{U}{U}")

	farben := kosten.Farben()
	if len(farben) != 1 || farben[0] != domain.FarbeBlau {
		t.Fatalf("Farben = %v", farben)
	}

	if kosten.Manawert() != 4 {
		t.Fatalf("Manawert = %d", kosten.Manawert())
	}
}

func TestManakostenWerdenGedeutet(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name    string
		roh     string
		farben  []domain.Farbe
		manawer int
	}{
		{"leer ist farblos", "", nil, 0},
		{"eine Null ist null", "{0}", nil, 0},
		{"ohne Klammern wie MSE es schreibt", "2U", []domain.Farbe{domain.FarbeBlau}, 3},
		{"zwei Farben", "{W}{B}", []domain.Farbe{domain.FarbeWeiss, domain.FarbeSchwarz}, 2},
		{"in der Reihenfolge des Regelwerks", "{G}{W}", []domain.Farbe{domain.FarbeWeiss, domain.FarbeGruen}, 2},
		{"hybrid ist beide Farben", "{W/U}", []domain.Farbe{domain.FarbeWeiss, domain.FarbeBlau}, 1},
		{"einfarbig hybrid zählt zwei", "{2/B}", []domain.Farbe{domain.FarbeSchwarz}, 2},
		{"phyrexianisch ist farbig", "{R/P}", []domain.Farbe{domain.FarbeRot}, 1},
		{"farblos ist keine Farbe", "{C}{C}", nil, 2},
		{"Schnee ist keine Farbe", "{S}", nil, 1},
		{"X zählt null", "{X}{R}", []domain.Farbe{domain.FarbeRot}, 1},
		{"Unbekanntes färbt nicht", "{Zauberstab}", nil, 0},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			kosten := domain.ManakostenAus(fall.roh)

			farben := kosten.Farben()
			if len(farben) != len(fall.farben) {
				t.Fatalf("Farben = %v, erwartet %v", farben, fall.farben)
			}

			for i := range farben {
				if farben[i] != fall.farben[i] {
					t.Fatalf("Farben = %v, erwartet %v", farben, fall.farben)
				}
			}

			if kosten.Manawert() != fall.manawer {
				t.Errorf("Manawert = %d, erwartet %d",
					kosten.Manawert(), fall.manawer)
			}
		})
	}
}

// **Ohne Manakosten ist ein Objekt farblos** — CR 202.2b, und der Manawert
// ist null (CR 202.3a). Das ist der Normalfall für jedes Land (CR 202.1b).
func TestOhneManakostenIstEinObjektFarblos(t *testing.T) {
	t.Parallel()

	kosten := domain.ManakostenAus("")

	if !kosten.Farblos() || kosten.Manawert() != 0 {
		t.Fatalf("farblos=%v, Manawert=%d", kosten.Farblos(), kosten.Manawert())
	}
}

// **Ein Kosten ohne Klammern endet auf einer Zahl** — `R10` ist rot und
// zehn, und die Zehn steht am Ende.
//
// Der Zweig dahinter räumt den Ziffernspeicher nach der Schleife; ohne ihn
// fiele die letzte Zahl heraus.
func TestEinKostenOhneKlammernEndetAufEinerZahl(t *testing.T) {
	t.Parallel()

	kosten := domain.ManakostenAus("R10")

	if kosten.Manawert() != 11 {
		t.Fatalf("Manawert = %d", kosten.Manawert())
	}
}

func TestTypzeilenWerdenGedeutet(t *testing.T) {
	t.Parallel()

	zeile, befunde := domain.TypzeileAus("Legendary Creature", "Human Wizard")

	switch {
	case len(befunde) != 0:
		t.Fatalf("Befunde: %+v", befunde)
	case !zeile.Hat(domain.SuperLegendaer):
		t.Error("nicht legendär")
	case !zeile.Ist(domain.TypKreatur):
		t.Error("keine Kreatur")
	case len(zeile.Untertypen) != 2 || zeile.Untertypen[0] != "Human":
		t.Errorf("Untertypen = %v", zeile.Untertypen)
	}
}

// **Eine Zeile mit Strich wird am Strich getrennt** — CR 205.3a.
func TestEineTypzeileMitStrichWirdGetrennt(t *testing.T) {
	t.Parallel()

	for _, roh := range []string{
		"Basic Land — Mountain", "Basic Land – Mountain", "Basic Land - Mountain",
	} {
		zeile, befunde := domain.TypzeileAus(roh, "")

		switch {
		case len(befunde) != 0:
			t.Fatalf("%q: Befunde %+v", roh, befunde)
		case !zeile.Hat(domain.SuperGrund) || !zeile.Ist(domain.TypLand):
			t.Errorf("%q: %+v", roh, zeile)
		case len(zeile.Untertypen) != 1 || zeile.Untertypen[0] != "Mountain":
			t.Errorf("%q: Untertypen %v", roh, zeile.Untertypen)
		}
	}
}

// **Die Gegenproben zu `Ist` und `Hat`** — ein Typ, den die Zeile nicht
// führt, ist keiner.
func TestEineTypzeileSagtAuchNein(t *testing.T) {
	t.Parallel()

	zeile, _ := domain.TypzeileAus("Creature", "Bear")

	switch {
	case zeile.Ist(domain.TypLand):
		t.Error("eine Kreatur ist ein Land")
	case zeile.Hat(domain.SuperLegendaer):
		t.Error("eine Kreatur ist legendär")
	}
}

// **Die dritte Abnahme des Issues**: ein Kartentyp, den die CR nicht kennt,
// ist ein Befund und kein stiller Ausfall (§8).
func TestEinUnbekannterKartentypIstEinBefund(t *testing.T) {
	t.Parallel()

	zeile, befunde := domain.TypzeileAus("Zauberding Creature", "")

	switch {
	case len(befunde) != 1:
		t.Fatalf("%d Befunde: %+v", len(befunde), befunde)
	case befunde[0].Ort != "Zauberding":
		t.Errorf("der Befund zeigt auf %q", befunde[0].Ort)
	case befunde[0].Art != domain.ArtVorlage:
		t.Errorf("die Art ist %q", befunde[0].Art)
	case !zeile.Ist(domain.TypKreatur):
		t.Error("der Rest der Zeile ging verloren")
	}
}

func TestKartenwerteWerdenGedeutet(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name  string
		roh   string
		steht bool
		stern bool
		zahl  int
	}{
		{"leer steht nicht", "", false, false, 0},
		{"eine Zahl", "3", true, false, 3},
		{"eine Null", "0", true, false, 0},
		{"ein Minus", "-1", true, false, -1},
		{"ein Stern", "*", true, true, 0},
		{"eins plus Stern", "1+*", true, true, 1},
		{"mit Leerraum", " 2 ", true, false, 2},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wert := domain.KartenwertAus(fall.roh)

			switch {
			case wert.Steht != fall.steht:
				t.Errorf("Steht = %v", wert.Steht)
			case wert.Stern != fall.stern:
				t.Errorf("Stern = %v", wert.Stern)
			case wert.Zahl != fall.zahl:
				t.Errorf("Zahl = %d", wert.Zahl)
			}
		})
	}
}

// **Die kanonischen Sets schreiben Manakosten in eckigen Klammern** — und
// zwar so: `[2][R]`, `[2/R]`, `[H/R]`.
//
// # Woher das kommt
//
// **Gemessen am 2026-09-17** an `Red.mse-set` aus
// `~/Library/Application Support/MagicSetEditor4Mac/kanonische-sets/` (3.730
// Karten, Scryfall-Daten mit Stand 2022-10-24, Herkunft in
// [0026](../../docs/entscheidungen/0026-mse2-rendert-die-referenz-selbst.md)) —
// bei der ersten Sichtprüfung des Spielkerns mit einem echten Deck.
//
// **`docs/FORMATE.md` sagte etwas anderes**, und es stimmte auch: die Zeile
// dort ist an `add_cards_scripts` erhoben, und **das** schreibt
// `casting_cost: "W"` ohne Trennzeichen. Zwei Schreibweisen, zwei Quellen —
// und eine dritte, die Scryfall und MTG JSON benutzen (`{2}{R}`). §8: fremde
// Daten sind der Normalfall.
//
// **Und das Phyrexianische heißt hier `H`** — CR 107.4g nennt `{H}` als das
// Symbol ohne Farbhintergrund; MSE schreibt es **vor** die Farbe.
func TestManakostenInEckigenKlammernWerdenGedeutet(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name    string
		roh     string
		farben  []domain.Farbe
		manawer int
	}{
		{"eine Farbe", "[R]", []domain.Farbe{domain.FarbeRot}, 1},
		{"generisch und Farbe", "[2][R]", []domain.Farbe{domain.FarbeRot}, 3},

		// **CR 202.3f**: beim Manawert zählt die **größere** Hälfte eines
		// Hybridsymbols. Flame Javelin kostet sechs und nicht neun.
		{"einfarbig hybrid", "[2/R][2/R][2/R]", []domain.Farbe{domain.FarbeRot}, 6},

		// **CR 202.3g**: ein phyrexianisches Symbol zählt eins.
		{"phyrexianisch", "[H/R]", []domain.Farbe{domain.FarbeRot}, 1},
		{"phyrexianisch mit generisch", "[4][H/R][H/R]", []domain.Farbe{domain.FarbeRot}, 6},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			kosten := domain.ManakostenAus(fall.roh)

			farben := kosten.Farben()
			if len(farben) != len(fall.farben) {
				t.Fatalf("Farben = %v, erwartet %v", farben, fall.farben)
			}

			for i := range farben {
				if farben[i] != fall.farben[i] {
					t.Fatalf("Farben = %v, erwartet %v", farben, fall.farben)
				}
			}

			if kosten.Manawert() != fall.manawer {
				t.Errorf("Manawert = %d, erwartet %d",
					kosten.Manawert(), fall.manawer)
			}
		})
	}
}

// **Und kein Symbol dieser Kosten bleibt unbekannt** — die Gegenprobe zum
// Lauf darüber: ohne sie wäre „die Farbe stimmt" auch dann wahr, wenn jede
// Klammer als unbekanntes Symbol durchginge.
func TestEckigeKlammernLassenNichtsUnbekannt(t *testing.T) {
	t.Parallel()

	for _, roh := range []string{"[R]", "[2][R]", "[2/R]", "[H/R]"} {
		for _, symbol := range domain.ManakostenAus(roh).Symbole {
			if symbol.Unbekannt {
				t.Errorf("%q: %q ist unbekannt", roh, symbol.Roh)
			}
		}
	}
}

// **Ein halb geschriebenes Symbol wird nicht geraten** — auch nicht in
// eckigen Klammern.
//
// **Dieselbe Zurückhaltung wie bei [domain.Manasymbole]**: aus `[2][R` kommt
// nur `[2]`. Den Rest zu ergänzen hieße, eine Eingabe zu erfinden (§8).
func TestEinHalbesEckigesSymbolWirdNichtGeraten(t *testing.T) {
	t.Parallel()

	kosten := domain.ManakostenAus("[2][R")

	if len(kosten.Symbole) != 1 || kosten.Manawert() != 2 {
		t.Fatalf("%d Symbole, Manawert %d", len(kosten.Symbole), kosten.Manawert())
	}
}
