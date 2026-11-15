package render

import (
	"image"
	"math"
)

// **Text vierfach zeichnen und herunterrechnen** — MSEs
// `draw_resampled_text`, portiert für QA-03.AK-02
// (#1656).
//
// # Was MSE tut
//
// Beim Bildexport steht die Güte auf `QUALITY_AA` — `card_anti_alias` ist
// voreingestellt an (`src/data/settings.cpp`, Zeile 126;
// `DataViewer::draw`, `src/render/card/viewer.cpp`, Zeilen 29–33) —, und
// dann zeichnet `RotatedDC::DrawText` kein Stück Text unmittelbar
// (`src/util/rotation.cpp`, Zeilen 189–219):
//
//  1. Das Stück wird in **vierfacher** Größe in einen eigenen Puffer gesetzt,
//     weiß auf schwarz, `int(breite) + 3` mal `int(hoehe) + 1` Bildpunkte
//     groß. Seine Stelle darin ist der Bruchteil der Zielstelle auf
//     **Viertelpunkte**: `xsub = int(text_scaling * (pos.x - xi))`.
//  2. Der Puffer wird auf die einfache Größe heruntergerechnet
//     (`downsample_to_alpha`) — ein Kastenfilter, gestaucht mit einer
//     eigenen Rechnung, die dabei fetter macht.
//  3. Das Ergebnis wird in der Textfarbe an die ganzzahlige Ecke `(xi, yi)`
//     gelegt.
//
// `src/gfx/resample_text.cpp`, Zeilen 166–211, Fassung f9d9356 vom
// 2021-07-06.
//
// # Was bleibt
//
// Gemessen am 2026-09-23 an einer Regeltextzeile gegen Core Text auf MSEs
// Weg: 1,92 im mittleren Betragsabstand. Der Rest war Quartz'
// Schriftglättung; sie gehört nicht MSE, sondern der Plattform, und sie ist
// seit #1806 in der Maske nachgebildet ([quartzVerdickung]) — 0,18.
//
// # Was hier anders ist
//
// **Die Glyphen setzt dieser Rasterer, nicht Core Text** — jede an die
// Viertelstelle ihres Vorschubs, abgeschnitten wie `xsub`. Core Text setzt
// innerhalb eines Stücks feiner; an der Regeltextzeile lag der Unterschied
// bei 0,065 von 255.

// vierfachlauf ist ein Textstück, wie MSE es an `draw_resampled_text` gibt.
type vierfachlauf struct {
	// xi, yi ist die ganzzahlige Ecke im Zielbild.
	xi, yi int

	// xsub, ysub ist die Stelle im Puffer, in Viertelpunkten.
	xsub, ysub int

	// breite und hoehe sind die Maße des Puffers in einfacher Größe —
	// ungestaucht, samt MSEs Zugabe.
	breite, hoehe int

	// grundlinie ist die Grundlinie im Puffer, in Viertelpunkten.
	grundlinie int

	// ausgabebreite ist die Breite nach der Stauchung.
	ausgabebreite int

	// deckung ist der Puffer, `breite·4` mal `hoehe·4`.
	deckung []uint8
}

// neuerVierfachlauf legt den Puffer für ein Stück an.
//
// `links` und `oben` sind die Ecke des Stücks im Zielbild, `breite` und
// `hoehe` sein Textmaß ungestaucht, `grundlinie` der Abstand der Grundlinie
// von oben — alles in Bildpunkten.
func neuerVierfachlauf(links, oben, breite, hoehe, grundlinie, stauchung float64) *vierfachlauf {
	return vierfachlaufMitRand(links, oben, breite, hoehe, grundlinie, stauchung, 0)
}

// vierfachlaufMitRand ist [neuerVierfachlauf] mit Platz für einen
// Weichzeichner — `weich` Bildpunkte je Seite, wie in `draw_resampled_text`:
//
//	int w = static_cast<int>(rect.width) + 3 + 2 * blur_radius, h = static_cast<int>(rect.height) + 1 + 2 * blur_radius;
//	int xi = static_cast<int>(rect.x) - blur_radius / text_scaling,
//	    yi = static_cast<int>(rect.y) - blur_radius / text_scaling;
//
// `src/gfx/resample_text.cpp`, Zeilen 170–174, Fassung f9d9356 vom
// 2021-07-06. **Die Ecke rückt um ein Viertel der Weichheit**, ganzzahlig
// geteilt — bis drei Bildpunkte gar nicht, und der Rand liegt dann ganz
// rechts und unten.
func vierfachlaufMitRand(links, oben, breite, hoehe, grundlinie, stauchung float64, weich int) *vierfachlauf {
	// `int w = static_cast<int>(rect.width) + 3, h = static_cast<int>(rect.height) + 1;`
	// — „some fonts are larger then the GetTextExtent tells us".
	//
	// **Und senkrecht mit Rand: eine halbe Zeile darüber und darunter** —
	// eine bewusste Abweichung (adapted), seit dem 2026-09-24. MSEs Puffer ist
	// so hoch wie das Textmaß der Plattform, und was darüber liegt, fällt
	// weg. Unter Windows schließt dieses Maß die Akzente ein (GDIs Oberlänge
	// ist `winAscent`); auf dem Mac nennt Core Text für Helvetica 0,77 em,
	// und die Punkte eines Ä reichen bis 0,89 em — gemessen am selben Tag.
	// Auf dem Mac stand damit „AOU", wo „ÄÖÜ" steht (`zeichenrand_test.go`).
	// **Gerechnet wird wie vorher**: Grundlinie, Zeilen und Umbruch bleiben,
	// wo sie waren; nur fällt nicht mehr weg, was über die Zeile ragt.
	rand := int(math.Ceil(hoehe / 2))
	lauf := &vierfachlauf{
		xi:     int(links) - weich/Textraster,
		yi:     int(oben) - rand - weich/Textraster,
		breite: int(breite) + 3 + 2*weich,
		hoehe:  int(hoehe) + 1 + 2*rand + 2*weich,
	}

	lauf.xsub = int(Textraster * (links - float64(lauf.xi)))
	lauf.ysub = int(Textraster * (oben - float64(int(oben)-weich/Textraster)))
	lauf.grundlinie = lauf.ysub + rand*Textraster + int(math.Round(Textraster*grundlinie))

	// `w += int(w * (stretch - 1) * ca);` — ungedreht ist `ca` eins.
	lauf.ausgabebreite = lauf.breite + int(float64(lauf.breite)*(stauchung-1))
	lauf.deckung = make([]uint8, lauf.breite*Textraster*lauf.hoehe*Textraster)

	return lauf
}

// setze legt eine vierfach gerasterte Glyphe an ihre Stelle. `versatz` ist
// ihr Abstand vom Anfang des Stücks in Bildpunkten, ungestaucht.
//
// **Was aus dem Puffer ragt, fällt weg** — waagerecht wie bei MSE, dessen
// Puffer genau so breit ist: ein kursiver Schwanz, der über die drei Punkte
// Zugabe reicht, ist dort abgeschnitten. Senkrecht hat der Puffer seit dem
// 2026-09-24 Rand ([neuerVierfachlauf]).
func (l *vierfachlauf) setze(maske image.Image, ecke image.Point, versatz float64) {
	spalten := l.breite * Textraster
	zeilen := l.hoehe * Textraster
	feld := maske.Bounds()

	links := l.xsub + int(Textraster*versatz) + ecke.X
	oben := l.grundlinie + ecke.Y

	for y := range feld.Dy() {
		zeile := oben + y
		if zeile < 0 || zeile >= zeilen {
			continue
		}

		for x := range feld.Dx() {
			spalte := links + x
			if spalte < 0 || spalte >= spalten {
				continue
			}

			m := alphaAn(maske, feld.Min.X+x, feld.Min.Y+y).A
			alt := l.deckung[zeile*spalten+spalte]

			// Weiß über das, was schon dasteht: `Over` in einem Kanal.
			l.deckung[zeile*spalten+spalte] = alt + uint8((int(255-alt)*int(m)+127)/255) /* #nosec G115 -- höchstens 255 */
		}
	}
}

// lege rechnet den Puffer herunter und legt ihn in der Farbe der Ebene auf.
func (r *Rasterer) lege(
	ziel *image.NRGBA, ebene Ebene, maske *maskenbild, rahmen image.Rectangle, lauf *vierfachlauf,
) {
	r.legeMit(ziel, ebene, ebene.Farbe, maske, rahmen, lauf, nil)
}

// legeMit ist [Rasterer.lege] mit eigener Farbe und einem Schritt auf der
// heruntergerechneten Deckung — den braucht der Schatten
// ([Rasterer.legeSchatten]).
func (r *Rasterer) legeMit(
	ziel *image.NRGBA, ebene Ebene, farbe Farbe, maske *maskenbild, rahmen image.Rectangle,
	lauf *vierfachlauf, schritt func(deckung []uint8),
) {
	// **Auf null gestaucht bleibt nichts zu zeichnen.** MSE teilte hier durch
	// die Ausgabebreite null; dieser Rasterer lässt das Stück aus.
	if lauf.ausgabebreite <= 0 {
		return
	}

	deckung := herunterrechnen(lauf.deckung, lauf.breite, lauf.hoehe, lauf.ausgabebreite)
	if schritt != nil {
		schritt(deckung)
	}

	for y := range lauf.hoehe {
		for x := range lauf.ausgabebreite {
			wert := deckung[y*lauf.ausgabebreite+x]
			if wert == 0 {
				continue
			}

			zielX, zielY := lauf.xi+x, lauf.yi+y
			if !(image.Point{X: zielX, Y: zielY}).In(ziel.Bounds()) {
				continue
			}

			r.punkt(ziel, zielX, zielY, farbe,
				ebene.Deckung*float64(wert)/255*maske.anteil(rahmen, zielX, zielY), nil)
		}
	}
}

// herunterrechnen ist `downsample_to_alpha`: erst waagerecht, dann senkrecht,
// jedes Mal ganzzahlig geteilt.
//
// `vierfach` ist `breite·4` mal `hoehe·4` groß. Ist `ausgabebreite` das
// Vierfache der Eingabe geteilt durch vier, wird je vier Punkte gemittelt;
// sonst läuft MSEs eigene Stauchung (Zeilen 51–84), die dabei fetter macht:
// `mul = 128 + min(256, 128·w1 / (4·w2))`.
func herunterrechnen(vierfach []uint8, breite, hoehe, ausgabebreite int) []uint8 {
	w1, h1 := breite*Textraster, hoehe*Textraster

	var waagerecht []uint8

	if w1 == ausgabebreite*Textraster {
		waagerecht = make([]uint8, ausgabebreite*h1)

		for i := range waagerecht {
			summe := 0
			for j := range Textraster {
				summe += int(vierfach[i*Textraster+j])
			}

			waagerecht[i] = uint8(summe / Textraster) //nolint:gosec // Mittel von Bytes
		}
	} else {
		waagerecht = gestaucht(vierfach, w1, h1, ausgabebreite)
	}

	aus := make([]uint8, ausgabebreite*hoehe)

	for y := range hoehe {
		for x := range ausgabebreite {
			summe := 0
			for j := range Textraster {
				summe += int(waagerecht[x+ausgabebreite*(j+Textraster*y)])
			}

			aus[x+ausgabebreite*y] = uint8(summe / Textraster) //nolint:gosec // Mittel von Bytes
		}
	}

	return aus
}

// gestaucht ist MSEs waagerechte Stauchung beim Herunterrechnen — die Zeilen
// 51–84 von `downsample_to_alpha`, mit denselben ganzen Zahlen.
func gestaucht(vierfach []uint8, w1, h1, w2 int) []uint8 {
	// `const int shift = 32-12-8; // => max size = 4096, max alpha = 255`
	const shift = 32 - 12 - 8

	anteil := (w2 << shift) / w1
	rest := (w2 << shift) % w1

	// „make the image 'bolder' to compensate for compressing it"
	mul := 128 + min(256, 128*w1/(Textraster*w2))

	aus := make([]uint8, w2*h1)
	ein := 0

	for y := range h1 {
		offen := anteil + rest

		for x := range w2 {
			noch := 1 << shift
			summe := 0

			for noch >= offen {
				summe += int(vierfach[min(ein, len(vierfach)-1)]) * offen
				noch -= offen
				offen = anteil
				ein++
			}

			if noch > 0 {
				summe += int(vierfach[min(ein, len(vierfach)-1)]) * noch
				offen -= noch
			}

			aus[y*w2+x] = uint8(min(255, ((summe>>shift)*mul)>>8)) /* #nosec G115 -- durch min begrenzt */
		}
	}

	return aus
}
