package storage

// **Das Kostenjournal externer Aufträge** — Werkstück O12 von Phase 5.2
// (#1439).
//
// # Die Reservierung ist **eine** Transaktion, und das ist der ganze Punkt
//
// Die Abnahme verlangt: „Zwei gleichzeitig gestartete unterschiedliche
// Prozesse können nicht zusammen mehr als das lokale genehmigte Budget
// reservieren." Ein Zähler im Arbeitsspeicher kann das nicht — er kennt den
// anderen Prozess nicht. Eine Transaktion kann es: sie zählt zusammen, was
// gebunden ist, vergleicht mit der Grenze und schreibt den Posten **im selben
// Zug**. SQLite serialisiert die Schreiber; wer als Zweiter kommt, sieht den
// Ersten.
//
// **`_txlock=immediate` steht am Treiber** (`datenbank.go`): eine Transaktion,
// die erst liest und dann schreibt, nähme sonst die Sperre zu spät — und zwei
// Leser kämen beide durch.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Kostenjournal ist die Ablage der Budgets und Posten.
type Kostenjournal struct{ db *sql.DB }

// NeuesKostenjournal baut sie über einer Datenbank.
func NeuesKostenjournal(db *sql.DB) *Kostenjournal { return &Kostenjournal{db: db} }

// BudgetSetzen legt die Grenze eines Kontos fest — oder nimmt sie weg.
//
// **Null heißt keine Grenze und damit kein kostenpflichtiger Auftrag.** Das
// ist die sichere Vorgabe: wer nichts freigegeben hat, gibt nichts aus.
func (j *Kostenjournal) BudgetSetzen(
	ctx context.Context, profilID int64, konto string,
	zeitraum domain.Zeitraum, betrag domain.Betrag, jetzt time.Time,
) error {
	if fehler := betrag.Pruefe(); fehler != nil {
		return fehler
	}

	return imGeschaeft(ctx, j.db, "das Budget setzen", func(geschaeft *sql.Tx) error {
		_, fehler := geschaeft.ExecContext(ctx, `
			INSERT INTO kostenbudgets (profil_id, konto, zeitraum, betrag, waehrung, geaendert_am)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT (profil_id, konto, zeitraum) DO UPDATE SET
				betrag = excluded.betrag,
				waehrung = excluded.waehrung,
				geaendert_am = excluded.geaendert_am`,
			profilID, konto, string(zeitraum), betrag.Wert, betrag.Waehrung, Zeitstempel(jetzt))

		return uebersetze(fehler, "das Budget für %q setzen", konto)
	})
}

// Budget liest die Grenze eines Kontos — leer, wenn keine gesetzt ist.
func (j *Kostenjournal) Budget(
	ctx context.Context, profilID int64, konto string, zeitraum domain.Zeitraum,
) (domain.Betrag, error) {
	var betrag domain.Betrag

	fehler := j.db.QueryRowContext(ctx, `
		SELECT betrag, waehrung FROM kostenbudgets
		 WHERE profil_id = ? AND konto = ? AND zeitraum = ?`,
		profilID, konto, string(zeitraum)).Scan(&betrag.Wert, &betrag.Waehrung)

	if errors.Is(fehler, sql.ErrNoRows) {
		return domain.Betrag{}, nil
	}

	if fehler != nil {
		return domain.Betrag{}, uebersetze(fehler, "das Budget für %q lesen", konto)
	}

	return betrag, nil
}

// Reservieren bucht einen Auftrag **vor** dem Ruf nach draußen.
//
// # Drei Ausgänge, und der erste ist der wichtigste
//
//  1. **Die Kennung gibt es schon** — dann kommt der vorhandene Posten
//     zurück, und zwar unverändert. Das ist der Schutz gegen den zweiten
//     kostenpflichtigen Auftrag: ein Abbruch nach dem Ruf, ein Neustart, ein
//     Wiederholungsversuch — alle drei zeigen dieselbe Identität vor.
//  2. **Die Grenze trägt es nicht** — [domain.ErrConflict] mit der Zahl, die
//     fehlt.
//  3. **Es geht** — der Posten steht als `reserviert` im Journal.
func (j *Kostenjournal) Reservieren(
	ctx context.Context, posten domain.Kostenposten, zeitraum domain.Zeitraum, jetzt time.Time,
) (domain.Kostenposten, error) {
	if fehler := posten.Pruefe(); fehler != nil {
		return domain.Kostenposten{}, fehler
	}

	var gebucht domain.Kostenposten

	fehler := imGeschaeft(ctx, j.db, "die Reservierung", func(geschaeft *sql.Tx) error {
		// **Zuerst die Identität** — vor jeder Rechnung. Wer wiederkommt, soll
		// nicht an einer Grenze scheitern, die sein eigener Posten füllt.
		vorhanden, gefunden, lesefehler := postenLesen(ctx, geschaeft, posten.ProfilID, posten.Kennung)
		if lesefehler != nil {
			return lesefehler
		}

		if gefunden {
			// **Dieselbe Kennung muss denselben Auftrag meinen** — Befund 8
			// des Zweitblicks über Phase 5.2. Vorher kam der vorhandene Posten
			// zurück, ohne dass irgendetwas verglichen wurde: eine
			// wiederverwendete Kennung ließ damit eine **fremde** Reservierung
			// wie die Freigabe des neuen Auftrags aussehen.
			if abweichung := vorhanden.MeintDasselbeWie(posten); abweichung != nil {
				return abweichung
			}

			gebucht = vorhanden

			return nil
		}

		if passfehler := passtInsBudget(ctx, geschaeft, posten, zeitraum, jetzt); passfehler != nil {
			return passfehler
		}

		posten.Stand = domain.KostenReserviert
		posten.Begonnen = jetzt

		_, schreibfehler := geschaeft.ExecContext(ctx, `
			INSERT INTO kostenposten (
				kennung, profil_id, konto, aufgabe,
				hoechstbetrag, betrag, waehrung, stand, grund, begonnen
			) VALUES (?, ?, ?, ?, ?, 0, ?, ?, '', ?)`,
			posten.Kennung, posten.ProfilID, posten.Konto, posten.Aufgabe,
			posten.Hoechstbetrag.Wert, posten.Hoechstbetrag.Waehrung,
			string(posten.Stand), Zeitstempel(jetzt))
		if schreibfehler != nil {
			return uebersetze(schreibfehler, "den Auftrag %q buchen", posten.Kennung)
		}

		gebucht = posten

		return nil
	})
	if fehler != nil {
		return domain.Kostenposten{}, fehler
	}

	return gebucht, nil
}

// passtInsBudget prüft die Grenze gegen die Summe — **in derselben
// Transaktion**, in der gebucht wird.
func passtInsBudget(
	ctx context.Context, geschaeft *sql.Tx, posten domain.Kostenposten,
	zeitraum domain.Zeitraum, jetzt time.Time,
) error {
	var grenze domain.Betrag

	fehler := geschaeft.QueryRowContext(ctx, `
		SELECT betrag, waehrung FROM kostenbudgets
		 WHERE profil_id = ? AND konto = ? AND zeitraum = ?`,
		posten.ProfilID, posten.Konto, string(zeitraum)).
		Scan(&grenze.Wert, &grenze.Waehrung)

	if errors.Is(fehler, sql.ErrNoRows) || (fehler == nil && grenze.Leer()) {
		return fmt.Errorf("%w: für %q ist kein Budget freigegeben — ohne "+
			"Freigabe geht kein kostenpflichtiger Auftrag hinaus",
			domain.ErrErlaubnis, posten.Konto)
	}

	if fehler != nil {
		return uebersetze(fehler, "das Budget für %q lesen", posten.Konto)
	}

	gebunden, summenfehler := gebundenIm(ctx, geschaeft, posten, zeitraum, jetzt)
	if summenfehler != nil {
		return summenfehler
	}

	zusammen, rechenfehler := gebunden.Plus(posten.Hoechstbetrag)
	if rechenfehler != nil {
		return rechenfehler
	}

	ueber, vergleichsfehler := zusammen.Uebersteigt(grenze)
	if vergleichsfehler != nil {
		return vergleichsfehler
	}

	if ueber {
		return fmt.Errorf("%w: das Budget für %q trägt diesen Auftrag nicht "+
			"— gebunden sind %d, die Grenze ist %d %s",
			domain.ErrConflict, posten.Konto, gebunden.Wert, grenze.Wert, grenze.Waehrung)
	}

	return nil
}

// gebundenIm summiert, was im laufenden Zeitraum an diesem Konto hängt.
//
// **Gerechnet wird über die Stände und nicht über eine gespeicherte Summe**:
// eine zweite Zahl daneben liefe an dem Tag auseinander, an dem ein Posten
// seinen Stand ändert und niemand sie nachzieht.
func gebundenIm(
	ctx context.Context, geschaeft *sql.Tx, posten domain.Kostenposten,
	zeitraum domain.Zeitraum, jetzt time.Time,
) (domain.Betrag, error) {
	zeilen, fehler := geschaeft.QueryContext(ctx, `
		SELECT hoechstbetrag, betrag, waehrung, stand FROM kostenposten
		 WHERE profil_id = ? AND konto = ? AND begonnen >= ?`,
		posten.ProfilID, posten.Konto, Zeitstempel(zeitraum.Fenster(jetzt)))
	if fehler != nil {
		return domain.Betrag{}, uebersetze(fehler, "die Posten von %q zählen", posten.Konto)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	summe := domain.Betrag{}

	for zeilen.Next() {
		var (
			einer domain.Kostenposten
			stand string
		)

		if lesefehler := zeilen.Scan(&einer.Hoechstbetrag.Wert, &einer.Betrag.Wert,
			&einer.Hoechstbetrag.Waehrung, &stand); lesefehler != nil {
			return domain.Betrag{}, uebersetze(lesefehler, "einen Posten lesen")
		}

		einer.Betrag.Waehrung = einer.Hoechstbetrag.Waehrung
		einer.Stand = domain.Kostenstand(stand)

		naechste, rechenfehler := summe.Plus(einer.Gebunden())
		if rechenfehler != nil {
			return domain.Betrag{}, rechenfehler
		}

		summe = naechste
	}

	if fehler := zeilen.Err(); fehler != nil {
		return domain.Betrag{}, uebersetze(fehler, "die Posten von %q zählen", posten.Konto)
	}

	return summe, nil
}

// postenLesen holt einen Posten über seine Identität.
func postenLesen(ctx context.Context, geschaeft *sql.Tx, profilID int64, kennung string) (
	domain.Kostenposten, bool, error,
) {
	var (
		posten            domain.Kostenposten
		stand             string
		begonnen, beendet string
	)

	fehler := geschaeft.QueryRowContext(ctx, `
		SELECT kennung, profil_id, konto, aufgabe, hoechstbetrag, betrag,
		       waehrung, stand, grund, begonnen, beendet
		  FROM kostenposten WHERE profil_id = ? AND kennung = ?`, profilID, kennung).
		Scan(&posten.Kennung, &posten.ProfilID, &posten.Konto, &posten.Aufgabe,
			&posten.Hoechstbetrag.Wert, &posten.Betrag.Wert,
			&posten.Hoechstbetrag.Waehrung, &stand, &posten.Grund, &begonnen, &beendet)

	if errors.Is(fehler, sql.ErrNoRows) {
		return domain.Kostenposten{}, false, nil
	}

	if fehler != nil {
		return domain.Kostenposten{}, false, uebersetze(fehler,
			"den Auftrag %q lesen", kennung)
	}

	posten.Betrag.Waehrung = posten.Hoechstbetrag.Waehrung
	posten.Stand = domain.Kostenstand(stand)

	// **Ein unlesbarer Zeitpunkt wird zur Nullzeit und nicht zum Fehler.**
	// Er ist die Auskunft „wann", nicht die Buchung; wer daran das Lesen
	// eines Postens scheitern ließe, machte aus einer krummen Zeile eine
	// Sperre für das ganze Journal.
	posten.Begonnen = zeitpunktAus(begonnen)
	posten.Beendet = zeitpunktAus(beendet)

	return posten, true, nil
}

// zeitpunktAus liest die Form dieser Datenbank zurück — leer bleibt leer.
func zeitpunktAus(roh string) time.Time {
	if roh == "" {
		return time.Time{}
	}

	gelesen, fehler := time.Parse(time.RFC3339Nano, roh)
	if fehler != nil {
		return time.Time{}
	}

	return gelesen
}
