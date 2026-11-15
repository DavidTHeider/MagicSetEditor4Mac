package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Kennung aus einem fremden Profil trifft nichts** —
// #1815.
//
// # Der Befund
//
// Die zehn Wurzeln tragen `profil_id`, und jede Abfrage auf sie klemmt
// darauf (0066; Wächter `TestJedeWurzelabfrageKlemmtAufEinProfil`). Die
// **Abkömmlinge** — Karten, Kartenwerte, Deckplätze, Runden — erben ihr Profil
// über den Fremdschlüssel und sind „nur so sicher wie die Abfrage, die ihre
// Kennung geliefert hat" (0035). Geliefert hat sie aber oft der **Aufrufer**:
// am 2026-09-23 schrieb ein Nutzer an der Kommandozeile mit `mse feld 1 1 name
// …` den Kartennamen eines anderen um, ohne dessen Profil zu betreten. 37 von
// 55 Funktionen der Karten- und Setablagen ließen es zu.
//
// # Die Regel
//
// **Wer eine Kennung eines Abkömmlings annimmt, fragt zuerst, ob seine Wurzel
// dem gezeigten Profil gehört** — in derselben Transaktion, wenn er schreibt.
// Sonst ist die Antwort `ErrNotFound`, dieselbe wie für eine Kennung, die es
// nicht gibt: sie verrät nicht, dass es sie woanders gibt.
//
// **In der Ablage und nicht an den Flächen**: vier Flächen rufen dieselben
// Funktionen, und eine Prüfung an jeder wäre vier Stellen, an denen sie
// fehlen kann. Ein Wächter (`besitzwaechter_test.go` in `internal/werkzeuge`)
// hält, dass jede Methode einer Abkömmlingsablage einen dieser Helfer ruft.

// setImProfil weist eine Setkennung ab, die nicht zum gezeigten Profil gehört.
func setImProfil(ctx context.Context, an abfrager, setID int64) error {
	return besitzt(ctx, an, `SELECT 1 FROM sets WHERE id = ? AND profil_id = ?`,
		fmt.Errorf("%w: das Set %d gibt es nicht", domain.ErrNotFound, setID), setID)
}

// karteImProfil weist eine Karte ab, deren Set nicht zum gezeigten Profil
// gehört.
func karteImProfil(ctx context.Context, an abfrager, karteID int64) error {
	return besitzt(ctx, an,
		`SELECT 1 FROM cards JOIN sets ON sets.id = cards.set_id
		  WHERE cards.id = ? AND sets.profil_id = ?`,
		fmt.Errorf("%w: die Karte %d gibt es nicht", domain.ErrNotFound, karteID), karteID)
}

// kartenImProfil weist eine Reihe von Karten ab, von denen eine nicht zum
// gezeigten Profil gehört — **alle oder keine**, wie die Schreibwege, die sie
// schützt.
func kartenImProfil(ctx context.Context, an abfrager, karteIDs []int64) error {
	for _, karteID := range karteIDs {
		if fehler := karteImProfil(ctx, an, karteID); fehler != nil {
			return fehler
		}
	}

	return nil
}

// deckImProfil weist eine Deckkennung ab, die nicht zum gezeigten Profil
// gehört.
func deckImProfil(ctx context.Context, an abfrager, deckID int64) error {
	return besitzt(ctx, an, `SELECT 1 FROM decks WHERE id = ? AND profil_id = ?`,
		fmt.Errorf("%w: das Deck %d gibt es nicht", domain.ErrNotFound, deckID), deckID)
}

// platzImDeck weist einen Deckplatz ab, der nicht in diesem Deck liegt —
// oder dessen Deck nicht zum gezeigten Profil gehört.
//
// **Beides in einer Frage**: ein Platz aus einem fremden Deck ist für eine
// Zuordnung in ein eigenes genauso falsch wie einer aus einem fremden Profil,
// und der Bestand hält keine der beiden Bedingungen selbst (0052).
func platzImDeck(ctx context.Context, an abfrager, deckID, platzID int64) error {
	return besitzt(ctx, an,
		`SELECT 1 FROM deckplaetze JOIN decks ON decks.id = deckplaetze.deck_id
		  WHERE deckplaetze.id = ? AND decks.id = ? AND decks.profil_id = ?`,
		fmt.Errorf("%w: den Platz %d gibt es in Deck %d nicht", domain.ErrNotFound, platzID, deckID),
		platzID, deckID)
}

// exemplarImProfil weist ein Exemplar ab, das nicht zum gezeigten Profil
// gehört.
func exemplarImProfil(ctx context.Context, an abfrager, exemplarID int64) error {
	return besitzt(ctx, an, `SELECT 1 FROM sammlung WHERE id = ? AND profil_id = ?`,
		fmt.Errorf("%w: das Exemplar %d gibt es nicht", domain.ErrNotFound, exemplarID), exemplarID)
}

// turnierImProfil weist eine Turnierkennung ab, die nicht zum gezeigten
// Profil gehört.
func turnierImProfil(ctx context.Context, an abfrager, turnierID int64) error {
	return besitzt(ctx, an, `SELECT 1 FROM turniere WHERE id = ? AND profil_id = ?`,
		fmt.Errorf("%w: das Turnier %d gibt es nicht", domain.ErrNotFound, turnierID), turnierID)
}

// teilnehmerImProfil weist einen Teilnehmer ab, dessen Turnier nicht zum
// gezeigten Profil gehört.
func teilnehmerImProfil(ctx context.Context, an abfrager, teilnehmerID int64) error {
	return besitzt(ctx, an,
		`SELECT 1 FROM turnierteilnehmer AS t JOIN turniere ON turniere.id = t.turnier_id
		  WHERE t.id = ? AND turniere.profil_id = ?`,
		fmt.Errorf("%w: den Teilnehmer %d gibt es nicht", domain.ErrNotFound, teilnehmerID), teilnehmerID)
}

// teilnehmerImTurnier weist einen Teilnehmer ab, der nicht in diesem Turnier
// steht — oder dessen Turnier nicht zum gezeigten Profil gehört. Dieselbe
// Gestalt wie [platzImDeck], und aus demselben Grund: der Fremdschlüssel sagt
// nur, dass es ihn gibt, nicht wo.
func teilnehmerImTurnier(ctx context.Context, an abfrager, turnierID, teilnehmerID int64) error {
	return besitzt(ctx, an,
		`SELECT 1 FROM turnierteilnehmer AS t JOIN turniere ON turniere.id = t.turnier_id
		  WHERE t.id = ? AND turniere.id = ? AND turniere.profil_id = ?`,
		fmt.Errorf("%w: den Teilnehmer %d gibt es in Turnier %d nicht",
			domain.ErrNotFound, teilnehmerID, turnierID),
		teilnehmerID, turnierID)
}

// paarungImTurnier weist eine Paarung ab, die nicht in diesem Turnier liegt —
// oder deren Turnier nicht zum gezeigten Profil gehört.
func paarungImTurnier(ctx context.Context, an abfrager, turnierID, paarungID int64) error {
	return besitzt(ctx, an,
		`SELECT 1 FROM turnierpaarungen AS p
		   JOIN turnierrunden AS r ON r.id = p.runde_id
		   JOIN turniere ON turniere.id = r.turnier_id
		  WHERE p.id = ? AND turniere.id = ? AND turniere.profil_id = ?`,
		fmt.Errorf("%w: die Paarung %d gibt es in Turnier %d nicht",
			domain.ErrNotFound, paarungID, turnierID),
		paarungID, turnierID)
}

// paarungImProfil weist eine Paarung ab, deren Turnier nicht zum gezeigten
// Profil gehört.
func paarungImProfil(ctx context.Context, an abfrager, paarungID int64) error {
	return besitzt(ctx, an,
		`SELECT 1 FROM turnierpaarungen AS p
		   JOIN turnierrunden AS r ON r.id = p.runde_id
		   JOIN turniere ON turniere.id = r.turnier_id
		  WHERE p.id = ? AND turniere.profil_id = ?`,
		fmt.Errorf("%w: die Paarung %d gibt es nicht", domain.ErrNotFound, paarungID), paarungID)
}

// partieImProfil weist eine Partie ab, die nicht zum gezeigten Profil gehört.
func partieImProfil(ctx context.Context, an abfrager, partieID string) error {
	return besitzt(ctx, an, `SELECT 1 FROM partien WHERE id = ? AND profil_id = ?`,
		fmt.Errorf("%w: die Partie %q gibt es nicht", domain.ErrNotFound, partieID), partieID)
}

// besitzt fragt Kennungen gegen das Profil aus dem Kontext.
//
// **Die Abfrage endet mit der Bedingung auf das Profil**, und das Profil ist
// der letzte Wert — so steht es in jeder Frage darüber, und so erkennt sie
// der Stolpertreiber (`stolperbesitz_test.go`).
func besitzt(ctx context.Context, an abfrager, abfrage string, fehlt error, kennungen ...any) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	var eins int

	fehler = an.QueryRowContext(ctx, abfrage, append(kennungen, profilID)...).Scan(&eins)
	if errors.Is(fehler, sql.ErrNoRows) {
		return fehlt
	}

	return uebersetze(fehler, "den Besitz von %v nachschlagen", kennungen)
}
