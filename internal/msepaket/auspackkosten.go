package msepaket

// **Was ein Bild ausgepackt kostet** — FMT-06b
// (#1974,
// Vermerk [0160](../../docs/entscheidungen/0160-was-ein-bild-ausgepackt-kostet.md)).
//
// # Warum eine zweite Grenze
//
// **[GrenzeBildpunkte] zählt Bildpunkte, und ein Bildpunkt kostet je nach
// Kodierer zwischen einem und vierundzwanzig Byte.** Unter ihr waren bis zu
// 2,4 GB je Bild erlaubt — ein progressives JPEG in CMYK. Gefunden hat das
// `FuzzDekodiereBild` beim Bau von QA-06b (#1967).
//
// # Woher die Faktoren kommen
//
// **Gemessen** am 2026-10-01 mit Go 1.27.1, je Format und Farbmodell ein
// Bild von 512 × 512, die zugeteilten Byte über `runtime.MemStats`
// (`auspackkosten_test.go` hält es fest):
//
//	PNG  grau 1,2 · grau16 2,2 · RGBA 4,2 · RGBA64 8,2 · Palette 1,3
//	PNG mit Zeilensprung (Adam7)  grau 2,2 · RGBA 8,3 · RGBA64 16,3
//	JPEG grau 1,1 · YCbCr 4:2:0 1,6 · CMYK 8,1
//	JPEG progressiv  grau 5,1 · YCbCr 4:4:4 15,1 · CMYK 24,1
//	GIF  2,1
//
// **Der Kopf sagt nicht alles**: ob ein PNG Zeilensprung trägt und ob ein
// JPEG progressiv ist, steht nicht in [image.Config]. Geschätzt wird deshalb
// der teurere Fall, aufgerundet: ein Byte mehr als gemessen.

import (
	"errors"
	"image"
	"image/color"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// GrenzeAuspacken ist, wie viel ein Bild ausgepackt höchstens kosten darf.
//
// **Ein Gibibyte.** Das größte Bild im vollen Pack ist eine Karte bei
// 1200 dpi mit 12,4 Millionen Bildpunkten (`paketgrenzen.go`); selbst als
// progressives JPEG in CMYK kostet sie unter 320 MB. Ein JPEG in YCbCr darf
// damit rund 67 Millionen Bildpunkte haben; ein PNG mit RGBA hält schon die
// Bildpunktgrenze darunter — hundert Millionen kosten 900 MB.
const GrenzeAuspacken = 1 << 30

// Auspackkosten schätzt nach oben, wie viele Byte die registrierten Kodierer
// für ein Bild mit diesem Kopf zuteilen.
//
// **Ein JPEG legt Blöcke zu sechzehn Bildpunkten an** — gerechnet wird über
// die aufgerundete Fläche, dieselbe wie in `FuzzDekodiereBild`.
func Auspackkosten(mass image.Config, art string) int64 {
	breite, hoehe := int64(mass.Width), int64(mass.Height)
	if art == "jpeg" {
		breite, hoehe = (breite+15)/16*16, (hoehe+15)/16*16
	}

	return breite * hoehe * jeBildpunkt(mass.ColorModel, art)
}

// jeBildpunkt ist der Faktor je Format und Farbmodell — **der teuerste, wo
// keiner bekannt ist**.
func jeBildpunkt(modell color.Model, art string) int64 {
	switch art {
	case "png":
		switch modell {
		case color.GrayModel:
			return 3
		case color.Gray16Model:
			return 5
		case color.RGBAModel, color.NRGBAModel:
			return 9
		case color.RGBA64Model, color.NRGBA64Model:
			return 17
		}

		return 3 // die Palette — ein Byte je Bildpunkt wie Grau
	case "gif":
		return 3
	case "jpeg":
		switch modell {
		case color.GrayModel:
			return 6
		case color.YCbCrModel:
			return 16
		}
	}

	return 25
}

// bildZuGross ist die Abweisung — mit eigenem Schlüssel an jeder Fläche, und
// [ErrBildZuGross] bleibt darin erkennbar.
func bildZuGross(was string, werte ...any) error {
	return domain.NewFieldError("bild", domain.SchluesselBildZuGross,
		"%s: "+was, append([]any{ErrBildZuGross}, werte...)...).MitGrund(ErrBildZuGross)
}

// ausgepacktZuTeuer sagt, ob ein Bild über [GrenzeAuspacken] läge.
func ausgepacktZuTeuer(mass image.Config, art string) error {
	if kosten := Auspackkosten(mass, art); kosten > GrenzeAuspacken {
		return bildZuGross("%d × %d Bildpunkte kosteten als %s ausgepackt bis zu %d Byte, höchstens %d",
			mass.Width, mass.Height, art, kosten, GrenzeAuspacken)
	}

	return nil
}

// ZuGrossOder gibt die Abweisung eines zu großen Bildes, wie sie ist — mit
// ihrem Schlüssel —, und jeden anderen Fehler so, wie der Rufer ihn
// umschreibt.
//
// **Weil jeder Rufer bisher umschrieb**: „kein lesbares Bild" stand auch über
// einem Bild, das in Ordnung und nur zu groß war (#1974).
func ZuGrossOder(fehler, sonst error) error {
	if errors.Is(fehler, ErrBildZuGross) {
		return fehler
	}

	return sonst
}
