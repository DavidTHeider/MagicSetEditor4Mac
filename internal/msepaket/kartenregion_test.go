package msepaket_test

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Stil erklärt seine Kartenregionen** — N1 des Nachlaufs über die
// QA-Arbeit (#2039, Vermerk 0165).
//
// `card region:` steht je Region einmal im Kopf, mit `name`, `x`, `y`,
// `width` und `height` (`IMPLEMENT_REFLECTION(CardRegion)`,
// `src/data/card_region.cpp`, Zeilen 19–24, in
// `G-e-n-e-v-e-n-s-i-S/MagicSetEditor2`, Fassung a4b79c9 vom 2026-09-04).
// Acht Stile des Packs tun es — zweiseitige Karten, `meld 3in1` —, und
// `magic-lackeyccg` schneidet danach mehrere Bilder je Karte zu. `x` und `y`
// sind dort `double`, `width` und `height` `int`.
//
// Geprüft gegen ein **eigenes** Paket in `testdaten/` (§3, §11): dieselbe
// Syntax, eigener Inhalt.
func TestEinStilErklaertSeineKartenregionen(t *testing.T) {
	t.Parallel()

	paket, err := msepaket.Oeffne(filepath.Join("..", "..", "testdaten", "regionen.mse-style"))
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

	erwartet := []msepaket.Kartenregion{
		{Name: "front", X: 0, Y: 0, Breite: 375, Hoehe: 523},
		{Name: "back", X: 377.5, Y: 0, Breite: 375, Hoehe: 523},
		// Was fehlt oder keine Zahl ist, ist null — wie bei jeder Zahl des
		// Kopfes ([zahlOderNull]); eine geratene sähe aus wie eine gelesene.
		{Name: "lücke", X: 0, Y: 12.5, Breite: 0, Hoehe: 40},
	}

	if len(stil.Regionen) != len(erwartet) {
		t.Fatalf("%+v, erwartet %+v", stil.Regionen, erwartet)
	}

	for stelle := range erwartet {
		if stil.Regionen[stelle] != erwartet[stelle] {
			t.Errorf("Region %d: %+v, erwartet %+v", stelle, stil.Regionen[stelle], erwartet[stelle])
		}
	}
}

// **Ein Stil ohne Regionen hat keine** — und kein leeres Element.
func TestEinStilOhneRegionenHatKeine(t *testing.T) {
	t.Parallel()

	if regionen := stilmodell(t).Regionen; len(regionen) != 0 {
		t.Errorf("%+v", regionen)
	}
}
