package msescript

// **`trace`, `warning` und `error`** — Werkstück A von Phase 4.9
// (#1324).
//
// Alles aus `src/script/functions/basic.cpp`, Zeilen 23–65, Fassung f9d9356
// vom 2021-07-06. Wohin die Meldungen gehen, steht in `meldung.go`.

import "context"

func init() {
	eintragen("trace", meldeHinweis)
	eintragen("warning", meldeWarnung)
	eintragen("error", meldeFehler)
}

// meldeHinweis ist `trace(text)`.
//
//	SCRIPT_PARAM_C(String, input);
//	queue_message(MESSAGE_INFO, _("Trace: ") + input);
//	SCRIPT_RETURN(input);
//
// Zeilen 23–32. **Es gibt seine Eingabe zurück** — am Orakel bestätigt:
// `to_string(trace("y"))` ergibt `"y"`. Damit lässt es sich mitten in einen
// Ausdruck setzen, ohne ihn zu ändern; genau dafür ist es da.
//
// **Ohne `condition:`**, anders als die beiden anderen. Steht so im Original.
func meldeHinweis(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("trace"); err != nil {
		return Nichts, err
	}

	eingabe := ruf.Stellung[0].AlsText()
	ruf.meldeAus(MeldungHinweis, eingabe)

	return Text(eingabe), nil
}

// meldeWarnung ist `warning(text, condition: b)` — Zeilen 33–40.
func meldeWarnung(_ context.Context, ruf Ruf) (Wert, error) {
	return meldeWenn(ruf, "warning", MeldungWarnung)
}

// meldeFehler ist `error(text, condition: b)`.
//
//	if (condition) { queue_message(MESSAGE_ERROR, input); }
//	return script_nil;
//
// Zeilen 58–65. **Es hält nichts an.** Wer `error` ruft, sagt etwas über die
// Daten, nicht über das Skript — der Rückgabewert ist `nil`, und die
// Auswertung läuft weiter.
func meldeFehler(_ context.Context, ruf Ruf) (Wert, error) {
	return meldeWenn(ruf, "error", MeldungFehler)
}

// meldeWenn ist der gemeinsame Rumpf von `warning` und `error`.
//
//	SCRIPT_PARAM_DEFAULT_C(bool, condition, true);
//
// **Die Vorgabe ist wahr**: wer nichts sagt, meldet.
func meldeWenn(ruf Ruf, name string, art Meldungsart) (Wert, error) {
	if err := ruf.eineEingabe(name); err != nil {
		return Nichts, err
	}

	if bedingung, da := ruf.Benannter("condition"); da {
		trifftZu, err := bedingung.AlsWahrheitswert()
		if err != nil {
			return Nichts, ruf.fehlerAn(
				name + " braucht für condition: eine Wahrheit: " + err.Error())
		}

		if !trifftZu {
			return Nichts, nil
		}
	}

	ruf.meldeAus(art, ruf.Stellung[0].AlsText())

	return Nichts, nil
}

// meldeAus legt eine Meldung samt ihrer Stelle in die Umgebung des Rufs.
//
// **Ohne Umgebung geht sie verloren**, und das ist derselbe Fall wie im Kopf
// von [Ruf]: wer eine eingebaute Funktion von Hand ruft, gibt keine mit.
func (r Ruf) meldeAus(art Meldungsart, text string) {
	if r.Umgebung == nil {
		return
	}

	r.Umgebung.melde(Meldung{
		Art:    art,
		Text:   text,
		Zeile:  r.Zeile,
		Spalte: r.Spalte,
	})
}
