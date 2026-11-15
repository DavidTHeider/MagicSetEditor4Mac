package msepaket_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Das Bild eines Symbols darf ein eingerückter `script:`-Block sein.**
//
// So schreibt `magic-mana-small.mse-symbol-font` seine zehn Dreifach-Hybriden
// (Zeilen 300–408):
//
//	symbol:
//		code: W/U/B
//		image:
//			script:
//				masked_blend(light: "mana_www.png", …)
//
// Wer nur `knoten.Wert` liest, bekommt die leere Zeichenkette — und meldet „das
// Symbol W/U/B nennt kein Bild". Über die kanonischen Sets waren das **2.500
// von 2.794** Befunden, zehn Meldungen auf **jeder** Karte
// (#405).
//
// **Dieselbe Klasse zum vierten Mal** — `left:` (#295), `font: color:` (#382),
// `font: name:` (#393).
func TestEinSymbolbildDarfEinSkriptblockSein(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "probe-hybrid.mse-symbol-font")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\ngame: probe\nshort name: Hybrid\n" +
		"symbol:\n\tcode: W/U\n\timage:\n\t\tscript:\n" +
		"\t\t\tmasked_blend(\n\t\t\t\tlight: \"w.png\",\n\t\t\t\tdark: \"u.png\",\n" +
		"\t\t\t\tmask: \"m.png\")\n\timage font size: 145\n"

	if err := os.WriteFile(filepath.Join(ort, "symbol-font"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	paket, err := msepaket.Oeffne(ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	schrift, befunde := paket.Symbolmodell()

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "nennt kein Bild") {
			t.Errorf("Befund %q — das Bild steht als Skriptblock da", befund.Meldung)
		}
	}

	if len(schrift.Symbole) != 1 {
		t.Fatalf("%d Symbole, erwartet eines", len(schrift.Symbole))
	}

	if !strings.Contains(schrift.Symbole[0].Bild, "masked_blend") {
		t.Errorf("das Bild ist %q, erwartet den Skriptrumpf", schrift.Symbole[0].Bild)
	}

	// **Geklammert**, denn `service` streift die Klammer ab und rechnet, was
	// dazwischen steht — dieselbe Naht wie bei jeder anderen Angabe.
	if !strings.HasPrefix(strings.TrimSpace(schrift.Symbole[0].Bild), "{") {
		t.Errorf("das Bild ist %q, erwartet eine geklammerte Angabe", schrift.Symbole[0].Bild)
	}
}

// **Ein Symbol ohne `image:` wird weiterhin benannt** — die Gegenprobe zum
// Skriptblock: der Weg darf die Meldung nicht mitnehmen, die er beheben soll.
func TestEinSymbolOhneBildWirdWeiterhinBenannt(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "probe-ohne.mse-symbol-font")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\ngame: probe\nshort name: Ohne\nsymbol:\n\tcode: W\n"
	if err := os.WriteFile(filepath.Join(ort, "symbol-font"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	paket, err := msepaket.Oeffne(ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	_, befunde := paket.Symbolmodell()

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "nennt kein Bild") {
			return
		}
	}

	t.Errorf("kein Befund über das fehlende Bild: %v", befunde)
}
