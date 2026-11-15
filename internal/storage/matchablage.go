package storage

// **Bindung und Meldung im Bestand** — Werkstück G7 von Phase 5.5
// (#1464).
//
// **Zwei Tabellen, zwei Fragen** — der Grund steht in der Wanderung 0067.
//
// # Warum hier kein Profil geklemmt wird
//
// **Weil die Paarung es schon tut.** `matchbindungen` hängt über
// `turnierpaarungen` an `turnierrunden`, an `turniere` und damit am Profil;
// wer an eine fremde Paarung käme, käme an ein fremdes Turnier — und dagegen
// steht die Klemme dort. Eine zweite hier wäre eine zweite Antwort auf
// dieselbe Frage (§10).

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Matchablage hält Bindungen und Meldungen.
type Matchablage struct{ db *sql.DB }

// NeueMatchablage baut sie.
func NeueMatchablage(db *sql.DB) *Matchablage { return &Matchablage{db: db} }

// BindungSchreiben legt eine Bindung an oder ersetzt sie.
//
// **Ersetzen und nicht ablehnen.** Wer eine Paarung erneut bindet, hat die
// erste Partie abgebrochen und fängt neu an — und eine Bindung, die im Weg
// stünde, zwänge zu einem Löschweg, den niemand braucht. **Die Meldungen
// gehen dabei mit**: sie gehören zur alten Partie.
func (a *Matchablage) BindungSchreiben(
	ctx context.Context, bindung domain.Matchbindung,
) error {
	// **Ohne Fehlerweg**, und der Grund steht bei [kodiereLeihe]: ein
	// [domain.Matchsitz] trägt Zeichenketten und eine Zahl; `encoding/json`
	// kann daran nicht scheitern. Ein Zweig dafür stünde unbetreten da und
	// sähe aus wie eine geprüfte Zusage (§10).
	sitze, _ := json.Marshal(bindung.Sitze) //nolint:errcheck,errchkjson // siehe oben

	geschaeft, fehler := a.db.BeginTx(ctx, nil)
	if fehler != nil {
		return uebersetze(fehler, "die Bindung schreiben")
	}

	defer func() { _ = geschaeft.Rollback() }() //nolint:errcheck // nach Commit wirkungslos

	// **Paarung und Partie aus dem gezeigten Profil, und die Paarung aus
	// diesem Turnier** (#1815). `partie_id` verweist im Bestand auf nichts
	// (0067); hier wird zugesagt, was der Fremdschlüssel nicht hält.
	if fehler := paarungImTurnier(ctx, geschaeft, bindung.TurnierID, bindung.PaarungID); fehler != nil {
		return fehler
	}

	if fehler := partieImProfil(ctx, geschaeft, bindung.PartieID); fehler != nil {
		return fehler
	}

	if _, fehler := geschaeft.ExecContext(ctx,
		`DELETE FROM matchbindungen WHERE paarung_id = ?`,
		bindung.PaarungID); fehler != nil {
		return uebersetze(fehler, "die bisherige Bindung entfernen")
	}

	if _, fehler := geschaeft.ExecContext(ctx,
		`INSERT INTO matchbindungen
		   (paarung_id, turnier_id, fassung, partie_id, sitze, hoechstens)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		bindung.PaarungID, bindung.TurnierID, bindung.Fassung,
		bindung.PartieID, string(sitze), bindung.Hoechstens); fehler != nil {
		return uebersetze(fehler, "die Bindung schreiben")
	}

	if fehler := geschaeft.Commit(); fehler != nil {
		return uebersetze(fehler, "die Bindung schreiben")
	}

	return nil
}

// Bindung gibt die Bindung einer Paarung.
func (a *Matchablage) Bindung(
	ctx context.Context, paarungID int64,
) (domain.Matchbindung, error) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := paarungImProfil(ctx, a.db, paarungID); fehler != nil {
		return domain.Matchbindung{}, fehler
	}

	bindung := domain.Matchbindung{PaarungID: paarungID}

	var sitze string

	fehler := a.db.QueryRowContext(ctx,
		`SELECT turnier_id, fassung, partie_id, sitze, hoechstens
		   FROM matchbindungen WHERE paarung_id = ?`, paarungID).Scan(
		&bindung.TurnierID, &bindung.Fassung, &bindung.PartieID,
		&sitze, &bindung.Hoechstens)

	if errors.Is(fehler, sql.ErrNoRows) {
		return domain.Matchbindung{}, fmt.Errorf(
			"%w: an der Paarung %d hängt keine Partie",
			domain.ErrNotFound, paarungID)
	}

	if fehler != nil {
		return domain.Matchbindung{}, uebersetze(fehler, "die Bindung lesen")
	}

	if fehler := json.Unmarshal([]byte(sitze), &bindung.Sitze); fehler != nil {
		return domain.Matchbindung{}, domain.MitGrund(domain.ErrValidation,
			fehler, "die gespeicherten Sitze ließen sich nicht lesen")
	}

	return bindung, nil
}

// MeldungSchreiben trägt die Meldung einer Seite ein — **eine je Sitz**.
//
// **Die letzte gilt** (`ON CONFLICT … DO UPDATE`), und der Grund steht bei
// [domain.Matchbindung.Uebernahme]: wer zweimal meldet, hat sich beim ersten
// Mal vertippt.
func (a *Matchablage) MeldungSchreiben(
	ctx context.Context, meldung domain.Matchmeldung,
) error {
	return imGeschaeft(ctx, a.db, "die Meldung schreiben", func(tx *sql.Tx) error {
		// **Nur an einer Paarung des gezeigten Profils** (#1815).
		if fehler := paarungImProfil(ctx, tx, meldung.PaarungID); fehler != nil {
			return fehler
		}

		return meldungEintragen(ctx, tx, meldung)
	})
}

// meldungEintragen schreibt eine Meldung oder die bisherige derselben Seite
// fort.
func meldungEintragen(ctx context.Context, tx *sql.Tx, meldung domain.Matchmeldung) error {
	if _, fehler := tx.ExecContext(ctx,
		`INSERT INTO matchmeldungen
		   (paarung_id, von, fassung, partie_id, spiele_a, spiele_b,
		    unentschieden, kette)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (paarung_id, von) DO UPDATE SET
		   fassung       = excluded.fassung,
		   partie_id     = excluded.partie_id,
		   spiele_a      = excluded.spiele_a,
		   spiele_b      = excluded.spiele_b,
		   unentschieden = excluded.unentschieden,
		   kette         = excluded.kette,
		   gemeldet      = datetime('now')`,
		meldung.PaarungID, meldung.Von, meldung.Fassung, meldung.PartieID,
		meldung.Ergebnis.SpieleA, meldung.Ergebnis.SpieleB,
		meldung.Ergebnis.Unentschieden, meldung.Kette); fehler != nil {
		return uebersetze(fehler, "die Meldung schreiben")
	}

	return nil
}

// Meldungen gibt, was die Seiten gemeldet haben — nach Sitz geordnet.
func (a *Matchablage) Meldungen(
	ctx context.Context, paarungID int64,
) ([]domain.Matchmeldung, error) {
	if fehler := paarungImProfil(ctx, a.db, paarungID); fehler != nil {
		return nil, fehler
	}

	zeilen, fehler := a.db.QueryContext(ctx,
		`SELECT von, fassung, partie_id, spiele_a, spiele_b, unentschieden,
		        kette
		   FROM matchmeldungen WHERE paarung_id = ? ORDER BY von`, paarungID)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Meldungen lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	gesammelt := []domain.Matchmeldung{}

	for zeilen.Next() {
		meldung := domain.Matchmeldung{PaarungID: paarungID}

		if lesefehler := zeilen.Scan(&meldung.Von, &meldung.Fassung,
			&meldung.PartieID, &meldung.Ergebnis.SpieleA,
			&meldung.Ergebnis.SpieleB, &meldung.Ergebnis.Unentschieden,
			&meldung.Kette); lesefehler != nil {
			return nil, uebersetze(lesefehler, "eine Meldung lesen")
		}

		gesammelt = append(gesammelt, meldung)
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, uebersetze(fehler, "die Meldungen lesen")
	}

	return gesammelt, nil
}
