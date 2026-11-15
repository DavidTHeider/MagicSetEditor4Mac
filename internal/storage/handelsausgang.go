package storage

// **Was noch hinausmuss** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// # Warum eine Nachricht liegen bleibt
//
// **Weil „abgeschickt" nicht „angekommen" heißt.** Ein Ausgang, der nach dem
// Senden vergisst, ist keiner: die Nachricht ist dann weg, und niemand weiß,
// ob die andere Seite sie hat. Erst ihre Bestätigung räumt die Zeile.
//
// # Warum die Versuche gezählt werden
//
// **Damit ein totes Gegenüber nicht endlos angerufen wird.** Die Zahl steht
// an der Zeile; **was** bei welcher Zahl geschieht, entscheidet der Dienst —
// der Bestand zählt und urteilt nicht.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Ausgangszeile ist eine Nachricht, die noch nicht bestätigt ist.
type Ausgangszeile struct {
	NachrichtID string
	Vorgang     string

	// Empfaenger ist die Peerkennung des Gegenübers.
	Empfaenger string

	// Nachricht ist Umschlag und Nutzlast als JSON.
	//
	// **Die ganze Nachricht und nicht ihre Kennung**: ein Ausgang, der sie
	// nur nennt, kann sie nach einem Neustart nicht wiederholen — und genau
	// dafür gibt es ihn.
	Nachricht string

	// Versuche ist, wie oft sie schon hinausging.
	Versuche int
}

// Ausgangsablage führt, was noch hinausmuss.
type Ausgangsablage struct{ db *sql.DB }

// NeueAusgangsablage baut sie über einer Datenbank.
func NeueAusgangsablage(db *sql.DB) *Ausgangsablage {
	return &Ausgangsablage{db: db}
}

// Einreihen legt eine Nachricht in den Ausgang.
//
// **Zweimal dieselbe Kennung ergibt eine Zeile.** Wer nach einem Neustart
// noch einmal einreiht, was schon dalag, soll es nicht doppelt schicken — und
// er soll dabei auch keinen Fehler bekommen, denn getan ist getan.
func (a *Ausgangsablage) Einreihen(ctx context.Context, zeile Ausgangszeile) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	_, fehler = a.db.ExecContext(ctx,
		`INSERT INTO handelsausgang
		   (profil_id, nachricht_id, vorgang, empfaenger, nachricht)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT (profil_id, nachricht_id) DO NOTHING`,
		profilID, zeile.NachrichtID, zeile.Vorgang, zeile.Empfaenger,
		zeile.Nachricht)

	return uebersetze(fehler, "eine Nachricht einreihen")
}

// Offene gibt, was für dieses Gegenüber noch hinausmuss.
//
// **In der Reihenfolge, in der es angelegt wurde** — eine Folge, die drüben
// verkehrt ankommt, ist keine.
func (a *Ausgangsablage) Offene(
	ctx context.Context, empfaenger string,
) ([]Ausgangszeile, error) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return nil, fehler
	}

	zeilen, fehler := a.db.QueryContext(ctx,
		`SELECT nachricht_id, vorgang, nachricht, versuche
		   FROM handelsausgang
		  WHERE profil_id = ? AND empfaenger = ?
		  ORDER BY angelegt, nachricht_id`,
		profilID, empfaenger)
	if fehler != nil {
		return nil, uebersetze(fehler, "den Ausgang lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // nur zum Aufräumen

	// **Eine leere Scheibe und kein `nil`** — sonst fällt aus einer Antwort
	// `null` heraus, wo eine Fläche eine Liste erwartet (§3).
	offen := []Ausgangszeile{}

	for zeilen.Next() {
		zeile := Ausgangszeile{Empfaenger: empfaenger}

		if fehler := zeilen.Scan(&zeile.NachrichtID, &zeile.Vorgang,
			&zeile.Nachricht, &zeile.Versuche); fehler != nil {
			return nil, uebersetze(fehler, "eine Ausgangszeile lesen")
		}

		offen = append(offen, zeile)
	}

	return offen, uebersetze(zeilen.Err(), "den Ausgang lesen")
}

// Versucht zählt einen Versuch — die Zeile bleibt liegen.
func (a *Ausgangsablage) Versucht(ctx context.Context, nachrichtID string) error {
	return a.eineZeile(ctx,
		`UPDATE handelsausgang
		    SET versuche = versuche + 1, zuletzt = datetime('now')
		  WHERE profil_id = ? AND nachricht_id = ?`,
		nachrichtID, "einen Versuch zählen")
}

// Bestaetigen räumt eine Nachricht weg, die angekommen ist.
//
// **Eine Bestätigung ins Leere ist ein Nein.** Sonst hielte jemand eine
// Nachricht für zugestellt, die nie dalag — und der Ausgang sagte nichts
// darüber, ob er je etwas davon wusste.
func (a *Ausgangsablage) Bestaetigen(ctx context.Context, nachrichtID string) error {
	return a.eineZeile(ctx,
		`DELETE FROM handelsausgang
		  WHERE profil_id = ? AND nachricht_id = ?`,
		nachrichtID, "eine Nachricht bestätigen")
}

// eineZeile fährt eine Anweisung und verlangt, dass sie eine Zeile trifft.
func (a *Ausgangsablage) eineZeile(
	ctx context.Context, anweisung, nachrichtID, wobei string,
) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	ergebnis, fehler := a.db.ExecContext(ctx, anweisung, profilID, nachrichtID)
	if fehler != nil {
		return uebersetze(fehler, wobei)
	}

	getroffen, fehler := ergebnis.RowsAffected()
	if fehler != nil {
		return uebersetze(fehler, wobei)
	}

	if getroffen == 0 {
		return domain.MitGrund(domain.ErrNotFound, nil,
			"Diese Nachricht liegt nicht im Ausgang.")
	}

	return nil
}

// Poststand ist, was der Ausgang über ein Gegenüber sagt.
//
// **Zahlen und keine Nachrichten.** Eine Diagnose, die den Inhalt zeigte,
// wäre eine zweite Sicht auf den Handel — und die gibt es schon. Was hier
// steht, sagt, ob etwas klemmt, und nicht, worum es geht.
type Poststand struct {
	// Offen ist die Zahl der Nachrichten, die noch auf ihre Bestätigung
	// warten.
	Offen int

	// HoechsteVersuche ist, wie oft die hartnäckigste schon hinausging.
	//
	// **Die höchste und nicht die Summe**: eine Zahl, die mit jeder
	// Nachricht wächst, sagt nichts darüber, ob eine davon klemmt.
	HoechsteVersuche int
}

// Stand zählt, was für ein Gegenüber noch aussteht.
func (a *Ausgangsablage) Stand(
	ctx context.Context, empfaenger string,
) (Poststand, error) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return Poststand{}, fehler
	}

	var stand Poststand

	// **`coalesce`, weil `max` über null Zeilen `NULL` gibt** — und ein
	// `NULL` in einer Zahl ist ein Lesefehler und keine Null.
	if fehler := a.db.QueryRowContext(ctx,
		`SELECT count(*), coalesce(max(versuche), 0)
		   FROM handelsausgang
		  WHERE profil_id = ? AND empfaenger = ?`,
		profilID, empfaenger).
		Scan(&stand.Offen, &stand.HoechsteVersuche); fehler != nil {
		return Poststand{}, uebersetze(fehler, "den Poststand zählen")
	}

	return stand, nil
}
