// **Eine Stiloption hat eine Wirkung, und Liste und Karte sagen dasselbe**
// (F05, #2247;
// gefunden beim Beleg für Befund 1 des Nachlaufs über Phase G, zweiter Teil).
//
// # Was schiefging
//
// Zwei Wahrheiten für eine Frage. Die Rasterung las den `styling:`-Block der
// Datei, die Optionsliste die hier gesetzten Werte — und keine die andere.
// Eine hier gesetzte Option änderte die Karte nicht (das Bild blieb Byte für
// Byte gleich), und ein Set aus einer Datei zeichnete »dunkel«, während die
// Liste »hell« nannte.
//
// MSE2 kennt nur eine: `Set::stylingDataFor` (`src/data/set.cpp`, Zeilen
// 99–105, Fassung f9d9356 vom 2021-07-06) — gelesen von der Rasterung wie
// vom Stilfenster.
package flaechen_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEineGesetzteStiloptionAendertDieKarte(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := neuesSet(t, umgebung, "Wirkung")

	mussKommandozeile(t, umgebung, "karte", setID, "anlegen")
	karteID := ersteKarte(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))

	ort := t.TempDir()
	vorher, nachher := filepath.Join(ort, "vorher.png"), filepath.Join(ort, "nachher.png")

	mussKommandozeile(t, umgebung, "zeichnen", karteID, vorher)
	mussKommandozeile(t, umgebung, "stiloption", setID, "setzen", "hintergrund", "dunkel")
	mussKommandozeile(t, umgebung, "zeichnen", karteID, nachher)

	if bytes.Equal(bildbytes(t, vorher), bildbytes(t, nachher)) {
		t.Error("die gesetzte Option hat die Karte nicht verändert")
	}
}

func TestDieListeNenntDenWertAusDerDatei(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	quelle := filepath.Join(t.TempDir(), "dunkel.mse-set")
	if fehler := os.MkdirAll(quelle, 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	inhalt := "mse version: 2.0.0\ngame: probe\nstylesheet: schlicht\nset info:\n\ttitle: Dunkel\n" +
		"card:\n\tname: Feuerball\nstyling:\n\tprobe-schlicht:\n\t\thintergrund: dunkel\n"
	if fehler := os.WriteFile(filepath.Join(quelle, "set"), []byte(inhalt), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	setID := setkennungAus(t, mussKommandozeile(t, umgebung, "oeffnen", quelle, "--json"))

	for _, zeile := range strings.Split(mussKommandozeile(t, umgebung, "stiloption", setID), "\n") {
		if strings.Contains(zeile, "hintergrund") && !strings.HasSuffix(strings.TrimSpace(zeile), "dunkel") {
			t.Errorf("die Liste nennt nicht den Wert der Datei: %q", zeile)
		}
	}
}

// bildbytes liest ein geschriebenes Bild.
func bildbytes(t *testing.T, pfad string) []byte {
	t.Helper()

	inhalt, fehler := os.ReadFile(pfad) // #nosec G304 -- aus t.TempDir()
	if fehler != nil {
		t.Fatalf("Bild lesen: %v", fehler)
	}

	return inhalt
}
