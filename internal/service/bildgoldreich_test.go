// **Was die Ausfuhr an einer Karte kostet, nicht an einer Halbfläche** —
// #851.
//
// # Warum eine zweite Golddatei
//
// `bildgold_test.go` misst den Weg **vom Bild zur Datei**: schreiben,
// zurücklesen, vergleichen. Dafür genügt eine Fläche, und ihr Plan trägt
// deshalb genau eine.
//
// Hier steht die andere Frage: **was das Format an einer Karte kostet**. Der
// Zweitblick über Phase P hat sie angestrichen — der „Stapel vieler Karten"
// war dreimal dieselbe rote Halbfläche auf durchsichtigem Grund, ohne Verlauf,
// ohne Rahmen, ohne schräge Kante. Eine an ihr gemessene Toleranz sagt über
// eine Karte nichts (§10).
package service_test

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// planReich gibt einen Plan mit dem, woran ein Bildformat scheitert.
//
// **Sechs Ebenen und nicht eine**, jede für eine andere Art von Kante:
//
//	Grund       eine volle Fläche          — der Bezug
//	Verlauf     weicher Übergang           — was JPEG gut kann
//	Ecke        runde Ecken                — gekrümmte Kante
//	Rand        ein Rahmen aus vier Bahnen — dünne Linien
//	Winkel      eine gedrehte Fläche       — die schräge Kante
//	Deckung     halbdurchsichtig           — der Alphakanal
//
// **Ohne Text und ohne Bild**, und das ist eine Grenze, die dazugehört: beides
// bräuchte eine Schriftsuche oder eine Bildquelle, und ein Goldbild, das an
// einer Systemschrift hängt, misst den Rechner mit. Was hier fehlt, steht in
// `docs/FORMATE.md` neben der Zahl.
//
// **Je Karte eine andere Farbe.** Drei gleiche Karten sind ein Stapel nur dem
// Namen nach; die Nummer geht in den Farbton ein, damit die drei Läufe
// wirklich drei Bilder vergleichen.
type planReich struct{}

func (planReich) Kartenplan(_ context.Context, karte *domain.CardRow) (
	*render.Plan, []domain.Befund, error,
) {
	const breite, hoehe = 100.0, 140.0

	ton := uint8(40 + karte.ID*60)

	return &render.Plan{
		Breite: breite, Hoehe: hoehe,
		Ebenen: []render.Ebene{
			{
				Art:     render.ArtFlaeche,
				Rahmen:  render.Rechteck{Breite: breite, Hoehe: hoehe},
				Farbe:   render.Farbe{R: 240, G: 238, B: 230, A: 255},
				Deckung: 1,
			},
			{
				Art:    render.ArtFlaeche,
				Rahmen: render.Rechteck{Oben: 10, Links: 6, Breite: 88, Hoehe: 50},
				Verlauf: &render.Verlauf{Stufen: []render.Verlaufsstufe{
					{Stelle: 0, Farbe: render.Farbe{R: ton, G: 30, B: 30, A: 255}},
					{Stelle: 1, Farbe: render.Farbe{R: 20, G: 40, B: ton, A: 255}},
				}},
				Deckung: 1,
			},
			{
				Art:     render.ArtFlaeche,
				Rahmen:  render.Rechteck{Oben: 66, Links: 10, Breite: 80, Hoehe: 30},
				Farbe:   render.Farbe{R: 20, G: 90, B: 40, A: 255},
				Ecke:    8,
				Deckung: 1,
			},
			{
				Art:     render.ArtFlaeche,
				Rahmen:  render.Rechteck{Oben: 4, Links: 4, Breite: 92, Hoehe: 132},
				Farbe:   render.Farbe{R: 10, G: 10, B: 10, A: 255},
				Rand:    &render.Randbreiten{Links: 2, Oben: 2, Rechts: 2, Unten: 2},
				Deckung: 1,
			},
			{
				Art:     render.ArtFlaeche,
				Rahmen:  render.Rechteck{Oben: 100, Links: 20, Breite: 60, Hoehe: 20},
				Farbe:   render.Farbe{R: 220, G: 200, B: 20, A: 255},
				Winkel:  30,
				Deckung: 1,
			},
			{
				Art:     render.ArtFlaeche,
				Rahmen:  render.Rechteck{Oben: 118, Links: 8, Breite: 84, Hoehe: 16},
				Farbe:   render.Farbe{R: 0, G: 0, B: 0, A: 255},
				Deckung: 0.5,
			},
		},
	}, nil, nil
}

// reicherStapel ist die Ausfuhr über [planReich].
func reicherStapel(t *testing.T) *service.Bildausfuhr {
	t.Helper()

	return service.NeueBildausfuhr(
		kartenDouble{leserDouble: karten("Feuerball", "Eisball", "Steinschlag")},
		planReich{}, render.NeuerRasterer(nil))
}

// schreibeReich fährt den reichen Stapel in ein Verzeichnis.
func schreibeReich(t *testing.T, ordner string, format service.Bildformat) []service.Bildname {
	t.Helper()

	ausfuhr := reicherStapel(t)

	namen, fehler := ausfuhr.Namen(t.Context(), 7, "{name}", format,
		domain.BildkonfliktNummerierenNeu, domain.Kartenauswahl{})
	if fehler != nil {
		t.Fatalf("Namen für %s: %v", format, fehler)
	}

	for _, name := range namen {
		if _, fehler := ausfuhr.Schreibe(t.Context(), name.KarteID, ordner,
			name.Datei, format, 0, domain.BildkonfliktNummerierenNeu, render.GueteMittel); fehler != nil {
			t.Fatalf("%s schreiben: %v", name.Datei, fehler)
		}
	}

	return namen
}

// Was ein Format an einer Karte kostet — **alle drei Zahlen**, nicht eine.
//
// # Warum drei Schwellen und nicht der Mittelwert
//
// `render.Vergleich.Innerhalb` sagt es an seiner eigenen Stelle: der
// Mittelwert fängt eine Änderung, die überall ein wenig ausmacht, die Spitze
// eine, die an wenigen Stellen viel ausmacht. Bis zum 2026-08-22 sicherte
// dieser Lauf **nur** den Mittelwert zu und protokollierte die anderen beiden;
// ein Fehler, der 5 % der Fläche zerstört und den Rest unberührt lässt, wäre
// durchgekommen.
//
// # Die Zahlen
//
// Gemessen am 2026-08-22 über den Stapel aus [planReich], 100 × 140 Punkte:
//
//	TIFF, BMP   Abstand 0,000   Spitze   0   betroffen   0 %
//	JPG Güte 90 Abstand 1,48    Spitze  40   betroffen  91 %
//	JPG Güte 30 Abstand 5,34    Spitze 112   betroffen 100 %
//
// **Das ist eine andere Größenordnung als die 0,077 aus `bildgold_test.go`**,
// und daran hängt der ganze Befund: dort steht eine einfarbige Halbfläche, und
// die kostet JPEG fast nichts. Eine Karte hat einen Verlauf, einen Rahmen aus
// zwei Punkt breiten Bahnen und eine schräge Kante — dort kostet es das
// Zwanzigfache.
const (
	// KartenAbstandJPG ist der mittlere Abstand, den Güte 90 an einer Karte
	// kostet. Gemessen 1,479 im schlechtesten der drei Läufe.
	KartenAbstandJPG = 1.6

	// KartenSpitzeJPG ist die größte Abweichung eines einzelnen Kanals.
	// Gemessen 40; Güte 30 gibt 112.
	KartenSpitzeJPG = 48

	// KartenAnteilJPG ist der Anteil überhaupt veränderter Bildpunkte.
	//
	// **Die schwächste der drei Schwellen, und das gehört dazu.** Bei Güte 90
	// ändert JPEG ohnehin fast jeden Punkt um einen Schritt — gemessen 91 %,
	// und Güte 30 kommt auf 100 %. Sie trennt also nur wenig. Was sie fängt,
	// ist eine Änderung, die den **letzten** unveränderten Rest auch noch
	// bewegt; die Arbeit tun [KartenAbstandJPG] und [KartenSpitzeJPG].
	KartenAnteilJPG = 0.95
)

// preisAnEinerKarte misst ein Format gegen den PNG-Lauf.
func preisAnEinerKarte(t *testing.T, format service.Bildformat) []render.Vergleich {
	t.Helper()

	gold := map[int64]image.Image{}
	ordnerPNG := t.TempDir()

	for _, name := range schreibeReich(t, ordnerPNG, service.FormatPNG) {
		gold[name.KarteID] = lies(t, filepath.Join(ordnerPNG, name.Datei))
	}

	ordner := t.TempDir()
	namen := schreibeReich(t, ordner, format)
	vergleiche := make([]render.Vergleich, 0, len(namen))

	for _, name := range namen {
		zurueck := lies(t, filepath.Join(ordner, name.Datei))

		// **TIFF behält Alpha, die anderen nicht** — der Maßstab wird für sie
		// im Test selbst auf Weiß gelegt, nicht über `PackeBild`. Wer die zu
		// prüfende Seite auch als Maßstab nähme, verglich sie mit sich selbst.
		maszstab := gold[name.KarteID]
		if format != service.FormatTIFF {
			maszstab = aufWeissGelegt(maszstab)
		}

		vergleich, fehler := render.Vergleiche(maszstab, zurueck)
		if fehler != nil {
			t.Fatalf("%s vergleichen: %v", name.Datei, fehler)
		}

		t.Logf("%s %s: %s", format, name.Datei, vergleich)

		vergleiche = append(vergleiche, vergleich)
	}

	return vergleiche
}

// haelt prüft einen Vergleich gegen alle drei Schwellen.
func haelt(t *testing.T, was string, v render.Vergleich, abstand float64, spitze int, anteil float64) {
	t.Helper()

	if v.Abstand > abstand {
		t.Errorf("%s: Abstand %.3f über %g", was, v.Abstand, abstand)
	}

	if v.Groesster > spitze {
		t.Errorf("%s: Spitze %d über %d", was, v.Groesster, spitze)
	}

	if teil := float64(v.Abweichende) / float64(v.Punkte); teil > anteil {
		t.Errorf("%s: %.4f der Fläche betroffen, über %g", was, teil, anteil)
	}
}

func TestJedesFormatAnEinerKarte(t *testing.T) {
	faelle := []struct {
		format  service.Bildformat
		abstand float64
		spitze  int
		anteil  float64
	}{
		// **Verlustfrei heißt null, nicht „wenig".** Weicht dort ein
		// einzelner Bildpunkt ab, stimmt etwas nicht.
		{service.FormatTIFF, 0, 0, 0},
		{service.FormatBMP, 0, 0, 0},
		{service.FormatJPG, KartenAbstandJPG, KartenSpitzeJPG, KartenAnteilJPG},
	}

	for _, fall := range faelle {
		t.Run(string(fall.format), func(t *testing.T) {
			for stelle, vergleich := range preisAnEinerKarte(t, fall.format) {
				haelt(t, fmt.Sprintf("%s, Karte %d", fall.format, stelle+1),
					vergleich, fall.abstand, fall.spitze, fall.anteil)
			}
		})
	}
}

// TestGuete30BrichtJedeDerDreiSchwellen ist die Gegenprobe — **ausgeführt**.
//
// # Warum sie im Lauf steht und nicht im Kommentar
//
// **Eine Gegenprobe, die nichts ändert, ist keine** (§10). Bis zum 2026-08-22
// stand hier nur der Satz „mit Güte 30 steigt der Abstand auf 0,846"; wer die
// Schwelle später großzügiger setzte, bekam davon nichts zu sehen. Jetzt
// scheitert der Lauf, sobald eine der drei Schwellen so weit steht, dass sie
// auch die schlechtere Güte durchließe.
//
// **Gepackt wird hier im Test**, weil [service.JPGGuete] eine Konstante ist.
// Der Weg ist derselbe wie in `PackeBild`: auf Weiß legen, dann kodieren.
func TestGuete30BrichtJedeDerDreiSchwellen(t *testing.T) {
	ordnerPNG := t.TempDir()

	for _, name := range schreibeReich(t, ordnerPNG, service.FormatPNG) {
		auf := aufWeissGelegt(lies(t, filepath.Join(ordnerPNG, name.Datei)))

		var schlecht bytes.Buffer
		if fehler := jpeg.Encode(&schlecht, auf, &jpeg.Options{Quality: 30}); fehler != nil {
			t.Fatalf("Güte 30 packen: %v", fehler)
		}

		zurueck, fehler := jpeg.Decode(bytes.NewReader(schlecht.Bytes()))
		if fehler != nil {
			t.Fatalf("Güte 30 lesen: %v", fehler)
		}

		vergleich, fehler := render.Vergleiche(auf, zurueck)
		if fehler != nil {
			t.Fatalf("Güte 30 vergleichen: %v", fehler)
		}

		t.Logf("Güte 30, %s: %s", name.Datei, vergleich)

		if vergleich.Abstand <= KartenAbstandJPG {
			t.Errorf("%s: Güte 30 hält den Abstand %.3f — die Schwelle %g ist zu weit",
				name.Datei, vergleich.Abstand, KartenAbstandJPG)
		}

		if vergleich.Groesster <= KartenSpitzeJPG {
			t.Errorf("%s: Güte 30 hält die Spitze %d — die Schwelle %d ist zu weit",
				name.Datei, vergleich.Groesster, KartenSpitzeJPG)
		}

		if teil := float64(vergleich.Abweichende) / float64(vergleich.Punkte); teil <= KartenAnteilJPG {
			t.Errorf("%s: Güte 30 hält den Anteil %.4f — die Schwelle %g ist zu weit",
				name.Datei, teil, KartenAnteilJPG)
		}
	}
}
