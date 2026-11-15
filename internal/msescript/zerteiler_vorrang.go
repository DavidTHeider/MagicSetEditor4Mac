package msescript

// Die zweistelligen Operatoren und ihr Vorrang — die Kette, die `Zerteile`
// zwischen `ausdruck` und `vorzeichen` legt. Aus `zerteiler_formen.go`
// herausgelöst, als QA-03 sie zu MSEs Stufen umbaute (#1656).

// stufen sind die zweistelligen Operatoren, schwächste Bindung zuerst — MSEs
// `PREC_AND`, `PREC_CMP`, `PREC_ADD` und `PREC_MUL`.
//
// **`and`, `or`, `xor` und `or else` auf einer Stufe**, nach links verkettet:
// `true or false and false` ist bei MSE `false`
// (`src/script/parser.cpp`, Zeilen 753–778, Fassung f9d9356 vom 2021-07-06;
// am Orakel abgefragt, `vorrang_test.go`).
var stufen = [][]string{
	{"or else", "or", "xor", "and"},
	{"==", "!=", "<", ">", "<=", ">="},
	{"+", "-"},
	{"*", "/", "mod"},
}

// summenstufe ist die Stufe von `+` und `-` — so weit reicht `or else` nach
// rechts: MSE liest dort `parseOper(…, PREC_ADD, …)`.
const summenstufe = 2

// verkettung liest einen Ausdruck aus Operatoren ab `stufe`, nach links
// verkettet.
//
// **Eine Schleife über jede Stufe ab `stufe`, wie MSEs `parseOper`**, und
// nicht eine Stufe je Aufruf. Der Unterschied zeigt sich allein an `or else`:
// es steht auf der schwächsten Stufe, liest rechts aber nur eine Summe — und
// was danach kommt, gilt dem ganzen Ausdruck davor. `a or else "" != ""` ist
// bei MSE `(a or else "") != ""`; eine Stufe je Aufruf las `a or else ("" !=
// "")`, und `lines_of_text` in `magic.mse-game` bekam einen Text als
// Bedingung.
func (z *zerteiler) verkettung(stufe int) (*Knoten, error) {
	links, err := z.vorzeichen()
	if err != nil {
		return nil, err
	}

	for {
		zeichen, eigene, gefunden := z.operatorAb(stufe)
		if !gefunden {
			return links, nil
		}

		z.stelle++
		z.umbruecheUeberspringen()

		rechts, err := z.verkettung(rechterRand(zeichen, eigene))
		if err != nil {
			return nil, err
		}

		links = &Knoten{
			Form: FormBinaer, Text: zeichen, Kinder: []*Knoten{links, rechts},
			Zeile: links.Zeile, Spalte: links.Spalte,
		}
	}
}

// rechterRand ist die Stufe, ab der die rechte Seite eines Operators liest —
// **die nächste über seiner eigenen, außer bei `or else`**.
func rechterRand(zeichen string, eigene int) int {
	if zeichen == "or else" {
		return summenstufe
	}

	return eigene + 1
}

// operatorAb sucht den nächsten Operator einer Stufe ab `stufe` und nennt
// seine Stufe. Er sieht über einen Umbruch hinweg, wenn dahinter der Operator
// steht — aber nur dann.
//
// In `magic-m15` endet eine Zeile auf `and`, in `yugioh` beginnt eine mit
// `else`. Beides muss durchgehen, ohne dass zwei nebeneinanderstehende
// Zuweisungen zu einem Ausdruck verschmelzen.
func (z *zerteiler) operatorAb(stufe int) (zeichen string, eigene int, gefunden bool) {
	gemerkt := z.stelle
	z.umbruecheUeberspringen()

	jetzt := z.jetzt()
	if !z.amEnde() && (jetzt.Art == ArtOperator || jetzt.Art == ArtSchluesselwort) {
		for eigene := stufe; eigene < len(stufen); eigene++ {
			for _, zeichen := range stufen[eigene] {
				if zeichen == "or else" {
					if jetzt.Text == "or" && z.folgtElse() {
						// Zwei Merkmale, ein Operator: eines mehr überspringen.
						z.stelle++

						return zeichen, eigene, true
					}

					continue
				}

				if jetzt.Text == zeichen {
					return zeichen, eigene, true
				}
			}
		}
	}

	z.stelle = gemerkt

	return "", 0, false
}
