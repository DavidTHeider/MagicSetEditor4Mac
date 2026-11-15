package storage

// **Einen Deckplatz in den anderen Bereich verschieben — in einem Zug.**
//
// Befund 7 des Zweitblicks über Phase 5.1
// (#1405).
//
// # Warum das nicht im Fenster geht
//
// Das Fenster hat den Wechsel bis zum 2026-09-10 aus zwei Rufen gebaut: erst
// `DeckplatzLegen` in den Zielbereich, dann `DeckplatzNehmen` für den alten.
// **Scheitert der zweite, liegt dieselbe Karte in beiden Bereichen** — ein
// halber Stand, und §7 Regel 3 sagt, dass nach einem gescheiterten Schreiben
// nichts zurückgenommen wird. Über eine Naht mit zwei Rufen lässt sich das
// nicht heilen; ein Vorgang muss **einer** sein.
//
// # Eine eigene Datei
//
// Weil `decks.go` mit 384 Zeilen an der Grenze aus §2 steht. Der Schnitt
// folgt der Frage: dort steht, was ein Deck **ist** und was in ihm liegt,
// hier, wie ein Platz von einem Bereich in den anderen kommt.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// PlatzVerschieben legt einen Platz in einen anderen Bereich.
//
// **Es ist kein Umschreiben einer Spalte**, und das ist der Grund für die
// Transaktion: liegt im Zielbereich schon derselbe Bezug, hält der Bestand
// dort **einen** Platz — die Mengen gehen zusammen, und der alte fällt weg.
// `platzImZug` kann beides, und es ist dieselbe Regel wie beim Legen.
func (a *Deckablage) PlatzVerschieben(
	ctx context.Context, deck, platz int64, bereich domain.Deckbereich,
) (domain.Deckplatz, error) {
	if !domain.IstDeckbereich(bereich) {
		return domain.Deckplatz{}, domain.NewFieldError("bereich",
			domain.SchluesselValidierung,
			"Ein Deckplatz liegt im Hauptdeck oder im Sideboard, nicht in %q.", string(bereich))
	}

	transaktion, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Deckplatz{}, uebersetze(err, "den Bereichswechsel beginnen")
	}

	defer func() { _ = transaktion.Rollback() }()

	// **Nur in einem Deck des gezeigten Profils** (#1815).
	if fehler := deckImProfil(ctx, transaktion, deck); fehler != nil {
		return domain.Deckplatz{}, fehler
	}

	verschoben, fehler := verschiebenImZug(ctx, transaktion, deck, platz, bereich)
	if fehler != nil {
		return domain.Deckplatz{}, fehler
	}

	return verschoben, uebersetze(transaktion.Commit(), "den Bereichswechsel abschließen")
}

// verschiebenImZug liest den Platz, setzt ihn drüben und räumt hier auf.
func verschiebenImZug(
	ctx context.Context, zug *sql.Tx, deck, platz int64, bereich domain.Deckbereich,
) (domain.Deckplatz, error) {
	alt, lesefehler := platzLesenImZug(ctx, zug, deck, platz)
	if lesefehler != nil {
		return domain.Deckplatz{}, lesefehler
	}

	// **Derselbe Bereich ist keine Arbeit und kein Fehler.** Wer ihn als
	// Fehler führte, machte aus einem Doppelklick eine Meldung.
	if alt.Bereich == bereich {
		return alt, nil
	}

	ziel := alt
	ziel.ID = 0
	ziel.Bereich = bereich

	// **Die Mengen gehen zusammen, sie ersetzen einander nicht.** `platzImZug`
	// **setzt** die Menge — das ist beim Legen richtig („vier davon") und beim
	// Verschieben falsch: wer vier ins Sideboard legt, wo zwei liegen, hat
	// sechs und nicht vier. Der rote Lauf dazu stand vor dieser Zeile.
	dortige, zaehlfehler := mengeImZug(ctx, zug, deck, bereich, alt)
	if zaehlfehler != nil {
		return domain.Deckplatz{}, zaehlfehler
	}

	ziel.Menge = alt.Menge + dortige

	// **Erst wegnehmen, dann setzen** — und zwar innerhalb des Zuges. Setzte
	// man zuerst, fände `platzImZug` den alten Platz nicht, wohl aber der
	// Eindeutigkeitsindex über (Deck, Bereich, Bezug), sobald Quelle und Ziel
	// je eine Zeile hätten.
	if _, err := zug.ExecContext(ctx,
		`DELETE FROM deckplaetze WHERE deck_id = ? AND id = ?`, deck, platz); err != nil {
		return domain.Deckplatz{}, uebersetze(err, "den alten Deckplatz löschen")
	}

	gesetzt, setzfehler := platzImZug(ctx, zug, deck, ziel)
	if setzfehler != nil {
		return domain.Deckplatz{}, setzfehler
	}

	return gesetzt, nil
}

// platzLesenImZug liest einen Platz samt seinem Bezug.
func platzLesenImZug(
	ctx context.Context, zug *sql.Tx, deck, platz int64,
) (domain.Deckplatz, error) {
	var (
		gelesen  domain.Deckplatz
		karteID  sql.NullInt64
		printing sql.NullString
		druckID  sql.NullInt64
	)

	err := zug.QueryRowContext(ctx,
		`SELECT id, karte_id, printing, druck_id, menge, bereich
		   FROM deckplaetze WHERE deck_id = ? AND id = ?`, deck, platz).
		Scan(&gelesen.ID, &karteID, &printing, &druckID, &gelesen.Menge, &gelesen.Bereich)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.Deckplatz{}, fmt.Errorf(
			"%w: diesen Deckplatz gibt es nicht", domain.ErrNotFound)
	case err != nil:
		return domain.Deckplatz{}, uebersetze(err, "den Deckplatz lesen")
	}

	gelesen.KarteID = karteID.Int64
	gelesen.Printing = printing.String
	gelesen.DruckID = druckID.Int64

	return gelesen, nil
}

// mengeImZug sagt, wie viele desselben Bezugs im Zielbereich schon liegen.
//
// **Null heißt: keiner.** Ein fehlender Platz ist kein Fehler — er ist die
// Antwort.
func mengeImZug(
	ctx context.Context, zug *sql.Tx, deck int64, bereich domain.Deckbereich,
	bezug domain.Deckplatz,
) (int, error) {
	spalte, wert := "printing", any(bezug.Printing)

	switch {
	case bezug.KarteID != 0:
		spalte, wert = "karte_id", any(bezug.KarteID)
	case bezug.Printing == "":
		spalte, wert = "druck_id", any(bezug.DruckID)
	}

	var menge int

	err := zug.QueryRowContext(ctx,
		`SELECT menge FROM deckplaetze
		  WHERE deck_id = ? AND bereich = ? AND `+spalte+` = ?`,
		deck, bereich, wert).Scan(&menge)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return 0, nil
	case err != nil:
		return 0, uebersetze(err, "die Menge im Zielbereich lesen")
	}

	return menge, nil
}
