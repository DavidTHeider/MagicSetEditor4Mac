package domain

// **Die Reihenfolge der Revisionen** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// # Warum der Initiator sequenziert und nicht die Uhr
//
// **Weil zwei Rechner keine gemeinsame Uhr haben.** Wer nach Zeitstempeln
// ordnete, bekäme bei zwei Vorschlägen im selben Augenblick zwei Revisionen
// mit der Nummer N — und beim ersten Unterschied wüsste niemand, welche gilt.
// **Eine Stelle entscheidet**, und das ist die, die den Vorgang eröffnet hat.
//
// # Und warum ein Vorschlag seine Eltern nennt
//
// **Weil „ich habe das Neueste gelesen" keine Aussage ist.** Der Abdruck der
// Revision, auf der ein Vorschlag aufsetzt, ist eine: stimmt er nicht mehr,
// hat inzwischen jemand anderes geschrieben — und der Vorschlag bekommt einen
// **Konflikt** statt einer Nummer.

import (
	"fmt"
	"slices"
)

// Nimmt sagt, ob dieser Vorschlag auf den heutigen Stand passt.
//
// # Zwei Fehlerarten, und der Unterschied ist für den Aufrufer
//
//   - **[ErrConflict]** heißt: noch einmal lesen und noch einmal vorschlagen.
//     Der Stand ist weiter, der Vorschlag war nicht falsch.
//   - **[ErrValidation]** heißt: daran ändert kein Wiederholen etwas — ein
//     fremder Vorgang, eine unbekannte Vertragsfassung, ein veränderter Kopf.
//
// **Wer beides zusammenlegte**, ließe einen Klienten in einer Schleife
// wiederholen, in der er nie ankommt.
func (a Handelsangebot) Nimmt(vorschlag Handelsangebot) error {
	if fehler := vorschlag.Pruefe(); fehler != nil {
		return fehler
	}

	if fehler := a.pruefeKopfgleich(vorschlag); fehler != nil {
		return fehler
	}

	return a.pruefeAnschluss(vorschlag)
}

// NimmtVorschlagVon ist [Handelsangebot.Nimmt] mit dem, der vorschlägt — und
// die Rolle, die die Matrix für `gegenangebot` nennt.
//
// **Aus `angeboten` schlägt nur die Gegenseite vor** (TRADE-02, #1716). Bis
// #1975 hielt das allein das Fenster, weil der Vorschlag nicht wusste, wer
// ihn machte; Kommandozeile, REST, MCP und die Leitung kamen daran vorbei,
// und beide Seiten liefen auseinander. **Ein Verstoß ist eine
// [ErrValidation]**, wie bei [Handelsangebot.Schreite]: kein Wiederholen
// ändert, wer man ist.
func (a Handelsangebot) NimmtVorschlagVon(vorschlag Handelsangebot, wer string) error {
	if fehler := a.Nimmt(vorschlag); fehler != nil {
		return fehler
	}

	// **Den Übergang gibt es** — [Handelsangebot.Nimmt] hat ihn eben
	// verlangt. **Und er nennt nur die Gegenseite oder jede**; das hält
	// `TestEinGegenangebotNenntNurGegenseiteOderJede`, und deshalb sagt der
	// Satz „nur die Gegenseite".
	uebergang, _ := a.uebergangFuer(SchrittGegenangebot)
	if !a.darf(uebergang.Wer, wer) {
		return NewFieldError("vorschlag", SchluesselVorschlagNurGegenseite,
			"%q darf aus %q nicht vorschlagen — das darf nur die Gegenseite", wer, a.Zustand)
	}

	return nil
}

// pruefeKopfgleich hält fest, was über alle Revisionen gleich bleibt.
//
// **Der Kopf ist keine Verhandlungsmasse.** Wer die Parteien änderte, machte
// aus einem Handel einen anderen und behielte die Zustimmungen der ersten
// Revisionen darüber; wer die Art änderte, machte aus einer Deckleihe einen
// Tausch, ohne dass jemand zugestimmt hätte.
func (a Handelsangebot) pruefeKopfgleich(vorschlag Handelsangebot) error {
	switch {
	case vorschlag.Vorgang != a.Vorgang:
		return fmt.Errorf(
			"%w: der Vorschlag gehört zum Vorgang %q, hier läuft %q",
			ErrValidation, vorschlag.Vorgang, a.Vorgang)

	case vorschlag.Art != a.Art:
		return fmt.Errorf(
			"%w: der Vorschlag ist ein %q, der Vorgang ist ein %q",
			ErrValidation, vorschlag.Art, a.Art)

	case vorschlag.Initiator != a.Initiator:
		return fmt.Errorf(
			"%w: der Vorschlag nennt %q als Initiator, der Vorgang %q",
			ErrValidation, vorschlag.Initiator, a.Initiator)

	case !slices.Equal(vorschlag.Parteien, a.Parteien):
		return fmt.Errorf(
			"%w: der Vorschlag nennt die Parteien %v, der Vorgang %v",
			ErrValidation, vorschlag.Parteien, a.Parteien)
	}

	return nil
}

// pruefeAnschluss hält den Vorschlag gegen den heutigen Stand.
func (a Handelsangebot) pruefeAnschluss(vorschlag Handelsangebot) error {
	switch {
	case HandelEnde(a.Zustand):
		return fmt.Errorf(
			"%w: der Vorgang steht auf %q — daraus folgt nichts mehr",
			ErrConflict, a.Zustand)

	// **Ein Vorschlag ist ein Gegenangebot**, und das geht nur, wo die Matrix
	// es kennt — aus `angeboten` und `verhandlung`. Nach der Einigung gibt
	// es nichts mehr vorzuschlagen (TRADE-02, #1716); bis QA-06b (#1967)
	// sagte das nur das Fenster, und ein Vorschlag aus `uebergabeOffen`
	// führte an der Klärung vorbei zurück in die Verhandlung.
	case !slices.ContainsFunc(Handelsmatrix(), func(u Handelsuebergang) bool {
		return u.Schritt == SchrittGegenangebot && u.Von == a.Zustand
	}):
		return fmt.Errorf(
			"%w: aus %q führt kein Gegenangebot — der Vorgang steht woanders",
			ErrConflict, a.Zustand)

	case vorschlag.Revision != a.Revision+1:
		return fmt.Errorf(
			"%w: der Vorschlag ist Revision %d, die nächste wäre %d",
			ErrConflict, vorschlag.Revision, a.Revision+1)

	case vorschlag.Eltern != a.Abdruck():
		return fmt.Errorf(
			"%w: der Vorschlag setzt auf %q auf, hier gilt %q — inzwischen "+
				"hat jemand anderes geschrieben",
			ErrConflict, vorschlag.Eltern, a.Abdruck())
	}

	return nil
}
