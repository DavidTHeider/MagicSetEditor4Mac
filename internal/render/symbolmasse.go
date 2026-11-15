package render

import "math"

// **Wie groß ein Symbol ist, weiß erst, wer sein Bild öffnet.**
//
// Deshalb steht diese Rechnung hier und nicht im Planbauer — ausführlich in
// `docs/entscheidungen/0022-der-rasterer-misst-die-symbole.md`.

// Symbol ist ein Symbol im Zeichenplan.
type Symbol struct {
	// Bild ist der **Bildbaum** dieses Symbols — eine Datei oder eine
	// gerechnete Verknüpfung.
	//
	// **Ein Baum und kein Pfad**, weil `SymbolInFont` sein Bild bei MSE2 als
	// `ScriptableImage` hält (`src/data/symbol_font.cpp`, Zeile 93, Fassung
	// f9d9356 vom 2021-07-06). Über `Full-Magic-Pack` sind 2.435 von 8.110
	// `image:`-Zeilen der Symbolschriften Skripte
	// (#422).
	Bild Bildbaum

	// Bildgroesse ist die `image font size:` **dieses** Symbols und überschreibt
	// die der Schrift. Null heißt „keine eigene".
	//
	// `magic-mana-large` nennt 135 und schreibt an einzelnen Symbolen 120 oder
	// 145 dazu (`SymbolInFont::SymbolInFont`, Zeilen 111–113, Fassung f9d9356
	// vom 2021-07-06).
	Bildgroesse float64

	// Text ist, was **auf** das Bild geschrieben wird — die Ziffer in einem
	// Manakreis.
	//
	//	if (text.empty() || !sym.text_font) return;
	//
	// `SymbolFont::drawSymbol`, `src/data/symbol_font.cpp`, Zeile 297, Fassung
	// f9d9356 vom 2021-07-06. Leer heißt „kein Text"
	// (#336).
	Text string

	// Textsatz sagt, **wie** dieser Text gezeichnet wird.
	Textsatz Symboltextsatz
}

// Symboltextsatz sind die Textangaben eines Symbols — `text font:` und die
// vier `text margin`-Ränder.
type Symboltextsatz struct {
	Schrift string
	Farbe   Farbe

	// Groesse ist ein **Anteil** der Symbolgröße: `double size = font_size *
	// text_font->size`.
	Groesse float64

	// Dehnung ist `max stretch` — so weit darf der Text gestaucht werden,
	// bevor MSE die Schriftgröße verkleinert.
	Dehnung float64

	// Ausrichtung ist `text alignment:`.
	//
	// **Der Nullwert ist `middle center`**, so wie bei MSE
	// (`text_alignment(ALIGN_MIDDLE_CENTER)`) — und anders als bei einer
	// Textebene, deren Vorgabe links oben ist. Wer das gleichsetzte, schöbe
	// jede Ziffer in die linke obere Ecke ihres Kreises.
	Ausrichtung Ausrichtung

	// Die vier Ränder sind Anteile der Symbolgröße und dürfen **negativ**
	// sein — `text margin left: -.25` schiebt den Text über den Bildrand.
	RandLinks  float64
	RandRechts float64
	RandOben   float64
	RandUnten  float64
}

// Symbolmasse sind die Maße, mit denen ein Symbolzug gesetzt wird.
//
// Sie kommen aus zwei Dateien: `Groesse` und `Ausrichtung` aus dem
// `symbol font:`-Block des **Feldstils**, der Rest aus dem
// `.mse-symbol-font`-Paket.
type Symbolmasse struct {
	// Groesse ist `symbol font: size:` — die Schriftgröße, auf die gerechnet wird.
	Groesse float64

	// Bildgroesse ist `image font size:` der Schrift; ein Symbol darf sie
	// überschreiben. Null heißt „nicht angegeben", und dann bleibt es beim
	// Quadrat der Größe — siehe Vermerk 0022.
	Bildgroesse float64

	// AbstandWaagerecht und AbstandSenkrecht sind `horizontal space:` und
	// `vertical space:`. Sie kommen **je Symbol** zu dessen Größe hinzu
	// (`add_diagonal(sym.symbol->size(…), spacing)`, Zeilen 384–401).
	AbstandWaagerecht, AbstandSenkrecht float64

	// Ausrichtung ist die `alignment:` des Symbolschriftblocks — sie sagt, wo
	// das **Bild in seinem Kasten** sitzt, nicht wo der Zug im Feld sitzt.
	//
	// **Ohne Angabe mittig**, und zwar in beiden Richtungen: MSE setzt
	// `ALIGN_MIDDLE_CENTER` als Vorgabe (`SymbolFontRef::SymbolFontRef`, Zeile
	// 556). Der Nullwert dieses Feldes ist die leere Zeichenkette und damit
	// keine der neun Ausrichtungen; [Symbolkasten.bildstelle] liest ihn als
	// „mittig".
	Ausrichtung Ausrichtung
}

// Symbolkasten ist der Platz eines Symbols und das Bild darin.
//
// **Zwei Maße und nicht eines.** Der Kasten trägt den Abstand, das Bild nicht;
// wer sie zusammenwürfe, zeichnete das Bild um den Abstand zu groß.
type Symbolkasten struct {
	// Breite und Hoehe des Kastens, Abstand eingerechnet.
	Breite, Hoehe float64

	// Bildbreite und Bildhoehe des Bildes darin.
	Bildbreite, Bildhoehe float64
}

// kasten misst ein Symbol gegen sein Bild.
//
// **Der Rückfall ist das Quadrat**, und zwar in drei Fällen: keine
// `image font size`, keine Bildquelle, kein lesbares Bild. Eine Null dort
// setzte das Symbol auf keine Breite, und der Rest der Zeile rutschte
// nach links — der Befund beim Zeichnen zeigte dann auf eine Stelle, an der
// nichts mehr steht.
func (m Symbolmasse) kasten(symbol *Symbol, quelle Bildquelle) Symbolkasten {
	quadrat := Symbolkasten{
		Breite: m.Groesse + m.AbstandWaagerecht, Hoehe: m.Groesse + m.AbstandSenkrecht,
		Bildbreite: m.Groesse, Bildhoehe: m.Groesse,
	}

	bildgroesse := m.Bildgroesse
	if symbol.Bildgroesse > 0 {
		bildgroesse = symbol.Bildgroesse
	}

	if bildgroesse <= 0 || quelle == nil {
		return quadrat
	}

	// **Gemessen wird am ersten Blatt, nicht am gerechneten Bild.** Diese
	// Stufe hat nur eine Bildquelle; den Baum auszuführen kann allein der
	// Rasterer.
	//
	// **Was das kostet** (§10): ein Skript, das die Größe **ändert**, wird
	// hier falsch gemessen — `crop` und `enlarge` tun das. Die Klasse, um
	// derentwillen der Baum hier steht, tut es nicht: `recolor_image` gibt ein
	// Bild derselben Größe zurück.
	blaetter := symbol.Bild.Blaetter()
	if len(blaetter) == 0 {
		return quadrat
	}

	bild, err := quelle.Bild(blaetter[0])
	if err != nil {
		return quadrat
	}

	masse := bild.Bounds()

	// **Der Kasten in ganzen Punkten**, wie `wxSize` rechnet:
	//
	//	return wxSize(actual_size * (int) (size) / (int) (img_size));
	//
	// `SymbolInFont::size`, `src/data/symbol_font.cpp`, Zeilen 147–153 —
	// die Schriftgröße abgeschnitten, die Division ganzzahlig.
	//
	// **Das Bild darin auch, aber ohne die erste Kürzung**: `getImage` legt es
	// in `(int) (actual_size.GetWidth() * size / img_size)` an (ebd. Zeilen
	// 115–127). Bei 13,8 Punkten ist das Bild deshalb größer als sein Kasten.
	// Beides Fassung f9d9356 vom 2021-07-06; gerechnet wird in Stilpunkten, so
	// wie MSE2 beim Zoom 1 (#2299).
	kastenmass := func(seite int) float64 {
		return float64(seite * int(m.Groesse) / max(int(bildgroesse), 1))
	}
	bildmass := func(seite int) float64 {
		return math.Floor(float64(seite) * m.Groesse / bildgroesse)
	}

	return Symbolkasten{
		Breite:     kastenmass(masse.Dx()) + m.AbstandWaagerecht,
		Hoehe:      kastenmass(masse.Dy()) + m.AbstandSenkrecht,
		Bildbreite: bildmass(masse.Dx()), Bildhoehe: bildmass(masse.Dy()),
	}
}

// bildstelle gibt die linke obere Ecke des Bildes in einem Kasten.
//
// Der Kasten ist so breit wie das Symbol samt Abstand und so hoch wie die
// **Zeile** — bei MSE erhält `split_left` die Höhe des Zeilenrechtecks, und
// `align_in_rect` legt das Bild darin ab (`symbol_font.cpp`, Zeilen 279–294).
func (k Symbolkasten) bildstelle(
	ausrichtung Ausrichtung, links, oben, zeilenhoehe float64,
) (bildLinks, bildOben float64) {
	mittig, rechts := waagerecht(ausrichtung)
	bildLinks = links + symbolversatz(mittig, rechts, k.Breite, k.Bildbreite)

	mittig, unten := senkrecht(ausrichtung)
	bildOben = oben + symbolversatz(mittig, unten, zeilenhoehe, k.Bildhoehe)

	return bildLinks, bildOben
}

// symbolversatz rückt ein Maß in einem größeren ein.
func symbolversatz(mittig, amEnde bool, kasten, inhalt float64) float64 {
	switch {
	case amEnde:
		return kasten - inhalt

	case mittig:
		return (kasten - inhalt) / 2
	}

	return 0
}

// waagerecht und senkrecht lesen eine Ausrichtung.
//
// **Der Nullwert ist mittig, nicht links oben** — das ist der Unterschied zur
// Feldausrichtung und der Grund, warum hier nicht [waagerechterVersatz] steht.
func waagerecht(ausrichtung Ausrichtung) (mittig, rechts bool) {
	switch ausrichtung {
	case LinksOben, LinksMitte, LinksUnten:
		return false, false

	case RechtsOben, RechtsMitte, RechtsUnten:
		return false, true
	}

	return true, false
}

func senkrecht(ausrichtung Ausrichtung) (mittig, unten bool) {
	switch ausrichtung {
	case LinksOben, MitteOben, RechtsOben:
		return false, false

	case LinksUnten, MitteUnten, RechtsUnten:
		return false, true
	}

	return true, false
}
