// **Ein Goldvergleich je Format, für viele Karten** (#839).
//
// # Was hier geprüft wird und was nicht
//
// **Nicht die Rasterung** — die hat ihren eigenen Goldvergleich in
// `internal/render`. Hier geht es um den Weg **danach**: kommt aus der
// geschriebenen Datei dasselbe Bild wieder heraus?
//
// Das ist keine Selbstverständlichkeit. JPG kennt keine Durchsichtigkeit und
// verliert bei Güte 90; BMP kennt sie ebenfalls nicht; TIFF und PNG behalten
// sie. Vier Formate, vier Erwartungen — und sie stehen als Zahl da, nicht als
// Zuversicht.
package service_test

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// JPGToleranz ist der Abstand, den Güte 90 kostet.
//
// **Gemessen und dann festgeschrieben, nicht geschätzt.** Eine Toleranz, die
// jemand großzügig gewählt hat, deckt beim nächsten Fehler auch den zu. Der
// Wert unten entstand aus dem Lauf selbst; die Gegenprobe senkt die Güte auf 30
// und sieht den roten Lauf.
//
// **Gemessen am 2026-08-22**: Abstand **0,077**, größte Abweichung **6**, 10 %
// der Bildpunkte betroffen. Die Schwelle steht knapp darüber — wer sie
// großzügig setzte, deckte beim nächsten Fehler auch den zu. Der erste Anlauf
// stand auf 6,0 und hätte fast jede Verschlechterung durchgelassen.
//
// **Warum der Mittelwert und nicht die Spitze**: JPEG rechnet in Blöcken von
// 8×8, und an der Kante zwischen zwei Farben weicht ein einzelner Bildpunkt um
// 6 ab. Genau das ist der Preis des Formats. Die betroffene **Fläche** sagt,
// ob er bezahlt wurde oder ob etwas anderes schiefging: zehn Prozent sind die
// Kante, achtzig wären ein anderer Fehler.
const JPGToleranz = 0.1

// **Und die beiden anderen Zahlen, seit #851 ebenfalls zugesichert.**
//
// Sie standen im Kopf oben und in `docs/FORMATE.md` als Messvertrag, geprüft
// wurde davon nur der Mittelwert; der Zweitblick über Phase P hat es
// angestrichen. Gemessen am 2026-08-22: Spitze **6**, betroffen **10 %**.
//
// **Was sie über eine Karte sagen: nichts.** Der Plan hier ist eine
// einfarbige Halbfläche, und JPEG kostet daran fast nichts. Die Zahlen für
// eine Karte stehen in `bildgoldreich_test.go` und sind zwanzigmal so hoch.
const (
	JPGSpitze = 8
	JPGAnteil = 0.12
)

// **PNG, TIFF und BMP tragen null**, und das ist keine Nachsicht, sondern eine
// Zusage: alle drei sind verlustfrei. Weicht dort auch nur ein Bildpunkt ab,
// stimmt etwas nicht — und der Lauf sagt es sofort.

// vieleKarten ist der Stapel, über den jeder Lauf hier geht.
//
// **Mehr als eine**, denn der Abnahmekasten verlangt „für eine Karte **und**
// für viele": eine Namensvergabe, die bei zweien zusammenstößt, fällt bei einer
// nicht auf.
func vieleKarten(t *testing.T) *service.Bildausfuhr {
	t.Helper()

	return service.NeueBildausfuhr(
		kartenDouble{leserDouble: karten("Feuerball", "Eisball", "Steinschlag")},
		planDouble{}, render.NeuerRasterer(nil))
}

// lies gibt das Bild einer geschriebenen Datei.
func lies(t *testing.T, pfad string) image.Image {
	t.Helper()

	roh, fehler := os.ReadFile(pfad) // #nosec G304 -- Pfad aus t.TempDir()
	if fehler != nil {
		t.Fatalf("%s lesen: %v", pfad, fehler)
	}

	nach := map[string]func(*bytes.Reader) (image.Image, error){
		".png": func(r *bytes.Reader) (image.Image, error) { return png.Decode(r) },
		".jpg": func(r *bytes.Reader) (image.Image, error) { return jpeg.Decode(r) },
		".bmp": func(r *bytes.Reader) (image.Image, error) { return bmp.Decode(r) },
		".tif": func(r *bytes.Reader) (image.Image, error) { return tiff.Decode(r) },
	}

	entpacke, bekannt := nach[filepath.Ext(pfad)]
	if !bekannt {
		t.Fatalf("%s trägt keine bekannte Endung", pfad)
	}

	bild, fehler := entpacke(bytes.NewReader(roh))
	if fehler != nil {
		t.Fatalf("%s entpacken: %v", pfad, fehler)
	}

	return bild
}

// aufWeissGelegt legt ein Bild auf einen weißen Grund.
//
// **Im Test selbst gerechnet und nicht über `PackeBild`.** Wer die zu prüfende
// Seite auch als Maßstab nähme, verglich sie mit sich selbst — und der Lauf
// bliebe grün, gleich was sie tut.
func aufWeissGelegt(bild image.Image) image.Image {
	grund := image.NewNRGBA(bild.Bounds())
	draw.Draw(grund, grund.Bounds(), &image.Uniform{C: color.White},
		image.Point{}, draw.Src)
	draw.Draw(grund, grund.Bounds(), bild, bild.Bounds().Min, draw.Over)

	return grund
}

// **Jedes Format schreibt dasselbe Bild — je Format mit seiner Toleranz.**
//
// # Woher das Goldbild kommt
//
// **Aus dem PNG-Lauf.** Es ist verlustfrei und behält den Alphakanal; damit ist
// es der Maßstab, gegen den sich die drei anderen messen lassen — und es
// entsteht auf demselben Weg, den ein Anwender geht, nicht über eine
// Testabkürzung.
//
// Für BMP und JPG wird es **im Test** auf Weiß gelegt: beide kennen keine
// Durchsichtigkeit, und ein Vergleich gegen ein durchsichtiges Goldbild misst
// genau das — nicht, ob das Format sein Bild richtig trägt.
func TestJedesFormatSchreibtDasselbeBild(t *testing.T) {
	gold := map[int64]image.Image{}
	ordnerPNG := t.TempDir()

	namenPNG := schreibeStapel(t, ordnerPNG, service.FormatPNG)
	for _, name := range namenPNG {
		gold[name.KarteID] = lies(t, filepath.Join(ordnerPNG, name.Datei))
	}

	faelle := []struct {
		format   service.Bildformat
		aufWeiss bool
		toleranz float64
		spitze   int
		anteil   float64
	}{
		{service.FormatTIFF, false, 0, 0, 0},
		{service.FormatBMP, true, 0, 0, 0},
		{service.FormatJPG, true, JPGToleranz, JPGSpitze, JPGAnteil},
	}

	for _, fall := range faelle {
		t.Run(string(fall.format), func(t *testing.T) {
			ordner := t.TempDir()

			for _, name := range schreibeStapel(t, ordner, fall.format) {
				zurueck := lies(t, filepath.Join(ordner, name.Datei))

				maszstab := gold[name.KarteID]
				if fall.aufWeiss {
					maszstab = aufWeissGelegt(maszstab)
				}

				vergleich, fehler := render.Vergleiche(maszstab, zurueck)
				if fehler != nil {
					t.Fatalf("%s vergleichen: %v", name.Datei, fehler)
				}

				// **Alle drei Zahlen, nicht eine** (#851). Bis zum
				// 2026-08-22 stand hier allein der Mittelwert, während der
				// Kopf von [JPGToleranz] und `docs/FORMATE.md` drei nannten
				// — ein Wächter, der die halbe Zusage hält, macht die
				// andere unsichtbar (§10).
				if vergleich.Abstand > fall.toleranz {
					t.Errorf("%s: %s — Abstand über %g",
						name.Datei, vergleich, fall.toleranz)
				}

				if vergleich.Groesster > fall.spitze {
					t.Errorf("%s: %s — Spitze über %d",
						name.Datei, vergleich, fall.spitze)
				}

				if teil := float64(vergleich.Abweichende) / float64(vergleich.Punkte); teil > fall.anteil {
					t.Errorf("%s: %s — Anteil über %g",
						name.Datei, vergleich, fall.anteil)
				}

				t.Logf("%s: %s", name.Datei, vergleich)
			}
		})
	}
}

// schreibeStapel fährt einen ganzen Stapel in ein Verzeichnis.
func schreibeStapel(t *testing.T, ordner string, format service.Bildformat) []service.Bildname {
	t.Helper()

	ausfuhr := vieleKarten(t)

	namen, fehler := ausfuhr.Namen(t.Context(), 7, "{name}", format,
		domain.BildkonfliktNummerierenNeu, domain.Kartenauswahl{})
	if fehler != nil {
		t.Fatalf("Namen für %s: %v", format, fehler)
	}

	// **Drei Karten und nicht eine**: der Abnahmekasten verlangt „für eine
	// Karte **und** für viele", und eine Namensvergabe, die bei zweien
	// zusammenstößt, fällt bei einer nicht auf.
	if len(namen) != 3 {
		t.Fatalf("%d Namen, erwartet waren drei", len(namen))
	}

	for _, name := range namen {
		if _, fehler := ausfuhr.Schreibe(t.Context(), name.KarteID, ordner,
			name.Datei, format, 0, domain.BildkonfliktNummerierenNeu, render.GueteMittel); fehler != nil {
			t.Fatalf("%s schreiben: %v", name.Datei, fehler)
		}
	}

	return namen
}
