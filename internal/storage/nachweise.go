// **Nachweise im Bestand** — Werkstück O1 von Phase 5.2
// (#1425),
// die Ablage zu [service.Nachweisablage].
//
// # Das Einlösen ist ein `UPDATE`, kein Lesen und Schreiben
//
// `UPDATE … WHERE kennung = ? AND eingeloest = 0` trifft **genau einmal**,
// gleich wie viele zugleich kommen — SQLite reiht Schreiber; wer als Zweiter
// kommt, ändert null Zeilen. Null Zeilen heißt dann eines von zweien: den
// Nachweis gibt es nicht, oder er ist verbraucht. Ein zweiter Blick sagt,
// welches — und das ist die Auskunft, die der Aufrufer braucht (§8): „gibt es
// nicht" und „jemand war vor dir" sind verschiedene nächste Schritte.

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Nachweisablage schreibt und liest Nachweise.
type Nachweisablage struct{ db *sql.DB }

// NeueNachweisablage baut sie über der Datenbank.
func NeueNachweisablage(db *sql.DB) *Nachweisablage { return &Nachweisablage{db: db} }

// NachweisAnlegen legt einen Nachweis ab — dieselbe Kennung zweimal ist ein
// Konflikt.
func (a *Nachweisablage) NachweisAnlegen(ctx context.Context, n domain.Nachweis) error {
	_, fehler := a.db.ExecContext(ctx, `
		INSERT INTO freigabenachweise (
			kennung, nutzer_id, profil_id, zweck, ziel, revision,
			grenze_betrag, grenze_waehrung, ausgestellt, ablauf, eingeloest
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		n.Kennung, n.NutzerID, n.ProfilID, n.Zweck, n.Ziel, n.Revision,
		n.Grenze.Betrag, n.Grenze.Waehrung,
		n.Ausgestellt.UTC().Format(time.RFC3339Nano),
		n.Ablauf.UTC().Format(time.RFC3339Nano),
		boolAlsZahl(n.Eingeloest),
	)

	return uebersetze(fehler, "Nachweis %q anlegen", n.Kennung)
}

// Nachweis liest einen Nachweis über seine Kennung.
func (a *Nachweisablage) Nachweis(ctx context.Context, kennung string) (domain.Nachweis, error) {
	var (
		n                   domain.Nachweis
		ausgestellt, ablauf string
		eingeloest          int
	)

	fehler := a.db.QueryRowContext(ctx, `
		SELECT kennung, nutzer_id, profil_id, zweck, ziel, revision,
		       grenze_betrag, grenze_waehrung, ausgestellt, ablauf, eingeloest
		FROM freigabenachweise WHERE kennung = ?`, kennung).Scan(
		&n.Kennung, &n.NutzerID, &n.ProfilID, &n.Zweck, &n.Ziel, &n.Revision,
		&n.Grenze.Betrag, &n.Grenze.Waehrung, &ausgestellt, &ablauf, &eingeloest,
	)
	if errors.Is(fehler, sql.ErrNoRows) {
		return domain.Nachweis{}, fmt.Errorf(
			"%w: den Nachweis %q führt der Bestand nicht", domain.ErrNotFound, kennung)
	}

	if fehler != nil {
		return domain.Nachweis{}, uebersetze(fehler, "Nachweis %q lesen", kennung)
	}

	if n.Ausgestellt, fehler = time.Parse(time.RFC3339Nano, ausgestellt); fehler != nil {
		return domain.Nachweis{}, fmt.Errorf("Ausstellzeit von %q lesen: %w", kennung, fehler)
	}

	if n.Ablauf, fehler = time.Parse(time.RFC3339Nano, ablauf); fehler != nil {
		return domain.Nachweis{}, fmt.Errorf("Ablauf von %q lesen: %w", kennung, fehler)
	}

	n.Eingeloest = eingeloest == 1

	return n, nil
}

// NachweisEinloesen verbraucht einen Nachweis — genau einmal.
func (a *Nachweisablage) NachweisEinloesen(ctx context.Context, kennung string) error {
	ergebnis, fehler := a.db.ExecContext(ctx,
		`UPDATE freigabenachweise SET eingeloest = 1 WHERE kennung = ? AND eingeloest = 0`,
		kennung)
	if fehler != nil {
		return uebersetze(fehler, "Nachweis %q einlösen", kennung)
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return zaehlfehler
	}

	if zeilen == 1 {
		return nil
	}

	// **Null Zeilen — und jetzt der zweite Blick.**
	if _, fehler := a.Nachweis(ctx, kennung); fehler != nil {
		return fehler
	}

	return fmt.Errorf("%w: der Nachweis %q ist schon eingelöst", domain.ErrConflict, kennung)
}

// boolAlsZahl schreibt einen Wahrheitswert, wie SQLite ihn hält.
func boolAlsZahl(wert bool) int {
	if wert {
		return 1
	}

	return 0
}
