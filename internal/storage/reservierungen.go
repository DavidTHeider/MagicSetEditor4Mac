package storage

// **Reservierungen im Bestand** — Werkstück `T2` von Phase 5.6
// (#1467).
//
// # Warum Nachsehen und Eintragen in einer Transaktion stehen
//
// **Weil dazwischen ein Fenster liegt.** Zwei Rufe, die beide „es sind noch
// zwei frei" lesen und beide zwei eintragen, hätten vier festgehalten — und
// die Domäne hätte in beiden Fällen Ja gesagt. Sie prüft, **was** hineinpasst;
// dass es zwischen der Prüfung und dem Eintrag niemand ändert, kann nur der
// Bestand.
//
// **`_txlock=immediate` steht schon am Treiber** (`datenbank.go`): eine
// Transaktion nimmt die Schreibsperre **beim Beginnen** und nicht beim ersten
// Schreiben. Genau das braucht dieser Ablauf — sonst läsen zwei Transaktionen
// gemeinsam und schrieben nacheinander.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Reservierungsablage hält die Zusagen eines Profils.
type Reservierungsablage struct{ db *sql.DB }

// NeueReservierungsablage baut sie.
func NeueReservierungsablage(db *sql.DB) *Reservierungsablage {
	return &Reservierungsablage{db: db}
}

// Festhalten trägt eine Zusage ein — oder sagt, warum nicht.
//
// **Ohne Bestandswirkung im Fehlerfall**: die Transaktion wird
// zurückgenommen, und zwar bevor irgendetwas steht.
func (a *Reservierungsablage) Festhalten(
	ctx context.Context, zusage domain.Reservierung, vertrag domain.Zusagevertrag,
) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("die Zusage ließ sich nicht beginnen: %w", err)
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach dem Commit wirkungslos

	exemplar, fehler := besitzInTx(ctx, tx, profilID, zusage.ExemplarID)
	if fehler != nil {
		return fehler
	}

	aufDerZeile, fehler := zusagenInTx(ctx, tx, profilID, zusage.ExemplarID)
	if fehler != nil {
		return fehler
	}

	if fehler := domain.ReservierungPasst(
		exemplar, aufDerZeile, zusage, time.Now()); fehler != nil {
		return fehler
	}

	// **Und der Vertrag — in dieser Transaktion** (Befund 1 des Nachlaufs
	// zum Zweitblick über Phase 5.6, seit #2492 je Variante). Die Zeile
	// darüber fragt die **Besitzzeile**; wer nur sie fragte, ließ zwei Rufe
	// für **zwei verschiedene** Zeilen beide durch, und der Vertrag über ein
	// Stück bekam zwei Abgänge. Eine Prüfung davor im Dienst ist eine
	// Vorschau, keine Grenze.
	if fehler := vertragPruefenInTx(ctx, tx, profilID, exemplar, zusage, vertrag); fehler != nil {
		return fehler
	}

	if _, fehler := tx.ExecContext(ctx,
		`INSERT INTO reservierungen
		   (profil_id, vorgang, revision, exemplar_id, menge, fassung, frist,
		    uebergeben)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		profilID, zusage.Vorgang, zusage.Revision, zusage.ExemplarID,
		zusage.Menge, zusage.Fassung, alsFrist(zusage.Frist),
		alsZahl(zusage.Uebergeben),
	); fehler != nil {
		return fmt.Errorf("die Zusage ließ sich nicht schreiben: %w", fehler)
	}

	if fehler := tx.Commit(); fehler != nil {
		return fmt.Errorf("die Zusage ließ sich nicht abschließen: %w", fehler)
	}

	return nil
}

// Reservierungen gibt die Zusagen auf ein Exemplar.
//
// **Alle und nicht nur die aktiven**: ob eine gilt, entscheidet
// [domain.Reservierung.Aktiv] gegen eine Uhr — und die gehört dem Aufrufer,
// nicht dem Bestand. Wer hier filterte, hätte die Uhr zweimal.
func (a *Reservierungsablage) Reservierungen(
	ctx context.Context, exemplarID int64,
) ([]domain.Reservierung, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	return zusagenInTx(ctx, a.db, profilID, exemplarID)
}

// ReserviertJeExemplar zählt, was je Besitzzeile **jetzt** festgehalten ist.
//
// # Warum das gezählt wird und nicht summiert
//
// **Weil eine abgelaufene Frist nichts festhält.** `SUM(menge)` über alle
// Zeilen wäre kürzer und falsch: eine Zusage mit einer Frist von gestern
// bindet nichts mehr, und wer sie mitrechnete, zeigte eine Sammlung als
// gebunden, die frei ist.
//
// **Die Frist wird dabei in Go entschieden und nicht in SQL.** Der Satz, was
// „aktiv" heißt, steht in [domain.Reservierung.Aktiv] — einmal, und für jeden
// Leser derselbe. Ein `WHERE frist = ” OR frist > ?` daneben wäre eine
// zweite Fassung derselben Regel, und die läuft beim ersten Sonderfall
// auseinander (hier: die bestätigte Übergabe, die **trotz** Frist hält).
func (a *Reservierungsablage) ReserviertJeExemplar(
	ctx context.Context,
) (map[int64]int, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, fehler := a.db.QueryContext(ctx,
		`SELECT exemplar_id, menge, frist, uebergeben
		   FROM reservierungen
		  WHERE profil_id = ?`, profilID)
	if fehler != nil {
		return nil, fmt.Errorf("die Zusagen ließen sich nicht lesen: %w", fehler)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird nur

	jetzt := time.Now()
	gezaehlt := map[int64]int{}

	for zeilen.Next() {
		var (
			exemplarID int64
			zusage     domain.Reservierung
			frist      string
			uebergeben int
		)

		if fehler := zeilen.Scan(
			&exemplarID, &zusage.Menge, &frist, &uebergeben); fehler != nil {
			return nil, fmt.Errorf("eine Zusage ließ sich nicht lesen: %w", fehler)
		}

		gelesen, fristfehler := ausFrist(frist)
		if fristfehler != nil {
			return nil, fristfehler
		}

		zusage.Frist, zusage.Uebergeben = gelesen, uebergeben == 1

		if zusage.Aktiv(jetzt) {
			gezaehlt[exemplarID] += zusage.Menge
		}
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, fmt.Errorf("die Zusagen ließen sich nicht lesen: %w", fehler)
	}

	return gezaehlt, nil
}

// Freigeben nimmt die Zusagen eines Vorgangs weg.
//
// **Eine bestätigte Übergabe gibt nicht frei** — auch nicht hier. Sonst
// genügte ein verspäteter Aufräumer, um Karten wieder verfügbar zu machen,
// die längst über den Tisch gegangen sind.
func (a *Reservierungsablage) Freigeben(
	ctx context.Context, vorgang string,
) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	return freigebenInTx(ctx, a.db, profilID, vorgang)
}

// freigebenInTx tut dasselbe über einen beliebigen Ableser.
//
// **Damit dieselbe Regel auch in der Transaktion einer Zustellung gilt**
// (Befund 3 des Nachlaufs zum Zweitblick über Phase 5.6) — eine zweite
// Abschrift liefe beim ersten neuen Grund auseinander.
func freigebenInTx(
	ctx context.Context, wer bestandsfrager, profilID int64, vorgang string,
) error {
	var uebergeben int

	if fehler := wer.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM reservierungen
		  WHERE profil_id = ? AND vorgang = ? AND uebergeben = 1`,
		profilID, vorgang,
	).Scan(&uebergeben); fehler != nil {
		return fmt.Errorf("die Zusagen ließen sich nicht zählen: %w", fehler)
	}

	if uebergeben > 0 {
		return domain.MitGrund(domain.ErrConflict, nil,
			"eine bestätigte Übergabe gibt nichts mehr frei — das klärt ein "+
				"Mensch")
	}

	if _, fehler := wer.ExecContext(ctx,
		`DELETE FROM reservierungen WHERE profil_id = ? AND vorgang = ?`,
		profilID, vorgang,
	); fehler != nil {
		return fmt.Errorf("die Zusagen ließen sich nicht lösen: %w", fehler)
	}

	return nil
}

// zeilenleser ist, was `*sql.DB` und `*sql.Tx` gemeinsam haben.
//
// **Damit derselbe Ableser in und außerhalb einer Transaktion gilt.** Zwei
// Fassungen liefen bei der ersten Änderung auseinander, und die eine, die
// niemand ändert, ist die in der Transaktion.
//
// **Weder `frager` noch `zeilenleser` noch `abfrager`**: alle drei Namen sind
// in diesem Paket vergeben, und jeder von ihnen kann etwas anderes. Ein
// vierter, der wie ein dritter heißt, wäre die Verwechslung, die man später
// nicht mehr findet.
type bestandsfrager interface {
	QueryContext(ctx context.Context, abfrage string, werte ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, abfrage string, werte ...any) *sql.Row

	// ExecContext gehört dazu, seit dasselbe Freigeben auch in der
	// Transaktion einer Zustellung gilt (Befund 3 des Nachlaufs zum
	// Zweitblick über Phase 5.6).
	ExecContext(ctx context.Context, abfrage string, werte ...any) (sql.Result, error)
}

// besitzInTx liest die Besitzzeile, gegen die geprüft wird.
func besitzInTx(
	ctx context.Context, wer bestandsfrager, profilID, exemplarID int64,
) (domain.Exemplar, error) {
	exemplar := domain.Exemplar{ID: exemplarID}

	var wunsch int

	// **Mit der Variante** (#2492): ob das Exemplar zum Vertrag passt, sagt
	// erst sie.
	fehler := wer.QueryRowContext(ctx,
		`SELECT menge, fassung, wunsch, printing, coalesce(druck_id, 0),
		        zustand, sprache, finish FROM sammlung
		  WHERE id = ? AND profil_id = ?`,
		exemplarID, profilID,
	).Scan(&exemplar.Menge, &exemplar.Fassung, &wunsch, &exemplar.Printing,
		&exemplar.DruckID, &exemplar.Zustand, &exemplar.Sprache, &exemplar.Finish)

	if errors.Is(fehler, sql.ErrNoRows) {
		return domain.Exemplar{}, domain.MitGrund(domain.ErrNotFound, nil,
			"die Besitzzeile gibt es nicht")
	}

	if fehler != nil {
		return domain.Exemplar{}, fmt.Errorf(
			"die Besitzzeile ließ sich nicht lesen: %w", fehler)
	}

	exemplar.Wunsch = wunsch == 1

	return exemplar, nil
}

// zusagenInTx liest die Zusagen auf ein Exemplar.
func zusagenInTx(
	ctx context.Context, wer bestandsfrager, profilID, exemplarID int64,
) ([]domain.Reservierung, error) {
	zeilen, fehler := wer.QueryContext(ctx,
		zusagenabfrage+` WHERE profil_id = ? AND exemplar_id = ? ORDER BY id ASC`,
		profilID, exemplarID)
	if fehler != nil {
		return nil, fmt.Errorf("die Zusagen ließen sich nicht lesen: %w", fehler)
	}

	return zusagenLesen(zeilen)
}

// alsFrist schreibt einen Zeitpunkt — der Nullwert wird zur leeren Angabe.
func alsFrist(wann time.Time) string {
	if wann.IsZero() {
		return ""
	}

	return wann.UTC().Format(time.RFC3339Nano)
}

// ausFrist liest einen Zeitpunkt zurück.
//
// **Was keiner ist, ist ein Befund** (§8): ein von Hand geänderter Bestand
// darf die Anwendung nicht umwerfen, und raten darf sie erst recht nicht.
func ausFrist(roh string) (time.Time, error) {
	if roh == "" {
		return time.Time{}, nil
	}

	wann, fehler := time.Parse(time.RFC3339Nano, roh)
	if fehler != nil {
		return time.Time{}, domain.MitGrund(domain.ErrValidation, fehler,
			"die gespeicherte Frist ist kein Zeitpunkt")
	}

	return wann, nil
}

// alsZahl macht aus einem Wahrheitswert die 0 oder die 1.
//
// **SQLite kennt keinen Wahrheitswert**, und `STRICT` weist einen erfundenen
// Typ ab — derselbe Satz wie an der Sammlung selbst.
func alsZahl(ja bool) int {
	if ja {
		return 1
	}

	return 0
}
