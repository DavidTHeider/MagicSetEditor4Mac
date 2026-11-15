package msepaket

import "strconv"

// Die `pack type:`-Blöcke eines Pakets
// (#1072).
//
// # Was hier steht und was nicht
//
// **Hier steht, was in der Datei steht.** Wie daraus ein Booster wird — die
// Tiefen, die Gewichte, das Ziehen —, steht in `service.Packzieher`: es ist
// eine fachliche Regel über mehrere Einheiten und braucht die
// Skriptauswertung für den Filter.

// Pack ist eine Packart — ein Booster, ein Blatt, eine Gruppe.
//
// Nachgelesen an `src/data/pack.hpp`, Zeilen 41–57, Fassung f9d9356 vom
// 2021-07-06.
type Pack struct {
	Name string

	// Auswahl ist `select` — **schon aufgelöst**.
	//
	// `auto` steht in der Datei; was es heißt, entscheidet
	// `after_reading(PackType&)` (`src/data/pack.cpp`, Zeilen 43–46): mit (Fassung f9d9356 vom 2021-07-06)
	// Filter `no replace`, mit Posten `all`. Die Auflösung steht hier, weil
	// sie eine Aussage über die **Datei** ist und nicht über das Ziehen.
	Auswahl string

	// Filter ist ein MSE-Script-Ausdruck über `card` — oder leer.
	Filter string

	// Posten sind die Unterpacks.
	Posten []Packposten

	// Waehlbar sagt, ob die Oberfläche die Art zur Auswahl stellt
	// (`selectable`). **Ohne Angabe: ja**, wenn es Posten gibt — MSE führt
	// dort einen Dreiwert, und die Oberfläche zeigt, was wählbar ist.
	Waehlbar bool
}

// Packposten ist ein Posten in einer Packart.
type Packposten struct {
	Name string

	// Anzahl ist `amount` — wie viele Züge aus diesem Posten kommen.
	// **Ohne Angabe eins** (`PackItem::PackItem`, `src/data/pack.cpp`, (Fassung f9d9356 vom 2021-07-06)
	// Zeilen 75–79).
	Anzahl int

	// Gewicht ist `weight` — die relative Wahrscheinlichkeit.
	// **Ohne Angabe eins**, ebenda.
	Gewicht float64
}

// Packs gibt die `pack type:`-Blöcke des Kopfes.
//
// **Ein Pack ohne Namen wird übergangen**: es ließe sich nicht nennen, und
// ein Posten, der auf nichts zeigt, wäre ein Fehler, den niemand findet.
func (p *Paket) Packs() []Pack {
	knoten := p.Kopf.Alle("pack type")
	packs := make([]Pack, 0, len(knoten))

	for _, einer := range knoten {
		gelesen := PackAus(einer)
		if gelesen.Name == "" {
			continue
		}

		packs = append(packs, gelesen)
	}

	return packs
}

// PackAus liest **einen** `pack type:`-Block.
//
// **Ausgelagert für die Packarten eines Sets**
// (#1081):
// dort kommen die Knoten aus dem Setleser und nicht aus [Paket.Packs]. Zwei
// Fassungen derselben Regel liefen auseinander, und dann zöge dasselbe Pack
// aus einem Spielpaket anders als aus einem Set.
//
// **Ein Pack ohne Namen kommt mit leerem Namen zurück** und nicht als Fehler —
// wer es übergehen will, prüft den Namen; [Paket.Packs] tut das.
func PackAus(knoten *Knoten) Pack {
	posten := packposten(knoten)
	waehlbar, _ := wahrheitMit(knoten, "selectable", len(posten) > 0)

	return Pack{
		Name:     WertOderBlock(knoten.Ersten("name")),
		Auswahl:  auswahlart(WertOderBlock(knoten.Ersten("select")), knoten, posten),
		Filter:   WertOderBlock(knoten.Ersten("filter")),
		Posten:   posten,
		Waehlbar: waehlbar,
	}
}

// auswahlart löst `auto` auf.
//
//	if (pt.select == SELECT_AUTO) {
//	  if (pt.filter)              pt.select = SELECT_NO_REPLACE;
//	  else if (!pt.items.empty()) pt.select = SELECT_ALL;
//	}
//
// `src/data/pack.cpp`, Zeilen 43–46, Fassung f9d9356 vom 2021-07-06.
//
// **Ein Pack ohne beides bleibt `auto`** — MSE lässt es dort ebenso stehen,
// und der Zieher zieht daraus nichts.
func auswahlart(genannt string, knoten *Knoten, posten []Packposten) string {
	if genannt != "" && genannt != "auto" {
		return genannt
	}

	switch {
	case WertOderBlock(knoten.Ersten("filter")) != "":
		return "no replace"

	case len(posten) > 0:
		return "all"

	default:
		return "auto"
	}
}

// packposten liest die `item:`-Blöcke.
func packposten(knoten *Knoten) []Packposten {
	kinder := knoten.Alle("item")
	posten := make([]Packposten, 0, len(kinder))

	for _, eines := range kinder {
		name := WertOderBlock(eines.Ersten("name"))
		if name == "" {
			continue
		}

		posten = append(posten, Packposten{
			Name:    name,
			Anzahl:  ganzzahlOder(WertOderBlock(eines.Ersten("amount")), 1),
			Gewicht: bruchOder(WertOderBlock(eines.Ersten("weight")), 1),
		})
	}

	return posten
}

// ganzzahlOder liest eine ganze Zahl — oder gibt die Vorgabe.
func ganzzahlOder(text string, vorgabe int) int {
	zahl, fehler := strconv.Atoi(text)
	if fehler != nil {
		return vorgabe
	}

	return zahl
}

// bruchOder liest eine Kommazahl — oder gibt die Vorgabe.
//
// **`weight` ist ein `Scriptable<double>`** (`pack.hpp:69`): im echten Pack
// stehen dort Zahlen, aber ein Skript wäre erlaubt. Was sich nicht als Zahl
// lesen lässt, bekommt hier die Vorgabe — und der Zieher rechnet damit weiter,
// statt das ganze Pack fallen zu lassen (§8).
func bruchOder(text string, vorgabe float64) float64 {
	zahl, fehler := strconv.ParseFloat(text, 64)
	if fehler != nil {
		return vorgabe
	}

	return zahl
}
