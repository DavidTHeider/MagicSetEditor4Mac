package msepaket_test

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Der `styling style:`-Block** — die Sichtseite der Stiloptionen
// (#1225).
//
// # Was er ist
//
//	vector<FieldP> styling_fields;
//	IndexMap<FieldP, StyleP> styling_style;
//	…
//	REFLECT(styling_fields);
//	REFLECT_IF_READING styling_style.init(styling_fields);
//	REFLECT(styling_style);
//
// `src/data/stylesheet.hpp`, Zeilen 47–50, und `src/data/stylesheet.cpp`,
// Zeilen 103–105, Fassung f9d9356 vom 2021-07-06. Er verhält sich zu
// `styling field:` wie `card style:` zu `card field:` — dieselbe Gestalt,
// dieselbe Lesart, ein anderer Gegenstand.
//
// **Und darin steht das, was die Wahl zeigt**: über
// `full-magic-pack/data` tragen **303 von 376** Stilpaketen dort
// `choice images:`-Blöcke, zusammen 1.097 mit 2.707 Einträgen (gezählt am
// 2026-08-27).
func TestDerStilBringtDieSichtseiteSeinerOptionenMit(t *testing.T) {
	stil := saatstil(t, "probe-schlicht.mse-style")

	hintergrund, da := stil.Optionsstil("hintergrund")
	if !da {
		t.Fatalf("kein Optionsstil `hintergrund`: %+v", stil.Optionsstile)
	}

	if len(hintergrund.Wahlbilder) != 2 {
		t.Fatalf("%d Wahlbilder: %+v", len(hintergrund.Wahlbilder), hintergrund.Wahlbilder)
	}

	if hintergrund.Wahlbilder[0].Wahl != "hell" ||
		hintergrund.Wahlbilder[0].Bild != "hell.png" {
		t.Errorf("erstes Wahlbild: %+v", hintergrund.Wahlbilder[0])
	}
}

// **Ein Optionsstil, den es nicht gibt, wird als solcher gemeldet.**
//
// Nicht als leerer Feldstil: `magic-m15` erklärt dreizehn Optionen und gibt
// nur vieren ein Bild. Wer die Unterscheidung verlöre, zeichnete für die
// übrigen neun eine leere Bildreihe statt gar keiner.
func TestEinOptionsstilOhneEintragFehltSichtbar(t *testing.T) {
	stil := saatstil(t, "probe-schlicht.mse-style")

	if _, da := stil.Optionsstil("titelfarbe"); da {
		t.Fatalf("`titelfarbe` hat keinen Optionsstil und meldet doch einen")
	}
}

// saatstil öffnet ein Stilpaket der Saat und gibt sein Modell.
func saatstil(t *testing.T, paketname string) msepaket.Stil {
	t.Helper()

	paket, err := msepaket.Oeffne(filepath.Join("..", "..", "saat", paketname))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, befunde := paket.Stilmodell()
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	return stil
}
