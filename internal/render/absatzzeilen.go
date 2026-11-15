package render

// absatzInZeilen bricht einen Absatz auf die Breite um.
//
// **Ein Wort bleibt zusammen, auch wenn der Schnitt darin wechselt.** „Feuer"
// kursiv und „ball" aufrecht sind ein Wort; sie an der Abschnittsgrenze zu
// trennen ergäbe eine Zeile, die niemand geschrieben hat — dieselbe Begründung
// wie bei `{T}` in [Setze].
func absatzInZeilen(
	absatz []Abschnitt, breite func(zeile int) float64, satz Satz,
	schriften Schriftsatz, bilder Bildquelle,
) (zeilen [][]Stueck, einzug float64) {
	// **Der Einzug entsteht vor dem Umbruch, nicht daneben.** Er verkürzt
	// jede Zeile **außer der ersten** — dort stehen die Zeichen selbst.
	einzug = randbreite(absatz, satz, schriften, bilder)

	worte := worteAus(absatz, satz, schriften, bilder)
	if len(worte) == 0 {
		// Eine leere Zeile bleibt eine leere Zeile: sie trennt zwei Absätze,
		// und wer sie wegwirft, schiebt den Regeltext an den Namen.
		return [][]Stueck{nil}, einzug
	}

	var laufend []Stueck

	frei := func() float64 {
		if len(zeilen) > 0 {
			return breite(len(zeilen)) - einzug
		}

		return breite(len(zeilen))
	}

	for _, wort := range worte {
		versuch := append(append([]Stueck{}, laufend...), wort...)
		amAnfang := len(laufend) == 0

		switch {
		case amAnfang:
			laufend = versuch
		case breiteDerStuecke(versuch, satz) > frei():
			zeilen = append(zeilen, laufend)
			// **Ohne die führende Lücke.** Ein Zeilenanfang trägt kein
			// Leerzeichen; es stünde sonst als Einzug da, den niemand gesetzt
			// hat.
			laufend = ohneLuecke(wort)
			amAnfang = true
		default:
			laufend = versuch
		}

		// **Ein Wort, das allein nicht passt, wird gebrochen**
		// ([zwangsgebrochen]) — so oft, bis der Rest passt.
		for amAnfang && breiteDerStuecke(laufend, satz) > frei() {
			kopf, rest := zwangsgebrochen(laufend, frei(), satz, schriften)
			if len(rest) == 0 {
				break
			}

			zeilen = append(zeilen, kopf)
			laufend = rest
		}
	}

	return append(zeilen, laufend), einzug
}

// randbreite misst, was aus diesem Absatz in seinen linken Rand geht.
func randbreite(
	absatz []Abschnitt, satz Satz, schriften Schriftsatz, bilder Bildquelle,
) float64 {
	var breite float64

	for _, teil := range teileAus(absatz, satz, schriften, bilder) {
		if !teil.ImRand {
			continue
		}

		breite += teil.Breite
	}

	return breite
}
