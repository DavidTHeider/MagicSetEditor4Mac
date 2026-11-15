package storage

// **Der Stand wird in derselben Transaktion geprüft, in der geschrieben wird**
// — Befund 2 des Zweitblicks über Phase 5.3.
//
// # Der Befund
//
// **Eine Übernahme prüfte den Stand und schrieb ihn dann unbedingt.**
// `service` liest die Revision der Karte, vergleicht sie mit der des Auftrags
// und ruft danach den Schreibweg. Zwischen beidem liegt alles: zwei
// Übernahmen desselben Feldes — eine einzelne und eine aus einem Batch —
// kommen beide durch die Prüfung und schreiben beide. Die zweite übermalt die
// erste samt ihrer Herkunft, **still**.
//
// Das Fenster hält sie nicht auseinander: die Riegel heißen
// `uebernahme/<karte>` und `gruppe/<batch>` und sperren einander nicht; CLI,
// MCP und REST haben gar keinen.
//
// # Warum hier und nicht im Aufrufer
//
// **Weil eine Bedingung nur dort gilt, wo sie mit der Wirkung zusammen
// stattfindet.** Eine Prüfung vor `BeginTx` ist eine Auskunft über die
// Vergangenheit; dieselbe Prüfung **in** der Transaktion ist eine Bedingung.
// Das Haus kennt die Gestalt schon — `WHERE id = ? AND fassung = ?` steht bei
// Decks, Formaten, Nutzern und der Sammlung. Eine Karte führt keine
// `fassung`, sondern einen **gerechneten** Stand ([domain.CardRow.Stand]);
// geprüft wird er deshalb, indem die Karte in der Transaktion noch einmal
// gelesen und ihr Stand gerechnet wird.

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// standHalten weist ab, wenn die Karte nicht mehr auf dem erwarteten Stand ist.
//
// **Ein leerer Erwartungswert heißt: keine Bedingung.** Die alten Schreibwege
// haben keine, und ein Pflichtfeld hier machte aus jedem von ihnen einen
// Aufrufer, der sich einen Stand ausdenken müsste.
func standHalten(
	ctx context.Context, tx *sql.Tx, karteID int64, erwartet string,
) error {
	if erwartet == "" {
		return nil
	}

	// **Ein Rechenstand trägt die Setangaben mit** (#2533) — ein Wertskript
	// liest `set.*`, und der Setteil wird in derselben Transaktion gehalten.
	karte, set := domain.StandTeile(erwartet)

	stand, fehler := standInDerTransaktion(ctx, tx, karteID)
	if fehler != nil {
		return fehler
	}

	if stand != karte {
		return ueberholt()
	}

	if set == "" {
		return nil
	}

	setstand, fehler := setstandInDerTransaktion(ctx, tx, karteID)
	if fehler != nil {
		return fehler
	}

	if setstand != set {
		return ueberholt()
	}

	return nil
}

// ueberholt ist die Antwort auf einen Stand, den es nicht mehr gibt.
func ueberholt() error {
	return fmt.Errorf(
		"%w: die Karte hat sich inzwischen geändert — was geschrieben werden "+
			"sollte, gehört zu einem Stand, den es so nicht mehr gibt; noch einmal versuchen",
		domain.ErrConflict)
}

// standInDerTransaktion rechnet den Stand der Karte aus der laufenden
// Transaktion.
//
// **Dieselbe Rechnung wie [domain.CardRow.Stand]**, und zwar durch sie: ein
// zweiter Weg zu derselben Zahl liefe beim ersten neuen Feld auseinander.
func standInDerTransaktion(
	ctx context.Context, tx *sql.Tx, karteID int64,
) (string, error) {
	zeilen, err := tx.QueryContext(ctx, `
SELECT card_values.feld, card_values.wert
  FROM cards
  LEFT JOIN card_values ON card_values.card_id = cards.id
 WHERE cards.id = ?`, karteID)
	if err != nil {
		return "", uebersetze(err, "Stand der Karte %d lesen", karteID)
	}

	defer func() { _ = zeilen.Close() }()

	karte := domain.CardRow{ID: karteID}
	gibts := false

	for zeilen.Next() {
		var feld, wert sql.NullString

		if err := zeilen.Scan(&feld, &wert); err != nil {
			return "", uebersetze(err, "Stand der Karte %d lesen", karteID)
		}

		gibts = true

		if feld.Valid {
			karte.Values = append(karte.Values,
				domain.Value{Field: feld.String, Text: wert.String})
		}
	}

	if err := zeilen.Err(); err != nil {
		return "", uebersetze(err, "Stand der Karte %d lesen", karteID)
	}

	if !gibts {
		return "", fmt.Errorf("%w: die Karte %d", domain.ErrNotFound, karteID)
	}

	return karte.Stand(), nil
}

// setstandInDerTransaktion rechnet den Stand der Setangaben zur Karte —
// **durch [domain.Setstand]**, aus demselben Grund wie oben.
func setstandInDerTransaktion(
	ctx context.Context, tx *sql.Tx, karteID int64,
) (string, error) {
	zeilen, err := tx.QueryContext(ctx, `
SELECT feld, wert FROM set_values
 WHERE set_id = (SELECT set_id FROM cards WHERE id = ?)
 ORDER BY reihenfolge`, karteID)
	if err != nil {
		return "", uebersetze(err, "Setangaben zur Karte %d lesen", karteID)
	}

	defer func() { _ = zeilen.Close() }()

	var werte []domain.Value

	for zeilen.Next() {
		var wert domain.Value

		if err := zeilen.Scan(&wert.Field, &wert.Text); err != nil {
			return "", uebersetze(err, "Setangaben zur Karte %d lesen", karteID)
		}

		werte = append(werte, wert)
	}

	if err := zeilen.Err(); err != nil {
		return "", uebersetze(err, "Setangaben zur Karte %d lesen", karteID)
	}

	return domain.Setstand(werte), nil
}
