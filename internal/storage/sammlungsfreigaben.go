package storage

// **Freigegebene Sammlungssichten im Bestand** — Werkstück `T3` von Phase 5.6
// (#1468).
//
// # Warum geprüft wird, bevor geschrieben wird
//
// **Weil eine unleserliche Freigabe schlimmer ist als keine.** `Zeigt` gibt
// für einen unbekannten Umfang nichts zurück — das ist der sichere Ausgang,
// und er bedeutet: die Zeile steht da und tut nichts. Wer sie erteilt hat,
// glaubt, er habe etwas freigegeben.
//
// # Und warum die Auswahl in einem Zug mit der Freigabe steht
//
// **Weil eine Auswahlfreigabe ohne Auswahl nichts zeigt.** Stünde sie ohne
// ihre Drucke da, hätte jemand eine Freigabe erteilt, die keine ist — und die
// Domäne weist genau das ab. Beides steht oder fällt zusammen.

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Freigabeablage hält die Sammlungsfreigaben eines Profils.
type Freigabeablage struct{ db *sql.DB }

// NeueFreigabeablage baut sie.
func NeueFreigabeablage(db *sql.DB) *Freigabeablage {
	return &Freigabeablage{db: db}
}

// Erteilen legt eine Freigabe an — oder ersetzt die gleichartige.
//
// **Gleichartig heißt: derselbe Empfänger, derselbe Zweck, derselbe Umfang.**
// Wer eine Frist ändert, meint dieselbe Freigabe; wer einen zweiten Umfang
// freigibt, meint eine zweite.
func (a *Freigabeablage) Erteilen(
	ctx context.Context, freigabe domain.Sammlungsfreigabe,
) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	if fehler := freigabe.Validate(); fehler != nil {
		return fehler
	}

	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "eine Freigabe erteilen")
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach einem Commit wirkungslos

	var kennung int64

	if fehler := tx.QueryRowContext(ctx, `
INSERT INTO sammlungsfreigaben
       (profil_id, empfaenger, umfang, zweck, behalten, ablauf, erteilt)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (profil_id, empfaenger, zweck, umfang) DO UPDATE
   SET behalten = excluded.behalten,
       ablauf   = excluded.ablauf,
       erteilt  = excluded.erteilt
RETURNING id`,
		profilID, freigabe.Empfaenger, string(freigabe.Umfang),
		string(freigabe.Zweck), freigabe.Behalten, alsFrist(freigabe.Ablauf),
		freigabe.Erteilt.Format(time.RFC3339Nano),
	).Scan(&kennung); fehler != nil {
		return uebersetze(fehler, "eine Freigabe erteilen")
	}

	if fehler := auswahlSetzen(ctx, tx, kennung, freigabe.Auswahl); fehler != nil {
		return fehler
	}

	return uebersetze(tx.Commit(), "eine Freigabe erteilen")
}

// auswahlSetzen schreibt die gewählten Drucke neu.
//
// **Erst weg, dann hin.** Ein Ersetzen ohne das Löschen ließe die Drucke einer
// früheren Auswahl stehen — und die Freigabe zeigte mehr, als zuletzt gewählt
// wurde.
func auswahlSetzen(
	ctx context.Context, tx *sql.Tx, freigabeID int64, auswahl []string,
) error {
	if _, fehler := tx.ExecContext(ctx,
		`DELETE FROM sammlungsfreigabeauswahl WHERE freigabe_id = ?`,
		freigabeID); fehler != nil {
		return uebersetze(fehler, "die Auswahl räumen")
	}

	for _, printing := range auswahl {
		if _, fehler := tx.ExecContext(ctx,
			`INSERT INTO sammlungsfreigabeauswahl (freigabe_id, printing)
			 VALUES (?, ?)`, freigabeID, printing); fehler != nil {
			return uebersetze(fehler, "die Auswahl schreiben")
		}
	}

	return nil
}

// Widerrufen nimmt genau eine Freigabe weg.
//
// **Und meldet, wenn es keine gab.** „Widerrufen" ohne Wirkung sähe aus wie
// ein Erfolg — etwa für jemanden, der sich im Zweck vertan hat.
func (a *Freigabeablage) Widerrufen(
	ctx context.Context, empfaenger string, zweck domain.Zweck,
	umfang domain.Freigabeumfang,
) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	ergebnis, err := a.db.ExecContext(ctx,
		`DELETE FROM sammlungsfreigaben
		  WHERE profil_id = ? AND empfaenger = ? AND zweck = ? AND umfang = ?`,
		profilID, empfaenger, string(zweck), string(umfang))
	if err != nil {
		return uebersetze(err, "eine Freigabe widerrufen")
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return zaehlfehler
	}

	if zeilen == 0 {
		return fmt.Errorf(
			"%w: für %q gibt es keine solche Freigabe", domain.ErrNotFound,
			empfaenger)
	}

	return nil
}

// Fuer gibt die Freigaben, die einem Empfänger gelten.
//
// **Alle, auch die abgelaufenen.** Ob eine gilt, entscheidet
// [domain.Sammlungsfreigabe.Gilt] gegen eine Uhr — und die steht nicht im
// Bestand. Wer hier filterte, hätte dieselbe Regel zweimal, und die zweite
// Fassung wäre irgendwann eine andere.
func (a *Freigabeablage) Fuer(
	ctx context.Context, empfaenger string,
) ([]domain.Sammlungsfreigabe, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, fehler := a.db.QueryContext(ctx, `
SELECT f.id, f.empfaenger, f.umfang, f.zweck, f.behalten, f.ablauf, f.erteilt
  FROM sammlungsfreigaben f
 WHERE f.profil_id = ? AND f.empfaenger = ?
 ORDER BY f.umfang`, profilID, empfaenger)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Freigaben lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird nur

	gelesen, kennungen, lesefehler := freigabenLesen(zeilen)
	if lesefehler != nil {
		return nil, lesefehler
	}

	return a.mitAuswahl(ctx, gelesen, kennungen)
}

// freigabenLesen liest die Zeilen samt ihren Kennungen.
func freigabenLesen(zeilen *sql.Rows) (
	[]domain.Sammlungsfreigabe, []int64, error,
) {
	gelesen := []domain.Sammlungsfreigabe{}
	kennungen := []int64{}

	for zeilen.Next() {
		var (
			kennung         int64
			umfang, zweck   string
			ablauf, erteilt string
			freigabe        domain.Sammlungsfreigabe
		)

		if fehler := zeilen.Scan(&kennung, &freigabe.Empfaenger, &umfang,
			&zweck, &freigabe.Behalten, &ablauf, &erteilt); fehler != nil {
			return nil, nil, uebersetze(fehler, "eine Freigabe lesen")
		}

		gelesenerAblauf, fristfehler := ausFrist(ablauf)
		if fristfehler != nil {
			return nil, nil, fristfehler
		}

		gelesenErteilt, zeitfehler := ausFrist(erteilt)
		if zeitfehler != nil {
			return nil, nil, zeitfehler
		}

		freigabe.Umfang = domain.Freigabeumfang(umfang)
		freigabe.Zweck = domain.Zweck(zweck)
		freigabe.Ablauf, freigabe.Erteilt = gelesenerAblauf, gelesenErteilt

		gelesen = append(gelesen, freigabe)
		kennungen = append(kennungen, kennung)
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, nil, uebersetze(fehler, "die Freigaben lesen")
	}

	return gelesen, kennungen, nil
}

// mitAuswahl hängt jeder Auswahlfreigabe ihre Drucke an.
func (a *Freigabeablage) mitAuswahl(
	ctx context.Context, gelesen []domain.Sammlungsfreigabe, kennungen []int64,
) ([]domain.Sammlungsfreigabe, error) {
	for stelle := range gelesen {
		if gelesen[stelle].Umfang != domain.UmfangAuswahl {
			continue
		}

		auswahl, fehler := a.auswahlLesen(ctx, kennungen[stelle])
		if fehler != nil {
			return nil, fehler
		}

		gelesen[stelle].Auswahl = auswahl
	}

	return gelesen, nil
}

// auswahlLesen gibt die gewählten Drucke einer Freigabe.
func (a *Freigabeablage) auswahlLesen(
	ctx context.Context, freigabeID int64,
) ([]string, error) {
	zeilen, fehler := a.db.QueryContext(ctx,
		`SELECT printing FROM sammlungsfreigabeauswahl
		  WHERE freigabe_id = ? ORDER BY printing`, freigabeID)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Auswahl lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird nur

	auswahl := []string{}

	for zeilen.Next() {
		var printing string

		if scanfehler := zeilen.Scan(&printing); scanfehler != nil {
			return nil, uebersetze(scanfehler, "einen gewählten Druck lesen")
		}

		auswahl = append(auswahl, printing)
	}

	if endefehler := zeilen.Err(); endefehler != nil {
		return nil, uebersetze(endefehler, "die Auswahl lesen")
	}

	return auswahl, nil
}
