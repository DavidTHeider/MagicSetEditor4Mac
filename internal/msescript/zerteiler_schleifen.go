package msescript

// Die Schleifenformen von MSE-Script.
//
// Von `zerteiler_formen.go` abgetrennt, weil die Datei über die Zeilengrenze
// aus AGENTS.md §2 gewachsen war. Der Schnitt folgt der Frage: dort steht, wie
// ein Ausdruck entsteht, hier, wie eine Wiederholung entsteht.
//
// **Zwei Formen, und beide stehen in echten Vorlagen**
// (#51):
// `for each x in liste do …` geht eine Liste durch, `for x from a to b do …`
// zählt.

// schleife liest `for each <name> in <liste> do <ausdruck>`.
//
// # Woher die Form stammt
//
// Aus allen 19 Exportvorlagen von `M15-Magic-Pack`
// (#51):
//
//	write_cards := to_text(for each card in sort_list(cards, …) do write_card())
//
// **`each` gehört dazu und ist nicht wegzulassen.** In keiner der Vorlagen
// steht `for x in …`; wer die Form großzügiger läse, akzeptierte etwas, das
// MSE selbst vielleicht abweist — und der erste, der sie so schriebe, bekäme
// hier eine Datei, die woanders nicht läuft.
func (z *zerteiler) schleife() (*Knoten, error) {
	beginn := z.jetzt()
	z.stelle++

	// **Zwei Formen, und beide stehen in echten Vorlagen**: `for each x in …`
	// geht eine Liste durch, `for x from a to b` zählt. Die zweite steht in
	// `magic-field-test.mse-export-template` und in
	// `magic.mse-game/statistics_script`.
	if !z.istWort("each") {
		return z.zaehlung(beginn)
	}

	z.stelle++

	name := z.jetzt()
	if z.amEnde() || name.Art != ArtBezeichner {
		return nil, z.hier(SchluesselFehltNach, "`each`", "ein Name")
	}

	z.stelle++

	if !z.istWort("in") {
		return nil, z.hier(SchluesselFehltNach, "dem Namen", "`in`")
	}

	z.stelle++
	z.umbruecheUeberspringen()

	liste, err := z.verkettung(0)
	if err != nil {
		return nil, err
	}

	z.umbruecheUeberspringen()

	if !z.istWort("do") {
		return nil, z.hier(SchluesselFehltNach, "der Liste", "`do`")
	}

	z.stelle++
	z.umbruecheUeberspringen()

	rumpf, err := z.ausdruck()
	if err != nil {
		return nil, err
	}

	return &Knoten{
		Form: FormSchleife, Text: name.Text, Kinder: []*Knoten{liste, rumpf},
		Zeile: beginn.Zeile, Spalte: beginn.Spalte,
	}, nil
}

// zaehlung liest `for <name> from <a> to <b> do <ausdruck>`.
//
// **Die obere Grenze zählt mit.** `for x from 0 to length(cmds)-1 do …` sagt
// es: wer bis `length` zählte und die letzte Stelle ausließe, schriebe das
// `-1` nicht dazu.
func (z *zerteiler) zaehlung(beginn Merkmal) (*Knoten, error) {
	name := z.jetzt()
	if z.amEnde() || name.Art != ArtBezeichner {
		return nil, z.hier(SchluesselFehltNach, "`for`", "`each` oder ein Name")
	}

	z.stelle++

	if !z.istWort("from") {
		return nil, z.hier(SchluesselFehltNach, "dem Namen", "`from`")
	}

	z.stelle++
	z.umbruecheUeberspringen()

	von, err := z.verkettung(0)
	if err != nil {
		return nil, err
	}

	z.umbruecheUeberspringen()

	if !z.istWort("to") {
		return nil, z.hier(SchluesselFehltNach, "dem Anfang", "`to`")
	}

	z.stelle++
	z.umbruecheUeberspringen()

	bis, err := z.verkettung(0)
	if err != nil {
		return nil, err
	}

	z.umbruecheUeberspringen()

	if !z.istWort("do") {
		return nil, z.hier(SchluesselFehltNach, "dem Ende", "`do`")
	}

	z.stelle++
	z.umbruecheUeberspringen()

	rumpf, err := z.ausdruck()
	if err != nil {
		return nil, err
	}

	return &Knoten{
		Form: FormZaehlung, Text: name.Text, Kinder: []*Knoten{von, bis, rumpf},
		Zeile: beginn.Zeile, Spalte: beginn.Spalte,
	}, nil
}

// fall liest `case <ausdruck> of <wert>: <ausdruck>, … else: <ausdruck>;`.
//
// # Die Form, belegt
//
// Achtmal in vier Dateien von `M15-Magic-Pack` (#214):
//
//	case input of
//	"token": true,
//	"emblem": true,
//	else: false;
//
// Die Paare stehen wie in einer Abbildung — Schlüssel, Doppelpunkt, Wert,
// Komma. `else` ist die Vorgabe; das abschließende `;` ist der Trenner zur
// nächsten Anweisung und wird von der Folge gelesen.
//
// # Was der Zerteiler hier **nicht** entscheidet
//
// **Ob ein Schlüssel doppelt vorkommt.** Zwei gleiche Fälle sind kein
// Zerteilfehler; der erste gewinnt, und das steht beim Auswerter.
func (z *zerteiler) fall() (*Knoten, error) {
	beginn := z.jetzt()
	z.stelle++

	pruefung, err := z.ausdruck()
	if err != nil {
		return nil, err
	}

	if !z.istWort("of") {
		return nil, z.hier(SchluesselFehltNach, "`case`", "`of`")
	}

	z.stelle++

	knoten := &Knoten{
		Form: FormFall, Kinder: []*Knoten{pruefung},
		Zeile: beginn.Zeile, Spalte: beginn.Spalte,
	}

	for {
		z.umbruecheUeberspringen()

		// **Der Vorgabezweig schließt den Fall.** Danach kommt nichts mehr,
		// was zu ihm gehörte.
		if z.istWort("else") {
			z.stelle++

			if !z.istOperator(":") {
				return nil, z.hier(SchluesselFehltNach, "`else`", "`:`")
			}

			z.stelle++
			z.umbruecheUeberspringen()

			vorgabe, fehler := z.ausdruck()
			if fehler != nil {
				return nil, fehler
			}

			knoten.Namen = append(knoten.Namen, "")
			knoten.Kinder = append(knoten.Kinder, nil, vorgabe)

			return knoten, nil
		}

		if z.amEnde() || z.istOperator("}", ")", ";") {
			// **Mindestens ein Fall.** `case x of` ohne einen einzigen Zweig ist
			// ein Ausdruck, den jemand nicht zu Ende geschrieben hat — und ohne
			// diese Zeile gäbe er stillschweigend `nil`.
			if len(knoten.Namen) == 0 {
				return nil, z.hier(SchluesselFehltNach, "`of`", "mindestens ein Fall")
			}

			return knoten, nil
		}

		schluessel, fehler := z.wert()
		if fehler != nil {
			return nil, fehler
		}

		if !z.istOperator(":") {
			return nil, z.hier(SchluesselFehltNach, "dem Fall", "`:`")
		}

		z.stelle++
		z.umbruecheUeberspringen()

		wert, fehler := z.ausdruck()
		if fehler != nil {
			return nil, fehler
		}

		knoten.Namen = append(knoten.Namen, "fall")
		knoten.Kinder = append(knoten.Kinder, schluessel, wert)

		z.umbruecheUeberspringen()

		if z.istOperator(",") {
			z.stelle++
		}
	}
}
