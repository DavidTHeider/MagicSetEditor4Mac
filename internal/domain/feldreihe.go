package domain

import (
	"regexp"
	"sort"
	"strconv"
)

// **Wie MSE mehrseitige Karten führt, und woher wir es wissen** — Werkstück 3
// von Phase 3.6
// (#1167).
//
// # Eine Karte, nicht zwei
//
// **Drei unabhängige Belege, alle am 2026-08-27 erhoben:**
//
//   - `src/data/card.hpp`, Fassung f9d9356 vom 2021-07-06, führt in `Card`:
//     `data`, `notes`, `time_created`, `time_modified`, `stylesheet`,
//     `styling_data`, `extra_data`, `keyword_usage`. **Keinen Bezug auf eine
//     andere Karte** — MSE2 kennt keine Beziehung Vorderseite ↔ Rückseite.
//   - `magic.mse-game` erklärt 249 Kartenfelder; **43** davon haben
//     numerierte Geschwister (`name`/`name 2`/`name 3`/`name 4`, `casting
//     cost`/`casting cost 2`, …). Die höchste Stelle ist **8**.
//
//     **Gezählt über den Leser dieses Projekts, nicht über ein Textmuster**
//     (§10 — die Reichweite einer Messung ist die Reichweite ihrer Aussage).
//     Der erste Anlauf zählte `grep "^\tname:"` und bekam 41 statt 40 Felder
//     ohne Grundfeld: Felder in eingebundenen Dateien tragen eine andere
//     Einrückung, und `alt name:` sieht aus wie `name:`. Nachgeprüft wird es
//     von `internal/msepaket/feldreihenfremd_test.go` hinter dem Tag
//     `fremdpakete`.
//   - `magic-set-generator` (MIT, §11) schreibt es genauso: sein Makro
//     `push_alt!` in `src/mse.rs` hängt bei der **zweiten Seite** ` 2` an
//     jeden Schlüssel und legt beide Seiten in **dieselbe** Karte
//     (`Layout::DoubleFaced`, `Flip`, `Split`, `Meld`, `Adventure`).
//
// # Warum hier keine Beziehung zwischen zwei Karten entsteht
//
// **Weil sie den Rundlauf nicht überlebte.** MSEs `Reader::unknownKey()`
// überspringt einen unbekannten Schlüssel und verliert ihn beim Speichern
// still ([0053](../../docs/entscheidungen/0053-marken-liegen-neben-dem-set.md)).
// Eine erfundene Kante `karte → rückseite` wäre nach einem Umweg durch MSE2
// fort — und niemand sähe, wann. Die Beziehung, die den Rundlauf überlebt,
// ist die **im** Kartensatz, und sie ist vom Spielpaket erklärt.
//
// # Die Gestalt ist übernommen, die Erkennungsregel ist neu
//
// **`adapted`, nicht `ported`** (Befund 8 des Zweitblicks über Phase 3.6):
// dass eine mehrseitige Karte **eine** Karte mit numerierten Feldreihen ist,
// steht in allen drei Quellen. Die Regel, **an der eine Reihe erkannt wird**,
// steht in keiner — MSE2 braucht sie nicht, weil dort das Stilpaket zeichnet,
// was es kennt. Ein Formular, das hundert Felder auflistet, braucht sie.
//
// # Warum „Reihe" und nicht „Seite"
//
// Weil die Zahl nicht immer eine Seite meint. `additional credit 2` bis
// `additional credit 5` sind fünf Zeilen einer Danksagung, nicht fünf Seiten;
// `magic-set-generator` benutzt `rule text 3` für das Schlüsselwort *Fuse*
// (`src/mse.rs`, Zeile 293). **Was eine Stelle bedeutet, entscheidet das
// Stilpaket**: `magic-m15-mainframe-dfc.mse-style` zeichnet `name` und
// `name 2` als Vorder- und Rückseite. Dieses Paket beschreibt deshalb die
// **Gestalt** und behauptet nichts über die Bedeutung.

// stellenmuster trifft eine Stelle am Ende eines Feldnamens.
//
// **Hinter einem Leerzeichen, ohne führende Null, bis zum Ende.** `name2` ist
// keine Stelle, `name 2 lang` auch nicht, und `name 02` wäre eine zweite
// Schreibweise derselben — drei Formen, die kein Paket benutzt und die eine
// lose Regel hereinließe.
var stellenmuster = regexp.MustCompile(`^(.+) ([1-9]\d*)$`)

// Stelle ist eine numerierte Position einer Feldreihe.
type Stelle struct {
	// Feld ist der volle Feldname — `name 2`.
	Feld string

	// Nummer ist die Stelle — 2 für `name 2`. Immer größer als 1.
	Nummer int
}

// Feldreihe ist ein Kartenfeld samt seinen numerierten Geschwistern.
//
// **Das Grundfeld ist die erste Stelle** und steht deshalb nicht in
// [Feldreihe.Stellen]: ein `name 1` daneben wäre eine zweite erste, und
// niemand könnte sagen, welche gilt. In `magic.mse-game` gibt es kein
// einziges solches Paar.
type Feldreihe struct {
	Grundfeld string
	Stellen   []Stelle
}

// Feldreihen leitet aus den Feldern eines Spiels ab, welche eine Reihe bilden.
//
// # Die Regel in einem Satz
//
// **Ein Feld `F N` gehört zur Reihe von `F`, wenn es das Feld `F` gibt.**
//
// Ohne die zweite Hälfte wäre `loyalty cost 1` bis `loyalty cost 16` eine
// sechzehnstellige Reihe — es sind die Zeilen eines Planeswalkers, und ein
// Feld `loyalty cost` gibt es nicht. Gezählt an `magic.mse-game` am
// 2026-08-27: **40** Felder mit Zahl haben kein Grundfeld.
//
// **In der Ordnung des Spielpakets**, wie alles andere: eine alphabetische
// wäre eine zweite Ordnung neben der, die MSE benutzt.
func (g Game) Feldreihen() []Feldreihe {
	vorhanden := make(map[string]bool, len(g.Fields))
	for i := range g.Fields {
		vorhanden[g.Fields[i].Name] = true
	}

	stellen := make(map[string][]Stelle)

	for i := range g.Fields {
		// **`len(teile) == 0` und nicht `teile == nil`.** Ein leeres, aber
		// nicht-nil Ergebnis gäbe es hier zwar nie — `FindStringSubmatch`
		// liefert entweder nil oder drei Gruppen —, und ein Zugriff auf
		// `teile[1]` hinter einer Prüfung, die das nicht sichert, ist eine
		// Zeile, die beim nächsten Umbau falsch wird.
		teile := stellenmuster.FindStringSubmatch(g.Fields[i].Name)
		if len(teile) == 0 || !vorhanden[teile[1]] {
			continue
		}

		nummer, fehler := strconv.Atoi(teile[2])
		if fehler != nil || nummer < 2 {
			// **Die Stelle 1 fällt weg**, siehe [Feldreihe]. Ein Fehler kann
			// hier nicht entstehen — das Muster lässt nur Ziffern durch —,
			// und ein Zweig ohne Fall wäre eine Behauptung (§3): deshalb
			// steht er in demselben `if` wie die Regel, die es wirklich gibt.
			continue
		}

		stellen[teile[1]] = append(stellen[teile[1]], Stelle{
			Feld: g.Fields[i].Name, Nummer: nummer,
		})
	}

	reihen := make([]Feldreihe, 0, len(stellen))

	for i := range g.Fields {
		gefunden, da := stellen[g.Fields[i].Name]
		if !da {
			continue
		}

		sort.Slice(gefunden, func(links, rechts int) bool {
			return gefunden[links].Nummer < gefunden[rechts].Nummer
		})

		reihen = append(reihen, Feldreihe{Grundfeld: g.Fields[i].Name, Stellen: gefunden})
	}

	return reihen
}
