//go:build fremdpakete

package flaechen_test

// **Die eigene Saat des Setdatei-Rundlaufs** — F05
// (#2247):
// Sets mit **eigenem** Inhalt im fremden Format, für zwei Spiele, damit der
// Rundlauf mehr als eine Paketart sieht. Fremde Inhalte stehen in den
// kanonischen Sets, eigene hier — getrennt, wie das Issue es verlangt.

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// magicsaat trägt, was die kanonischen Sets nicht tragen: Unicode in drei
// Schriften und mit Zeichen außerhalb der Grundebene, einen Tabulator in
// einem mehrzeiligen Wert, eine geschützte spitze Klammer, ein Bild, das
// dieser Lauf selbst zeichnet, und **unbekannte Schlüssel** auf Set- und
// Kartenebene — die ein Programm mitschreiben soll, das sie nicht versteht.
var magicsaat = []string{
	"card:",
	"\tname: Grüße aus Köln — 漢字 und עברית",
	"\tcasting cost: 2WU",
	"\trule text:",
	"\t\tErste Zeile mit einem Tab\tdarin.",
	"\t\tZweite Zeile mit \\<kein Tag> und „Anführungen“ und 𝄞.",
	"\tflavor text: <i-flavor>Ein Satz mit 3½ und ‰.</i-flavor>",
	"\timage: image1",
	"\tpower: 2",
	"\ttoughness: 3",
	"\trarity: rare",
	"\teigener kartenschluessel: eigener Wert",
	"card:",
	"\tname: Zweite Karte",
	"\trule text: Ohne Bild, ohne Unicode.",
}

// sagasaat ist dasselbe für ein zweites Spiel des Packs: andere Felder,
// anderer Stil, dieselben Sonderfälle.
var sagasaat = []string{
	"card:",
	"\tname: Kapitel — Ελληνικά und Кириллица",
	"\tsupertype: Saga",
	"\trule text:",
	"\t\tI — Eine Zeile.",
	"\t\tII — Noch eine, mit \\<Klammer>.",
	"\tflavor text: Ein Satz.",
	"\timage: image1",
	"\trarity: uncommon",
	"\teigener kartenschluessel: eigener Wert",
	"card:",
	"\tname: Zweites Kapitel",
}

// eigeneSaat schreibt ein Set mit eigenem Inhalt als ZIP, samt einem Bild.
func eigeneSaat(t *testing.T, spiel, stil string, karten []string) string {
	t.Helper()

	set := strings.Join(append([]string{
		"mse version: 2.0.2",
		"game: " + spiel,
		"stylesheet: " + stil,
		"set info:",
		"\ttitle: Eigene Rundlaufsaat",
		"\tdescription: Nur für den Rundlauf — eigener Inhalt, keine fremde Vorlage.",
		"\teigener setschluessel: bleibt das?",
	}, append(karten, "")...), "\n")

	var bild bytes.Buffer

	zeichnung := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for x := range 8 {
		zeichnung.Set(x, x, color.RGBA{R: 200, G: 30, B: 60, A: 255})
	}

	if fehler := png.Encode(&bild, zeichnung); fehler != nil {
		t.Fatalf("das Bild: %v", fehler)
	}

	var archiv bytes.Buffer

	schreiber := zip.NewWriter(&archiv)

	for _, datei := range []struct {
		name   string
		inhalt []byte
	}{{"set", []byte(set)}, {"image1", bild.Bytes()}} {
		eintrag, fehler := schreiber.Create(datei.name)
		if fehler != nil {
			t.Fatalf("%s: %v", datei.name, fehler)
		}

		if _, fehler := eintrag.Write(datei.inhalt); fehler != nil {
			t.Fatalf("%s: %v", datei.name, fehler)
		}
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatalf("das Archiv: %v", fehler)
	}

	pfad := filepath.Join(t.TempDir(), spiel+"-saat.mse-set")
	if fehler := os.WriteFile(pfad, archiv.Bytes(), 0o600); fehler != nil {
		t.Fatalf("die Saat: %v", fehler)
	}

	return pfad
}
