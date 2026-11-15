package domain

// **Die Reservierung** — Werkstück `T2` von Phase 5.6
// (#1467).
//
// # Warum es sie gibt
//
// **Weil eine Zusage etwas festhalten muss, das es nur einmal gibt.** Ein
// Stück Papier lässt sich zweimal versprechen und einmal übergeben; wer das
// nicht verhindert, hat zwei Handelspartner und eine Karte.
//
// # Zwei Fälle, die gleich aussehen und es nicht sind
//
// **Eine abgelaufene Frist gibt frei, eine bestätigte Übergabe nicht.** Wer
// beides gleich behandelte, gäbe Karten frei, die längst über den Tisch
// gegangen sind — und der Bestand sagte, man habe sie noch. Was danach noch
// zu klären ist, klärt ein Mensch; keine Uhr, kein Neustart und kein
// Verbindungsverlust tun es.

import (
	"fmt"
	"time"
)

// Reservierung hält Stücke eines Exemplars für einen Handelsvorgang fest.
type Reservierung struct {
	// Vorgang ist der Handel, für den festgehalten wird.
	Vorgang string

	// Revision ist die Angebotsrevision, auf die sich die Zusage bezieht.
	//
	// **Sie steht dabei, weil eine Zusage einer Revision gilt** (`T1`): wer
	// das Angebot ändert, ändert auch, was festzuhalten ist.
	Revision int

	// ExemplarID ist die Besitzzeile, aus der festgehalten wird.
	ExemplarID int64

	// Menge ist die Stückzahl — mindestens eine.
	Menge int

	// Fassung ist die Fassung des Exemplars, gegen die reserviert wurde.
	//
	// **Sie ist die Bedingung und nicht die Buchführung** (§7, Regel 2):
	// ändert jemand die Menge, während eine Zusage offen ist, rechnet die
	// Zusage gegen einen Stand, den es nicht mehr gibt.
	Fassung int64

	// Frist ist, bis wann sie gilt — der Nullwert heißt: ohne Frist.
	Frist time.Time

	// Uebergeben sagt, dass eine Übergabe bestätigt ist.
	//
	// **Danach gibt keine Uhr mehr frei** — der Grund steht im Kopf.
	Uebergeben bool
}

// Aktiv sagt, ob diese Reservierung jetzt noch etwas festhält.
func (r Reservierung) Aktiv(jetzt time.Time) bool {
	return r.Uebergeben || r.Frist.IsZero() || jetzt.Before(r.Frist)
}

// Verfuegbar sagt, wie viele Stücke eines Exemplars frei sind.
//
// **Ein Wunsch hat nichts Freies** — er ist kein Besitz ([Exemplar.Besitzt],
// [0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).
// Wer eine Wunschzeile reservierte, sagte etwas zu, das er nicht hat.
//
// **Nie weniger als null**: ein von Hand geänderter Bestand kann mehr
// festgehalten haben, als heute dasteht (§8). Eine negative freie Menge wäre
// keine Auskunft, sondern eine Einladung, irgendwo damit zu rechnen.
func Verfuegbar(
	exemplar Exemplar, reservierungen []Reservierung, jetzt time.Time,
) int {
	if !exemplar.Besitzt() {
		return 0
	}

	frei := exemplar.Menge

	for _, reservierung := range reservierungen {
		if reservierung.ExemplarID == exemplar.ID && reservierung.Aktiv(jetzt) {
			frei -= reservierung.Menge
		}
	}

	return max(0, frei)
}

// ReservierungPasst sagt, ob diese Reservierung noch hineinpasst.
//
// # Zwei Fehlerarten, und der Unterschied ist für den Aufrufer
//
//   - **[ErrValidation]** heißt: daran ändert kein Wiederholen etwas — eine
//     Wunschzeile, ein fremdes Exemplar, keine Menge.
//   - **[ErrConflict]** heißt: der Stand ist weiter. Die Fassung stimmt nicht
//     mehr, oder es ist nicht mehr genug frei. **Noch einmal lesen und noch
//     einmal fragen** kann helfen.
func ReservierungPasst(
	exemplar Exemplar, bestehende []Reservierung, neu Reservierung,
	jetzt time.Time,
) error {
	if fehler := neu.pruefe(exemplar); fehler != nil {
		return fehler
	}

	if neu.Fassung != exemplar.Fassung {
		return fmt.Errorf(
			"%w: die Zusage rechnet gegen die Fassung %d, das Exemplar steht "+
				"bei %d — jemand hat inzwischen geschrieben",
			ErrConflict, neu.Fassung, exemplar.Fassung)
	}

	if frei := Verfuegbar(exemplar, bestehende, jetzt); neu.Menge > frei {
		return fmt.Errorf(
			"%w: festgehalten werden sollen %d, frei sind %d von %d",
			ErrConflict, neu.Menge, frei, exemplar.Menge)
	}

	return nil
}

// pruefe sagt, ob diese Reservierung überhaupt eine ist.
func (r Reservierung) pruefe(exemplar Exemplar) error {
	switch {
	case !exemplar.Besitzt():
		return fmt.Errorf(
			"%w: eine Wunschzeile lässt sich nicht festhalten — sie ist kein "+
				"Besitz", ErrValidation)

	case r.Vorgang == "":
		return fmt.Errorf(
			"%w: eine Reservierung ohne Vorgang hält für niemanden fest",
			ErrValidation)

	case r.Revision < 1:
		return fmt.Errorf(
			"%w: die Revision ist %d — eine Zusage gilt einer Revision",
			ErrValidation, r.Revision)

	case r.ExemplarID != exemplar.ID:
		return fmt.Errorf(
			"%w: die Reservierung meint das Exemplar %d, geprüft wird gegen %d",
			ErrValidation, r.ExemplarID, exemplar.ID)

	case r.Menge < 1:
		return fmt.Errorf(
			"%w: festgehalten werden sollen %d Stücke — null ist keine "+
				"Reservierung, sondern ihre Abwesenheit",
			ErrValidation, r.Menge)
	}

	return nil
}
