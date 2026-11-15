package render

// **Woraus ein Zug besteht — Abschnitt, Rand, Schnitt, Stück.**
//
// Getrennt von `abschnitte.go`, das ihn **setzt**: dort steht, wie aus
// diesen Angaben Zeilen werden, hier, was die Angaben bedeuten.

// Abschnitt ist ein Stück Text mit eigenem Schnitt.
//
// # Warum ein Textzug Abschnitte braucht
//
// Kursiv steht **mitten** im Regeltext: `<i-flavor>` beginnt nach dem
// Regelabsatz und läuft bis zum Ende. Eine flache Zeichenkette hat keine
// Stelle, an der ein Schnitt wechseln könnte — es ist derselbe Umbau wie beim
// Bildbaum (#270), eine Ebene höher
// (#312).
type Abschnitt struct {
	Text   string
	Kursiv bool

	// Fett ist der Schnitt aus `<b>`.
	//
	// **Ein eigenes Feld neben [Abschnitt.Kursiv]**, denn die beiden sind bei
	// MSE nicht dasselbe: kursiv ist ein Name, fett ein Gewicht. Zusammen
	// ergeben sie einen dritten Schnitt (Vermerk 0023).
	Fett bool

	// Weich sagt, dass dieser Abschnitt bei der **Ausrichtung** nicht zählt und
	// **nicht gezeichnet** wird — `<soft>` und `<sep-soft>`.
	//
	//	bool soft : 1; ///< Is this a 'soft' character? soft characters are
	//	               ///  ignored for alignment
	//
	// `CharInfo`, `src/render/text/element.hpp`, Zeilen 36–47, Fassung f9d9356
	// vom 2021-07-06. Die Breite einer Zeile endet am letzten **nicht
	// weichen** Zeichen (`viewer.cpp`, Zeilen 17–35 und 610).
	//
	// **Wozu:** `sub_type_filter` hängt an die Typzeile ein
	// `"<soft>" + subtype_separator + "</soft>"` — einen Trenner für den Fall,
	// dass noch etwas folgt. Zählte er mit, stünde jede mittig gesetzte
	// Typzeile um seine halbe Breite zu weit links
	// (#345).
	Weich bool

	// Symbole sind die Bilder, wenn dieser Abschnitt in **Symbolen** steht
	// statt in Buchstaben — `<sym>T</sym>` mitten im Regeltext.
	//
	// **Bilder und keine Codes.** Welcher Code welches Bild meint, weiß die
	// Symbolschrift, und die kennt `internal/service` (§2). Was hier ankommt,
	// ist das Ergebnis dieser Zuordnung — derselbe Schnitt wie beim Bildbaum.
	Symbole []Symbol

	// Symbolmasse sind die Maße, mit denen diese Symbole gesetzt werden.
	//
	// **Angaben und kein Maß.** Wie groß ein Symbol wird, hängt an seinem Bild;
	// gemessen wird deshalb hier und nicht im Planbauer (Vermerk 0022).
	Symbolmasse Symbolmasse

	// Blockende sagt, dass hinter diesem Abschnitt ein **Block** endet — die
	// Auszeichnung `<line>`.
	//
	// MSE gibt dem Element den Umbruch `LineBreak::LINE`, und im Satz endet
	// daran ein Block (`src/render/text/viewer.cpp`, Zeilen 380–384, Fassung
	// f9d9356 vom 2021-07-06). Das Feld `text` von `magic.mse-game` setzt damit
	// Regeltext und Flavortext auseinander
	// (#338).
	Blockende bool

	// WeicheZeile sagt, dass ein `\n` in diesem Abschnitt **weich** umbricht —
	// die Auszeichnung `<soft-line>`.
	//
	//	LineBreak line_break = line > 0      ? LineBreak::LINE :
	//	                       soft_line > 0 ? LineBreak::SOFT : LineBreak::HARD;
	//
	// `src/render/text/element.cpp`, Zeilen 235–236, Fassung f9d9356 vom
	// 2021-07-06. Der Umbruch fällt, aber er bekommt den Abstand eines weichen
	// (`line height soft` statt `hard`) — und er schließt **keinen Absatz**:
	//
	//	void addParagraphs(const String& text, size_t start, size_t end) {
	//	  if (line == 0 && soft_line > 0) return;
	//
	// ebd. Zeilen 273–274.
	//
	// **Ein eigenes Feld neben [Abschnitt.Weich]**, und die beiden sind nicht
	// dasselbe: `<soft>` betrifft die **Ausrichtung**, `<soft-line>` den
	// **Umbruch**. MSE prüft `<soft-line` deshalb vor `<soft`, und dieses
	// Programm tut es ebenso
	// (#431).
	WeicheZeile bool

	// Margen sind die Ränder des Absatzes, in dem dieser Abschnitt **beginnt**
	// — die Auszeichnung `<margin:links:rechts:oben>`. Nil heißt „keine".
	//
	//	paragraphs.back().margin_left  = m.left;
	//	paragraphs.back().margin_right = m.right;
	//	paragraphs.back().margin_top   = m.top;
	//
	// `src/render/text/element.cpp`, Zeilen 192–194, Fassung f9d9356 vom
	// 2021-07-06. Angewandt wird es beim Absatzwechsel (`viewer.cpp`, Zeilen
	// 692–694): der linke Rand verschiebt die Zeile **und** verkürzt sie, der
	// obere kommt auf den Absatzanfang.
	//
	// **Am Absatz und nicht am Abschnitt.** MSE führt die Ränder an
	// `paragraphs`, nicht am Element; ein `<margin>` mitten in einem Absatz
	// setzt den Rand des **ganzen** Absatzes, in dem es steht.
	Margen *Margen

	// ImRand sagt, dass die Zeichen dieses Abschnitts zum **linken Rand**
	// ihres Absatzes zählen — der Inhalt von `<bullet>`.
	//
	//	if (i < elements.paragraphs[i_para].margin_end_char) {
	//	  line.margin_left += c.size.width; // character in left margin
	//	}
	//
	// `src/render/text/viewer.cpp`, Zeilen 611–613, Fassung f9d9356 vom
	// 2021-07-06. Gezeichnet werden sie, wo sie stehen; die **folgenden**
	// Zeilen des Absatzes beginnen hinter ihnen — ein hängender Einzug.
	// `magic.mse-game` klammert damit jede Zeile, die mit „• " beginnt
	// (`script`, Zeilen 1943–1946).
	//
	// # Wo das von MSE abweicht
	//
	// MSE führt eine **Stelle** (`margin_end_char`), nicht eine Eigenschaft
	// des Abschnitts: alles vom Absatzanfang bis zum schließenden `</bullet`
	// zählt zum Rand, auch was **vor** dem `<bullet>` stünde. Hier zählt nur,
	// was **darin** steht.
	//
	// Der Unterschied ist im Pack keiner: die Regel, die `<bullet>` erzeugt,
	// verankert es am Zeilenanfang (`match: "^(• )([^\n]+\n?)"`). Wer eine
	// Vorlage findet, die etwas davorsetzt, hat den Fall gefunden, für den
	// diese Vereinfachung nicht trägt
	// (#431).
	ImRand bool
}

// Margen sind die drei Ränder eines Absatzes.
//
// **Drei und nicht vier.** MSE kennt `left`, `right` und `top`; einen unteren
// gibt es nicht (`Margins`, `src/render/text/element.cpp`, Zeile 44).
//
// Fassung f9d9356 vom 2021-07-06.
type Margen struct {
	Links  float64
	Rechts float64
	Oben   float64
}

// istSymbolisch sagt, ob ein Abschnitt in Symbolen steht.
func (a Abschnitt) istSymbolisch() bool { return len(a.Symbole) > 0 }

// Schriftsatz sind die Schnitte, in denen ein Feld gesetzt wird.
//
// **Kursiv wird benannt, nicht errechnet.** MSE2 schrägt keine Glyphen: bei
// gesetztem `FONT_ITALIC` nimmt es die Schrift, deren Namen der Stil unter
// `italic name:` angibt (`src/data/font.cpp`, Zeilen 100–101, Fassung f9d9356
// vom 2021-07-06). Fehlt die Angabe, bleibt es bei der aufrechten — die
// Bedingung dort verlangt ausdrücklich `!italic_name().empty()`.
type Schriftsatz struct {
	Aufrecht Schrift
	Kursiv   Schrift

	// Fett und KursivFett sind die fetten Schnitte.
	//
	// **Vier Schnitte und nicht zwei**, weil MSE das fette Gewicht auch auf die
	// Kursive legt: bei `<b><i>` nimmt es die Schrift aus `italic name:` **und**
	// das Gewicht (`src/data/font.cpp`, Zeilen 100–101).
	//
	// Fassung f9d9356 vom 2021-07-06.
	Fett       Schrift
	KursivFett Schrift

	// Ersatz gibt eine Schrift, die ein Zeichen führt, das die gewählte nicht
	// hat — **nil heißt „kein Ausweich"**.
	//
	// # Warum es das gibt
	//
	// **MSE weicht aus, weil wxWidgets es tut.** Ein Stil nennt eine Schrift;
	// fehlt darin ein Zeichen, sucht die Plattform eine andere, die es führt —
	// unter macOS Core Text, unter Windows GDI. Dieses Programm rastert selbst
	// und hatte diesen Weg nicht: über die 23.168 kanonischen Karten fehlte
	// **ein** Zeichen, `ǵ` (U+01F5), und es fehlte 159-mal
	// (#515).
	//
	// # Warum eine Funktion und keine Liste
	//
	// **Weil die Antwort vom Zeichen abhängt.** Welche Schrift `ǵ` führt, weiß
	// nur, wer alle geöffnet hat — und das ist `internal/schriften`, nicht der
	// Rasterer (§2). Eine Liste hier wäre eine Vermutung darüber, welche
	// Schriften dieser Rechner hat.
	//
	// **Der Preis steht daneben**: die Ersatzschrift hat andere Formen und
	// andere Maße. Ein ausgewichenes Zeichen sieht anders aus als bei MSE, wo
	// die Plattform wählt — aber es steht da, und das ist der Unterschied, um
	// den es geht (§8).
	Ersatz Ersatzschrift
}

// Ersatzschrift gibt eine Schrift, die ein Zeichen führt — siehe
// [Schriftsatz.Ersatz].
type Ersatzschrift func(zeichen rune) (Glyphenschrift, bool)

// fuer gibt den Schnitt eines Abschnitts.
//
// **Was fehlt, fällt auf den nächstliegenden zurück** — dieselbe
// Bestenfalls-Arbeit wie bei der Kursiven: ein aufrechter Text ist eine Karte,
// ein fehlender ist keine (§8).
func (s Schriftsatz) fuer(kursiv, fett bool) Schrift {
	for _, schnitt := range s.schnitte(kursiv, fett) {
		if schnitt != nil {
			return schnitt
		}
	}

	return s.Aufrecht
}

// schnitte gibt die Schnitte eines Abschnitts in der Reihenfolge ihres Vorrangs.
func (s Schriftsatz) schnitte(kursiv, fett bool) []Schrift {
	switch {
	case kursiv && fett:
		return []Schrift{s.KursivFett, s.Kursiv, s.Fett}

	case kursiv:
		return []Schrift{s.Kursiv}

	case fett:
		return []Schrift{s.Fett}
	}

	return nil
}

// Stueck ist ein Abschnitt, wie er in einer Zeile steht.
//
// **Mit eigener Stelle und eigener Breite**, denn nur so kann der Rasterer ihn
// mit seiner eigenen Schrift zeichnen.
type Stueck struct {
	Text   string
	Kursiv bool
	Fett   bool

	// Symbole und Symbolmasse gelten für einen symbolischen Abschnitt —
	// siehe [Abschnitt.Symbole].
	Symbole     []Symbol
	Symbolmasse Symbolmasse

	// Blockende — siehe [Abschnitt.Blockende].
	Blockende bool

	// Weich — siehe [Abschnitt.Weich].
	Weich bool

	// ImRand — siehe [Abschnitt.ImRand].
	ImRand bool

	// Links ist die linke Kante in Vorlagenpunkten, Breite die gemessene.
	Links, Breite float64

	// Lauf benennt den Abschnitt, aus dem das Stück stammt — bei MSE das
	// Element, das eine Zeile in einem Zug zeichnet ([Rasterer.zeichneZeilen],
	// #2302). Null heißt: keiner.
	Lauf int
}
