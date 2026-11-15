package service

// **Was MSE2 an einem neuen Stil beanstandet hat** —
// #1350.
//
// Der erste Anlauf schrieb `font:` unter **jedes** Feld. MSE2 hat es beim
// Öffnen gemeldet:
//
//	WARNING: Warnings while reading file:
//	         …/eigenes-schlicht.mse-style/style
//	         On line 35: 	Unexpected key: 'font'
//
// Gefunden hat es **nicht ein Test, sondern das Orakel** (§15) — und dieser
// Test hält es fest, damit es nicht zurückkommt: ein `ImageStyle` kennt nur
// `default` (`src/data/field/image.cpp`, Zeilen 24–27, Fassung f9d9356 vom
// 2021-07-06), `font` steht am `TextStyle` (`src/data/field/text.cpp`,
// Zeilen 143–159).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestEineSchriftStehtNurAnEinemTextfeld(t *testing.T) {
	t.Parallel()

	baum := stilbaum(
		Neuanlage{Art: "mse-style", Name: "schlicht", Vollname: "Schlicht", Spiel: "eigenes"},
		[]saatfeld{
			{name: "name", art: "text", hoehe: 28},
			{name: "kunst", art: "image", hoehe: 240},
		})

	geschrieben := string(msepaket.Schreibe(baum))

	nameblock, kunstblock := feldblock(geschrieben, "name"), feldblock(geschrieben, "kunst")

	if !strings.Contains(nameblock, "font:") {
		t.Errorf("das Textfeld hat keine Schrift:\n%s", nameblock)
	}

	if strings.Contains(kunstblock, "font:") {
		t.Errorf("das Bildfeld hat eine Schrift — MSE2 meldet „Unexpected key: 'font'\":\n%s",
			kunstblock)
	}
}

// **Und die Höhen stapeln sich, statt übereinanderzuliegen.**
//
// Ohne diesen Fall wäre ein Stil grün, der jedes Feld auf `top: 24` legte —
// die Karte zeigte dann ein Feld und darunter nichts.
func TestDieFelderStapelnSich(t *testing.T) {
	t.Parallel()

	geschrieben := string(msepaket.Schreibe(stilbaum(
		Neuanlage{Art: "mse-style", Name: "s", Vollname: "S", Spiel: "e"},
		[]saatfeld{
			{name: "eins", art: "text", hoehe: 28},
			{name: "zwei", art: "text", hoehe: 28},
		})))

	if oben := obenIn(feldblock(geschrieben, "zwei")); oben <= obenIn(feldblock(geschrieben, "eins")) {
		t.Errorf("das zweite Feld liegt bei %d, nicht unter dem ersten:\n%s",
			oben, geschrieben)
	}
}

// feldblock schneidet die Zeilen eines Feldes aus dem `card style` heraus.
func feldblock(geschrieben, feld string) string {
	var (
		gebaut strings.Builder
		drin   bool
	)

	for _, zeile := range strings.Split(geschrieben, "\n") {
		if zeile == "\t"+feld+":" {
			drin = true

			continue
		}

		if drin {
			if !strings.HasPrefix(zeile, "\t\t") {
				break
			}

			gebaut.WriteString(zeile + "\n")
		}
	}

	return gebaut.String()
}

// obenIn liest die `top:`-Zeile eines Blocks.
func obenIn(block string) int {
	for _, zeile := range strings.Split(block, "\n") {
		if wert, gefunden := strings.CutPrefix(strings.TrimSpace(zeile), "top: "); gefunden {
			zahl := 0
			for _, ziffer := range wert {
				zahl = zahl*10 + int(ziffer-'0')
			}

			return zahl
		}
	}

	return -1
}
