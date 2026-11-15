package domain

// **Was ein Angebot prüfen muss, bevor es gilt** — Werkstück `T1` von Phase
// 5.6 (#1466).
//
// # Warum das neben dem Verbund steht
//
// **Weil es eine andere Frage beantwortet** (§2): dort steht, **was** ein
// Angebot ist; hier steht, **wann** es eines ist. Der Schnitt folgt der
// Frage; die Zeilengrenze macht ihn nur sichtbar.

import (
	"fmt"
	"slices"
	"strings"
)

// Pruefe sagt, ob dieses Angebot überhaupt eines ist.
//
// **Ohne Bestandswirkung** — jede Ablehnung hier geschieht, bevor irgendetwas
// geschrieben wurde. Das ist die erste Abnahme des Issues.
func (a Handelsangebot) Pruefe() error {
	if fehler := a.pruefeKopf(); fehler != nil {
		return fehler
	}

	if fehler := a.pruefeKette(); fehler != nil {
		return fehler
	}

	return a.pruefePositionen()
}

// pruefeKopf prüft, was unabhängig von der Kette gilt.
func (a Handelsangebot) pruefeKopf() error {
	switch {
	case a.Vertragsfassung != Handelsvertragsfassung:
		return fmt.Errorf(
			"%w: das Angebot trägt die Vertragsfassung %d — dieses Programm "+
				"schreibt und liest %d", ErrValidation,
			a.Vertragsfassung, Handelsvertragsfassung)

	case strings.TrimSpace(a.Vorgang) == "":
		return fmt.Errorf(
			"%w: ein Angebot ohne Vorgang gehört zu keinem Handel",
			ErrValidation)

	case !slices.Contains(Handelsarten(), a.Art):
		return fmt.Errorf(
			"%w: %q ist keine Handelsart — es gibt %v",
			ErrValidation, a.Art, Handelsarten())

	case !slices.Contains(Handelszustaende(), a.Zustand):
		return fmt.Errorf(
			"%w: %q ist kein Handelszustand",
			ErrValidation, a.Zustand)

	case strings.TrimSpace(a.Initiator) == "":
		return fmt.Errorf(
			"%w: ein Angebot ohne Initiator hat niemanden, der es sequenziert",
			ErrValidation)

	case len(a.Parteien) < 2:
		return fmt.Errorf(
			"%w: ein Handel braucht zwei Parteien, genannt sind %d",
			ErrValidation, len(a.Parteien))

	case !slices.Contains(a.Parteien, a.Initiator):
		return fmt.Errorf(
			"%w: der Initiator %q steht nicht unter den Parteien %v",
			ErrValidation, a.Initiator, a.Parteien)
	}

	return a.pruefeParteienVerschieden()
}

// pruefeParteienVerschieden weist denselben Namen zweimal ab.
//
// # Warum das mehr ist als eine Formsache
//
// **Weil ein solcher Handel nie fertig wird.** `zustimmen`, `uebergeben` und
// `bestaetigen` wechseln den Zustand erst, wenn **jede** Partei gestimmt hat
// ([Handelsangebot.mitStimme]); wer zweimal dasteht, stimmt einmal, und
// `len(stimmen) == len(Parteien)` wird nie wahr. Der Vorgang bliebe stehen,
// und niemand sähe, warum.
//
// **Gefunden am 2026-09-19 im Inline-Review zu
// #1637.**
// Bis dahin kam der Fall über keine Fläche herein — `T1` hatte keine. Seit
// es eine gibt, kommt er über `HandelEroeffnen`: wer sich selbst als
// Gegenüber einträgt, bekommt eine Absage statt eines toten Vorgangs.
func (a Handelsangebot) pruefeParteienVerschieden() error {
	gesehen := make(map[string]bool, len(a.Parteien))

	for _, partei := range a.Parteien {
		if gesehen[partei] {
			return fmt.Errorf(
				"%w: %q steht zweimal unter den Parteien — ein Handel mit "+
					"sich selbst wird nie beidseitig", ErrValidation, partei)
		}

		gesehen[partei] = true
	}

	return nil
}

// pruefeKette prüft Revision und Elternabdruck gegeneinander.
//
// **Die erste Revision hat keine Eltern, jede weitere hat welche.** Ohne
// diese Prüfung hinge eine zweite Revision an nichts, und eine erste
// behauptete eine Vorgeschichte, die es nicht gibt — beides wäre eine
// Unterschrift über einen Stand, den niemand nachrechnen kann.
func (a Handelsangebot) pruefeKette() error {
	switch {
	case a.Revision < 1:
		return fmt.Errorf(
			"%w: die Revision ist %d — gezählt wird ab eins",
			ErrValidation, a.Revision)

	case a.Revision == 1 && a.Eltern != "":
		return fmt.Errorf(
			"%w: die erste Revision nennt den Elternabdruck %q — vor ihr "+
				"gibt es nichts", ErrValidation, a.Eltern)

	case a.Revision > 1 && a.Eltern == "":
		return fmt.Errorf(
			"%w: die Revision %d nennt keinen Elternabdruck und hinge damit "+
				"an nichts", ErrValidation, a.Revision)
	}

	return nil
}

// pruefePositionen prüft jede Zeile des Angebots.
func (a Handelsangebot) pruefePositionen() error {
	if len(a.Positionen) == 0 {
		return fmt.Errorf(
			"%w: ein Angebot ohne Position sagt nicht, worum es geht",
			ErrValidation)
	}

	for stelle, position := range a.Positionen {
		if fehler := position.pruefe(stelle, a.Parteien); fehler != nil {
			return fehler
		}
	}

	return nil
}

// pruefe sagt, ob diese Position eine ist.
//
// **Die Stelle steht in der Meldung**, weil ein Angebot viele Zeilen hat und
// „eine Position ohne Menge" sonst niemandem sagt, welche.
func (p Handelsposition) pruefe(stelle int, parteien []string) error {
	switch {
	case !slices.Contains(parteien, p.Von):
		return fmt.Errorf(
			"%w: die Position %d kommt von %q — das ist keine Partei dieses "+
				"Handels (%v)", ErrValidation, stelle+1, p.Von, parteien)

	case p.Menge < 1:
		return fmt.Errorf(
			"%w: die Position %d hat die Menge %d — null Exemplare sind "+
				"keine Position, sondern ihre Abwesenheit",
			ErrValidation, stelle+1, p.Menge)

	case strings.TrimSpace(p.Printing) == "" && p.DruckID == 0:
		return fmt.Errorf(
			"%w: die Position %d nennt weder einen Druck noch eine "+
				"Druckzeile", ErrValidation, stelle+1)
	}

	return nil
}
