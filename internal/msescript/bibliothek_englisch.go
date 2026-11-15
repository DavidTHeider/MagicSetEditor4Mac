package msescript

import (
	"context"
	"strconv"
	"strings"
)

// **Die englischen Sprachfunktionen**
// (#1068).
//
// Portiert aus `src/script/functions/english.cpp`, Fassung f9d9356 vom
// 2021-07-06. Im bezogenen `M15-Magic-Pack` werden sie **206-mal** gerufen —
// `english_number_a` allein 116-mal.
//
// # Der Anlass
//
// `process_english_hints` steht im `combine:` der Schlüsselwörter
// (`magic.mse-game/script`, Zeile 1718). Ohne die Funktion scheiterte das
// Zusammensetzen, und der Erinnerungstext fiel still auf den bloßen
// Kartentext zurück
// (#1066).

func init() {
	eintragen("english_number", englischeFunktion(englischesZahlwort))
	eintragen("english_number_a", englischeFunktion(zahlwortMitArtikel))
	eintragen("english_number_multiple", englischeFunktion(zahlwortMehrfach))
	eintragen("english_number_ordinal", englischeFunktion(ordnungszahl))
	eintragen("english_singular", worteFunktion(einzahl))
	eintragen("english_plural", worteFunktion(mehrzahl))
}

// englischeZahlwoerter sind die Namen, die sich nicht aus einer Regel ergeben.
//
// **Die Tabelle ist lückenhaft, und das ist Absicht** (Zeilen 29–51): `14`
// steht nicht darin, weil `englischesZahlwort(4) + "teen"` es ergibt; `60` nicht, weil
// `englischesZahlwort(6) + "ty"` es ergibt. Wer sie abschreibt und die Regeln wegläßt,
// bekommt für 14 eine Ziffer.
var zahlwoerter = map[int]string{
	0: "zero", 1: "one", 2: "two", 3: "three", 4: "four", 5: "five",
	6: "six", 7: "seven", 8: "eight", 9: "nine", 10: "ten",
	11: "eleven", 12: "twelve", 13: "thirteen", 15: "fifteen", 18: "eighteen",
	20: "twenty", 30: "thirty", 40: "forty", 50: "fifty", 80: "eighty",
}

// englischesZahlwort ist `english_number` (Zeilen 28–65).
func englischesZahlwort(zahl int) string {
	if wort, da := zahlwoerter[zahl]; da {
		return wort
	}

	switch {
	case zahl < 0 || zahl >= 100:
		// **Zu groß: die Ziffern bleiben** („number too large, keep as digits").
		return strconv.Itoa(zahl)

	case zahl < 20:
		return englischesZahlwort(zahl%10) + "teen"

	case zahl%10 == 0:
		return englischesZahlwort(zahl/10) + "ty"

	default:
		return englischesZahlwort(zahl/10*10) + "-" + englischesZahlwort(zahl%10)
	}
}

// ordnungszahlen sind die unregelmäßigen Ordnungszahlen (Zeilen 69–78).
var ordnungszahlen = map[int]string{
	1: "first", 2: "second", 3: "third", 5: "fifth", 8: "eighth",
	9: "ninth", 11: "eleventh", 12: "twelfth",
}

// ordnungszahl ist `english_ordinal` (Zeilen 68–92).
func ordnungszahl(zahl int) string {
	if wort, da := ordnungszahlen[zahl]; da {
		return wort
	}

	switch {
	case zahl <= 0 || zahl >= 100:
		return strconv.Itoa(zahl) + "th"

	case zahl < 20:
		return englischesZahlwort(zahl) + "th"

	case zahl%10 == 0:
		// **`twenty` verliert sein `y`** — daher `twentieth`.
		wort := englischesZahlwort(zahl)

		return wort[:len(wort)-1] + "ieth"

	default:
		return englischesZahlwort(zahl/10*10) + "-" + ordnungszahl(zahl%10)
	}
}

// zahlwortMitArtikel ist `english_number_a`: `a` für die Eins.
func zahlwortMitArtikel(zahl int) string {
	if zahl == 1 {
		return "a"
	}

	return englischesZahlwort(zahl)
}

// zahlwortMehrfach ist `english_number_multiple`: nichts für die Eins.
func zahlwortMehrfach(zahl int) string {
	if zahl == 1 {
		return ""
	}

	return englischesZahlwort(zahl)
}

// einzahl ist `english_singular` (Zeilen 155–175).
func einzahl(wort string) string {
	switch {
	case len(wort) > 3 && strings.HasSuffix(wort, "ies"):
		return wort[:len(wort)-3] + "y"

	case len(wort) > 3 && strings.HasSuffix(wort, "oes"),
		len(wort) > 4 && strings.HasSuffix(wort, "ches"),
		len(wort) > 4 && strings.HasSuffix(wort, "shes"),
		len(wort) > 4 && strings.HasSuffix(wort, "sses"):
		return wort[:len(wort)-2]

	case len(wort) > 5 && strings.HasSuffix(wort, "ves") &&
		(hatAn(wort, len(wort)-5, "el") || hatAn(wort, len(wort)-5, "ar")):
		return wort[:len(wort)-3] + "f"

	case len(wort) > 1 && strings.HasSuffix(wort, "s"):
		return wort[:len(wort)-1]

	case len(wort) >= 3 && strings.HasSuffix(wort, "men"):
		return wort[:len(wort)-2] + "an"

	default:
		return wort
	}
}

// mehrzahl ist `english_plural` (Zeilen 176–192).
func mehrzahl(wort string) string {
	if len(wort) <= 2 {
		return wort + "s"
	}

	vorletzter, letzter := wort[len(wort)-2], wort[len(wort)-1]

	switch {
	case letzter == 'y' && !istVokal(vorletzter):
		return wort[:len(wort)-1] + "ies"

	case letzter == 'o' && !istVokal(vorletzter),
		(vorletzter == 's' || vorletzter == 'c') && letzter == 'h',
		letzter == 's':
		return wort + "es"

	default:
		return wort + "s"
	}
}

// istVokal ist `is_vowel` (Zeilen 16–19) — **nur a, e, i, o, u**, in beiden
// Schreibungen. `y` gehört nicht dazu, und daran hängt `pony` → `ponies`.
func istVokal(zeichen byte) bool {
	return strings.IndexByte("aeiouAEIOU", zeichen) >= 0
}

// hatAn sagt, ob an einer Stelle eine Zeichenfolge steht.
func hatAn(wort string, stelle int, teil string) bool {
	return stelle >= 0 && stelle+len(teil) <= len(wort) && wort[stelle:stelle+len(teil)] == teil
}

// englischeFunktion baut eine eingebaute Funktion über einer Zahlregel.
//
// # Der Parameterfall
//
//	// a keyword parameter, of the form "<param->123</param->"
//
// `do_english_num`, Zeilen 105–133: steht am Anfang ein `<param-…>`, wird der
// Wert **darin** ersetzt, und davor kommt `<hint-1>` für die Eins, sonst
// `<hint-2>`. Was sich nicht als Zahl lesen lässt, gilt als Mehrzahl.
//
// **Der Hinweis ist kein Schmuck.** [ProcessEnglischeHinweise] entscheidet an
// ihm, ob `(s)` stehen bleibt oder fällt — ohne ihn stünde auf jeder Karte die
// Mehrzahl.
func englischeFunktion(regel func(int) string) Eingebaut {
	return func(_ context.Context, ruf Ruf) (Wert, error) {
		if err := ruf.eineEingabe("english_number"); err != nil {
			return Nichts, err
		}

		eingabe := ruf.Stellung[0].AlsText()

		wert, davor, ende, istParameter := parameterwert(eingabe)
		if !istParameter {
			zahl, istZahl := ganzzahlAus(eingabe)
			if !istZahl {
				// **Was keine Zahl ist, kommt zurück, wie es kam** —
				// `input.ToLong(&i)` scheitert, und MSE gibt die Eingabe
				// weiter (Zeile 132).
				return Text(eingabe), nil
			}

			return Text(regel(zahl)), nil
		}

		zahl, istZahl := ganzzahlAus(wert)
		if !istZahl {
			// **Ein unlesbarer Parameterwert gilt als Mehrzahl** (Zeile 128).
			return Text("<hint-2>" + eingabe), nil
		}

		hinweis := "<hint-2>"
		if zahl == 1 {
			hinweis = "<hint-1>"
		}

		return Text(hinweis + eingabe[:davor] + regel(zahl) + eingabe[ende:]), nil
	}
}

// worteFunktion ist dasselbe für Einzahl und Mehrzahl — **ohne Hinweis**.
//
// `do_english`, Zeilen 194–210: derselbe Parameterfall, aber kein `<hint-…>`.
func worteFunktion(regel func(string) string) Eingebaut {
	return func(_ context.Context, ruf Ruf) (Wert, error) {
		if err := ruf.eineEingabe("english_singular"); err != nil {
			return Nichts, err
		}

		eingabe := ruf.Stellung[0].AlsText()

		wert, davor, ende, istParameter := parameterwert(eingabe)
		if !istParameter {
			return Text(regel(eingabe)), nil
		}

		return Text(eingabe[:davor] + regel(wert) + eingabe[ende:]), nil
	}
}

// parameterwert gibt den Wert in einem führenden `<param-…>`.
//
// **Nur am Anfang** — `is_substr(input, 0, "<param-")`. Steht die Auszeichnung
// woanders, ist der ganze Text die Eingabe.
func parameterwert(eingabe string) (wert string, davor, ende int, istParameter bool) {
	if !strings.HasPrefix(eingabe, "<param-") {
		return "", 0, 0, false
	}

	zu := strings.IndexByte(eingabe, '>')
	if zu < 0 {
		return "", 0, 0, false
	}

	naechste := strings.IndexByte(eingabe[zu+1:], '<')
	if naechste < 0 {
		return "", 0, 0, false
	}

	return eingabe[zu+1 : zu+1+naechste], zu + 1, zu + 1 + naechste, true
}

// ganzzahlAus liest eine ganze Zahl — oder sagt, dass es keine ist.
//
// **Kein Fehlerwert, sondern eine Frage.** Ob eine Eingabe eine Zahl ist,
// entscheidet hier den Weg und ist kein Scheitern; `nilerr` hätte sonst recht.
func ganzzahlAus(text string) (int, bool) {
	zahl, fehler := strconv.Atoi(strings.TrimSpace(text))

	return zahl, fehler == nil
}
