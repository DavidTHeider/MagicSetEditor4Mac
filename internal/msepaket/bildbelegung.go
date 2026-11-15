package msepaket

// **Was ein entschlüsseltes Bild im Arbeitsspeicher belegt** — PERF03
// (#2353),
// Befund A-19 des Audits 0.9.0: bis dahin zählte die Bildquelle Einträge, und
// unter „64 Bilder" ließen sich Gigabyte halten — je Bild ist ausgepackt bis
// zu 1 GiB erlaubt ([0160](../../docs/entscheidungen/0160-was-ein-bild-ausgepackt-kostet.md)).
//
// **Die Puffer der Bildpunkte, je Art so, wie `image` sie anlegt** — nicht
// die Größe der Datei: ein PNG von 100 KB kann entschlüsselt 30 MB belegen.
// Eine Art, die `image` nicht kennt, zählt vier Byte je Bildpunkt.

import (
	"image"
	"image/color"
	"reflect"
)

// farbwert ist, was ein Eintrag einer Palette belegt.
var farbwert = int64(reflect.TypeFor[color.Color]().Size())

// belegteBytes ist, was ein entschlüsseltes Bild belegt.
func belegteBytes(bild image.Image) int64 {
	switch b := bild.(type) {
	case *image.NRGBA:
		return int64(len(b.Pix))
	case *image.RGBA:
		return int64(len(b.Pix))
	case *image.NRGBA64:
		return int64(len(b.Pix))
	case *image.RGBA64:
		return int64(len(b.Pix))
	case *image.Gray:
		return int64(len(b.Pix))
	case *image.Gray16:
		return int64(len(b.Pix))
	case *image.Alpha:
		return int64(len(b.Pix))
	case *image.Alpha16:
		return int64(len(b.Pix))
	case *image.CMYK:
		return int64(len(b.Pix))
	case *image.Paletted:
		return int64(len(b.Pix)) + int64(len(b.Palette))*farbwert
	case *image.NYCbCrA:
		return int64(len(b.Y) + len(b.Cb) + len(b.Cr) + len(b.A))
	case *image.YCbCr:
		return int64(len(b.Y) + len(b.Cb) + len(b.Cr))
	}

	feld := bild.Bounds()

	return int64(feld.Dx()) * int64(feld.Dy()) * 4
}
