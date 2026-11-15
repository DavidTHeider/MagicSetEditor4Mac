package msepaket_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// TestEinNachschlagTrifftBeideSchreibweisen ist der Kern von #396, eine Ebene
// tiefer.
//
// **Beide Schreibweisen kommen wirklich vor**, gemessen am 2026-08-24:
// die 213 Vorlagen in `M15-Magic-Pack` schreiben `card field:` mit Leerzeichen
// und **keine einzige** mit Unterstrich; die zehn kanonischen Sets schreiben
// `casting_cost:`, `rule_text:`, `set_info:` mit Unterstrich und keines mit
// Leerzeichen. MSE trifft beide, weil sein Leser jeden Schlüssel kanonisch
// macht (`src/util/io/reader.cpp`, Zeile 217, Fassung f9d9356 vom 2021-07-06).
func TestEinNachschlagTrifftBeideSchreibweisen(t *testing.T) {
	t.Parallel()

	baum, err := msepaket.Zerteile("probe", []byte(
		"set_info:\n\ttitle: Der Titel\n"+
			"card field:\n\tname: eins\n"+
			"card_field:\n\tname: zwei\n"))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	t.Run("Unterstrich in der Datei, Leerzeichen im Nachschlag", func(t *testing.T) {
		t.Parallel()

		knoten := baum.Ersten("set info")
		if knoten == nil {
			t.Fatal("set_info: nicht gefunden")
		}

		if titel := knoten.Ersten("title"); titel == nil || titel.Wert != "Der Titel" {
			t.Errorf("der Titel: %v", titel)
		}
	})

	t.Run("beide Schreibweisen zaehlen als derselbe Schluessel", func(t *testing.T) {
		t.Parallel()

		if hat := len(baum.Alle("card field")); hat != 2 {
			t.Errorf("card field: %d gefunden, will 2", hat)
		}
	})

	t.Run("der Schluessel bleibt, wie er in der Datei stand", func(t *testing.T) {
		t.Parallel()

		if hat := baum.Ersten("set info").Schluessel; hat != "set_info" {
			t.Errorf("der Schlüssel ist %q — der Rundlauf braucht ihn roh", hat)
		}
	})
}

// TestEineKurveAusEinerVonMseGeschriebenenDatei ist die erste sichtbare Folge.
//
// **MSE schreibt `line_after`, nicht `line after`.** Sein Schreiber gibt den
// Namen aus, wie er im Code steht, ohne Rückumwandlung
// (`src/util/io/writer.cpp`, Zeile 56, Fassung f9d9356 vom 2021-07-06) — und im
// Code steht `REFLECT_N("line_after", segment_after)` (`src/data/symbol.cpp`,
// Zeile 29). Jede Bézierkurve eines echten Symbols wäre als Gerade gelesen
// worden.
func TestEineKurveAusEinerVonMseGeschriebenenDatei(t *testing.T) {
	t.Parallel()

	baum, err := msepaket.Zerteile("probe.mse-symbol", []byte(
		"mse_version: 2.0.0\n"+
			"part:\n\ttype: shape\n\tname: Bogen\n\tcombine: overlap\n"+
			"\tpoint:\n\t\tposition: (0,0)\n\t\tlock: free\n"+
			"\t\tline_after: curve\n"+
			"\t\thandle_before: (-0.1,0)\n\t\thandle_after: (0.1,0)\n"+
			"\tpoint:\n\t\tposition: (1,1)\n\t\tlock: free\n\t\tline_after: line\n"))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	zeichnung, befunde := msepaket.Symbolzeichnungsmodell(baum)
	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	punkt := zeichnung.Teile[0].Punkte[0]

	if !punkt.KurveDanach {
		t.Error("die Strecke nach dem ersten Punkt ist eine Gerade — MSE schreibt dort curve")
	}

	if punkt.GriffDanach != (msepaket.Vektor{X: 0.1}) {
		t.Errorf("der Griff danach ist %v, will (0.1,0)", punkt.GriffDanach)
	}
}
