package msescript

// **Drei Namen, die Set und Editor kennen — und die Sprache nicht** —
// Werkstück A von Phase 4.9
// (#1324).
//
// # Warum sie hier trotzdem stehen
//
// `keyword_usage` liest die Schlüsselwortstatistik einer Karte,
// `primary_choice` die Gruppen eines Wahlfelds, `new_card` legt eine Karte im
// Spiel an. Alles drei kennt ein **Blatt** nicht und darf es nicht kennen
// (§2) — `msescript` importiert nichts aus diesem Projekt außer nichts.
//
// **Der Weg dafür steht seit `expand_keywords` da**: die Bibliothek bringt
// einen **Stummel** mit, der tut, was ohne die fehlende Kenntnis richtig ist;
// `internal/service` überschreibt den Namen, wenn es mehr weiß. Zwei Wege
// wären zwei Verhalten — ein fehlender Name wäre eine Vorlage, die gar nicht
// läuft.
//
// **Das ist eine Anpassung** (§15, `adapted`), und sie ist hier begründet:
// dieselbe Funktion, dieselbe Aufschrift, ein anderer Ort für das Wissen.
//
// Alles nachgelesen in Fassung f9d9356 vom 2021-07-06.

import (
	"context"
	"strings"
)

func init() {
	eintragen("keyword_usage", schluesselwortgebrauch)
	eintragen("primary_choice", ersteWahl)
	eintragen("new_card", neueKarte)
}

// schluesselwortgebrauch ist `keyword_usage(card, unique: b)`.
//
//	SCRIPT_PARAM_C(CardP, card);
//	SCRIPT_OPTIONAL_PARAM_(bool, unique);
//	// make a list "kw1, kw2, kw3" of keywords used on card
//
// `src/script/functions/basic.cpp`, Zeilen 695–711, Fassung f9d9356 vom
// 2021-07-06. **Die Statistik entsteht
// beim Einsetzen der Schlüsselwörter** — MSE führt sie an der Karte mit.
//
// # Der Stummel gibt nichts zurück
//
// **Und das ist die richtige Antwort ohne Statistik**: „auf dieser Karte
// wurde kein Schlüsselwort gebraucht". Eine Vorlage, die daraus eine Liste
// baut, bekommt eine leere — und keine Fehlermeldung über einen Namen, den
// sie zu Recht ruft.
//
// **Wer mehr weiß, überschreibt den Namen**: `internal/service` setzt die
// Schlüsselwortnaht und kennt die gebrauchten Wörter.
func schluesselwortgebrauch(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("keyword_usage"); err != nil {
		return Nichts, err
	}

	return Text(""), nil
}

// ersteWahl ist `primary_choice(input)`.
//
//	ChoiceValueP value = dynamic_pointer_cast<ChoiceValue>(input);
//	int id = value->field().choices->choiceId(value->value);
//	// find the last group that still contains id
//	FOR_EACH_CONST_REVERSE(c, choices) {
//	  if (id >= c->first_id) { SCRIPT_RETURN(c->name); }
//	}
//
// Zeilen 195–210 in `src/script/functions/editor.cpp`, Fassung f9d9356 vom
// 2021-07-06. **Es gibt die
// Gruppe**, in der eine Wahl steht — `magic.mse-game` gruppiert etwa die
// Seltenheiten.
//
// # Der Stummel gibt die Wahl selbst
//
// **Ohne Gruppen ist eine Wahl ihre eigene Gruppe.** Das ist die einzige
// Antwort, die ohne die Felddefinition wahr ist — und sie ist brauchbar: eine
// Vorlage, die `primary_choice(x)` zeigt, zeigt dann `x`.
//
// **Und es ist eine Abweichung**, die hier steht, damit niemand sie für
// Parität hält: MSE **weist ab**, wenn die Eingabe kein Wahlwert ist
// („Argument to 'primary_choice' should be a choice value"). Hier kann das
// niemand prüfen — die Sprache kennt keinen Wahlwert.
func ersteWahl(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("primary_choice"); err != nil {
		return Nichts, err
	}

	return Text(ruf.Stellung[0].AlsText()), nil
}

// neueKarte ist `new_card(game: g, input: felder)`.
//
//	SCRIPT_PARAM(GameP, game);
//	CardP new_card = make_intrusive<Card>(*game);
//	SCRIPT_PARAM(ScriptValueP, input);
//	// set field values
//
// `src/script/functions/construction.cpp`, Zeilen 23–35, Fassung f9d9356
// vom 2021-07-06.
//
// # Was es hier gibt und was nicht
//
// **Eine Karte anlegen kann die Sprache nicht** — sie hat kein Set, in das
// sie eine legte, und ein Blatt, das eines hätte, wäre keines mehr (§2).
//
// **Sie beschreibt sie.** Herauskommt die Abbildung der Feldwerte, wie sie
// hereinkam; wer sie zu einer Karte macht, ist der Aufrufer. Das ist dieselbe
// Trennung wie beim Bildbaum: die Sprache baut den Plan, die Anwendung führt
// ihn aus.
//
// **`game:` wird gelesen und nicht gebraucht.** Es steht im Original, weil
// dort die Felder des Spiels die Karte formen; hier trägt die Beschreibung
// nur, was dasteht. Wer sie einsetzt, kennt sein Spiel ohnehin.
func neueKarte(_ context.Context, ruf Ruf) (Wert, error) {
	felder, da := ruf.Benannter("input")
	if !da && len(ruf.Stellung) > 0 {
		felder, da = ruf.Stellung[0], true
	}

	if !da {
		return Nichts, ruf.fehlerAn("new_card braucht die Felder als input:")
	}

	if felder.Art != ArtAbbildung {
		return Nichts, ruf.fehlerAn(
			"new_card braucht für input: eine Abbildung aus Feldnamen, hier " +
				"steht " + string(felder.Art))
	}

	beschrieben := Wert{
		Art:       ArtAbbildung,
		Eintraege: make([]Eintrag, 0, len(felder.Eintraege)),
	}

	for _, eintrag := range felder.Eintraege {
		beschrieben.Eintraege = append(beschrieben.Eintraege, Eintrag{
			Name: strings.TrimSpace(eintrag.Name),
			Wert: eintrag.Wert,
		})
	}

	return beschrieben, nil
}
