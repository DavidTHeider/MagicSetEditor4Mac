package msescript

// Die Formen, eine Stufe je Vorrang. Die Reihenfolge steht bei `Zerteile`.
//
// Jede Stufe ruft die nächststärkere und schaut danach, ob ihr eigener Operator
// folgt. Das ist der übliche Bau; er steht hier ausgeschrieben statt über eine
// Tabelle, weil `if`, `:=` und der Nachsatz sich ohnehin nicht in dieselbe
// Schleife fügen.

// folge liest Ausdrücke, bis das Ende oder ein Schlusszeichen kommt.
func (z *zerteiler) folge(schluss []string) (*Knoten, error) {
	var glieder []*Knoten

	for {
		gemerkt := z.stelle
		z.umbruecheUeberspringen()
		getrennt := z.stelle > gemerkt

		if z.amEnde() || (schluss != nil && z.istOperator(schluss...)) {
			break
		}

		// **Getrennt wird am Umbruch, nicht am Leerzeichen.** In einem
		// `init script:` steht je eine Zuweisung auf einer Zeile; `a b` auf
		// derselben Zeile ist keine Folge, sondern ein Ausdruck, den jemand
		// nicht zu Ende geschrieben hat. Ohne diese Bedingung verschluckte der
		// Zerteiler jeden solchen Rest schweigend.
		if len(glieder) > 0 && !getrennt {
			// **In einer Klammer sagt das der Aufrufer, nicht diese
			// Schleife** — Werkstück B von Phase 4.9
			// (#1324).
			//
			// Die Meldung dieser Schleife zeigte auf das Merkmal, an dem der
			// Ausdruck aufhört, und schwiege über die Klammer, die niemand
			// geschlossen hat. Der Aufrufer weiß mehr — er hält die
			// öffnende in der Hand.
			//
			// MSE trennt genauso: in der Klammer liest es **einen**
			// Ausdruck (`parseOper`) und erwartet danach das Schlusszeichen
			// (`src/script/parser.cpp`, Zeilen 447–449, Fassung f9d9356 vom
			// 2021-07-06); was den Ausdruck beendet, ist dann das „instead
			// of" seiner Meldung. Am Orakel gemessen:
			//
			//	(a b)   → (0): Expected closing ')' for this '(' instead of 'b'
			//	(1+2]   → (0): Expected closing ')' for this '(' instead of ']'
			//	1+2]    → (3): Expected 'end of input' instead of ']'
			//
			// Die dritte Zeile ist der Grund für die Bedingung: **ganz
			// oben** gibt es keine Klammer, die man nennen könnte, und dort
			// bleibt der Satz, der schon da war.
			if schluss != nil {
				break
			}

			return nil, z.hier(SchluesselAusdruckGehtWeiter)
		}

		glied, err := z.ausdruck()
		if err != nil {
			return nil, err
		}

		glieder = append(glieder, glied)
	}

	switch len(glieder) {
	case 0:
		return nil, z.hier(SchluesselFehltHier, "ein Ausdruck")
	case 1:
		return glieder[0], nil
	default:
		return &Knoten{
			Form: FormFolge, Kinder: glieder,
			Zeile: glieder[0].Zeile, Spalte: glieder[0].Spalte,
		}, nil
	}
}

// ausdruck ist die schwächste Stufe: die Zuweisung.
func (z *zerteiler) ausdruck() (*Knoten, error) {
	if fehler := z.hinab(); fehler != nil {
		return nil, fehler
	}

	defer z.hinauf()

	links, err := z.bedingung()
	if err != nil {
		return nil, err
	}

	if !z.istOperator(":=") {
		return links, nil
	}

	if links.Form != FormName {
		return nil, z.hier(SchluesselZuweisungBrauchtNamen)
	}

	z.stelle++
	z.umbruecheUeberspringen()

	rechts, err := z.ausdruck()
	if err != nil {
		return nil, err
	}

	return &Knoten{
		Form: FormZuweisung, Text: links.Text, Kinder: []*Knoten{rechts},
		Zeile: links.Zeile, Spalte: links.Spalte,
	}, nil
}

func (z *zerteiler) bedingung() (*Knoten, error) {
	if z.istWort("for") {
		return z.schleife()
	}

	if z.istWort("case") {
		return z.fall()
	}

	if !z.istWort("if") {
		return z.verkettung(0)
	}

	beginn := z.jetzt()
	z.stelle++

	pruefung, err := z.verkettung(0)
	if err != nil {
		return nil, err
	}

	z.umbruecheUeberspringen()

	if !z.istWort("then") {
		return nil, z.hier(SchluesselFehltNach, "`if`", "`then`")
	}

	z.stelle++
	z.umbruecheUeberspringen()

	// **Der `then`-Zweig darf selbst eine Bedingung sein**, und das ist keine
	// Bequemlichkeit: `magic-DayNight.mse-style` schreibt
	// `if color_count == 0 then if land then l else c`, und **53 der 190**
	// Initskripte eines bezogenen Packs scheiterten daran
	// (#186).
	//
	// **Und eine Zuweisung**: `if card.loyalty != "" then output := output + […]`
	// steht so in `magic-m15-compleated-planeswalkers.mse-style`; 31
	// Initskripte scheiterten daran. Deshalb `ausdruck` und nicht `bedingung`
	// — jenes ist die schwächere Stufe und schließt diese ein.
	//
	// Das `else` bindet ans **innere** `if` — die Rekursion greift zuerst.
	// Anders herum läse sich die Zeile oben als „wenn `color_count == 0`, dann
	// (wenn `land` dann `l`), sonst `c`", und die Kette darunter hinge am
	// falschen Zweig.
	dann, err := z.ausdruck()
	if err != nil {
		return nil, err
	}

	// **Der Umbruch wird nur vorläufig übersprungen.**
	//
	// Vor einem `else` muss er weg — `if a then 1\nelse 2` ist eine Form, die
	// echte Pakete schreiben. Kommt aber **kein** `else`, gehört er zurück:
	// sonst verschluckt die Bedingung den Trenner zur nächsten Anweisung, und
	// `if a then 1\nb` meldet „hier hört der Ausdruck auf, aber der Text geht
	// weiter". Gefunden beim Bau von
	// #186 —
	// ein Fehler, der seit Phase 3 dalag und den kein Prüfstück traf, weil
	// keines ein `if` ohne `else` **vor** einer weiteren Anweisung hatte.
	vorDemUmbruch := z.stelle
	z.umbruecheUeberspringen()

	// **`else` darf fehlen.** In `kamen-decade.mse-style/style` Zeile 63 steht
	// `visible: { if card.background == "Normal" then true }`. Der erste Anlauf
	// des Zerteilers wies diese Zeile ab — eine Form, die `docs/FORMATE.md`
	// führt und die ein echtes Paket benutzt. Gefunden vom Zweitblick über
	// Phase 3.
	//
	// Der fehlende Zweig ergibt `nil`. Was MSE tut, ist damit nicht gesagt; für
	// ein `visible`-Feld sind `nil` und `false` dasselbe, und eine andere
	// Vorgabe zu erfinden wäre eine Behauptung mehr, nicht weniger.
	if !z.istWort("else") {
		z.stelle = vorDemUmbruch

		return &Knoten{
			Form: FormBedingung,
			Kinder: []*Knoten{
				pruefung, dann,
				{Form: FormNichts, Zeile: beginn.Zeile, Spalte: beginn.Spalte},
			},
			Zeile: beginn.Zeile, Spalte: beginn.Spalte,
		}, nil
	}

	z.stelle++
	z.umbruecheUeberspringen()

	// Rechts, nicht links: `else if … then … else …` hängt in den letzten
	// Zweig, so wie die Ketten in `yugioh` es meinen.
	sonst, err := z.ausdruck()
	if err != nil {
		return nil, err
	}

	return &Knoten{
		Form: FormBedingung, Kinder: []*Knoten{pruefung, dann, sonst},
		Zeile: beginn.Zeile, Spalte: beginn.Spalte,
	}, nil
}

// folgtElse sieht, ob auf `or` unmittelbar `else` folgt.
//
// `map["stylesheet"] or else "m15-altered"` aus `game_script` Zeile 62: der
// linke Wert gilt, wenn er da ist, sonst der rechte. Ohne diesen Griff läse der
// Zerteiler `or` und stünde dann vor `else` — und meldete einen fehlenden Wert
// an einer Stelle, die in Ordnung ist.
// **Unmittelbar, ohne Umbruch dazwischen.** Der erste Anlauf sah über Zeilen
// hinweg und erfand damit eine Form, die in keinem Paket steht — dafür bekam er
// eine Mehrdeutigkeit geschenkt: `if a then b or` gefolgt von `else c` wäre
// nicht mehr von `b or else c` zu unterscheiden.
func (z *zerteiler) folgtElse() bool {
	naechste := z.stelle + 1

	return naechste < len(z.merkmale) &&
		z.merkmale[naechste].Art == ArtSchluesselwort &&
		z.merkmale[naechste].Text == "else"
}

func (z *zerteiler) vorzeichen() (*Knoten, error) {
	if !z.istWort("not") && !z.istOperator("-") {
		return z.nachsatz()
	}

	beginn := z.jetzt()
	z.stelle++
	z.umbruecheUeberspringen()

	if fehler := z.hinab(); fehler != nil {
		return nil, fehler
	}

	darunter, err := z.vorzeichen()
	z.hinauf()

	if err != nil {
		return nil, err
	}

	return &Knoten{
		Form: FormUnaer, Text: beginn.Text, Kinder: []*Knoten{darunter},
		Zeile: beginn.Zeile, Spalte: beginn.Spalte,
	}, nil
}

// nachsatz liest, was **hinter** einem Wert stehen darf: `.name`, `(…)`, `@(…)`.
func (z *zerteiler) nachsatz() (*Knoten, error) {
	knoten, err := z.wert()
	if err != nil {
		return nil, err
	}

	for {
		switch {
		case z.istOperator("."):
			z.stelle++

			// **Auch eine Zahl.** `default_images[colors.0]` und
			// `template(pure.0, type: "pt")` stehen so in `magic-m15.mse-style`:
			// eine Liste ist eine Abbildung mit Zahlen als Namen, und `.0`
			// greift auf ihr erstes Glied zu. 53 Initskripte scheiterten allein
			// an dieser Stelle
			// (#186).
			name := z.jetzt()
			if z.amEnde() || (name.Art != ArtBezeichner && name.Art != ArtZahl) {
				return nil, z.hier(SchluesselFehltNach, "dem Punkt", "ein Name")
			}

			z.stelle++
			knoten = &Knoten{
				Form: FormZugriff, Text: name.Text, Kinder: []*Knoten{knoten},
				Zeile: knoten.Zeile, Spalte: knoten.Spalte,
			}

		case z.istOperator("["):
			// `template_prefix[type]` aus magic-m15 Zeile 54. Ein Nachsatz wie
			// der Punkt, nur mit einem Ausdruck statt einem Namen.
			knoten, err = z.index(knoten)

		case z.istOperator("("):
			knoten, err = z.argumente(knoten, FormAufruf)

		case z.istOperator("@"):
			z.stelle++

			if !z.istOperator("(") {
				return nil, z.hier(SchluesselFehltNach, "`@`", "eine Klammer mit Argumenten")
			}

			knoten, err = z.argumente(knoten, FormBindung)

		default:
			return knoten, nil
		}

		if err != nil {
			return nil, err
		}
	}
}
