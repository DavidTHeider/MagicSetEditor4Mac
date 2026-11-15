// Wie eine Stelle in einem Befund heißt.
//
// **Eine eigene Datei, weil sie eine eigene Frage beantwortet** (§2): nicht
// *ob* etwas gültig ist — das steht in `gueltigkeit.go` —, sondern *wie* der
// Befund darüber seinen Ort nennt. Beides in einer Datei ging über die
// Zeilengrenze, als der Zweitblick über Phase 4.2 die Klasse „jeder Befund
// nennt sein Set" schloss
// (#1289).

package mtgjson

import "strings"

// ohneCodeUndNamen ist der Ort eines Sets, das sich nicht benennt.
//
// **An einer Stelle**, weil ihn zwei Wege setzen: der Schemaprüfer über
// [ortVon] und der Leser über [rohSet.ort]. Zwei Schreibweisen desselben
// Satzes liefen an dem Tag auseinander, an dem eine von ihnen umformuliert
// wurde.
const ohneCodeUndNamen = "(Set ohne Code und ohne Namen)"

// ortVon gibt die Stelle, an der ein Befund entstanden ist.
//
// **`code`, sonst der Schlüssel, sonst ein benannter Platzhalter.** Ein Set
// darf beides nicht haben — `{"":{}}` ist gültiges JSON —, und dann ist ein
// leerer Ort schlechter als ein hässlicher: er sagt gar nichts.
func ortVon(satz rohsatz) string {
	return benannt(textVon(satz.felder["code"]),
		benannt(satz.schluessel, ohneCodeUndNamen))
}

// benannt gibt den Text, wenn er etwas Lesbares trägt — sonst den Ersatz.
//
// **Leerraum ist kein Name.** Der erste Anlauf fragte `!= ""`, und ein Set mit
// dem Schlüssel `" "` kam damit als benannt durch: der Befund lautete
// `"  → "` und sagte weder, welches Set noch welches Feld gemeint war.
//
// Gefunden vom Fuzzlauf am 2026-08-30, nachdem der Zweitblick über Phase 4.2
// die **Klasse** angestrichen hatte — der Fall davor (`{"":{}}`) war behoben,
// dieser nicht. Das Korpuskorn liegt unter
// `testdata/fuzz/FuzzLiesInhalt/ebc337301d6f9ef2`.
func benannt(text, ersatz string) string {
	if strings.TrimSpace(text) == "" {
		return ersatz
	}

	return text
}

// feldOhneNamen ist das Feld, das keinen Namen trägt.
//
// **Ein JSON-Objekt darf einen leeren Schlüssel haben**, und eine beschädigte
// Datei nutzt das. Ein Befund über `""` nennt sonst nichts.
const feldOhneNamen = "(Feld ohne Namen)"
