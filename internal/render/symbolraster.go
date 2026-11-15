package render

import "math"

// **Rastern heißt bei MSE: zwei Hilfsflächen füllen und sie auf den Schirm
// bringen.**
//
// Der Schirm beginnt grün, der Rand kommt mit `wxOR` darauf, das Innere mit
// `wxAND_INVERT` — was in beiden liegt, wird innen. Sichtbar bleibt vom Rand
// deshalb nur seine **äußere** Hälfte (`SymbolViewer::draw` und
// `combineBuffers`, `src/render/symbol/viewer.cpp`, Zeilen 85–118, Fassung
// f9d9356 vom 2021-07-06).
//
// **Ohne Glättung.** MSE zeichnet mit `wxDC`, und der rundet jeden Punkt auf
// einen Bildpunkt ([Zerlege]); eine geglättete Kante wäre hier hübscher und
// falsch.
//
// # Eine Unterscheidung von MSE fehlt hier, und zwar mit Absicht
//
// MSE zeichnet die erste Form **unmittelbar** auf den Schirm und puffert erst
// ab der zweiten; deshalb geht es vorher einmal durch alle Teile und sucht
// nach `intersection` (Zeilen 98–107), denn mit dem Schnitt trägt der direkte
// Weg nicht. Hier wird **immer** gepuffert. Das Ergebnis ist dasselbe — die
// Puffer werden am Ende mit denselben Rechenarten übertragen —, und der
// Sonderweg gäbe es nur, um eine Fläche zu sparen
// (#323).

// rasterer hält den Zustand eines Laufs — dieselben Größen wie `SymbolViewer`.
type rasterer struct {
	breite, hoehe int
	randbreite    float64

	ursprung Punkt
	matrix   Matrix

	schirmRot, schirmGruen *Symbolpuffer
	rand, inneres          *Symbolpuffer

	// gefuellt ist `buffersFilled` — steht etwas in den Hilfsflächen?
	gefuellt bool
}

// RastereSymbol zeichnet eine Symbolzeichnung in eine Maske — `render_symbol`.
//
// **Die Randbreite ist ein Anteil**, kein Maß in Bildpunkten: MSE rechnet sie
// mit dem Zoom hoch (`rotation.trS(border_radius)`), und `magic-m15` gibt 0,05
// an. Sie wächst dabei mit der **kürzeren** Seite — der Zoom nimmt die längere,
// der Anteil wird um das Verhältnis gekürzt (Zeilen 26–34).
//
// **Und sie ist ganzzahlig**:
//
//	border->SetPen(wxPen(*wxWHITE, (int) rotation.trS(border_radius)));
//
// `drawSymbolShape`, Zeile 268. `(int)` schneidet ab — bei 0,135 auf 100
// Punkten sind es 13 und nicht 13,5. Der Unterschied ist ein halber Punkt und
// damit über beide Seiten genau einer; bei einem Symbol von 66 Punkten, das
// auf 22 verkleinert wird, ist das ein Sechstel der sichtbaren Strichbreite.
//
// **Was hier von MSE abweicht** (uncertain, §15): eine Randbreite, die auf 0
// abschneidet, zeichnet hier keinen Strich. MSE gibt sie an `wxPen` weiter,
// und was der aus der Breite 0 macht, steht in keiner Kopfdatei, die hier
// vorliegt — wx ist auf diesem Rechner nur als Binärdatei da. Erreichbar ist
// der Fall erst unter 20 Punkten Seitenlänge; [Symbolflaeche] gibt keine
// solche her.
//
// **Die verlangte Fläche ist eine Obergrenze, keine Zusage** (#735): sie wird
// auf das Seitenverhältnis der Zeichnung geklemmt, statt das Symbol zu dehnen.
// Was dabei herauskommt, steht in [Symbolmaske.Breite] und
// [Symbolmaske.Hoehe] — wer die Maße braucht, liest sie dort und nimmt nicht
// die, die er hineingegeben hat.
//
// `kleinerErlaubt` ist MSEs `allow_smaller`. Ohne das Kennzeichen wird nur
// geklemmt, wenn die Zeichnung selbst ausgesprochen flach oder hoch ist —
// `SymbolToImage::generate` setzt es, wenn beide Maße genannt sind
// (`src/gfx/generated_image.cpp`, Zeile 469).
//
// Fassung f9d9356 vom 2021-07-06.
func RastereSymbol(
	teile []Symbolteil, breite, hoehe int, randbreite float64, kleinerErlaubt bool,
) Symbolmaske {
	breite, hoehe = aufsSeitenverhaeltnis(
		Seitenverhaeltnis(teile), breite, hoehe, kleinerErlaubt)

	zoom := float64(max(breite, hoehe))

	r := &rasterer{
		breite: breite, hoehe: hoehe,
		randbreite: math.Trunc(randbreite * float64(min(breite, hoehe))),
		matrix:     Skaliert(zoom),
		ursprung:   randursprung(breite, hoehe),

		schirmRot:   NeuerSymbolpuffer(breite, hoehe),
		schirmGruen: NeuerSymbolpuffer(breite, hoehe),
		rand:        NeuerSymbolpuffer(breite, hoehe),
		inneres:     NeuerSymbolpuffer(breite, hoehe),
	}

	// **Der Schirm beginnt grün** — `clearDC(dc, Color(0,128,0))`.
	r.schirmGruen.Fuelle(rahmenpunkte(breite, hoehe), 128, PufferSetzen)

	for stelle := len(teile) - 1; stelle >= 0; stelle-- {
		r.teil(teile[stelle], true)
	}

	if r.gefuellt {
		r.uebertrage()
	}

	return Symbolmaske{Breite: breite, Hoehe: hoehe, rot: r.schirmRot, gruen: r.schirmGruen}
}

// aufsSeitenverhaeltnis bringt die verlangte Fläche auf das der Zeichnung.
//
//	double par = (double)width/height;
//	if (par > ar && (ar > 1 || (allow_smaller && height < width))) {
//	  width  = int(height * ar);
//	} else if (par < ar && (ar < 1 || (allow_smaller && width < height))) {
//	  height = int(width / ar);
//	}
//
// `render_symbol`, Zeilen 18–25.
//
// **`int(…)` schneidet ab**, es rundet nicht — deshalb steht hier eine
// Umwandlung und kein `math.Round`.
//
// **Eine Fläche ohne Höhe bleibt, wie sie ist.** MSE käme dort auf `inf` und
// verglich es weiter; hier wäre es eine Division durch null im ersten Zweig,
// und eine Fläche ohne Höhe zeichnet ohnehin nichts.
func aufsSeitenverhaeltnis(
	verhaeltnis float64, breite, hoehe int, kleinerErlaubt bool,
) (geklemmteBreite, geklemmteHoehe int) {
	if breite <= 0 || hoehe <= 0 {
		return breite, hoehe
	}

	flaeche := float64(breite) / float64(hoehe)

	switch {
	case flaeche > verhaeltnis && (verhaeltnis > 1 || (kleinerErlaubt && hoehe < breite)):
		return int(float64(hoehe) * verhaeltnis), hoehe

	case flaeche < verhaeltnis && (verhaeltnis < 1 || (kleinerErlaubt && breite < hoehe)):
		return breite, int(float64(breite) / verhaeltnis)

	default:
		return breite, hoehe
	}
}

// randursprung schiebt das Symbol in die Mitte, wenn das Bild nicht quadratisch ist.
//
//	viewer.setZoom(width);
//	viewer.setOrigin(Vector2D(0,-(width-height) * 0.5));
//
// `render_symbol`, Zeilen 26–34.
func randursprung(breite, hoehe int) Punkt {
	if breite > hoehe {
		return Punkt{Y: -float64(breite-hoehe) * 0.5}
	}

	return Punkt{X: -float64(hoehe-breite) * 0.5}
}

// rahmenpunkte ist das Vieleck über die ganze Fläche.
func rahmenpunkte(breite, hoehe int) []Punkt {
	return []Punkt{
		{X: 0, Y: 0},
		{X: float64(breite), Y: 0},
		{X: float64(breite), Y: float64(hoehe)},
		{X: 0, Y: float64(hoehe)},
	}
}

// teil ist `combineSymbolPart`.
func (r *rasterer) teil(teil Symbolteil, darfUeberdecken bool) {
	switch teil.Art {
	case TeilForm:
		r.form(teil, darfUeberdecken)

	case TeilSymmetrie:
		r.symmetrie(teil, darfUeberdecken)

	case TeilGruppe:
		// **Von hinten nach vorn** — die unterste Form zuerst.
		for stelle := len(teil.Teile) - 1; stelle >= 0; stelle-- {
			r.teil(teil.Teile[stelle], darfUeberdecken)
		}
	}
}

// form zeichnet eine Form und räumt vorher den Puffer, wenn sie überdeckt.
func (r *rasterer) form(teil Symbolteil, darfUeberdecken bool) {
	if teil.Kombination == KombinationUeberdecken && r.gefuellt && darfUeberdecken {
		// **Was überdeckt wird, kommt vorher auf den Schirm** (Zeilen 121–135).
		r.uebertrage()
		r.gefuellt = false

		r.rand = NeuerSymbolpuffer(r.breite, r.hoehe)
		r.inneres = NeuerSymbolpuffer(r.breite, r.hoehe)
	}

	r.verrechne(teil)

	r.gefuellt = true
}

// uebertrage ist `combineBuffers`.
//
//	if (borders)  dc.Blit(…, borders,  0, 0, wxOR);
//	if (interior) dc.Blit(…, interior, 0, 0, wxAND_INVERT);
func (r *rasterer) uebertrage() {
	for y := range r.hoehe {
		for x := range r.breite {
			rand, inneres := r.rand.Bei(x, y), r.inneres.Bei(x, y)

			r.schirmRot.setze(x, y, rand, PufferOder)
			r.schirmGruen.setze(x, y, rand, PufferOder)

			r.schirmRot.setze(x, y, ^inneres, PufferUnd)
			r.schirmGruen.setze(x, y, ^inneres, PufferUnd)
		}
	}
}

// verrechne ist `combineSymbolShape` — die sechs Arten von `combine:`.
func (r *rasterer) verrechne(teil Symbolteil) {
	punkte := r.zug(teil)

	switch teil.Kombination {
	case KombinationUeberdecken, KombinationVereinen:
		r.zeichne(punkte, r.rand, r.inneres, 255, 255, PufferSetzen, false)

	case KombinationAbziehen:
		// **Der Rand wird verundet, das Innere umgedreht** (Zeilen 226–230).
		r.zeichne(punkte, r.rand, r.inneres, 0, 0, PufferUnd, false)

	case KombinationSchneiden:
		r.schneide(punkte)

	case KombinationUnterschied:
		// **Das Innere mit `wxXOR`**, und der Rand wird hinterher gelöscht
		// (Zeilen 240–244 mit `clear = true`).
		r.zeichne(punkte, r.rand, r.inneres, 0, 255, PufferEntweder, true)

	case KombinationUmranden:
		// **Der Rand wird als Inneres gezeichnet** (Zeile 247).
		r.zeichne(punkte, nil, r.rand, 0, 255, PufferSetzen, false)
	}
}

// schneide behält, was in beiden liegt.
//
//	drawSymbolShape(shape, keepBorder.get(), keepInterior.get(), 255, 255, false, false);
//	border  .Blit(…, &*keepBorder  , 0, 0, wxAND);
//	interior.Blit(…, &*keepInterior, 0, 0, wxAND);
func (r *rasterer) schneide(punkte []Punkt) {
	rand := NeuerSymbolpuffer(r.breite, r.hoehe)
	inneres := NeuerSymbolpuffer(r.breite, r.hoehe)

	r.zeichne(punkte, rand, inneres, 255, 255, PufferSetzen, false)

	for y := range r.hoehe {
		for x := range r.breite {
			r.rand.setze(x, y, rand.Bei(x, y), PufferUnd)
			r.inneres.setze(x, y, inneres.Bei(x, y), PufferUnd)
		}
	}
}

// zeichne ist `drawSymbolShape`.
//
// **Der Rand wird gefüllt und gestrichen.** Die Füllung trägt `randfarbe`, der
// Strich ist immer weiß — er ist der Rand, die Füllung ist nur das, was das
// Innere gleich wieder wegnimmt (Zeilen 264–286).
func (r *rasterer) zeichne(
	punkte []Punkt, rand, inneres *Symbolpuffer,
	randfarbe, innenfarbe uint8, art Pufferart, loeschen bool,
) {
	if rand != nil && r.randbreite > 0 {
		rand.Fuelle(punkte, randfarbe, art)
		rand.Striche(punkte, r.randbreite, 255, art)

		if loeschen {
			rand.Fuelle(punkte, 0, PufferSetzen)
		}
	}

	if inneres != nil {
		inneres.Fuelle(punkte, innenfarbe, art)
	}
}

// zug macht aus den Punkten einer Form ihr Vieleck.
func (r *rasterer) zug(teil Symbolteil) []Punkt {
	punkte := make([]Punkt, 0, len(teil.Punkte))

	for stelle := range teil.Punkte {
		von := teil.Punkte[stelle]
		nach := teil.Punkte[(stelle+1)%len(teil.Punkte)]

		punkte = append(punkte, Zerlege(von, nach, r.ursprung, r.matrix)...)
	}

	return punkte
}

// symmetrie zeichnet die Kinder einer Symmetrie so oft, wie sie sagt.
//
//	Radians b = 2 * s->handle.angle();
//	int copies = s->kind == SYMMETRY_REFLECTION ? s->copies / 2 * 2 : s->copies;
//	…
//	double a = i * 2 * M_PI / copies;
//	Matrix2D rot(cos(a),-sin(a), sin(a),cos(a));       // Drehung
//	Matrix2D rot(cos(a+b),sin(a+b), sin(a+b),-cos(a+b)); // Spiegelung
//	multiply = rot * old_m;
//	origin = old_o + (s->center - s->center * rot) * old_m;
//
// Zeilen 150–200.
//
// **Bei einer Spiegelung ist jede zweite Kopie gespiegelt**, und die Zahl der
// Kopien wird auf eine gerade abgerundet.
func (r *rasterer) symmetrie(teil Symbolteil, darfUeberdecken bool) {
	kopien := teil.Kopien
	if teil.Spiegelart == SpiegelungAchse {
		kopien = kopien / 2 * 2
	}

	if kopien < 1 {
		return
	}

	alteMatrix, alterUrsprung := r.matrix, r.ursprung
	spiegelwinkel := 2 * math.Atan2(teil.Griff.Y, teil.Griff.X)

	for stelle := len(teil.Teile) - 1; stelle >= 0; stelle-- {
		for i := kopien - 1; i >= 0; i-- {
			winkel := float64(i) * 2 * math.Pi / float64(kopien)

			drehung := Matrix{
				A: math.Cos(winkel), B: -math.Sin(winkel),
				C: math.Sin(winkel), D: math.Cos(winkel),
			}

			if teil.Spiegelart == SpiegelungAchse && i%2 != 0 {
				gesamt := winkel + spiegelwinkel
				drehung = Matrix{
					A: math.Cos(gesamt), B: math.Sin(gesamt),
					C: math.Sin(gesamt), D: -math.Cos(gesamt),
				}
			}

			r.matrix = drehung.verkettet(alteMatrix)
			r.ursprung = plus(alterUrsprung,
				alteMatrix.angewandt(minus(teil.Mitte, drehung.angewandt(teil.Mitte))))

			r.teil(teil.Teile[stelle], darfUeberdecken && i == kopien-1)
		}
	}

	r.matrix, r.ursprung = alteMatrix, alterUrsprung
}
