package render_test

// **Jede Art hat ihre Meldung**
// (#513).
//
// Die Zuordnung hängt am Meldungstext, und das ist der Preis: eine Meldung,
// die jemand umformuliert, fällt heraus und zählt wieder als Portierung. Das
// ist die richtige Richtung — sie fällt in die **strengere** Art —, und dieser
// Wächter hält die vorhandenen Zuordnungen fest.
//
// **Er prüft nicht, ob eine Einordnung richtig ist.** Ob „die Vorlage rechnet
// ins Leere" wirklich zutrifft, steht in
// #497 — dort
// am Orakel belegt. Hier steht nur, dass die Meldung dieser Art zugeordnet
// wird.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestJedeArtHatIhreMeldung(t *testing.T) {
	faelle := []struct {
		meldung string
		art     render.Befundart
	}{
		{
			"in der Schrift Beleren Small Caps Bold fehlen Zeichen: U+2014",
			render.ArtUmgebung,
		},
		{
			"die Schrift Matrix ist nicht zu finden: keine Schriftdatei",
			render.ArtUmgebung,
		},
		{"das Bild /magic-m15-split-fusable.mse-style/true ist nicht lesbar: " +
			"das Bild ist in keinem Paket zu finden", render.ArtVorlage},
		{"die Symbolschrift magic-mana-small führt kein Symbol für \"\\n\"; " +
			"es wird übersprungen", render.ArtVorlage},

		// **Alles andere ist Portierung** — die strengste Art, und die
		// Vorgabe.
		{"die Verknüpfung rotate ist noch nicht gebaut", render.ArtPortierung},
		{"ein Textrahmen von -309×23", render.ArtPortierung},
		{"", render.ArtPortierung},
	}

	for _, fall := range faelle {
		t.Run(fall.art.String()+": "+gekuerzteMeldung(fall.meldung), func(t *testing.T) {
			if art := render.ArtAus(fall.meldung); art != fall.art {
				t.Errorf("%q ist %q, erwartet %q", fall.meldung, art, fall.art)
			}
		})
	}
}

// gekuerzteMeldung kürzt eine Meldung für den Testfallnamen.
func gekuerzteMeldung(meldung string) string {
	if len(meldung) <= 30 {
		return meldung
	}

	return meldung[:30]
}
