// **Die Rücknahme greift auf Schlüsselwörter**
// (#1077).
//
// # Was fehlte
//
// ⌘Z nahm eine Schlüsselwortänderung **nicht** zurück. Anlegen, Ändern und
// Löschen gingen an `history` vorbei; der Verlauf kannte sie nicht. MSE hat
// sie im selben Rücknahmestapel wie alles andere — `AddKeywordAction` und
// `ChangeKeywordModeAction` sind `Action`s (`src/data/action/keyword.cpp`,
// Fassung f9d9356 vom 2021-07-06).
//
// # Eine Handlungsart, nicht drei
//
// Das Issue nannte drei — setzen, anlegen, löschen. Es sind zwei **Stände**:
// anlegen heißt „vorher nicht da", löschen „nachher nicht da", ändern „beides
// da". Dieselbe Gestalt wie bei einer ganzen Karte
// (#621), und
// aus demselben Grund: eine Rücknahme, die drei Arten unterscheiden müsste,
// unterschiede sie an drei Stellen.

package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// schluesselwortstand liest den Stand eines Schlüsselworts.
//
// **Ein Wort, das es nicht gibt, ist kein Fehler**: es ist die Seite eines
// Schritts, auf der es nicht dasteht. Wer daraus einen Fehler machte, könnte
// ein Anlegen nicht protokollieren.
func schluesselwortstand(
	ctx context.Context, tx *sql.Tx, setID, wortID int64,
) (domain.Schluesselwortstand, error) {
	stand := domain.Schluesselwortstand{ID: wortID}

	err := tx.QueryRowContext(ctx,
		`SELECT reihenfolge, wort, muster, erinnerung, regeln, modus
		   FROM set_keywords WHERE id = ? AND set_id = ?`, wortID, setID).
		Scan(&stand.Reihenfolge, &stand.Wort, &stand.Muster,
			&stand.Erinnerung, &stand.Regeln, &stand.Modus)

	switch {
	case err == nil:
		stand.GabEs = true

		return stand, nil

	case errors.Is(err, sql.ErrNoRows):
		return domain.Schluesselwortstand{ID: wortID}, nil

	default:
		return domain.Schluesselwortstand{},
			uebersetze(err, "den Stand des Schlüsselworts %d lesen", wortID)
	}
}

// schreibeSchluesselwortverlauf legt den Schritt an — **in der Transaktion des
// Aufrufers**, aus demselben Grund wie [schreibeVerlauf]: die Änderung und ihr
// Weg zurück gehören in denselben Zug, sonst gibt es einen Stand ohne Rückweg.
func schreibeSchluesselwortverlauf(
	ctx context.Context, tx *sql.Tx, setID int64,
	vorher, nachher domain.Schluesselwortstand, wer domain.Urheber,
) error {
	hin, _ := json.Marshal(nachher)    //nolint:errcheck // ein Verbund aus Text und Zahlen
	zurueck, _ := json.Marshal(vorher) //nolint:errcheck // dito

	// **`karte_id` bleibt 0.** Ein Schlüsselwort gehört keiner Karte; seine
	// Kennung steht im Stand. Dieselbe Wahl wie bei den Setangaben.
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO history
		        (set_id, karte_id, zeitpunkt, handlung, hinweg, rueckweg, urheber, gruppe)
		 VALUES (?, 0, ?, ?, ?, ?, ?, '')`,
		setID, time.Now().UTC().Format(time.RFC3339Nano), domain.HandlungSchluesselwort,
		string(hin), string(zurueck), string(wer)); err != nil {
		return uebersetze(err, "den Verlauf zum Schlüsselwort schreiben")
	}

	return kuerzeSetverlauf(ctx, tx, setID)
}

// schluesselwortGehen stellt einen Stand her.
//
// **Drei Fälle, wie bei einer ganzen Karte**: es soll nicht dastehen (dann
// weg damit), es steht schon da (dann überschreiben), oder es fehlt (dann
// anlegen — **mit seiner Kennung**, damit ein zweiter Schritt desselben Worts
// es wiederfindet).
//
// **Der Stand ist nie `nil`.** [schluesselwortschrittAus] setzt beide Seiten
// oder gibt einen Fehler, und nur ein Eintrag der Art `schluesselwort` kommt
// hierher. Eine Prüfung darauf wäre ein Zweig, den kein Aufruf betritt — und §3
// sagt, was ein solcher Zweig wert ist.
func schluesselwortGehen(
	ctx context.Context, tx *sql.Tx, setID int64, stand *domain.Schluesselwortstand,
) error {
	if !stand.GabEs {
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM set_keywords WHERE id = ? AND set_id = ?`,
			stand.ID, setID); err != nil {
			return uebersetze(err, "das Schlüsselwort %d entfernen", stand.ID)
		}

		return nil
	}

	// **`INSERT OR REPLACE` und nicht erst prüfen.** Beide Fälle — es fehlt
	// oder es steht anders da — enden in derselben Zeile, und eine Abfrage
	// davor wäre ein zweites Zeitfenster in derselben Transaktion.
	if _, err := tx.ExecContext(ctx,
		`INSERT OR REPLACE INTO set_keywords
		        (id, set_id, reihenfolge, wort, muster, erinnerung, regeln, modus)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		stand.ID, setID, stand.Reihenfolge, stand.Wort, stand.Muster,
		stand.Erinnerung, stand.Regeln, stand.Modus); err != nil {
		return uebersetze(err, "das Schlüsselwort %d zurückbringen", stand.ID)
	}

	return nil
}

// schluesselwortschrittAus füllt einen Eintrag der Art `schluesselwort`.
func schluesselwortschrittAus(
	eintrag domain.Verlaufseintrag, hin, zurueck string,
	zurueckgenommen int, urheber string,
) (domain.Verlaufseintrag, error) {
	nachher, err := schluesselwortstandAus(hin)
	if err != nil {
		return domain.Verlaufseintrag{}, err
	}

	vorher, err := schluesselwortstandAus(zurueck)
	if err != nil {
		return domain.Verlaufseintrag{}, err
	}

	eintrag.VorherWort, eintrag.NachherWort = vorher, nachher
	eintrag.Zurueckgenommen = zurueckgenommen != 0
	eintrag.Urheber = domain.Urheber(urheber)

	return eintrag, nil
}

// schluesselwortstandAus liest einen Stand aus seinem JSON.
func schluesselwortstandAus(roh string) (*domain.Schluesselwortstand, error) {
	var stand domain.Schluesselwortstand

	if err := json.Unmarshal([]byte(roh), &stand); err != nil {
		return nil, uebersetze(err, "einen Schlüsselwortstand lesen")
	}

	return &stand, nil
}
