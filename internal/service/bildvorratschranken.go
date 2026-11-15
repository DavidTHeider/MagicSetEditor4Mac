package service

// **Was durch diesen Vorrat passt** — Werkstück A12 von Phase 5.3
// (#1451).
//
// # Warum die Schranken hier stehen und nicht beim Anbieter
//
// **Weil zwei verschiedene Fragen beantwortet werden.** Was ein Dienst
// annimmt, sagt, was er verarbeitet; was hier durchkommt, muss durch die
// Ablage passen und in einer Karte landen
// ([0102](../../docs/entscheidungen/0102-wo-ein-erzeugtes-bild-liegt.md) §6).
//
// **Und sie greifen in einer Reihenfolge.** Eine komprimierte Antwort von
// 50 MB ist eine Bildbombe, gleich ob der Dienst sie annimmt — und eine
// 2 KB große Datei mit 40.000 × 40.000 Pixeln ist auch eine: sie wiegt erst
// beim Auspacken, und dann wiegt sie Gigabytes.

import (
	"bytes"
	"fmt"
	"image"

	// **Die zwei Decoder stehen hier und nicht nur nebenan.**
	//
	// `image.DecodeConfig` kennt nur, was jemand registriert hat. Zwei
	// andere Dateien dieses Pakets tun das nebenbei (`bildausfuhr.go`,
	// `nutzerbild.go`) — wer sie umbaut, nähme dieser hier die Maße weg,
	// und **der Lauf bliebe grün**: `bildvorrat_test.go` bindet `image/jpeg`
	// selbst ein, um sein Prüfbild zu schreiben. Ein Test kann diesen
	// Unterschied nicht sehen; die Zeile kann ihn verhindern.
	_ "image/jpeg"
	_ "image/png"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Bildgrenzen sind die Schranken eines Vorrats.
//
// **Sie stehen hier und nicht beim Anbieter**
// ([0102](../../docs/entscheidungen/0102-wo-ein-erzeugtes-bild-liegt.md) §6):
// was ein Dienst annimmt, sagt nichts darüber, was durch diese Anwendung
// passt.
type Bildgrenzen struct {
	// JeBild ist die Bytezahl, ab der ein einzelnes Bild abgewiesen wird.
	JeBild int64

	// Insgesamt ist die Quote über alle wartenden Bilder.
	Insgesamt int64

	// Pixel ist die Zahl der Bildpunkte, ab der abgewiesen wird — **vor** dem
	// Auspacken gemessen, sonst ist der Speicher schon belegt.
	Pixel int64

	// Frist ist, wie lange ein Vorschlag gilt.
	Frist time.Duration
}

// UeblicheBildgrenzen gibt die Grenzen, mit denen die Anwendung läuft.
//
// **Acht Mebibyte je Bild**: ein PNG in 1024 × 1536 wiegt selten mehr als
// zwei, und wer acht überschreitet, schickt etwas anderes als ein Kartenbild.
//
// **8.294.400 Pixel**: genau die Obergrenze, die OpenAI selbst dokumentiert
// (0102 §6). Was darüber liegt, ist Drift und keine große Karte.
//
// **Vierundsechzig Mebibyte insgesamt** — etwa zehn wartende Vorschläge. Mehr
// gleichzeitig offen zu haben, heißt, den Überblick schon verloren zu haben.
//
// **Eine Stunde Frist**: lang genug, um einen Vorschlag in Ruhe anzusehen,
// kurz genug, dass ein vergessenes Fenster nicht den halben Tag Platz belegt.
func UeblicheBildgrenzen() Bildgrenzen {
	return Bildgrenzen{
		JeBild:    8 << 20,
		Insgesamt: 64 << 20,
		Pixel:     8_294_400,
		Frist:     time.Hour,
	}
}

// Bildbesitz sagt, zu wem ein wartendes Bild gehört.
//
// **Profil und Auftrag zusammen** — das Issue verlangt „Profil-/
// Vorschlagsbindung", und beides ist nötig: das Profil, weil ein zweiter

// bildendungen sind die Medientypen, die dieser Vorrat annimmt.
//
// **PNG und JPEG, und sonst nichts.** Das Programm fragt nach PNG (0102 §6);
// JPEG steht daneben, weil ein Anbieter es liefern darf, ohne dass jemand es
// verlangt hat. Was weder das eine noch das andere ist — SVG zumal, mit
// seinen aktiven Inhalten — kommt nicht herein.
var bildendungen = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
}

// bildsignaturen ordnet den ersten Bytes einen Medientyp zu.
//
// **Gelesen wird die Datei, nicht ihre Aufschrift.** Ein `Content-Type` und
// eine Endung sagt, wer liefert; die Signatur sagt, was dasteht.
var bildsignaturen = []struct {
	vorspann []byte
	typ      string
}{
	{[]byte("\x89PNG\r\n\x1a\n"), "image/png"},
	{[]byte("\xff\xd8\xff"), "image/jpeg"},
}

// pruefe hält die drei Schranken in der Reihenfolge, in der sie greifen
// müssen: Bytes, Signatur, Pixel.
func (v *Bildvorrat) pruefe(inhalt []byte) (medientyp string, breite, hoehe int, fehler error) {
	if int64(len(inhalt)) > v.grenzen.JeBild {
		return "", 0, 0, fmt.Errorf(
			"%w: das Bild wiegt %d Bytes, erlaubt sind %d",
			domain.ErrValidation, len(inhalt), v.grenzen.JeBild)
	}

	medientyp = bildtypAus(inhalt)
	if medientyp == "" {
		return "", 0, 0, fmt.Errorf(
			"%w: die Antwort trägt keine Bildsignatur — erwartet werden PNG "+
				"oder JPEG", domain.ErrValidation)
	}

	// **Nur der Kopf.** `DecodeConfig` liest die Maße, ohne die Bildpunkte
	// zu belegen — das ist der ganze Zweck dieser Zeile.
	kopf, _, fehler := image.DecodeConfig(bytes.NewReader(inhalt))
	if fehler != nil {
		return "", 0, 0, domain.MitGrund(domain.ErrValidation, fehler,
			"den Bildkopf lesen")
	}

	if int64(kopf.Width)*int64(kopf.Height) > v.grenzen.Pixel {
		return "", 0, 0, fmt.Errorf(
			"%w: das Bild behauptet %d × %d Pixel, erlaubt sind %d Pixel",
			domain.ErrValidation, kopf.Width, kopf.Height, v.grenzen.Pixel)
	}

	return medientyp, kopf.Width, kopf.Height, nil
}

// bildtypAus gibt den Medientyp der Signatur — leer, wenn keine passt.
func bildtypAus(inhalt []byte) string {
	for _, signatur := range bildsignaturen {
		if bytes.HasPrefix(inhalt, signatur.vorspann) {
			return signatur.typ
		}
	}

	return ""
}
