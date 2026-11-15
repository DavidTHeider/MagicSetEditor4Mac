// **Die fünf Formatangaben aus dem Setkopf lesen** —
// #1325.
//
// Eigene Datei, weil `import.go` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort steht, **welcher** Block wohin geht, hier, was
// aus diesen fünf wird.

package service

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// kopfangaben ordnet den entkanonisierten Schlüsseln ihren Namen zu.
//
// **Als Zuordnung und nicht als Aufzählung im `switch`**: `importschluessel.go`
// führt dieselben fünf, und zwei Listen liefen auseinander. Wer hier einen
// Namen ergänzt, ergänzt ihn dort ebenso — `setkopfvertrag_test.go` hält beide
// zusammen.
var kopfangaben = map[string]string{
	"mse version":        "mse",
	"game version":       "spiel",
	"stylesheet version": "stil",
	"version control":    "vcs",
	"apprentice code":    "apprentice",
}

// kopfangabe gibt den Namen einer Kopfangabe — leer, wenn es keine ist.
func kopfangabe(schluessel string) string { return kopfangaben[schluessel] }

// kopfSetzen trägt einen Knoten in den Setkopf ein.
//
// **`version_control` ist ein Block, die übrigen vier sind Zeilen.** MSE2 hält
// dort ein Objekt (`VCSP vcs`, `src/data/set.hpp` Zeile 62, Fassung f9d9356
// vom 2021-07-06); wir nehmen seine Art. Steht dort nichts, bleibt das Feld
// leer — ein erfundenes `none` wäre eine Aussage über eine fremde Datei.
func kopfSetzen(kopf *domain.Setkopf, schluessel string, knoten *msepaket.Knoten) {
	switch kopfangabe(schluessel) {
	case "mse":
		kopf.MSEFassung = knoten.Wert
	case "spiel":
		kopf.Spielfassung = knoten.Wert
	case "stil":
		kopf.Stilfassung = knoten.Wert
	case "vcs":
		kopf.Versionskontrolle = wertVon(knoten.Ersten("type"))
	case "apprentice":
		kopf.Apprenticecode = knoten.Wert
	}
}
