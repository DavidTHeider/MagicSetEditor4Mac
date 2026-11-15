package service

import (
	"context"
	"math"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// istBildreihe sagt, ob ein Feld eine Mehrfachwahl ist, die ihre gewählten
// Wahlen als **Reihe ihrer Bilder** zeichnet — `image list` und der Bildteil
// von `both list` (`RENDER_LIST | RENDER_IMAGE`).
//
// **Nicht gebaut**: der Text einer `text list` oder `both list` und die
// `checklist` mit ihren Kästchen — kein Stil des Packs zeichnet sie auf eine
// Karte; `checklist` steht nur in Einstellflächen (`docs/PARITAET.md`,
// `unsupported`, #2297). Eine solche Mehrfachwahl zeichnet wie eine einfache
// Wahl.
func istBildreihe(feld *msepaket.Feldstil, vorlage Vorlage) bool {
	darstellung := strings.ToLower(feld.Darstellung)

	return strings.Contains(darstellung, "list") && !strings.Contains(darstellung, "checklist") &&
		(strings.Contains(darstellung, "image") || strings.Contains(darstellung, "both")) &&
		feldart(vorlage, feld.Feld) == domain.KindMultipleChoice
}

// bildreihe legt die Bilder der gewählten Wahlen nebeneinander — je eines in
// **eigener Größe**:
//
//	RealPoint pos = align_in_rect(style().alignment, RealSize(0,0), dc.getInternalRect());
//	…
//	FOR_EACH(choice, selected) { drawChoice(dc, pos, choice); }
//	…
//	Image image = it->second.generate(options);           // Options(0,0,…)
//	dc.DrawPreRotatedImage(image, RealRect(pos.x + size.width, pos.y, options.width, options.height), …);
//	…
//	pos = move_in_direction(style().direction, pos, size, style().spacing);
//
// `MultipleChoiceValueViewer::draw` und `drawChoice`,
// `src/render/value/multiple_choice.cpp`, Zeilen 23–83, Fassung f9d9356 vom
// 2021-07-06. **Drei Eigenheiten, die hier mitkommen**: `pos` beginnt mit der
// Größe null ausgerichtet — bei `middle` beginnt das Bild auf halber Höhe —,
// die Reihenfolge ist die der Erklärung (`value().get`, nach `normalForm`),
// und eine Wahl ohne Bild rückt um den Abstand allein weiter.
//
// **Die Größe kommt aus der Bildquelle**, und deshalb nur für ein Wahlbild,
// das eine Datei ist; ein gerechnetes wird benannt und übersprungen — seine
// Größe stünde erst nach dem Rastern fest.
func bildreihe(
	ctx context.Context, feld *msepaket.Feldstil, wert string, rahmen render.Rechteck,
	vorlage Vorlage, umgebung *msescript.Umgebung, symbole Symbolquelle, bilder render.Bildquelle,
) ([]render.Ebene, []domain.Befund) {
	abstand, befunde := masszahl(ctx, feld.Feld, "spacing", feld.Abstand, 0, umgebung)

	// **Ausgerechnet wie jede Stilangabe, und gedreht wie das Feld** —
	// Befund 4 des Zweitblicks über E/F (`MultipleChoiceStyle::update`,
	// `src/data/field/multiple_choice.cpp`, Zeilen 42–47; `Rotation::tr`
	// über `ValueViewer::getRotation`, `src/render/value/viewer.cpp`, Zeilen
	// 44–46; Fassung f9d9356 vom 2021-07-06).
	richtung := gerechneteAngabe(ctx, feld.Richtung, umgebung)
	winkel := winkelAus(ctx, feld, umgebung)
	x, y := reihenanfang(rahmen, ausrichtungstext(ctx, feld, umgebung))

	var ebenen []render.Ebene

	for _, wahl := range gewaehlteWahlen(vorlage, feld.Feld, wert) {
		breite, hoehe := 0.0, 0.0

		baum, weitere := wahlbild(ctx, feld, wahl, umgebung)
		befunde = append(befunde, weitere...)

		if baum.Knotenart != "" {
			zeichen := alsZeichenbaum(baum, vorlage.Stil.Paket, symbole)

			links, oben := imFeldGedreht(rahmen, x, y, winkel)

			if ebene, gemessen := wahlbildebene(feld.Feld, wahl, zeichen, links, oben, bilder); gemessen != nil {
				befunde = append(befunde, *gemessen)
			} else {
				ebene.Winkel = winkel
				ebenen = append(ebenen, ebene)
				breite, hoehe = ebene.Rahmen.Breite, ebene.Rahmen.Hoehe
			}
		}

		x, y = inRichtung(richtung, x, y, breite, hoehe, abstand)
	}

	return ebenen, befunde
}

// wahlbildebene baut die Ebene eines Wahlbilds in seiner eigenen Größe —
// oder sagt, warum das nicht geht.
func wahlbildebene(
	feld, wahl string, zeichen render.Bildbaum, x, y float64, bilder render.Bildquelle,
) (render.Ebene, *domain.Befund) {
	ort := feld + " → choice images → " + wahl

	if !zeichen.IstBlatt() {
		return render.Ebene{}, &domain.Befund{
			Ort:     ort,
			Meldung: "ein gerechnetes Wahlbild in einer Bildreihe ist hier nicht gebaut — seine Größe stünde erst nach dem Rastern fest (#2297)",
		}
	}

	if bilder == nil {
		return render.Ebene{}, &domain.Befund{Ort: ort, Meldung: "für die Bildreihe ist keine Bildquelle eingerichtet"}
	}

	bild, err := bilder.Bild(zeichen.Pfad)
	if err != nil {
		return render.Ebene{}, &domain.Befund{Ort: ort, Meldung: "das Bild " + zeichen.Pfad + " ist nicht lesbar: " + err.Error()}
	}

	groesse := bild.Bounds().Size()

	return render.Ebene{
		Art:      render.ArtBild,
		Feld:     feld,
		Rahmen:   render.Rechteck{Links: x, Oben: y, Breite: float64(groesse.X), Hoehe: float64(groesse.Y)},
		Bild:     zeichen,
		Mischart: zeichen.Mischart(),
		Deckung:  1,
	}, nil
}

// gewaehlteWahlen gibt die gewählten Wahlen in der Reihenfolge der Erklärung.
func gewaehlteWahlen(vorlage Vorlage, feld, wert string) []string {
	var wahlen []string

	for _, felder := range [][]domain.Field{vorlage.Kartenfelder, vorlage.Stil.Zusatzfelder} {
		for i := range felder {
			if felder[i].Name == feld {
				wahlen = felder[i].Choices
			}
		}
	}

	geordnet := normalform(wahlen, wert, "")
	if geordnet == "" {
		return nil
	}

	return strings.Split(geordnet, Wahltrenner)
}

// reihenanfang ist `align_in_rect(alignment, RealSize(0,0), rect)` —
// `align_delta_x` und `align_delta_y`, `src/util/alignment.cpp`, Zeilen
// 15–33, Fassung f9d9356 vom 2021-07-06. Ohne Angabe gilt `ALIGN_STRETCH`,
// das weder waagerecht noch senkrecht etwas verschiebt.
func reihenanfang(rahmen render.Rechteck, ausrichtung string) (x, y float64) {
	woerter := strings.Fields(strings.ToLower(ausrichtung))
	x, y = rahmen.Links, rahmen.Oben

	for _, wort := range woerter {
		switch wort {
		case "center":
			x = rahmen.Links + rahmen.Breite/2
		case "right":
			x = rahmen.Links + rahmen.Breite
		case "middle":
			y = rahmen.Oben + rahmen.Hoehe/2
		case "bottom":
			y = rahmen.Oben + rahmen.Hoehe
		}
	}

	return x, y
}

// inRichtung ist `move_in_direction`, `src/util/alignment.cpp`, Zeilen
// 99–124, Fassung f9d9356 vom 2021-07-06 — Vorgabe `left to right`.
func inRichtung(richtung string, x, y, breite, hoehe, abstand float64) (neuX, neuY float64) {
	vor, runter := breite+abstand, hoehe+abstand

	switch strings.ToLower(strings.TrimSpace(richtung)) {
	case "right to left":
		return x - vor, y
	case "top to bottom", "vertical":
		return x, y + runter
	case "bottom to top":
		return x, y - runter
	case "top-right to bottom-left":
		return x - vor, y + runter
	case "bottom-left to top-right":
		return x + vor, y - runter
	case "top-left to bottom-right":
		return x + vor, y + runter
	case "bottom-right to top-left":
		return x - vor, y - runter
	default:
		return x + vor, y
	}
}

// gerechneteAngabe wertet eine Stilangabe aus, die ein Skript sein darf — wie
// [ausrichtungstext], ohne dessen Skriptblock.
func gerechneteAngabe(ctx context.Context, roh string, umgebung *msescript.Umgebung) string {
	rumpf, geklammert := skriptrumpf(strings.TrimSpace(roh))
	if !geklammert {
		return strings.TrimSpace(roh)
	}

	return ausgerechnet(ctx, rumpf, umgebung)
}

// imFeldGedreht bringt eine Stelle im Feld an ihren Ort auf der Karte: um den
// Ursprung des Feldkastens gedreht, gegen den Uhrzeigersinn —
// `(c·x + s·y + links, −s·x + c·y + oben)`, wie `Rotation::tr` rechnet.
//
// **Gerundet auf ein Milliardstel**: der Kosinus von 90 Grad ist im
// Gleitkomma nicht null, und eine Stelle von 41,999999 wäre eine andere.
func imFeldGedreht(rahmen render.Rechteck, x, y, winkel float64) (links, oben float64) {
	if winkel == 0 {
		return x, y
	}

	bogen := winkel * math.Pi / 180
	c, s := math.Cos(bogen), math.Sin(bogen)
	dx, dy := x-rahmen.Links, y-rahmen.Oben

	return milliardstel(rahmen.Links + c*dx + s*dy), milliardstel(rahmen.Oben - s*dx + c*dy)
}

func milliardstel(wert float64) float64 {
	return math.Round(wert*1e9) / 1e9
}
