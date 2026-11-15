package storage

// **Der Turnierspeicher** — Werkstück 3 von Phase 4.7
// (#1294),
// nach Vermerk [0062](../../docs/entscheidungen/0062-was-ein-turnier-hier-ist.md).
//
// # Was er nicht ablegt
//
// **Die Rangliste.** Sie wird gerechnet (`domain.Rangliste`). Eine Spalte
// `matchpunkte` wäre der Anfang des Auseinanderlaufens, gegen das Vermerk
// 0062 steht.
//
// # Was in einer Transaktion steht und warum
//
// **Eine Runde entsteht ganz oder gar nicht.** Das ist die Zusage aus dem
// Issue — „ein Absturz mitten in der Ergebniserfassung hinterlässt keine
// halbe Runde" —, und sie ist eine Transaktionsgrenze, kein Zustandsfeld.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Turnierablage liest und schreibt Turniere samt Verlauf.
type Turnierablage struct{ db *sql.DB }

// NeueTurnierablage gibt die Ablage über diesem Bestand.
func NeueTurnierablage(db *sql.DB) *Turnierablage {
	return &Turnierablage{db: db}
}

// Turnier gibt ein Turnier samt Teilnehmern und Verlauf.
func (a *Turnierablage) Turnier(ctx context.Context, kennung int64) (
	domain.Turnier, error,
) {
	var (
		turnier domain.Turnier
		format  sql.NullInt64
	)

	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Turnier{}, profilfehler
	}

	err := a.db.QueryRowContext(ctx,
		`SELECT id, fassung, name, spiel, einzelspiel, notiz, format_id,
		        coalesce(abgeschlossen, '')
		   FROM turniere WHERE id = ? AND profil_id = ?`, kennung, profilID).
		Scan(&turnier.ID, &turnier.Fassung, &turnier.Name, &turnier.Spiel,
			&turnier.Einzelspiel, &turnier.Notiz, &format, &turnier.Abgeschlossen)

	turnier.FormatID = format.Int64

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.Turnier{}, fmt.Errorf(
			"%w: ein Turnier mit dieser Kennung gibt es nicht", domain.ErrNotFound)

	case err != nil:
		return domain.Turnier{}, uebersetze(err, "das Turnier lesen")
	}

	if turnier.Teilnehmer, err = a.teilnehmer(ctx, kennung); err != nil {
		return domain.Turnier{}, err
	}

	if turnier.Runden, err = a.runden(ctx, kennung); err != nil {
		return domain.Turnier{}, err
	}

	return turnier, nil
}

// Anmelden trägt einen Teilnehmer ein — **samt seiner eingereichten Liste**.
//
// **Die Stelle kommt vom Speicher und nicht vom Aufrufer.** Zwei Flächen, die
// gleichzeitig anmelden, vergäben sonst dieselbe, und der eindeutige Index
// wiese die zweite ab, obwohl beide gültig sind.
//
// **Und beides entsteht in einem Zug.** Ein Teilnehmer ohne seine Liste wäre
// ein eingereichtes Dokument, von dem die Hälfte fehlt — genau das, was
// Vermerk 0062 verhindert. Bis zum 2026-09-01 wurde die Liste **gar nicht**
// geschrieben: `turnierdeckplaetze` hatte weder Schreiber noch Leser (Befund
// 1 des Zweitblicks über Phase 4.7).
func (a *Turnierablage) Anmelden(
	ctx context.Context, turnierID int64, teilnehmer domain.Teilnehmer,
) (domain.Teilnehmer, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Teilnehmer{}, profilfehler
	}

	zug, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Teilnehmer{}, uebersetze(err, "die Anmeldung beginnen")
	}
	defer zug.Rollback() //nolint:errcheck // nach einem Commit ist es ein No-op

	// **Nur in ein Turnier des gezeigten Profils, und nur mit einem Deck aus
	// ihm** (#1815).
	if fehler := turnierImProfil(ctx, zug, turnierID); fehler != nil {
		return domain.Teilnehmer{}, fehler
	}

	if teilnehmer.DeckID != 0 {
		if fehler := deckImProfil(ctx, zug, teilnehmer.DeckID); fehler != nil {
			return domain.Teilnehmer{}, fehler
		}
	}

	angemeldet, err := meldeAnImZug(ctx, zug, turnierID, teilnehmer)
	if err != nil {
		return domain.Teilnehmer{}, err
	}

	// **Eine Anmeldung ist eine Änderung am Turnier** — die Fassung zählt sie.
	if err := fortschreiben(ctx, zug, turnierID, profilID, domain.Protokolleintrag{
		Vorgang: domain.TurnierAngemeldet, Angaben: angemeldet.Name,
	}); err != nil {
		return domain.Teilnehmer{}, err
	}

	if err := zug.Commit(); err != nil {
		return domain.Teilnehmer{}, uebersetze(err, "die Anmeldung abschließen")
	}

	return angemeldet, nil
}

// meldeAnImZug ist der Rumpf von [Turnierablage.Anmelden].
func meldeAnImZug(
	ctx context.Context, zug *sql.Tx, turnierID int64, teilnehmer domain.Teilnehmer,
) (domain.Teilnehmer, error) {
	var stelle int

	if err := zug.QueryRowContext(ctx,
		`SELECT coalesce(max(stelle), 0) + 1 FROM turnierteilnehmer
		  WHERE turnier_id = ?`, turnierID).Scan(&stelle); err != nil {
		return domain.Teilnehmer{}, uebersetze(err, "die nächste Stelle lesen")
	}

	ergebnis, err := zug.ExecContext(ctx,
		`INSERT INTO turnierteilnehmer (turnier_id, name, deck_id, ausgestiegen, stelle)
		 VALUES (?, ?, ?, ?, ?)`,
		turnierID, teilnehmer.Name, kennungOderNull(teilnehmer.DeckID), teilnehmer.Ausgestiegen, stelle)
	if err != nil {
		return domain.Teilnehmer{}, uebersetze(err, "den Teilnehmer anmelden")
	}

	kennung, idfehler := ergebnis.LastInsertId()
	if idfehler != nil {
		return domain.Teilnehmer{}, uebersetze(idfehler,
			"die Kennung des Teilnehmers lesen")
	}

	teilnehmer.ID = kennung
	teilnehmer.TurnierID = turnierID
	teilnehmer.Stelle = stelle

	for platz := range teilnehmer.Deckabzug {
		gelegt, err := legeAbzugsplatz(ctx, zug, kennung,
			teilnehmer.Deckabzug[platz], platz+1)
		if err != nil {
			return domain.Teilnehmer{}, err
		}

		teilnehmer.Deckabzug[platz] = gelegt
	}

	return teilnehmer, nil
}

// legeAbzugsplatz schreibt eine Zeile der eingereichten Liste.
func legeAbzugsplatz(
	ctx context.Context, zug *sql.Tx, teilnehmerID int64,
	platz domain.Abzugsplatz, stelle int,
) (domain.Abzugsplatz, error) {
	druck, druckfehler := druckZeile(ctx, zug, druckbezug{
		druckID: platz.DruckID, printing: platz.Printing, name: platz.Bezeichnung,
	})
	if druckfehler != nil {
		return domain.Abzugsplatz{}, druckfehler
	}

	ergebnis, err := zug.ExecContext(ctx,
		`INSERT INTO turnierdeckplaetze
		        (teilnehmer_id, bezeichnung, printing, druck_id, menge, bereich, stelle)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		teilnehmerID, platz.Bezeichnung, platz.Printing, druck, platz.Menge,
		string(platz.Bereich), stelle)
	if err != nil {
		return domain.Abzugsplatz{}, uebersetze(err, "eine Zeile der Liste schreiben")
	}

	kennung, idfehler := ergebnis.LastInsertId()

	platz.ID = kennung
	platz.Stelle = stelle

	return platz, uebersetze(idfehler, "die Kennung der Zeile lesen")
}

// teilnehmer liest die Teilnehmer eines Turniers — **in Anmeldereihenfolge**.
func (a *Turnierablage) teilnehmer(ctx context.Context, turnierID int64) (
	[]domain.Teilnehmer, error,
) {
	zeilen, err := a.db.QueryContext(ctx,
		`SELECT id, turnier_id, name, coalesce(deck_id, 0), ausgestiegen, stelle
		   FROM turnierteilnehmer WHERE turnier_id = ? ORDER BY stelle`, turnierID)
	if err != nil {
		return nil, uebersetze(err, "die Teilnehmer lesen")
	}
	defer zeilen.Close() //nolint:errcheck // gelesen wird der Fehler von Err()

	var alle []domain.Teilnehmer

	for zeilen.Next() {
		var einer domain.Teilnehmer

		if err := zeilen.Scan(&einer.ID, &einer.TurnierID, &einer.Name,
			&einer.DeckID, &einer.Ausgestiegen, &einer.Stelle); err != nil {
			return nil, uebersetze(err, "einen Teilnehmer lesen")
		}

		alle = append(alle, einer)
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "die Teilnehmer lesen")
	}

	return a.mitAbzuegen(ctx, turnierID, alle)
}

// mitAbzuegen hängt jedem Teilnehmer seine eingereichte Liste an.
//
// **Eine Abfrage für alle und nicht eine je Teilnehmer.** Bei acht
// Teilnehmern wären das acht Rufe für eine Auskunft; dieselbe Überlegung wie
// bei der Bündelung der Formatprüfung (#1373).
//
// **Und jeder bekommt eine Liste, auch wer keine hat.** `null` und `[]` sind
// in JSON zwei verschiedene Antworten, und eine Seite, die über `null`
// iteriert, bricht ab.
func (a *Turnierablage) mitAbzuegen(
	ctx context.Context, turnierID int64, alle []domain.Teilnehmer,
) ([]domain.Teilnehmer, error) {
	for stelle := range alle {
		alle[stelle].Deckabzug = []domain.Abzugsplatz{}
	}

	zeilen, err := a.db.QueryContext(ctx,
		`SELECT p.teilnehmer_id, p.id, p.bezeichnung, p.printing, p.druck_id,
		        p.menge, p.bereich, p.stelle
		   FROM turnierdeckplaetze p
		   JOIN turnierteilnehmer t ON t.id = p.teilnehmer_id
		  WHERE t.turnier_id = ?
		  ORDER BY p.teilnehmer_id, p.stelle`, turnierID)
	if err != nil {
		return nil, uebersetze(err, "die eingereichten Listen lesen")
	}
	defer zeilen.Close() //nolint:errcheck // gelesen wird der Fehler von Err()

	nach := make(map[int64]int, len(alle))
	for stelle := range alle {
		nach[alle[stelle].ID] = stelle
	}

	for zeilen.Next() {
		var (
			wem     int64
			platz   domain.Abzugsplatz
			druck   sql.NullInt64
			bereich string
		)

		if err := zeilen.Scan(&wem, &platz.ID, &platz.Bezeichnung,
			&platz.Printing, &druck, &platz.Menge, &bereich, &platz.Stelle); err != nil {
			return nil, uebersetze(err, "eine Zeile der Liste lesen")
		}

		platz.Bereich = domain.Deckbereich(bereich)
		platz.DruckID = druck.Int64

		// **Eine Zeile ohne Teilnehmer wird übergangen.** Sie kann es nach
		// `ON DELETE CASCADE` nicht geben; sie hier anzulegen verdeckte einen
		// Bestand, den jemand von Hand verändert hat.
		if stelle, gibts := nach[wem]; gibts {
			alle[stelle].Deckabzug = append(alle[stelle].Deckabzug, platz)
		}
	}

	return alle, uebersetze(zeilen.Err(), "die eingereichten Listen lesen")
}

// kennungOderNull macht aus einer Null ein SQL-NULL.
//
// **Weil der Fremdschlüssel eine Null nicht kennt.** `deck_id = 0` verwiese
// auf ein Deck mit der Kennung 0, und das gibt es nicht — SQLite wiese es ab.
func kennungOderNull(kennung int64) any {
	if kennung == 0 {
		return nil
	}

	return kennung
}

// Turniere listet die Turniere eines Spiels — **leer heißt: alle**.
//
// **Ohne Teilnehmer und ohne Verlauf.** Eine Übersicht, die jedes Turnier
// vollständig läse, zöge bei zehn Turnieren zehn Verläufe mit — und zeigte
// keinen davon.
func (a *Turnierablage) Turniere(ctx context.Context, spiel string) (
	[]domain.Turnier, error,
) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, err := a.db.QueryContext(ctx,
		`SELECT id, fassung, name, spiel, einzelspiel, notiz, format_id,
		        coalesce(abgeschlossen, '')
		   FROM turniere
		  WHERE profil_id = ? AND (? = '' OR spiel = ?)
		  ORDER BY name`, profilID, spiel, spiel)
	if err != nil {
		return nil, uebersetze(err, "die Turniere lesen")
	}
	defer zeilen.Close() //nolint:errcheck // gelesen wird der Fehler von Err()

	alle := []domain.Turnier{}

	for zeilen.Next() {
		var (
			eines  domain.Turnier
			format sql.NullInt64
		)

		if err := zeilen.Scan(&eines.ID, &eines.Fassung, &eines.Name,
			&eines.Spiel, &eines.Einzelspiel, &eines.Notiz, &format, &eines.Abgeschlossen); err != nil {
			return nil, uebersetze(err, "ein Turnier lesen")
		}

		eines.FormatID = format.Int64

		alle = append(alle, eines)
	}

	return alle, uebersetze(zeilen.Err(), "die Turniere lesen")
}
