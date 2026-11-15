package msepaket

// Kartenregion ist ein `card region:` eines Stils — ein Ausschnitt der
// gezeichneten Karte mit Namen, etwa die Vorder- und die Rückseite einer
// zweiseitigen Karte nebeneinander.
//
// **Aus MSE 2.5 und nicht aus der Referenz f9d9356**: dort gibt es keine
// Regionen. `vector<CardRegionP> card_regions` (`src/data/stylesheet.hpp`,
// Zeile 44), je Region `name`, `x`, `y` als `double` und `width`, `height`
// als `int` (`src/data/card_region.hpp` und `card_region.cpp`, Zeilen 19–24,
// in `G-e-n-e-v-e-n-s-i-S/MagicSetEditor2`, Fassung a4b79c9 vom 2026-09-04).
// Acht Stile des Packs erklären welche; Exportvorlagen schneiden danach mehrere
// Bilder je Karte zu (Vermerk 0165).
type Kartenregion struct {
	Name   string
	X, Y   float64
	Breite int
	Hoehe  int
}

// kartenregionenAus liest jedes `card region:` des Kopfes in der Reihenfolge,
// in der es dasteht — die Vorlagen zählen sie über `card_regions[i]`.
func kartenregionenAus(kopf *Baum) []Kartenregion {
	alle := kopf.Alle("card region")
	regionen := make([]Kartenregion, 0, len(alle))

	for _, knoten := range alle {
		regionen = append(regionen, Kartenregion{
			Name:   kinderwert(knoten, "name"),
			X:      zahlOderNull(knoten.Ersten("x")),
			Y:      zahlOderNull(knoten.Ersten("y")),
			Breite: zahl(kinderwert(knoten, "width")),
			Hoehe:  zahl(kinderwert(knoten, "height")),
		})
	}

	return regionen
}

// kinderwert gibt den Wert eines Kindes; leer, wenn es fehlt.
func kinderwert(knoten *Knoten, schluessel string) string {
	if kind := knoten.Ersten(schluessel); kind != nil {
		return kind.Wert
	}

	return ""
}
