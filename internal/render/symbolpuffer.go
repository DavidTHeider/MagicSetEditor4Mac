package render

import (
	"math"
	"sort"
)

// **Ein Symbol entsteht in Graustufenpuffern, nicht in Farbe.**
//
// MSE zeichnet ein Symbol in zwei Hilfsflächen — eine für den Rand, eine für
// das Innere — und bringt sie am Ende mit `wxOR` und `wxAND_INVERT` auf den
// Schirm (`combineBuffers`, `src/render/symbol/viewer.cpp`, Zeilen 85–89,
// Fassung f9d9356 vom 2021-07-06). Erst ein Filter macht daraus Farbe.
//
// Ein Puffer hält je Bildpunkt ein Byte, und die Rechenarten sind die von
// wxWidgets (#323).

// Pufferart ist die logische Funktion, mit der geschrieben wird.
type Pufferart int

const (
	// PufferSetzen ist `wxCOPY` — der neue Wert gilt.
	PufferSetzen Pufferart = iota

	// PufferUnd ist `wxAND` — beim Abziehen einer Form.
	PufferUnd

	// PufferOder ist `wxOR` — beim Übertragen des Randes auf den Schirm.
	PufferOder

	// PufferEntweder ist `wxXOR` — beim Unterschied zweier Formen.
	PufferEntweder
)

// Symbolpuffer ist eine Graustufenfläche.
type Symbolpuffer struct {
	breite, hoehe int
	werte         []uint8
}

// NeuerSymbolpuffer legt eine schwarze Fläche an — `clearDC(*newDC, *wxBLACK_BRUSH)`.
func NeuerSymbolpuffer(breite, hoehe int) *Symbolpuffer {
	return &Symbolpuffer{breite: breite, hoehe: hoehe, werte: make([]uint8, breite*hoehe)}
}

// Bei gibt den Wert eines Bildpunkts; außerhalb ist er null.
func (p *Symbolpuffer) Bei(x, y int) uint8 {
	if x < 0 || y < 0 || x >= p.breite || y >= p.hoehe {
		return 0
	}

	return p.werte[y*p.breite+x]
}

// setze schreibt einen Bildpunkt mit der gewählten Rechenart.
//
// **Ohne Prüfung auf die Fläche.** Jeder Aufrufer schneidet vorher zu —
// [Symbolpuffer.Fuelle] über [Symbolpuffer.zeilenbereich] und
// [Symbolpuffer.schnitte], [Symbolpuffer.scheibe] über ihre Schleifengrenzen.
// Ein Wächter hier wäre nie betreten worden, und ein nie betretener Wächter
// sieht aus wie eine Zusage (§3).
func (p *Symbolpuffer) setze(x, y int, wert uint8, art Pufferart) {
	stelle := y*p.breite + x

	switch art {
	case PufferSetzen:
		p.werte[stelle] = wert

	case PufferUnd:
		p.werte[stelle] &= wert

	case PufferOder:
		p.werte[stelle] |= wert

	case PufferEntweder:
		p.werte[stelle] ^= wert
	}
}

// Fuelle füllt ein Vieleck — `DrawPolygon`.
//
// **Gerade und ungerade**, denn das ist die Vorgabe von wxWidgets
// (`wxODDEVEN_RULE`): ein Punkt liegt innen, wenn ein Strahl von ihm aus die
// Kanten ungerade oft schneidet. Ein Symbol mit einem Loch braucht diese
// Regel — der innere Zug macht das Loch, gleich in welcher Richtung er läuft.
//
// **Gemessen wird in der Mitte eines Bildpunkts.** Das ist die Stelle, an der
// diese Umsetzung von wxWidgets abweichen kann: dessen Füllregel für Kanten,
// die genau auf einem Bildpunkt liegen, ist plattformabhängig und nirgends
// zugesagt. Ein Symbol ist deshalb bis auf seine Randpunkte gleich, nicht
// Punkt für Punkt.
func (p *Symbolpuffer) Fuelle(punkte []Punkt, wert uint8, art Pufferart) {
	if len(punkte) < 3 {
		return
	}

	oben, unten := p.zeilenbereich(punkte)

	for y := oben; y < unten; y++ {
		mitte := float64(y) + 0.5

		for _, strecke := range p.schnitte(punkte, mitte) {
			for x := strecke[0]; x < strecke[1]; x++ {
				p.setze(x, y, wert, art)
			}
		}
	}
}

// zeilenbereich sagt, welche Zeilen ein Vieleck überhaupt berührt.
func (p *Symbolpuffer) zeilenbereich(punkte []Punkt) (oben, unten int) {
	kleinste, groesste := punkte[0].Y, punkte[0].Y

	for _, punkt := range punkte[1:] {
		kleinste = math.Min(kleinste, punkt.Y)
		groesste = math.Max(groesste, punkt.Y)
	}

	return max(0, int(math.Floor(kleinste))), min(p.hoehe, int(math.Ceil(groesste))+1)
}

// schnitte gibt die zu füllenden Abschnitte einer Zeile.
//
// **Paarweise**, und das ist die Regel „gerade und ungerade": zwischen dem
// ersten und dem zweiten Schnittpunkt liegt Innen, zwischen dem zweiten und
// dem dritten Außen.
func (p *Symbolpuffer) schnitte(punkte []Punkt, hoehe float64) [][2]int {
	var stellen []float64

	for stelle := range punkte {
		von, nach := punkte[stelle], punkte[(stelle+1)%len(punkte)]

		// **Eine waagerechte Kante schneidet nicht.** Sie liegt entweder ganz
		// über oder ganz unter der Zeilenmitte.
		if (von.Y > hoehe) == (nach.Y > hoehe) {
			continue
		}

		anteil := (hoehe - von.Y) / (nach.Y - von.Y)
		stellen = append(stellen, von.X+anteil*(nach.X-von.X))
	}

	sort.Float64s(stellen)

	abschnitte := make([][2]int, 0, len(stellen)/2)

	for stelle := 0; stelle+1 < len(stellen); stelle += 2 {
		links := max(0, int(math.Ceil(stellen[stelle]-0.5)))
		rechts := min(p.breite, int(math.Ceil(stellen[stelle+1]-0.5)))

		if links < rechts {
			abschnitte = append(abschnitte, [2]int{links, rechts})
		}
	}

	return abschnitte
}

// Striche zieht den Umriss eines Vielecks mit einer Strichbreite.
//
//	border->SetPen(wxPen(*wxWHITE, (int) rotation.trS(border_radius)));
//	border->DrawPolygon((int)points.size(), &points[0]);
//
// `drawSymbolShape`, `src/render/symbol/viewer.cpp`, Zeilen 265–269, Fassung
// f9d9356 vom 2021-07-06.
//
// **Der Strich sitzt mittig auf der Kante.** Die innere Hälfte wird später vom
// Inneren überschrieben — sichtbar bleibt die äußere. Wer ihn nach außen legte,
// bekäme einen doppelt so breiten Rand.
//
// **Runde Ecken und runde Enden**, wie es die Vorgabe von wxWidgets ist
// (`wxJOIN_ROUND`, `wxCAP_ROUND`). Ohne sie klaffte in jeder Ecke ein Loch
// zwischen zwei Strichen.
func (p *Symbolpuffer) Striche(punkte []Punkt, breite float64, wert uint8, art Pufferart) {
	if len(punkte) < 2 || breite <= 0 {
		return
	}

	halbe := breite / 2

	for stelle := range punkte {
		von, nach := punkte[stelle], punkte[(stelle+1)%len(punkte)]

		p.strich(von, nach, halbe, wert, art)
		p.scheibe(von, halbe, wert, art)
	}
}

// strich füllt das Rechteck eines Streckenstücks.
func (p *Symbolpuffer) strich(von, nach Punkt, halbe float64, wert uint8, art Pufferart) {
	laenge := math.Hypot(nach.X-von.X, nach.Y-von.Y)
	if laenge == 0 {
		return
	}

	quer := Punkt{X: -(nach.Y - von.Y) / laenge * halbe, Y: (nach.X - von.X) / laenge * halbe}

	p.Fuelle([]Punkt{
		plus(von, quer), plus(nach, quer), minus(nach, quer), minus(von, quer),
	}, wert, art)
}

// scheibe füllt den Kreis einer Ecke.
//
// **Von Hand und nicht als Vieleck**, damit eine Ecke bei kleiner Strichbreite
// nicht eckig wird: bei drei Bildpunkten Durchmesser hat ein Achteck sichtbare
// Flächen zu wenig.
func (p *Symbolpuffer) scheibe(mitte Punkt, halbe float64, wert uint8, art Pufferart) {
	oben := max(0, int(math.Floor(mitte.Y-halbe)))
	unten := min(p.hoehe, int(math.Ceil(mitte.Y+halbe))+1)

	for y := oben; y < unten; y++ {
		links := max(0, int(math.Floor(mitte.X-halbe)))
		rechts := min(p.breite, int(math.Ceil(mitte.X+halbe))+1)

		for x := links; x < rechts; x++ {
			if math.Hypot(float64(x)+0.5-mitte.X, float64(y)+0.5-mitte.Y) <= halbe {
				p.setze(x, y, wert, art)
			}
		}
	}
}
