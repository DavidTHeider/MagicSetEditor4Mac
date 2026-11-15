package msepaket

// **Die zwei Schriftblöcke eines Feldstils — und warum sie getrennt bleiben.**
//
// `font:` und `symbol font:` tragen dieselben Schlüssel und meinen
// Verschiedenes. Sie stehen hier zusammen, weil sie zusammen gelesen werden:
// `stil.go` sagt, **wo** ein Feld sitzt, diese Datei, **womit** es schreibt.

// Symbolangabe ist der `symbol font:`-Unterblock eines Feldstils.
//
// **Nicht zu verwechseln mit [Symbolschrift]**, dem Inhalt eines
// `.mse-symbol-font`: hier steht, **welche** Schrift ein Feld benutzt, dort,
// **was** in ihr steht.
//
// **Ein eigener Block neben `font:`, mit denselben Schlüsseln.** Beide tragen
// `name:` und `size:`; wer sie in dasselbe Feld liest, zeichnet den Text in
// einer Schrift, die keine Textschrift ist. `casting cost` in
// `magic-m15.mse-style` führt beides untereinander — `MPlantin` für den Text,
// `magic-mana-large` für die Symbole.
//
// **Der Name ist ein Paket, keine Schriftdatei**:
// `magic-mana-large.mse-symbol-font` ist ein Verzeichnis aus PNG-Dateien, je
// eine Datei für ein Symbol. Gedeutet wird es noch nicht (#173).
type Symbolangabe struct {
	Name        string
	Groesse     string
	Ausrichtung string
}

// Schrift ist der `font:`-Unterblock eines Feldstils.
type Schrift struct {
	Name    string
	Groesse string
	Farbe   string

	// KursivName ist die Schrift für kursiven Text — `italic name:`.
	//
	// **Kursiv wird benannt, nicht errechnet.** MSE2 schrägt keine Glyphen: bei
	// gesetztem `FONT_ITALIC` nimmt es die Schrift, deren Namen der Stil angibt
	// (`src/data/font.cpp`, Zeilen 100–101, Fassung f9d9356 vom 2021-07-06).
	// `MPlantin` kursiv ist die Datei `MPlantin-Italic`, eine eigene Schrift
	// mit eigenen Formen.
	//
	// 475-mal in `M15-Magic-Pack` — fast überall dort, wo Flavortext steht.
	KursivName string

	// Kleinstgroesse ist `scale down to:` — die kleinste Schriftgröße, auf
	// die ein zu langer Text verkleinert werden darf.
	//
	//	double FontTextElement::minScale() const {
	//	  return min(font->size(), font->scale_down_to) / max(0.01, font->size());
	//	}
	//
	// `src/render/text/font.cpp`, Zeilen 51–53, Fassung f9d9356 vom
	// 2021-07-06. MSE sucht dann den größten Maßstab zwischen dieser Schranke
	// und 1, bei dem der Text noch in den Kasten passt
	// (`prepareLinesTryScales`, `src/render/text/viewer.cpp`, Zeilen 444–500).
	//
	// **571 Angaben in 174 der 190 Stilpakete** von `M15-Magic-Pack`, im
	// vollen Pack 921 in 348 von 376 — fast jeder Stil verlässt sich darauf
	// (#498).
	//
	// **Leer heißt „nicht verkleinern"**, nicht „auf null": MSEs Vorgabe ist
	// 100000 (`src/data/font.cpp`, Zeile 18), also größer als jede Schrift.
	Kleinstgroesse string

	// Gewicht ist `weight:` und Lage ist `style:` — MSE macht aus `bold` und
	// `italic` die Schalter `FONT_BOLD` und `FONT_ITALIC` der ganzen Schrift
	// (`Font::update`, `src/data/font.cpp`, Zeilen 26–41, Fassung f9d9356 vom
	// 2021-07-06).
	//
	// **Beide dürfen Skripte sein**, wie Name und Farbe: im Pack stehen sie
	// 768- und 16-mal, darunter `{if loy_leveler() then "" else "bold"}`. Was
	// sie ergeben, rechnet `service` — hier steht der Quelltext.
	Gewicht string
	Lage    string

	// SchattenFarbe, SchattenX, SchattenY und SchattenWeichheit sind
	// `shadow color:`, `shadow displacement x:`/`y:` und `shadow blur:`.
	// **Nur die Farbe ist ein Skript** (`Scriptable<Color>`,
	// `src/data/font.hpp`, Zeilen 44–46, Fassung f9d9356 vom 2021-07-06);
	// die drei Zahlen liest MSE als `double`.
	SchattenFarbe     string
	SchattenX         string
	SchattenY         string
	SchattenWeichheit string
}

// schriftAus liest den `font:`-Unterblock.
func schriftAus(knoten *Knoten) Schrift {
	var schrift Schrift

	for _, kind := range knoten.Kinder {
		// **In kanonischer Form verglichen**, wie MSE2 jeden Schlüssel liest
		// (`canonical_name_form`, `src/util/io/reader.cpp` Zeile 217, Fassung
		// f9d9356 vom 2021-07-06): `italic_name:` und `italic name:` sind dasselbe (#1972).
		switch Entkanonisch(kind.Schluessel) {
		// **Auch der Name darf ein eingerückter `script:`-Block sein** —
		// dieselbe Form wie bei `color:` und `left:`
		// (#393).
		case "name":
			schrift.Name = massangabeAus(kind)
		case "italic name":
			schrift.KursivName = massangabeAus(kind)
		case "size":
			schrift.Groesse = kind.Wert
		case "scale down to":
			schrift.Kleinstgroesse = massangabeAus(kind)
		case "weight":
			schrift.Gewicht = massangabeAus(kind)
		case "style":
			schrift.Lage = massangabeAus(kind)
		case "shadow color":
			schrift.SchattenFarbe = massangabeAus(kind)
		case "shadow displacement x":
			schrift.SchattenX = kind.Wert
		case "shadow displacement y":
			schrift.SchattenY = kind.Wert
		case "shadow blur":
			schrift.SchattenWeichheit = kind.Wert
		// **Wie `left:` und `alignment:`**: die Farbe darf ein eingerückter
		// `script:`-Block sein, und wer nur `Wert` liest, bekommt die leere
		// Zeichenkette — also **Schwarz**, ohne dass etwas fehlt. Sechsmal in
		// `M15-Magic-Pack`, alle in `magic-m15-future.mse-style`
		// (#382).
		case "color":
			schrift.Farbe = massangabeAus(kind)
		}
	}

	return schrift
}

// symbolschriftAus liest den `symbol font:`-Unterblock.
//
// **Eigen und nicht über `schriftAus`.** Die Schlüssel überschneiden sich, die
// Bedeutung nicht: `alignment:` steht hier für die Ausrichtung der Symbole im
// Textfluss und hat in `font:` keine Entsprechung.
func symbolschriftAus(knoten *Knoten) Symbolangabe {
	var schrift Symbolangabe

	for _, kind := range knoten.Kinder {
		// **Kanonisch wie beim Feldstil**, auch wenn hier jeder Schlüssel ein
		// Wort ist — ein zweiter, der dazukommt, soll nicht roh verglichen
		// werden (#1972).
		switch Entkanonisch(kind.Schluessel) {
		case "name":
			schrift.Name = kind.Wert
		case "size":
			schrift.Groesse = kind.Wert
		case "alignment":
			schrift.Ausrichtung = kind.Wert
		}
	}

	return schrift
}
