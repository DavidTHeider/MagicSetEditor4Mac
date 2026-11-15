package service_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die vier Klassen aus #88, gemessen statt behauptet.**
//
// # Warum dieser Test hier steht
//
// #88 zählt vier
// Fälle auf, in denen ein Vorkommen weder gespeichert noch als Rohblock
// mitgeführt wird. Die Aufzählung stand seit Phase 2 in einem Kommentar — und
// ein Kommentar wird nicht rot, wenn sich etwas ändert.
//
// Das Issue sagt selbst, was der erste Schritt ist: **der Schreiber, nicht die
// Reparatur.** Den gibt es seit #48, und er ist byte-genau — weil die
// Quelldatei mitreist ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)).
// Damit ist die Frage eine andere geworden: nicht mehr „geht beim Schreiben
// etwas verloren", sondern **„ist der Bestand blind"**.
//
// Blind heißt: Suche, Statistik und der MCP-Server sehen es nicht. Genau das
// misst dieser Test.
func TestDieVierKlassenAusAchtundachtzig(t *testing.T) {
	speicher := &speicherDouble{}

	if _, err := service.NeuerImporteur(speicher).
		Set(t.Context(), filepath.Join(testdaten, "vierklassen.mse-set")); err != nil {
		t.Fatalf("Import: %v", err)
	}

	gesehen := make([]string, 0, len(speicher.roh))
	for _, block := range speicher.roh {
		gesehen = append(gesehen, block.Text)
	}

	// **Auch die Setangaben und Karten zählen als „gesehen".** Die Frage ist
	// nicht, wo etwas ankommt, sondern **ob**.
	for _, wert := range speicher.set.Values {
		gesehen = append(gesehen, wert.Field+": "+wert.Text)
	}

	for _, karte := range speicher.set.Cards {
		for _, wert := range karte.Values {
			gesehen = append(gesehen, wert.Field+": "+wert.Text)
		}
	}

	roh := strings.Join(gesehen, "\n")

	faelle := map[string]string{
		"ein zweites game":                  "zweites-spiel",
		"ein zweites set info":              "zweiter Block",
		"ein Kommentar auf der Wurzelebene": "# ein Kommentar auf der Wurzelebene",
		"der skalare Rest eines card":       "skalarer Rest",
	}

	for name, erwartet := range faelle {
		t.Run(name, func(t *testing.T) {
			if !strings.Contains(roh, erwartet) {
				t.Errorf("%q kommt im Bestand nicht an:\n%s", erwartet, roh)
			}
		})
	}
}
