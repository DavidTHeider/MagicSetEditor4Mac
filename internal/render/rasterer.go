package render

import (
	"context"
	"errors"
	"image"
	"math"
)

// Bildquelle liefert die Bilddaten zu einem Paketpfad.
//
// Eine Schnittstelle, weil `render` nicht weiß, wo Pakete liegen — das
// entscheidet der Paketverwalter aus Phase 7. Hier steht nur, was gebraucht
// wird.
type Bildquelle interface {
	Bild(pfad string) (image.Image, error)
}

// Rasterer zeichnet Pläne.
type Rasterer struct {
	quelle    Bildquelle
	schriften Schriftquelle

	// baumbefunde sind die Befunde, die beim Ablaufen **eines** Bildbaums
	// angefallen sind und deren Knoten trotzdem ein Bild geliefert hat.
	//
	// **Sie stehen hier und nicht im Rückgabewert**, damit kein Elternteil
	// sie verlieren kann — der Grund steht bei [Rasterer.ausBaumTief]
	// (#473).
	//
	// **Sie gehören zu einem Lauf, nicht zum Rasterer** —
	// #748.
	// [Rasterer.ausBaum] arbeitet auf einer **Kopie** des Rasterers; was hier
	// steht, ist deshalb nur für die Dauer eines Bildbaums gültig und wird von
	// keinem zweiten Lauf gesehen.
	baumbefunde []string

	// gefragteGroesse ist die Größe, in der der laufende Bildbaum gebraucht
	// wird — MSEs `Options.width`/`height`.
	//
	// **Nur `blank_image` liest sie.** Jeder andere Erzeuger bringt seine
	// eigene Größe mit; ein leeres Bild hat keine
	// (#509).
	gefragteGroesse image.Point

	// ersatz gibt eine Schrift für ein Zeichen, das die gewählte nicht führt —
	// siehe [Rasterer.MitErsatzschrift]. Nil heißt „kein Ausweich".
	ersatz Ersatzschrift

	// gemeldeteGroesse ist eine Naht **für Tests** —
	// #751.
	//
	// # Warum es sie gibt
	//
	// Es gibt fünf produktive Wege in [Rasterer.ausBaum], und jeder soll die
	// Größe mitgeben, in der sein Bild gebraucht wird. Ob er das tut, ist am
	// **Ergebnis nicht abzulesen**: das Bild wird danach ohnehin auf den Rahmen
	// gebracht, und ein Weg, der versehentlich `image.Point{}` gibt, liefert
	// dasselbe Bild in schlechterer Auflösung.
	//
	// **Ein Feld und keine Paketvariable**: sie gehört zu **einem** Rasterer,
	// und zwei Tests nebeneinander teilen sie sonst.
	//
	// **Sie reist durch die Kopie in `ausBaum` mit** und ist damit auch für die
	// Wege gesetzt, die tief im Baum liegen.
	gemeldeteGroesse func(image.Point)

	// ebenen hält Ebenen vor, die schon auf ihre Größe gebracht sind — siehe
	// `ebenenspeicher.go`.
	//
	// **Ein Zeiger und kein Verbund**: er wird **geteilt** und ist deshalb
	// gesperrt. Läge er als Wert im Rasterer, bekäme jede Kopie in
	// [Rasterer.ausBaum] ihren eigenen — und der Zwischenspeicher wäre keiner
	// mehr (#748).
	ebenen *ebenenspeicher

	// pforte lässt eine Rasterung erst beginnen, wenn ein Platz frei ist —
	// siehe `pforte.go`. Nil heißt „keine Grenze".
	pforte Pforte
}

// NeuerRasterer baut den Rasterer über einer Bildquelle.
//
// **`nil` ist erlaubt und heißt „keine".** Ein Plan ohne Bildebene braucht
// keine, und ein Aufrufer, der trotzdem eine Bildebene schickt, bekommt dafür
// einen Befund — nicht einen Absturz.
func NeuerRasterer(quelle Bildquelle) *Rasterer {
	return &Rasterer{quelle: quelle, ebenen: &ebenenspeicher{}}
}

// errOhneBildquelle meldet eine Bildebene ohne eingerichtete Quelle.
var errOhneBildquelle = errors.New("es ist keine Bildquelle eingerichtet")

// bilddaten holt ein Bild — auch dann, wenn es gar keine Quelle gibt.
//
// **Die eine Stelle, an der `r.quelle` gelesen wird.** Der Zugriff stand
// zweimal in `bild.go`, und beide Male ohne diese Frage; gefunden hat es die
// Gegenprobe zum Goldbild aus #173, nicht ein Test über den Rasterer.
func (r *Rasterer) bilddaten(pfad string) (image.Image, error) {
	if r.quelle == nil {
		return nil, errOhneBildquelle
	}

	return r.quelle.Bild(pfad)
}

// MitSchriften hängt die Schriftsuche an.
//
// Getrennt vom Bau, weil ein Plan ohne Textebene keine braucht — und weil ein
// Aufrufer, der keine anhängt, dafür einen Befund je Textebene bekommt statt
// stillschweigend eine Karte ohne Text.
func (r *Rasterer) MitSchriften(schriften Schriftquelle) *Rasterer {
	r.schriften = schriften

	return r
}

// MitErsatzschrift setzt den Ausweich für fehlende Zeichen.
//
// **Getrennt von [Rasterer.MitSchriften]**, weil es eine andere Frage
// beantwortet: die Quelle sagt „gib mir die Schrift *Beleren Bold*", der
// Ausweich „gib mir **irgendeine**, die `ǵ` führt". Wer beides in eine
// Schnittstelle legte, zwänge jeden Aufrufer zu einer Antwort auf die zweite
// (#515).
func (r *Rasterer) MitErsatzschrift(ersatz Ersatzschrift) *Rasterer {
	r.ersatz = ersatz

	return r
}

// Zeichne rastert einen Plan mit der gegebenen Vergrößerung.
//
// # Vergrößerung statt Bildpunkten
//
// Der Plan misst in Vorlagenpunkten. `1` heißt Bildschirm, `4` heißt Druck —
// **derselbe Plan**, und das ist die Zusage aus #34: eine Vorschau und ein
// Druckbild, die auseinanderlaufen, tun das an genau einer Stelle, wenn es zwei
// Pläne gibt.
//
// # Befunde halten den Lauf nicht an
//
// Ein fehlendes Bild ist ein benannter Vorbehalt, kein Abbruch (AGENTS.md §8).
// Eine Karte mit elf von zwölf Ebenen und einer Meldung ist brauchbar; eine
// leere Karte ohne Meldung ist es nicht.
func (r *Rasterer) Zeichne(
	ctx context.Context, plan *Plan, vergroesserung float64, guete Guete,
) (*image.NRGBA, []Befund, error) {
	bild, _, befunde, err := r.zeichne(ctx, plan, vergroesserung, guete, false)

	return bild, befunde, err
}

// zeichne ist der gemeinsame Rumpf von [Rasterer.Zeichne] und
// [Rasterer.ZeichneMitHerkunft].
//
// **Ein Rumpf und nicht zwei**: zwei Schleifen, die dasselbe malen sollen,
// laufen auseinander, sobald jemand eine ändert — und dann misst die eine ein
// Bild, das die andere nie zeichnet (§10).
func (r *Rasterer) zeichne(
	ctx context.Context, plan *Plan, vergroesserung float64, guete Guete,
	aufzeichnen bool,
) (*image.NRGBA, *Herkunft, []Befund, error) {
	befunde, err := Pruefe(plan)
	if err != nil {
		return nil, nil, nil, err
	}

	if vergroesserung <= 0 || math.IsNaN(vergroesserung) {
		return nil, nil, nil, ErrPlan
	}

	breite := int(math.Round(plan.Breite * vergroesserung))
	hoehe := int(math.Round(plan.Hoehe * vergroesserung))

	if breite <= 0 || hoehe <= 0 {
		return nil, nil, nil, ErrPlan
	}

	// **Erst nach der Prüfung**: wer ohnehin scheitert, belegt keinen Platz.
	freigeben, err := r.platz(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	defer freigeben()

	bild := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))
	zeichenbar, nummern := Zeichenbar(plan, befunde)

	var (
		herkunft *Herkunft
		vorher   []uint8
	)

	if aufzeichnen {
		herkunft = NeueHerkunft(breite, hoehe)
		vorher = make([]uint8, len(bild.Pix))
	}

	for stelle := range zeichenbar.Ebenen {
		if err := ctx.Err(); err != nil {
			return nil, nil, nil, err
		}

		if aufzeichnen {
			copy(vorher, bild.Pix)
		}

		// **Die Nummer aus dem ursprünglichen Plan**, nicht die aus dem
		// gekürzten. Sonst zeigt ein Laufzeitbefund auf die falsche Ebene,
		// sobald davor eine weggefallen ist.
		//
		// **Ein Befund je Mangel**, und jeder mit seiner eigenen Art (QA-03,
		// #1656): eine Ebene kann an der Maske, am Bild und an der Schrift
		// etwas zu melden haben, und eine einzige Meldung bekäme die Art ihres
		// ersten Treffers.
		for _, meldung := range r.ebene(bild, zeichenbar.Ebenen[stelle], vergroesserung, guete) {
			befunde = append(befunde, Befund{
				Ebene:   nummern[stelle],
				Feld:    zeichenbar.Ebenen[stelle].Feld,
				Meldung: meldung,
				Art:     ArtAus(meldung),
			})
		}

		if aufzeichnen {
			herkunft.vermerke(vorher, bild.Pix, zeichenbar.Ebenen[stelle].Art)
		}
	}

	return bild, herkunft, befunde, nil
}

// ebene zeichnet eine Ebene und gibt ihre Mängel — einen je Meldung.
func (r *Rasterer) ebene(
	ziel *image.NRGBA, ebene Ebene, vergroesserung float64, guete Guete,
) []string {
	deckung := ebene.Deckung
	if deckung <= 0 {
		// **Deckung 0 ist unsichtbar, nicht undurchsichtig.** Der Nullwert eines
		// Go-Verbunds ist 0, und eine Ebene, die jemand ohne Deckung
		// hinschreibt, soll nichts zeichnen statt alles zu überdecken. Das ist
		// eine Falle, die man einmal baut und dreimal sucht.
		return nil
	}

	if !Sichtbar(ebene.Rahmen) {
		return nil
	}

	// **Eine Maske kann eine Warnung tragen und trotzdem taugen.** Ist kein
	// Bild herausgekommen, bricht die Ebene ab; ist eines da, wird gezeichnet
	// und die Warnung steht neben dem Ergebnis.
	maske, maskenmaengel := r.maske(ebene, vergroesserung, guete)
	if maske == nil && len(maskenmaengel) > 0 {
		return maskenmaengel
	}

	return append(maskenmaengel, r.ebeneGedreht(ziel, ebene, maske, vergroesserung, guete)...)
}

// ebeneZeichnen malt eine Ebene ihrer Art nach.
func (r *Rasterer) ebeneZeichnen(
	ziel *image.NRGBA, ebene Ebene, maske *maskenbild, vergroesserung float64,
	guete Guete,
) []string {
	switch ebene.Art {
	case ArtFlaeche:
		r.flaeche(ziel, ebene, maske, vergroesserung)

	case ArtBild, ArtSymbol:
		return r.bild(ziel, ebene, maske, vergroesserung, guete)

	case ArtText:
		return r.text(ziel, ebene, maske, vergroesserung, guete)

	case ArtSymbolzug:
		return r.symbolzug(ziel, ebene, maske, vergroesserung, guete)
	}

	return nil
}

// flaeche füllt den Rahmen mit einer Farbe oder einem Verlauf.
//
// **Die Maske wirkt auf die Ebene, bevor sie aufgelegt wird** — nicht auf das
// fertige Bild. Andersherum radierte sie weg, was darunter schon stand: der
// Verlauf überdeckte, was die Maske freigelegt hat, und das sieht auf einer
// einzelnen Karte fast richtig aus.
func (r *Rasterer) flaeche(ziel *image.NRGBA, ebene Ebene, maske *maskenbild, vergroesserung float64) {
	rahmen := skaliert(ebene.Rahmen, vergroesserung)

	for y := rahmen.Min.Y; y < rahmen.Max.Y; y++ {
		for x := rahmen.Min.X; x < rahmen.Max.X; x++ {
			if !(image.Point{X: x, Y: y}).In(ziel.Bounds()) {
				continue
			}

			if !inFlaechenform(ebene, rahmen, x, y, vergroesserung) {
				continue
			}

			farbe := ebene.Farbe
			if ebene.Verlauf != nil {
				farbe = farbeAn(ebene.Verlauf, verlaufsstelle(ebene.Verlauf, rahmen, x, y))
			}

			r.punkt(ziel, x, y, farbe, ebene.Deckung*maske.anteil(rahmen, x, y), nil)
		}
	}
}

// punkt legt eine Farbe über den Bildpunkt, der schon dasteht.
func (r *Rasterer) punkt(
	ziel *image.NRGBA, x, y int, farbe Farbe, anteil float64,
	mischart func(a, b int) int,
) {
	if anteil <= 0 {
		return
	}

	stelle := ziel.PixOffset(x, y)
	unten := Farbe{
		R: ziel.Pix[stelle], G: ziel.Pix[stelle+1],
		B: ziel.Pix[stelle+2], A: ziel.Pix[stelle+3],
	}

	// **Die Mischart rechnet auf den Farben, das Alpha bleibt der Quelle.**
	// MSE nimmt den Zielausschnitt auf, rechnet je Kanal `f(Ziel, Quelle)` und
	// legt das Ergebnis dann gewöhnlich auf — `combine_image` kopiert zuvor
	// den Alphakanal der Quelle hinein (`src/gfx/combine_image.cpp`, Zeilen
	// 108–112 und 142–157, Fassung f9d9356 vom 2021-07-06).
	if mischart != nil {
		farbe = Farbe{
			R: gemischterKanal(mischart, unten.R, farbe.R),
			G: gemischterKanal(mischart, unten.G, farbe.G),
			B: gemischterKanal(mischart, unten.B, farbe.B),
			A: farbe.A,
		}
	}

	gemischt := mischen(unten, gedaempft(farbe, anteil))

	ziel.Pix[stelle] = gemischt.R
	ziel.Pix[stelle+1] = gemischt.G
	ziel.Pix[stelle+2] = gemischt.B
	ziel.Pix[stelle+3] = gemischt.A
}

// verlaufsstelle sagt, wie weit ein Bildpunkt im Verlauf steht.
func verlaufsstelle(verlauf *Verlauf, rahmen image.Rectangle, x, y int) float64 {
	bogen := verlauf.Winkel * math.Pi / 180
	breite := float64(rahmen.Dx())
	hoehe := float64(rahmen.Dy())

	// Auf die Mitte des Bildpunkts, nicht auf seine Ecke: sonst liegt die erste
	// Farbe des Verlaufs auf einem halben Bildpunkt weniger als die letzte.
	rechts := float64(x-rahmen.Min.X) + 0.5
	runter := float64(y-rahmen.Min.Y) + 0.5

	// Die Länge kann nicht null werden: der Aufrufer betritt die Schleife nur
	// für einen Rahmen mit Fläche, und Sinus und Kosinus sind nie beide null.
	laenge := math.Abs(breite*math.Cos(bogen)) + math.Abs(hoehe*math.Sin(bogen))

	// Die Projektion auf die Verlaufsrichtung, verschoben, damit sie bei 0 beginnt.
	entlang := rechts*math.Cos(bogen) + runter*math.Sin(bogen)
	if math.Cos(bogen) < 0 {
		entlang += breite * -math.Cos(bogen)
	}

	if math.Sin(bogen) < 0 {
		entlang += hoehe * -math.Sin(bogen)
	}

	return begrenzt01(entlang / laenge)
}

// skaliert rechnet einen Rahmen aus Vorlagenpunkten in Bildpunkte.
func skaliert(rahmen Rechteck, vergroesserung float64) image.Rectangle {
	links := int(math.Round(rahmen.Links * vergroesserung))
	oben := int(math.Round(rahmen.Oben * vergroesserung))

	return image.Rect(
		links, oben,
		links+int(math.Round(rahmen.Breite*vergroesserung)),
		oben+int(math.Round(rahmen.Hoehe*vergroesserung)),
	)
}
