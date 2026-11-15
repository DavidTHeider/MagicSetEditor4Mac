package msescript

import (
	"context"
	"strings"
)

// **`process_english_hints` löst Einzahl und Mehrzahl auf**
// (#1068).
//
// Portiert aus `src/script/functions/english.cpp`, Zeilen 224–326, Fassung
// f9d9356 vom 2021-07-06. Der Quelltext erklärt sich dort selbst:
//
//   - `"<hint-1>xxx(yyy)zzz"` → `"xxxzzz"` (Einzahl)
//   - `"<hint-2>xxx(yyy)zzz"` → `"xxxyyyzzz"` (Mehrzahl)
//   - `"[^., ]a <param-..>[aeiou]"` → `"\1 an \2"` (der Artikel `an`)
//   - `"<hint-?>"` → `""`
//
// # Warum es diese Funktion braucht
//
// Sie steht im `combine:` der Schlüsselwörter (`magic.mse-game/script`,
// Zeile 1718). Ohne sie scheitert das Zusammensetzen, und der Erinnerungstext
// fällt still auf den bloßen Kartentext zurück
// (#1066).

func init() {
	eintragen("process_english_hints", englischeHinweise)
}

// englischeHinweise ist die eingebaute Funktion.
func englischeHinweise(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("process_english_hints"); err != nil {
		return Nichts, err
	}

	return Text(HinweiseAufloesen(ruf.Stellung[0].AlsText())), nil
}

// HinweiseAufloesen ist `process_english_hints` (Zeilen 237–326).
//
// **Ein Durchgang von links nach rechts**, wie im Original: der Zustand ist
// die zuletzt gesehene Hinweiszahl, und sie gilt bis zur nächsten Klammer oder
// `<singular>`/`<plural>`.
func HinweiseAufloesen(text string) string {
	var gebaut strings.Builder

	// einOderMehr ist 0 (kein Hinweis), 1 (Einzahl) oder 2 (Mehrzahl).
	einOderMehr := 0

	for stelle := 0; stelle < len(text); {
		switch {
		case strings.HasPrefix(text[stelle:], "<hint-"):
			if stelle+6 < len(text) {
				switch text[stelle+6] {
				case '1':
					einOderMehr = 1
				case '2':
					einOderMehr = 2
				}
			}

			stelle = nachAuszeichnung(text, stelle)

		case strings.HasPrefix(text[stelle:], "<param-"):
			stelle = parameterhinweis(text, stelle, einOderMehr, &gebaut)

		case strings.HasPrefix(text[stelle:], "<singular>"):
			stelle = zweigNehmen(text, stelle, "singular", einOderMehr == 1, &gebaut)
			einOderMehr = 0

		case strings.HasPrefix(text[stelle:], "<plural>"):
			stelle = zweigNehmen(text, stelle, "plural", einOderMehr == 2, &gebaut)
			einOderMehr = 0

		case text[stelle] == '(' && einOderMehr != 0:
			stelle = klammerAufloesen(text, stelle, einOderMehr == 2, &gebaut)
			einOderMehr = 0

		case text[stelle] == '<':
			nach := nachAuszeichnung(text, stelle)
			gebaut.WriteString(text[stelle:nach])
			stelle = nach

		default:
			gebaut.WriteByte(text[stelle])
			stelle++
		}
	}

	return gebaut.String()
}

// parameterhinweis behandelt ein `<param-…>` — vor allem den Artikel.
//
// **`a` wird `an` vor einem Vokal** (Zeilen 268–283). Geprüft wird am schon
// **gebauten** Text: steht dort `a ` am Ende und beginnt der Parameterwert mit
// einem Vokal, wächst das `a` um ein `n`.
//
// **Und ein leerer Parameter frißt das Leerzeichen davor** (Zeilen 284–288) —
// sonst stünde in „two ⟨leer⟩cards" ein doppelter Abstand.
func parameterhinweis(text string, stelle, einOderMehr int, gebaut *strings.Builder) int {
	nach := nachAuszeichnung(text, stelle)

	switch {
	case einOderMehr == 1 && strings.HasPrefix(text[nach:], "a</param-"):
		// **Der Sonderfall aus `english_number_a(1)`**: der Parameter *ist*
		// das `a`, und dahinter steht das Wort.
		nachEnde := nachAuszeichnung(text, nach+1)
		if nachEnde+1 < len(text) &&
			text[nachEnde] == ' ' && istVokal(text[nachEnde+1]) {
			gebaut.WriteString(text[stelle:nach])
			gebaut.WriteString("an")

			return nach + 1
		}

	case nach < len(text) && istVokal(text[nach]):
		gebautText := gebaut.String()
		if artikelWaechst(gebautText) {
			zurueck := gebaut.String()
			gebaut.Reset()
			gebaut.WriteString(zurueck[:len(zurueck)-1] + "n" + zurueck[len(zurueck)-1:])
		}

	case strings.HasPrefix(text[nach:], "</param-"):
		gebautText := gebaut.String()
		if strings.HasSuffix(gebautText, " ") {
			gebaut.Reset()
			gebaut.WriteString(gebautText[:len(gebautText)-1])
		}
	}

	gebaut.WriteString(text[stelle:min(nach, len(text))])

	return nach
}

// artikelWaechst sagt, ob am Ende ein alleinstehendes `a ` steht.
//
// „is there \"a\" before this?" — Zeilen 275–281: das `a` darf kein Teil eines
// Wortes sein.
func artikelWaechst(gebaut string) bool {
	if len(gebaut) < 2 {
		return false
	}

	letzte := gebaut[len(gebaut)-2:]
	if letzte != "a " && letzte != "A " {
		return false
	}

	return len(gebaut) == 2 || !istBuchstabe(gebaut[len(gebaut)-3])
}

// zweigNehmen nimmt den Inhalt von `<singular>` oder `<plural>` — oder wirft
// ihn weg (Zeilen 289–302).
func zweigNehmen(text string, stelle int, name string, nehmen bool, gebaut *strings.Builder) int {
	anfang := nachAuszeichnung(text, stelle)

	ende := strings.Index(text[anfang:], "</"+name+">")
	if ende < 0 {
		return anfang
	}

	if nehmen {
		gebaut.WriteString(text[anfang : anfang+ende])
	}

	return anfang + ende + len("</"+name+">")
}

// klammerAufloesen behandelt `(…)` hinter einem Hinweis (Zeilen 303–315).
func klammerAufloesen(text string, stelle int, nehmen bool, gebaut *strings.Builder) int {
	ende := strings.IndexByte(text[stelle:], ')')
	if ende < 0 {
		gebaut.WriteByte(text[stelle])

		return stelle + 1
	}

	if nehmen {
		gebaut.WriteString(text[stelle+1 : stelle+ende])
	}

	return stelle + ende + 1
}

// nachAuszeichnung gibt die Stelle hinter einer Auszeichnung.
func nachAuszeichnung(text string, stelle int) int {
	zu := strings.IndexByte(text[stelle:], '>')
	if zu < 0 {
		return len(text)
	}

	return stelle + zu + 1
}

// istBuchstabe ist `isAlpha` für die ASCII-Buchstaben.
func istBuchstabe(zeichen byte) bool {
	return (zeichen >= 'a' && zeichen <= 'z') || (zeichen >= 'A' && zeichen <= 'Z')
}
