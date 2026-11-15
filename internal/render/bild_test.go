// Die Prüfungen über **Bilder** im Rasterlauf: Befunde, Teilbilder, Kanten.
//
// Von `rasterer_test.go` abgetrennt, weil die Datei über die Zeilengrenze aus
// AGENTS.md §2 gewachsen war. Der Schnitt folgt der Frage: dort steht, wie
// Farben übereinanderkommen, hier, was mit einem Bild geschieht.
package render_test

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Ein fehlendes Bild ist ein Befund mit Ebenennummer, kein leerer Bereich.
func TestEinFehlendesBildIstEinBefundMitStelle(t *testing.T) {
	faelle := []struct {
		name  string
		ebene render.Ebene
	}{
		{"ein Bild, das es nicht gibt", render.Ebene{
			Art: render.ArtBild, Rahmen: render.Rechteck{Breite: 1, Hoehe: 1},
			Bild: render.Blatt("/fehlt.png"), Deckung: 1,
		}},
		{"eine Maske, die es nicht gibt", render.Ebene{
			Art: render.ArtFlaeche, Rahmen: render.Rechteck{Breite: 1, Hoehe: 1},
			Farbe: render.Farbe{A: 255}, Maske: render.Blatt("/fehlt.png"), Deckung: 1,
		}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			plan := &render.Plan{Breite: 1, Hoehe: 1, Ebenen: []render.Ebene{
				{
					Art: render.ArtFlaeche, Rahmen: render.Rechteck{Breite: 1, Hoehe: 1},
					Farbe: render.Farbe{G: 255, A: 255}, Deckung: 1,
				},
				fall.ebene,
			}}

			bild, befunde, err := render.NeuerRasterer(bilder{}).Zeichne(t.Context(), plan, 1, render.GueteMittel)
			if err != nil {
				t.Fatalf("abgewiesen: %v", err)
			}

			if len(befunde) != 1 || befunde[0].Ebene != 2 {
				t.Fatalf("Befunde: %v", befunde)
			}

			// Und die erste Ebene steht noch.
			if got := farbeBei(t, bild, 0, 0); got != (render.Farbe{G: 255, A: 255}) {
				t.Errorf("%s — der Rest der Karte ging mit verloren", got)
			}
		})
	}
}

// Ein Bild wird in seinen Rahmen gerechnet, auch bei anderer Vergrößerung.
func TestEinBildWirdInSeinenRahmenGerechnet(t *testing.T) {
	quelle := einfarbig(2, 2, color.NRGBA{R: 10, G: 20, B: 30, A: 255})

	plan := &render.Plan{Breite: 4, Hoehe: 4, Ebenen: []render.Ebene{
		{
			Art: render.ArtBild, Rahmen: render.Rechteck{Breite: 4, Hoehe: 4},
			Bild: render.Blatt("/b.png"), Deckung: 1,
		},
	}}

	for _, vergroesserung := range []float64{1, 2, 0.5} {
		bild, befunde, err := render.NeuerRasterer(bilder{"/b.png": quelle}).
			Zeichne(t.Context(), plan, vergroesserung, render.GueteMittel)
		if err != nil || len(befunde) != 0 {
			t.Fatalf("%v / %v", befunde, err)
		}

		erwartet := int(4 * vergroesserung)
		if bild.Bounds().Dx() != erwartet {
			t.Errorf("Breite %d bei Vergrößerung %g, erwartet war %d",
				bild.Bounds().Dx(), vergroesserung, erwartet)
		}

		if got := farbeBei(t, bild, 0, 0); got != (render.Farbe{R: 10, G: 20, B: 30, A: 255}) {
			t.Errorf("%s bei Vergrößerung %g", got, vergroesserung)
		}
	}
}

// Ein Teilbild wird an seinem eigenen Ursprung abgetastet.
//
// Ein `image.Image` muss nicht bei (0,0) beginnen: ein `SubImage` trägt seinen
// Ursprung mit. Der erste Anlauf rechnete ab null und klemmte danach fast alles
// auf die linke obere Ecke — ein Teilbild wurde zu einer Fläche seiner Ecke.
// Gefunden vom Nachlauf des Zweitblicks über Phase 3.
func TestEinTeilbildWirdAnSeinemUrsprungAbgetastet(t *testing.T) {
	// Ein Bild, dessen linke Hälfte rot und rechte Hälfte grün ist.
	ganz := image.NewNRGBA(image.Rect(0, 0, 4, 2))
	for y := range 2 {
		for x := range 4 {
			farbe := color.NRGBA{R: 255, A: 255}
			if x >= 2 {
				farbe = color.NRGBA{G: 255, A: 255}
			}

			ganz.SetNRGBA(x, y, farbe)
		}
	}

	// Der Ausschnitt ist die **rechte** Hälfte: Bounds (2,0)–(4,2), also ganz grün.
	teil := ganz.SubImage(image.Rect(2, 0, 4, 2))

	plan := &render.Plan{Breite: 2, Hoehe: 2, Ebenen: []render.Ebene{
		{
			Art: render.ArtBild, Bild: render.Blatt("/teil.png"),
			Rahmen: render.Rechteck{Breite: 2, Hoehe: 2}, Deckung: 1,
		},
	}}

	bild, befunde, err := render.NeuerRasterer(bilder{"/teil.png": teil}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil || len(befunde) != 0 {
		t.Fatalf("%v / %v", befunde, err)
	}

	for _, punkt := range []struct{ x, y int }{{0, 0}, {1, 0}, {0, 1}, {1, 1}} {
		if got := farbeBei(t, bild, punkt.x, punkt.y); got.G != 255 || got.R != 0 {
			t.Errorf("bei (%d,%d) steht %s — der Ausschnitt wurde ab null gelesen",
				punkt.x, punkt.y, got)
		}
	}
}

// Ein Laufzeitbefund trägt die Nummer aus dem **ursprünglichen** Plan.
//
// Fällt Ebene 1 schon bei der Prüfung weg, meldete ein fehlendes Bild in Ebene 2
// sich als Ebene 1 — und wer danach in einer Vorlage mit 60 Ebenen sucht, sucht
// an der falschen. Gefunden vom Nachlauf des Zweitblicks über Phase 3.
func TestEinLaufzeitbefundTraegtDieUrspruenglicheNummer(t *testing.T) {
	plan := &render.Plan{Breite: 1, Hoehe: 1, Ebenen: []render.Ebene{
		// Ebene 1 fällt schon bei der Prüfung weg: ein Bild ohne Pfad.
		{Art: render.ArtBild, Rahmen: render.Rechteck{Breite: 1, Hoehe: 1}, Deckung: 1},
		// Ebene 2 fällt erst beim Zeichnen weg: das Bild gibt es nicht.
		{
			Art: render.ArtBild, Bild: render.Blatt("/fehlt.png"),
			Rahmen: render.Rechteck{Breite: 1, Hoehe: 1}, Deckung: 1,
		},
	}}

	_, befunde, err := render.NeuerRasterer(bilder{}).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	if len(befunde) != 2 {
		t.Fatalf("%d Befunde: %v", len(befunde), befunde)
	}

	if befunde[0].Ebene != 1 || befunde[1].Ebene != 2 {
		t.Errorf("Ebenen %d und %d, erwartet waren 1 und 2 — %v",
			befunde[0].Ebene, befunde[1].Ebene, befunde)
	}
}

// Eine Kante zwischen durchsichtig und deckend bekommt keinen Farbsaum.
//
// **Über vorgewichtete Kanäle gemischt.** Der erste Anlauf mischte die
// entwichteten Kanäle und das Alpha getrennt; dann färbt die Farbe eines
// Bildpunkts, den man gar nicht sieht, seinen Nachbarn. Gefunden vom Zweitblick
// über Phase 4.
func TestEineKanteZwischenDurchsichtigUndDeckendBekommtKeinenSaum(t *testing.T) {
	// Links fast durchsichtiges Rot, rechts deckendes Blau.
	//
	// **Fast, nicht ganz:** `punktfarbe` setzt einen völlig durchsichtigen Punkt
	// schon auf Null, und dann gäbe es nichts zu färben. Der erste Anlauf dieses
	// Tests nahm Alpha 0 und blieb deshalb auch mit der falschen Mischung grün —
	// eine Gegenprobe, die den Fall nicht erreicht.
	quelle := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	quelle.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 1})
	quelle.SetNRGBA(1, 0, color.NRGBA{B: 255, A: 255})

	// Vierfach vergrößert: dazwischen entstehen gemischte Bildpunkte.
	plan := &render.Plan{Breite: 8, Hoehe: 1, Ebenen: []render.Ebene{
		{
			Art: render.ArtBild, Bild: render.Blatt("/kante.png"),
			Rahmen: render.Rechteck{Breite: 8, Hoehe: 1}, Deckung: 1,
		},
	}}

	bild, befunde, err := render.NeuerRasterer(bilder{"/kante.png": quelle}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil || len(befunde) != 0 {
		t.Fatalf("%v / %v", befunde, err)
	}

	// **Geprüft wird, wo man etwas sieht.** Links steht die Quellfarbe selbst,
	// und die ist rot — mit Alpha 1. Das ist richtig so. Falsch wäre, wenn ein
	// **sichtbarer** Bildpunkt rot würde: dort hat die fast unsichtbare Farbe
	// den Nachbarn gefärbt.
	//
	// Der erste Anlauf verlangte „nirgends Rot" und schlug deshalb mit **beiden**
	// Mischungen fehl — eine Zusicherung, die auch das Richtige abweist, prüft
	// nichts.
	sichtbare := 0

	for x := range 8 {
		got := farbeBei(t, bild, x, 0)
		if got.A < 128 {
			continue
		}

		sichtbare++

		if got.R > 2 {
			t.Errorf("bei %d steht %s — die fast unsichtbare Farbe hat gefärbt", x, got)
		}
	}

	if sichtbare == 0 {
		t.Fatal("kein sichtbarer Bildpunkt — dann prüft der Test nichts")
	}
}

// **Ein Rasterer ohne Bildquelle meldet das, statt abzustürzen.**
//
// # Wie das aufgefallen ist
//
// Nicht durch einen Einfall, sondern durch die Gegenprobe zum Goldbild aus
// #173: dem
// Rasterer die Bildquelle wegzunehmen sollte den Rahmen verschwinden lassen und
// ließ den Lauf mit einem Zugriff auf `nil` abbrechen.
//
// `NeuerRasterer(nil)` ist kein erfundener Fall. Er steht so in
// `app.TestDieSaatkarteBleibtWieSieIst` und war dort folgenlos, weil jener Plan
// keine Bildebene trägt — die Zeile war seit Phase 4 eine Bildebene weit von
// einem Absturz entfernt.
//
// **Und §8 sagt, was stattdessen gilt**: ein fehlendes Bild ist ein benannter
// Vorbehalt. Eine fehlende **Quelle** ist derselbe Fall, eine Stufe früher.
//
// # Warum je ein Befund und nicht zwei
//
// Eine Ebene, deren Maske nicht zu haben ist, wird gar nicht erst gezeichnet;
// das Bild dazu zu melden, hieße von einem Versuch berichten, den es nicht gab.
// Deshalb zwei Fälle statt einem: sonst liefe der Bildweg nie durch.
func TestEinRastererOhneBildquelleMeldetJedeBildebene(t *testing.T) {
	faelle := map[string]struct {
		maske string
		wort  string
	}{
		"die Maske kommt zuerst": {maske: "/probe.mse-style/maske.png", wort: "Maske"},
		"ohne Maske das Bild":    {maske: "", wort: "Bild"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			plan := &render.Plan{Breite: 10, Hoehe: 10, Ebenen: []render.Ebene{{
				Art:     render.ArtBild,
				Rahmen:  render.Rechteck{Breite: 10, Hoehe: 10},
				Bild:    render.Blatt("/probe.mse-style/rahmen.png"),
				Maske:   render.Blatt(fall.maske),
				Deckung: 1,
			}}}

			bild, befunde, err := render.NeuerRasterer(nil).Zeichne(t.Context(), plan, 1, render.GueteMittel)
			if err != nil {
				t.Fatalf("zeichnen: %v", err)
			}

			if bild == nil {
				t.Fatal("kein Bild")
			}

			if len(befunde) != 1 {
				t.Fatalf("%d Befunde, erwartet einer: %v", len(befunde), befunde)
			}

			if !strings.Contains(befunde[0].Meldung, fall.wort) ||
				!strings.Contains(befunde[0].Meldung, "keine Bildquelle") {
				t.Errorf("Meldung %q", befunde[0].Meldung)
			}
		})
	}
}
