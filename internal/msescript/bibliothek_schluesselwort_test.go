package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`expand_keywords` mit leerer Wortliste räumt die alten Erinnerungstexte weg
// und gibt den Rest zurück.**
//
// # Die Regel
//
//	String tagged = remove_keyword_tags(text);
//	// any keywords in database?
//	if (!root) return tagged;
//
// `KeywordDatabase::expand`, `src/data/keyword.cpp`, Zeilen 775–799, Fassung
// f9d9356 vom 2021-07-06. Und `remove_keyword_tags` (Zeilen 756–764) nimmt
// weg: den **Inhalt** von `<atom-reminder…>`, `<atom-keyword…>` und
// `<atom-kwpph>`, dazu die Auszeichnungen `<keyword-param…>` und `<param-…>`.
//
// # Warum das hier steht und die Wortliste nicht
//
// **Die Schlüsselwörter eines Spiels werden noch nicht gelesen.** Solange das
// so ist, ist die Datenbank leer — und für eine leere Datenbank ist genau
// dieses Verhalten das von MSE, nicht eine Näherung.
//
// Ohne die Funktion scheitert `text_filter` in `magic.mse-game` als Ganzes:
// sie steht mitten in einer Kette aus rund dreißig teilangewandten Funktionen,
// und die Kette entsteht beim Auswerten der Zuweisung
// (#345).
func TestExpandKeywordsRaeumtDieErinnerungstexteWeg(t *testing.T) {
	faelle := []struct {
		name, quelle, erwartet string
	}{
		{
			name:     "ein Erinnerungstext verschwindet samt Inhalt",
			quelle:   `expand_keywords(input: "Fliegend<atom-reminder> (kann nicht geblockt werden)</atom-reminder>.", combine: {""})`,
			erwartet: "Fliegend.",
		},
		{
			name:     "ein Platzhalter verschwindet samt Inhalt",
			quelle:   `expand_keywords(input: "a<atom-kwpph>xyz</atom-kwpph>b", combine: {""})`,
			erwartet: "ab",
		},
		{
			// **Eine Parameterauszeichnung verschwindet, ihr Inhalt bleibt.**
			// `remove_tag` gegen `remove_tag_contents` — MSE unterscheidet die
			// beiden hier ausdrücklich.
			name:     "eine Parameterauszeichnung verschwindet, der Inhalt bleibt",
			quelle:   `expand_keywords(input: "a<param-color>rot</param-color>b", combine: {""})`,
			erwartet: "arotb",
		},
		{
			name:     "gewöhnlicher Text bleibt",
			quelle:   `expand_keywords(input: "<i>Fliegend</i>", combine: {""})`,
			erwartet: "<i>Fliegend</i>",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Und in einer Kette reicht es seinen Text weiter.**
//
// Das ist der Fall, um den es geht: `text_filter` reiht
// `remove_tag@(…) + … + expand_keywords@(…) + …`.
func TestExpandKeywordsReihtSichInEineKette(t *testing.T) {
	// **Die Ersetzungen treffen `q` und `r`, nicht `a`** — sonst schriebe die
	// erste Stufe die Auszeichnung selbst um, und der Wächter prüfte die
	// Ersetzung statt die Kette.
	quelle := `(replace@(match:"q", replace:"r") + expand_keywords@(combine: {""}) + ` +
		`replace@(match:"r", replace:"s"))("q<atom-kwpph>x</atom-kwpph>")`

	wert, err := msescript.AuswertenQuelle(t.Context(), quelle, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "s" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "s")
	}
}

// **`combine:` ist Pflicht, und ein Argument braucht es auch.**
//
// MSE liest `combine` als `SCRIPT_PARAM` ohne Vorgabe (`basic.cpp`, Zeile
// 667). Eine Vorlage ohne die Angabe ist falsch geschrieben; ein
// stillschweigend durchgereichter Text verdeckte das.
func TestExpandKeywordsVerlangtSeineArgumente(t *testing.T) {
	faelle := []struct{ name, quelle string }{
		{"ohne combine", `expand_keywords(input: "a")`},
		{"ohne Text", `expand_keywords(combine: {""})`},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if _, err := msescript.AuswertenQuelle(
				t.Context(), fall.quelle, msescript.Bibliothek(),
			); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}
