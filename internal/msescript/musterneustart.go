package msescript

// **Die Suche nach jedem Treffer neu, und die Stelle ist dann der Anfang** —
// Befunde 8 und 9 des Zweitblicks über Phase 1 von 0.9.0 (E2315, #2315).
//
// `filter_text`, `break_text`, `split_text` und `replace` mit Kontext,
// Ersatzfunktion oder `recursive` laufen bei MSE2 in einer Schleife über
// `regex_search(start, end)` — **ohne** `match_prev_avail`
// (`src/script/functions/regex.cpp`, Zeilen 72–110 und 154–232;
// `src/util/regex.hpp`, Zeile 80; Fassung f9d9356 vom 2021-07-06). Boost hält
// `start` dann für den Anfang des Textes: `^` und `\b` greifen dort wieder,
// ein Rückblick sieht nichts davor. Hier wird deshalb **ab der Stelle in einem
// eigenen Stück Text** gesucht — RE2 und `regexp2` lesen ein Stück genauso.
//
// Nur das einfache `replace` sieht den ganzen Text ([Muster.Ersetze]).

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

// ErrLeererTreffer meldet einen angenommenen leeren Treffer **an der Stelle,
// an der die Suche stand** — sie käme nicht weiter.
//
// **`filter_text` wirft ihn bei MSE2** („Regular expression matches empty
// string", `regex.cpp`, Zeilen 164–167), und zwar bei `pos.second == start`:
// **ein leerer Treffer weiter hinten ist keiner**, er rückt die Suche vor
// (Nachlauf über Phase 1 von 0.9.0, Befund 6; vom Orakel bestätigt).
// `break_text`, `split_text` und die Ersatzschleife laufen dort endlos, wenn
// `start` stehen bleibt — **hier ist es überall derselbe Fehler** (adapted):
// eine Vorlage, die das braucht, hielte MSE2 an.
var ErrLeererTreffer = errors.New("das Muster trifft die leere Zeichenkette")

// jederTrefferMitNeustart läuft über die Treffer, die ihr Kontext zulässt —
// je Treffer neu ab seinem Ende. `tue` bekommt die Stelle in Zeichen und die
// Gruppen; die nullte ist der ganze Treffer.
func (m *Muster) jederTrefferMitNeustart(text string, tue func(index, laenge int, gruppen []string)) error {
	zeichen := []rune(text)
	start := 0

	for {
		index, laenge, gruppen, gefunden, err := m.naechsterTreffer(zeichen, start)
		if err != nil || !gefunden {
			return err
		}

		if index+laenge == start {
			return fmt.Errorf("%w: %s", ErrLeererTreffer, m.quelle)
		}

		tue(index, laenge, gruppen)

		start = index + laenge
	}
}

// naechsterTreffer ist `ScriptRegex::matches`: der erste Treffer ab `start`,
// den der Kontext zulässt (`regex.cpp`, Zeilen 29–51).
//
// **Der Kontext sieht den ganzen Text** — „davor + `<match>` + danach" reicht
// vom Anfang bis zum Ende, nicht nur ab `start`. **Hinter einem abgelehnten
// Treffer wird weitergesucht**, und ein leerer rückt um ein Zeichen vor.
func (m *Muster) naechsterTreffer(zeichen []rune, start int) (index, laenge int, gruppen []string, gefunden bool, err error) {
	anfang := start

	for {
		index, laenge, gruppen, gefunden, err = m.trefferAb(zeichen, anfang)
		if err != nil || !gefunden || m.kontext == nil {
			return index, laenge, gruppen, gefunden, err
		}

		passt, kontextfehler := m.kontext.Passt(string(zeichen[:index]) + "<match>" + string(zeichen[index+laenge:]))
		if kontextfehler != nil || passt {
			return index, laenge, gruppen, passt, kontextfehler
		}

		switch {
		case index+laenge != anfang:
			anfang = index + laenge
		case anfang == len(zeichen):
			return 0, 0, nil, false, nil
		default:
			anfang++
		}
	}
}

// trefferAb sucht den ersten Treffer in `zeichen[start:]` — **als finge der
// Text dort an**.
func (m *Muster) trefferAb(zeichen []rune, start int) (index, laenge int, gruppen []string, gefunden bool, err error) {
	stueck := string(zeichen[start:])

	if m.schnell != nil {
		stellen := m.schnell.FindStringSubmatchIndex(stueck)
		if stellen == nil {
			return 0, 0, nil, false, nil
		}

		gruppen = make([]string, 0, len(stellen)/2)
		for paar := 0; paar+1 < len(stellen); paar += 2 {
			if stellen[paar] < 0 {
				gruppen = append(gruppen, "")

				continue
			}

			gruppen = append(gruppen, stueck[stellen[paar]:stellen[paar+1]])
		}

		return start + utf8.RuneCountInString(stueck[:stellen[0]]),
			utf8.RuneCountInString(stueck[stellen[0]:stellen[1]]), gruppen, true, nil
	}

	treffer, err := m.breit.FindStringMatch(stueck)
	if err != nil {
		return 0, 0, nil, false, fmt.Errorf("%w: %s: %w", ErrLauf, m.quelle, err)
	}

	if treffer == nil {
		return 0, 0, nil, false, nil
	}

	gruppen = make([]string, 0, treffer.GroupCount())
	for _, gruppe := range treffer.Groups() {
		gruppen = append(gruppen, gruppe.String())
	}

	return start + treffer.Index, treffer.Length, gruppen, true, nil
}
