package svg

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// bild bettet den Bildbaum einer Ebene als Rasterbild ein.
//
// **Über [Bildbaumquelle] und nicht über einen Pfad.** Ergibt das
// `image:`-Skript eine Verknüpfung (`recolor_image`, `masked_blend`,
// `symbol_variation`), ist die Wurzel pfadlos, und über `ebene.Bild.Pfad`
// verschwände die Ebene ganz — dieselbe Falle wie an [render.Ebene.Maske]
// (#852).
func (s *schreiber) bild(aus *strings.Builder, nummer int, ebene, innen render.Ebene) {
	if s.quellen.Baeume == nil {
		s.melde(nummer, ebene, "für Bilder ist keine Bildquelle eingerichtet")

		return
	}

	breite, hoehe := s.masse(ebene.Rahmen)

	gelesen, befund := s.quellen.Baeume.BildAusBaum(ebene.Bild, breite, hoehe)

	// **Gezeichnet wird auch, wenn ein Befund mitkommt** — dieselbe Haltung
	// wie im Rasterer (§8): eine Verknüpfung, die wir nicht können, liefert
	// ihr Ausgangsbild, sichtbar falsch statt unsichtbar fehlend.
	if befund != "" {
		s.melde(nummer, ebene, befund)
	}

	if gelesen == nil {
		return
	}

	if meldung := eingebettet(aus, ebene.Feld, innen.Rahmen, gelesen,
		passung(ebene.Passung)); meldung != "" {
		s.melde(nummer, ebene, meldung)
	}
}

// alsRaster zeichnet eine ganze Ebene durch den Rasterer und bettet sie ein.
//
// # Wofür
//
// Für den Symbolzug und für einen Textzug mit Symbolen. Wohin ein `<sym>`
// mitten im Regeltext kommt, weiß erst der Umbruch; die Symbolmaschine des
// Rasterers ein zweites Mal zu bauen hieße, zwei Rechnungen über dieselbe
// Sache zu führen — und die laufen auseinander, sobald eine etwas lernt.
//
// # Wie
//
// **Ein Plan mit einer Ebene, in Kartengröße.** Das Ergebnis ist außerhalb der
// Ebene durchsichtig und wird deshalb an `0,0` gelegt, nicht in den Rahmen:
// wo genau die Symbole sitzen, hat der Rasterer schon entschieden.
//
// **Die Ebene geht ungedreht und voll deckend hinein** — Drehung, Deckung und
// Maske hängen an der Hülle drumherum, und sie ein zweites Mal anzuwenden
// verdoppelte sie. Dieselbe Überlegung wie in `render.ebeneGedreht`.
func (s *schreiber) alsRaster(
	ctx context.Context, aus *strings.Builder, plan *render.Plan,
	nummer int, ebene, innen render.Ebene,
) {
	if s.quellen.Raster == nil {
		s.melde(nummer, ebene, "für Symbole ist kein Rasterer eingerichtet")

		return
	}

	allein := innen
	allein.Winkel = 0
	allein.Deckung = 1
	allein.Maske = render.Bildbaum{}

	// **Parität**: eine SVG ist eine Ausfuhr und wird mit MSE verglichen.
	gezeichnet, befunde, err := s.quellen.Raster.Zeichne(ctx,
		&render.Plan{
			Breite: plan.Breite, Hoehe: plan.Hoehe,
			Ebenen: []render.Ebene{allein},
		},
		s.vergroesserung(), render.GueteMittel)
	if err != nil {
		s.melde(nummer, ebene, "die Ebene lässt sich nicht zeichnen: "+err.Error())

		return
	}

	// **Die Befunde des Teilplans tragen dessen Ebenennummer — null.** Sie
	// wird hier auf die richtige gesetzt; sonst zeigte jeder Befund aus
	// diesem Weg auf die erste Ebene der Karte.
	for _, befund := range befunde {
		s.melde(nummer, ebene, befund.Meldung)
	}

	// **Ebenfalls `none`.** Hier stimmt das Seitenverhältnis zwar ohnehin —
	// gezeichnet wird die ganze Karte —, aber eine Zusage, die nur zufällig
	// hält, ist keine.
	if meldung := eingebettet(aus, ebene.Feld,
		render.Rechteck{Breite: plan.Breite, Hoehe: plan.Hoehe},
		gezeichnet, "none"); meldung != "" {
		s.melde(nummer, ebene, meldung)
	}
}

// maske legt eine `<mask>` an und gibt ihre Kennung; leer, wenn keine da ist.
//
// # Warum ein eigenes Graubild und nicht das Maskenbild selbst
//
// **MSE nimmt den roten Kanal als Deckung**, nicht die Helligkeit:
//
//	// Copy red chanel to alpha
//	for (size_t i = 0 ; i < n ; ++i) { to[i] = from[3*i]; }
//
// `src/gfx/mask_image.cpp`, Zeilen 41–44, Fassung f9d9356 vom 2021-07-06.
// SVG kennt für `<mask>` nur *Leuchtdichte* und *Alpha*. Rot in alle drei
// Kanäle zu kopieren macht die Leuchtdichte gleich dem roten Kanal — damit
// stimmt die SVG-Maske mit `maskenbild.anteil` überein, statt bei einer
// **farbigen** Maske auseinanderzulaufen (rot ergäbe nach Leuchtdichte 54
// statt 255).
func (s *schreiber) maske(nummer int, ebene render.Ebene) string {
	if ebene.Maske.Leer() {
		return ""
	}

	if s.quellen.Baeume == nil {
		s.melde(nummer, ebene, "für die Maske ist keine Bildquelle eingerichtet")

		return ""
	}

	breite, hoehe := s.masse(ebene.Rahmen)

	gelesen, befund := s.quellen.Baeume.BildAusBaum(ebene.Maske, breite, hoehe)
	if befund != "" {
		s.melde(nummer, ebene, "die Maske "+ebene.Maske.Beschreibung()+": "+befund)
	}

	if gelesen == nil {
		return ""
	}

	kennung := s.name("maske")

	// **Der Rahmen der Maske ist der der Ebene**, ungedreht — sie sitzt in
	// derselben Hülle und wird mit ihr gedreht.
	rahmen := ebene.Rahmen
	if ebene.Winkel != 0 {
		rahmen.Links, rahmen.Oben = 0, 0
	}

	s.vorlagen.WriteString(`<mask id="` + kennung + `">`)

	// **`none` und nicht der Standard.** Ohne das Attribut gilt in SVG
	// `xMidYMid meet`: ein Maskenbild mit anderem Seitenverhältnis würde
	// **eingepasst**, und der Rest der Maske bliebe schwarz — also die halbe
	// Ebene unsichtbar. Der Rasterer tastet das Maskenbild über den ganzen
	// Rahmen ab und dehnt es damit (`maskenbild.anteil` über `abtasten`).
	// Angestrichen vom Nachlauf zum Zweitblick über Phase P
	// (#860).
	if meldung := eingebettet(&s.vorlagen, ebene.Feld, rahmen,
		ausRot(gelesen), "none"); meldung != "" {
		s.melde(nummer, ebene, meldung)
	}

	s.vorlagen.WriteString("</mask>")

	return kennung
}

// ausRot macht aus einem Bild eines, dessen Leuchtdichte sein roter Kanal ist.
func ausRot(quelle image.Image) image.Image {
	feld := quelle.Bounds()
	grau := image.NewNRGBA(feld)

	for y := feld.Min.Y; y < feld.Max.Y; y++ {
		for x := feld.Min.X; x < feld.Max.X; x++ {
			rot, _, _, _ := quelle.At(x, y).RGBA()
			wert := uint8(rot >> 8) //#nosec G115 -- ein 16-Bit-Kanal, um acht geschoben

			grau.SetNRGBA(x, y, color.NRGBA{R: wert, G: wert, B: wert, A: 255})
		}
	}

	return grau
}

// masse gibt die Größe, in der ein Bild gebraucht wird — in Bildpunkten.
func (s *schreiber) masse(rahmen render.Rechteck) (breite, hoehe int) {
	v := s.vergroesserung()

	return int(rahmen.Breite*v + 0.5), int(rahmen.Hoehe*v + 0.5)
}

// vergroesserung gibt die Auflösung der eingebetteten Bilder; null heißt 1.
func (s *schreiber) vergroesserung() float64 {
	if s.quellen.Vergroesserung <= 0 {
		return 1
	}

	return s.quellen.Vergroesserung
}

// passung übersetzt [render.Bildpassung] in `preserveAspectRatio`.
//
//	} else if (options.preserve_aspect == ASPECT_FIT) { … }
//
// `conform_image`, `src/gfx/generated_image.cpp`, Zeilen 49–59, Fassung
// f9d9356 vom 2021-07-06; ausgerichtet wird danach mit `align_in_rect`
// (`src/render/value/choice.cpp`, Zeile 83).
//
// **Nil heißt gestreckt** — MSEs `ASPECT_STRETCH`, und in SVG `none`.
func passung(p *render.Bildpassung) string {
	if p == nil {
		return "none"
	}

	return ausrichtungen[p.Ausrichtung]
}

// ausrichtungen ordnet MSEs neun Ausrichtungen die SVG-Schreibweise zu.
//
// **Der Nullwert ist `xMidYMid`** — MSEs `align_in_rect` mittet ohne Angabe.
var ausrichtungen = map[render.Ausrichtung]string{
	render.LinksOben:   "xMinYMin meet",
	render.MitteOben:   "xMidYMin meet",
	render.RechtsOben:  "xMaxYMin meet",
	render.LinksMitte:  "xMinYMid meet",
	render.MitteMitte:  "xMidYMid meet",
	render.RechtsMitte: "xMaxYMid meet",
	render.LinksUnten:  "xMinYMax meet",
	render.MitteUnten:  "xMidYMax meet",
	render.RechtsUnten: "xMaxYMax meet",
}

// eingebettet schreibt ein `<image>` mit dem Bild als Datenadresse.
//
// **Gibt eine Meldung zurück, statt still zu scheitern.** Ein Bild, das die
// Quelle **hergibt** und das sich nicht packen lässt, ist etwas anderes als
// ein fehlendes: dort ist die Stelle leer, hier stimmt etwas nicht. Der erste
// Anlauf kehrte hier kommentarlos zurück — dieselbe Klasse wie der Befund, um
// dessentwillen es #852 gibt.
func eingebettet(
	aus *strings.Builder, feld string, rahmen render.Rechteck,
	bild image.Image, seitenverhaeltnis string,
) string {
	var roh bytes.Buffer
	if err := png.Encode(&roh, bild); err != nil {
		return "das Bild lässt sich nicht einbetten: " + err.Error()
	}

	fmt.Fprintf(aus, `<image data-feld="%s" x="%s" y="%s" width="%s" height="%s"`,
		entschaerfe(feld), zahl(rahmen.Links), zahl(rahmen.Oben),
		zahl(rahmen.Breite), zahl(rahmen.Hoehe))

	if seitenverhaeltnis != "" {
		fmt.Fprintf(aus, ` preserveAspectRatio="%s"`, seitenverhaeltnis)
	}

	fmt.Fprintf(aus, ` href="data:image/png;base64,%s"/>`,
		base64.StdEncoding.EncodeToString(roh.Bytes()))

	return ""
}
