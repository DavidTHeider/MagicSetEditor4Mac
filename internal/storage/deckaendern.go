package storage

// **Ein Deck ändern** — Werkstück 3 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei, weil `decks.go` an der Zeilengrenze aus §2 steht.** Der
// Schnitt folgt der Frage: dort steht, wie ein Deck entsteht und was darin
// liegt; hier, wie es sich unter fremder Hand ändert.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Aendern schreibt Name, Notiz und das gewählte Format eines Decks fort.
//
// **Nicht das Spiel.** Es zu wechseln hieße, jeden Platz darin zu prüfen — und
// das ist eine andere Frage mit einer anderen Antwort.
//
// **Das Format geht mit**, weil es zum Datensatz gehört und nicht zu den
// Plätzen. Null heißt „keines" und nicht „behalte das alte": wer ein Deck
// liest, ändert und zurückschreibt, trägt den Wert ohnehin mit sich.
//
// **Die Fassung ist die Bedingung** (§7, Regel 2), wie am Exemplar. Sie steigt
// mit dem Schreiben, und der Rückgabewert trägt die neue: wer sie nicht
// zurückbekäme, schriebe beim nächsten Mal wieder gegen die alte.
//
// **Sie deckt den Datensatz und nicht die Plätze** — warum, steht an
// [domain.Deck.Fassung].
func (a *Deckablage) Aendern(ctx context.Context, deck domain.Deck) (
	domain.Deck, error,
) {
	deck.Normalisieren()

	// **Jede Wurzelabfrage klemmt auf ein Profil** (#1351, Vermerk 0066).
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Deck{}, profilfehler
	}

	ergebnis, err := a.db.ExecContext(ctx,
		`UPDATE decks
		    SET name = ?, notiz = ?, format_id = ?, fassung = fassung + 1,
		        geaendert = datetime('now')
		  WHERE id = ? AND fassung = ? AND profil_id = ?`,
		deck.Name, deck.Notiz, formatspalte(deck.FormatID),
		deck.ID, deck.Fassung, profilID)
	if err != nil {
		return domain.Deck{}, uebersetze(err, "das Deck ändern")
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return domain.Deck{}, zaehlfehler
	}

	if zeilen == 0 {
		return domain.Deck{}, a.warumNicht(ctx, deck.ID, profilID)
	}

	deck.Fassung++

	return deck, nil
}

// warumNicht sagt, ob das Deck fehlt oder seine Fassung veraltet ist.
//
// **Zwei Antworten auf denselben Fehlschlag** — dieselbe Form wie in
// `sammlung.go`, und aus demselben Grund: „gibt es nicht" heißt aufräumen,
// „jemand anderes hat geschrieben" heißt nachladen (§8).
//
// **Das Profil kommt vom Rufer.** Es hier ein zweites Mal aus dem Kontext zu
// holen wäre ein Zweig, den kein Aufruf betreten kann: der Rufer hat es schon
// gelesen und wäre sonst gar nicht bis hierher gekommen (§3).
func (a *Deckablage) warumNicht(
	ctx context.Context, kennung, profilID int64,
) error {
	var vorhanden int64

	err := a.db.QueryRowContext(ctx,
		`SELECT count(*) FROM decks WHERE id = ? AND profil_id = ?`,
		kennung, profilID).Scan(&vorhanden)
	if err != nil {
		return uebersetze(err, "das Deck nachschlagen")
	}

	if vorhanden == 0 {
		return fmt.Errorf("%w: das Deck %d gibt es nicht", domain.ErrNotFound, kennung)
	}

	return fmt.Errorf(
		"%w — das Deck %d hat sich seit dem Lesen geändert", domain.ErrUeberholt, kennung)
}

// SpielZurKarte gibt das Spiel des Sets, in dem eine Karte liegt.
//
// **Am [SetLeser] und nicht an der Deckablage**: es ist eine Frage über ein
// Set, erreicht über eine Karte. Ein Speicher, der die Tabellen eines anderen
// liest, wäre eine zweite Verbindung auf dasselbe.
//
// **Der Deckbau braucht sie**, um eine eigene Karte abzuweisen, die zu einem
// anderen Spiel gehört als das Deck.
func (l *SetLeser) SpielZurKarte(ctx context.Context, karteID int64) (string, error) {
	var spiel string

	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return "", profilfehler
	}

	err := l.db.QueryRowContext(ctx,
		`SELECT s.spiel FROM cards c JOIN sets s ON s.id = c.set_id
		  WHERE c.id = ? AND s.profil_id = ?`,
		karteID, profilID).Scan(&spiel)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", fmt.Errorf(
			"%w: die Karte %d führt der Bestand nicht", domain.ErrNotFound, karteID)

	case err != nil:
		return "", uebersetze(err, "das Spiel zur Karte %d lesen", karteID)
	}

	return spiel, nil
}

// NameZurKarte gibt den Namen einer eigenen Karte — leer, wenn sie keinen hat.
//
// **Am [SetLeser] und aus demselben Grund wie [SetLeser.SpielZurKarte]**: eine
// Frage über eine Karte, gestellt dort, wo Karten liegen.
//
// **Die Turnierleitung braucht ihn**, weil der Abzug einer Deckliste Text ist
// und kein Verweis (Vermerk 0062, Entscheidung 3). Ohne den Namen bliebe von
// einer eigenen Karte nichts übrig, sobald jemand sie löscht.
//
// **Ein leerer Name ist kein Fehler.** Eine frisch angelegte Karte hat keinen,
// und ein Turnier abzuweisen, weil eine Karte namenlos ist, wäre die falsche
// Strenge — die Fläche zeigt dann die Kennung.
func (l *SetLeser) NameZurKarte(ctx context.Context, karteID int64) (string, error) {
	// **Nur aus dem gezeigten Profil, und in derselben Abfrage** (#1815): eine
	// Karte eines anderen fehlt hier wie eine, die es nicht gibt.
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return "", profilfehler
	}

	var name string

	err := l.db.QueryRowContext(ctx,
		`SELECT coalesce((SELECT wert FROM card_values
		                   WHERE card_values.card_id = cards.id
		                     AND card_values.feld = 'name'), '')
		   FROM cards JOIN sets ON sets.id = cards.set_id
		  WHERE cards.id = ? AND sets.profil_id = ?`, karteID, profilID).Scan(&name)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", fmt.Errorf(
			"%w: die Karte %d führt der Bestand nicht", domain.ErrNotFound, karteID)

	case err != nil:
		return "", uebersetze(err, "den Namen zur Karte %d lesen", karteID)
	}

	return name, nil
}

// FeldZurKarte gibt einen beliebigen Feldwert einer eigenen Karte —
// #1376.
//
// **Dieselbe Frage wie [SetLeser.NameZurKarte], nur mit dem Feld als
// Argument.** Der Deckbau braucht die Manakosten für die Kurve; eine dritte
// Funktion je Feld liefe mit den beiden anderen auseinander.
//
// **Ein leerer Wert ist kein Fehler.** Ein Feld, das eine Karte nicht führt,
// gibt die leere Zeichenkette — genau wie ein Feld, das leer ausgefüllt ist.
// Die beiden zu unterscheiden bräuchte eine zweite Abfrage für einen
// Unterschied, den keine Fläche zeigt.
func (l *SetLeser) FeldZurKarte(
	ctx context.Context, karteID int64, feld string,
) (string, error) {
	// **Nur aus dem gezeigten Profil, und in derselben Abfrage** (#1815).
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return "", profilfehler
	}

	var wert string

	err := l.db.QueryRowContext(ctx,
		`SELECT coalesce((SELECT wert FROM card_values
		                   WHERE card_values.card_id = cards.id
		                     AND card_values.feld = ?), '')
		   FROM cards JOIN sets ON sets.id = cards.set_id
		  WHERE cards.id = ? AND sets.profil_id = ?`, feld, karteID, profilID).Scan(&wert)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", fmt.Errorf(
			"%w: die Karte %d führt der Bestand nicht", domain.ErrNotFound, karteID)

	case err != nil:
		return "", uebersetze(err, "das Feld %q zur Karte %d lesen", feld, karteID)
	}

	return wert, nil
}

// formatspalte macht aus einer Kennung ein `NULL`, wenn keines gewählt ist.
//
// **Null ist kein Format**, und ein `format_id = 0` fände keinen
// Fremdschlüssel — der Bestand wiese die Zeile ab.
func formatspalte(kennung int64) sql.NullInt64 {
	return sql.NullInt64{Int64: kennung, Valid: kennung != 0}
}
