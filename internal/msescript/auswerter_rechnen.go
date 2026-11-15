package msescript

import "math"

// Was ein Rechenzeichen tut — und was es aus seinen beiden Seiten macht.
//
// **Getrennt von `auswerter_verknuepfung.go`**, weil es eine andere Frage
// beantwortet: dort steht, *welche* Bedeutung ein Zeichen bekommt, hier, was
// die gewählte Bedeutung mit den Werten anstellt. Die Umwandlungsregeln selbst
// stehen in `umwandlung.go`.

func (a *auswerter) rechnung(knoten *Knoten, links, rechts Wert) (Wert, error) {
	// **Umgewandelt statt abgewiesen** — `OPERATOR_DI` in `context.cpp:436`
	// ruft auf beiden Seiten `toDouble`, und was sich nicht umwandeln lässt,
	// wirft dort. Bis #240 verlangte dieses Programm zwei Zahlwerte; das wies
	// 1.606 Feldstile ab, in denen `nil` stand, wo MSE null rechnet.
	linkeZahl, err := links.AlsZahl()
	if err != nil {
		return Nichts, a.fehler(knoten, SchluesselRechnenLinks, knoten.Text, err.Error())
	}

	rechteZahl, err := rechts.AlsZahl()
	if err != nil {
		return Nichts, a.fehler(knoten, SchluesselRechnenRechts, knoten.Text, err.Error())
	}

	// **Ein Zahltyp, MSE hat zwei** — `SCRIPT_INT` und `SCRIPT_DOUBLE`.
	// `OPERATOR_DI` rechnet ganzzahlig, solange keine Seite `double` ist
	// (`src/script/context.cpp`, Zeilen 436–442, Fassung f9d9356 vom
	// 2021-07-06); ein Text, der wie eine Kommazahl aussieht, wirft dort bei
	// `toInt`, und `"1.5" * 2` gelingt hier.
	//
	// **Gemessen über den Packlauf** (190 Stile, 5.950 Feldstile): **null**
	// Rechnungen mit einem solchen Text
	// (#271).
	// Für `-`, `*` und die Vergleiche ist der Unterschied ohnehin keiner,
	// solange beide Seiten ganze Zahlen sind.
	//
	// **Was fehlt, ist `div`** — MSEs **ganzzahlige** Division
	// (`I_DIV`, `parser.cpp:794`); `/` ist `I_FDIV` und rechnet in
	// Kommazahlen, so wie hier. Im vollen Pack steht `div` **null**-mal.
	switch knoten.Text {
	case "-":
		return Zahl(linkeZahl - rechteZahl), nil

	case "*":
		return Zahl(linkeZahl * rechteZahl), nil
	}

	// **Durch null zu teilen ist bei MSE kein Fehler.** `I_FDIV` rechnet
	// zwei `double` durcheinander und lässt IEEE 754 antworten:
	//
	//	case I_FDIV:
	//	  a = to_script(a->toDouble() / b->toDouble());
	//
	// `src/script/context.cpp`, Zeilen 485–487, Fassung f9d9356 vom
	// 2021-07-06; am Orakel bestätigt: `1/0` → `inf`, `0/0` → `nan`.

	// **`mod` folgt der Zahlenart** — `context.cpp:495` von MSE2:
	//
	//	if (at == SCRIPT_DOUBLE || bt == SCRIPT_DOUBLE) fmod(…) else a % b
	//
	// `10.5 mod 3` ergibt damit 1,5 und nicht 1 — auch wenn im Pack nur ganze
	// Werte davorstehen.
	if knoten.Text == "mod" {
		if linkeZahl == math.Trunc(linkeZahl) && rechteZahl == math.Trunc(rechteZahl) {
			// **Der Rest durch null ist der Zähler** — auf diesem Rechner.
			//
			// `a->toInt() % b->toInt()` (`src/script/context.cpp`, Zeile
			// 499, Fassung f9d9356 vom 2021-07-06) ist bei `b == 0` in C++
			// **undefiniert**. Auf arm64 gibt `sdiv` null, und der Rest
			// rechnet sich als `a − 0·0 = a`; das Orakel antwortet auf
			// `3 mod 0` deshalb `3`. Auf x86 fiele dasselbe Programm mit
			// einer Ausnahme aus.
			//
			// **adapted, nicht ported**: die Regel steht in keinem
			// Quelltext, sondern in einer Befehlssatzarchitektur. Gebaut
			// wird für darwin/arm64 (§12), und die Referenzbilder stammen
			// von genau diesem Binärprogramm.
			if rechteZahl == 0 {
				return Zahl(linkeZahl), nil
			}

			return Zahl(float64(int64(linkeZahl) % int64(rechteZahl))), nil
		}

		// `fmod(x, 0)` ist dagegen definiert und ergibt `nan`.
		return Zahl(math.Mod(linkeZahl, rechteZahl)), nil
	}

	return Zahl(linkeZahl / rechteZahl), nil
}

// vergleich ordnet zwei Werte.
//
// **MSE rechnet hier**, es vergleicht nicht: `<`, `>`, `<=` und `>=` gehen in
// `context.cpp:528` durch dasselbe `OPERATOR_DI` wie `-` und `*`. Zwei leere
// Werte sind damit gleich groß, und das waren 327 der gemessenen Befunde.
//
// **Der Vergleich zweier Zeichenketten bleibt trotzdem stehen**, und das ist
// eine bewusste Abweichung: MSE riefe dort `toInt` und würfe für jedes Wort.
// Wer `if card.name < "M"` schreibt, meint offensichtlich die Ordnung der
// Zeichen, und ein Fehler an dieser Stelle nähme einer Vorlage etwas weg, ohne
// etwas zu gewinnen. Beide Seiten müssen Text sein — sonst gewänne die
// Zeichenordnung gegen die Zahl, und `"9" < 10` wäre falsch.
func (a *auswerter) vergleich(knoten *Knoten, links, rechts Wert) (Wert, error) {
	var kleiner, gleichgross bool

	if links.Art == ArtTextwert && rechts.Art == ArtTextwert {
		kleiner = links.Text < rechts.Text
		gleichgross = links.Text == rechts.Text
	} else {
		linkeZahl, err := links.AlsZahl()
		if err != nil {
			return Nichts, a.fehler(knoten, SchluesselVergleichLinks, knoten.Text, err.Error())
		}

		rechteZahl, err := rechts.AlsZahl()
		if err != nil {
			return Nichts, a.fehler(knoten, SchluesselVergleichRechts, knoten.Text, err.Error())
		}

		kleiner = linkeZahl < rechteZahl
		gleichgross = linkeZahl == rechteZahl
	}

	switch knoten.Text {
	case "<":
		return Wahrheit(kleiner), nil

	case ">":
		return Wahrheit(!kleiner && !gleichgross), nil

	case "<=":
		return Wahrheit(kleiner || gleichgross), nil
	}

	return Wahrheit(!kleiner), nil
}
