package domain

import "strings"

// **Die Ausrichtung einer Spalte der Kartenliste** — Werkstück 3 von
// Phase 3.7
// (#1168).
//
//	if      (f.second->card_list_align & ALIGN_RIGHT)  align = wxLIST_FORMAT_RIGHT;
//	else if (f.second->card_list_align & ALIGN_CENTER) align = wxLIST_FORMAT_CENTRE;
//	else                                               align = wxLIST_FORMAT_LEFT;
//
// `src/gui/control/card_list.cpp`, Zeilen 222–225, Fassung f9d9356 vom
// 2021-07-06. **Drei Werte und nicht neun**: MSEs `Alignment` ist ein
// Bitfeld über waagerecht *und* senkrecht (`src/util/alignment.hpp`), aber
// eine Listenspalte fragt nur nach der waagerechten — und in dieser
// Reihenfolge.

// Listenausrichtung sagt, wie eine Spalte ihren Wert ausrichtet.
type Listenausrichtung string

// Die drei, die es gibt.
//
// **`ListeLinks` ist die Vorgabe** — MSEs `else`-Zweig, und `Alignment`
// beginnt im Konstruktor bei `ALIGN_TOP_LEFT`.
const (
	ListeLinks  Listenausrichtung = "links"
	ListeRechts Listenausrichtung = "rechts"
	ListeMitte  Listenausrichtung = "mitte"
)

// ListenausrichtungAus liest `card list alignment:`.
//
// **Ein Wort in einer Angabe aus mehreren.** MSE schreibt dort Kombinationen
// wie `middle right` oder `top center`; gelesen wird die waagerechte Hälfte,
// und die senkrechte bleibt liegen — eine Zeile einer Liste ist so hoch wie
// ihr Text.
//
// **Ein unbekanntes Wort ist links.** §8: fremde Daten sind der Normalfall,
// und eine Spalte links ist kein Schaden.
func ListenausrichtungAus(angabe string) Listenausrichtung {
	for _, wort := range strings.Fields(strings.ToLower(angabe)) {
		switch wort {
		case "right":
			return ListeRechts
		case "center", "centre":
			return ListeMitte
		}
	}

	return ListeLinks
}
