package storage

// **Der Druck an den fünf Nähten** — Werkstück A2 von Phase 5.3
// (#1441, 0086).
//
// Sammlung, Verweis, Deckplatz, Preisstand und Deckabzug nennen einen Druck.
// Bis zur Wanderung 0051 war das allein der Scryfall-Alias `printing`; seither
// steht `druck_id` daneben. **Die Ablage hält beides zusammen**: wer einen
// Alias schreibt, bekommt die Zeile des Drucks dazu — angelegt, wenn es sie
// noch nicht gibt, sonst wiedergefunden —, und wer nur eine Zeile nennt, weil
// der Druck keinen Alias hat, schreibt genau die.
//
// # Warum das hier steht und nicht im Dienst
//
// Es ist dieselbe Zuordnung, die die Wanderung für den Altbestand vollzogen
// hat: ein Alias ist eine Scryfall-Druckkennung. Das ist keine fachliche
// Entscheidung, sondern die Bedeutung der Spalte — und sie an fünf Diensten
// zu wiederholen hieße, sie fünfmal zu vergessen.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// abfrager ist, was `*sql.DB` und `*sql.Tx` gemeinsam können.
//
// **`QueryContext` kam mit Befund 13 des Zweitblicks über Phase 5.2 dazu**: das
// Fenster der Sammlung liest zweimal und soll dabei **einen** Stand sehen —
// dafür müssen beide Leser über dieselbe Stelle gehen können
// (`imLesegeschaeft` in `abfrager.go`).
type abfrager interface {
	ExecContext(ctx context.Context, abfrage string, werte ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, abfrage string, werte ...any) *sql.Row
	QueryContext(ctx context.Context, abfrage string, werte ...any) (*sql.Rows, error)
}

// druckbezug ist, was eine Naht über ihren Druck weiß.
type druckbezug struct {
	druckID  int64
	printing string

	// name, setcode und nummer sind der Rückfall, mit dem ein neuer Druck
	// angelegt wird — der Schnappschuss der Naht, wie in der Wanderung.
	name, setcode, nummer string
}

// druckZeile gibt die Zeile des Drucks, die eine Naht schreiben soll.
//
// **Der Alias geht vor**: nennt die Naht einen, ist der Druck der mit dieser
// Scryfall-Kennung — vorhanden oder neu. Ohne Alias gilt die genannte Zeile;
// ohne beides `NULL`, und das ist bei einem Deckplatz auf eine eigene Karte
// der Normalfall.
func druckZeile(ctx context.Context, q abfrager, bezug druckbezug) (sql.NullInt64, error) {
	if bezug.printing == "" {
		return sql.NullInt64{Int64: bezug.druckID, Valid: bezug.druckID != 0}, nil
	}

	kennung := domain.Druckkennung{
		System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: bezug.printing,
	}

	id, gefunden, err := besitzer(ctx, q, kennung)
	if err != nil {
		return sql.NullInt64{}, err
	}

	if gefunden {
		return sql.NullInt64{Int64: id, Valid: true}, nil
	}

	// **`RETURNING` statt `LastInsertId`**: ein Weg weniger, den kein echter
	// Treiber hergibt (§3).
	if err := q.QueryRowContext(ctx,
		`INSERT INTO drucke (name, setcode, nummer, sprache) VALUES (?, ?, ?, '') RETURNING id`,
		bezug.name, bezug.setcode, bezug.nummer).Scan(&id); err != nil {
		return sql.NullInt64{}, uebersetze(err, "den Druck %s anlegen", bezug.printing)
	}

	if _, err := q.ExecContext(ctx,
		`INSERT INTO druckkennungen (druck_id, system, semantik, wert) VALUES (?, ?, ?, ?)`,
		id, string(kennung.System), string(kennung.Semantik), kennung.Wert); err != nil {
		return sql.NullInt64{}, uebersetze(err, "die Kennung des Drucks %s anlegen", bezug.printing)
	}

	return sql.NullInt64{Int64: id, Valid: true}, nil
}
