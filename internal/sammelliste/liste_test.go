package sammelliste_test

// **Eine Kartenliste lesen** — Werkstück 2 von Phase 4.5
// (#1292).
//
// # Warum es dieses Paket gibt
//
// Die Abnahme von #1292 verlangt, dass sich ein Bestand einlesen lässt,
// **ohne dass jemand eine Kennung von Hand kennt**. Heute braucht jedes
// Exemplar eine Scryfall-UUID; die schreibt niemand ab.
//
// Was Menschen stattdessen haben, ist eine **Liste** — aus einem
// Deckbaukasten, einem Marktplatz, einer Textdatei oder einem Zettel:
//
//	4 Lightning Bolt (M11) 149
//	2x Counterspell
//	1 Black Lotus (LEA) 232 *F*
//
// # Warum `new` — und wo gesucht wurde
//
// **MSE2 kennt keine Sammlung** und keinen Listenimport. Gesucht am
// 2026-08-30 im Quelltext `f9d9356` nach `decklist`, `deck_list`,
// `inventory`, `wishlist` und `owned`: die Treffer sind Kommentare über
// Zeigerbesitz in `render/text/element.hpp` und `script/to_value.hpp` — nichts
// zur Sache.
//
// **`magic-set-generator` (`1b8fb7f`) hat genau eine Zeile dazu**, und die ist
// ein Vorhaben: `src/lib.rs:266`, „TODO also read card names from
// args.decklists". Auch dort ist nichts umgesetzt.
//
// Das Format ist damit keine Portierung, sondern eine **Produktentscheidung**
// über eine gewachsene Gemeinschaftskonvention (§15).
//
// # Warum ein Blatt und nicht `service`
//
// Weil Zerteilen rein ist: Text hinein, Zeilen heraus, kein Netz und kein
// Bestand. Die **Zuordnung** auf einen Druck braucht Scryfall und steht
// deshalb in `service` — das ist dieselbe Trennung wie bei `msepaket` (§2).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sammelliste"
)

// **Die Gestalten, die eine Zeile haben kann.**
func TestZeilenformen(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		roh      string
		erwartet sammelliste.Zeile
	}{
		{
			name: "Menge, Name, Set und Nummer",
			roh:  "4 Lightning Bolt (M11) 149",
			erwartet: sammelliste.Zeile{
				Menge: 4, Name: "Lightning Bolt", Setcode: "m11", Nummer: "149",
			},
		},
		{
			name:     "Menge mit x",
			roh:      "2x Counterspell",
			erwartet: sammelliste.Zeile{Menge: 2, Name: "Counterspell"},
		},
		{
			// **`12 X Dark Ritual` sind zwölf Dark Ritual.** Das `X` ist die
			// Mengenmarke, auch mit Leerzeichen davor — so schreiben es
			// Marktplätze. Eine Karte namens „X Dark Ritual" gibt es nicht,
			// und der erste Anlauf dieses Tests erwartete genau die.
			name:     "Menge mit X und Leerzeichen",
			roh:      "12 X Dark Ritual",
			erwartet: sammelliste.Zeile{Menge: 12, Name: "Dark Ritual"},
		},
		{
			name:     "ohne Menge ist eines",
			roh:      "Black Lotus",
			erwartet: sammelliste.Zeile{Menge: 1, Name: "Black Lotus"},
		},
		{
			name: "Foil hinten",
			roh:  "1 Black Lotus (LEA) 232 *F*",
			erwartet: sammelliste.Zeile{
				Menge: 1, Name: "Black Lotus", Setcode: "lea",
				Nummer: "232", Finish: domain.FinishFoil,
			},
		},
		{
			name: "Foil ohne Set",
			roh:  "3 Shivan Dragon *foil*",
			erwartet: sammelliste.Zeile{
				Menge: 3, Name: "Shivan Dragon",
				Finish: domain.FinishFoil,
			},
		},
		{
			name: "Setcode ohne Nummer",
			roh:  "1 Giant Growth (m11)",
			erwartet: sammelliste.Zeile{
				Menge: 1, Name: "Giant Growth", Setcode: "m11",
			},
		},
		{
			name:     "Ränder werden geglättet",
			roh:      "   7   Serra   Angel   ",
			erwartet: sammelliste.Zeile{Menge: 7, Name: "Serra   Angel"},
		},
		{
			name: "eckige Klammern statt runder",
			roh:  "1 Llanowar Elves [M11] 189",
			erwartet: sammelliste.Zeile{
				Menge: 1, Name: "Llanowar Elves", Setcode: "m11", Nummer: "189",
			},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			zeilen, befunde := sammelliste.Lies(fall.roh)

			if len(befunde) != 0 {
				t.Fatalf("unerwartete Befunde: %v", befunde)
			}

			if len(zeilen) != 1 {
				t.Fatalf("%d Zeilen, erwartet 1", len(zeilen))
			}

			erwartet := fall.erwartet
			erwartet.Nummer = fall.erwartet.Nummer
			erwartet.Zeilennummer = 1
			erwartet.Roh = strings.TrimSpace(fall.roh)

			if zeilen[0] != erwartet {
				t.Errorf("gelesen  %+v\nerwartet %+v", zeilen[0], erwartet)
			}
		})
	}
}

// **Leerzeilen, Kommentare und Überschriften gehen durch, ohne zu stören.**
//
// Wer eine Deckliste einfügt, bringt „Deck", „Sideboard" und Leerzeilen mit.
// Sie als Karten zu lesen ergäbe drei Karten, die es nicht gibt; sie zu
// bemängeln ergäbe drei Befunde, die niemand lesen will.
func TestWasUebergangenWird(t *testing.T) {
	t.Parallel()

	zeilen, befunde := sammelliste.Lies(`Deck

// eine Notiz
# noch eine
4 Lightning Bolt

Sideboard
2 Counterspell
`)

	if len(befunde) != 0 {
		t.Fatalf("unerwartete Befunde: %v", befunde)
	}

	if len(zeilen) != 2 {
		t.Fatalf("%d Zeilen, erwartet 2: %+v", len(zeilen), zeilen)
	}

	// **Die Zeilennummer ist die der Datei**, nicht die des Ergebnisses. Ohne
	// das zeigte ein Befund auf eine Zeile, die niemand findet.
	if zeilen[0].Zeilennummer != 5 {
		t.Errorf("die erste Karte steht in Zeile %d, erwartet 5", zeilen[0].Zeilennummer)
	}

	if zeilen[1].Zeilennummer != 8 {
		t.Errorf("die zweite Karte steht in Zeile %d, erwartet 8", zeilen[1].Zeilennummer)
	}
}

// **Eine Zeile ohne Namen ist ein Befund — und der Rest wird weitergelesen.**
//
// §8: ein Set, von dem 300 Karten lesbar sind und zwei nicht, wird mit 300
// Karten und zwei benannten Befunden geöffnet. Dasselbe gilt hier.
func TestEineUnlesbareZeileHaeltDenRestNichtAuf(t *testing.T) {
	t.Parallel()

	// **`7` allein ist keine Karte.** Es gibt keinen Kartennamen, der nur aus
	// Ziffern besteht — wohl aber welche **mit** Ziffern („1996 World
	// Champion"), und die dürfen nicht mit auffallen.
	zeilen, befunde := sammelliste.Lies("4 Lightning Bolt\n7\n2 Counterspell")

	if len(zeilen) != 2 {
		t.Errorf("%d Zeilen, erwartet 2: %+v", len(zeilen), zeilen)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %v", len(befunde), befunde)
	}

	if befunde[0].Ort != "Zeile 2" {
		t.Errorf("der Befund zeigt auf %q, erwartet \"Zeile 2\"", befunde[0].Ort)
	}
}

// **Eine Menge von null oder weniger ist kein Bestand.**
func TestEineMengeUnterEins(t *testing.T) {
	t.Parallel()

	zeilen, befunde := sammelliste.Lies("0 Lightning Bolt")

	if len(zeilen) != 0 {
		t.Errorf("%d Zeilen, erwartet 0: %+v", len(zeilen), zeilen)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %v", len(befunde), befunde)
	}
}

// **Eine leere Eingabe ist keine Liste — und kein Absturz.**
func TestEineLeereEingabe(t *testing.T) {
	t.Parallel()

	zeilen, befunde := sammelliste.Lies("   \n\n  ")

	if len(zeilen) != 0 || len(befunde) != 0 {
		t.Errorf("%d Zeilen und %d Befunde, erwartet 0 und 0", len(zeilen), len(befunde))
	}
}

// **Ein Name mit Ziffern geht durch.**
//
// Die Gegenprobe zur Regel darüber: geprüft wird „nichts als Ziffern", nicht
// „enthält Ziffern". Ohne diesen Test wäre eine Verschärfung auf
// `strings.ContainsAny` grün — und „1996 World Champion" fiele heraus.
func TestEinNameMitZiffern(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"1996 World Champion", "Borrowing 100,000 Arrows", "B.F.M. 2"} {
		zeilen, befunde := sammelliste.Lies("1 " + name)

		if len(befunde) != 0 {
			t.Errorf("%q gibt Befunde: %v", name, befunde)

			continue
		}

		if len(zeilen) != 1 || zeilen[0].Name != name {
			t.Errorf("%q wurde als %+v gelesen", name, zeilen)
		}
	}
}

// **Eine Zeile, von der nach Menge, Foil und Druck nichts übrig bleibt.**
//
// `4 (M11) 149` nennt eine Stückzahl und einen Druck — und keine Karte. Das
// ist kein erfundener Fall: es entsteht, wenn jemand eine Spalte zu viel
// löscht.
func TestEineZeileOhneNamen(t *testing.T) {
	t.Parallel()

	for _, roh := range []string{"4 (M11) 149", "1 *F*", "(LEA)"} {
		zeilen, befunde := sammelliste.Lies(roh)

		if len(zeilen) != 0 {
			t.Errorf("%q gab %d Zeilen, erwartet 0: %+v", roh, len(zeilen), zeilen)
		}

		if len(befunde) != 1 {
			t.Errorf("%q gab %d Befunde, erwartet 1: %v", roh, len(befunde), befunde)
		}
	}
}

// **Eine Überschrift gliedert, und die Gliederung geht nicht verloren** —
// Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Warum der Abschnitt roh bleibt und kein [domain.Deckbereich] ist
//
// **Weil dieses Paket nicht weiß, wofür gelesen wird.** Wer eine Liste in die
// **Sammlung** einliest, hat keine Bereiche; wer ein **Deck** baut, hat zwei.
// Und was aus `maybeboard` wird, ist eine fachliche Entscheidung und keine
// des Zerteilers — hier stünde sie an der falschen Stelle.
func TestEineUeberschriftBleibtAmFolgendenZeilenHaengen(t *testing.T) {
	t.Parallel()

	zeilen, befunde := sammelliste.Lies(`Deck
4 Lightning Bolt (M11) 149

Sideboard
2 Counterspell

Maybeboard
1 Brainstorm`)

	if len(befunde) != 0 {
		t.Fatalf("es gab Befunde: %+v", befunde)
	}

	erwartet := []string{"deck", "sideboard", "maybeboard"}

	if len(zeilen) != len(erwartet) {
		t.Fatalf("%d Zeilen statt %d: %+v", len(zeilen), len(erwartet), zeilen)
	}

	for stelle, abschnitt := range erwartet {
		if zeilen[stelle].Abschnitt != abschnitt {
			t.Errorf("Zeile %d steht unter %q, erwartet %q",
				stelle+1, zeilen[stelle].Abschnitt, abschnitt)
		}
	}
}

// **Und ohne Überschrift steht nichts da.**
//
// **Leer und nicht `deck`**: „hier stand keine Überschrift" und „hier stand
// `Deck`" sind zwei Auskünfte, und wer sie gleich schriebe, nähme dem
// Aufrufer die Entscheidung, was eine Liste ohne Gliederung bedeutet.
func TestOhneUeberschriftBleibtDerAbschnittLeer(t *testing.T) {
	t.Parallel()

	zeilen, _ := sammelliste.Lies("4 Lightning Bolt\n2 Counterspell")

	for _, zeile := range zeilen {
		if zeile.Abschnitt != "" {
			t.Errorf("%q steht unter %q", zeile.Name, zeile.Abschnitt)
		}
	}
}
