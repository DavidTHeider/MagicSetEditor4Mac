package msescript

// **`check_spelling` — die Rechtschreibprüfung des Editors.**

import (
	"context"
	"regexp"
)

// markenmuster trifft die öffnende und die schließende Form einer
// Auszeichnung — dieselbe Form wie in `bibliothek_tag.go`.
var fehlermarke = regexp.MustCompile(`</?error-spelling[^>]*>`)

func init() {
	eintragen("check_spelling", ohneRechtschreibpruefung)
	eintragen("check_spelling_word", wortOhneRechtschreibpruefung)
}

// ohneRechtschreibpruefung ist `check_spelling`.
//
// # Was MSE tut
//
//	SCRIPT_FUNCTION(check_spelling) {
//	  …
//	  if (!settings.stylesheetSettingsFor(*stylesheet).card_spellcheck_enabled)
//	    SCRIPT_RETURN(input);
//	  …
//	  // remove old spelling error tags
//	  input = remove_tag(input, _("<error-spelling"));
//	  // no language -> spelling checking
//	  if (language.empty()) {
//	    SCRIPT_RETURN(input);
//	  }
//
// `src/script/functions/spelling.cpp`, Zeilen 57–71, Fassung f9d9356 vom
// 2021-07-06. Danach setzt es `<error-spelling:…>` um jedes Wort, das kein
// Wörterbuch kennt.
//
// # Was hier geschieht, und warum das kein Verlust ist
//
// **Die alten Marken werden entfernt, neue nicht gesetzt** — genau MSEs Weg
// für „keine Sprache" und „Prüfung aus". Ein Wörterbuch führt dieses Programm
// nicht.
//
// **Für das Bild ist es dasselbe.** Die Marken sind eine Auskunft an den
// Editor, und die Vorlagen nehmen sie vor dem Zeichnen wieder heraus:
// `magic.mse-game/script`, Zeile 1684, ruft
// `remove_tag@(tag: "<error-spelling")`. Ein Textzug, der sie trüge, sähe
// gezeichnet gleich aus.
//
// **Was verloren geht, gehört gesagt** (§15): eine Oberfläche, die falsch
// geschriebene Wörter unterringelt, gibt es damit nicht — und wenn sie kommt,
// ist dies die Stelle.
//
// **Kennzeichnung: adapted.** Das Verhalten ist für eine Plattform ohne
// Wörterbuch angepasst, nicht erfunden
// (#456).
func ohneRechtschreibpruefung(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("check_spelling"); err != nil {
		return Nichts, err
	}

	return Text(fehlermarke.ReplaceAllLiteralString(ruf.Stellung[0].AlsText(), "")), nil
}

// wortOhneRechtschreibpruefung ist `check_spelling_word`.
//
//	SCRIPT_FUNCTION(check_spelling_word) {
//	  SCRIPT_PARAM_C(String,language);
//	  SCRIPT_PARAM_C(String,input);
//	  if (language.empty()) {
//
// `src/script/functions/spelling.cpp`, Zeilen 133–136, Fassung f9d9356 vom
// 2021-07-06 — ohne Sprache gilt jedes Wort als richtig. Dasselbe hier, aus
// demselben Grund wie bei [ohneRechtschreibpruefung].
func wortOhneRechtschreibpruefung(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("check_spelling_word"); err != nil {
		return Nichts, err
	}

	return Wahrheit(true), nil
}
