package storage

// **Abschluss und Protokoll eines Turniers** — DOM-04.AK-04
// (#1699),
// Vermerk [0139](../../docs/entscheidungen/0139-wann-ein-turnier-abgeschlossen-ist.md).
//
// **Die zweite Reihe**: der Dienst weist ein abgeschlossenes Turnier ab, und
// das Fortschreiben der Fassung trifft keines — ein Weg am Dienst vorbei
// schriebe sonst trotzdem.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Abschliessen setzt das Datum des Abschlusses — gegen die Fassung.
func (a *Turnierablage) Abschliessen(ctx context.Context, turnierID, fassung int64) error {
	return a.abschlussSetzen(ctx, turnierID, fassung, true)
}

// Wiedereroeffnen nimmt das Datum wieder weg — gegen die Fassung.
func (a *Turnierablage) Wiedereroeffnen(ctx context.Context, turnierID, fassung int64) error {
	return a.abschlussSetzen(ctx, turnierID, fassung, false)
}

// abschlussSetzen ist der gemeinsame Rumpf beider Richtungen.
//
// **Nur in die Richtung, in der es etwas ändert**: ein offenes Turnier lässt
// sich nicht wiedereröffnen und ein abgeschlossenes nicht noch einmal
// abschließen — sonst stünden zwei Zeilen im Protokoll für nichts.
func (a *Turnierablage) abschlussSetzen(ctx context.Context, turnierID, fassung int64, abschliessen bool) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	zug, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "den Abschluss beginnen")
	}
	defer zug.Rollback() //nolint:errcheck // nach einem Commit ist es ein No-op

	abfrage := `UPDATE turniere
	               SET abgeschlossen = strftime('%Y-%m-%dT%H:%M:%SZ', 'now'),
	                   fassung = fassung + 1, geaendert = datetime('now')
	             WHERE id = ? AND fassung = ? AND profil_id = ? AND abgeschlossen IS NULL`
	eintrag := domain.Protokolleintrag{Vorgang: domain.TurnierAbgeschlossen}

	if !abschliessen {
		abfrage = `UPDATE turniere
		              SET abgeschlossen = NULL,
		                  fassung = fassung + 1, geaendert = datetime('now')
		            WHERE id = ? AND fassung = ? AND profil_id = ? AND abgeschlossen IS NOT NULL`
		eintrag.Vorgang = domain.TurnierWiedereroeffnet
	}

	ergebnis, err := zug.ExecContext(ctx, abfrage, turnierID, fassung, profilID)
	if err != nil {
		return uebersetze(err, "den Abschluss setzen")
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return zaehlfehler
	}

	if zeilen == 0 {
		return warumNichtFortgeschrieben(ctx, zug, turnierID, fassung, profilID, !abschliessen)
	}

	if fehler := protokollieren(ctx, zug, turnierID, eintrag); fehler != nil {
		return fehler
	}

	return uebersetze(zug.Commit(), "den Abschluss abschließen")
}

// protokollieren schreibt Zeilen des Protokolls — **im Zug des Rufers**, in
// der Reihenfolge, in der sie kommen.
func protokollieren(ctx context.Context, wo execer, turnierID int64, eintraege ...domain.Protokolleintrag) error {
	for _, eintrag := range eintraege {
		if _, err := wo.ExecContext(ctx,
			`INSERT INTO turnierprotokoll (turnier_id, vorgang, angaben, paarung)
			 VALUES (?, ?, ?, ?)`,
			turnierID, string(eintrag.Vorgang), eintrag.Angaben, kennungOderNull(eintrag.PaarungID)); err != nil {
			return uebersetze(err, "das Protokoll schreiben")
		}
	}

	return nil
}

// Protokoll gibt die Zeilen eines Turniers — in der Reihenfolge, in der sie
// entstanden.
func (a *Turnierablage) Protokoll(ctx context.Context, turnierID int64) (
	[]domain.Protokolleintrag, error,
) {
	if fehler := turnierImProfil(ctx, a.db, turnierID); fehler != nil {
		return nil, fehler
	}

	zeilen, err := a.db.QueryContext(ctx,
		`SELECT zeit, vorgang, angaben, coalesce(paarung, 0)
		   FROM turnierprotokoll WHERE turnier_id = ? ORDER BY id`, turnierID)
	if err != nil {
		return nil, uebersetze(err, "das Protokoll lesen")
	}
	defer zeilen.Close() //nolint:errcheck // gelesen wird der Fehler von Err()

	protokoll := []domain.Protokolleintrag{}

	for zeilen.Next() {
		var (
			eintrag domain.Protokolleintrag
			vorgang string
		)

		if err := zeilen.Scan(&eintrag.Zeit, &vorgang, &eintrag.Angaben, &eintrag.PaarungID); err != nil {
			return nil, uebersetze(err, "eine Zeile des Protokolls lesen")
		}

		eintrag.Vorgang = domain.Turniervorgang(vorgang)
		protokoll = append(protokoll, eintrag)
	}

	return protokoll, uebersetze(zeilen.Err(), "das Protokoll lesen")
}

// ergebniseintrag ist die Zeile zu einem Ergebnis — **der Spielstand als
// `a:b:u`**, sprachfrei; `nil` heißt zurückgenommen.
func ergebniseintrag(paarungID int64, stand *domain.Ergebnis) domain.Protokolleintrag {
	if stand == nil {
		return domain.Protokolleintrag{Vorgang: domain.TurnierZurueckgenommen, PaarungID: paarungID}
	}

	return domain.Protokolleintrag{
		Vorgang: domain.TurnierErgebnis, PaarungID: paarungID,
		Angaben: fmt.Sprintf("%d:%d:%d", stand.SpieleA, stand.SpieleB, stand.Unentschieden),
	}
}
