package render

// **Ein gesetzter Zug ist mehr als eine Folge von Zeilen.**
//
// Vorlagen rechnen mit seinen **Gruppen**: `card_style.<feld>.layout.blocks[i]`
// setzt bei den Planeswalker-Stilen von `M15-Magic-Pack` jede Fähigkeitszeile.
// Die Gruppierung steht hier und nicht in `abschnitte.go`, weil sie eine eigene
// Frage beantwortet: dort steht, **wie** ein Zug umbricht — hier, **wozu** die
// Zeilen zusammengehören
// (#338).

// absatzenden sagt je Zeile, ob sie einen Absatz schließt.
func absatzenden(roh []rohzeile) []bool {
	enden := make([]bool, len(roh))
	for stelle := range roh {
		enden[stelle] = roh[stelle].absatzende && !roh[stelle].weich
	}

	return enden
}

// blockenden sagt je Zeile, ob hinter ihr ein Block endet.
//
// **Die letzte Zeile schließt immer.** Sonst bliebe der letzte Block offen,
// und `blocks` zählte einen zu wenig.
func blockenden(roh []rohzeile) []bool {
	enden := make([]bool, len(roh))

	for stelle := range roh {
		enden[stelle] = roh[stelle].blockende || stelle == len(roh)-1
	}

	return enden
}

// gruppiert fasst Zeilen zu Gruppen zusammen.
//
// Die Maße folgen `update_size` (`src/render/text/viewer.cpp`, Zeilen 350–353):
// die Breite ist die **größte** Zeilenbreite, die Höhe reicht von der Oberkante
// der ersten bis zur Unterkante der letzten Zeile.
//
// Fassung f9d9356 vom 2021-07-06.
func gruppiert(zeilen []Zeile, enden []bool) []Zeilengruppe {
	var (
		gruppen  []Zeilengruppe
		laufende *Zeilengruppe
	)

	for stelle, zeile := range zeilen {
		if laufende == nil {
			gruppen = append(gruppen, Zeilengruppe{Oben: zeile.Oben})
			laufende = &gruppen[len(gruppen)-1]
		}

		laufende.Zeilen = append(laufende.Zeilen, zeile)
		laufende.Breite = max(laufende.Breite, zeile.Breite)
		laufende.Hoehe = max(laufende.Hoehe, zeile.Oben+zeile.Hoehe-laufende.Oben)

		if enden[stelle] {
			laufende = nil
		}
	}

	return gruppen
}

// TrennerZwischen gibt die Mitten zwischen aufeinanderfolgenden Blöcken.
//
// **Ausgeführt, weil sie ohne `<line>` nie zwei Blöcke sieht.** Dieses
// Programm schneidet noch nicht an `<line>`; die Rechnung ist damit aus einem
// gesetzten Zug heraus nicht prüfbar, und ein Wächter, der sie nie betritt,
// prüft nichts (§10).
func TrennerZwischen(bloecke []Zeilengruppe) []float64 {
	if len(bloecke) < 2 {
		return nil
	}

	trenner := make([]float64, 0, len(bloecke)-1)
	for stelle := range len(bloecke) - 1 {
		trenner = append(trenner,
			(bloecke[stelle].Oben+bloecke[stelle].Hoehe+bloecke[stelle+1].Oben)/2)
	}

	return trenner
}

// linksbuendig nimmt einer Ausrichtung ihren waagerechten Teil.
func linksbuendig(ausrichtung Ausrichtung) Ausrichtung {
	switch ausrichtung {
	case MitteMitte, RechtsMitte, LinksMitte:
		return LinksMitte

	case MitteUnten, RechtsUnten, LinksUnten:
		return LinksUnten
	}

	return LinksOben
}
