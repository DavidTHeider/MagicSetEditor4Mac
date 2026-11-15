package msescript

// TextskriptFehler gibt den ersten Fehler in einem **Textskript** — einer
// Zeichenkette mit eingebetteten Ausdrücken, ohne umschließende
// Anführungszeichen — oder nil
// (#1934).
//
// # Warum streng
//
// [AuswertenTextskript] lässt einen Ausdruck, der sich nicht lesen lässt, als
// Text stehen (siehe [zerteiler.textglied]); für einen Erinnerungstext mit
// einem Muster wie `\x{20AC}` ist das richtig. **MSE2 unterscheidet das
// nicht**: im Zeichenkettenmodus öffnet jede ungeschützte `{` einen Ausdruck,
// und das Ende der Eingabe darin ist ein Fehler (`TokenIterator` mit
// `string_mode`, `src/script/parser.cpp` Zeilen 139–143 und 249–260, Fassung
// f9d9356 vom 2021-07-06). Wer einen Bildwert wie `border-mask{pt()}.png`
// prüft, will diese Antwort — sonst würde ein kaputter Ausdruck still zum
// Dateinamen.
//
// **Die Musterflucht bleibt Text**, wie in [textteile]: `\x{20AC}` ist keine
// Einbettung, und eine Prüfung, die dort einen Fehler fände, fände ihn in
// Vorlagen, die MSE2 wie dieses Programm zeichnen.
//
// # Die Stelle
//
// Zeile und Spalte gelten in `quelle`: ein Fehler im Ausdruck wird um die
// Stelle hinter seiner `{` versetzt.
func TextskriptFehler(quelle string) error {
	runen := []rune(quelle)
	zeile, spalte := 1, 1

	for stelle := 0; stelle < len(runen); stelle++ {
		switch {
		case runen[stelle] == '\\' && stelle+1 < len(runen):
			stelle++
			spalte += 2

			continue

		case runen[stelle] == '\n':
			zeile, spalte = zeile+1, 1

			continue

		case runen[stelle] != '{' || nachMusterflucht(string(runen[:stelle])):
			spalte++

			continue
		}

		ende, gefunden := klammerende(runen, stelle)
		if !gefunden {
			return klammerOffen(Merkmal{Text: "{", Zeile: zeile, Spalte: spalte}, "}")
		}

		if _, err := Zerteile(string(runen[stelle+1 : ende])); err != nil {
			return versetzterFehler(err, zeile, spalte+1)
		}

		// **Weiter hinter der `}`**, mit der Stelle, die der Ausdruck
		// erreicht hat — er darf über Zeilen gehen.
		for _, zeichen := range runen[stelle:ende] {
			if zeichen == '\n' {
				zeile, spalte = zeile+1, 1
			} else {
				spalte++
			}
		}

		spalte++
		stelle = ende
	}

	return nil
}

// versetzterFehler gibt einen Fehler des Zerteilers mit der Stelle in der
// umgebenden Quelle — `zeile` und `spalte` sind die des ersten Zeichens des
// Ausdrucks.
//
// **Über [Befund]**, wie `zerteilfehler` in `internal/service`: der Zerteiler
// gibt nichts anderes als [Fehler], und ein Zweig für den anderen Fall bliebe
// unbetreten (§3). **Verändert wird er an Ort und Stelle**: [Zerteile] merkt
// sich Bäume, keine Fehler, und jeder Aufruf gibt einen frischen.
func versetzterFehler(err error, zeile, spalte int) error {
	befund := Befund(err)
	if befund.Zeile == 1 {
		befund.Spalte = spalte + befund.Spalte - 1
	}

	befund.Zeile = zeile + befund.Zeile - 1

	return befund
}
