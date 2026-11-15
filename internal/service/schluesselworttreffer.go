// Das Suchen und Ordnen der Treffer
// (#1061).
//
// **Von `schluesselwoerter.go` getrennt**, als sie über 400 Zeilen wuchs (§2).
// Der Schnitt folgt der Frage: dort, **was** ein Schlüsselwort ist und wie
// sein Ausdruck entsteht, hier, **wo** es im Text sitzt und welcher Treffer
// gewinnt.

package service

import (
	"sort"
	"strings"
)

// treffer ist ein Fund im Text.
type treffer struct {
	von, bis int
	wort     vorbereitet
	gruppen  []string
}

// Einsetzen setzt die Erinnerungstexte in einen Regeltext.
//
// # Die Reihenfolge
//
// `sort_keyword_matches`, `src/data/keyword.cpp`, Zeilen 445–458, Fassung
// f9d9356 vom 2021-07-06: nach Stelle, dann **eigene vor festen**, dann der
// **längste** Treffer, dann nach Namen. Hier entscheidet zuletzt die Stelle in
// der Eingabereihe — sie ist stabil und braucht den Namen nicht.
//
// # Was hier **nicht** steht
//
// **Die Auszeichnungsarithmetik von MSE.** `expand_keywords` rechnet zwischen
// dem ausgezeichneten und dem entlarvten Text hin und her, behandelt
// `<kw-X>`-Zustände, Trennerskripte und Platzhalter. Hier wird auf dem Text
// gearbeitet, wie er kommt, und **innerhalb von `<atom…>` nichts gesetzt** —
// das ist die Zusage, die MSE an derselben Stelle macht (Zeile 526).
//
// Was fehlt, ist benannt und nicht verdeckt: Trennerskripte, Platzhalter für
// fehlende Parameter und der `<kw-X>`-Zustand, mit dem eine Karte einen
// Erinnerungstext einzeln an- und abschaltet.
func (l Wortliste) Einsetzen(text string, zusammen func(Schluesselworttreffer) string) string {
	if len(l.woerter) == 0 {
		return text
	}

	gefunden := l.suche(text)
	if len(gefunden) == 0 {
		return text
	}

	var gebaut strings.Builder

	zuletzt := 0

	for _, einer := range gefunden {
		if einer.von < zuletzt {
			continue
		}

		gebaut.WriteString(text[zuletzt:einer.von])
		gebaut.WriteString(zusammen(alsTreffer(text, einer)))

		zuletzt = einer.bis
	}

	gebaut.WriteString(text[zuletzt:])

	return gebaut.String()
}

// suche sammelt alle Treffer und ordnet sie.
func (l Wortliste) suche(text string) []treffer {
	frei := freieBereiche(text)

	var gefunden []treffer

	for _, wort := range l.woerter {
		for _, stellen := range wort.ausdruck.FindAllStringSubmatchIndex(text, -1) {
			// **Der Nachlauf gehört nicht zum Treffer.** Er steht als letzte
			// Gruppe da, weil Go keinen Vorausblick kennt.
			nachlauf := stellen[len(stellen)-2]

			ende := stellen[1]
			if nachlauf >= 0 {
				ende = nachlauf
			}

			if !imFreien(frei, stellen[0], ende) {
				continue
			}

			gefunden = append(gefunden, treffer{
				von: stellen[0], bis: ende, wort: wort,
				gruppen: parameterwerte(text, stellen),
			})
		}
	}

	sort.SliceStable(gefunden, func(links, rechts int) bool {
		return besserZuerst(gefunden[links], gefunden[rechts])
	})

	return gefunden
}

// besserZuerst ist `sort_keyword_matches`.
func besserZuerst(links, rechts treffer) bool {
	switch {
	case links.von != rechts.von:
		return links.von < rechts.von

	case links.wort.wort.Fest != rechts.wort.wort.Fest:
		return !links.wort.wort.Fest

	case links.bis-links.von != rechts.bis-rechts.von:
		return links.bis-links.von > rechts.bis-rechts.von

	default:
		return links.wort.stelle < rechts.wort.stelle
	}
}

// parameterwerte sammelt die Parameterwerte — die Gruppen ohne den Nachlauf.
func parameterwerte(text string, stellen []int) []string {
	var werte []string

	for gruppe := 1; gruppe < len(stellen)/2-1; gruppe++ {
		von, bis := stellen[2*gruppe], stellen[2*gruppe+1]
		if von < 0 {
			werte = append(werte, "")

			continue
		}

		werte = append(werte, text[von:bis])
	}

	return werte
}

// alsTreffer baut, was der Aufrufer sieht.
func alsTreffer(text string, einer treffer) Schluesselworttreffer {
	return Schluesselworttreffer{
		Text:       text[einer.von:einer.bis],
		Wort:       einer.wort.wort.Wort,
		Modus:      einer.wort.wort.Modus,
		Erinnerung: einer.wort.wort.Erinnerung,
		Parameter:  einer.gruppen,
		Fest:       einer.wort.wort.Fest,

		// **Am Rest des Textes geprüft, nicht nur am Treffer**: der Ausdruck
		// endet auf einem Zeichen, das kein Wortlaut ist, und das steht hinter
		// dem Treffer.
		SchreibungStimmt: einer.wort.genau.MatchString(text[einer.von:]),
	}
}
