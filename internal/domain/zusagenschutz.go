package domain

// **Was eine offene Zusage einem Schreibweg verbietet** — Werkstück `T2` von
// Phase 5.6
// (#1467).
//
// # Warum das neben der Reservierung steht und nicht darin
//
// **Weil es die andere Richtung ist.** [ReservierungPasst] fragt, ob eine
// Zusage in den Besitz hineinpasst; hier wird gefragt, ob der Besitz nach
// einem Schreiben noch zu den Zusagen passt. Dieselben Zahlen, zwei Zeitpunkte
// — und der zweite ist der, an dem jemand seine Sammlung pflegt, ohne an einen
// laufenden Handel zu denken.
//
// # Und warum nicht jede Änderung verboten ist
//
// **Weil eine Änderung eine Zusage entwerten darf, solange nichts übergeben
// ist.** [Reservierung.Fassung] sagt, gegen welchen Stand gerechnet wurde; wer
// die Zeile ändert, hebt die Fassung — und der Handel muss neu prüfen lassen.
// Das ist die Antwort des Issues auf „relevante Änderung invalidiert die
// Revision". Was **nicht** geschehen darf, ist, dass zugesagte Stücke
// verschwinden.

import (
	"fmt"
	"time"
)

// AenderungPasst sagt, ob ein geänderter Besitz die Zusagen noch trägt.
//
// **Gemessen wird gegen die Menge, nicht gegen die Zeile.** Zustand, Sprache
// und Ausführung darf jemand ändern, solange die Stückzahl reicht: was dabei
// zugesagt war, gilt für eine Fassung, die es danach nicht mehr gibt, und der
// Handel liest das an seiner eigenen Prüfung ab.
func AenderungPasst(neu Exemplar, zusagen []Reservierung, jetzt time.Time) error {
	zugesagt := Zugesagt(neu.ID, zusagen, jetzt)
	if zugesagt <= neu.Menge {
		return nil
	}

	return fmt.Errorf(
		"%w: für dieses Exemplar sind %d Stück zugesagt, stehen bleiben "+
			"sollen %d — gib die Zusage frei oder ändere sie",
		ErrConflict, zugesagt, neu.Menge)
}

// LoeschenPasst sagt, ob eine Zeile fortdarf.
//
// **Eine einzige offene Zusage genügt.** `reservierungen.exemplar_id` hängt
// mit `ON DELETE CASCADE` an der Zeile (Wanderung 0069): wer sie löscht,
// nimmt die Zusage mit — und der Partner wartet auf eine Karte, von der
// dieses Programm nichts mehr weiß.
func LoeschenPasst(
	exemplarID int64, zusagen []Reservierung, jetzt time.Time,
) error {
	if zugesagt := Zugesagt(exemplarID, zusagen, jetzt); zugesagt > 0 {
		return fmt.Errorf(
			"%w: für dieses Exemplar sind %d Stück zugesagt — es lässt sich "+
				"nicht löschen, solange die Zusage steht",
			ErrConflict, zugesagt)
	}

	return nil
}

// Zugesagt zählt, wie viele Stücke eines Exemplars jetzt festgehalten sind.
//
// **Die Kennung wird mitgeprüft.** Eine Liste von Zusagen kann die mehrerer
// Zeilen sein; wer sie ungeprüft summierte, verböte eine Änderung wegen einer
// Zusage über eine andere Karte.
func Zugesagt(exemplarID int64, zusagen []Reservierung, jetzt time.Time) int {
	summe := 0

	for _, zusage := range zusagen {
		if zusage.ExemplarID == exemplarID && zusage.Aktiv(jetzt) {
			summe += zusage.Menge
		}
	}

	return summe
}
