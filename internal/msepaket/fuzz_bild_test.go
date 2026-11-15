package msepaket_test

// **Ein Bild aus einem fremden Paket** — QA-06b
// (#1967),
// Fläche 7 aus `docs/FUZZZIELE.md`: [msepaket.DekodiereBild], die eine
// Stelle, die auspackt.
//
// # Die Formate des Programms, nicht die des Pakets
//
// **Was ausgepackt wird, entscheidet die Registrierung**, und das Programm
// registriert neben GIF, JPEG und PNG auch BMP, TIFF und WebP
// (`internal/service/dateiansicht.go`). Dieses Ziel holt dieselben herein —
// ein Testlauf nur mit den dreien dieses Pakets prüfte ein Programm, das es
// nicht gibt.
//
// **PNG trägt Prüfsummen**, und ein Fuzzer, der ein Maß im Kopf ändert,
// scheiterte daran, bevor der Kopf gelesen wird. Auf Wunsch (`heilen`) setzt
// das Ziel deshalb jede Prüfsumme neu — sonst sähe es von PNG nur die
// Ablehnung.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - zweimal dieselben Bytes, zwei Ergebnisse
//   - **ein angenommenes Bild, das größer ist, als vermessen wurde** — dann
//     hielte die Grenze nicht, was „erst vermessen, dann ausgepackt" sagt
//   - ein angenommenes Bild über [msepaket.GrenzeBildseite],
//     [msepaket.GrenzeBildpunkte] oder [msepaket.GrenzeAuspacken]; ein
//     abgewiesenes Maß über der Grenze ohne [msepaket.ErrBildZuGross]
//   - **mehr Speicher, als [msepaket.Auspackkosten] schätzt** — dazu ein
//     Sockel von 1 MiB und das Vierfache der Eingabe. Seit FMT-06b
//     (#1974)
//     ist das die gemessene Schätzung je Format und Farbmodell, und dieselbe
//     Zahl hält die Grenze; vorher standen hier pauschal 24 Byte
//
// # Warum es einmal 24 Byte waren und nicht 4
//
// **Hier stand vor dem ersten Lauf: 4 Byte je vermessenem Bildpunkt** — der
// Satz aus `paketgrenzen.go`, hundert Millionen Bildpunkte kosteten 400 MB.
// Der Lauf fand ein PNG mit 16 Bit je Kanal bei gut 8 Byte, und der Satz war
// falsch und nicht das Programm: die Grenze zählt Bildpunkte, wie sie sagt.
// Das Teuerste, was die registrierten Kodierer anlegen, ist ein progressives
// JPEG mit vier Farbauszügen — je Auszug 4 Byte Koeffizienten (`image/jpeg`,
// `scan.go` Zeile 167, Go 1.27.1), dazu das Bild, der Schwarzauszug und das
// CMYK-Bild. **Die Zusage hält damit, was die Kodierer brauchen, und nicht,
// was `paketgrenzen.go` einmal versprach**; ob die Grenze in Byte zählen
// soll, steht offen im Bericht zu QA-06b
// (#1967).
// Was sie weiter fängt, ist die Klasse des ersten Funds: Speicher, der mit
// keinem Bildpunkt zu tun hat.
//
// # Was nicht zugesichert wird
//
// **Der Bereich zwischen 4 Millionen Bildpunkten und der Grenze wird nicht
// ausgepackt.** Er ist erlaubt und kostet bis zu 2,4 GB je Ausführung; ein
// Fuzzer, der dort landet, legte den Rechner lahm, ohne etwas zu sagen, was
// das Verhältnis bei 4 Millionen nicht schon sagt.
//
//	go test ./internal/msepaket -run Fuzz -fuzz FuzzDekodiereBild -fuzztime 2m

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"runtime"
	"testing"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	_ "golang.org/x/image/webp" // dieselbe Registrierung wie das Programm

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// ausgepacktHoechstens ist, bis wohin das Ziel wirklich auspackt.
const ausgepacktHoechstens = 4 << 20

// **Was ausgepackt wird, ist vermessen — und kostet nicht mehr, als die
// Grenze sagt.**
func FuzzDekodiereBild(f *testing.F) {
	for _, saat := range bildsaat(f) {
		f.Add(saat, false)
	}

	f.Add([]byte("\x89PNG\r\n\x1a\n"), true)
	f.Add([]byte("GIF89a"), false)
	f.Add([]byte{}, false)

	f.Fuzz(func(t *testing.T, roh []byte, heilen bool) {
		if heilen {
			roh = pngGeheilt(roh)
		}

		mass, kopfart, vermessen := image.DecodeConfig(bytes.NewReader(roh))
		flaeche := int64(mass.Width) * int64(mass.Height)

		zuGross := mass.Width > msepaket.GrenzeBildseite || mass.Height > msepaket.GrenzeBildseite ||
			flaeche > msepaket.GrenzeBildpunkte || msepaket.Auspackkosten(mass, kopfart) > msepaket.GrenzeAuspacken
		if vermessen == nil && !zuGross && flaeche > ausgepacktHoechstens {
			return
		}

		var vorher, nachher runtime.MemStats

		runtime.ReadMemStats(&vorher)

		bild, art, fehler := msepaket.DekodiereBild(roh)

		runtime.ReadMemStats(&nachher)

		pruefeBildantwort(t, roh, bild, art, fehler, mass, vermessen, zuGross)

		if zugesagt := uint64(msepaket.Auspackkosten(mass, kopfart)) + 1<<20 + 4*uint64(len(roh)); nachher.TotalAlloc-vorher.TotalAlloc > zugesagt { //nolint:gosec // geschätzt, höchstens ein Gibibyte
			t.Errorf("%s, %d × %d: %d Byte zugeteilt, zugesagt %d", art, mass.Width, mass.Height,
				nachher.TotalAlloc-vorher.TotalAlloc, zugesagt)
		}
	})
}

// pruefeBildantwort hält die Antwort gegen die eigene Vermessung.
func pruefeBildantwort(t *testing.T, roh []byte, bild image.Image, art string, fehler error,
	mass image.Config, vermessen error, zuGross bool,
) {
	t.Helper()

	nochmal, nochmalArt, nochmalFehler := msepaket.DekodiereBild(roh)
	if (fehler == nil) != (nochmalFehler == nil) || art != nochmalArt ||
		(fehler != nil && fehler.Error() != nochmalFehler.Error()) ||
		(bild != nil && nochmal.Bounds() != bild.Bounds()) {
		t.Fatalf("dieselben Bytes, zwei Ergebnisse: %v / %v", fehler, nochmalFehler)
	}

	if vermessen == nil && zuGross && !errors.Is(fehler, msepaket.ErrBildZuGross) {
		t.Errorf("%d × %d über der Grenze, und der Fehler ist %v", mass.Width, mass.Height, fehler)
	}

	if fehler != nil {
		return
	}

	if bild == nil || art == "" {
		t.Fatalf("angenommen ohne Bild oder Art: %v, %q", bild, art)
	}

	if vermessen != nil {
		t.Fatalf("angenommen, obwohl der Kopf nicht zu vermessen war: %v", vermessen)
	}

	grenzen := bild.Bounds()
	if grenzen.Dx() > mass.Width || grenzen.Dy() > mass.Height {
		t.Errorf("%s: vermessen %d × %d, ausgepackt %v", art, mass.Width, mass.Height, grenzen)
	}

	if zuGross {
		t.Errorf("%s: %d × %d über der Grenze und angenommen", art, mass.Width, mass.Height)
	}
}

// bildsaat sind kleine Bilder in jedem Format, das ein Kodierer hier schreibt.
func bildsaat(f *testing.F) [][]byte {
	f.Helper()

	bunt := image.NewNRGBA(image.Rect(0, 0, 5, 3))
	bunt.SetNRGBA(4, 2, color.NRGBA{R: 200, G: 10, A: 128})

	tief := image.NewNRGBA64(image.Rect(0, 0, 3, 2))
	tief.SetNRGBA64(1, 1, color.NRGBA64{R: 0xffff, A: 0x8000})

	grau := image.NewGray(image.Rect(0, 0, 4, 4))
	zuBreit := image.NewGray(image.Rect(0, 0, msepaket.GrenzeBildseite+1, 1))
	palette := image.NewPaletted(image.Rect(0, 0, 4, 2), color.Palette{color.Black, color.White})

	schreiber := []func(*bytes.Buffer) error{
		func(b *bytes.Buffer) error { return png.Encode(b, bunt) },
		func(b *bytes.Buffer) error { return png.Encode(b, tief) },
		func(b *bytes.Buffer) error { return png.Encode(b, grau) },
		func(b *bytes.Buffer) error { return png.Encode(b, zuBreit) },
		func(b *bytes.Buffer) error { return png.Encode(b, palette) },
		func(b *bytes.Buffer) error { return jpeg.Encode(b, bunt, &jpeg.Options{Quality: 50}) },
		func(b *bytes.Buffer) error { return jpeg.Encode(b, grau, nil) },
		func(b *bytes.Buffer) error { return gif.Encode(b, palette, nil) },
		func(b *bytes.Buffer) error { return bmp.Encode(b, bunt) },
		func(b *bytes.Buffer) error { return tiff.Encode(b, bunt, &tiff.Options{Compression: tiff.Deflate}) },
	}

	saat := make([][]byte, 0, len(schreiber))

	for _, schreibe := range schreiber {
		var puffer bytes.Buffer
		if fehler := schreibe(&puffer); fehler != nil {
			f.Fatalf("ein Saatbild schreiben: %v", fehler)
		}

		saat = append(saat, puffer.Bytes())
	}

	return saat
}

// pngGeheilt setzt die Prüfsumme jedes PNG-Stücks neu — was kein PNG ist,
// bleibt, wie es ist.
func pngGeheilt(roh []byte) []byte {
	const signatur = "\x89PNG\r\n\x1a\n"

	if !bytes.HasPrefix(roh, []byte(signatur)) {
		return roh
	}

	geheilt := bytes.Clone(roh)

	for stelle := len(signatur); stelle+12 <= len(geheilt); {
		laenge := int(binary.BigEndian.Uint32(geheilt[stelle:]))
		if laenge < 0 || laenge > len(geheilt)-stelle-12 {
			break
		}

		ende := stelle + 8 + laenge
		binary.BigEndian.PutUint32(geheilt[ende:], crc32.ChecksumIEEE(geheilt[stelle+4:ende]))
		stelle = ende + 4
	}

	return geheilt
}
