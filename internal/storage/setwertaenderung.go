package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Setangabe ändern** —
// #567.
//
// `set_values` wurde seit Phase 2 geschrieben und seit Phase 9 gelesen —
// **geändert nie**. Ein Set hat 89 Angaben in `magic.mse-game`: Titel,
// Setcode, Copyright, Symbol.
//
// # Warum das eine eigene Datei ist und kein Zweig in `feldaenderung.go`
//
// Die beiden Wege sehen gleich aus und meinen Verschiedenes: dort hängt ein
// Wert an einer **Karte**, hier am **Set**. Ein gemeinsamer Weg mit einem
// Schalter „Karte oder Set" trüge jede Abfrage zweimal — und beim ersten
// Sonderfall stünde er an der falschen Stelle.

// Setwerthandlung ist die Handlung, unter der eine Setänderung im Verlauf
// steht.
//
// **Eine eigene und nicht `feld`.** Die Rücknahme muss unterscheiden, wohin
// sie schreibt: `card_values` oder `set_values`. Bis hierher stand in der
// Spalte immer `feld`, und der Kommentar daneben sagte, sie werde tragen, „was
// Löschen und Anlegen brauchen" — sie trägt jetzt das Erste davon.
const Setwerthandlung = "setwert"

// setwertSetzen schreibt einen Setwert und hängt ihn hinten an, wenn er neu ist.
const setwertSetzen = `
INSERT INTO set_values (set_id, feld, wert, reihenfolge)
VALUES (?, ?, ?, (SELECT coalesce(max(reihenfolge) + 1, 0)
                    FROM set_values WHERE set_id = ?))
ON CONFLICT (set_id, feld) DO UPDATE SET wert = excluded.wert`

// setwertSetzenAnStelle schreibt einen Setwert **an eine gegebene Stelle**.
//
// Der Weg zurück aus dem Verlauf (#160): eine Angabe, die entfernt und deren
// Entfernung zurückgenommen wurde, gehört dorthin, wo sie stand — nicht ans
// Ende. `setwertSetzen` kann das nicht, weil es die Stelle selbst vergibt.
const setwertSetzenAnStelle = `
INSERT INTO set_values (set_id, feld, wert, reihenfolge)
VALUES (?, ?, ?, ?)
ON CONFLICT (set_id, feld) DO UPDATE SET wert = excluded.wert,
                                         reihenfolge = excluded.reihenfolge`

// SetwertSetzenMitVorwert schreibt eine Setangabe und gibt die vorige zurück.
//
// **In einer Transaktion**, aus demselben Grund wie bei einem Kartenfeld
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)): wer erst
// liest und dann schreibt, kann die zugesagte Rücknahme auf einen Stand führen,
// den es nie gab.
//
// Der zweite Rückgabewert unterscheidet **„war leer"** von **„gab es nicht"** —
// eine leere Zeile ist nicht dasselbe wie keine Zeile.
func (k *Kartenschreiber) SetwertSetzenMitVorwert(
	ctx context.Context, setID int64, feld, wert string,
) (vorher string, hatte bool, fehler error) {
	// **Auf dem Schreibweg und nicht an der Fassade** (#89): so gilt die Regel
	// für jeden Schreiber, auch für den, den es noch nicht gibt.
	if befund := domain.PruefeFeldname(feld); befund != nil {
		return "", false, befund
	}

	tx, err := k.db.BeginTx(ctx, nil)
	if err != nil {
		return "", false, uebersetze(err, "die Setangabe %q ändern", feld)
	}

	defer func() { fehler = zurueckgenommen(tx, fehler) }()

	// **Die Frage nach dem Profil stellt `alterSetwert`** — in derselben
	// Abfrage wie die nach dem alten Wert, und schon vor #1815.
	vorher, hatte, alteStelle, err := alterSetwert(ctx, tx, setID, feld)
	if err != nil {
		return "", false, err
	}

	if fremd := fremderWert(ctx, feld, vorher); fremd != nil {
		return "", false, fremd
	}

	if _, err := tx.ExecContext(ctx, setwertSetzen, setID, feld, wert, setID); err != nil {
		return "", false, uebersetze(err, "die Setangabe %q ändern", feld)
	}

	if err := k.setwertVerlaufen(ctx, tx, setID, feld, vorher, hatte, alteStelle); err != nil {
		return "", false, err
	}

	if err := tx.Commit(); err != nil {
		return "", false, uebersetze(err, "die Setangabe %q ändern", feld)
	}

	return vorher, hatte, nil
}

// alterSetwert liest, was vor der Änderung dastand — und ob es das Set gibt.
//
// **Eine Abfrage für beides.** Ob es das Set gibt und was im Feld stand, sind
// zwei Fragen an denselben Stand; zwei Abfragen wären zwei Stände, und der
// Unterschied fiele nur unter Last auf.
func alterSetwert(ctx context.Context, tx *sql.Tx, setID int64, feld string) (
	vorher string, hatte bool, stelle *int64, fehler error,
) {
	var (
		setDa      sql.NullInt64
		alterWert  sql.NullString
		alteStelle sql.NullInt64
	)

	// **Jede Wurzelabfrage klemmt auf ein Profil** (#1351, Vermerk 0066).
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return "", false, nil, profilfehler
	}

	err := tx.QueryRowContext(ctx, `
		SELECT (SELECT 1 FROM sets WHERE id = ? AND profil_id = ?),
		       (SELECT wert FROM set_values WHERE set_id = ? AND feld = ?),
		       (SELECT reihenfolge FROM set_values WHERE set_id = ? AND feld = ?)`,
		setID, profilID, setID, feld, setID, feld).
		Scan(&setDa, &alterWert, &alteStelle)
	if err != nil {
		return "", false, nil, uebersetze(err, "die vorige Setangabe %q lesen", feld)
	}

	if !setDa.Valid {
		return "", false, nil, fmt.Errorf("%w: das Set %d liegt nicht im Bestand",
			domain.ErrNotFound, setID)
	}

	return alterWert.String, alterWert.Valid, stelleAus(alteStelle), nil
}

// setwertVerlaufen schreibt den Verlaufseintrag einer Setänderung.
//
// **Ohne Kartenkennung.** Eine Setangabe gehört keiner Karte; `karte_id` bleibt
// auf ihrer Vorgabe null, und die Handlung sagt, wohin die Rücknahme schreibt.
func (k *Kartenschreiber) setwertVerlaufen(
	ctx context.Context, tx *sql.Tx, setID int64, feld, vorher string, hatte bool,
	alteStelle *int64,
) error {
	var (
		neuerWert  sql.NullString
		neueStelle sql.NullInt64
	)

	err := tx.QueryRowContext(ctx,
		`SELECT wert, reihenfolge FROM set_values WHERE set_id = ? AND feld = ?`,
		setID, feld).Scan(&neuerWert, &neueStelle)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return uebersetze(err, "die neue Setangabe %q lesen", feld)
	}

	hin, _ := json.Marshal(feldweg{ //nolint:errcheck // ein Verbund aus Text und Wahrheit
		Feld: feld, Wert: neuerWert.String, GabEs: neuerWert.Valid,
		Reihenfolge: stelleAus(neueStelle),
	})
	zurueck, _ := json.Marshal(feldweg{ //nolint:errcheck // dito
		Feld: feld, Wert: vorher, GabEs: hatte, Reihenfolge: alteStelle,
	})

	_, err = tx.ExecContext(ctx,
		`INSERT INTO history
		        (set_id, karte_id, zeitpunkt, handlung, hinweg, rueckweg, urheber, gruppe)
		 VALUES (?, 0, ?, ?, ?, ?, ?, '')`,
		setID, time.Now().UTC().Format(time.RFC3339Nano), Setwerthandlung,
		string(hin), string(zurueck), string(k.urheber))
	if err != nil {
		return uebersetze(err, "den Verlauf zur Setangabe %q schreiben", feld)
	}

	return kuerzeSetverlauf(ctx, tx, setID)
}
