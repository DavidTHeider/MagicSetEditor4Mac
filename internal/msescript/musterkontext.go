package msescript

import "strings"

// MitKontext gibt dieses Muster mit `in_context:` — ein Treffer zählt nur,
// wenn der Kontext in „Text davor + `<match>` + Text danach" trifft. Ein
// Treffer, der nicht zählt, wird übersprungen, und gesucht wird hinter ihm
// weiter.
//
// Portiert aus `ScriptRegex::matches`, `src/script/functions/regex.cpp`,
// Zeilen 29–51, Fassung f9d9356 vom 2021-07-06 (ported, #2315). Dort nehmen
// ihn `replace`, `filter_text` und `break_text` (ebd. Zeilen 113–208).
//
// **Eine Kopie und nicht das Muster selbst**: [UebersetzeMuster] gibt Muster
// aus einem Speicher her, und dasselbe Muster steht anderswo ohne Kontext.
func (m *Muster) MitKontext(kontext *Muster) *Muster {
	kopie := *m
	kopie.kontext = kontext

	return &kopie
}

// ersetzeImKontext ist [Muster.Ersetze] mit Kontext.
func (m *Muster) ersetzeImKontext(text, vorlage string) (string, error) {
	var gebaut strings.Builder

	zeichen := []rune(text)
	gelesen := 0

	err := m.jederTrefferMitNeustart(text, func(index, laenge int, gruppen []string) {
		gebaut.WriteString(string(zeichen[gelesen:index]))
		gebaut.WriteString(fuelle(vorlage, func(nummer int) string {
			if nummer < len(gruppen) {
				return gruppen[nummer]
			}

			return ""
		}))

		gelesen = index + laenge
	})
	if err != nil {
		return "", err
	}

	gebaut.WriteString(string(zeichen[gelesen:]))

	return gebaut.String(), nil
}
