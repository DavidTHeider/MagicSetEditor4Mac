//go:build fremdpakete

package render

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// **Die Glättung trifft Quartz** — gegen Core Text selbst, nicht gegen eine
// Zahl, die einmal aus ihm kam
// (#1806).
//
// `scripts/coretext/raster.swift` zeichnet eine Zeile auf MSE2s vierfachem Weg
// mit Core Text, so wie wxWidgets den Kontext anlegt; dieser Test zeichnet
// dieselbe Zeile mit [OpenTypeSchrift.Maske], [vierfachlauf] und
// [herunterrechnen] und vergleicht Bildpunkt für Bildpunkt.
//
// **Die Gegenprobe steht im Test**: ohne Verdickung muss der Abstand groß
// sein, sonst sähe der Vergleich die Glättung gar nicht.
//
// # Was er nicht hält
//
// **Den Satz.** Gezeichnet wird ein Stück, wie es der Zeichner an den
// vierfachen Weg gibt; wo die Stücke einer Zeile stehen, hält der Umbruch.
// Und **gedrehten Text**: Quartz verdickt im Gerätebild, und nahe 45 Grad gar
// nicht (`docs/SCHRIFTUNTERSUCHUNG.md`).
func TestDieGlaettungTrifftQuartz(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	if _, err := exec.LookPath("swift"); err != nil {
		t.Skip("kein swift auf diesem Rechner — Core Text ist nicht zu fragen")
	}

	for _, fall := range []struct {
		datei         string
		groesse       float64
		breite, hoehe int
		text          string
		hoechstens    float64
	}{
		{"mplantin.ttf", 14, 420, 44, "Flying, first strike. Whenever this creature attacks", 0.3},
		{"beleren-bold_P1.01.ttf", 16, 260, 44, "Armored Griffin of Waffling Kyscu", 0.4},
	} {
		pfad := filepath.Join(filepath.Dir(wurzel), "Magic - Fonts", fall.datei)

		roh, err := os.ReadFile(pfad) //nolint:gosec // Pfad aus der Umgebung
		if err != nil {
			t.Fatalf("die Schrift des Packs: %v", err)
		}

		schrift, err := SchriftAusDatei(roh)
		if err != nil {
			t.Fatalf("lesen: %v", err)
		}

		const links, grund = 0.3, 30.0

		referenz := quartzZeile(t, pfad, fall.groesse, links, grund, fall.breite, fall.hoehe, fall.text)

		for _, glatt := range []bool{true, false} {
			eigene := eigeneZeile(schrift, fall.groesse, links, grund, fall.breite, fall.hoehe, fall.text, glatt)
			abstand := betragsabstand(referenz, eigene)

			switch {
			case glatt && abstand > fall.hoechstens:
				t.Errorf("%s: %.3f von 255 neben Quartz, höchstens %.1f", fall.datei, abstand, fall.hoechstens)
			case !glatt && abstand < 1.5:
				t.Errorf("%s: ohne Verdickung nur %.3f neben Quartz — der Vergleich sieht die Glättung nicht",
					fall.datei, abstand)
			}
		}
	}
}

// quartzZeile lässt Core Text die Zeile zeichnen und gibt die Deckung.
func quartzZeile(t *testing.T, schrift string, groesse, links, grund float64, breite, hoehe int, text string) []byte {
	t.Helper()

	ausgabe := filepath.Join(t.TempDir(), "quartz.pgm")

	//nolint:gosec // die Probe aus diesem Repositorium, Werte aus dem Test
	befehl := exec.CommandContext(t.Context(), "swift", "../../scripts/coretext/raster.swift", schrift,
		fmt.Sprint(groesse), fmt.Sprint(links), fmt.Sprint(grund), fmt.Sprint(breite), fmt.Sprint(hoehe),
		"vorgabe", ausgabe, text)

	if meldung, err := befehl.CombinedOutput(); err != nil {
		t.Fatalf("raster.swift: %v\n%s", err, meldung)
	}

	roh, err := os.ReadFile(ausgabe) //nolint:gosec // eben geschrieben
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	kopf := fmt.Sprintf("P5\n%d %d\n255\n", breite, hoehe)
	if !bytes.HasPrefix(roh, []byte(kopf)) {
		t.Fatalf("raster.swift gab keine Deckung von %d × %d", breite, hoehe)
	}

	return roh[len(kopf):]
}

// eigeneZeile zeichnet die Zeile als ein Stück auf dem vierfachen Weg.
func eigeneZeile(
	schrift *OpenTypeSchrift, groesse, links, grund float64, breite, hoehe int, text string, glatt bool,
) []byte {
	lauf := neuerVierfachlauf(links, 0, float64(breite-3), float64(hoehe-1), grund, 1)
	satz := Satz{Groesse: groesse}
	versatz := 0.0
	vorher := rune(0)

	for _, zeichen := range gesetzt(text, satz, schrift) {
		korrektur := kernungZwischen(vorher, zeichen, satz, schrift)

		if kennung, da := schrift.kennung(zeichen); da {
			verdickung := 0.0
			if glatt {
				verdickung = quartzVerdickung(groesse * Textraster)
			}

			schrift.schloss.Lock()
			maske, ecke, gezeichnet := schrift.maskeDerKennung(kennung, groesse*Textraster, verdickung)
			schrift.schloss.Unlock()

			if gezeichnet {
				lauf.setze(maske, ecke, versatz+korrektur)
			}
		}

		versatz += korrektur + schrift.Breite(zeichen)*groesse
		vorher = zeichen
	}

	voll := herunterrechnen(lauf.deckung, lauf.breite, lauf.hoehe, lauf.ausgabebreite)

	// **Ohne den Rand, den der Puffer seit #1835 oben und unten trägt** — eine
	// halbe Zeile darüber und darunter ([neuerVierfachlauf]). Core Text
	// zeichnet nur die Zeile; hier stand bis zum Kanonlauf vom 2026-10-04 das
	// ganze, doppelt so hohe Bild, und der Abstand war `+Inf`.
	rand := -lauf.yi

	return voll[rand*lauf.ausgabebreite : (rand+hoehe)*lauf.ausgabebreite]
}

// betragsabstand ist das Mittel der Beträge über alle Bildpunkte, in Stufen
// von 0 bis 255.
func betragsabstand(a, b []byte) float64 {
	if len(a) != len(b) {
		return math.Inf(1)
	}

	summe := 0.0
	for i := range a {
		summe += math.Abs(float64(a[i]) - float64(b[i]))
	}

	return summe / float64(len(a))
}
