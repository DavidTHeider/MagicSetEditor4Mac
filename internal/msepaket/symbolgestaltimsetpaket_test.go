package msepaket_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Eine Symbolzeichnung im Setpaket** — Befund des Zweitblicks durch Codex vom
// 2026-08-21.
//
// # Der Befund
//
// [msepaket.Bildquelle.Bild] kennt den [msepaket.SetpaketVerweis] und öffnet
// das Paket an dem Ort, den er nennt (`bildquelle.go`, Zeilen 124–131).
// `Symbolgestalt` daneben reichte ihn ungeprüft an `ZerlegePaketpfad` weiter —
// und das verlangt einen führenden `/` (`einbinden.go`, Zeilen 28–34). Ergebnis
// war **immer** eine leere Zeichnung.
//
// `set.symbol` nennt eine `.mse-symbol` **im Set**, und `symbol_variation`
// zeichnet sie: MSE liest sie aus dem lokalen Paket
// (`SymbolToImage::generate`, `src/gfx/generated_image.cpp`, Zeilen 453–462,
// Fassung f9d9356 vom 2021-07-06).
//
// # Warum es niemand gemerkt hat
//
// **Kein einziges der sechs kanonischen Sets trägt ein Setsymbol** — nachgesehen
// am 2026-08-21 in `White`, `Black`, `Green`, `Red`, `Archenemy` und
// `Vanguard`: keine `symbol:`-Zeile, keine `.mse-symbol`-Datei. Ein leerer
// Verweis gibt MSEs Standardsymbol, und das zeichnet.
//
// Der Test der Ausfuhr benutzte eine Attrappe, die für **jeden** Verweis
// antwortete — sie verdeckte die Naht (§3).
func TestEineSymbolzeichnungImSetpaket(t *testing.T) {
	t.Parallel()

	ort := filepath.Join(t.TempDir(), "Probe.mse-set")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatal(err)
	}

	// Ein Dreieck — genug, damit das Modell etwas zu lesen hat.
	zeichnung := "mse version: 0.3.7\n" +
		"part:\n" +
		"\ttype: shape\n" +
		"\tname: Dreieck\n" +
		"\tpoint:\n" +
		"\t\tposition: (0,0)\n" +
		"\tpoint:\n" +
		"\t\tposition: (1,0)\n" +
		"\tpoint:\n" +
		"\t\tposition: (0,1)\n"

	if err := os.WriteFile(filepath.Join(ort, "symbol.mse-symbol"),
		[]byte(zeichnung), 0o600); err != nil {
		t.Fatal(err)
	}

	quelle := msepaket.NeueBildquelle(t.TempDir())

	gestalt := quelle.Symbolgestalt(msepaket.SetpaketVerweis(ort, "symbol.mse-symbol"))
	if len(gestalt.Teile) == 0 {
		t.Fatal("die Zeichnung im Setpaket wurde nicht gelesen")
	}

	// **Und ein gewöhnlicher Paketpfad geht weiter wie bisher.**
	if leer := quelle.Symbolgestalt("/gibtsnicht.mse-style/s.mse-symbol"); len(leer.Teile) != 0 {
		t.Errorf("ein unbekannter Pfad gab %d Teile", len(leer.Teile))
	}
}

// **Die drei Wege, auf denen eine Zeichnung im Setpaket leer bleibt.**
//
// Dieselbe Regel wie beim gewöhnlichen Paketpfad: **was nicht lesbar ist, ist
// leer**, und der Rasterer benennt es. Ein Fehler hier bräche den ganzen Plan
// für eine Datei, die auf den meisten Karten gar nicht gebraucht wird (§8).
func TestEineUnlesbareZeichnungImSetpaketBleibtLeer(t *testing.T) {
	t.Parallel()

	ort := filepath.Join(t.TempDir(), "Probe.mse-set")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatal(err)
	}

	// **Eine Zeile ohne Schlüssel** — der Zerteiler weist sie ab
	// („der Schlüssel ist leer", `zerteiler.go`, Zeile 351).
	if err := os.WriteFile(filepath.Join(ort, "kaputt.mse-symbol"),
		[]byte(": ohne Schlüssel\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	quelle := msepaket.NeueBildquelle(t.TempDir())

	faelle := []struct{ name, verweis string }{
		{
			// **Ein Verweis ohne Ort.** `pruefeSetpaketverweis` weist ihn ab,
			// bevor irgendetwas auf die Platte geht.
			name:    "ohne Ort",
			verweis: msepaket.SetpaketVerweis("", "symbol.mse-symbol"),
		},
		{
			name:    "die Datei liegt nicht da",
			verweis: msepaket.SetpaketVerweis(ort, "gibtsnicht.mse-symbol"),
		},
		{
			name:    "die Datei ist kein Baum",
			verweis: msepaket.SetpaketVerweis(ort, "kaputt.mse-symbol"),
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if gestalt := quelle.Symbolgestalt(fall.verweis); len(gestalt.Teile) != 0 {
				t.Errorf("gelesen wurden %d Teile", len(gestalt.Teile))
			}
		})
	}
}
