package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KartenTauschen vertauscht die Stellen zweier Karten desselben Sets
// (#628).
//
// # Was portiert ist — und was nicht
//
// **ported: die Absicht.** MSE2 ordnet durch Ziehen in der Kartenliste um und
// löst dabei **je überfahrenem Element** eine eigene Handlung aus, die zwei
// Stellen **tauscht** (`src/gui/control/card_list.cpp:370–388`, Fassung
// f9d9356 vom 2021-07-06). Wer eine Karte von Stelle 5 auf Stelle 2 zieht,
// löst drei Tausche aus; der Tausch mit dem Nachbarn ist der Elementarschritt,
// und genau er steht hier.
//
// **Nicht portiert: der Ausfall.** `ReorderCardsAction::perform` prüft
// `card_id1 >= size || card_id2 < size` und kehrt bei jeder gültigen Auswahl
// zurück, ohne zu tauschen (`src/data/action/set.cpp:54–63`, dieselbe
// Fassung) — die zweite Hälfte der Bedingung ist verdreht. Der Kommentar
// daneben sagt, dass der Verfasser es wusste:
//
//	// TODO : Too lazy to fix this right now.
//
// Eine Referenz, die ihre eigene Absicht nicht ausführt, gibt kein Verhalten
// vor, das zu erhalten wäre (§15.4).
//
// # Warum die Stelle über eine Parkposition geht
//
// `cards` trägt `UNIQUE (set_id, reihenfolge)`. Die erste Karte direkt auf die
// Stelle der zweiten zu setzen verletzte die Zusage, solange die zweite noch
// dort steht. Beide zuerst auf `-1 - reihenfolge` zu schieben hält sie
// untereinander verschieden — die Abbildung ist umkehrbar eindeutig — und
// keine echte Karte steht auf einer negativen Stelle.
func (k *Kartenschreiber) KartenTauschen(ctx context.Context, eineID, andereID int64) error {
	tx, err := k.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "die Karten %d und %d tauschen", eineID, andereID)
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach Commit ein ErrTxDone

	// **Beide Karten gehören dem gezeigten Profil** (#1815).
	if fehler := kartenImProfil(ctx, tx, []int64{eineID, andereID}); fehler != nil {
		return fehler
	}

	// **Eine Karte mit sich selbst zu tauschen ist keine Bewegung.** Sie liefe
	// durch: `IN (?, ?)` gibt eine Zeile, beide Nachschläge finden sie, und
	// heraus käme ein Verlaufsschritt, der nichts tut — den aber ⌘Z verbraucht.
	if eineID == andereID {
		return fmt.Errorf("%w: die Karte %d lässt sich nicht mit sich selbst tauschen",
			domain.ErrValidation, eineID)
	}

	stellen, err := stellenLesen(ctx, tx, eineID, andereID)
	if err != nil {
		return err
	}

	eine, andere := stellen[eineID], stellen[andereID]

	// **Über die Setgrenze wird nicht getauscht.** Stellen zählen je Set; ein
	// Tausch darüber hinweg setzte beide auf eine Stelle, die im anderen Set
	// schon besetzt sein kann — und hinterließe zwei Sets, deren Ordnung
	// niemand angeordnet hat.
	if eine.set != andere.set {
		return fmt.Errorf("%w: die Karten %d und %d liegen in verschiedenen Sets",
			domain.ErrValidation, eineID, andereID)
	}

	hin := domain.Stellenstand{
		Eine: eineID, Andere: andereID,
		EineStelle: andere.stelle, AndereStelle: eine.stelle,
	}
	zurueck := domain.Stellenstand{
		Eine: eineID, Andere: andereID,
		EineStelle: eine.stelle, AndereStelle: andere.stelle,
	}

	if err := schreibeStellenverlauf(ctx, tx, eineID, zurueck, hin, k.urheber); err != nil {
		return err
	}

	if err := stellenGehen(ctx, tx, &hin); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return uebersetze(err, "die Karten %d und %d tauschen", eineID, andereID)
	}

	return nil
}

// kartenstelle ist, wo eine Karte steht — Set und Stelle darin.
type kartenstelle struct{ set, stelle int64 }

// stellenLesen holt beide Karten in **einer** Abfrage.
//
// **Nicht zweimal nacheinander**, und das aus zwei Gründen: zwischen zwei
// Abfragen liegt Zeit, in der sich die zweite Karte bewegen kann — und zwei
// gleichlautende Abfragen lassen sich vom Stolpertreiber nicht auseinanderhalten,
// womit der zweite Fehlerweg unprüfbar wäre (§10).
func stellenLesen(
	ctx context.Context, tx *sql.Tx, eineID, andereID int64,
) (map[int64]kartenstelle, error) {
	zeilen, err := tx.QueryContext(ctx,
		`SELECT id, set_id, reihenfolge FROM cards WHERE id IN (?, ?)`, eineID, andereID)
	if err != nil {
		return nil, uebersetze(err, "die Stellen der Karten %d und %d lesen", eineID, andereID)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen ist gelesen

	stellen := map[int64]kartenstelle{}

	for zeilen.Next() {
		var (
			kennung int64
			stelle  kartenstelle
		)

		if fehler := zeilen.Scan(&kennung, &stelle.set, &stelle.stelle); fehler != nil {
			return nil, uebersetze(fehler, "eine Kartenstelle lesen")
		}

		stellen[kennung] = stelle
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, uebersetze(fehler, "die Stellen der Karten %d und %d lesen", eineID, andereID)
	}

	// **Eine Karte, die es nicht gibt, ist ein Befund** (§8).
	for _, kennung := range []int64{eineID, andereID} {
		if _, da := stellen[kennung]; !da {
			return nil, fmt.Errorf("%w: die Karte %d gibt es nicht",
				domain.ErrNotFound, kennung)
		}
	}

	return stellen, nil
}

// stellenGehen setzt beide Karten auf die Stellen, die der Stand nennt.
//
// **Ohne Richtung**: Hin- und Rückweg unterscheiden sich nur darin, wohin die
// beiden gehören, und diese Stelle führt aus, was dasteht.
func stellenGehen(ctx context.Context, tx *sql.Tx, stand *domain.Stellenstand) error {
	if stand == nil {
		return nil
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE cards SET reihenfolge = -1 - reihenfolge WHERE id IN (?, ?)`,
		stand.Eine, stand.Andere); err != nil {
		return uebersetze(err, "die Karten %d und %d parken", stand.Eine, stand.Andere)
	}

	for _, paar := range [][2]int64{
		{stand.Eine, stand.EineStelle},
		{stand.Andere, stand.AndereStelle},
	} {
		if _, err := tx.ExecContext(ctx,
			`UPDATE cards SET reihenfolge = ? WHERE id = ?`, paar[1], paar[0]); err != nil {
			return uebersetze(err, "die Karte %d auf Stelle %d setzen", paar[0], paar[1])
		}
	}

	return nil
}

// schreibeStellenverlauf legt den Schritt an — in der Transaktion des Aufrufers.
func schreibeStellenverlauf(
	ctx context.Context, tx *sql.Tx, karteID int64,
	vorher, nachher domain.Stellenstand, wer domain.Urheber,
) error {
	hin, _ := json.Marshal(nachher)    //nolint:errcheck // vier Zahlen
	zurueck, _ := json.Marshal(vorher) //nolint:errcheck // dito

	_, err := tx.ExecContext(ctx,
		`INSERT INTO history
		        (set_id, karte_id, zeitpunkt, handlung, hinweg, rueckweg, urheber, gruppe)
		 SELECT set_id, ?, ?, ?, ?, ?, ?, '' FROM cards WHERE id = ?`,
		karteID, time.Now().UTC().Format(time.RFC3339Nano), domain.HandlungReihenfolge,
		string(hin), string(zurueck), string(wer), karteID)
	if err != nil {
		return uebersetze(err, "den Verlauf zur Karte %d schreiben", karteID)
	}

	return kuerzeVerlauf(ctx, tx, karteID)
}

// stellenstandAus liest einen Stellenstand aus seinem JSON.
func stellenstandAus(roh string) (*domain.Stellenstand, error) {
	if roh == "" {
		return &domain.Stellenstand{}, nil
	}

	var stand domain.Stellenstand
	if err := json.Unmarshal([]byte(roh), &stand); err != nil {
		return nil, uebersetze(err, "einen Stellenweg lesen")
	}

	return &stand, nil
}
