package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`to_html` — Kartentext als HTML** —
// #723.
//
// # Woher jede Zeile stammt
//
// Aus dem Orakel (§15), nicht aus dem Lesen von `export.cpp`:
//
//	magicseteditor --cli Vanguard.mse-set -q --raw
//
// Der Quelltext daneben sagt, **warum** die Antwort so lautet; er sagt nicht,
// dass sie so lautet. Zwei der Zeilen hätte niemand geraten — die zweite hier
// und das ausbleibende Wiederöffnen im Stapeltest darunter.
func TestToHTML(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, quelle, erwartet string }{
		{
			name:     "Fett und kursiv gehen als solche durch",
			quelle:   `to_html("<b>fett</b> und <i>kursiv</i>")`,
			erwartet: "<b>fett</b> und <i>kursiv</i>",
		},

		// **Ein Kleinerzeichen ohne schließende Klammer frisst den Rest.**
		//
		//	size_t end = str.find_first_of(_('>'), start);
		//	return end == String::npos ? String::npos : end + 1;
		//
		// `skip_tag`, `src/util/tagged_string.cpp`, Zeilen 229–233, Fassung
		// f9d9356 vom 2021-07-06: ohne `>` gibt es `npos`, und die Schleife in
		// `to_html` endet daran.
		//
		// **Das ist kein Versehen dieser Umsetzung**, sondern das gemessene
		// Verhalten von MSE. Wer es „verbesserte", schriebe eine zweite
		// Semantik (§15, Satz 4).
		{
			name:     "ein Kleinerzeichen ohne Klammer verschluckt den Rest",
			quelle:   `to_html("x & y < z")`,
			erwartet: "x &amp; y ",
		},
		{
			name:     "ein Zeilenumbruch wird br und bleibt stehen",
			quelle:   `to_html("Zeile1\nZeile2")`,
			erwartet: "Zeile1<br>\nZeile2",
		},

		// **Alles über 0x7f wird eine Zahlenmarke.** `c >= 0x80` in
		// `export.cpp`, Zeile 252.
		{
			name:     "ein Umlaut wird eine Zahlenmarke",
			quelle:   `to_html("café")`,
			erwartet: "caf&#233;",
		},
		{
			name:     "größer, Anführung und Hochkomma bleiben roh",
			quelle:   `to_html("a>b" + "\"" + "c'd")`,
			erwartet: `a>b"c'd`,
		},

		// **Der Inhalt eines weichen Trenners fällt weg** —
		// `remove_tag_contents(str_in, "<sep-soft")` steht als erste Zeile in
		// `to_html` (`export.cpp`, Zeile 209).
		{
			name:     "der Inhalt eines weichen Trenners fällt weg",
			quelle:   `to_html("a<sep-soft>XY</sep-soft>b")`,
			erwartet: "ab",
		},

		// **Eine geschachtelte Marke zählt mit** — `tag_type_at` schneidet am
		// ersten `-` ab, `<sep-soft>` heißt darin `sep`, und ein `<sep>` darin
		// ist eine Ebene tiefer.
		{
			name:     "eine Schachtelung im weichen Trenner zählt mit",
			quelle:   `to_html("a<sep-soft>X<sep>Y</sep>Z</sep-soft>b")`,
			erwartet: "ab",
		},

		// **Ohne schließende Marke fällt alles weg** — auch das `a` davor.
		// MSE gibt zurück, was es bis dahin gesammelt hat, und das Stück vor
		// der Marke wird erst **nach** der Prüfung angehängt
		// (`tagged_string.cpp`, Zeilen 585–588).
		{
			name:     "ein offener weicher Trenner frisst alles",
			quelle:   `to_html("a<sep-soft>offen")`,
			erwartet: "",
		},

		// **Ein harter Trenner bleibt unberührt**: er heißt zwar auch `sep`,
		// gesucht wird aber nach der Zeichenkette `<sep-soft`.
		{
			name:     "ein harter Trenner behält seinen Inhalt",
			quelle:   `to_html("a<sep>X</sep>b")`,
			erwartet: "aXb",
		},

		// **Der Vergleich ist eine Vorsilbe.** `is_substr(str, i, "i")` trifft
		// jede Marke, die mit `i` anfängt — `<i-flavor>` wird `<i>`.
		{
			name:     "eine Aromamarke wird kursiv",
			quelle:   `to_html("<i-flavor>Aroma</i-flavor>")`,
			erwartet: "<i>Aroma</i>",
		},
		{
			name:     "eine unbekannte Marke fällt weg, ihr Inhalt bleibt",
			quelle:   `to_html("<kw-a>Fliegend</kw-a>")`,
			erwartet: "Fliegend",
		},
		{
			name:     "ein unbekanntes Paar fällt ganz weg",
			quelle:   `to_html("<foo>x</foo>")`,
			erwartet: "x",
		},

		// **Ein geschütztes Kleinerzeichen ist ein Kleinerzeichen.**
		// `untag_char` macht aus `\1` ein `<`, und **danach** wird maskiert.
		{
			name:     "ein geschütztes Kleinerzeichen wird maskiert",
			quelle:   "to_html(\"a\x01b\")",
			erwartet: "a&lt;b",
		},

		// **Und das Verbindungsleerzeichen wird ein Leerzeichen** — dieselbe
		// Zeile in `untag_char`, `src/util/tagged_string.cpp`, Zeilen 14–18,
		// Fassung f9d9356 vom 2021-07-06.
		{
			name:     "ein Verbindungsleerzeichen wird ein Leerzeichen",
			quelle:   "to_html(\"a\ueb00b\")",
			erwartet: "a b",
		},

		// **Ohne Symbolschrift bleibt der Symbolzug Text im span.**
		//
		//	if (symbol.opened > 0 && symbol_font) { symbols += c; }
		//
		// `export.cpp`, Zeilen 244–246: ohne Schrift geht das Zeichen den
		// gewöhnlichen Weg.
		{
			name:     "ein Symbolzug ohne Schrift wird ein span",
			quelle:   `to_html("Kosten: <sym>2WW</sym>!")`,
			erwartet: `Kosten: <span class="symbol">2WW</span>!`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Der Auszeichnungsstapel repariert die Verschachtelung** — und tut es an
// einer Stelle anders, als sein eigener Kommentar behauptet.
//
//	// situation was         <a><b><c>text
//	// situation will become <a><b><c>text</c></b><c>
//
// `TagStack::add`, `export.cpp`, Zeilen 133–152, Fassung f9d9356 vom
// 2021-07-06. Die Schleife, die das `<c>` wieder öffnen soll, läuft über
// `tags` — und `tags` ist zu diesem Zeitpunkt **schon** verkürzt. Sie ist tot,
// und das Wiederöffnen findet nicht statt.
//
// **Nachgemessen, nicht erschlossen**: `a<b>b<i>c</b>d</i>e` gibt
// `a<b>b<i>c</i></b>de`, und das `d` steht dort ohne Kursivierung. Hätte diese
// Umsetzung den Kommentar umgesetzt statt den Code, käme `<i>d</i>` heraus.
func TestDerAuszeichnungsstapel(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, quelle, erwartet string }{
		{
			name:     "eine kreuzende Schließung wird umsortiert",
			quelle:   `to_html("a<b>b<i>c</b>d</i>e")`,
			erwartet: "a<b>b<i>c</i></b>de",
		},
		{
			name:     "das Wiederöffnen findet nicht statt",
			quelle:   `to_html("<b><i>x</b>y</i>")`,
			erwartet: "<b><i>x</i></b>y",
		},
		{
			name:     "und ebenso andersherum",
			quelle:   `to_html("<i><b>a</i>b</b>")`,
			erwartet: "<i><b>a</b></i>b",
		},

		// **Zwei gleiche Marken hintereinander werden eine.** Die schließende
		// steht noch vor, die öffnende hebt sie auf.
		{
			name:     "eine Schließung hebt sich gegen die nächste Öffnung auf",
			quelle:   `to_html("<b>a</b><b>b</b>")`,
			erwartet: "<b>ab</b>",
		},

		// **Aber nicht über einen Text hinweg**: dazwischen wird die
		// Vormerkung geschrieben.
		{
			name:     "mit Text dazwischen bleiben es zwei",
			quelle:   `to_html("<b>a</b>c<b>b</b>")`,
			erwartet: "<b>a</b>c<b>b</b>",
		},
		{
			name:     "was offen bleibt, wird am Ende geschlossen",
			quelle:   `to_html("<b>offen")`,
			erwartet: "<b>offen</b>",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}
