package msescript

import "strconv"

// wert liest das, was ohne Nachbarn für sich steht.
func (z *zerteiler) wert() (*Knoten, error) {
	jetzt := z.jetzt()

	if z.amEnde() {
		return nil, z.hier(SchluesselFehltHier, "ein Wert")
	}

	// **`if` und `for` sind Ausdrücke, keine Satzanfänge** (#214). In
	// `magic.mse-game` stehen beide rechts von einem `+`:
	//
	//	"…</sym> or 2 life. " +
	//	if multiple then "For each …"
	//	else             "If life was paid, …"
	//
	//	e1 := e1 + for x from 0 to length(of)-1 do …
	//
	// Für `case … of` gilt dasselbe — es ist die dritte Form, die einen Wert
	// ergibt.
	//
	// **Sie reichen bis ans Ende** — `bedingung` liest den ganzen `if` samt
	// `else` und die ganze Schleife samt `do`. Anders gälte der `else`-Zweig
	// nur für sich, und die Vorlage schriebe die Hälfte.
	if z.istWort("if", "for", "case") {
		return z.bedingung()
	}

	switch jetzt.Art {
	case ArtZahl:
		z.stelle++

		return blatt(FormZahl, jetzt), nil

	case ArtZeichenkette:
		z.stelle++

		return z.textbaustein(jetzt)

	case ArtBezeichner:
		z.stelle++

		return blatt(FormName, jetzt), nil

	case ArtSchluesselwort:
		if jetzt.Text == "nil" {
			z.stelle++

			return &Knoten{Form: FormNichts, Zeile: jetzt.Zeile, Spalte: jetzt.Spalte}, nil
		}

		// **Die Wörter des Schleifenkopfs sind auch Namen** (#214).
		//
		// In `magic.mse-game` sind `in` und `from` die Argumentnamen eines
		// Blocks — `ar_position(of:… in:…)`, `ar_position(of:of[x], in:from)` —
		// und werden darin als `length(in)` und `in:from` gelesen. Gezählt:
		// `to:` steht 111-mal, `in:` 19-mal als Argumentname im Spielpaket und
		// seinen Modulen.
		//
		// **Es folgt notwendig** aus der Regel darüber, dass ein Schlüsselwort
		// ein Argumentname sein darf: wer es nicht auch als Wert liest, hat
		// eine Sprache, in der sich ein Argument übergeben, aber nicht benutzen
		// lässt.
		//
		// **Nur diese drei, und der Grund ist der Bau des Zerteilers.** `if`,
		// `for`, `not`, `nil` und die Verknüpfungen **beginnen** eine Form; wer
		// sie als Namen zuließe, machte `if x then y` zweideutig. `in`, `from`
		// und `to` beginnen keine — sie stehen ausschließlich an festen Stellen
		// eines Schleifenkopfs, und dort sucht `schleife` sie ausdrücklich.
		if istSchleifenwort(jetzt.Text) {
			z.stelle++

			return blatt(FormName, jetzt), nil
		}

	case ArtOperator:
		return z.klammer()
	}

	return nil, z.hier(SchluesselFehltHier, "ein Wert")
}

// istSchleifenwort sagt, ob ein Schlüsselwort auch ein Name sein darf.
//
// **`of` gehört dazu** (#214): es steht 48-mal als Argumentname und 11-mal als
// Wert (`ar_position(of:of[x], …)`) und trägt seine Form — `case … of` — nur in
// deren Mitte.
//
// **`each` und `do` gehören nicht dazu**: für sie gibt es keinen Beleg, und was
// nicht belegt ist, wird nicht gebaut (§11).
func istSchleifenwort(text string) bool {
	return text == "in" || text == "from" || text == "to" || text == "of"
}

func blatt(form Form, merkmal Merkmal) *Knoten {
	return &Knoten{Form: form, Text: merkmal.Text, Zeile: merkmal.Zeile, Spalte: merkmal.Spalte}
}

// klammer liest `( … )` als Folge und `{ … }` als Wert.
//
// **Der Unterschied ist der Kern der Sprache** (`docs/FORMATE.md`): die runde
// Klammer fasst zusammen und ergibt sofort einen Wert, die geschweifte ergibt
// eine Funktion, die erst beim Aufruf rechnet. Wer beide gleich behandelt,
// wertet jede Vorlage zu früh aus.
func (z *zerteiler) klammer() (*Knoten, error) {
	beginn := z.jetzt()

	switch beginn.Text {
	case "(":
		// **Ohne eigenen Knoten.** Die Gruppierung steht schon in der Form des
		// Baums: `(a or b) and c` ergibt ein `and` über einem `or`, und ein
		// Klammerknoten darüber sagte dasselbe ein zweites Mal. Der erste
		// Anlauf setzte ihn und bekam für eine Folge in Klammern zwei
		// ineinandergeschachtelte Folgen.
		return z.eingefasst(beginn, ")", "")

	case "{":
		return z.eingefasst(beginn, "}", FormKlammer)

	case "[":
		return z.abbildung(beginn)
	}

	return nil, z.hier(SchluesselFehltHier, "ein Wert")
}

func (z *zerteiler) eingefasst(beginn Merkmal, schluss string, form Form) (*Knoten, error) {
	z.stelle++

	inhalt, err := z.folge([]string{schluss})
	if err != nil {
		return nil, err
	}

	if !z.istOperator(schluss) {
		return nil, klammerOffen(beginn, schluss)
	}

	z.stelle++

	if form == "" {
		return inhalt, nil
	}

	return &Knoten{
		Form: form, Kinder: []*Knoten{inhalt},
		Zeile: beginn.Zeile, Spalte: beginn.Spalte,
	}, nil
}

// argumente liest `( a, name: b )` — benannt und nach Stellung gemischt.
//
// Die Namen wandern in ein eigenes Feld statt in den Baum: ein benanntes
// Argument ist kein Ausdruck, und es als einen zu führen zwänge jeden Leser,
// beim Auswerten wieder zu unterscheiden.
func (z *zerteiler) argumente(ziel *Knoten, form Form) (*Knoten, error) {
	beginn := z.jetzt()

	z.stelle++

	knoten := &Knoten{
		Form: form, Kinder: []*Knoten{ziel}, Namen: []string{""},
		Zeile: ziel.Zeile, Spalte: ziel.Spalte,
	}

	for {
		z.umbruecheUeberspringen()

		if z.istOperator(")") {
			z.stelle++

			return knoten, nil
		}

		// **Am Quellende ist die Klammer der Befund, nicht der fehlende
		// Wert.** MSEs Argumentschleife hält bei `TOK_EOF` an und erwartet
		// erst danach die schließende Klammer:
		//
		//	while (input.peek() != _(")") && input.peek() != TOK_EOF) { … }
		//	expectToken(input, _(")"), &token);
		//
		// `src/script/parser.cpp`, `parseCallArguments` und Zeile 825,
		// Fassung f9d9356 vom 2021-07-06. Am Orakel gibt `f(1,` deshalb
		// `(1): Expected closing ')' for this '('` und nicht
		// `Expected 'expression'`.
		if z.amEnde() {
			return nil, klammerOffen(beginn, ")")
		}

		name := ""
		if bezeichner, ist := z.benanntesArgument(); ist {
			name = bezeichner
		}

		wert, err := z.ausdruck()
		if err != nil {
			return nil, err
		}

		knoten.Kinder = append(knoten.Kinder, wert)
		knoten.Namen = append(knoten.Namen, name)

		z.umbruecheUeberspringen()

		if z.istOperator(",") {
			z.stelle++
		}
	}
}

// benanntesArgument erkennt `name:` und geht darüber hinweg.
//
// Zwei Merkmale weit vorausgeschaut, weil ein Name auch für sich stehen darf:
// in `f(order)` ist `order` ein Ausdruck, in `f(order: "XY")` eine Beschriftung.
//
// **Eine Zeichenkette und eine Zahl zählen auch als Name.**
// `magic.mse-game/language` schreibt `["Deutsch": [ … ]]`, sein Initskript
// `[0: "ноль", 1: "один"]` (#195, #214).
//
// **Die Regel ist nicht „Zeichenketten sind erlaubt", sondern: was fest in der
// Datei steht, ist ein Schlüssel** — es lässt sich lesen, ohne etwas
// auszurechnen. Ein **Ausdruck** ist weiterhin keiner.
func (z *zerteiler) benanntesArgument() (string, bool) {
	// **Über `jetzt` und nicht über den Index.** Das Quellende prüfen beide
	// Aufrufer selbst, um dort die offene Klammer zu melden; eine eigene
	// Prüfung hier wäre unbetretbar. `jetzt` gibt am Ende das
	// Endmerkmal, dessen Art keine der vier unten ist; das Ergebnis bleibt
	// dasselbe, und ein Zugriff daneben ist ausgeschlossen statt bloß
	// unwahrscheinlich.
	//
	// **Auch ein Schlüsselwort**: `number_of_items(in: colors)` steht 25-mal im
	// bezogenen Pack, und `in` ist sonst der Kopf einer Schleife. Zweideutig
	// ist es nicht — als Name steht es nur dort, wo unmittelbar ein `:` folgt,
	// und diese Form gibt es in keiner Schleife (#214).
	if art := z.jetzt().Art; art != ArtBezeichner &&
		art != ArtZeichenkette && art != ArtZahl && art != ArtSchluesselwort {
		return "", false
	}

	if z.stelle+1 >= len(z.merkmale) {
		return "", false
	}

	naechstes := z.merkmale[z.stelle+1]
	if naechstes.Art != ArtOperator || naechstes.Text != ":" {
		return "", false
	}

	name := z.jetzt().Text
	z.stelle += 2

	// **Hinter dem Doppelpunkt ist ein Umbruch keiner** (#214). So steht es im
	// Initskript von `magic.mse-game`:
	//
	//	word_list_artifact:
	//	  "Altmetall,"
	//	+ "Attraction,"
	//
	// Dieselbe Regel gilt seit jeher hinter `:=` und hinter einem Operator
	// (`stufenoperator`); hier fehlte sie — und diese eine Lücke hielt **alle
	// 190 Stile** des bezogenen Packs auf, weil jeder dasselbe Spielskript lädt.
	//
	// **Der Wert darf umbrechen, nicht fehlen**: was danach kommt, liest `wert`,
	// und am Quellende meldet es „hier fehlt ein Wert" wie zuvor.
	z.umbruecheUeberspringen()

	return name, true
}

// abbildung liest `[ name: wert, … ]`.
//
// Die Schlüssel sind **Namen, keine Ausdrücke**: in `magic-m15` steht
// `[card: "card.jpg", pt: "pt.png"]`, in `game_script` `[stylesheet: "…"]`.
// Ein Ausdruck als Schlüssel kommt nirgends vor, und einen zuzulassen hieße,
// eine Form zu bauen, die keine Vorlage benutzt.
func (z *zerteiler) abbildung(beginn Merkmal) (*Knoten, error) {
	z.stelle++

	knoten := &Knoten{Form: FormAbbildung, Zeile: beginn.Zeile, Spalte: beginn.Spalte}

	for {
		z.umbruecheUeberspringen()

		if z.istOperator("]") {
			z.stelle++

			return knoten, nil
		}

		// **Am Quellende ist die Klammer der Befund** — dieselbe Regel wie
		// bei den Argumenten, und aus derselben Quelle: MSEs Schleife über
		// die Glieder hält bei `TOK_EOF` an (`src/script/parser.cpp`, Zeile
		// 483, Fassung f9d9356 vom 2021-07-06). `[1,` gibt am Orakel
		// `(0): Expected closing ']' for this '['`.
		if z.amEnde() {
			return nil, klammerOffen(beginn, "]")
		}

		// **Ohne Namen zählt die Stelle.** Eine Liste **ist** in MSE-Script
		// eine Abbildung mit Zahlen als Namen — genau deshalb greift
		// `colors.0` auf ihr erstes Glied zu. Der erste Anlauf verlangte
		// überall ein `name:` und wies damit jede Liste ab; 33 Initskripte
		// eines bezogenen Packs scheiterten daran
		// (#186).
		//
		// **Gemischt ist erlaubt**, und die Zahl zählt die Stelle im Ganzen,
		// nicht die der namenlosen Glieder: `[a: 1, 2]` hat ein `a` und eine
		// `1`, denn das zweite Glied steht an Stelle eins.
		name, benannt := z.benanntesArgument()
		if !benannt {
			name = strconv.Itoa(len(knoten.Kinder))
		}

		wert, err := z.ausdruck()
		if err != nil {
			return nil, err
		}

		knoten.Kinder = append(knoten.Kinder, wert)
		knoten.Namen = append(knoten.Namen, name)

		z.umbruecheUeberspringen()

		if z.istOperator(",") {
			z.stelle++
		}
	}
}

// index liest `ziel[ausdruck]`.
func (z *zerteiler) index(ziel *Knoten) (*Knoten, error) {
	beginn := z.jetzt()

	z.stelle++

	// **Hinter der Klammer darf der Ausdruck umbrechen** (#214). So steht es in
	// `magic.mse-game`:
	//
	//	else	colors[
	//	font_color_positions[
	//
	// Dieselbe Regel wie hinter `:=`, hinter einem Operator und hinter `name:`
	// — ein Umbruch mitten in einem Ausdruck ist nichts. Der Aufruf hatte sie
	// schon (`argumente` überspringt in seiner Schleife); der Index nicht.
	z.umbruecheUeberspringen()

	schluessel, err := z.ausdruck()
	if err != nil {
		return nil, err
	}

	z.umbruecheUeberspringen()

	// **Hier und nicht vorher.** `x[` meldet weiterhin den fehlenden Wert am
	// Quellende: MSE liest hinter `[` erst einen Ausdruck und erwartet die
	// Klammer danach, und am Orakel gibt `x[` entsprechend
	// `(2): Expected 'expression' instead of 'end of input'`.
	if !z.istOperator("]") {
		return nil, klammerOffen(beginn, "]")
	}

	z.stelle++

	return &Knoten{
		Form: FormIndex, Kinder: []*Knoten{ziel, schluessel},
		Zeile: ziel.Zeile, Spalte: ziel.Spalte,
	}, nil
}
