package msepaket

import (
	"strconv"
	"strings"
)

// Symbol ist ein Zeichen einer Symbolschrift samt seinem Bild.
//
// `Code` ist das Zeichen, wie es im Kartentext steht — `W` für Weiß, `U` für
// Blau. `Bild` ist der Paketpfad; gelesen wird es erst beim Rendern, denn eine
// Symbolschrift kann hundert Bilder tragen, und beim Öffnen eines Sets braucht
// man keines davon.
type Symbol struct {
	Code   string
	Bild   string
	Stelle Stelle

	// Bildgroesse überschreibt [Symbolschrift.Bildgroesse] für dieses Symbol —
	// `image font size:` im `symbol:`-Block. Null heißt „nicht angegeben".
	//
	// `magic-mana-small` tut das für seine eigenen Manasymbole
	// (`image font size: 120` gegen 135 der Schrift).
	Bildgroesse float64

	// Bedingung ist das rohe `enabled:` — der Ausdruck, der sagt, ob **diese
	// Fassung** gilt. Leer heißt „gilt immer".
	//
	// # Warum roh und nicht ausgewertet
	//
	// `msepaket` wertet nichts aus (§2). Ob `mana_t() == "older"` zutrifft,
	// weiß erst, wer das Initskript des Stils gelaufen hat — und das ist
	// `internal/service`.
	//
	// # Warum das überhaupt gebraucht wird
	//
	// `magic-mana-large` führt `T` und `Q` je **dreimal** und trägt 54
	// solcher Bedingungen. MSE geht die Liste in Dateireihenfolge durch und
	// nimmt den ersten Eintrag, dessen Bedingung erfüllt ist
	// (`SymbolFont::split`, Zeilen 191–230, Fassung f9d9356 vom 2021-07-06).
	// Ohne die Bedingung gewinnt der erste Eintrag, und das ist bei `T` die
	// Fassung `mana_t_older.png`
	// (#293).
	Bedingung string

	// Textgruppe ist `draw text:` — die **Gruppennummer** des Musters, deren
	// Treffer auf das Bild geschrieben wird.
	//
	//	int draw_text; ///< The index of the captured regex expression to draw,
	//	               ///  or -1 to not draw text
	//
	// `SymbolInFont`, `src/data/symbol_font.cpp`, Zeile 84, Fassung f9d9356
	// vom 2021-07-06.
	//
	// **Minus eins heißt „kein Text"** — das ist MSEs Vorgabe. Null wäre die
	// **ganze** Trefferzeichenkette und damit etwas anderes.
	//
	// **Daran hängt jede Zahl in den Manakosten**: `code: ([XYZ0-9])/W` mit
	// `draw text: 1` zeichnet den Kreis und schreibt die Ziffer hinein
	// (#336).
	Textgruppe int

	// Textschrift ist `text font:` — Name, Größe, Farbe und `max stretch`.
	Textschrift Symboltextschrift

	// Textausrichtung ist `text alignment:`; leer heißt `middle center`.
	Textausrichtung string

	// Die vier Textränder sind Anteile der Schriftgröße und dürfen **negativ**
	// sein: `text margin left: -.25` schiebt den Text über den Bildrand hinaus.
	TextrandLinks  float64
	TextrandRechts float64
	TextrandOben   float64
	TextrandUnten  float64

	// Ausdruck sagt, dass [Symbol.Code] ein **regulärer Ausdruck** ist —
	// `regex: yes`.
	//
	// 32-mal in `magic-mana-large`, darunter der letzte Eintrag der Datei als
	// Auffangausdruck (`[0-9](?!/[WUBRGCSTQ2])|.`). Wer die Angabe nicht liest,
	// vergleicht solch einen Code wörtlich und trifft nie.
	Ausdruck bool
}

// Symbolschrift ist der Inhalt eines `.mse-symbol-font`.
type Symbolschrift struct {
	Name    string
	Symbole []Symbol

	// Bildgroesse ist die Schriftgröße, auf die die Bilder gerechnet sind —
	// `image font size:`.
	//
	// # Warum das gebraucht wird
	//
	// Die Größe, mit der MSE ein Symbol zeichnet, ist **nicht** die
	// Schriftgröße im Quadrat, sondern die **tatsächliche Bildgröße**,
	// verhältnisgleich umgerechnet:
	//
	//	return wxSize(actual_size * (int) (size) / (int) (img_size));
	//
	// Gelesen: `src/data/symbol_font.cpp`, Zeilen 147–153, Fassung f9d9356 vom
	// 2021-07-06. Bei `magic-mana-small` ist es 135
	// (#330).
	//
	// **Null heißt „nicht angegeben"** — eine geratene Zahl sähe aus wie eine
	// gelesene.
	Bildgroesse float64

	// AbstandWaagerecht und AbstandSenkrecht sind `horizontal space:` und
	// `vertical space:` — der Abstand, der **je Symbol** zu seiner Größe
	// hinzukommt (`add_diagonal(…, spacing)`). **Ohne Angabe je 1**:
	// `spacing(1,1)`, `SymbolFont::SymbolFont`, `src/data/symbol_font.cpp`,
	// Zeile 28, Fassung f9d9356 vom 2021-07-06 — 75 der 76 Symbolschriften
	// des vollen Packs nennen keinen senkrechten Abstand (#2300).
	AbstandWaagerecht float64
	AbstandSenkrecht  float64

	// Einfuegemenue ist `insert symbol menu:` — was eine Oberfläche zum
	// Einfügen anbietet ([Menueeintrag]). `nil`, wenn die Schrift keins hat.
	Einfuegemenue []Menueeintrag
}

// Symbol sucht ein Zeichen. Der zweite Rückgabewert sagt, ob es das gibt.
func (s Symbolschrift) Symbol(code string) (Symbol, bool) {
	for i := range s.Symbole {
		symbol := s.Symbole[i]

		if symbol.Code == code {
			return symbol, true
		}
	}

	return Symbol{}, false
}

// Symbolmodell liest ein Symbolschriftpaket.
//
// Ein Symbol ohne `code` oder ohne `image` ist ein Befund, kein Abbruch: eine
// Schrift mit neunundneunzig brauchbaren Zeichen und einem kaputten ist neunmal
// nützlicher als gar keine.
func (p *Paket) Symbolmodell() (Symbolschrift, []Befund) {
	schrift := Symbolschrift{
		Name:              p.KurzerName,
		Bildgroesse:       zahlOderNull(p.Kopf.Ersten("image font size")),
		AbstandWaagerecht: zahlOder(p.Kopf.Ersten("horizontal space"), 1),
		AbstandSenkrecht:  zahlOder(p.Kopf.Ersten("vertical space"), 1),
		Einfuegemenue:     einfuegemenue(p.Kopf.Ersten("insert symbol menu")),
	}

	var befunde []Befund

	for _, knoten := range p.Kopf.Alle("symbol") {
		code := wertOderLeer(knoten.Ersten("code"))
		// **Auch als eingerückter `script:`-Block** — dieselbe Form wie bei
		// `left:` und `font: color:`, und dieselbe Falle: wer nur den Wert
		// liest, meldet „nennt kein Bild".
		//
		// `magic-mana-small.mse-symbol-font` schreibt seine zehn
		// Dreifach-Hybriden so (Zeilen 300–408); über die kanonischen Sets
		// waren das **2.500 von 2.794** Befunden, zehn Meldungen auf **jeder**
		// Karte (#405).
		bild := massangabeAus(knoten.Ersten("image"))

		switch {
		case code == "":
			befunde = append(befunde, Befund{
				Stelle:  knoten.Stelle,
				Meldung: "ein Symbol ohne `code` — es lässt sich keinem Zeichen zuordnen",
			})
		case bild == "":
			befunde = append(befunde, Befund{
				Stelle:  knoten.Stelle,
				Meldung: "das Symbol " + code + " nennt kein Bild",
			})
		default:
			schrift.Symbole = append(schrift.Symbole, symbolMitText(knoten, code, bild))
		}
	}

	return schrift, befunde
}

// jaOderNein liest eine Ja-Angabe — MSE schreibt `yes`, `true` oder `1`.
//
// **Alles andere ist nein**, auch ein leerer Knoten: die Angabe ist in jeder
// gelesenen Datei die Ausnahme, und eine unbekannte Schreibweise als „ja" zu
// deuten machte aus einem gewöhnlichen Code stillschweigend einen Ausdruck.
func jaOderNein(knoten *Knoten) bool {
	if knoten == nil {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(knoten.Wert)) {
	case "yes", "true", "1":
		return true
	}

	return false
}

// zahlOderNull liest eine Kommazahl aus einem Knoten; fehlt er, ist es null.
//
// **Null heißt „nicht angegeben"**, und das ist hier eine Auskunft: der
// Aufrufer entscheidet dann, was gilt. Eine geratene Zahl sähe aus wie eine
// gelesene.
func zahlOderNull(knoten *Knoten) float64 {
	if knoten == nil {
		return 0
	}

	zahl, err := strconv.ParseFloat(strings.TrimSpace(knoten.Wert), 64)
	if err != nil {
		return 0
	}

	return zahl
}
