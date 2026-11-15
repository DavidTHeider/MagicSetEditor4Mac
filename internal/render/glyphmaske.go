package render

import (
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// kennung gibt die Glyphe eines Zeichens — auch die eines Glyphenzeichens
// ([glyphenzeichen]). Falsch heißt: die Schrift führt es nicht.
//
// **Unter der Sperre der Schrift**: `GlyphIndex` benutzt ihren Puffer.
func (s *OpenTypeSchrift) kennung(zeichen rune) (sfnt.GlyphIndex, bool) {
	if nummer, glyphe := glyphennummer(zeichen); glyphe {
		if nummer == 0 || int(nummer) >= s.schrift.NumGlyphs() {
			return 0, false
		}

		return sfnt.GlyphIndex(nummer), true
	}

	kennung, err := s.schrift.GlyphIndex(&s.puffer, zeichen)
	if err != nil || kennung == 0 {
		return 0, false
	}

	return kennung, true
}

// Fuehrt sagt, ob die Schrift eine Glyphe für das Zeichen hat — die Frage,
// an der das Messen ausweicht ([Zeichenfuehrer]).
func (s *OpenTypeSchrift) Fuehrt(zeichen rune) bool {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	_, da := s.kennung(zeichen)

	return da
}

// Maske rastert ein Zeichen — [Glyphenschrift].
func (s *OpenTypeSchrift) Maske(zeichen rune, groesse float64) (
	image.Image, image.Point, bool,
) {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	kennung, da := s.kennung(zeichen)
	if !da {
		return nil, image.Point{}, false
	}

	// **Wie Quartz sie zeichnet**, samt Schriftglättung (#1806) — die Maske
	// ist für den vierfachen Weg MSEs, und dort ist `groesse` die Geviertgröße
	// in Bildpunkten.
	return s.maskeDerKennung(kennung, groesse, quartzVerdickung(groesse))
}

// maskeDerKennung rastert eine Glyphe nach ihrer Nummer, ihr Umriss um
// `verdickung` Bildpunkte je Seite verdickt ([verdickt]).
//
// **Derselbe Weg wie `opentype.Face.Glyph`** (`golang.org/x/image` v0.45.0,
// `font/opentype/opentype.go`, Zeilen 153–263 — gleich mit v0.44.0, verglichen
// beim Heben am 2026-09-23), nur nach Nummer statt nach
// Zeichen — eine Ligatur hat keines
// (#1805).
// Der Ursprung liegt auf null, der Maßstab ist der von `opentype.NewFace` bei
// 72 dpi, und `TestEineGlypheNachNummerGleichtDerNachZeichen` hält jeden
// Bildpunkt gegen die Bibliothek.
//
// **72 dpi und keine Hinweise.** Bei 72 dpi ist ein Punkt ein Bildpunkt, und
// der Plan misst in Vorlagenpunkten; die Vergrößerung steckt schon in der
// Größe, die hier ankommt. Hinting verschöbe Umrisse abhängig vom Raster —
// und damit Goldbilder, die bei anderer Vergrößerung entstanden sind.
//
// **Eine eigene Maske je Aufruf.** Das Gesicht der Bibliothek schrieb in
// einen Puffer, den der nächste Aufruf überschrieb, und die Maske musste
// deshalb abgeschrieben werden; hier gehört sie dem Aufrufer von Anfang an.
func (s *OpenTypeSchrift) maskeDerKennung(kennung sfnt.GlyphIndex, groesse, verdickung float64) (
	image.Image, image.Point, bool,
) {
	umriss, err := umrissLesen(s.schrift, &s.puffer, kennung, fixed.Int26_6(0.5+groesse*64))
	if err != nil {
		return nil, image.Point{}, false
	}

	if verdickung != 0 {
		umriss = verdickt(umriss, verdickung)
	}

	grenzen := umriss.Bounds()
	rahmen := image.Rect(grenzen.Min.X.Floor(), grenzen.Min.Y.Floor(),
		grenzen.Max.X.Ceil(), grenzen.Max.Y.Ceil())

	versatz := fixed.Point26_6{X: -fixed.I(rahmen.Min.X), Y: -fixed.I(rahmen.Min.Y)}
	punkt := func(p fixed.Point26_6) (float32, float32) {
		return float32(p.X+versatz.X) / 64, float32(p.Y+versatz.Y) / 64
	}

	maske := image.NewAlpha(image.Rectangle{Max: rahmen.Size()})

	s.raster.Reset(rahmen.Dx(), rahmen.Dy())
	s.raster.DrawOp = draw.Src

	for _, stueck := range umriss {
		switch stueck.Op {
		case sfnt.SegmentOpMoveTo:
			s.raster.MoveTo(punkt(stueck.Args[0]))

		case sfnt.SegmentOpLineTo:
			s.raster.LineTo(punkt(stueck.Args[0]))

		case sfnt.SegmentOpQuadTo:
			bx, by := punkt(stueck.Args[0])
			cx, cy := punkt(stueck.Args[1])
			s.raster.QuadTo(bx, by, cx, cy)

		case sfnt.SegmentOpCubeTo:
			bx, by := punkt(stueck.Args[0])
			cx, cy := punkt(stueck.Args[1])
			dx, dy := punkt(stueck.Args[2])
			s.raster.CubeTo(bx, by, cx, cy, dx, dy)
		}
	}

	s.raster.Draw(maske, maske.Bounds(), image.Opaque, image.Point{})

	return maske, rahmen.Min, true
}

// alphaAn liest die Deckung eines Bildpunkts.
//
// `RGBA` gibt 16 Bit je Kanal; die Maske führt 8. Geteilt wird durch 257 statt
// nach rechts geschoben — das ist dasselbe Ergebnis für jeden gültigen Wert und
// braucht keine Umwandlung, an der ein Prüfwerkzeug zu Recht hängen bleibt.
func alphaAn(bild image.Image, x, y int) color.Alpha {
	_, _, _, deckung := bild.At(x, y).RGBA()

	return color.Alpha{A: byte(min(deckung/257, 255))}
}
