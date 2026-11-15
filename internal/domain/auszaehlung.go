package domain

// **Aus Werten werden Gruppen** — die Auszählung nach einer erklärten
// Dimension (#1338,
// Werkstück C von Phase 4.11).
//
// # ported: `GraphData::GraphData`
//
// `src/gui/control/graph.cpp`, Zeilen 85–135, dazu `GraphAxis::addGroup`
// (24–32) und `to_bin` (67–73), Fassung f9d9356 vom 2021-07-06.
//
// **Hier steht das Zählen und nicht das Rechnen.** Welchen Wert eine Karte
// hat, sagt ein MSE-Script — und `domain` kennt keinen Interpreter (§2). Was
// hier ankommt, ist die fertige Liste der Werte, eine Zeichenkette je Karte;
// was daraus wird, ist eine fachliche Regel über mehrere Karten und gehört
// deshalb hierher.

import (
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// Auszaehlart sind die Angaben einer Dimension, soweit das Zählen sie braucht.
type Auszaehlart struct {
	// Numerisch liest die Werte als Zahl und ordnet sie als solche.
	Numerisch bool

	// Klassenbreite fasst benachbarte Zahlen zusammen — 0 heißt: nicht.
	Klassenbreite float64

	// Gruppen ist die **erklärte** Reihenfolge. Leer heißt: klug ordnen.
	Gruppen []string

	// Farben ordnet einem Gruppennamen seine Farbe zu, in MSEs Schreibweise.
	Farben map[string]string
}

// Auszaehlgruppe ist ein Balken: ein Wert und wie viele Karten ihn tragen.
type Auszaehlgruppe struct {
	Name   string `json:"name"`
	Karten int    `json:"karten"`

	// Farbe ist die erklärte Farbe — leer, wo keine dasteht.
	Farbe string `json:"farbe"`

	// Kartenkennungen sind die Karten des Balkens, aufsteigend und je Karte
	// einmal — UXR-F6 (#2013).
	//
	// **Nur, wo die Zählung die Karten kannte** ([AuszaehlenKarten]); wer nur
	// Werte zählt, bekommt hier `nil`. `Karten` und die Länge dieser Liste
	// sind nicht dasselbe: der Balken zählt **Elemente** wie MSE2s
	// `GraphGroup::size`, die Liste nennt **Karten** wie `GraphData::indices`.
	// Sie unterscheiden sich, wenn eine Karte mit `split list:` zweimal in
	// derselben Gruppe liegt.
	Kartenkennungen []int64 `json:"-"`
}

// Dimensionswert ist der Wert, den eine Dimension an einer Karte hat.
//
// **Der Name sagt, wessen Wert es ist**: [Kartenwert] ist schon vergeben und
// meint Stärke und Widerstand einer Karte im Spiel.
type Dimensionswert struct {
	Karte int64
	Wert  string
}

// gezaehlt hält, wie viele Elemente einen Wert tragen — und welche Karten.
type gezaehlt struct {
	elemente int
	karten   []int64
}

// Auszaehlen zählt die Werte und ordnet sie.
//
// **Drei Ordnungen, in MSEs Reihenfolge**: numerisch, sonst die erklärte,
// sonst [Vergleiche]. Die Reihenfolge ist die des Originals — eine erklärte
// Reihenfolge an einer numerischen Achse ist dort ohne Wirkung.
func Auszaehlen(werte []string, art Auszaehlart) []Auszaehlgruppe {
	zaehlung := map[string]gezaehlt{}
	for _, wert := range werte {
		eintrag := zaehlung[wert]
		eintrag.elemente++
		zaehlung[wert] = eintrag
	}

	return gruppiert(zaehlung, art)
}

// AuszaehlenKarten zählt wie [Auszaehlen] und merkt sich dabei die Karten.
//
// UXR-F6 (#2013).
// MSE2 behält zu jedem Element den `original_index` der Karte
// (`GraphDataElement`, `src/gui/control/graph.hpp`) und gibt nach einer Wahl
// die Indizes der Gruppe zurück (`GraphData::indices`, `graph.cpp` Zeilen
// 249–258, Fassung f9d9356 vom 2021-07-06); hier nennt jede Gruppe ihre
// Karten gleich mit. **Ordnung, Klassen und leere Gruppen sind dieselben**
// — es ist derselbe Kern.
func AuszaehlenKarten(werte []Dimensionswert, art Auszaehlart) []Auszaehlgruppe {
	zaehlung := map[string]gezaehlt{}
	for _, wert := range werte {
		eintrag := zaehlung[wert.Wert]
		eintrag.elemente++
		eintrag.karten = append(eintrag.karten, wert.Karte)
		zaehlung[wert.Wert] = eintrag
	}

	gruppen := gruppiert(zaehlung, art)

	for stelle := range gruppen {
		gruppen[stelle].Kartenkennungen = einzeln(gruppen[stelle].Kartenkennungen)
	}

	return gruppen
}

// einzeln ordnet Kennungen aufsteigend und lässt jede einmal stehen.
//
// **Auch dort, wo keine dasteht, kommt eine Liste** und kein `nil`: wer
// [AuszaehlenKarten] ruft, will wissen, dass die Gruppe **keine** Karte hat,
// und nicht, dass die Karten unbekannt sind.
func einzeln(kennungen []int64) []int64 {
	geordnet := slices.Clone(kennungen)
	slices.Sort(geordnet)

	return append([]int64{}, slices.Compact(geordnet)...)
}

// gruppiert ordnet die gezählten Werte zu Gruppen und gibt ihnen ihre Farbe.
func gruppiert(zaehlung map[string]gezaehlt, art Auszaehlart) []Auszaehlgruppe {
	var gruppen []Auszaehlgruppe

	switch {
	case art.Numerisch:
		gruppen = numerischGeordnet(zaehlung, art.Klassenbreite)
	case len(art.Gruppen) > 0:
		gruppen = erklaertGeordnet(zaehlung, art.Gruppen)
	default:
		gruppen = klugGeordnet(zaehlung)
	}

	// **Kanonisch auf beiden Seiten** (#1980) — eine Abweichung von MSE mit
	// Absicht: `GraphData::GraphData` schlägt die kanonisch gelesene Tafel
	// mit dem rohen Gruppennamen nach (`src/gui/control/graph.cpp`, Zeile
	// 138, Fassung f9d9356 vom 2021-07-06), und `multi color` fände dort nie
	// seine Farbe. Beim Wahlbild und beim Sprachpaket vergleicht MSE selbst
	// auf beiden Seiten kanonisch.
	farben := make(map[string]string, len(art.Farben))
	for name, farbe := range art.Farben {
		farben[strings.ReplaceAll(name, " ", "_")] = farbe
	}

	for stelle := range gruppen {
		gruppen[stelle].Farbe = farben[strings.ReplaceAll(gruppen[stelle].Name, " ", "_")]
	}

	return gruppen
}

// erklaertGeordnet gibt genau die erklärten Gruppen, auch die leeren.
//
// **Auch die leeren, und das ist die Absicht**: MSEs `counts[gn]` legt für
// einen unbekannten Namen eine Null an. Eine erklärte Farbe ohne Karten ist
// eine Aussage über das Set und kein Grund, die Zeile wegzulassen.
func erklaertGeordnet(zaehlung map[string]gezaehlt, ordnung []string) []Auszaehlgruppe {
	gruppen := make([]Auszaehlgruppe, 0, len(ordnung))
	for _, name := range ordnung {
		gruppen = anhaengen(gruppen, name, zaehlung[name])
	}

	return gruppen
}

// klugGeordnet ordnet nach [Vergleiche] — MSEs `smart_less`.
func klugGeordnet(zaehlung map[string]gezaehlt) []Auszaehlgruppe {
	namen := make([]string, 0, len(zaehlung))
	for name := range zaehlung {
		namen = append(namen, name)
	}

	sort.Slice(namen, func(a, b int) bool {
		return Vergleiche(namen[a], namen[b]) < 0
	})

	gruppen := make([]Auszaehlgruppe, 0, len(namen))
	for _, name := range namen {
		gruppen = anhaengen(gruppen, name, zaehlung[name])
	}

	return gruppen
}

// numerischGeordnet ordnet die Zahlen und lässt den Rest stehen.
//
// **Die Lücke dazwischen bekommt einen leeren Balken** — MSEs
// `add 0 bars before this value`. Ein Balken, der fehlt, sähe aus wie einer,
// den niemand erklärt hat; einer mit null sagt „hier ist nichts".
func numerischGeordnet(zaehlung map[string]gezaehlt, breite float64) []Auszaehlgruppe {
	namen := make([]string, 0, len(zaehlung))
	for name := range zaehlung {
		namen = append(namen, name)
	}

	sort.Slice(namen, func(a, b int) bool {
		return Vergleiche(namen[a], namen[b]) < 0
	})

	gruppen := make([]Auszaehlgruppe, 0, len(namen))
	vorige := 0

	for _, name := range namen {
		zahl, fehler := strconv.ParseFloat(name, 64)
		if fehler != nil {
			// **Kein Zahlwert zählt trotzdem** — MSEs `non-numeric, add
			// anyway`. Eine Achse, die `X` verlöre, verlöre die Karten dazu.
			gruppen = anhaengen(gruppen, name, zaehlung[name])

			continue
		}

		naechste := int(math.Floor(zahl))
		for i := vorige; i < naechste; i++ {
			gruppen = anhaengen(gruppen, Klassenname(float64(i), breite), gezaehlt{})
		}

		vorige = naechste + 1

		gruppen = anhaengen(gruppen, Klassenname(zahl, breite), zaehlung[name])
	}

	return gruppen
}

// anhaengen legt eine Gruppe an — oder zählt zur vorigen dazu.
//
// **Dasselbe wie MSEs `addGroup`**: trägt die letzte Gruppe denselben Namen,
// wächst sie. Daran hängt die Klassenbildung — `1` und `4` landen beide in
// `1–5`, und sie stehen nebeneinander, weil die Liste geordnet ist.
func anhaengen(gruppen []Auszaehlgruppe, name string, zaehlung gezaehlt) []Auszaehlgruppe {
	if letzte := len(gruppen) - 1; letzte >= 0 && gruppen[letzte].Name == name {
		gruppen[letzte].Karten += zaehlung.elemente
		gruppen[letzte].Kartenkennungen = append(gruppen[letzte].Kartenkennungen, zaehlung.karten...)

		return gruppen
	}

	return append(gruppen, Auszaehlgruppe{
		Name: name, Karten: zaehlung.elemente, Kartenkennungen: zaehlung.karten,
	})
}

// Klassenname beschriftet eine Klasse — MSEs `to_bin`.
//
//	int bin = ceil(value / bin_size);
//	String::Format("%.0f%c%.0f", (bin-1) * bin_size + 1, EN_DASH, bin * bin_size)
//
// **Der Strich ist ein Halbgeviertstrich** (`EN_DASH`), kein Bindestrich.
// **Und die Null bleibt die Null**: `bin_size <= 0 || value == 0` gibt die
// Zahl selbst — sonst hieße „keine Manakosten" `1–5`.
func Klassenname(wert, breite float64) string {
	if breite <= 0 || wert == 0 {
		return strconv.Itoa(int(wert))
	}

	klasse := math.Ceil(wert / breite)

	return strconv.Itoa(int((klasse-1)*breite+1)) + "–" + strconv.Itoa(int(klasse*breite))
}
