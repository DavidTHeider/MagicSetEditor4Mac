package domain

// **Einen Schritt gehen** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// # Warum das neben der Matrix steht
//
// **Weil die Matrix eine Aussage ist und dies eine Handlung** (§2). Wer beide
// in eine Datei legte, hätte eine Stelle mit zwei Gründen, sich zu ändern:
// einem neuen Übergang und einer neuen Regel darüber, wie ein Übergang
// geschieht.

import (
	"fmt"
	"slices"
)

// Schreite geht einen Schritt — oder sagt, warum nicht.
//
// # Drei Antworten und nicht zwei
//
//  1. **Es geht** — das Angebot kommt mit dem neuen Zustand zurück.
//  2. **Es geht nicht von hier aus** — [ErrConflict], denn der Vorgang ist
//     woanders, und der Absender rechnet gegen einen Stand von vorhin.
//  3. **Du darfst das nicht** — [ErrValidation], denn das ist keine Frage des
//     Standes, sondern der Person. Wer beides zusammenlegte, ließe einen
//     Fremden so lange wiederholen, bis der Zustand passt.
//
// # Und die Wiederholung ist keine von den dreien
//
// **Zweimal dasselbe ist einmal.** Wer schon zugestimmt hat und noch einmal
// zustimmt, bekommt denselben Stand zurück — ohne Fehler und ohne zweite
// Zustimmung (§7, Regel 4).
func (a Handelsangebot) Schreite(
	schritt Handelsschritt, wer string,
) (Handelsangebot, error) {
	if !slices.Contains(a.Parteien, wer) {
		return a, fmt.Errorf(
			"%w: %q gehört nicht zu den Parteien %v dieses Vorgangs",
			ErrValidation, wer, a.Parteien)
	}

	uebergang, da := a.uebergangFuer(schritt)
	if !da {
		return a, fmt.Errorf(
			"%w: aus %q führt kein %q — der Vorgang steht woanders",
			ErrConflict, a.Zustand, schritt)
	}

	if !a.darf(uebergang.Wer, wer) {
		return a, fmt.Errorf(
			"%w: %q darf hier nicht %q — das darf nur %q",
			ErrValidation, wer, schritt, uebergang.Wer)
	}

	if !uebergang.Alle {
		return a.mitZustand(uebergang.Nach), nil
	}

	return a.mitStimme(uebergang, wer), nil
}

// mitZustand wechselt den Zustand — und **räumt den Beschluss weg**.
//
// # Der Fund, für den es dieses Verfahren gibt
//
// **Ein Beschluss ist mit seinem Wechsel verbraucht.** Die Matrix kennt zwei
// Schritte, die alle Parteien brauchen — `zustimmen` und `bestaetigen` —, und
// die Stimmen standen in **einem** Feld. Wer zugestimmt hatte, stand darin;
// beim Bestätigen fand [Handelsangebot.mitStimme] ihn wieder, hielt es für
// eine Wiederholung und gab denselben Stand zurück. `beidseitigVerbucht` war
// damit **nicht erreichbar**.
//
// **Verloren geht dabei nichts**: wer zugestimmt hat, steht in der Revision,
// in der er es tat — die Kette bleibt (`T5`).
func (a Handelsangebot) mitZustand(nach Handelszustand) Handelsangebot {
	a.Zustand = nach
	a.Zustimmungen = nil

	return a
}

// uebergangFuer sucht den Übergang aus dem heutigen Zustand.
func (a Handelsangebot) uebergangFuer(
	schritt Handelsschritt,
) (Handelsuebergang, bool) {
	for _, uebergang := range Handelsmatrix() {
		if uebergang.Schritt == schritt && uebergang.Von == a.Zustand {
			return uebergang, true
		}
	}

	return Handelsuebergang{}, false
}

// darf sagt, ob dieser Mensch in dieser Rolle steht.
//
// **Ohne Zweig für eine vierte Rolle**, und das ist keine Nachlässigkeit: die
// Rolle kommt aus [Handelsmatrix], einer Tabelle in dieser Datei nebenan. Ein
// Zweig, den kein Weg betritt, ist eine Zusage, die niemand hält (§10) — und
// er wäre hier zusätzlich unprüfbar, weil es keine vierte Rolle zu bauen
// gibt. Wer eine hinzufügt, ändert diese Stelle mit.
func (a Handelsangebot) darf(rolle Handelnder, wer string) bool {
	switch rolle {
	case HandelnderInitiator:
		return wer == a.Initiator

	case HandelnderGegenseite:
		return wer != a.Initiator
	}

	return true
}

// mitStimme trägt eine Stimme ein und wechselt, wenn alle da sind.
//
// **Die Scheibe wird kopiert und nicht geteilt.** [Handelsangebot] ist ein
// Wertetyp; wer die Scheibe weiterreichte, ließe zwei Angebote dieselben
// Zustimmungen halten — und das zweite bekäme die des ersten dazu.
func (a Handelsangebot) mitStimme(
	uebergang Handelsuebergang, wer string,
) Handelsangebot {
	if slices.Contains(a.Zustimmungen, wer) {
		return a
	}

	stimmen := make([]string, 0, len(a.Zustimmungen)+1)
	stimmen = append(stimmen, a.Zustimmungen...)
	stimmen = append(stimmen, wer)

	a.Zustimmungen = stimmen

	if len(stimmen) == len(a.Parteien) {
		return a.mitZustand(uebergang.Nach)
	}

	return a
}
