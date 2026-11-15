package domain

// **Der Stand einer Problemprüfung über ein ganzes Set** — PERF02
// (#2352),
// Befunde A-14, B-13 und Z-03 des Audits 0.9.0.
//
// # Warum es das gibt
//
// **Die Probleme einer Karte entstehen beim Zeichnen** — Plan, Rasterung,
// Nennungen —, und bis PERF02 zeichnete nur die Vorschau, eine Karte zur
// Zeit. Die Übersicht konnte deshalb Karten zählen, aber nicht sagen, wie
// viele davon ein Problem haben. Eine Prüfung über alle Karten dauert; ihr
// Ergebnis gilt nur, solange das Set dasselbe ist, und nur, wenn sie zu Ende
// lief.
//
// # Was ein Stand weiß
//
// **Den Abdruck seiner Eingaben** — Inhalt des Sets, Stil, Pakete, Schriften;
// wer liest, vergleicht ihn mit dem heutigen. **Wie weit er kam** —
// geprüfte gegen gesamte Karten. Und je Karte mit Problem die Befunde.
//
// **Die Zahl daraus ist nur sicher, wenn der Stand aktuell ist.** Eine
// abgebrochene Prüfung, die bis dahin nichts fand, sagt nicht „0 Probleme",
// sondern „nicht zu Ende geprüft".

import "time"

// Problemurteil sagt, wie weit ein Stand gilt.
type Problemurteil string

// Die Urteile über einen Stand — technische Schlüssel; den Satz setzt das
// Fenster in seiner Sprache.
const (
	ProblemeUngeprueft     Problemurteil = "ungeprueft"
	ProblemeAktuell        Problemurteil = "aktuell"
	ProblemeVeraltet       Problemurteil = "veraltet"
	ProblemeUnvollstaendig Problemurteil = "unvollstaendig"
)

// Kartenproblem sind die Befunde einer Karte.
type Kartenproblem struct {
	KarteID int64
	Befunde []string
}

// Problemstand ist das Ergebnis der letzten Prüfung eines Sets.
type Problemstand struct {
	SetID int64
	// Abdruck fasst die Eingaben der Prüfung zusammen; ein anderer heute
	// heißt, der Stand ist veraltet.
	Abdruck  string
	Begonnen time.Time
	Beendet  time.Time
	Geprueft int
	Gesamt   int
	// Vollstaendig sagt, ob jede Karte geprüft wurde — falsch nach einem
	// Abbruch oder einem Fehler.
	Vollstaendig bool
	// Karten nennt nur die Karten mit Problemen.
	Karten []Kartenproblem
}

// Urteil sagt, wie weit der Stand gegenüber dem heutigen Abdruck gilt.
func (s Problemstand) Urteil(abdruck string) Problemurteil {
	switch {
	case s.SetID == 0:
		return ProblemeUngeprueft
	case !s.Vollstaendig:
		return ProblemeUnvollstaendig
	case s.Abdruck != abdruck:
		return ProblemeVeraltet
	default:
		return ProblemeAktuell
	}
}

// Zahl gibt die Karten mit Problemen und ob die Zahl sicher ist — **nur ein
// aktueller Stand gibt eine sichere**; sonst ist sie veraltet oder eine
// Untergrenze.
func (s Problemstand) Zahl(abdruck string) (int, bool) {
	return len(s.Karten), s.Urteil(abdruck) == ProblemeAktuell
}
