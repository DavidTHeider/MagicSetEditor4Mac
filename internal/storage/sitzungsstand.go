package storage

// **Die festgehaltene Sitzung** — ID-02
// (#1701).
//
// **Im Bestand und nicht im Arbeitsspeicher**, aus demselben Grund wie der
// Profilzeiger: eine Sperre, die ⌘Q und ein Neustart aufheben, wäre keine.
// **Wer gesperrt ist, sagt der Zeiger auf den arbeitenden Nutzer** — hier
// steht nur der Zustand.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sitzungsschluessel ist der Eintrag im Bestand, der eine gesperrte oder
// beendete Sitzung festhält. **Fehlt er, ist sie angemeldet** — so, wie jeder
// Bestand von vor ID-02 beginnt.
const Sitzungsschluessel = "sitzung"

// SitzungLesen gibt den festgehaltenen Zustand.
func (a *Nutzerablage) SitzungLesen(ctx context.Context) (domain.Sitzungszustand, error) {
	var wert string
	if err := a.db.QueryRowContext(ctx,
		`SELECT coalesce((SELECT wert FROM bestand WHERE schluessel = ?), ?)`,
		Sitzungsschluessel, string(domain.SitzungAngemeldet)).Scan(&wert); err != nil {
		return "", uebersetze(err, "die Sitzung lesen")
	}

	return domain.Sitzungszustand(wert), nil
}

// SitzungFesthalten hält einen Zustand fest — angemeldet nimmt den Eintrag
// weg.
func (a *Nutzerablage) SitzungFesthalten(ctx context.Context, zustand domain.Sitzungszustand) error {
	return sitzungFesthaltenIn(ctx, a.db, zustand)
}

// AnmeldungFesthalten stellt beide Zeiger auf den, der sich anmeldet, und
// hält die Sitzung als angemeldet fest — **in einem Geschäft** (Nachlauf über
// ID-02, Befund 2). Vorher standen die Zeiger schon, wenn das Festhalten
// scheiterte: ein Mensch im Zeiger, der nie angemeldet war.
func (a *Nutzerablage) AnmeldungFesthalten(ctx context.Context, nutzerID, profilID int64) error {
	return imGeschaeft(ctx, a.db, "die Anmeldung festhalten", func(geschaeft *sql.Tx) error {
		if fehler := zeigerSetzenIn(ctx, geschaeft, nutzerID, profilID); fehler != nil {
			return fehler
		}

		return sitzungFesthaltenIn(ctx, geschaeft, domain.SitzungAngemeldet)
	})
}

// sitzungFesthaltenIn schreibt den Zustand — in der Datenbank oder in einem
// Geschäft.
func sitzungFesthaltenIn(ctx context.Context, ausfuehren ausfuehrer, zustand domain.Sitzungszustand) error {
	var err error
	if zustand == domain.SitzungAngemeldet {
		_, err = ausfuehren.ExecContext(ctx, `DELETE FROM bestand WHERE schluessel = ?`, Sitzungsschluessel)
	} else {
		_, err = ausfuehren.ExecContext(ctx,
			`INSERT INTO bestand (schluessel, wert) VALUES (?, ?)
			 ON CONFLICT (schluessel) DO UPDATE SET wert = excluded.wert`,
			Sitzungsschluessel, string(zustand))
	}

	return uebersetze(err, "die Sitzung festhalten")
}
