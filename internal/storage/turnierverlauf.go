package storage

// **Runden, Paarungen und Ergebnisse** — Werkstück 3 von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, damit `turniere.go` unter der Zeilengrenze aus §2
// bleibt. Der Schnitt folgt der Frage: dort steht, **was ein Turnier ist**,
// hier, **wie sein Verlauf entsteht**.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// RundeLegen legt eine Runde samt allen ihren Paarungen an — **in einem Zug**.
//
// **Das ist die Zusage aus dem Issue**: „Ein Absturz mitten in der
// Ergebniserfassung hinterlässt keine halbe Runde." Sie löst kein
// Zustandsfeld ein, sondern diese Transaktionsgrenze (Vermerk 0062,
// Entscheidung 10).
//
// **Die Nummer kommt vom Aufrufer und wird geprüft.**
//
// Hier stand zuerst „die Nummer kommt vom Speicher", und das war die Lücke:
// zwei Rufe konnten denselben Stand lesen, denselben Plan rechnen und ihn als
// zwei aufeinanderfolgende Runden ablegen (Befund 8 des Zweitblicks über
// Phase 4.7). Wer eine Runde legt, sagt jetzt, **die wievielte er meint** —
// und bekommt einen Konflikt, wenn inzwischen jemand anderes eine gelegt hat.
//
// **Dieselbe Gestalt wie eine Fassung** (§7, Regel 2): der Aufrufer nennt den
// Stand, aus dem er rechnet, und die Transaktion entscheidet.
func (a *Turnierablage) RundeLegen(
	ctx context.Context, turnierID int64, erwartet int, paarungen []domain.Paarung,
) (domain.Runde, error) {
	if len(paarungen) == 0 {
		return domain.Runde{}, fmt.Errorf(
			"%w: eine Runde ohne eine einzige Paarung ist keine",
			domain.ErrValidation)
	}

	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Runde{}, profilfehler
	}

	zug, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Runde{}, uebersetze(err, "die Runde beginnen")
	}
	defer zug.Rollback() //nolint:errcheck // nach einem Commit ist es ein No-op

	// **Nur in einem Turnier des gezeigten Profils, und nur mit seinen
	// Teilnehmern** (#1815) — gefragt vor der Rundennummer, sonst verriete die
	// Konfliktmeldung, bei welcher Runde ein fremdes Turnier steht.
	if fehler := rundeGehoertDemTurnier(ctx, zug, turnierID, paarungen); fehler != nil {
		return domain.Runde{}, fehler
	}

	runde, err := legeRundeImZug(ctx, zug, turnierID, erwartet, paarungen)
	if err != nil {
		return domain.Runde{}, err
	}

	// **Eine gelegte Runde ist eine Änderung am Turnier** — die Fassung zählt
	// sie, zusätzlich zur Rundennummer oben. Die Nummer schützt genau diesen
	// Weg; die Fassung sagt jedem Leser, dass sich etwas getan hat.
	if err := fortschreiben(ctx, zug, turnierID, profilID, domain.Protokolleintrag{
		Vorgang: domain.TurnierRundeGelegt, Angaben: strconv.Itoa(runde.Nummer),
	}); err != nil {
		return domain.Runde{}, err
	}

	if err := zug.Commit(); err != nil {
		return domain.Runde{}, uebersetze(err, "die Runde abschließen")
	}

	return runde, nil
}

// legeRundeImZug ist der Rumpf von [Turnierablage.RundeLegen].
// rundeGehoertDemTurnier weist ein fremdes Turnier ab und jede Paarung, deren
// Seiten nicht in ihm stehen. **Ein Bye hat keine zweite Seite.**
func rundeGehoertDemTurnier(
	ctx context.Context, zug *sql.Tx, turnierID int64, paarungen []domain.Paarung,
) error {
	if fehler := turnierImProfil(ctx, zug, turnierID); fehler != nil {
		return fehler
	}

	for _, paarung := range paarungen {
		for _, seite := range []int64{paarung.A, paarung.B} {
			if seite == 0 {
				continue
			}

			if fehler := teilnehmerImTurnier(ctx, zug, turnierID, seite); fehler != nil {
				return fehler
			}
		}
	}

	return nil
}

func legeRundeImZug(
	ctx context.Context, zug *sql.Tx, turnierID int64, erwartet int,
	paarungen []domain.Paarung,
) (domain.Runde, error) {
	var nummer int

	if err := zug.QueryRowContext(ctx,
		`SELECT coalesce(max(nummer), 0) + 1 FROM turnierrunden
		  WHERE turnier_id = ?`, turnierID).Scan(&nummer); err != nil {
		return domain.Runde{}, uebersetze(err, "die nächste Rundennummer lesen")
	}

	// **Der Vergleich ist der ganze Punkt.** Er steht **in** der Transaktion:
	// draußen wäre er eine Frage an einen Stand, der sich bis zum Schreiben
	// noch ändern kann.
	if erwartet != nummer {
		return domain.Runde{}, fmt.Errorf(
			"%w: dieses Turnier ist bei Runde %d, gerechnet wurde für Runde %d "+
				"— jemand anderes war schneller",
			domain.ErrConflict, nummer, erwartet)
	}

	ergebnis, err := zug.ExecContext(ctx,
		`INSERT INTO turnierrunden (turnier_id, nummer) VALUES (?, ?)`,
		turnierID, nummer)
	if err != nil {
		return domain.Runde{}, uebersetze(err, "die Runde anlegen")
	}

	rundeID, idfehler := ergebnis.LastInsertId()
	if idfehler != nil {
		return domain.Runde{}, uebersetze(idfehler, "die Kennung der Runde lesen")
	}

	runde := domain.Runde{
		ID: rundeID, TurnierID: turnierID, Nummer: nummer,
		Paarungen: make([]domain.Paarung, 0, len(paarungen)),
	}

	for stelle, paarung := range paarungen {
		gelegt, err := legePaarungImZug(ctx, zug, rundeID, paarung, stelle+1)
		if err != nil {
			return domain.Runde{}, err
		}

		runde.Paarungen = append(runde.Paarungen, gelegt)
	}

	return runde, nil
}

// legePaarungImZug legt eine Paarung.
func legePaarungImZug(
	ctx context.Context, zug *sql.Tx, rundeID int64,
	paarung domain.Paarung, stelle int,
) (domain.Paarung, error) {
	ergebnis, err := zug.ExecContext(ctx,
		`INSERT INTO turnierpaarungen (runde_id, a, b, stelle) VALUES (?, ?, ?, ?)`,
		rundeID, paarung.A, kennungOderNull(paarung.B), stelle)
	if err != nil {
		return domain.Paarung{}, uebersetze(err, "eine Paarung anlegen")
	}

	kennung, idfehler := ergebnis.LastInsertId()

	paarung.ID = kennung
	paarung.RundeID = rundeID
	paarung.Stelle = stelle

	return paarung, uebersetze(idfehler, "die Kennung der Paarung lesen")
}

// ErgebnisSetzen trägt ein Ergebnis ein — oder nimmt es zurück.
//
// **`nil` heißt: doch nicht gespielt.** Berichtigen ist am Turniertisch der
// Normalfall; ein Weg, der nur eintragen könnte, zwänge dazu, die Paarung neu
// zu legen — und damit die halbe Runde herzustellen, gegen die es die
// Transaktion gibt.
// **Und es geht gegen die Fassung** (§7, Regel 2). Zwei Leute am selben Tisch,
// beide haben die Runde offen: wer als Zweiter einträgt, überschriebe sonst
// die Berichtigung des Ersten, ohne dass jemand es merkte (Befund 7 des
// Nachlaufs über Phase 4.7).
func (a *Turnierablage) ErgebnisSetzen(
	ctx context.Context, fassung, paarungID int64, stand *domain.Ergebnis,
) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	zug, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "das Eintragen beginnen")
	}
	defer zug.Rollback() //nolint:errcheck // nach einem Commit ist es ein No-op

	// **Der Besitz vor der Fassung** (#1815). Die Fassung wiese ein fremdes
	// Turnier auch ab — aber als Nebenwirkung, und ihre Antwort spräche dann
	// über die Fassung statt über den Besitz.
	if fehler := paarungImProfil(ctx, zug, paarungID); fehler != nil {
		return fehler
	}

	// **Die Fassung zuerst.** Stimmt sie nicht, bleibt die Paarung unberührt
	// — und die Meldung sagt, was zu tun ist, statt still zu überschreiben.
	turnierID, lesefehler := turnierZurPaarung(ctx, zug, paarungID)
	if lesefehler != nil {
		return lesefehler
	}

	if _, fortschrittfehler := fortschreibenGegen(
		ctx, zug, turnierID, fassung, profilID, ergebniseintrag(paarungID, stand)); fortschrittfehler != nil {
		return fortschrittfehler
	}

	if _, schreibfehler := zug.ExecContext(ctx,
		`UPDATE turnierpaarungen
		    SET spiele_a = ?, spiele_b = ?, unentschieden = ?
		  WHERE id = ?`,
		zahlOderNull(stand, func(e domain.Ergebnis) int { return e.SpieleA }),
		zahlOderNull(stand, func(e domain.Ergebnis) int { return e.SpieleB }),
		zahlOderNull(stand, func(e domain.Ergebnis) int { return e.Unentschieden }),
		paarungID); schreibfehler != nil {
		return uebersetze(schreibfehler, "das Ergebnis setzen")
	}

	return uebersetze(zug.Commit(), "das Eintragen abschließen")
}

// turnierZurPaarung sagt, zu welchem Turnier eine Paarung gehört.
//
// **Sie ist zugleich die Prüfung, dass es die Paarung gibt.** Ein zweiter
// Weg dafür wäre eine zweite Fassung derselben Frage (§10).
func turnierZurPaarung(ctx context.Context, wo execer, paarungID int64) (
	int64, error,
) {
	var turnierID int64

	err := wo.QueryRowContext(ctx,
		`SELECT r.turnier_id
		   FROM turnierpaarungen p
		   JOIN turnierrunden r ON r.id = p.runde_id
		  WHERE p.id = ?`, paarungID).Scan(&turnierID)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf(
			"%w: eine Paarung mit dieser Kennung gibt es nicht", domain.ErrNotFound)
	}

	return turnierID, uebersetze(err, "das Turnier der Paarung lesen")
}

// zahlOderNull macht aus einem fehlenden Ergebnis ein SQL-NULL.
func zahlOderNull(stand *domain.Ergebnis, welche func(domain.Ergebnis) int) any {
	if stand == nil {
		return nil
	}

	return welche(*stand)
}

// runden liest den Verlauf eines Turniers.
func (a *Turnierablage) runden(ctx context.Context, turnierID int64) (
	[]domain.Runde, error,
) {
	zeilen, err := a.db.QueryContext(ctx,
		`SELECT r.id, r.nummer, p.id, p.a, coalesce(p.b, 0),
		        p.spiele_a, p.spiele_b, p.unentschieden, p.stelle
		   FROM turnierrunden r
		   LEFT JOIN turnierpaarungen p ON p.runde_id = r.id
		  WHERE r.turnier_id = ?
		  ORDER BY r.nummer, p.stelle`, turnierID)
	if err != nil {
		return nil, uebersetze(err, "den Verlauf lesen")
	}
	defer zeilen.Close() //nolint:errcheck // gelesen wird der Fehler von Err()

	var (
		runden []domain.Runde
		nach   = map[int64]int{}
	)

	for zeilen.Next() {
		runde, paarung, hatPaarung, err := liesVerlaufszeile(zeilen, turnierID)
		if err != nil {
			return nil, err
		}

		stelle, schon := nach[runde.ID]
		if !schon {
			stelle = len(runden)
			nach[runde.ID] = stelle
			runden = append(runden, runde)
		}

		if hatPaarung {
			runden[stelle].Paarungen = append(runden[stelle].Paarungen, paarung)
		}
	}

	return runden, uebersetze(zeilen.Err(), "den Verlauf lesen")
}

// liesVerlaufszeile deutet eine Zeile des Verbunds aus Runde und Paarung.
//
// **Der `LEFT JOIN` gibt eine Runde ohne Paarung als Zeile mit NULLen.** Sie
// entsteht nicht über [Turnierablage.RundeLegen] — das weist eine leere Runde
// ab —, aber ein von Hand berichtigter Bestand kann sie haben, und dann ist
// eine Runde ohne Paarungen die richtige Antwort und kein Absturz.
func liesVerlaufszeile(zeilen *sql.Rows, turnierID int64) (
	runde domain.Runde, paarung domain.Paarung, hatPaarung bool, fehler error,
) {
	var (
		paarungID                    sql.NullInt64
		a, b                         sql.NullInt64
		spieleA, spieleB, unentschie sql.NullInt64
		stelle                       sql.NullInt64
	)

	if err := zeilen.Scan(&runde.ID, &runde.Nummer, &paarungID, &a, &b,
		&spieleA, &spieleB, &unentschie, &stelle); err != nil {
		return domain.Runde{}, domain.Paarung{}, false,
			uebersetze(err, "eine Zeile des Verlaufs lesen")
	}

	runde.TurnierID = turnierID

	if !paarungID.Valid {
		return runde, domain.Paarung{}, false, nil
	}

	paarung = domain.Paarung{
		ID: paarungID.Int64, RundeID: runde.ID,
		A: a.Int64, B: b.Int64, Stelle: int(stelle.Int64),
	}

	if spieleA.Valid {
		paarung.Ergebnis = &domain.Ergebnis{
			SpieleA:       int(spieleA.Int64),
			SpieleB:       int(spieleB.Int64),
			Unentschieden: int(unentschie.Int64),
		}
	}

	return runde, paarung, true, nil
}
