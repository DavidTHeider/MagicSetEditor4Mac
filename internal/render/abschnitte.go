package render

// SetzeAbschnitte bricht einen Zug aus Abschnitten um und richtet ihn aus.
//
// **Der Umbruch geht über die Abschnittsgrenze.** „Feuer" kursiv und „ball"
// aufrecht sind ein Wort und bleiben zusammen; sie an der Grenze zu trennen
// ergäbe eine Zeile, die niemand geschrieben hat — dieselbe Begründung wie bei
// `{T}` in [Setze].
func SetzeAbschnitte(
	satz Satz, rahmen Rechteck, abschnitte []Abschnitt, schriften Schriftsatz,
	bilder Bildquelle,
) Satzergebnis {
	// **`scale down to:` — passt der Text nicht, wird er kleiner gesetzt**
	// (#498).
	//
	// Steht vor allem anderen, weil es den ganzen Satz wiederholt: es sucht
	// die größte Schriftgröße, bei der das Ergebnis in den Kasten passt, und
	// gibt dann dessen Satz zurück.
	if gesetzt, verkleinert := verkleinertGesetzt(
		satz, rahmen, abschnitte, schriften, bilder); verkleinert {
		return gesetzt
	}

	// **Gesetzt wird auf ¼ Punkt** (#2302) — nach der Suche, denn jeder ihrer
	// Versuche kommt selbst hier durch, mit Größe mal Maßstab.
	satz.Groesse = aufViertelpunkt(satz.Groesse)

	// **Das Polster schneidet den Kasten, bevor gesetzt wird.** MSE zieht es
	// von der Fläche ab und beginnt die erste Zeile darunter
	// (`src/render/text/viewer.cpp`, Zeilen 567 und 745–748).
	//
	// Fassung f9d9356 vom 2021-07-06.
	innen := Rechteck{
		Links:  rahmen.Links + satz.Polster.Links,
		Oben:   rahmen.Oben + satz.Polster.Oben,
		Breite: max(rahmen.Breite-satz.Polster.Links-satz.Polster.Rechts, 0),
		Hoehe:  max(rahmen.Hoehe-satz.Polster.Oben-satz.Polster.Unten, 0),
	}

	// **Die Maske misst am äußeren Kasten** — `rowLeft(y,
	// dc.getInternalSize())` bekommt die Feldgröße samt Polster.
	raum := zeilenraum{
		maske:   zeilenmaskeAus(bilder, satz.Zeilenmaske, rahmen.Breite, rahmen.Hoehe),
		feld:    rahmen,
		polster: satz.Polster,
	}

	rahmen = innen

	// **Erst umbrechen, dann stauchen** (#2308). Bei MSE hängt der Umbruch
	// allein an `multi line` und am harten Umbruch
	// (`TextViewer::prepareLinesAtScale`, `src/render/text/viewer.cpp`,
	// Zeilen 590–618) — ein einzeiliges Feld bricht nur am harten, ein
	// mehrzeiliges wie jedes andere, gestaucht oder nicht ([Satz.Einzeilig]).
	// Bis dahin ersetzte ein gestauchter Satz den gebrochenen durch **eine**
	// Zeile: 22 mehrzeilige Feldstile mit `shrink` (Embleme, Zählerbögen,
	// Battle) brachen gar nicht um. Das Zeichen, das eine Zeile bricht,
	// zeichnet [umbrochen] nicht mit (`viewer.cpp`, Zeilen 664–666) —
	// Fassung f9d9356 vom 2021-07-06.
	roh := umbrochen(abschnitte, raum, satz, schriften, bilder)

	// **Gemessen wird ungestaucht, und an der breitesten Zeile.** MSE rechnet
	// den Faktor aus dem fertigen Layout (`layout->width`, die Breite der
	// breitesten Zeile; `TextStyle::getStretch`, `src/data/field/text.cpp`,
	// Zeilen 73–81), und das entsteht ohne ihn — wer schon gestaucht misst,
	// bekommt beim nächsten Mal Faktor 1.
	//
	// Fassung f9d9356 vom 2021-07-06.
	gestaucht := stauchung(satz, rahmen, breitesteZeile(roh, satz))

	ergebnis := Satzergebnis{
		Zeilen:    make([]Zeile, 0, len(roh)),
		Groesse:   satz.Groesse,
		Stauchung: gestaucht,
	}

	arten := umbruchartenVon(roh)
	oben, hoehen := zeilenlagen(roh, arten, raum, satz, schriften, bilder)
	ergebnis.Unterkante = unterkante(oben, hoehen, arten, satz, raum.feld.Hoehe)

	// **Wo eine Zeile beginnt, sagt die Maske an ihrer Oberkante** — vor der
	// Dehnung, denn mit dieser Oberkante wurde gebrochen ([zeilenraum]).
	versaetze := make([]float64, len(roh))
	for nummer := range roh {
		versaetze[nummer], _ = raum.bei(oben[nummer])
	}

	// **Erst dehnen, dann ausrichten.** MSE verteilt den freien Platz auf die
	// Abstände und rechnet den senkrechten Versatz aus der **gedehnten** Höhe
	// (`alignParagraph`, Zeilen 780–819).
	var absolut []float64

	if satz.Absatzhoehe > 0 {
		absolut, ergebnis.Hoehe = inBaender(
			oben, hoehen, arten, margenOben(roh), satz, rahmen.Oben)
	} else {
		oben, ergebnis.Hoehe = gedehnt(
			oben, hoehen, arten, satz.Zeilenfaktoren, rahmen.Hoehe)

		versatz := rahmen.Oben +
			senkrechterVersatz(satz.Ausrichtung, rahmen.Hoehe, ergebnis.Hoehe)

		absolut = make([]float64, len(oben))
		for nummer := range oben {
			absolut[nummer] = versatz + oben[nummer]
		}
	}

	for nummer := range roh {
		stuecke := roh[nummer].stuecke
		breite := breiteDerStuecke(stuecke, satz)
		// **Ausgerichtet wird über die harte Breite.** Ein weicher Trenner am
		// Zeilenende zählt nicht mit — `line.width()` endet am letzten nicht
		// weichen Zeichen (`src/render/text/viewer.cpp`, Zeilen 17–35 und 610,
		// Fassung f9d9356 vom 2021-07-06).
		// **Der Absatzrand kommt vor die Ausrichtung.** Er verschiebt die
		// Zeile und verkürzt zugleich die Fläche, in der ausgerichtet wird —
		// bei MSE ist beides dieselbe Rechnung (`viewer.cpp:548–552`, `:834`).
		marge := roh[nummer].margen

		ziel := max(rahmen.Breite-marge.Links-marge.Rechts-roh[nummer].einzug, 0)
		hart := breiteDerStuecke(harteStuecke(stuecke), satz)

		links := rahmen.Links + versaetze[nummer] + marge.Links + roh[nummer].einzug +
			waagerechterVersatz(zeilenausrichtung(satz, hart, ziel), ziel, hart)

		ergebnis.Zeilen = append(ergebnis.Zeilen, Zeile{
			Text:    textAus(stuecke),
			Stuecke: verteile(stuecke, links, satz),
			Links:   links,
			Oben:    absolut[nummer],
			Hoehe:   hoehen[nummer],
			Breite:  breite,
		})
	}

	ergebnis.Absaetze = gruppiert(ergebnis.Zeilen, absatzenden(roh))
	ergebnis.Bloecke = gruppiert(ergebnis.Zeilen, blockenden(roh))
	ergebnis.Trenner = TrennerZwischen(ergebnis.Bloecke)

	return ergebnis
}

// zeilenhoehe gibt die Höhe einer Zeile.
//
// **Die Schrift gibt die Untergrenze, ein Symbol darf sie heben.** MSE nimmt
// die Zeilenhöhe aus den `CharInfo` ihrer Elemente, und ein Symbolelement
// meldet seinen ganzen Kasten (`SymbolFont::getCharInfo`, Zeilen 384–393,
// Fassung f9d9356 vom 2021-07-06). Ohne das liefe ein Manasymbol im Regeltext
// in die Zeile darunter: bei `magic-m15` misst es 18,5 Punkte, die Textzeile
// 16,8.
//
// **Je Zeile und nicht je Zug.** Ein Regeltext, dessen erste Zeile ein Symbol
// trägt, bekäme sonst überall den größeren Abstand.
func zeilenhoehe(
	stuecke []Stueck, satz Satz, schriften Schriftsatz, bilder Bildquelle,
) float64 {
	hoehe := gerastertesTextmass(schriften.Aufrecht.Zeilenhoehe(), satz.Groesse)
	if satz.Zeilenabstand > 0 {
		hoehe = satz.Groesse * satz.Zeilenabstand
	}

	for _, stueck := range stuecke {
		for i := range stueck.Symbole {
			symbol := &stueck.Symbole[i]

			hoehe = max(hoehe, stueck.Symbolmasse.kasten(symbol, bilder).Hoehe)
		}
	}

	return hoehe
}

// istLeer sagt, ob eine Zeile nichts trägt — weder Text noch Symbol.
func istLeer(zeile rohzeile) bool {
	for _, stueck := range zeile.stuecke {
		if stueck.Text != "" || len(stueck.Symbole) > 0 {
			return false
		}
	}

	return true
}

// verteile setzt die Stücke einer Zeile nebeneinander.
func verteile(stuecke []Stueck, links float64, satz Satz) []Stueck {
	gesetzt := make([]Stueck, 0, len(stuecke))
	lauf := links

	for _, stueck := range stuecke {
		// **Die Breite steht schon da** (#598): [teileAus] hat sie als
		// Differenz zweier gerasterter Vorsilben gesetzt. Sie hier neu zu
		// rechnen hieße, den Abschnitt zu vergessen, zu dem sie gehört.
		stueck.Links = lauf

		// **Die Laufweite steht zwischen den Stücken wie zwischen Buchstaben.**
		// Ohne sie säße das zweite Stück um eine Laufweite zu weit links, und
		// bei den 0,2 Punkten aus `magic-m15` fiele das erst im Vergleich
		// zweier Karten auf — derselbe Grund wie in [breiteVon].
		lauf += stueck.Breite
		if len(stueck.Symbole) == 0 && buchstabenIn(stueck.Text) > 0 {
			lauf += satz.Laufweite
		}

		gesetzt = append(gesetzt, stueck)
	}

	return gesetzt
}

// Aufrecht macht aus einer Zeichenkette einen einzigen aufrechten Abschnitt.
//
// **Der häufige Fall bleibt kurz.** Ein Kartenname hat keine Auszeichnung, und
// `[]Abschnitt{{Text: name}}` an jeder Stelle wäre Zeremonie ohne Aussage.
func Aufrecht(text string) []Abschnitt { return []Abschnitt{{Text: text}} }

// Klartext gibt den Text einer Ebene ohne seine Schnitte.
func (e Ebene) Klartext() string { return klartext(e.Abschnitte) }

// margenOben zieht die oberen Absatzränder heraus.
//
// [inBaender] rechnet jeden Absatz von null an und braucht den Rand deshalb
// einzeln — beim Verschieben ginge er sonst verloren.
func margenOben(roh []rohzeile) []float64 {
	raender := make([]float64, len(roh))
	for stelle := range roh {
		raender[stelle] = roh[stelle].margenOben()
	}

	return raender
}

// inBaender legt jeden Absatz in ein Band fester Höhe.
//
// **Portiert aus `alignParagraph`s Aufrufstelle**:
//
//	for (size_t last = 0 ; last < lines.size() ; ++last) {
//	  if (lines[last].break_after != LineBreak::SOFT || last+1 == lines.size()) {
//	    alignParagraph(start, last + 1, chars, style,
//	                   RealRect(0, style.padding_top + n*style.paragraph_height,
//	                            s.width, style.paragraph_height));
//	    start = last + 1;
//	    ++n;
//	  }
//	}
//
// `src/render/text/viewer.cpp`, Zeilen 755–763, Fassung f9d9356 vom
// 2021-07-06.
//
// **Gedehnt wird im Band und nicht im Kasten.** Bei MSE ist das Rechteck, das
// `alignParagraph` bekommt, das Band; wer stattdessen über den ganzen Kasten
// dehnte, verteilte den freien Platz einmal statt je Absatz.
//
// **Ob ein Zug mit Absatzband passt, steht nicht hier**: MSE rechnet dafür
// `max_height * n` mit `n` aus der Kastenhöhe (ebd. Zeilen 718–732) — das tut
// [unterkante]. Bis zum 2026-10-11 stand hier die Unterkante des letzten
// Bands, als benannte Abweichung
// (#479).
func inBaender(
	oben, hoehen []float64, arten []umbruchart, raender []float64,
	satz Satz, kastenOben float64,
) (absolut []float64, hoehe float64) {
	absolut = make([]float64, len(oben))

	start, band := 0, 0

	for letzte := range oben {
		// **Die Absatzgrenze ist `break_after != SOFT`** (ebd. Zeile 758). Die
		// letzte Zeile trägt [umbruchKeiner] und schließt damit von selbst.
		if arten[letzte] == umbruchWeich {
			continue
		}

		// Der Absatz für sich, von seinem oberen Rand an gerechnet.
		eigene := make([]float64, letzte-start+1)
		for i := range eigene {
			eigene[i] = oben[start+i] - oben[start] + raender[start]
		}

		eigene, absatzhoehe := gedehnt(eigene, hoehen[start:letzte+1],
			arten[start:letzte+1], satz.Zeilenfaktoren, satz.Absatzhoehe)

		bandOben := kastenOben + float64(band)*satz.Absatzhoehe +
			senkrechterVersatz(satz.Ausrichtung, satz.Absatzhoehe, absatzhoehe)

		for i := range eigene {
			absolut[start+i] = bandOben + eigene[i]
		}

		hoehe = bandOben + absatzhoehe - kastenOben
		start, band = letzte+1, band+1
	}

	return absolut, hoehe
}

// breitesteZeile ist die harte Breite der breitesten Zeile — ungestaucht und
// ohne weichen Schwanz ([harteStuecke]), wie `Line::width` bis
// `end_or_soft` (`src/render/text/viewer.cpp`, Zeile 35, Fassung f9d9356 vom
// 2021-07-06).
func breitesteZeile(roh []rohzeile, satz Satz) float64 {
	breiteste := 0.0
	for _, zeile := range roh {
		breiteste = max(breiteste, breiteDerStuecke(harteStuecke(zeile.stuecke), satz))
	}

	return breiteste
}

// zeilenausrichtung ist `should_fill` aus `Line::alignHorizontal`
// (`src/render/text/viewer.cpp`, Zeilen 834–869, Fassung f9d9356 vom
// 2021-07-06): **eine Zeile, die füllt, wird nicht ausgerichtet** — MSE lässt
// ihre Stellen im Stretch-Zweig stehen, und sie beginnt an der linken Kante.
// Sie füllt, wenn das Feld staucht und sie entweder ohne Überlaufbedingung
// steht oder selbst breiter ist als ihr Platz. **Je Zeile und nicht je Feld**
// (Zweitblick über die Folgen vor 0.8.0, Befund 11): eine kurze Zeile neben
// einer überlaufenden bleibt in ihrer Ausrichtung.
func zeilenausrichtung(satz Satz, breite, ziel float64) Ausrichtung {
	if satz.Stauchen && (!satz.NurBeiUeberlauf || breite > ziel) {
		return linksbuendig(satz.Ausrichtung)
	}

	return satz.Ausrichtung
}

// aufViertelpunkt ist die Größe, mit der MSE2 eine Schrift setzt: als
// wxFont mit ganzzahliger Punktgröße in vierfacher Auflösung,
// `to_int(4 · Größe · Maßstab)` (`Font::toWxFont`, `src/data/font.cpp`,
// Zeile 84; `text_scaling = 4`, `src/gfx/resample_text.cpp`, Zeile 22;
// `RotatedDC::SetFont`, `src/util/rotation.cpp`, Zeile 326 — Fassung
// f9d9356 vom 2021-07-06). `to_int` addiert 0,4999995 und nicht ½
// (`src/util/vector2d.hpp`, Zeile 20), die genaue Mitte rundet also ab.
//
// **Gemessen und gezeichnet wird mit derselben Größe**: MSE2 misst über
// `GetTextExtent` an genau dieser Schrift. Ein verkleinerter Text stand bei
// MSE2 deshalb bis zu ⅛ Punkt kleiner da als gerechnet — Clockwork Beast
// rund ein Prozent enger, und ein Wort passte noch in die Zeile (#2302).
//
// **Eine Größe unter ⅛ Punkt bleibt, wie sie ist**: wx bekäme 0 und nähme
// seine Vorgabe; eine solche Schrift setzt keine Vorlage.
func aufViertelpunkt(groesse float64) float64 {
	viertel := int(groesse*4 + 0.4999995)
	if viertel <= 0 {
		return groesse
	}

	return float64(viertel) / 4
}
