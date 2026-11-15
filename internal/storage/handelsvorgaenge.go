package storage

// **Der Handelsvertrag im Bestand** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// **Eine Zeile je Revision, das Angebot als JSON** — der Grund steht in der
// Wanderung 0068. Was daneben in Spalten steht, ist das, wonach gesucht wird.

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Handelsablage hält die Handelsvorgänge eines Profils.
type Handelsablage struct{ db *sql.DB }

// NeueHandelsablage baut sie.
func NeueHandelsablage(db *sql.DB) *Handelsablage { return &Handelsablage{db: db} }

// Schreiben legt eine Revision an oder schreibt ihren Zustand fort.
//
// **Ein Aufruf und kein Anlegen-und-Ändern.** Eine **neue** Revision kommt
// dazu; eine Zustimmung ändert dieselbe Zeile, weil sie keine neue Revision
// ist. Zwei Wege dafür wären zwei Stellen, an denen jemand vergisst, welche
// gilt — derselbe Satz wie bei [Leihablage.LeiheSchreiben].
func (a *Handelsablage) Schreiben(
	ctx context.Context, angebot domain.Handelsangebot,
) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	return revisionSchreiben(ctx, a.db, profilID, angebot)
}

// revisionSchreiben legt eine Revision an oder schreibt sie fort.
//
// **Über einen [abfrager] und nicht über die Verbindung**, damit dieselbe
// Anweisung auch **in** einer Transaktion laufen kann — der Briefkasten
// schreibt Vermerk und Wirkung in einem Zug (T5, #1470).
func revisionSchreiben(
	ctx context.Context, an abfrager, profilID int64,
	angebot domain.Handelsangebot,
) error {
	// **Ohne Fehlerzweig, und das ist keine Nachlässigkeit**: ein
	// [domain.Handelsangebot] trägt Zeichenketten, Zahlen, einen
	// Wahrheitswert und Scheiben davon — `encoding/json` kann daran nicht
	// scheitern. Ein Zweig, den kein Lauf betreten kann, ist eine Zusage,
	// die niemand hält (§10); wer dem Verbund ein `any` gibt, ändert diese
	// Stelle mit.
	roh, _ := json.Marshal(angebot) //nolint:errcheck // der Grund steht darüber

	if _, fehler := an.ExecContext(ctx,
		`INSERT INTO handelsrevisionen
		   (profil_id, vorgang, revision, abdruck, eltern, fassung, zustand,
		    angebot)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (profil_id, vorgang, revision) DO UPDATE SET
		   zustand   = excluded.zustand,
		   angebot   = excluded.angebot,
		   geaendert = datetime('now')`,
		profilID, angebot.Vorgang, angebot.Revision, angebot.Abdruck(),
		angebot.Eltern, angebot.Vertragsfassung, string(angebot.Zustand),
		string(roh),
	); fehler != nil {
		return fmt.Errorf("die Revision ließ sich nicht schreiben: %w", fehler)
	}

	return nil
}

// Vorgang gibt die **letzte** Revision eines Vorgangs.
func (a *Handelsablage) Vorgang(
	ctx context.Context, kennung string,
) (domain.Handelsangebot, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Handelsangebot{}, profilfehler
	}

	var roh string

	fehler := a.db.QueryRowContext(ctx,
		`SELECT angebot FROM handelsrevisionen
		  WHERE profil_id = ? AND vorgang = ?
		  ORDER BY revision DESC LIMIT 1`,
		profilID, kennung,
	).Scan(&roh)

	if errors.Is(fehler, sql.ErrNoRows) {
		return domain.Handelsangebot{}, domain.MitGrund(domain.ErrNotFound,
			nil, "den Handelsvorgang gibt es nicht")
	}

	if fehler != nil {
		return domain.Handelsangebot{}, fmt.Errorf(
			"der Handelsvorgang ließ sich nicht lesen: %w", fehler)
	}

	return entpackeAngebot(roh)
}

// Revisionen gibt die ganze Kette, von der ersten an.
//
// **Die Reihenfolge ist die Kette** — wer sie anders sortierte, müsste sie
// beim Lesen wieder zusammensetzen, und **wie** das geschieht, wäre eine
// zweite Verabredung (T5,
// #1470).
func (a *Handelsablage) Revisionen(
	ctx context.Context, vorgang string,
) ([]domain.Handelsangebot, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, fehler := a.db.QueryContext(ctx,
		`SELECT angebot FROM handelsrevisionen
		  WHERE profil_id = ? AND vorgang = ?
		  ORDER BY revision ASC`,
		profilID, vorgang)
	if fehler != nil {
		return nil, fmt.Errorf("die Kette ließ sich nicht lesen: %w", fehler)
	}

	return angeboteLesen(zeilen, "die Kette")
}

// Letzte gibt je Vorgang die letzte Revision — zuletzt geänderte zuerst.
//
// **Für die Wahl eines Handels** (UX-08.AK-04,
// #1680):
// wer einen Handel empfängt, soll ihn wählen und nicht seine Kennung tippen.
// **Wer daran beteiligt ist, entscheidet die Leitung** — die Ablage liest,
// was das Profil hat.
func (a *Handelsablage) Letzte(
	ctx context.Context,
) ([]domain.Handelsangebot, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, fehler := a.db.QueryContext(ctx,
		`SELECT angebot FROM handelsrevisionen AS h
		  WHERE profil_id = ?
		    AND revision = (SELECT MAX(revision) FROM handelsrevisionen
		                     WHERE profil_id = h.profil_id
		                       AND vorgang = h.vorgang)
		  ORDER BY geaendert DESC, vorgang`,
		profilID)
	if fehler != nil {
		return nil, fmt.Errorf("die Handelsvorgänge ließen sich nicht lesen: %w",
			fehler)
	}

	return angeboteLesen(zeilen, "die Liste der Handelsvorgänge")
}

// angeboteLesen liest Revisionen aus ihren Zeilen — `was` nennt, wobei.
//
// **Eine leere Liste und kein `nil`** — „hier gibt es keinen" ist eine
// Antwort, und sie sieht anders aus als keine.
func angeboteLesen(
	zeilen *sql.Rows, was string,
) ([]domain.Handelsangebot, error) {
	defer func() { _ = zeilen.Close() }() //nolint:errcheck // der Fehler steht unten

	angebote := []domain.Handelsangebot{}

	for zeilen.Next() {
		var roh string

		if fehler := zeilen.Scan(&roh); fehler != nil {
			return nil, fmt.Errorf("eine Revision ließ sich nicht lesen: %w",
				fehler)
		}

		angebot, fehler := entpackeAngebot(roh)
		if fehler != nil {
			return nil, fehler
		}

		angebote = append(angebote, angebot)
	}

	if fehler := zeilen.Err(); fehler != nil {
		return nil, fmt.Errorf("%s brach ab: %w", was, fehler)
	}

	return angebote, nil
}

// entpackeAngebot liest eine Revision aus ihrem JSON.
//
// **Ein Befund und kein Absturz** (§8): ein von Hand geänderter Bestand darf
// die Anwendung nicht umwerfen — er bekommt einen Satz, der sagt, was los
// ist.
func entpackeAngebot(roh string) (domain.Handelsangebot, error) {
	var angebot domain.Handelsangebot

	if fehler := json.Unmarshal([]byte(roh), &angebot); fehler != nil {
		return domain.Handelsangebot{}, domain.MitGrund(domain.ErrValidation,
			fehler, "die gespeicherte Revision ließ sich nicht lesen")
	}

	return angebot, nil
}
