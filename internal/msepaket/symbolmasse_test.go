package msepaket_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Eine Symbolschrift sagt, auf welche Größe ihre Bilder gerechnet sind.**
//
// # Der Befund
//
// Die Größe, mit der MSE ein Symbol zeichnet, ist **nicht** die Schriftgröße im
// Quadrat. Sie ist die tatsächliche Bildgröße, verhältnisgleich umgerechnet:
//
//	return wxSize(actual_size * (int) (size) / (int) (img_size));
//
// `img_size` steht als `image font size:` in der `symbol-font`-Datei — bei
// `magic-mana-small` **135** —, und einzelne `symbol:`-Blöcke überschreiben es.
// Dazu kommt je Symbol `horizontal space:` / `vertical space:`.
//
// Gelesen: `src/data/symbol_font.cpp`, Zeilen 50–52 und 147–153, Fassung
// f9d9356 vom 2021-07-06
// (#330).
func TestEineSymbolschriftNenntIhreBildgroesse(t *testing.T) {
	pfad := probesymbolschrift(t, `mse version: 2.0.0
image font size: 135
horizontal space: 2
vertical space: 1
symbol:
	code: W
	image: w.png
symbol:
	code: U
	image: u.png
	image font size: 120
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	schrift, befunde := paket.Symbolmodell()
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if schrift.Bildgroesse != 135 {
		t.Errorf("die Bildgröße ist %g, erwartet 135", schrift.Bildgroesse)
	}

	if schrift.AbstandWaagerecht != 2 || schrift.AbstandSenkrecht != 1 {
		t.Errorf("die Abstände sind %g/%g, erwartet 2/1",
			schrift.AbstandWaagerecht, schrift.AbstandSenkrecht)
	}

	// **Ein Symbol darf die Bildgröße überschreiben** — `magic-mana-small` tut
	// es für die eigenen Manasymbole (`image font size: 120`).
	w, _ := schrift.Symbol("W")
	if w.Bildgroesse != 0 {
		t.Errorf("W nennt %g, erwartet keine eigene Bildgröße", w.Bildgroesse)
	}

	u, _ := schrift.Symbol("U")
	if u.Bildgroesse != 120 {
		t.Errorf("U nennt %g, erwartet 120", u.Bildgroesse)
	}
}

// **Fehlt die Angabe, wird nichts erfunden.**
//
// Eine Null sagt „nicht angegeben", und der Aufrufer entscheidet. Eine
// geratene Zahl sähe aus wie eine gelesene.
func TestOhneAngabeBleibtDieBildgroesseNull(t *testing.T) {
	pfad := probesymbolschrift(t, "mse version: 2.0.0\nsymbol:\n\tcode: W\n\timage: w.png\n")

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	schrift, _ := paket.Symbolmodell()
	if schrift.Bildgroesse != 0 {
		t.Errorf("die Bildgröße ist %g, erwartet null", schrift.Bildgroesse)
	}

	// **Die Abstände dagegen haben eine Vorgabe, und sie ist gelesen**:
	// `spacing(1,1)` (`SymbolFont::SymbolFont`, `src/data/symbol_font.cpp`,
	// Zeile 28, Fassung f9d9356 vom 2021-07-06). Hier stand bis zum
	// 2026-10-11 null — jede Zeile, die ein Symbol hob, stand einen Punkt zu
	// niedrig (#2300).
	if schrift.AbstandWaagerecht != 1 || schrift.AbstandSenkrecht != 1 {
		t.Errorf("die Abstände sind %g/%g, erwartet 1/1",
			schrift.AbstandWaagerecht, schrift.AbstandSenkrecht)
	}
}

// **Eine Angabe, die keine Zahl ist, gilt als nicht angegeben.**
//
// Ein fremdes Paket kann alles enthalten (§8). `image font size: viel` ist
// keine Größe — und eine halb gelesene Zahl wäre schlimmer als keine.
func TestEineUnlesbareBildgroesseGiltAlsNichtAngegeben(t *testing.T) {
	pfad := probesymbolschrift(t,
		"mse version: 2.0.0\nimage font size: viel\nsymbol:\n\tcode: W\n\timage: w.png\n")

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if schrift, _ := paket.Symbolmodell(); schrift.Bildgroesse != 0 {
		t.Errorf("die Bildgröße ist %g, erwartet null", schrift.Bildgroesse)
	}
}

// probesymbolschrift legt ein Symbolschriftpaket mit dem gegebenen Inhalt an.
func probesymbolschrift(t *testing.T, inhalt string) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "probe.mse-symbol-font")
	if err := os.MkdirAll(pfad, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(pfad, "symbol-font"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return pfad
}
