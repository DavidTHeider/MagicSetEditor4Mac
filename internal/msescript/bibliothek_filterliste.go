package msescript

import (
	"context"
	"fmt"
	"strconv"
)

// **`filter_list` siebt eine Liste nach einer Bedingung** —
// #730.
//
//	SCRIPT_FUNCTION(filter_list) {
//	  ScriptCustomCollectionP ret(new ScriptCustomCollection());
//	  ScriptValueP it = input->makeIterator();
//	  while (ScriptValueP v = it->next()) {
//	    ctx.setVariable(SCRIPT_VAR_input, v);
//	    if (filter->eval(ctx)->toBool()) { ret->value.push_back(v); }
//	  }
//	  // TODO : somehow preserve keys
//	  return ret;
//	}
//
// `src/script/functions/basic.cpp`, Zeilen 574–588, Fassung f9d9356 vom
// 2021-07-06.
//
// # Wer es ruft
//
// Drei Stellen im Pack, und die wichtigste ist keine Vorlage:
// `magic.mse-game/script` Zeile 259 baut damit sein `includes()`, und das
// Spielskript belegt **jede** Ausfuhr vor. Dazu `magic-spoiler` Zeile 153 und
// `magic-cockatrice-v2` Zeile 945.
func init() {
	eintragen("filter_list", siebeListe)
}

// siebeListe ist `filter_list(liste, filter: {…})`.
//
// **Dieselbe Gestalt wie [ordneListe]** — ein Block je Glied, ausgewertet über
// der Umgebung des Aufrufers. Zwei Unterschiede, und beide stehen im Original:
//
//   - Der Rumpf gibt eine **Wahrheit** statt eines Schlüssels.
//   - **Nur `input` wird gesetzt.** `sort_list` hat die Verzweigung
//     `set ? _("card") : _("input")` (Zeile 512), `filter_list` nicht.
func siebeListe(ctx context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("filter_list"); err != nil {
		return Nichts, err
	}

	liste := ruf.Stellung[0]
	if liste.Art != ArtAbbildung {
		return Nichts, ruf.fehlerAn(
			fmt.Sprintf("filter_list siebt eine Liste, hier steht %s", liste.Art))
	}

	bedingung, benannt := ruf.Benannter("filter")
	if !benannt {
		return Nichts, ruf.fehlerAn("filter_list braucht ein Argument `filter:`")
	}

	// **Wer die eingebaute Funktion von Hand ruft, gibt keine Umgebung mit.**
	// Dieselbe Regel wie bei [schluesselVon].
	aussen := ruf.Umgebung
	if aussen == nil {
		aussen = Bibliothek()
	}

	gesiebt := Wert{Art: ArtAbbildung}

	for i := range liste.Eintraege {
		glied := liste.Eintraege[i].Wert

		umgebung := aussen.Darueber()
		umgebung.Setze("input", glied)

		haelt, err := gliedschluessel(ctx, bedingung, umgebung, ruf)
		if err != nil {
			return Nichts, err
		}

		// **Was sich nicht in eine Wahrheit umrechnen lässt, ist ein
		// Fehler.**
		//
		//	else { throw ScriptErrorConversion(value, typeName(), _TYPE_("boolean")); }
		//
		// `ScriptString::toBool`, `src/script/value.cpp`, Zeilen 345–352,
		// Fassung f9d9356 vom 2021-07-06 — und `filter_list` ruft
		// `filter->eval(ctx)->toBool()` ungeschützt (`basic.cpp`, Zeile 581).
		wahr, umwandlung := haelt.AlsWahrheitswert()
		if umwandlung != nil {
			return Nichts, ruf.fehlerAn("filter_list: " + umwandlung.Error())
		}

		if !wahr {
			continue
		}

		// **Die Namen werden neu vergeben** — `push_back` in eine frische
		// Sammlung, und der Kommentar im Original sagt selbst, dass die
		// Schlüssel liegen bleiben. Dieselbe Regel wie bei [ordneListe]: eine
		// Liste ist in MSE-Script eine Abbildung mit Zahlen als Namen, und wer
		// die alten mitnähme, ließe `gesiebt.0` ins Leere zeigen.
		gesiebt.Eintraege = append(gesiebt.Eintraege, Eintrag{
			Name: strconv.Itoa(len(gesiebt.Eintraege)),
			Wert: glied,
		})
	}

	return gesiebt, nil
}
