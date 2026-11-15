package msescript_test

// **`curly_quotes` macht aus geraden Anführungszeichen runde**
// (#456).
//
//	String curly_quotes(String str, bool curl) {
//	  bool open = true, in_tag = false;
//	  for (wxUniCharRef c : str) {
//	    if (c == '\'' || c == LEFT_SINGLE_QUOTE || c == RIGHT_SINGLE_QUOTE) {
//	      c = curl ? (open ? LEFT_SINGLE_QUOTE : RIGHT_SINGLE_QUOTE) : '\'';
//	    } else if (c == '"' || …) { … }
//	    else if (c == '<') { in_tag = true; }
//	    else if (c == '>') { in_tag = false; }
//	    else if (!in_tag) { open = is_space_like(c); }
//	  }
//	  return str;
//	}
//
// `src/util/tagged_string.cpp`, Zeilen 763–780, Fassung f9d9356 vom
// 2021-07-06. `is_space_like` sind Leerraum, `(`, `[`, `{`, Halbgeviert- und
// Geviertstrich (ebd. Zeile 760).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestRundeAnfuehrungszeichen(t *testing.T) {
	for _, fall := range []struct{ name, quelle, erwartet string }{
		{
			name:     "am Wortanfang öffnend, am Wortende schließend",
			quelle:   `curly_quotes("er sagte \"ja\" und ging")`,
			erwartet: "er sagte “ja” und ging",
		},
		{
			name:     "der Apostroph mitten im Wort schließt",
			quelle:   `curly_quotes("don't")`,
			erwartet: "don’t",
		},
		{
			name:     "nach einer Klammer wird geöffnet",
			quelle:   `curly_quotes("(\"a\")")`,
			erwartet: "(“a”)",
		},
		{
			// **Innerhalb einer Auszeichnung ändert sich der Zustand nicht.**
			// `<i>` darf nicht als „Wortzeichen" gelten, sonst schlösse das
			// Anführungszeichen dahinter.
			name:     "eine Auszeichnung zählt nicht als Zeichen",
			quelle:   `curly_quotes("a <i>\"b\"</i>")`,
			erwartet: "a <i>“b”</i>",
		},
		{
			name:     "am Anfang wird geöffnet",
			quelle:   `curly_quotes("\"a\"")`,
			erwartet: "“a”",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **`check_spelling` nimmt die alten Fehlermarken heraus und setzt keine neuen**
// (#456).
//
// Genau MSEs Weg für „keine Sprache" und „Prüfung aus"
// (`src/script/functions/spelling.cpp`, Zeilen 62–71, Fassung f9d9356 vom
// 2021-07-06). Ein Wörterbuch führt dieses Programm nicht — was das kostet,
// steht an [msescript.ohneRechtschreibpruefung].
func TestDieRechtschreibpruefungNimmtIhreMarkenHeraus(t *testing.T) {
	for _, fall := range []struct{ name, quelle, erwartet string }{
		{
			name:     "alte Marken fallen weg",
			quelle:   `check_spelling("ein <error-spelling:en>Wart</error-spelling> mehr", language: "en")`,
			erwartet: "ein Wart mehr",
		},
		{
			name:     "andere Auszeichnungen bleiben",
			quelle:   `check_spelling("ein <i>Wort</i>")`,
			erwartet: "ein <i>Wort</i>",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Und `check_spelling_word` hält jedes Wort für richtig** — ohne Sprache tut
// MSE dasselbe (ebd. Zeilen 133–136).
func TestEinWortGiltAlsRichtig(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`check_spelling_word("Wart")`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "true" {
		t.Errorf("%q, erwartet true", wert.AlsText())
	}
}

// **Ohne Argument bleibt es ein Fehler** — die Gegenprobe zu den drei neuen
// Funktionen.
func TestDieNeuenFunktionenBrauchenIhreEingabe(t *testing.T) {
	for _, quelle := range []string{
		"curly_quotes()", "check_spelling()", "check_spelling_word()",
	} {
		t.Run(quelle, func(t *testing.T) {
			if _, err := msescript.AuswertenQuelle(t.Context(), quelle,
				msescript.Bibliothek()); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}
