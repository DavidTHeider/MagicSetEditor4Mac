//go:build fremdpakete

package app

// **Was der kanonische Lauf als Klasse ausweist**
// (#427).
//
// Die Klasse ist der **Grund**, nicht der Ort. Bis hierher schnitt
// [kanonklasse] am ersten `": "` und warf damit den Grund weg — die Ausgabe
// nannte vier Feldnamen und ließ offen, was an ihnen falsch ist.

import "testing"

// **Eine Gegenprobe in beide Richtungen** (§10): eine Klassenbildung, die
// alles in einen Topf wirft, ist so wertlos wie eine, die nichts zusammenfasst.
// Deshalb steht hier beides — was zusammenfallen **muss** und was
// auseinanderbleiben **muss**.
func TestDieKanonklasseNenntDenGrund(t *testing.T) {
	for _, fall := range []struct {
		name     string
		meldung  string
		erwartet string
	}{
		{
			name:     "der Ort fällt weg",
			meldung:  `text: die Auszeichnungen soft-line sind entfernt; gezeichnet wird der bloße Text`,
			erwartet: "die Auszeichnungen … sind entfernt",
		},
		{
			name:     "und mit ihm die genannten Auszeichnungen",
			meldung:  `type: die Auszeichnungen word-list-class, word-list-race sind entfernt; gezeichnet wird der bloße Text`,
			erwartet: "die Auszeichnungen … sind entfernt",
		},
		{
			name:     "ein unbekannter Name behält seinen Namen",
			meldung:  `arrow 7 → senkrecht → kante: die Angabe ließ sich nicht ausrechnen: Zeile 798, Spalte 35: der Name "is_room" ist hier nicht bekannt`,
			erwartet: `der Name "is_room" ist hier nicht bekannt`,
		},
		{
			name:     "ein anderer Name ist eine andere Klasse",
			meldung:  `x: die Angabe ließ sich nicht ausrechnen: Zeile 1, Spalte 1: der Name "cos" ist hier nicht bekannt`,
			erwartet: `der Name "cos" ist hier nicht bekannt`,
		},
		{
			name:     "ein fehlendes Bild fällt auf seine Klasse",
			meldung:  `Ebene 15: das Bild /magic-m15-mainframe-planeswalker.mse-style/image7 ist nicht lesbar: /magic-m15-mainframe-planeswalker.mse-style/image7: das Bild ist in keinem Paket zu finden`,
			erwartet: "das Bild ist in keinem Paket zu finden",
		},
		{
			name:     "ein zweites fehlendes Bild ist dieselbe Klasse",
			meldung:  `Ebene 12: das Bild /magic-m15-mainframe-dfc.mse-style/image1 ist nicht lesbar: /magic-m15-mainframe-dfc.mse-style/image1: das Bild ist in keinem Paket zu finden`,
			erwartet: "das Bild ist in keinem Paket zu finden",
		},
		{
			// **Zweifach umschlossen**: die Maske nennt das Bild, das Bild
			// nennt die Datei, und die Datei nennt den Grund. Ein einziger
			// Schnitt bliebe bei „das Bild … ist nicht lesbar" stehen.
			name:     "eine fehlende Maske fällt auf dieselbe Klasse wie ein fehlendes Bild",
			meldung:  `Ebene 13: die Maske /a.mse-style/image_masks/m.png ist nicht lesbar: das Bild /a.mse-style/image_masks/m.png ist nicht lesbar: /a.mse-style/image_masks/m.png: das Bild ist in keinem Paket zu finden`,
			erwartet: "das Bild ist in keinem Paket zu finden",
		},
		{
			name:     "fehlende Zeichen nennen ihre Schrift, nicht die Zeichen",
			meldung:  `Ebene 15: in der Schrift Beleren Small Caps Bold fehlen Zeichen: ǵ`,
			erwartet: "in der Schrift Beleren Small Caps Bold fehlen Zeichen",
		},
		{
			name:     "eine Meldung ohne Ort bleibt, wie sie ist",
			meldung:  `to_int erwartet ein Argument, bekam 2 Argumente`,
			erwartet: "to_int erwartet ein Argument, bekam 2 Argumente",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			if bekommen := kanonklasse(fall.meldung); bekommen != fall.erwartet {
				t.Errorf("Klasse %q, erwartet %q", bekommen, fall.erwartet)
			}
		})
	}
}
