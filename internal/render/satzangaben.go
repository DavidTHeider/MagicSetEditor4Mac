package render

// Die **Satzangaben** eines Zuges stehen hier — was der Rasterer braucht, um
// Text zu setzen.
//
// Eigene Datei, weil `plan.go` mit ihnen über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort steht, **was** eine Ebene ist, hier, **wie**
// ein Text darin gesetzt wird.

// Satz sind die Textangaben, die der Rasterer braucht.
type Satz struct {
	Schrift string

	// Kursivschrift ist der Name der Schrift für kursive Abschnitte —
	// `italic name:` im Stil.
	//
	// **Kursiv wird benannt, nicht errechnet** (`src/data/font.cpp`, Zeilen
	// 100–101, Fassung f9d9356 vom 2021-07-06). Fehlt der Name, bleibt es bei
	// der aufrechten Schrift.
	//
	// # Fett und kursiv stehen am Abschnitt
	//
	// Kursiv gilt nicht für ein Feld, sondern für eine **Stelle im Satz**, und
	// das trägt [Abschnitt].
	//
	// **Fett steht seit dem 2026-08-17 daneben** und ebenfalls an [Abschnitt].
	// MSE setzt dafür ein **Gewicht**, keinen Namen; [Schriftquelle.Suche]
	// nimmt es entgegen und macht daraus wieder einen Namen
	// ([0023](../../docs/entscheidungen/0023-fett-wird-ueber-den-namen-gesucht.md)).
	Kursivschrift string

	Groesse float64

	// Kleinstgroesse ist `scale down to:` — die kleinste Schriftgröße, auf
	// die ein zu langer Text verkleinert werden darf. Null heißt „gar nicht".
	//
	//	double FontTextElement::minScale() const {
	//	  return min(font->size(), font->scale_down_to) / max(0.01, font->size());
	//	}
	//
	// `src/render/text/font.cpp`, Zeilen 51–53, Fassung f9d9356 vom
	// 2021-07-06.
	//
	// **571 Angaben in 174 der 190 Stilpakete** von `M15-Magic-Pack`
	// (#498).
	Kleinstgroesse float64
	Ausrichtung    Ausrichtung

	// Laufweite ist der zusätzliche Abstand zwischen Zeichen, in Punkten.
	Laufweite float64

	// Zeilenabstand als Vielfaches der Schriftgröße; 0 heißt „wie die Schrift".
	Zeilenabstand float64

	// Stauchen sagt, dass ein zu breiter Zug **waagerecht gestaucht** wird —
	// die Ansage `shrink` (oder `stretch`) in der Ausrichtung.
	//
	// **Ein Faktor auf die Breite, nicht auf die Größe.** MSE rechnet
	//
	//	factor = (width - padding_left - padding_right) / layout->width
	//
	// und gibt ihn als waagerechte Zoomstufe an die Zeichenfläche weiter
	// (`src/data/field/text.cpp`, Zeilen 73–81, Fassung f9d9356 vom
	// 2021-07-06). Die Schriftgröße bleibt; ein langer Kartenname steht in
	// voller Höhe und schmal.
	//
	// **Sie steht 676-mal in den 190 Stilen von `M15-Magic-Pack`**, fast immer
	// am Namen und an der Typzeile. Sie ist zugleich das Glied, ohne das der
	// zweistufige Satz aus Phase B unsichtbar bleibt: ein Feld, das wegen
	// breiter Manakosten schmaler wird, ändert nur dann etwas am Bild, wenn
	// sein Text darauf reagiert
	// (#291,
	// #319).
	Stauchen bool

	// Einzeilig ist `multi line: false` des Spielfelds — die Vorgabe von MSE:
	// ein solches Feld bricht weder weich noch erzwungen um, nur an einem
	// harten Umbruch, und läuft sonst über (`TextViewer::prepareLinesAtScale`,
	// `src/render/text/viewer.cpp`, Zeilen 614–621 und 590, Fassung f9d9356
	// vom 2021-07-06; J05, #2269).
	//
	// **Die Angabe steht im Spielpaket und nicht im Stil**; `internal/service`
	// liest sie und gibt sie hierher. Bis dahin stand [Satz.Stauchen] an
	// ihrer Stelle — `shrink` steht meist an denselben Feldern, aber nicht an
	// allen.
	Einzeilig bool

	// Polster ist `padding left/right/top/bottom` — der Rand **innerhalb** des
	// Feldes, der frei bleibt.
	//
	// MSE zieht ihn von der Fläche ab, bevor es setzt: die erste Zeile beginnt
	// bei `padding_top`, umgebrochen wird auf
	// `breite − padding_left − padding_right` (`src/render/text/viewer.cpp`,
	// Zeilen 567, 736–740 und 745–748, Fassung f9d9356 vom 2021-07-06).
	//
	// **`padding top:` steht 279-mal in den 190 Stilen von `M15-Magic-Pack`**,
	// `padding left:` 203-mal.
	Polster Polster

	// Zeilenfaktoren sind `line height soft/hard/line`.
	Zeilenfaktoren Zeilenfaktoren

	// Absatzhoehe ist `paragraph height:` — die **feste** Höhe eines Absatzes.
	//
	// **Sie richtet jeden Absatz für sich aus**, in einem Rechteck dieser
	// Höhe an der Stelle `n · Absatzhoehe`:
	//
	//	alignParagraph(start, last + 1, chars, style,
	//	               RealRect(0, style.padding_top + n*style.paragraph_height,
	//	                        s.width, style.paragraph_height));
	//
	// `src/render/text/viewer.cpp`, Zeilen 749–764, Fassung f9d9356 vom
	// 2021-07-06. Ein Wert von null oder darunter heißt: der ganze Kasten wird
	// als eines ausgerichtet (ebd. Zeile 749).
	//
	// **Das setzt die Fähigkeiten eines Planeswalkers neben ihre
	// Loyalitätssymbole**: `magic-m15-planeswalker` schreibt
	// `paragraph height: 49`, und die Symbole stehen fest bei 333, 383 und
	// 432. Über `M15-Magic-Pack` steht die Angabe 16-mal in 9 Paketen, über
	// das volle Pack 46-mal in 36
	// (#479).
	Absatzhoehe float64

	// Zeilenmaske ist der Paketpfad der Maske des Felds (`mask:`) — aus ihr
	// liest der Umbruch, wie breit jede Zeile ist ([zeilenmaske]). Leer heißt:
	// jede Zeile hat die ganze Breite.
	//
	// **Ein Pfad und kein Bildbaum**: jede `mask:`-Angabe eines Textfelds im
	// Pack ist eine Datei, 52 in 43 Stilen
	// (#2299).
	Zeilenmaske string

	// NurBeiUeberlauf ist die Ansage `overflow`: gestaucht wird nur, wenn der
	// Zug sonst überliefe (`factor < 1.0`).
	//
	// **Ohne sie wird auch gedehnt.** Im Pack steht das zweimal; 676-mal steht
	// `shrink-overflow`.
	NurBeiUeberlauf bool

	// Schatten ist der Schatten der Schrift — durchsichtig, wo der Stil
	// keinen nennt ([Textschatten]).
	Schatten Textschatten
}

// Polster ist der Rand innerhalb eines Feldes.
type Polster struct{ Links, Rechts, Oben, Unten float64 }

// Zeilenfaktoren sagen, wie weit der Satz nach einer Zeile weiterrückt.
//
// **Es hängt daran, wie die Zeile endet:**
//
//	double line_height_multiplier = line.break_after == LineBreak::HARD ? style.line_height_hard
//	                              : line.break_after == LineBreak::LINE ? style.line_height_line
//	                              :                                       style.line_height_soft;
//	line.top += line.line_height * line_height_multiplier;
//
// `src/render/text/viewer.cpp`, Zeilen 678–681, Fassung f9d9356 vom
// 2021-07-06. `magic-m15` schreibt 1,2 / 1,5 / **0,9** — ein umgebrochener
// Regeltext steht dort enger als zwei Absätze.
//
// **Null heißt eins**, nicht „keine Höhe": MSE setzt alle drei auf 1,0, wo
// nichts dasteht (`src/data/field/text.cpp`, Zeilen 62–64), und ein Nullwert
// im Verbund staucht sonst den ganzen Zug auf eine Zeile.
type Zeilenfaktoren struct {
	Weich, Hart, Linie float64

	// WeichMax, HartMax und LinieMax sind `line height … max` — bis dorthin
	// darf ein **zu kurzer** Zug gedehnt werden, damit er den Kasten füllt
	// (`alignParagraph`, `src/render/text/viewer.cpp`, Zeilen 780–815).
	//
	// `magic-m15` schreibt für den Regeltext `hard: 1,2 / max 1,3` und
	// `line: 1,5 / max 1,6`; im Pack stehen `hard max` und `line max` je
	// 271-mal, `soft max` 22-mal.
	//
	// Fassung f9d9356 vom 2021-07-06.
	WeichMax, HartMax, LinieMax float64
}

// oderEins liest den Nullwert als eins.
func oderEins(faktor float64) float64 {
	if faktor <= 0 {
		return 1
	}

	return faktor
}
