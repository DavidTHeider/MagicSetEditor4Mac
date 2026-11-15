package domain

// **Die fünf Farben** — Werkstück M2 der Regelmaschine
// (#1604).
//
// # Warum eine eigene Datei für fünf Konstanten
//
// **Weil die Farbe keine Eigenschaft der Manakosten ist, sondern eine des
// Objekts.** CR 202.2 (Fassung 2026-08-19) leitet sie aus den Kosten ab, und
// CR 105.3 lässt Effekte sie ändern — wer sie in `manakosten.go` unterbrächte,
// hätte sie am falschen Ort, sobald es Effekte gibt (`M10`).

import "strings"

// Farbe ist eine der fünf — CR 105.1 (Fassung 2026-08-19).
//
// **Der Buchstabe ist der des Manasymbols** (CR 202.2a), und er ist damit
// derselbe, den MSE, MTG JSON und Scryfall schreiben. Ein eigener Schlüssel
// daneben wäre eine zweite Schreibweise für dieselbe Sache.
type Farbe string

const (
	FarbeWeiss   Farbe = "W"
	FarbeBlau    Farbe = "U"
	FarbeSchwarz Farbe = "B"
	FarbeRot     Farbe = "R"
	FarbeGruen   Farbe = "G"
)

// Farben gibt die fünf in der Reihenfolge des Regelwerks — CR 105.1.
//
// **WUBRG und nicht alphabetisch**: so zählt die CR sie auf, so stehen sie auf
// jeder Karte, und so sortiert MSE. Eine andere Reihenfolge wäre eine stille
// Abweichung (§15).
func Farben() []Farbe {
	return []Farbe{FarbeWeiss, FarbeBlau, FarbeSchwarz, FarbeRot, FarbeGruen}
}

// FarbeAus liest einen Farbbuchstaben — **groß oder klein**.
//
// **Klein auch**, weil MSE-Vorlagen beides schreiben: `casting_cost: "w"`
// kommt in gewachsenen Paketen vor, und §8 sagt, dass fremde Daten der
// Normalfall sind.
func FarbeAus(zeichen string) (Farbe, bool) {
	farbe := Farbe(strings.ToUpper(zeichen))

	for _, eine := range Farben() {
		if eine == farbe {
			return eine, true
		}
	}

	return "", false
}

// SortiereFarben gibt die Farben in der Reihenfolge des Regelwerks.
//
// **Ohne Wiederholung**: eine Karte mit zwei weißen Symbolen ist einmal weiß
// (CR 202.2c spricht von „zwei oder mehr **verschiedenen**").
func SortiereFarben(farben []Farbe) []Farbe {
	sortiert := make([]Farbe, 0, len(farben))

	for _, kandidat := range Farben() {
		for _, eine := range farben {
			if eine == kandidat {
				sortiert = append(sortiert, kandidat)

				break
			}
		}
	}

	return sortiert
}
