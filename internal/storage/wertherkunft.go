// **Woher der Wert stammt, der jetzt dasteht**
// (#1025,
// #1036).
//
// Eine eigene Datei, weil `feldaenderung.go` mit ihr über 400 Zeilen ginge
// (§2) — und weil der Schnitt einer Frage folgt: dort steht, **wie** ein Feld
// geändert wird, hier, **was dabei über seine Herkunft** geschieht.

// # Die Zusage aus 0014, und wo sie bis zum 2026-08-24 nicht galt
//
// Die Migration sagt: „Beide werden in derselben Transaktion geführt […]
// Auseinanderlaufen können sie nicht." Für den **direkten** Schreibweg stimmte
// das; die **Rücknahme** ging nie durch ihn — `kartenwegGehen` schrieb
// unmittelbar nach `card_values`. Nach der Rücknahme eines Scryfall-Bildes
// blieb die Nennung an einem wiederhergestellten eigenen Bild hängen: genau
// die falsche Nennung, die #1025 ausschließen sollte.
//
// **Seit #1036 trägt der Verlaufseintrag die Herkunft mit** (`feldweg.Herkunft`)
// — ohne sie wäre die des vorigen Werts nach der Änderung nicht mehr zu
// rekonstruieren.
//
// **Der Text in 0014 bleibt, wie er ist.** Eine angewendete Migration wird
// nicht mehr geändert (die Lehre aus 0007); was sie zu weit zusagte, steht
// hier — an der Stelle, an der die Zusage eingelöst wird.
//
// # Was weiterhin offen ist
//
// Eine **gelöschte** Karte nimmt ihre Herkunftszeilen über `ON DELETE CASCADE`
// mit, und das Wiederherstellen baut nur ihre Werte wieder auf. Ein
// zurückgeholtes Scryfall-Bild steht danach ohne seine Nennung da —
// #1036
// führt es weiter.

package storage

import (
	"context"
	"database/sql"
	"errors"
)

// herkunftFuehren hält fest, woher der Wert stammt, der **jetzt** dasteht.
//
// # Warum hier und nicht neben dem Schreiben
//
// **Weil hier jede Änderung vorbeikommt.** Diese Methode läuft in derselben
// Transaktion wie die Änderung selbst und für jedes geänderte Feld — einzeln
// geschrieben, in einem Verbund geschrieben, zurückgenommen. Wer die Zeile
// beim Aufrufer pflegte, hätte sie an vier Stellen und vergäße die fünfte;
// die Tabelle behauptete dann etwas über einen Wert, der nicht mehr dasteht.
//
// **Ein eigener Wert löscht die fremde Herkunft**, und das ist die Hälfte,
// die man vergisst: bliebe die Zeile stehen, behauptete die Anwendung über ein
// selbst gewähltes Foto, es sei ein Bildausschnitt von Scryfall — samt fremdem
// Copyright. Eine falsche Nennung ist schlimmer als keine
// (#1025).
//
// **Auch eine Rücknahme kommt hier vorbei.** Sie schreibt ohne Herkunft, und
// damit fällt die Zeile — richtig so: der Wert von vorher stammt nicht von
// dort, wo der zurückgenommene herkam.
func (k *Kartenschreiber) herkunftFuehren(
	ctx context.Context, tx *sql.Tx, karteID int64, feld string,
) error {
	if k.herkunft == "" {
		_, err := tx.ExecContext(ctx,
			`DELETE FROM wert_herkunft WHERE card_id = ? AND feld = ?`, karteID, feld)
		if err != nil {
			return uebersetze(err, "die Herkunft von %q löschen", feld)
		}

		return nil
	}

	_, err := tx.ExecContext(ctx,
		`INSERT INTO wert_herkunft (card_id, feld, herkunft) VALUES (?, ?, ?)
		 ON CONFLICT (card_id, feld) DO UPDATE SET herkunft = excluded.herkunft`,
		karteID, feld, k.herkunft)
	if err != nil {
		return uebersetze(err, "die Herkunft von %q festhalten", feld)
	}

	return nil
}

// herkunftLesen gibt die Herkunft, die ein Feld **jetzt** trägt.
//
// **Leer heißt: von hier** — dieselbe Übereinkunft wie überall sonst. Eine
// fehlende Zeile ist keine Ausnahme, sondern der Normalfall.
func (k *Kartenschreiber) herkunftLesen(
	ctx context.Context, tx *sql.Tx, karteID int64, feld string,
) (string, error) {
	var herkunft sql.NullString

	err := tx.QueryRowContext(ctx,
		`SELECT herkunft FROM wert_herkunft WHERE card_id = ? AND feld = ?`,
		karteID, feld).Scan(&herkunft)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", uebersetze(err, "die Herkunft von %q lesen", feld)
	}

	return herkunft.String, nil
}

// zeilenfrager ist alles, was mehrere Zeilen hergibt — Vorrat wie
// Transaktion. **`*sql.DB` und `*sql.Tx` erfüllen sie beide**, und genau
// darum gibt es sie: dieselbe Abfrage wird einmal von außen gelesen (der
// Leser) und einmal **innerhalb** der Löschtransaktion, wo der Stand einer
// sein muss (#1036).
type zeilenfrager interface {
	QueryContext(ctx context.Context, abfrage string, werte ...any) (*sql.Rows, error)
}

// herkuenfteVon gibt je Feld die Adresse, aus der sein aktueller Wert stammt.
//
// **Eine Fassung für beide Aufrufer.** Zwei gleichlautende Abfragen
// nebeneinander liefen beim nächsten Schemaschritt auseinander, und die
// zweite fiele niemandem auf — sie stünde in einer anderen Datei.
func herkuenfteVon(
	ctx context.Context, wo zeilenfrager, karteID int64,
) (map[string]string, error) {
	zeilen, err := wo.QueryContext(ctx,
		`SELECT feld, herkunft FROM wert_herkunft WHERE card_id = ?`, karteID)
	if err != nil {
		return nil, uebersetze(err, "die Herkünfte der Karte %d lesen", karteID)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen ist gelesen

	herkuenfte := map[string]string{}

	for zeilen.Next() {
		var feld, herkunft string
		if fehler := zeilen.Scan(&feld, &herkunft); fehler != nil {
			return nil, uebersetze(fehler, "eine Herkunft lesen")
		}

		herkuenfte[feld] = herkunft
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, uebersetze(fehler, "die Herkünfte der Karte %d lesen", karteID)
	}

	return herkuenfte, nil
}
