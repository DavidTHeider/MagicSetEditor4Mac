// **Zwei Befunde aus dem Nachlauf zum Zweitblick über Phase P** —
// #860.
package svg_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/svg"
)

// **Eine Maske wird gedehnt, nicht eingepasst.**
//
// Der Rasterer tastet das Maskenbild über den **ganzen** Rahmen ab und streckt
// es damit nötigenfalls nichtproportional (`maskenbild.anteil` über
// `abtasten`, `internal/render/bild.go`). Ohne `preserveAspectRatio="none"`
// gilt im SVG der Standard `xMidYMid meet`: ein 2:1-Bild in einem 1:1-Rahmen
// wird eingepasst, und der Rest der Maske bleibt **schwarz** — also die halbe
// Ebene unsichtbar.
func TestEineMaskeWirdGedehntUndNichtEingepasst(t *testing.T) {
	t.Parallel()

	mitMaske := flaeche()
	mitMaske.Maske = render.Blatt("/probe.mse-style/maske.png")

	geschrieben, _, fehler := svg.AusPlan(t.Context(), eineEbene(mitMaske),
		svg.Quellen{Baeume: baumprobe{}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	inMaske := regexp.MustCompile(`<mask id="[^"]*"><image[^>]*preserveAspectRatio="none"`)
	if !inMaske.MatchString(geschrieben) {
		t.Errorf("die Maske wird eingepasst statt gedehnt:\n%s", geschrieben)
	}
}

// **Ein Rand mit runden Ecken ist in den Ecken beschnitten.**
//
//	if (image.Point{X: x, Y: y}).In(innen) { return false }
//	return inRundemRechteck(rahmen, ebene.Ecke*vergroesserung, x, y)
//
// `inFlaechenform`, `internal/render/farbflaeche.go` — der Punkt muss nach dem
// Abzug des Innenrechtecks **weiterhin** im gerundeten Außenrechteck liegen.
// MSE2 malt ebenso über ein gerundetes Außenrechteck
// (`src/render/value/color.cpp`, Zeilen 47–50, Fassung f9d9356 vom
// 2021-07-06).
//
// Vier gewöhnliche Rechtecke haben diesen Beschnitt nicht: bei 100 × 100,
// Radius 10 und Randbreite 2 liegt (0,0) in der oberen Bahn und außerhalb des
// Bogens — √((0−10)² + (0−10)²) = 14,14 > 10.
func TestEinRandMitRundenEckenIstBeschnitten(t *testing.T) {
	t.Parallel()

	gerahmt := flaeche()
	gerahmt.Ecke = 10
	gerahmt.Rand = &render.Randbreiten{Links: 2, Oben: 2, Rechts: 2, Unten: 2}

	geschrieben, _, fehler := svg.AusPlan(t.Context(), eineEbene(gerahmt), svg.Quellen{})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	for _, teil := range []string{"<clipPath id=", `clip-path="url(#`} {
		if !strings.Contains(geschrieben, teil) {
			t.Errorf("%q fehlt — der Rand ist nicht beschnitten:\n%s", teil, geschrieben)
		}
	}

	// **Der Beschnitt trägt den Radius**, sonst schnitte er ein Rechteck aus
	// einem Rechteck und täte nichts.
	beschnitt := regexp.MustCompile(`<clipPath id="[^"]*"><rect[^>]*rx="10"`)
	if !beschnitt.MatchString(geschrieben) {
		t.Errorf("der Beschnitt ist nicht gerundet:\n%s", geschrieben)
	}
}

// **Ohne runde Ecken bleibt es bei vier Bahnen.**
//
// Ein Beschnitt ohne Rundung schnitte ein Rechteck aus einem Rechteck — ein
// Knoten, der nichts tut, und eine Zusage, die niemand prüfen kann.
func TestEinEckigerRandBrauchtKeinenBeschnitt(t *testing.T) {
	t.Parallel()

	gerahmt := flaeche()
	gerahmt.Rand = &render.Randbreiten{Links: 2, Oben: 2, Rechts: 2, Unten: 2}

	geschrieben, _, fehler := svg.AusPlan(t.Context(), eineEbene(gerahmt), svg.Quellen{})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if strings.Contains(geschrieben, "<clipPath") {
		t.Errorf("ein eckiger Rand wurde beschnitten:\n%s", geschrieben)
	}
}
