package render

import (
	"image"
	"image/color"
	"testing"
)

// **Die Probe findet den Versatz einer Ebene** —
// #1029.
//
// # Was hier geprüft wird und was nicht
//
// **Hier: dass die Probe rechnet, was sie soll** — über die Bildpunkte einer
// Art, und mit derselben Vorzeichenordnung wie [ProbeVersatz].
//
// **Nicht hier: dass ein Ebenenversatz in der Gesamtprobe verschwindet.** Das
// ist eine Aussage über **echte Karten**, und sie ist an ihnen gemessen: über
// 680 kanonische Karten findet die Gesamtprobe **null**, die durch eine
// Verschiebung gewinnen, während die Bilder den Regeltext doppelt zeigen.
//
// Der erste Anlauf hat genau das synthetisch nachbauen wollen und ist daran
// gescheitert: ein Grund, der eine Verschiebung bestraft, bestraft sie **auch**
// innerhalb der Textzeilen — und dann findet die Ebenenprobe den Versatz
// ebenso wenig. Ein Aufbau, der beides zugleich zeigt, wäre gebaut, um es zu
// zeigen, und bewiese nichts (§10).
func TestDieEbenenprobeFindetDenVersatzIhrerArt(t *testing.T) {
	t.Parallel()

	unser, referenz, herkunft := paarMitVersetzterTextzeile()

	probe, err := ProbeVersatzJeEbene(unser, referenz, herkunft, ArtText, 2)
	if err != nil {
		t.Fatalf("ProbeVersatzJeEbene: %v", err)
	}

	// **(0,+1):** unser Streifen sitzt eine Zeile **tiefer** als der der
	// Referenz — was sie bei y zeigt, steht bei uns an y+1. Dieselbe Ordnung
	// wie in `TestProbeVersatzFindetDieVerschiebung`.
	if probe.X != 0 || probe.Y != 1 {
		t.Errorf("bester Versatz (%d,%d), erwartet (0,1)", probe.X, probe.Y)
	}

	if probe.Bester != 0 {
		t.Errorf("Abstand beim besten Versatz %.3f, erwartet 0", probe.Bester)
	}

	if probe.Ohne <= probe.Bester {
		t.Errorf("ohne Versatz %.3f, mit %.3f — die Probe hat nichts gefunden",
			probe.Ohne, probe.Bester)
	}
}

// **Und sie misst nur die Bildpunkte ihrer Art.**
//
// Sonst wäre die Zahl daneben eine über die ganze Karte, und niemand sähe es
// ihr an (§10).
func TestDieEbenenprobeZaehltNurIhreArt(t *testing.T) {
	t.Parallel()

	unser, referenz, herkunft := paarMitVersetzterTextzeile()

	ganz, err := ProbeVersatz(unser, referenz, 2)
	if err != nil {
		t.Fatalf("ProbeVersatz: %v", err)
	}

	ebene, err := ProbeVersatzJeEbene(unser, referenz, herkunft, ArtText, 2)
	if err != nil {
		t.Fatalf("ProbeVersatzJeEbene: %v", err)
	}

	// Drei Zeilen von zwanzig im inneren Rechteck.
	if ebene.Punkte != 3*20 {
		t.Errorf("die Ebenenprobe zählt %d Punkte, erwartet %d",
			ebene.Punkte, 3*20)
	}

	if ganz.Punkte != 20*20 {
		t.Errorf("die Gesamtprobe zählt %d Punkte, erwartet %d",
			ganz.Punkte, 20*20)
	}
}

// **Eine Art, die es nicht gibt, gibt null Punkte und keinen Versatz.**
//
// Nicht einen Fehler: eine Karte ohne Symbolzug ist kein Mangel, und eine
// Meldung darüber stünde an jeder zweiten Karte.
func TestEineArtOhneBildpunkteGibtNullPunkte(t *testing.T) {
	t.Parallel()

	unser, referenz, herkunft := paarMitVersetzterTextzeile()

	probe, err := ProbeVersatzJeEbene(unser, referenz, herkunft, ArtSymbolzug, 2)
	if err != nil {
		t.Fatalf("ProbeVersatzJeEbene: %v", err)
	}

	if probe.Punkte != 0 || probe.X != 0 || probe.Y != 0 {
		t.Errorf("Probe %+v — erwartet waren null Punkte ohne Versatz", probe)
	}
}

// **Eine Herkunft, die nicht zum Bild passt, ist ein Fehler.**
//
// Sie wäre sonst still falsch: die Arten läsen sich an den falschen Stellen,
// und die Zahl sähe aus wie jede andere.
func TestEineFremdeHerkunftWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	unser, referenz, _ := paarMitVersetzterTextzeile()

	for name, herkunft := range map[string]*Herkunft{
		"keine":    nil,
		"zu klein": NeueHerkunft(4, 4),
		"zu groß":  NeueHerkunft(200, 200),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := ProbeVersatzJeEbene(
				unser, referenz, herkunft, ArtText, 2); err == nil {
				t.Error("angenommen, obwohl die Herkunft nicht passt")
			}
		})
	}
}

// **Und die drei anderen Abweisungen ebenso.**
//
// Jede von ihnen ist ein Weg, auf dem die Zahl **still falsch** wäre: ein Paar
// verschiedener Größe verglichen aneinander vorbei, eine negative Reichweite
// gäbe eine Schleife, die nie läuft, und ein Bild, das kleiner ist als der
// Suchbereich, hätte kein inneres Rechteck mehr.
func TestDieEbenenprobeWeistUnbrauchbareEingabenAb(t *testing.T) {
	t.Parallel()

	unser, referenz, herkunft := paarMitVersetzterTextzeile()

	t.Run("Bilder verschiedener Größe", func(t *testing.T) {
		t.Parallel()

		anderes := mitZeile(12, 12, 5)

		if _, err := ProbeVersatzJeEbene(
			unser, anderes, herkunft, ArtText, 2); err == nil {
			t.Error("angenommen, obwohl die Bilder nicht zusammenpassen")
		}
	})

	t.Run("negative Reichweite", func(t *testing.T) {
		t.Parallel()

		if _, err := ProbeVersatzJeEbene(
			unser, referenz, herkunft, ArtText, -1); err == nil {
			t.Error("angenommen, obwohl die Reichweite negativ ist")
		}
	})

	t.Run("Reichweite größer als das Bild", func(t *testing.T) {
		t.Parallel()

		klein := mitZeile(6, 6, 2)
		kleineHerkunft := NeueHerkunft(6, 6)

		if _, err := ProbeVersatzJeEbene(
			klein, mitZeile(6, 6, 2), kleineHerkunft, ArtText, 4); err == nil {
			t.Error("angenommen, obwohl das Bild kleiner ist als der Suchbereich")
		}
	})
}

// paarMitVersetzterTextzeile baut die Lage der doppelseitigen Karten nach.
//
// Ein Grund mit senkrechten Streifen — **in y unveränderlich**, damit eine
// senkrechte Verschiebung ihn nicht trifft — und drei volle Zeilen Tinte, die
// bei uns **einen Bildpunkt tiefer** sitzen.
func paarMitVersetzterTextzeile() (unser, referenz image.Image, herkunft *Herkunft) {
	const breite, hoehe = 24, 24

	referenz = mitZeile(breite, hoehe, 10)
	unser = mitZeile(breite, hoehe, 11)

	herkunft = NeueHerkunft(breite, hoehe)

	for y := range hoehe {
		for x := range breite {
			art := ArtBild
			if y >= 11 && y < 14 {
				art = ArtText
			}

			herkunft.arten[y*breite+x] = art
		}
	}

	return unser, referenz, herkunft
}

// mitZeile malt senkrechte Streifen mit drei vollen Zeilen Tinte.
func mitZeile(breite, hoehe, oben int) *image.NRGBA {
	bild := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))

	for y := range hoehe {
		for x := range breite {
			grund := uint8(60)
			if (x/2)%2 == 0 {
				grund = 220
			}

			farbe := color.NRGBA{R: grund, G: grund, B: grund, A: 255}
			if y >= oben && y < oben+3 {
				farbe = color.NRGBA{A: 255}
			}

			bild.SetNRGBA(x, y, farbe)
		}
	}

	return bild
}
