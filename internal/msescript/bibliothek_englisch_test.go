package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Die englischen Zahlwörter** —
// #1068.
//
// Portiert aus `src/script/functions/english.cpp`, Zeilen 28–100, Fassung
// f9d9356 vom 2021-07-06. Die Tabelle dort ist lückenhaft, und **das ist
// Absicht**: `14` steht nicht darin, weil `english_number(4) + "teen"` es
// ergibt; `60` nicht, weil `english_number(6) + "ty"` es ergibt. Wer die
// Tabelle abschreibt und die Regeln wegläßt, bekommt für 14 eine Zahl.
//
// **Im bezogenen Pack 206 Aufrufe** über die sieben Funktionen dieser Datei —
// `english_number_a` allein 116-mal.
func TestDieEnglischenZahlwoerter(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ hinein, heraus string }{
		{"0", "zero"},
		{"1", "one"},
		{"12", "twelve"},
		{"13", "thirteen"},
		// **Aus der Regel, nicht aus der Tabelle.**
		{"14", "fourteen"},
		{"16", "sixteen"},
		{"17", "seventeen"},
		{"19", "nineteen"},
		{"20", "twenty"},
		{"60", "sixty"},
		{"70", "seventy"},
		{"90", "ninety"},
		{"21", "twenty-one"},
		{"99", "ninety-nine"},
		// **Zu groß oder negativ: die Ziffern bleiben.**
		{"100", "100"},
		{"-3", "-3"},
		// **Was keine Zahl ist, bleibt, wie es kam.**
		{"drei", "drei"},
	} {
		pruefeEnglisch(t, "english_number", fall.hinein, fall.heraus)
	}
}

// **Die Ordnungszahlen.**
//
// `english_ordinal`, Zeilen 68–92. `twenty` verliert sein `y` und bekommt
// `ieth` — daher `twentieth` und nicht `twentyth`.
func TestDieEnglischenOrdnungszahlen(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ hinein, heraus string }{
		{"1", "first"},
		{"2", "second"},
		{"3", "third"},
		{"5", "fifth"},
		{"4", "fourth"},
		{"13", "thirteenth"},
		{"20", "twentieth"},
		{"40", "fortieth"},
		{"21", "twenty-first"},
		{"23", "twenty-third"},
		{"0", "0th"},
		{"100", "100th"},
	} {
		pruefeEnglisch(t, "english_number_ordinal", fall.hinein, fall.heraus)
	}
}

// **`a` für die Eins — und nichts für die Eins.**
func TestDieEnglischenZahlwoerterMitArtikel(t *testing.T) {
	t.Parallel()

	pruefeEnglisch(t, "english_number_a", "1", "a")
	pruefeEnglisch(t, "english_number_a", "2", "two")
	pruefeEnglisch(t, "english_number_multiple", "1", "")
	pruefeEnglisch(t, "english_number_multiple", "2", "two")
}

// **Einzahl und Mehrzahl.**
//
// `english_singular` und `english_plural`, Zeilen 155–192.
func TestDieEnglischeEinzahlUndMehrzahl(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ hinein, heraus string }{
		{"ponies", "pony"},
		{"heroes", "hero"},
		{"witches", "witch"},
		{"wishes", "wish"},
		{"glasses", "glass"},
		// **`shelves` wird `shelf`, `elves` aber `elve`** — die Regel verlangt
		// `size > 5`, und `elves` hat genau fünf Zeichen (`english.cpp`,
		// Zeile 166, Fassung f9d9356 vom 2021-07-06). Das ist das gemessene
		// Verhalten des Originals; wer es „verbessert", weicht ab (§15.4).
		{"shelves", "shelf"},
		{"elves", "elve"},
		{"cards", "card"},
		{"women", "woman"},
		{"card", "card"},
	} {
		pruefeEnglisch(t, "english_singular", fall.hinein, fall.heraus)
	}

	for _, fall := range []struct{ hinein, heraus string }{
		{"pony", "ponies"},
		{"hero", "heroes"},
		{"witch", "witches"},
		{"wish", "wishes"},
		{"glass", "glasses"},
		{"card", "cards"},
		{"boy", "boys"},
		{"radio", "radios"},
	} {
		pruefeEnglisch(t, "english_plural", fall.hinein, fall.heraus)
	}
}

// **Ein Schlüsselwortparameter trägt seinen Hinweis.**
//
//	// a keyword parameter, of the form "<param->123</param->"
//
// `do_english_num`, Zeilen 105–133: der Wert **im** der Auszeichnung wird
// ersetzt, und davor kommt `<hint-1>` für die Eins, sonst `<hint-2>`.
func TestEinSchluesselwortparameterTraegtSeinenHinweis(t *testing.T) {
	t.Parallel()

	pruefeEnglisch(t, "english_number",
		"<param-x>1</param-x>", "<hint-1><param-x>one</param-x>")
	pruefeEnglisch(t, "english_number",
		"<param-x>3</param-x>", "<hint-2><param-x>three</param-x>")

	// **Was sich nicht als Zahl lesen lässt, gilt als Mehrzahl.**
	pruefeEnglisch(t, "english_number",
		"<param-x>viel</param-x>", "<hint-2><param-x>viel</param-x>")

	// **Ohne Hinweis bei Einzahl und Mehrzahl** — `do_english` setzt keinen.
	pruefeEnglisch(t, "english_plural",
		"<param-x>card</param-x>", "<param-x>cards</param-x>")
}

// pruefeEnglisch ruft eine Funktion und vergleicht.
func pruefeEnglisch(t *testing.T, name, hinein, heraus string) {
	t.Helper()

	wert, fehler := msescript.AuswertenQuelle(t.Context(),
		name+`(input: `+alsSkripttext(hinein)+`)`, msescript.Bibliothek())
	if fehler != nil {
		t.Fatalf("%s(%q): %v", name, hinein, fehler)
	}

	if wert.AlsText() != heraus {
		t.Errorf("%s(%q) = %q, erwartet %q", name, hinein, wert.AlsText(), heraus)
	}
}

// alsSkripttext setzt einen Text in Anführungszeichen für ein Skript.
//
// **Nicht `alsQuelltext`**: den Namen führt schon `fremdpakete_test.go`, und
// zwei gleiche Namen in einem Testpaket übersetzen sich nicht — gesehen erst
// im Lauf hinter dem Build-Tag, weil die Datei nur dort mitkommt.
func alsSkripttext(text string) string { return `"` + text + `"` }

// **Die Randfälle der englischen Funktionen.**
//
// Sie stehen zusammen, weil sie dieselbe Frage beantworten: was geschieht mit
// einer Eingabe, die nicht der Regelfall ist? Die Antwort ist überall
// dieselbe — sie kommt zurück, wie sie kam (§8).
func TestDieRandfaelleDerEnglischenFunktionen(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, funktion, hinein, heraus string }{
		// **Ein `<param-…>` ohne Ende ist kein Parameter.**
		{"ohne schließende Klammer", "english_number", "<param-x", "<param-x"},
		{"ohne folgende Auszeichnung", "english_number", "<param-x>3", "<param-x>3"},
		// **Zwei Zeichen oder weniger bekommen einfach ein `s`.**
		{"ein kurzes Wort", "english_plural", "ax", "axs"},
		{"ein Buchstabe", "english_plural", "a", "as"},
		// **Einzahl von etwas, das schon Einzahl ist.**
		{"kein Plural", "english_singular", "ox", "ox"},
		{"nur ein Zeichen", "english_singular", "s", "s"},
		// **Wörter auf `-o` nach einem Vokal bekommen nur ein `s`.**
		{"Vokal vor o", "english_plural", "radio", "radios"},
		// **Und die Ordnungszahl der Vierzig.**
		{"vierzigste", "english_number_ordinal", "40", "fortieth"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			pruefeEnglisch(t, fall.funktion, fall.hinein, fall.heraus)
		})
	}
}

// **Ohne Eingabe wird abgewiesen.**
func TestDieEnglischenFunktionenBrauchenEineEingabe(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"english_number", "english_singular", "process_english_hints",
	} {
		if _, fehler := msescript.AuswertenQuelle(t.Context(),
			name+`()`, msescript.Bibliothek()); fehler == nil {
			t.Errorf("%s ohne Eingabe blieb ohne Fehler", name)
		}
	}
}
