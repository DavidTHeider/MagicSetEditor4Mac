package msescript

import "strconv"

// nachForm schaltet über die Form des Knotens.
func (a *auswerter) nachForm(knoten *Knoten, umgebung *Umgebung) (Wert, error) {
	switch knoten.Form {
	case FormZahl:
		zahl, err := strconv.ParseFloat(knoten.Text, 64)
		if err != nil {
			return Nichts, a.fehler(knoten, SchluesselKeineZahl, strconv.Quote(knoten.Text))
		}

		return Zahl(zahl), nil

	case FormZeichenkette:
		return Text(knoten.Text), nil

	case FormNichts:
		return Nichts, nil

	case FormName:
		return a.name(knoten, umgebung)

	case FormKlammer:
		// **Ohne Umgebung.** Eine geschweifte Klammer schließt nichts ein; sie
		// bekommt ihre Umgebung beim Aufruf (`umgebung.go`).
		return NeueFunktion(knoten.Kinder[0]), nil

	case FormFolge:
		return a.folge(knoten, umgebung)

	case FormZuweisung:
		return a.zuweisung(knoten, umgebung)

	case FormBedingung:
		return a.bedingung(knoten, umgebung)

	case FormFall:
		return a.fall(knoten, umgebung)

	case FormSchleife:
		return a.schleife(knoten, umgebung)

	case FormZaehlung:
		return a.zaehlung(knoten, umgebung)

	case FormUnaer:
		return a.unaer(knoten, umgebung)

	case FormBinaer:
		return a.binaer(knoten, umgebung)

	case FormAbbildung:
		return a.abbildung(knoten, umgebung)

	case FormIndex:
		return a.index(knoten, umgebung)

	case FormZugriff:
		return a.zugriff(knoten, umgebung)

	case FormAufruf, FormBindung:
		return a.aufruf(knoten, umgebung)
	}

	return Nichts, a.fehler(knoten, SchluesselFormUnbekannt, strconv.Quote(string(knoten.Form)))
}

func (a *auswerter) name(knoten *Knoten, umgebung *Umgebung) (Wert, error) {
	if wert, da := umgebung.Suche(knoten.Text); da {
		a.merke(knoten.Text)

		return wert, nil
	}

	if wert, da := grundnamen[knoten.Text]; da {
		return wert, nil
	}

	return Nichts, a.fehler(knoten, SchluesselNameUnbekannt, strconv.Quote(knoten.Text))
}

// folge wertet der Reihe nach; der letzte Wert zählt.
//
// Zuweisungen wirken dabei auf **dieselbe** Ebene, sonst sähe der zweite
// Ausdruck den Namen nicht, den der erste angelegt hat.
func (a *auswerter) folge(knoten *Knoten, umgebung *Umgebung) (Wert, error) {
	letzter := Nichts

	for _, kind := range knoten.Kinder {
		wert, err := a.werte(kind, umgebung)
		if err != nil {
			return Nichts, err
		}

		letzter = wert
	}

	return letzter, nil
}

func (a *auswerter) zuweisung(knoten *Knoten, umgebung *Umgebung) (Wert, error) {
	wert, err := a.werte(knoten.Kinder[0], umgebung)
	if err != nil {
		return Nichts, err
	}

	umgebung.Setze(knoten.Text, wert)

	return wert, nil
}

func (a *auswerter) bedingung(knoten *Knoten, umgebung *Umgebung) (Wert, error) {
	pruefung, err := a.werte(knoten.Kinder[0], umgebung)
	if err != nil {
		return Nichts, err
	}

	trifft, brauchbar := pruefung.AlsWahrheit()

	// **Eine Zahl gilt hier als Bedingung, sonst nirgends** (#216).
	//
	// Belegt in zwei Exportvorlagen von `M15-Magic-Pack`:
	// `if to_number(substring(…)) then …` — eine Anzahl, auf „gibt es
	// überhaupt" geprüft. Null ist falsch, alles andere wahr.
	//
	// **Die Lockerung steht hier und nicht in [Wert.AlsWahrheit]**, weil sie
	// dort `1 and true` und `not 1` mitzöge. Für die ist keine Fundstelle da,
	// und drei Zusagen in `auswerter_stellen_test.go` verlangen dort einen
	// Fehler. Ein Fall macht keine Klasse.
	if !brauchbar && pruefung.Art == ArtZahlwert {
		trifft, brauchbar = pruefung.Zahl != 0, true
	}

	// **Ein Fehlerwert wirft seine eigene Meldung.** „hier steht Fehler" sagte
	// nichts darüber, **welcher** Zugriff ins Leere ging (siehe [Fehlerwert]).
	if pruefung.IstFehler() {
		return Nichts, a.fehler(knoten.Kinder[0], SchluesselDurchgereicht, pruefung.Text)
	}

	if !brauchbar {
		return Nichts, a.fehler(knoten.Kinder[0],
			SchluesselBedingungWahrheit, string(pruefung.Art))
	}

	// Nur der genommene Zweig wird gerechnet — der andere darf ruhig
	// fehlschlagen. In `magic-m15` steht `if styling.overlay == "" then nil
	// else styling.overlay + "/overlay.png"`, und der else-Zweig setzt voraus,
	// dass die Bedingung falsch war.
	if trifft {
		return a.werte(knoten.Kinder[1], umgebung)
	}

	return a.werte(knoten.Kinder[2], umgebung)
}

func (a *auswerter) unaer(knoten *Knoten, umgebung *Umgebung) (Wert, error) {
	darunter, err := a.werte(knoten.Kinder[0], umgebung)
	if err != nil {
		return Nichts, err
	}

	if knoten.Text == "not" {
		trifft, brauchbar := darunter.AlsWahrheit()
		if !brauchbar {
			return Nichts, a.fehler(knoten, SchluesselNichtWahrheitswert,
				string(darunter.Art))
		}

		return Wahrheit(!trifft), nil
	}

	if darunter.Art != ArtZahlwert {
		return Nichts, a.fehler(knoten, SchluesselVorzeichenZahl,
			string(darunter.Art))
	}

	return Zahl(-darunter.Zahl), nil
}

// schleife geht eine Liste durch und **addiert** die Ergebnisse.
//
// # Was herauskommt
//
// Das, was `+` aus den Ergebnissen macht — angefangen bei `nil`:
//
//	// the loop body should have a net stack effect of 0, but the entire expression of +1
//	// solution: add all results from the body, start with nil
//	script.addInstruction(I_PUSH_CONST, script_nil);      //    push nil
//	…
//	parseOper(input, script, PREC_SET, I_BINARY, I_ADD);  // EEE;  add
//
// `src/script/parser.cpp`, Zeilen 554–608, Fassung f9d9356 vom 2021-07-06.
//
// Damit ergibt `for … do [wert]` eine **flache** Liste (zwei Listen verketten
// sich), `for … do zahl` die **Summe** und `for … do text` den
// aneinandergesetzten Text. Genau so benutzen die Exportvorlagen sie:
// `to_text(for each card in cards do write_card())`
// (#51).
//
// **Bis #390
// wurde gesammelt statt addiert.** Für `to_text` machte das keinen Unterschied
// — für `for x from 1 to 9 do [wert]` schon: heraus kam eine Liste von Listen,
// und der nächste Zugriff darauf rechnete mit einer Liste statt mit einer Zahl.
// 43 der 165 Befunde über `M15-Magic-Pack` hingen daran.
//
// # Warum der Name in einer eigenen Ebene lebt
//
// `card` gilt für den Durchgang und danach nicht mehr. Wer ihn in der Umgebung
// des Aufrufers ließe, machte aus zwei Schleifen hintereinander eine, deren
// zweite den letzten Wert der ersten sieht.
func (a *auswerter) schleife(knoten *Knoten, umgebung *Umgebung) (Wert, error) {
	liste, err := a.werte(knoten.Kinder[0], umgebung)
	if err != nil {
		return Nichts, err
	}

	if liste.Art != ArtAbbildung {
		return Nichts, a.fehler(knoten.Kinder[0],
			SchluesselFuerJedenListe, string(liste.Art))
	}

	// **`nil` ist der Anfang**, und `+` macht daraus, was die Vorlage meint.
	ergebnis := Nichts

	// **Kein eigener Namensraum je Durchgang.**
	//
	// MSE hat eine Variablentabelle und wechselt die Ebene allein beim
	// **Aufruf** (`case I_CALL: new_scope.reset(new LocalScope(*this))`,
	// `src/script/context.cpp`, Zeilen 168–170, Fassung f9d9356 vom
	// 2021-07-06). Eine Schleife öffnet keine — und deshalb sieht der nächste
	// Durchgang, was der vorige zugewiesen hat.
	//
	// `sub_type_filter` schält so Wort für Wort von `input` ab; mit einer
	// eigenen Ebene je Durchgang wurde aus `Goblin Mage` viermal `Goblin`
	// (#345).
	for i := range liste.Eintraege {
		umgebung.Setze(knoten.Text, liste.Eintraege[i].Wert)

		wert, fehler := a.werte(knoten.Kinder[1], umgebung)
		if fehler != nil {
			return Nichts, fehler
		}

		ergebnis, fehler = a.plus(knoten, ergebnis, wert, umgebung)
		if fehler != nil {
			return Nichts, fehler
		}
	}

	return ergebnis, nil
}

// zaehlung geht eine Zahlenspanne durch — einschließlich der oberen Grenze.
//
// **Eine leere Spanne ergibt eine leere Liste, keinen Fehler.**
// `for x from 0 to count do …` mit `count` gleich −1 steht so in
// `magic.mse-game/statistics_script`; die Vorlage prüft davor selbst, aber
// verlassen kann man sich darauf nicht.
func (a *auswerter) zaehlung(knoten *Knoten, umgebung *Umgebung) (Wert, error) {
	von, err := a.zahlAus(knoten.Kinder[0], umgebung, "der Anfang")
	if err != nil {
		return Nichts, err
	}

	bis, err := a.zahlAus(knoten.Kinder[1], umgebung, "das Ende")
	if err != nil {
		return Nichts, err
	}

	// **`nil` ist der Anfang, `+` fügt zusammen** — derselbe Weg wie bei
	// [auswerter.schleife], und dort steht auch die Quelle (#390).
	ergebnis := Nichts

	for zahl := von; zahl <= bis; zahl++ {
		// **Eine Zahl, die sich nicht mehr erhöhen lässt, hält die Schleife an.**
		//
		// Jenseits von 2^53 verändert `+1` eine `float64` nicht mehr; `zahl <=
		// bis` bliebe dauerhaft wahr, und die Schleife liefe endlos — ohne
		// Fehler, ohne Ausgabe, ohne Ende.
		//
		// **Im Kopf und nicht davor.** Der erste Anlauf prüfte `von < bis &&
		// von+1 == von` und behandelte damit den Fall statt seiner Klasse: eine
		// Spanne, deren Anfang gleich dem Ende ist, und eine, die **unterwegs**
		// an die Grenze stößt, liefen weiter endlos. Der Nachlauf zum Zweitblick
		// über Phase 6 hat beide angestrichen.
		//
		// **Abgewiesen und nicht abgeschnitten**: an dieser Stelle ist nicht zu
		// unterscheiden, ob ein Durchgang gemeint war oder unendlich viele. Eine
		// Schleife, die stillschweigend aufhörte, gäbe ein falsches Ergebnis
		// statt einer Meldung.
		if zahl+1 == zahl {
			return Nichts, a.fehler(knoten.Kinder[0], SchluesselSchleifeSteht,
				strconv.FormatFloat(zahl, 'g', -1, 64))
		}

		// **Kein eigener Namensraum je Durchgang** — der Grund steht bei
		// [auswerter.schleife].
		umgebung.Setze(knoten.Text, Zahl(zahl))

		wert, fehler := a.werte(knoten.Kinder[2], umgebung)
		if fehler != nil {
			return Nichts, fehler
		}

		ergebnis, fehler = a.plus(knoten, ergebnis, wert, umgebung)
		if fehler != nil {
			return Nichts, fehler
		}
	}

	return ergebnis, nil
}

// zahlAus wertet einen Knoten aus und verlangt eine Zahl.
func (a *auswerter) zahlAus(
	knoten *Knoten, umgebung *Umgebung, was string,
) (float64, error) {
	wert, err := a.werte(knoten, umgebung)
	if err != nil {
		return 0, err
	}

	if wert.Art != ArtZahlwert {
		return 0, a.fehler(knoten, SchluesselZaehlschleifeZahl, was,
			string(wert.Art))
	}

	return wert.Zahl, nil
}

// fall wertet `case … of` aus.
//
// **Der erste Treffer gewinnt**, und verglichen wird wie mit `==`. Ohne Treffer
// gilt der `else`-Zweig; gibt es auch den nicht, ist das Ergebnis `nil` — MSE
// kennt für diesen Fall keinen Abbruch, und einen zu erfinden hieße, eine Regel
// zu bauen, die keine Vorlage kennt (#214).
//
// **Nur der genommene Zweig wird gerechnet**, wie bei der Bedingung: die
// übrigen dürfen fehlschlagen.
func (a *auswerter) fall(knoten *Knoten, umgebung *Umgebung) (Wert, error) {
	gesucht, err := a.werte(knoten.Kinder[0], umgebung)
	if err != nil {
		return Nichts, err
	}

	for stelle, art := range knoten.Namen {
		schluessel := knoten.Kinder[1+stelle*2]
		wert := knoten.Kinder[2+stelle*2]

		if art == "" {
			return a.werte(wert, umgebung)
		}

		verglichen, fehler := a.werte(schluessel, umgebung)
		if fehler != nil {
			return Nichts, fehler
		}

		if gleich(gesucht, verglichen) {
			return a.werte(wert, umgebung)
		}
	}

	return Nichts, nil
}
