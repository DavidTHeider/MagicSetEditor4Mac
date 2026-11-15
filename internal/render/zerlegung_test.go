package render_test

import (
	"errors"
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// TestZerlegeOrdnetDenAbstandDerMalendenEbeneZu prüft die Partition.
func TestZerlegeOrdnetDenAbstandDerMalendenEbeneZu(t *testing.T) {
	t.Parallel()

	plan := flaechenplan(4, 4,
		gefuellt(render.Rechteck{Breite: 4, Hoehe: 4},
			render.Farbe{R: 50, G: 50, B: 50, A: 255}),
		bildebene(render.ArtSymbol, render.Rechteck{Breite: 4, Hoehe: 2}, deckendesBild),
	)

	unser, herkunft := mitHerkunftGezeichnet(t, plan)

	// Die obere Hälfte weicht um 10 je Kanal ab, die untere gar nicht.
	// Die deckende Kachel malt 200, die Fläche darunter 50.
	referenz := grau(4, 4, 50)
	for y := range 2 {
		for x := range 4 {
			referenz.SetNRGBA(x, y, color.NRGBA{R: 210, G: 210, B: 210, A: 255})
		}
	}

	zerlegung, err := render.ZerlegeAbstand(herkunft, unser, referenz)
	if err != nil {
		t.Fatalf("ZerlegeAbstand: %v", err)
	}

	oben := zerlegung.Teile[render.ArtSymbol]
	if oben.Punkte != 8 {
		t.Errorf("obere Ebene: %d Bildpunkte, erwartet 8", oben.Punkte)
	}

	// 8 Punkte × 3 Kanäle × 10² = 2.400.
	if oben.Quadratsumme != 2400 {
		t.Errorf("obere Ebene: Quadratsumme %.0f, erwartet 2400", oben.Quadratsumme)
	}

	if flaeche := zerlegung.Teile[render.ArtFlaeche]; flaeche.Quadratsumme != 0 {
		t.Errorf("Fläche: Quadratsumme %.0f, erwartet 0", flaeche.Quadratsumme)
	}
}

// TestZerlegeSummiertSichZumGanzen hält die Zusage fest, dass die Anteile
// aufgehen.
//
// **Ohne sie wäre die Zerlegung eine Sammlung von Zahlen ohne Ganzes**: ein
// Bildpunkt, der zwei Arten zugeschlagen oder von keiner erfasst würde, fiele
// nicht auf.
func TestZerlegeSummiertSichZumGanzen(t *testing.T) {
	t.Parallel()

	plan := flaechenplan(6, 6,
		gefuellt(render.Rechteck{Breite: 6, Hoehe: 4},
			render.Farbe{R: 10, G: 20, B: 30, A: 255}),
		bildebene(render.ArtBild, render.Rechteck{Links: 1, Oben: 1, Breite: 2, Hoehe: 2},
			deckendesBild),
		bildebene(render.ArtSymbol, render.Rechteck{Links: 4, Breite: 2, Hoehe: 1},
			deckendesBild),
	)

	unser, herkunft := mitHerkunftGezeichnet(t, plan)

	zerlegung, err := render.ZerlegeAbstand(herkunft, unser,
		einfarbig(6, 6, color.NRGBA{R: 40, G: 60, B: 80, A: 255}))
	if err != nil {
		t.Fatalf("ZerlegeAbstand: %v", err)
	}

	var summePunkte int

	var summeQuadrate float64

	for _, teil := range zerlegung.Teile {
		summePunkte += teil.Punkte
		summeQuadrate += teil.Quadratsumme
	}

	if summePunkte != 36 {
		t.Errorf("%d Bildpunkte über alle Teile, erwartet 36", summePunkte)
	}

	if summePunkte != zerlegung.Gesamt.Punkte {
		t.Errorf("Teile zählen %d Bildpunkte, das Ganze %d",
			summePunkte, zerlegung.Gesamt.Punkte)
	}

	if summeQuadrate != zerlegung.Gesamt.Quadratsumme {
		t.Errorf("Teile summieren %.0f, das Ganze %.0f",
			summeQuadrate, zerlegung.Gesamt.Quadratsumme)
	}
}

// TestZerlegeRechnetDenAbstandWieDerKanonlauf hält beide Maße zusammen.
//
// **Der Kanonlauf misst quadratisch.** Ein Teilabstand über ein anderes Maß
// wäre mit ihm nicht vergleichbar — genau der Fehler, den §10 anstreicht.
func TestZerlegeRechnetDenAbstandWieDerKanonlauf(t *testing.T) {
	t.Parallel()

	// Jeder der vier Bildpunkte weicht in jedem Kanal um 10 ab.
	zerlegung, err := render.ZerlegeAbstand(render.NeueHerkunft(2, 2),
		grau(2, 2, 100), grau(2, 2, 90))
	if err != nil {
		t.Fatalf("ZerlegeAbstand: %v", err)
	}

	if zerlegung.Gesamt.Abstand() != 10 {
		t.Errorf("Gesamtabstand %.3f, erwartet 10", zerlegung.Gesamt.Abstand())
	}

	if got := zerlegung.Teile[render.ArtUnbedeckt].Abstand(); got != 10 {
		t.Errorf("unbedeckt: Abstand %.3f, erwartet 10", got)
	}

	// **Ein Teil ohne Bildpunkte hat den Abstand null und nicht `NaN`.** Eine
	// Tabelle fragt jede Art ab, auch die, die auf dieser Karte nicht vorkommt
	// — und `NaN` machte die Zeile unlesbar.
	if got := zerlegung.Teile[render.ArtSymbolzug].Abstand(); got != 0 {
		t.Errorf("Art ohne Bildpunkte: Abstand %.3f, erwartet 0", got)
	}
}

// TestZerlegeNenntDenAnteilAmGanzen prüft die Zahl, die die Tabelle trägt.
func TestZerlegeNenntDenAnteilAmGanzen(t *testing.T) {
	t.Parallel()

	plan := flaechenplan(4, 4,
		gefuellt(render.Rechteck{Breite: 4, Hoehe: 4},
			render.Farbe{R: 50, G: 50, B: 50, A: 255}),
		bildebene(render.ArtSymbol, render.Rechteck{Breite: 4, Hoehe: 2}, deckendesBild),
	)

	unser, herkunft := mitHerkunftGezeichnet(t, plan)

	// Die deckende Kachel malt 200, die Fläche darunter 50.
	referenz := grau(4, 4, 50)
	for y := range 2 {
		for x := range 4 {
			referenz.SetNRGBA(x, y, color.NRGBA{R: 210, G: 210, B: 210, A: 255})
		}
	}

	zerlegung, err := render.ZerlegeAbstand(herkunft, unser, referenz)
	if err != nil {
		t.Fatalf("ZerlegeAbstand: %v", err)
	}

	if anteil := zerlegung.Anteil(render.ArtSymbol); anteil != 1 {
		t.Errorf("Anteil der oberen Ebene %.3f, erwartet 1", anteil)
	}

	if anteil := zerlegung.Anteil(render.ArtFlaeche); anteil != 0 {
		t.Errorf("Anteil der Fläche %.3f, erwartet 0", anteil)
	}
}

// TestZerlegeOhneAbweichungGibtDenAnteilNull prüft den Nenner null.
//
// **Zwei gleiche Bilder haben keinen Abstand**, und ein Anteil daran ist keine
// Zahl. Er wird null gemeldet, nicht `NaN` — eine Tabelle mit `NaN` ist
// unlesbar, und §10 verlangt, dass eine Zahl sagt, worüber sie erhoben wurde.
func TestZerlegeOhneAbweichungGibtDenAnteilNull(t *testing.T) {
	t.Parallel()

	zerlegung, err := render.ZerlegeAbstand(render.NeueHerkunft(2, 2),
		grau(2, 2, 7), grau(2, 2, 7))
	if err != nil {
		t.Fatalf("ZerlegeAbstand: %v", err)
	}

	if anteil := zerlegung.Anteil(render.ArtUnbedeckt); anteil != 0 {
		t.Errorf("Anteil %.3f, erwartet 0", anteil)
	}
}

// TestZerlegeWeistUnbrauchbareEingabenAb prüft die Fehlerwege.
func TestZerlegeWeistUnbrauchbareEingabenAb(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name           string
		herkunft       *render.Herkunft
		unser, referen image.Image
	}{
		{"keine Herkunft", nil, grau(2, 2, 0), grau(2, 2, 0)},
		{"kein Bild", render.NeueHerkunft(2, 2), nil, grau(2, 2, 0)},
		{"keine Referenz", render.NeueHerkunft(2, 2), grau(2, 2, 0), nil},
		{
			"verschiedene Größen", render.NeueHerkunft(2, 2),
			grau(2, 2, 0), grau(3, 2, 0),
		},
		{
			"Herkunft passt nicht zum Bild", render.NeueHerkunft(3, 3),
			grau(2, 2, 0), grau(2, 2, 0),
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, err := render.ZerlegeAbstand(fall.herkunft, fall.unser, fall.referen)
			if !errors.Is(err, render.ErrPlan) {
				t.Errorf("Fehler %v, erwartet ErrPlan", err)
			}
		})
	}
}

// TestGanzabstandStimmtMitDerZerlegungUeberein hält beide Wege zusammen.
//
// **Ohne diese Zusage stünden im Kanonlauf zwei Zahlen**, die dasselbe heißen
// und es nicht sind: eine für Karten mit Herkunft, eine für Karten ohne.
func TestGanzabstandStimmtMitDerZerlegungUeberein(t *testing.T) {
	t.Parallel()

	unser := einfarbig(5, 3, color.NRGBA{R: 12, G: 34, B: 56, A: 255})
	referenz := einfarbig(5, 3, color.NRGBA{R: 90, G: 8, B: 200, A: 255})

	zerlegung, err := render.ZerlegeAbstand(render.NeueHerkunft(5, 3), unser, referenz)
	if err != nil {
		t.Fatalf("ZerlegeAbstand: %v", err)
	}

	ganz, err := render.Ganzabstand(unser, referenz)
	if err != nil {
		t.Fatalf("Ganzabstand: %v", err)
	}

	if ganz != zerlegung.Gesamt {
		t.Errorf("Ganzabstand %+v, Zerlegung %+v", ganz, zerlegung.Gesamt)
	}
}

// TestGanzabstandWeistUnbrauchbareBilderAb prüft die Fehlerwege ohne Herkunft.
func TestGanzabstandWeistUnbrauchbareBilderAb(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name           string
		unser, referen image.Image
	}{
		{"kein Bild", nil, grau(2, 2, 0)},
		{"keine Referenz", grau(2, 2, 0), nil},
		{"verschiedene Größen", grau(2, 2, 0), grau(2, 3, 0)},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, err := render.Ganzabstand(fall.unser, fall.referen)
			if !errors.Is(err, render.ErrPlan) {
				t.Errorf("Fehler %v, erwartet ErrPlan", err)
			}
		})
	}
}

// TestZerlegeLiestEinTeilbildAnSeinemUrsprung ist die Gegenprobe zu einem
// Fehler, den dieses Projekt schon einmal hatte.
//
// **Ein [image.Image] muss nicht bei (0,0) beginnen.** `VergleicheAusserhalb`
// verglich beim ersten Anlauf `image.Pt(x, y)` gegen seine Aussparung, während
// es die Bildpunkte an `Bounds().Min + x` las — bei Bounds ab (10,10) sparte
// es damit die falsche Fläche aus. Angestrichen vom Zweitblick über Phase A.
//
// Hier ist derselbe Fall statisch richtig gebaut; ohne diesen Test belegte das
// niemand (angestrichen vom Zweitblick über Phase I).
func TestZerlegeLiestEinTeilbildAnSeinemUrsprung(t *testing.T) {
	t.Parallel()

	// Ein großes Bild, aus dem ein 4 × 4-Ausschnitt ab (10,10) genommen wird.
	gross := grau(20, 20, 100)
	for y := 10; y < 12; y++ {
		for x := 10; x < 14; x++ {
			gross.SetNRGBA(x, y, color.NRGBA{R: 110, G: 110, B: 110, A: 255})
		}
	}

	teil := gross.SubImage(image.Rect(10, 10, 14, 14))

	// Die Herkunft deckt die obere Hälfte mit einer eigenen Art — sie beginnt
	// **bei null**, wie jede Herkunft.
	_, herkunft := mitHerkunftGezeichnet(t, flaechenplan(4, 4,
		gefuellt(render.Rechteck{Breite: 4, Hoehe: 4},
			render.Farbe{R: 100, G: 100, B: 100, A: 255}),
		bildebene(render.ArtSymbol, render.Rechteck{Breite: 4, Hoehe: 2}, deckendesBild),
	))

	zerlegung, err := render.ZerlegeAbstand(herkunft, teil, grau(4, 4, 100))
	if err != nil {
		t.Fatalf("ZerlegeAbstand: %v", err)
	}

	// Die Abweichung liegt in der oberen Hälfte des **Ausschnitts**, also dort,
	// wo die Symbolebene gemalt hat.
	if anteil := zerlegung.Anteil(render.ArtSymbol); anteil != 1 {
		t.Errorf("Anteil der oberen Ebene %.3f, erwartet 1 — der Ausschnitt "+
			"wurde nicht an seinem Ursprung gelesen", anteil)
	}

	if zerlegung.Gesamt.Punkte != 16 {
		t.Errorf("%d Bildpunkte, erwartet 16", zerlegung.Gesamt.Punkte)
	}
}
