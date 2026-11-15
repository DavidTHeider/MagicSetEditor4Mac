// Die Fehlerpfade der Schriftbibliothek.
//
// **Paketintern, weil die Nähte hier liegen.** Sie setzen eine Schriftdatei
// voraus, die zerteilbar ist und trotzdem keine Metrik hat — mit einer echten
// Datei nicht herzustellen. Ein Zweig, den kein Test betritt, ist eine
// Behauptung darüber, was im Ernstfall geschieht.
package render

import (
	"errors"
	"image"
	"os"
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

var errNahtGestolpert = errors.New("die Naht ist planmäßig gescheitert")

// schriftbytes liest eine Schriftdatei vom System — oder überspringt.
func schriftbytes(t *testing.T) []byte {
	t.Helper()

	const pfad = "/System/Library/Fonts/Supplemental/Andale Mono.ttf"

	roh, err := os.ReadFile(pfad)
	if err != nil {
		t.Skipf("keine Schriftdatei unter %s: %v", pfad, err)
	}

	return roh
}

// **Eine Schrift ohne lesbare Metrik wird beim Laden abgewiesen** — nicht bei
// jedem Zeichen, und nicht stillschweigend mit Zeilenhöhe null.
func TestEineSchriftOhneMetrikWirdAbgewiesen(t *testing.T) {
	vorher := masseLesen
	masseLesen = func(*sfnt.Font, *sfnt.Buffer, fixed.Int26_6) (font.Metrics, error) {
		return font.Metrics{}, errNahtGestolpert
	}

	t.Cleanup(func() { masseLesen = vorher })

	if _, err := SchriftAusDatei(schriftbytes(t)); !errors.Is(err, errNahtGestolpert) {
		t.Errorf("Fehler %v", err)
	}
}

// Ein unlesbarer Vorschub ergibt Breite null — und keinen Absturz.
func TestEinUnlesbarerVorschubGibtBreiteNull(t *testing.T) {
	schrift, err := SchriftAusDatei(schriftbytes(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	vorher := vorschubLesen
	vorschubLesen = func(*sfnt.Font, *sfnt.Buffer, sfnt.GlyphIndex, fixed.Int26_6) (
		fixed.Int26_6, error,
	) {
		return 0, errNahtGestolpert
	}

	t.Cleanup(func() { vorschubLesen = vorher })

	if breite := schrift.Breite('A'); breite != 0 {
		t.Errorf("Breite %g", breite)
	}
}

// **Ein Umriss, der sich nicht lesen lässt, gibt keine Maske** — und die
// Textebene meldet das fehlende Zeichen, statt es zu erfinden.
func TestEinUnlesbarerUmrissGibtKeineMaske(t *testing.T) {
	schrift, err := SchriftAusDatei(schriftbytes(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	vorher := umrissLesen
	umrissLesen = func(*sfnt.Font, *sfnt.Buffer, sfnt.GlyphIndex, fixed.Int26_6) (sfnt.Segments, error) {
		return nil, errNahtGestolpert
	}

	t.Cleanup(func() { umrissLesen = vorher })

	if _, _, da := schrift.Maske('A', 12); da {
		t.Error("es kam eine Maske, obwohl kein Umriss zu lesen war")
	}
}

// **Auch ein kubischer Umriss wird gezeichnet** — Schriften mit CFF-Umrissen
// (`.otf`) beschreiben ihre Glyphen in Kubiken, und Go Regular hat keine. Ein
// Quadrat aus vier geraden Kubiken muss ein volles Quadrat geben; vertauschte
// Stützpunkte gäben eine andere Fläche.
func TestEinKubischerUmrissWirdGezeichnet(t *testing.T) {
	schrift, err := SchriftAusDatei(goregular.TTF)
	if err != nil {
		t.Fatalf("Go Regular: %v", err)
	}

	gerade := func(von, nach fixed.Point26_6) sfnt.Segment {
		schritt := nach.Sub(von)

		return sfnt.Segment{Op: sfnt.SegmentOpCubeTo, Args: [3]fixed.Point26_6{
			von.Add(schritt.Mul(fixed.I(1) / 3)), von.Add(schritt.Mul(fixed.I(2) / 3)), nach,
		}}
	}

	ecken := []fixed.Point26_6{fixed.P(0, -10), fixed.P(10, -10), fixed.P(10, 0), fixed.P(0, 0)}

	vorher := umrissLesen
	umrissLesen = func(*sfnt.Font, *sfnt.Buffer, sfnt.GlyphIndex, fixed.Int26_6) (sfnt.Segments, error) {
		umriss := make(sfnt.Segments, 0, len(ecken)+1)
		umriss = append(umriss, sfnt.Segment{Op: sfnt.SegmentOpMoveTo, Args: [3]fixed.Point26_6{ecken[0]}})

		for i := range ecken {
			umriss = append(umriss, gerade(ecken[i], ecken[(i+1)%len(ecken)]))
		}

		return umriss, nil
	}

	t.Cleanup(func() { umrissLesen = vorher })

	kennung, _ := schrift.kennung('A')

	maske, ecke, da := schrift.maskeDerKennung(kennung, 12, 0)
	if !da || ecke != image.Pt(0, -10) || maske.Bounds() != image.Rect(0, 0, 10, 10) {
		t.Fatalf("da %v, Ecke %v, Rahmen %v", da, ecke, maske.Bounds())
	}

	for y := range 10 {
		for x := range 10 {
			if deckung := alphaAn(maske, x, y).A; deckung != 255 {
				t.Fatalf("Bildpunkt %d,%d deckt %d statt 255", x, y, deckung)
			}
		}
	}
}

// **Eine Glyphe nach Nummer gleicht der nach Zeichen, Bildpunkt für
// Bildpunkt** — [OpenTypeSchrift.maskeDerKennung] geht den Weg von
// `opentype.Face.Glyph` selbst, weil eine Ligatur kein Zeichen hat (#1805).
// Ein Unterschied hier verschöbe jedes Goldbild und jeden Kanonlauf.
//
// **Über jede Glyphe von Go Regular**, die ein Zeichen bis U+2FFF hat, bei
// fünf Größen — ganzen und gebrochenen, denn der Maßstab wird gerundet.
func TestEineGlypheNachNummerGleichtDerNachZeichen(t *testing.T) {
	t.Parallel()

	schrift, err := SchriftAusDatei(goregular.TTF)
	if err != nil {
		t.Fatalf("Go Regular: %v", err)
	}

	verglichen := 0

	for _, groesse := range []float64{3.2, 7, 12.5, 48, 64.8} {
		gesicht, err := opentype.NewFace(schrift.schrift, &opentype.FaceOptions{
			Size: groesse, DPI: 72, Hinting: font.HintingNone,
		})
		if err != nil {
			t.Fatalf("Gesicht: %v", err)
		}

		for zeichen := rune(0x20); zeichen < 0x3000; zeichen++ {
			rahmen, erwartet, ecke, _, da := gesicht.Glyph(fixed.Point26_6{}, zeichen)

			kennung, gefunden := schrift.kennung(zeichen)

			var (
				maske image.Image
				lage  image.Point
			)

			if gefunden {
				maske, lage, gefunden = schrift.maskeDerKennung(kennung, groesse, 0)
			}

			if da != gefunden {
				t.Fatalf("%U bei %v: da %v, die Bibliothek %v", zeichen, groesse, gefunden, da)
			}

			if !da {
				continue
			}

			if lage != rahmen.Min || maske.Bounds().Size() != rahmen.Size() {
				t.Fatalf("%U bei %v: Lage %v %v, die Bibliothek %v", zeichen, groesse, lage, maske.Bounds(), rahmen)
			}

			for y := range rahmen.Dy() {
				for x := range rahmen.Dx() {
					if alphaAn(maske, x, y) != alphaAn(erwartet, ecke.X+x, ecke.Y+y) {
						t.Fatalf("%U bei %v: Bildpunkt %d,%d weicht ab", zeichen, groesse, x, y)
					}
				}
			}

			verglichen++
		}
	}

	// **Die Reichweite der Messung** (§10): ohne Glyphen verglichen wäre der
	// Lauf grün und sagte nichts.
	if verglichen < 3000 {
		t.Fatalf("nur %d Glyphen verglichen", verglichen)
	}
}

// Die Maske erfüllt `image.Image` vollständig.
//
// `ColorModel` ruft der Zeichner nie — er liest über `At`. Die Methode gehört
// trotzdem zur Schnittstelle, und eine, die niemand je aufruft, ist eine, die
// beim ersten Aufruf falsch sein darf.
func TestDieMaskeIstEinVollstaendigesBild(t *testing.T) {
	schrift, err := SchriftAusDatei(schriftbytes(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	maske, _, da := schrift.Maske('A', 12)
	if !da {
		t.Fatal("kein Glyph für A")
	}

	if maske.ColorModel() == nil {
		t.Error("die Maske nennt kein Farbmodell")
	}

	if maske.Bounds().Min != (image.Point{}) {
		t.Errorf("die Maske beginnt nicht bei (0,0): %v", maske.Bounds())
	}
}
