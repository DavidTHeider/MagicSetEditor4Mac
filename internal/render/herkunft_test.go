package render_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// grau baut ein Bild einer Graustufe — die kürzeste Form für einen Vergleich,
// bei dem es auf die Fläche ankommt und nicht auf die Farbe.
func grau(breite, hoehe int, wert uint8) *image.NRGBA {
	bild := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))

	for y := range hoehe {
		for x := range breite {
			bild.SetNRGBA(x, y, color.NRGBA{R: wert, G: wert, B: wert, A: 255})
		}
	}

	return bild
}

// deckendesBild ist der Pfad einer undurchsichtigen Kachel in [pruefbilder].
const deckendesBild = "/deckend.png"

// durchsichtigesBild ist der Pfad einer ganz durchsichtigen Kachel.
const durchsichtigesBild = "/durchsichtig.png"

// pruefbilder ist die Bildquelle dieser Datei.
//
// **Zwei Kacheln reichen**, weil hier nicht das Bild geprüft wird, sondern die
// Frage, wer gemalt hat.
func pruefbilder() bilder {
	return bilder{
		deckendesBild:      einfarbig(1, 1, color.NRGBA{R: 200, G: 200, B: 200, A: 255}),
		durchsichtigesBild: einfarbig(1, 1, color.NRGBA{R: 255, G: 255, B: 255}),
	}
}

// mitHerkunftGezeichnet rastert einen Plan **mit** seiner Herkunft.
func mitHerkunftGezeichnet(t *testing.T, plan *render.Plan) (*image.NRGBA, *render.Herkunft) {
	t.Helper()

	bild, herkunft, befunde, err := render.NeuerRasterer(pruefbilder()).
		ZeichneMitHerkunft(t.Context(), plan, 1)
	if err != nil {
		t.Fatalf("ZeichneMitHerkunft: %v", err)
	}

	for _, befund := range befunde {
		t.Logf("Befund: %v", befund)
	}

	return bild, herkunft
}

// flaechenplan baut einen Plan aus lauter Farbflächen.
//
// **Farbflächen, weil sie ohne Bildquelle und ohne Schrift zeichnen.** Ein
// Plan mit Bildern brauchte ein Paket, und §11 lässt keines ins Repositorium.
func flaechenplan(breite, hoehe float64, ebenen ...render.Ebene) *render.Plan {
	return &render.Plan{Breite: breite, Hoehe: hoehe, Ebenen: ebenen}
}

// gefuellt ist eine deckende Farbfläche.
//
// **Immer [render.ArtFlaeche]**, denn `ebeneZeichnen` schaltet über die Art:
// eine Farbe unter einer anderen Art malt nichts. Die zweite und dritte Art
// dieser Datei kommen aus [bildebene].
func gefuellt(r render.Rechteck, farbe render.Farbe) render.Ebene {
	return render.Ebene{Art: render.ArtFlaeche, Rahmen: r, Farbe: farbe, Deckung: 1}
}

// bildebene legt eine Kachel aus [pruefbilder] in ein Rechteck.
//
// **Eine Bildebene und keine Textebene**: `ebeneZeichnen` schaltet über die
// Art, und eine Textebene ohne eingerichtete Schrift malt nichts — dann prüfte
// der Test die fehlende Schrift statt die Herkunft. [render.ArtBild] und
// [render.ArtSymbol] gehen beide durch den Bildweg und sind trotzdem zwei
// **verschiedene** Arten; genau das braucht die Zerlegung.
func bildebene(art render.Ebenenart, r render.Rechteck, pfad string) render.Ebene {
	return render.Ebene{Art: art, Rahmen: r, Bild: render.Blatt(pfad), Deckung: 1}
}

// TestHerkunftNenntDieEbeneDieZuletztGemaltHat ist die Kernzusage.
func TestHerkunftNenntDieEbeneDieZuletztGemaltHat(t *testing.T) {
	t.Parallel()

	// Unten eine Fläche über alles, darüber ein „Textkasten" in der oberen
	// Hälfte. Die Ebenenart steht frei — der Rasterer malt beide als Fläche.
	plan := flaechenplan(4, 4,
		gefuellt(render.Rechteck{Breite: 4, Hoehe: 4},
			render.Farbe{R: 10, G: 10, B: 10, A: 255}),
		bildebene(render.ArtSymbol, render.Rechteck{Breite: 4, Hoehe: 2}, deckendesBild),
	)

	_, herkunft := mitHerkunftGezeichnet(t, plan)

	if art := herkunft.Art(1, 0); art != render.ArtSymbol {
		t.Errorf("oben: %q, erwartet %q", art, render.ArtSymbol)
	}

	if art := herkunft.Art(1, 3); art != render.ArtFlaeche {
		t.Errorf("unten: %q, erwartet %q", art, render.ArtFlaeche)
	}
}

// TestHerkunftUebergehtEineEbeneDieNichtsMalt ist der Befund, der den ersten
// Anlauf verworfen hat.
//
// **Eine ganzkartengroße Ebene, die nichts ändert, beansprucht nichts.** Genau
// daran ist die Zerlegung nach Rechtecken gescheitert: eine gewöhnliche
// M15-Karte trägt zuoberst ein Bild über die ganze Fläche, und die Zerlegung
// meldete darauf 100 % Bild.
func TestHerkunftUebergehtEineEbeneDieNichtsMalt(t *testing.T) {
	t.Parallel()

	plan := flaechenplan(4, 4,
		gefuellt(render.Rechteck{Breite: 4, Hoehe: 4},
			render.Farbe{R: 10, G: 10, B: 10, A: 255}),
		// Ganz durchsichtig über die ganze Karte — MSEs Überlagerung.
		// **Deckung 1**, damit die Ebene betreten wird: durchsichtig ist sie
		// über ihr Alpha, nicht über ihre Deckung.
		bildebene(render.ArtBild, render.Rechteck{Breite: 4, Hoehe: 4},
			durchsichtigesBild),
	)

	_, herkunft := mitHerkunftGezeichnet(t, plan)

	if art := herkunft.Art(2, 2); art != render.ArtFlaeche {
		t.Errorf("%q, erwartet %q — die durchsichtige Ebene hat nichts gemalt",
			art, render.ArtFlaeche)
	}
}

// TestHerkunftLaesstUnbemalteBildpunkteUnbedeckt prüft den Nullwert.
func TestHerkunftLaesstUnbemalteBildpunkteUnbedeckt(t *testing.T) {
	t.Parallel()

	plan := flaechenplan(4, 4,
		gefuellt(render.Rechteck{Breite: 4, Hoehe: 2},
			render.Farbe{R: 10, G: 10, B: 10, A: 255}),
	)

	_, herkunft := mitHerkunftGezeichnet(t, plan)

	if art := herkunft.Art(1, 3); art != render.ArtUnbedeckt {
		t.Errorf("%q, erwartet %q", art, render.ArtUnbedeckt)
	}
}

// TestHerkunftAusserhalbIstUnbedeckt prüft den Zugriff daneben.
func TestHerkunftAusserhalbIstUnbedeckt(t *testing.T) {
	t.Parallel()

	herkunft := render.NeueHerkunft(2, 2)

	daneben := []image.Point{{X: -1}, {Y: -1}, {X: 2}, {Y: 2}}
	for _, stelle := range daneben {
		if art := herkunft.Art(stelle.X, stelle.Y); art != render.ArtUnbedeckt {
			t.Errorf("%v: %q, erwartet %q", stelle, art, render.ArtUnbedeckt)
		}
	}

	var keine *render.Herkunft
	if art := keine.Art(0, 0); art != render.ArtUnbedeckt {
		t.Errorf("nil-Herkunft: %q, erwartet %q", art, render.ArtUnbedeckt)
	}
}

// TestZeichneMitHerkunftGibtDasselbeBild ist die Zusage, dass die Aufzeichnung
// nichts verändert.
//
// **Ohne sie wäre die Messung an einem anderen Bild erhoben** als dem, das die
// Anwendung zeigt — und die Zerlegung erklärte einen Abstand, den es so nicht
// gibt (§10).
func TestZeichneMitHerkunftGibtDasselbeBild(t *testing.T) {
	t.Parallel()

	plan := saatplan()

	mitHerkunft, _, _, err := render.NeuerRasterer(pruefbilder()).
		ZeichneMitHerkunft(t.Context(), plan, 1)
	if err != nil {
		t.Fatalf("ZeichneMitHerkunft: %v", err)
	}

	gewoehnlich, _, err := render.NeuerRasterer(pruefbilder()).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("Zeichne: %v", err)
	}

	vergleich, err := render.Vergleiche(gewoehnlich, mitHerkunft)
	if err != nil {
		t.Fatalf("Vergleiche: %v", err)
	}

	if vergleich.Abweichende != 0 {
		t.Errorf("%d Bildpunkte unterscheiden sich — die Aufzeichnung malt mit",
			vergleich.Abweichende)
	}
}
