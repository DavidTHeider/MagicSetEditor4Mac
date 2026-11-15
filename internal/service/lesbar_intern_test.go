package service

import "testing"

// **Fremder Text in einer Meldung wird lesbar gemacht**
// (#675).
//
// Der Anlass steht über [lesbar]: über den ganzen Kanon stand eine Meldung
// da, die mitten im Satz abbrach, weil das genannte Zeichen ein
// Zeilenumbruch war.
func TestLesbarMachtEineMeldungLesbar(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		text     string
		erwartet string
	}{
		{
			name:     "gewöhnlicher Text steht in Anführungszeichen",
			text:     "Feuerball",
			erwartet: `"Feuerball"`,
		},
		// **Der Fall, der es ausgelöst hat.**
		{
			name:     "ein Zeilenumbruch wird ausgeschrieben",
			text:     "\n",
			erwartet: `"\n"`,
		},
		{
			name:     "ein Tabulator ebenso",
			text:     "a\tb",
			erwartet: `"a\tb"`,
		},
		// **Umlaute bleiben Umlaute.** `strconv.Quote` schriebe `ä` und
		// machte aus einer deutschen Meldung eine Bytefolge.
		{
			name:     "Umlaute bleiben stehen",
			text:     "Wächter",
			erwartet: `"Wächter"`,
		},
		{
			name:     "ein Anführungszeichen wird geschützt",
			text:     `a"b`,
			erwartet: `"a\"b"`,
		},
		{
			name:     "eine leere Stelle wird benannt",
			text:     "",
			erwartet: "eine leere Stelle",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if ist := lesbar(fall.text); ist != fall.erwartet {
				t.Errorf("lesbar(%q) = %s, erwartet %s", fall.text, ist, fall.erwartet)
			}
		})
	}
}

// **Ein langer Text wird gekürzt.**
//
// Ein Feldwert kann tausend Zeichen haben; eine Meldung, die ihn ganz
// enthält, liest niemand.
func TestLesbarKuerztLangenText(t *testing.T) {
	t.Parallel()

	lang := ""
	for range LaengeInMeldung + 20 {
		lang += "x"
	}

	gekuerzt := []rune(lesbar(lang))

	// Zwei Anführungszeichen, die Grenze und das Auslassungszeichen.
	if len(gekuerzt) != LaengeInMeldung+3 {
		t.Errorf("Länge = %d, erwartet %d", len(gekuerzt), LaengeInMeldung+3)
	}

	if gekuerzt[len(gekuerzt)-2] != '…' {
		t.Errorf("es fehlt das Auslassungszeichen: %s", string(gekuerzt))
	}
}
