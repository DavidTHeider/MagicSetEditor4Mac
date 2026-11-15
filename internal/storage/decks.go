// **Der Deckspeicher** — Werkstück 2 von Phase 4.6
// (#1293),
// nach Vermerk
// [0061](../../docs/entscheidungen/0061-was-ein-deck-hier-ist.md).
//
// # Warum ein Deck seine Plätze mitbringt
//
// Ein Deck ohne seine Plätze ist eine Überschrift. Jeder Aufrufer, der eines
// liest, will wissen, was darin liegt — und zwei Rufe dafür wären zwei
// Gelegenheiten, den zweiten zu vergessen.

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Deckablage liest und schreibt Decks.
type Deckablage struct{ db *sql.DB }

// NeueDeckablage gibt die Ablage über diesem Bestand.
func NeueDeckablage(db *sql.DB) *Deckablage {
	return &Deckablage{db: db}
}

// Anlegen legt ein Deck an und gibt es mit seiner Kennung zurück.
func (a *Deckablage) Anlegen(ctx context.Context, deck domain.Deck) (
	domain.Deck, error,
) {
	// **Geprüft wird im Dienst**, nicht hier (§2): der Speicher setzt um, was
	// entschieden ist. Was der Bestand selbst hält, steht als `CHECK` in der
	// Wanderung.

	// **Jede Wurzelabfrage klemmt auf ein Profil** (#1351, Vermerk 0066).
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Deck{}, profilfehler
	}

	return deckEinfuegen(ctx, a.db, deck, profilID)
}

// Deck gibt ein Deck **samt seinen Plätzen**.
func (a *Deckablage) Deck(ctx context.Context, kennung int64) (domain.Deck, error) {
	var (
		deck   domain.Deck
		format sql.NullInt64
	)

	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Deck{}, profilfehler
	}

	var art string

	// **Die Herkunft kommt im selben Ruf** — Werkstück 12 (#1351). Ein
	// zweiter Ruf wäre eine zweite Stelle, an die jemand denken muss; ein
	// `LEFT JOIN` ist keine. **`coalesce`, weil die Herkunft meistens fehlt**:
	// die leere ist die Auskunft „hier entstanden".
	err := a.db.QueryRowContext(ctx,
		`SELECT d.id, d.fassung, d.name, d.spiel, d.notiz, d.format_id,
		        coalesce(h.art, ''), coalesce(h.quelle, ''),
		        coalesce(h.quell_profil, 0), coalesce(h.uebernommen, '')
		   FROM decks d LEFT JOIN deck_herkunft h ON h.deck_id = d.id
		  WHERE d.id = ? AND d.profil_id = ?`,
		kennung, profilID).Scan(&deck.ID, &deck.Fassung, &deck.Name, &deck.Spiel,
		&deck.Notiz, &format, &art, &deck.Herkunft.Quelle,
		&deck.Herkunft.QuellProfil, &deck.Herkunft.Uebernommen)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Deck{}, fmt.Errorf(
				"%w: ein Deck mit dieser Kennung gibt es nicht", domain.ErrNotFound)
		}

		return domain.Deck{}, uebersetze(err, "das Deck lesen")
	}

	deck.Herkunft.Art = domain.Herkunftsart(art)

	plaetze, err := a.plaetze(ctx, kennung)
	if err != nil {
		return domain.Deck{}, err
	}

	deck.Plaetze = plaetze
	deck.FormatID = format.Int64

	// **Hier aus den Plätzen** — die Übersicht zählt in SQL, dieser Weg hat
	// sie ohnehin.
	deck.Zaehlen()

	return deck, nil
}

// plaetze gibt die Plätze eines Decks — **in der gelegten Reihenfolge**.
//
// **Geordnet vom Bestand und nicht vom Aufrufer**: sonst sortierte jede Fläche
// für sich, und zwei von ihnen liefen auseinander.
func (a *Deckablage) plaetze(ctx context.Context, deck int64) (
	[]domain.Deckplatz, error,
) {
	return plaetzeIn(ctx, a.db, deck)
}

// plaetzeIn liest die Plätze über eine Verbindung oder einen Zug — die Kopie
// (DOM-03) liest sie im selben Zug, in dem sie entstehen.
func plaetzeIn(ctx context.Context, wo abfrager, deck int64) (
	[]domain.Deckplatz, error,
) {
	zeilen, err := wo.QueryContext(ctx,
		`SELECT id, karte_id, printing, druck_id, menge, bereich
		   FROM deckplaetze WHERE deck_id = ? ORDER BY stelle, id`, deck)
	if err != nil {
		return nil, uebersetze(err, "die Deckplätze lesen")
	}

	defer zeilen.Close() //nolint:errcheck // der Fehler steht in zeilen.Err()

	plaetze := []domain.Deckplatz{}

	for zeilen.Next() {
		var (
			platz    domain.Deckplatz
			karteID  sql.NullInt64
			printing sql.NullString
			druck    sql.NullInt64
		)

		if err := zeilen.Scan(&platz.ID, &karteID, &printing, &druck,
			&platz.Menge, &platz.Bereich); err != nil {
			return nil, uebersetze(err, "einen Deckplatz lesen")
		}

		platz.KarteID = karteID.Int64
		platz.Printing = printing.String
		platz.DruckID = druck.Int64

		plaetze = append(plaetze, platz)
	}

	// **Ohne eigenen Zweig** — dieselbe Form wie in `autoersetzungen.go`:
	// `uebersetze` gibt für `nil` nichts zurück.
	return plaetze, uebersetze(zeilen.Err(), "die Deckplätze lesen")
}

// Decks gibt die Decks eines Spiels — **ohne ihre Plätze**.
//
// **Ohne Plätze**, weil eine Übersicht sie nicht zeigt: für zwanzig Decks
// wären das zwanzig Abfragen für etwas, das niemand liest.
//
// **Ein leeres Spiel heißt: alle.** Die Übersicht über den ganzen Bestand ist
// dieselbe Frage ohne Einschränkung, und ein zweiter Weg dafür wäre eine
// zweite Fassung derselben Abfrage.
func (a *Deckablage) Decks(ctx context.Context, spiel string) ([]domain.Deck, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	// **Die Zahlen kommen aus derselben Abfrage** und nicht aus den Plätzen:
	// für zwanzig Decks wären das zwanzig Abfragen. Ein `LEFT JOIN` ist eine.
	//
	// **`coalesce`, weil ein Deck ohne Plätze keine Zeile hat** — `SUM` gäbe
	// dort `NULL`, und das ist keine Null.
	zeilen, err := a.db.QueryContext(ctx,
		`SELECT d.id, d.fassung, d.name, d.spiel, d.notiz, d.format_id,
		        coalesce(sum(CASE WHEN p.bereich = 'hauptdeck' THEN p.menge END), 0),
		        coalesce(sum(CASE WHEN p.bereich = 'sideboard' THEN p.menge END), 0)
		   FROM decks d
		   LEFT JOIN deckplaetze p ON p.deck_id = d.id
		  WHERE d.profil_id = ? AND (? = '' OR d.spiel = ?)
		  GROUP BY d.id
		  ORDER BY d.spiel, d.name`, profilID, spiel, spiel)
	if err != nil {
		return nil, uebersetze(err, "die Decks lesen")
	}

	defer zeilen.Close() //nolint:errcheck // der Fehler steht in zeilen.Err()

	decks := []domain.Deck{}

	for zeilen.Next() {
		var (
			deck   domain.Deck
			format sql.NullInt64
		)

		if err := zeilen.Scan(&deck.ID, &deck.Fassung, &deck.Name,
			&deck.Spiel, &deck.Notiz, &format,
			&deck.Hauptdeck, &deck.Sideboard); err != nil {
			return nil, uebersetze(err, "ein Deck lesen")
		}

		deck.FormatID = format.Int64
		decks = append(decks, deck)
	}

	return decks, uebersetze(zeilen.Err(), "die Decks lesen")
}

// PlatzSetzen legt einen Platz an oder schreibt seine Menge fort.
//
// **Derselbe Bezug im selben Bereich ist eine höhere Menge, kein zweiter
// Platz.** Wer dieselbe Karte zweimal hineinlegt, meint vier Wald und nicht
// zwei Zeilen über dasselbe.
func (a *Deckablage) PlatzSetzen(
	ctx context.Context, deck int64, platz domain.Deckplatz,
) (domain.Deckplatz, error) {
	// **In einem Zug** — Befund 4 des Zweitblicks über Phase 4.6.
	//
	// Suchen und Schreiben standen ohne Transaktion nebeneinander. Der Befund
	// las daraus ein Zeitfenster: zwei gleichzeitige Rufe lesen beide „nicht
	// vorhanden", und der zweite läuft gegen den Eindeutigkeitsindex.
	//
	// **Der Wortlaut stimmt, die Wirkung nicht.** Gemessen am 2026-09-01:
	// `Oeffnen` setzt `db.SetMaxOpenConns(1)` (`oeffnenMit` in `datenbank.go`,
	// seit #1485 für jede Datenbank) —
	// mit **einer** Verbindung laufen alle Rufe serialisiert, und das Fenster
	// entsteht nicht. Ein Lauf mit acht Goroutinen bleibt grün, auch ohne
	// diese Transaktion; er kann seinen Fall nicht enthalten.
	//
	// **Sie steht trotzdem hier, weil sie die Zusage von der Verbindungszahl
	// löst.** Wer sie je erhöht, hätte den Befund sonst geerbt, ohne ihn zu
	// sehen.
	transaktion, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Deckplatz{}, uebersetze(err, "den Deckplatz beginnen")
	}

	defer func() { _ = transaktion.Rollback() }()

	// **Nur in ein Deck des gezeigten Profils, und nur mit einer Karte aus
	// ihm** (#1815). Die Deckübernahme lässt einen Platz, der auf eine eigene
	// Karte des anderen Profils zeigt, deshalb schon drüben — hier wird es
	// zugesagt statt vorausgesetzt.
	if fehler := deckImProfil(ctx, transaktion, deck); fehler != nil {
		return domain.Deckplatz{}, fehler
	}

	if platz.KarteID != 0 {
		if fehler := karteImProfil(ctx, transaktion, platz.KarteID); fehler != nil {
			return domain.Deckplatz{}, fehler
		}
	}

	gesetzt, fehler := platzImZug(ctx, transaktion, deck, platz)
	if fehler != nil {
		return domain.Deckplatz{}, fehler
	}

	// **Ein Satz und kein Zweig** — dieselbe Form wie in
	// `autoersetzungen.go`: ein eigener Fehlerzweig fürs Abschließen wäre
	// einer, den kein Lauf betreten kann (§3).
	return gesetzt, uebersetze(transaktion.Commit(), "den Deckplatz abschließen")
}

// platzImZug sucht und schreibt innerhalb einer Transaktion.
//
// **Erst suchen, dann setzen.** Ein `INSERT … ON CONFLICT` bräuchte zwei
// Fassungen — eine je Spalte —, und zwei Fassungen derselben Regel laufen
// auseinander.
func platzImZug(
	ctx context.Context, zug *sql.Tx, deck int64, platz domain.Deckplatz,
) (domain.Deckplatz, error) {
	// **Drei Bezüge, eine Spalte je Bezug** (A2, #1441): die eigene Karte,
	// der Alias — oder die Zeile eines Drucks, der keinen Alias hat.
	spalte, wert := "printing", any(platz.Printing)

	switch {
	case platz.KarteID != 0:
		spalte, wert = "karte_id", any(platz.KarteID)
	case platz.Printing == "":
		spalte, wert = "druck_id", any(platz.DruckID)
	}

	var vorhanden int64

	err := zug.QueryRowContext(ctx,
		`SELECT id FROM deckplaetze
		  WHERE deck_id = ? AND bereich = ? AND `+spalte+` = ?`,
		deck, platz.Bereich, wert).Scan(&vorhanden)

	switch {
	case err == nil:
		_, schreibfehler := zug.ExecContext(ctx,
			`UPDATE deckplaetze SET menge = ? WHERE id = ?`, platz.Menge, vorhanden)

		platz.ID = vorhanden

		return platz, uebersetze(schreibfehler, "den Deckplatz fortschreiben")

	case !errors.Is(err, sql.ErrNoRows):
		return domain.Deckplatz{}, uebersetze(err, "den Deckplatz suchen")
	}

	return platzAnlegenImZug(ctx, zug, deck, platz)
}

// platzAnlegenImZug legt einen neuen Platz ans Ende.
func platzAnlegenImZug(
	ctx context.Context, zug *sql.Tx, deck int64, platz domain.Deckplatz,
) (domain.Deckplatz, error) {
	var (
		karteID  = sql.NullInt64{Int64: platz.KarteID, Valid: platz.KarteID != 0}
		printing = sql.NullString{String: platz.Printing, Valid: platz.Printing != ""}
	)

	druck, druckfehler := druckZeile(ctx, zug, druckbezug{druckID: platz.DruckID, printing: platz.Printing})
	if druckfehler != nil {
		return domain.Deckplatz{}, druckfehler
	}

	ergebnis, err := zug.ExecContext(ctx,
		`INSERT INTO deckplaetze (deck_id, karte_id, printing, druck_id, menge, bereich, stelle)
		 VALUES (?, ?, ?, ?, ?, ?,
		         (SELECT coalesce(max(stelle), 0) + 1 FROM deckplaetze WHERE deck_id = ?))`,
		deck, karteID, printing, druck, platz.Menge, platz.Bereich, deck)
	if err != nil {
		return domain.Deckplatz{}, uebersetze(err, "den Deckplatz legen")
	}

	kennung, idfehler := ergebnis.LastInsertId()

	platz.ID = kennung

	return platz, uebersetze(idfehler, "die Kennung des Deckplatzes lesen")
}

// PlatzLoeschen nimmt einen Platz heraus.
func (a *Deckablage) PlatzLoeschen(ctx context.Context, deck, platz int64) error {
	return imGeschaeft(ctx, a.db, "den Deckplatz löschen", func(tx *sql.Tx) error {
		// **Nur aus einem Deck des gezeigten Profils** (#1815).
		if fehler := deckImProfil(ctx, tx, deck); fehler != nil {
			return fehler
		}

		ergebnis, err := tx.ExecContext(ctx,
			`DELETE FROM deckplaetze WHERE deck_id = ? AND id = ?`, deck, platz)
		if err != nil {
			return uebersetze(err, "den Deckplatz löschen")
		}

		zeilen, zaehlfehler := betroffen(ergebnis)
		if zaehlfehler != nil {
			return zaehlfehler
		}

		if zeilen == 0 {
			return fmt.Errorf("%w: diesen Deckplatz gibt es nicht", domain.ErrNotFound)
		}

		return nil
	})
}

// Loeschen nimmt ein Deck heraus — **samt seinen Plätzen**.
//
// Das erledigt `ON DELETE CASCADE`; hier steht es, damit niemand es für einen
// Zufall hält.
func (a *Deckablage) Loeschen(ctx context.Context, kennung int64) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	ergebnis, err := a.db.ExecContext(ctx,
		`DELETE FROM decks WHERE id = ? AND profil_id = ?`, kennung, profilID)
	if err != nil {
		return uebersetze(err, "das Deck löschen")
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return zaehlfehler
	}

	if zeilen == 0 {
		return fmt.Errorf("%w: dieses Deck gibt es nicht", domain.ErrNotFound)
	}

	return nil
}
