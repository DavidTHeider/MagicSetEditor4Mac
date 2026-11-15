package storage

// **Was der Bestand erlebt hat** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Warum es hier keinen Löschweg gibt
//
// **Weil ein Journal, aus dem man etwas herausnehmen kann, keins ist.** Eine
// falsche Buchung bekommt eine Gegenbewegung mit Bezug und Grund
// ([domain.Journalzeile.Gegenbewegung]); beide Zeilen bleiben stehen. Dass
// hier kein `DELETE` steht, ist eine Entscheidung und kein Versäumnis.
//
// # Warum der Beleg seine Form mitspeichert
//
// **Weil eine nachgerechnete Form nicht die unterschriebene ist.** Wer sie
// beim Prüfen neu aus den Zeilen baute, prüfte die Unterschrift gegen das,
// was heute daraus folgt — und nicht gegen das, was damals unterschrieben
// wurde. Genau dieser Unterschied ist der Zweck einer Unterschrift.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Journal gibt die Zeilen zu einem Vorgang, älteste zuerst.
func (a *Abschlussablage) Journal(
	ctx context.Context, vorgang string,
) ([]domain.Journalzeile, error) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return nil, fehler
	}

	zeilen, fehler := a.db.QueryContext(ctx,
		`SELECT id, art, anlass, vorgang, bezug, grund, printing, menge,
		        zustand, sprache, finish, exemplar_id, lagerort, notiz, gebucht
		   FROM handelsjournal
		  WHERE profil_id = ? AND vorgang = ?
		  ORDER BY id`, profilID, vorgang)
	if fehler != nil {
		return nil, uebersetze(fehler, "das Journal lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // der Fehler kommt aus Err()

	gelesen := []domain.Journalzeile{}

	for zeilen.Next() {
		zeile, fehler := eineJournalzeile(zeilen)
		if fehler != nil {
			return nil, fehler
		}

		gelesen = append(gelesen, zeile)
	}

	return gelesen, uebersetze(zeilen.Err(), "das Journal lesen")
}

// eineJournalzeile liest eine Zeile des Journals.
func eineJournalzeile(zeilen *sql.Rows) (domain.Journalzeile, error) {
	var (
		zeile   domain.Journalzeile
		art     string
		anlass  string
		zustand string
		finish  string
		gebucht string
	)

	if fehler := zeilen.Scan(&zeile.ID, &art, &anlass, &zeile.Vorgang, &zeile.Bezug,
		&zeile.Grund, &zeile.Printing, &zeile.Menge, &zustand, &zeile.Sprache,
		&finish, &zeile.ExemplarID, &zeile.Lagerort, &zeile.Notiz,
		&gebucht); fehler != nil {
		return zeile, uebersetze(fehler, "eine Journalzeile lesen")
	}

	zeile.Art = domain.Journalart(art)
	zeile.Anlass = domain.Journalanlass(anlass)
	zeile.Zustand = domain.Zustand(zustand)
	zeile.Finish = domain.Finish(finish)

	uhr, fehler := time.Parse(time.DateTime, gebucht)
	if fehler != nil {
		return zeile, uebersetze(fehler, "die Uhr einer Journalzeile lesen")
	}

	zeile.Gebucht = uhr

	return zeile, nil
}

// Buchen trägt eine Zeile ein, die **kein** Handel bewirkt hat.
//
// **Kauf, Verkauf, Geschenk, Einfuhr und Korrektur gehen hier durch** — der
// Bestand hat sich auf einem anderen Weg geändert, und das Journal hält fest,
// warum. Ein Tausch geht nicht hier durch: er bewegt den Bestand mit, und das
// tut [Abschlussablage.Abschliessen] in einem Zug.
func (a *Abschlussablage) Buchen(
	ctx context.Context, zeile domain.Journalzeile,
) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	if zeile.Anlass == domain.AnlassTausch {
		return domain.MitGrund(domain.ErrValidation, nil,
			"ein Tausch wird abgeschlossen und nicht gebucht — sonst bewegt "+
				"sich das Journal ohne den Bestand")
	}

	return journalSchreiben(ctx, a.db, profilID, []domain.Journalzeile{zeile})
}

// journalSchreiben trägt Zeilen ein — in einer Transaktion oder ohne.
func journalSchreiben(
	ctx context.Context, an abfrager, profilID int64,
	zeilen []domain.Journalzeile,
) error {
	for _, zeile := range zeilen {
		if fehler := zeile.Pruefe(); fehler != nil {
			return fehler
		}

		if _, fehler := an.ExecContext(ctx,
			`INSERT INTO handelsjournal
			   (profil_id, art, anlass, vorgang, bezug, grund, printing,
			    menge, zustand, sprache, finish, exemplar_id, lagerort, notiz)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			profilID, string(zeile.Art), string(zeile.Anlass), zeile.Vorgang,
			zeile.Bezug, zeile.Grund, zeile.Printing, zeile.Menge,
			string(zeile.Zustand), zeile.Sprache, string(zeile.Finish),
			zeile.ExemplarID, zeile.Lagerort, zeile.Notiz,
		); fehler != nil {
			return uebersetze(fehler, "eine Journalzeile schreiben")
		}
	}

	return nil
}

// Beleg gibt den abgelegten Beleg zu einem Vorgang — **oder nichts**.
//
// **`nil` und kein [domain.ErrNotFound]**, aus demselben Grund wie bei
// [Briefkastenablage.Nachschlagen]: „noch nicht abgeschlossen" ist der
// gewöhnliche Fall eines laufenden Handels und kein Fehler.
func (a *Abschlussablage) Beleg(
	ctx context.Context, vorgang string,
) (*domain.Belegstand, error) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return nil, fehler
	}

	var (
		beleg      = domain.Belegstand{Vorgang: vorgang}
		beidseitig int
		gebucht    string
	)

	var (
		gegenzeichner   sql.NullString
		gegengezeichnet sql.NullString
		befund          befundspalten
	)

	// **Ein Lesen und kein zweites.** Wer den Fremdbeleg gesondert holte,
	// hätte zwei Antworten aus zwei Augenblicken — und die zweite könnte
	// einen Beleg zeigen, den die erste noch nicht kannte.
	fehler = a.db.QueryRowContext(ctx,
		`SELECT b.abdruck, b.form, b.unterschrift, b.beidseitig, b.gebucht,
		        f.von, f.empfangen,
		        g.von, g.abdruck, g.mangel, g.grund, g.empfangen
		   FROM handelsbelege b
		   LEFT JOIN handelsfremdbelege f
		     ON f.profil_id = b.profil_id AND f.vorgang = b.vorgang
		   LEFT JOIN handelsgegenbelegbefunde g
		     ON g.profil_id = b.profil_id AND g.vorgang = b.vorgang
		  WHERE b.profil_id = ? AND b.vorgang = ?`, profilID, vorgang).
		Scan(append([]any{
			&beleg.Abdruck, &beleg.Form, &beleg.Unterschrift,
			&beidseitig, &gebucht, &gegenzeichner, &gegengezeichnet,
		},
			befund.ziele()...)...)

	if errors.Is(fehler, sql.ErrNoRows) {
		return nil, nil //nolint:nilnil // „noch nicht" ist hier kein Fehler
	}

	if fehler != nil {
		return nil, uebersetze(fehler, "einen Beleg lesen")
	}

	beleg.Beidseitig = beidseitig == 1
	beleg.Gegenzeichner = gegenzeichner.String

	beleg.Gebucht, fehler = time.Parse(time.DateTime, gebucht)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Uhr eines Belegs lesen")
	}

	if gegengezeichnet.Valid {
		beleg.Gegengezeichnet, fehler = time.Parse(
			time.DateTime, gegengezeichnet.String)
		if fehler != nil {
			return nil, uebersetze(
				fehler, "die Uhr eines Fremdbelegs lesen")
		}
	}

	// **Und der jüngste abgewiesene Gegenbeleg** (#2500).
	if beleg.Befund, fehler = befund.befund(vorgang); fehler != nil {
		return nil, fehler
	}

	return &beleg, nil
}

// Gegenzeichnen legt den geprüften Beleg der Gegenseite ab.
//
// **Es ändert die eigene Form nicht.** Was hier unterschrieben wurde, steht
// fest; hinzu kommt allein, dass ein Zweiter denselben Handel unterschrieben
// hat. Wer hier die Form neu setzte, machte aus zwei Unterschriften unter
// zwei Texte eine unter einen.
//
// **Und es nimmt jetzt einen Beleg statt eines Vorgangs**
// (#1650).
// Bis zum 2026-09-19 setzte es `beidseitig` auf die bloße Behauptung hin,
// hier habe jemand für beide Parteien bestätigt. **Beides steht in einer
// Transaktion**: ein Fremdbeleg ohne gesetztes `beidseitig` wäre unsichtbar,
// ein gesetztes `beidseitig` ohne Fremdbeleg wäre wieder die Behauptung.
func (a *Abschlussablage) Gegenzeichnen(
	ctx context.Context, fremd domain.Fremdbeleg,
) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	zug, fehler := a.db.BeginTx(ctx, nil)
	if fehler != nil {
		return uebersetze(fehler, "einen Beleg gegenzeichnen")
	}

	defer func() { _ = zug.Rollback() }() //nolint:errcheck // nach dem Commit ohne Wirkung

	if fehler := gegenzeichnen(ctx, zug, profilID, fremd); fehler != nil {
		return fehler
	}

	return uebersetze(zug.Commit(), "einen Beleg gegenzeichnen")
}

// gegenzeichnen schreibt beides — den Vermerk am eigenen Beleg und den
// fremden daneben.
func gegenzeichnen(
	ctx context.Context, an abfrager, profilID int64, fremd domain.Fremdbeleg,
) error {
	vorgang := fremd.Vorgang

	ergebnis, fehler := an.ExecContext(ctx,
		`UPDATE handelsbelege SET beidseitig = 1
		  WHERE profil_id = ? AND vorgang = ?`, profilID, vorgang)
	if fehler != nil {
		return uebersetze(fehler, "einen Beleg gegenzeichnen")
	}

	getroffen, fehler := ergebnis.RowsAffected()
	if fehler != nil {
		return uebersetze(fehler, "einen Beleg gegenzeichnen")
	}

	if getroffen == 0 {
		return fmt.Errorf("%w: zum Vorgang %s steht kein Beleg",
			domain.ErrNotFound, vorgang)
	}

	// **Zweimal derselbe Beleg ist kein Fehler.** Eine Zustellung darf sich
	// wiederholen — der Briefkasten ist idempotent, und hier gilt dasselbe:
	// derselbe Unterzeichner unter derselben Revision ist dieselbe Tatsache.
	abgelegt, fehler := an.ExecContext(ctx,
		`INSERT INTO handelsfremdbelege
		   (profil_id, vorgang, von, abdruck, unterschrift)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT (profil_id, vorgang) DO NOTHING`,
		profilID, vorgang, fremd.Von, fremd.Abdruck, fremd.Unterschrift,
	)
	if fehler != nil {
		return uebersetze(fehler, "einen Fremdbeleg ablegen")
	}

	if neu, _ := abgelegt.RowsAffected(); neu == 1 { //nolint:errcheck // SQLite meldet die Zahl immer
		return nil
	}

	return derselbeFremdbeleg(ctx, an, profilID, fremd)
}

// derselbeFremdbeleg prüft, ob der schon abgelegte derselbe ist.
//
// **Ein zweiter, abweichender Beleg ist ein Konflikt** (TRADE-03.AK-03,
// #1717): bis hierher nahm `DO NOTHING` ihn still hin, und wer drüben eine
// andere Revision unterschrieb, hinterließ hier keine Spur.
func derselbeFremdbeleg(
	ctx context.Context, an abfrager, profilID int64, fremd domain.Fremdbeleg,
) error {
	var von, abdruck, unterschrift string

	if fehler := an.QueryRowContext(ctx,
		`SELECT von, abdruck, unterschrift FROM handelsfremdbelege
		  WHERE profil_id = ? AND vorgang = ?`, profilID, fremd.Vorgang,
	).Scan(&von, &abdruck, &unterschrift); fehler != nil {
		return uebersetze(fehler, "einen Fremdbeleg nachschlagen")
	}

	if von != fremd.Von || abdruck != fremd.Abdruck || unterschrift != fremd.Unterschrift {
		return fmt.Errorf("%w: zum Vorgang %s liegt schon ein anderer Beleg "+
			"des Gegenübers — der Handel gehört in die Klärung",
			domain.ErrConflict, fremd.Vorgang)
	}

	return nil
}
