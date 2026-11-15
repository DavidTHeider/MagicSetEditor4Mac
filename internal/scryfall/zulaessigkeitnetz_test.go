//go:build netzlauf || dienstdrift

package scryfall_test

// **Die vier Zulässigkeitswörter am echten Dienst** — Werkstück 5 von Phase
// 4.6 (#1293).
//
// # Was dieser Lauf hält, und was nicht
//
// Er hält, dass die vier Wörter aus `internal/scryfall/aufzaehlungen.go`
// **heute noch** die sind, die Scryfall schickt. Er hält **nicht**, dass es
// keine weiteren gibt — dafür müsste man jede Karte fragen. Genau deshalb ist
// ein unbekanntes Wort ein Befund und kein Fehler.
//
// Er läuft nicht in der Pipeline und zählt nicht zur Abdeckung (§3, §16).

import (
	"slices"
	"testing"
)

func TestDieZulaessigkeitswoerterAmEchtenDienst(t *testing.T) {
	klient := echterKlient(t)

	gesehen := map[string]bool{}

	// **Drei Karten mit absichtlich verschiedener Lage**: eine, die in Vintage
	// eingeschränkt und in Legacy verboten ist; eine, die alle vier Wörter
	// zeigt; und eine gewöhnliche.
	for _, name := range []string{"Black Lotus", "Channel", "Llanowar Elves"} {
		karte, _, _, fehler := klient.KarteNachName(t.Context(), name)
		if fehler != nil {
			t.Fatalf("%s: %v", name, fehler)
		}

		for _, wert := range karte.Legalitaeten {
			gesehen[wert] = true
		}
	}

	bekannt := []string{"legal", "not_legal", "restricted", "banned"}

	for wert := range gesehen {
		if !slices.Contains(bekannt, wert) {
			t.Errorf("Scryfall schickt %q — die Liste in aufzaehlungen.go kennt es nicht", wert)
		}
	}

	for _, wert := range bekannt {
		if !gesehen[wert] {
			t.Errorf("%q kam in keiner der drei Karten vor — die Messung deckt es nicht mehr", wert)
		}
	}
}
