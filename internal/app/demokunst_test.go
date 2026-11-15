//go:build fremdpakete

package app_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// **Die sechste Demokarte bringt ihr eigenes Bild mit.**
//
// Die fünf anderen zeigen ein leeres Kunstfenster — auf einer echten Karte ist
// dort die Illustration, und ohne sie wirkt keine Karte echt
// (#380).
//
// **Gemalt und nicht gefunden** (§11): ein Farbverlauf aus zwei Schleifen. Wer
// wissen will, woher das Bild kommt, liest [maleIllustration].

// demokunstfarbe ist die Ecke oben links der gemalten Illustration.
//
// **Ein Wert, den der M15-Rahmen nicht führt.** Gesucht wird er im Bild; ein
// Ton, den auch der Rahmen trägt, machte den Fund zur Glückssache.
var demokunstfarbe = color.NRGBA{R: 255, G: 0, B: 255, A: 255}

// kunstsetFuer schreibt ein Set-Paket mit einer gemalten Illustration.
//
// Gibt den **Ort** zurück — das ist, was [service.Vorlage.Setpaket] erwartet.
func kunstsetFuer(t *testing.T, datei string) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "Demoset.mse-set")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	var puffer bytes.Buffer
	if err := png.Encode(&puffer, maleIllustration()); err != nil {
		t.Fatalf("kodieren: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ort, datei), puffer.Bytes(), 0o600); err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	return ort
}

// maleIllustration malt einen Verlauf mit einer erkennbaren Ecke.
//
// **Ein Verlauf und kein einfarbiges Feld**, weil ein einfarbiges nicht zeigt,
// ob das Bild richtig herum und über das ganze Fenster gestreckt ankommt.
func maleIllustration() image.Image {
	const kante = 64

	gemalt := image.NewNRGBA(image.Rect(0, 0, kante, kante))

	for y := range kante {
		for x := range kante {
			gemalt.SetNRGBA(x, y, color.NRGBA{
				R: uint8(255 - y*3), G: uint8(x * 3), B: 255, A: 255,
			})
		}
	}

	// **Die Ecke ist gesetzt, nicht gerechnet.** Sie ist die Marke, an der der
	// Lauf das Bild wiedererkennt; aus dem Verlauf zu erschließen, welcher Ton
	// dort landet, hieße die Rechnung zweimal zu schreiben.
	gemalt.SetNRGBA(0, 0, demokunstfarbe)

	return gemalt
}

// fuehrtIllustration sagt, ob die gemalte Ecke im Bild steht.
func fuehrtIllustration(bild *image.NRGBA) bool {
	rahmen := bild.Bounds()

	for y := rahmen.Min.Y; y < rahmen.Max.Y; y++ {
		for x := rahmen.Min.X; x < rahmen.Max.X; x++ {
			if bild.NRGBAAt(x, y) == demokunstfarbe {
				return true
			}
		}
	}

	return false
}

// helleImFuss zählt die hellen Bildpunkte im Fußband der Karte.
//
// # Wo gemessen wird, und warum genau dort
//
// **Unten links, zwischen x = 20 und 150 und y = 488 bis 515.** Dort stehen
// Setcode und Illustrator (`magic-modules.mse-include/information/card_fields`
// setzt sie bei 499 mit 10 Punkten Höhe). Weiter rechts säßen das Prägesiegel
// (159, 471) und der P/T-Kasten (273, 466) — beide hell, beide ohne etwas mit
// dem Fuß zu tun zu haben.
//
// **Der Grund ist schwarz**, denn der Kartenrahmen der M15-Vorlagen ist es ab
// Zeile 486 (nachgemessen an `375 m15 simple/rcard.jpg`, x = 187). Ein heller
// Punkt dort ist deshalb Schrift und sonst nichts.
func helleImFuss(bild *image.NRGBA) int {
	var hell int

	for y := 488; y < 516; y++ {
		for x := 20; x < 150; x++ {
			farbe := bild.NRGBAAt(x, y)
			if int(farbe.R)+int(farbe.G)+int(farbe.B) > 300 {
				hell++
			}
		}
	}

	return hell
}
