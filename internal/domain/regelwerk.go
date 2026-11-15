package domain

// **Welche Regeln in welcher Strenge gelten** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Was hier steht und was in G4
//
// **Hier der Griff, dort die Matrix.** G1 braucht ein Regelwerk, um eine
// Partie zu prüfen und einen Befehl zuzuordnen; **welche** Regeln ein Format
// wirklich trägt und wie eine Unterstützungsmatrix aussieht, ist Werkstück G4
// (#1461).
//
// **Der Schnitt ist kein Aufschub, sondern die Reihenfolge**: ohne den Griff
// ließe sich G4 nirgends einhängen, und mit der ganzen Matrix hier stünde in
// G1 Fachlogik, für die es noch keinen Aufrufer gibt.
//
// # Die Strenge ist eine Zusage und keine Erleichterung
//
// **Sandkasten lockert keine Invariante** (#1458): keine über Objekte, keine
// über Sichtbarkeit, keine über Integrität und keine über Rechte. Was er
// lockert, sind **Regelprüfungen** — wer im Sandkasten eine Karte auf den
// Tisch legt, die dort nicht hingehört, darf das; wessen Hand er dabei sieht,
// ändert sich nicht.

import (
	"fmt"
	"strings"
)

// Strenge sagt, wie streng die Regeln geprüft werden.
type Strenge string

const (
	// StrengeSandkasten prüft nichts — der Tisch gehorcht dem Menschen.
	StrengeSandkasten Strenge = "sandkasten"

	// StrengeBegleitet prüft und **warnt**, hält aber nichts auf.
	StrengeBegleitet Strenge = "begleitet"

	// StrengeStreng weist ab, was die Regeln nicht erlauben.
	//
	// **Nur für nachgewiesene Unterstützung** (#1409, Grenze): ein Format
	// ohne Nachweis steht hier nicht zur Wahl, und ein „streng" ohne Deckung
	// wäre eine Zusage, die beim ersten Sonderfall bricht.
	StrengeStreng Strenge = "streng"
)

// Strengegrade gibt alle drei.
func Strengegrade() []Strenge {
	return []Strenge{StrengeSandkasten, StrengeBegleitet, StrengeStreng}
}

// IstStrenge sagt, ob ein Wert einer der drei ist.
func IstStrenge(grad Strenge) bool {
	for _, gilt := range Strengegrade() {
		if gilt == grad {
			return true
		}
	}

	return false
}

// Regelwerk ist das Regelpaket einer Partie — versioniert.
type Regelwerk struct {
	// Kennung nennt das Regelpaket — `sandkasten`, `commander`, …
	Kennung string

	// Fassung ist seine Nummer.
	//
	// **Versioniert und nicht „die neueste"**: eine laufende Partie bleibt
	// bei der Fassung, unter der sie angefangen hat — sonst änderte ein
	// Paketwechsel den Zustand einer Partie, die niemand angefasst hat.
	Fassung int

	Strenge Strenge

	// Abgewaehlt sind die Bereiche, die die Runde **selbst** übernimmt —
	// PLAY-02.AK-02 und AK-03
	// (#1710).
	//
	// # Warum sie hier stehen und nicht daneben
	//
	// **Weil sie Teil der Vereinbarung sind.** AK-03 verlangt: „Profil und
	// aktive Regeln sind Bestandteil des vereinbarten Partiestands." Hier
	// stehend werden sie mit dem Regelwerk **eingefroren**, reisen mit jeder
	// Sicht und jedem Abzug mit und stehen im Ereignisstrom — wie die
	// Fassung darüber.
	//
	// **Leer heißt: das Programm trägt, was es tragen kann.**
	//
	// # Und warum es keinen Weg gibt, sie später zu ändern
	//
	// **Weil eine laufende Partie ihre Regeln nicht wechselt.** AK-03 nennt
	// für eine Änderung „den vorgesehenen gemeinsamen Zustimmungsweg"; der
	// sicherste ist, dass es keinen Weg gibt, der daran vorbeiführte — wer
	// anders spielen will, fängt eine neue Partie an. Ein Wächter hält es:
	// [internal/werkzeuge/regelnfest_test.go].
	Abgewaehlt []Regelbereich
}

// Pruefe sagt, ob dieses Regelwerk benutzbar ist.
func (r Regelwerk) Pruefe() error {
	if strings.TrimSpace(r.Kennung) == "" {
		return fmt.Errorf(
			"%w: eine Partie ohne Regelwerk hat keine Grundlage, auf der sich "+
				"ein Befehl zuordnen ließe", ErrValidation)
	}

	if fehler := (Regelwahl{Abgewaehlt: r.Abgewaehlt}).Pruefe(); fehler != nil {
		return fehler
	}

	if r.Fassung < 1 {
		return fmt.Errorf(
			"%w: das Regelwerk %q hat die Fassung %d — eine laufende Partie "+
				"bleibt bei ihrer, und dafür braucht sie eine",
			ErrValidation, r.Kennung, r.Fassung)
	}

	if !IstStrenge(r.Strenge) {
		return fmt.Errorf(
			"%w: %q ist keine Strenge — es gibt Sandkasten, begleitet und "+
				"streng", ErrValidation, r.Strenge)
	}

	return nil
}
