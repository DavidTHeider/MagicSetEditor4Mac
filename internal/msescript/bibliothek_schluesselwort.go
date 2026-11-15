package msescript

import (
	"context"
	"regexp"
)

// **`expand_keywords` setzt Erinnerungstexte in den Kartentext.**
//
// # Was hier gebaut ist und was nicht
//
// Gebaut ist der Fall der **leeren Wortliste**:
//
//	String tagged = remove_keyword_tags(text);
//	// any keywords in database?
//	if (!root) return tagged;
//
// `KeywordDatabase::expand`, `src/data/keyword.cpp`, Zeilen 775–799, Fassung
// f9d9356 vom 2021-07-06. Die Schlüsselwörter eines Spiels werden von diesem
// Programm **noch nicht gelesen**; solange das so ist, ist die Datenbank leer,
// und dann ist genau dieses Verhalten das von MSE — keine Näherung.
//
// # Warum die Funktion trotzdem dastehen muss
//
// `text_filter` in `magic.mse-game` reiht rund dreißig teilangewandte
// Funktionen mit `+`, und `expand_keywords@(…)` steht mittendrin. Die Kette
// entsteht beim **Auswerten der Zuweisung**; fehlt ein Name darin, scheitert
// sie als Ganzes — und mit ihr die 72 Kartenfelder, die sie rufen
// (#345).
//
// # Was eine echte Wortliste bräuchte
//
// Die `keyword:`-Blöcke des Spielpakets samt ihren Parametertypen, einen
// Präfixbaum über die Wortanfänge, den Abgleich gegen den entlarvten Text und
// das Einsetzen über `combine:`. Das ist eine eigene Aufgabe und steht als
// eigenes Issue.
func init() {
	eintragen("expand_keywords", schluesselwoerterEinsetzen)
}

// schluesselwoerterEinsetzen ist `expand_keywords` mit leerer Wortliste.
func schluesselwoerterEinsetzen(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("expand_keywords"); err != nil {
		return Nichts, err
	}

	// **`combine:` ist Pflicht** — MSE liest es als `SCRIPT_PARAM`, ohne
	// Vorgabe (Zeile 667). Eine Vorlage ohne die Angabe ist falsch
	// geschrieben, und ein stillschweigend durchgereichter Text verdeckte das.
	if _, da := ruf.Benannter("combine"); !da {
		return Nichts, ruf.fehlerAn("expand_keywords braucht ein Argument `combine:`")
	}

	return Text(ohneSchluesselwortauszeichnungen(ruf.Stellung[0].AlsText())), nil
}

// ohneSchluesselwortauszeichnungen ist `remove_keyword_tags`.
//
//	String s = remove_tag_contents(tagged_str, "<atom-reminder");
//	s = remove_tag_contents(s, "<atom-keyword");
//	s = remove_tag_contents(s, "<atom-kwpph>");
//	s = remove_tag(s, "<keyword-param");
//	s = remove_tag(s, "<param-");
//	return s;
//
// `src/data/keyword.cpp`, Zeilen 756–764.
//
// **Die ersten drei nehmen den Inhalt mit, die letzten beiden nicht.** Ein
// Erinnerungstext ist etwas, das MSE selbst hineingeschrieben hat und beim
// nächsten Mal neu schreibt; ein Parameterwert ist der Text des Anwenders und
// bleibt stehen.
//
// Fassung f9d9356 vom 2021-07-06.
func ohneSchluesselwortauszeichnungen(text string) string {
	for _, name := range []string{"atom-reminder", "atom-keyword", "atom-kwpph"} {
		text = ohneAuszeichnungssamtInhalt(text, name)
	}

	for _, name := range []string{"keyword-param", "param-"} {
		text = ohneAuszeichnungsklammern(text, name)
	}

	return text
}

// ohneAuszeichnungssamtInhalt wirft eine Auszeichnung mit ihrem Inhalt weg.
//
// **Ohne Verschachtelung.** MSEs `remove_tag_contents` sucht das passende Ende
// über `match_close_tag`; ein Erinnerungstext trägt keine gleichnamige
// Auszeichnung in sich, und ein Muster ohne Rückverfolgung reicht. Wo es
// anders käme, bliebe der Rest stehen statt zu viel wegzufallen.
func ohneAuszeichnungssamtInhalt(text, name string) string {
	muster := gemerkterAusdruck(
		`<` + regexp.QuoteMeta(name) + `[^>]*>.*?</` + regexp.QuoteMeta(name) + `[^>]*>`)

	return muster.ReplaceAllLiteralString(text, "")
}

// ohneAuszeichnungsklammern wirft die Klammern weg und lässt den Inhalt stehen.
func ohneAuszeichnungsklammern(text, name string) string {
	muster := gemerkterAusdruck(`</?` + regexp.QuoteMeta(name) + `[^>]*>`)

	return muster.ReplaceAllLiteralString(text, "")
}
