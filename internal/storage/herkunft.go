// **Woher ein Deck und woher ein Exemplar kam** — Werkstück 12 von Phase 4.12
// (#1351),
// nach Vermerk [0071](../../docs/entscheidungen/0071-woher-ein-deck-kommt.md).

package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Herkunftsablage führt beide Herkunftstabellen.
//
// **Eine Ablage für beide und nicht zwei.** Es ist dieselbe Frage an zwei
// Datensätze; zwei Ablagen hätten zweimal dieselbe Übersetzung, und die
// zweite liefe irgendwann anders.
type Herkunftsablage struct{ db *sql.DB }

// NeueHerkunftsablage baut sie über einer Datenquelle.
func NeueHerkunftsablage(db *sql.DB) *Herkunftsablage {
	return &Herkunftsablage{db: db}
}

// Deckherkunft gibt die Herkunft eines Decks.
//
// **Keine ist kein Fehler**, sondern die Auskunft „hier entstanden". Ein
// [domain.ErrNotFound] zwänge jeden Aufrufer, ihn abzufangen und zu
// übersetzen — und wer es einmal vergäße, machte aus einem eigenen Deck einen
// Fehler.
func (a *Herkunftsablage) Deckherkunft(ctx context.Context, deckID int64) (
	domain.Herkunft, error,
) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := deckImProfil(ctx, a.db, deckID); fehler != nil {
		return domain.Herkunft{}, fehler
	}

	return a.eine(ctx,
		`SELECT art, quelle, quell_profil, uebernommen
		 FROM deck_herkunft WHERE deck_id = ?`, deckID)
}

// Exemplarherkunft gibt die Herkunft eines Exemplars — dieselbe Zusage.
func (a *Herkunftsablage) Exemplarherkunft(ctx context.Context, exemplarID int64) (
	domain.Herkunft, error,
) {
	if fehler := exemplarImProfil(ctx, a.db, exemplarID); fehler != nil {
		return domain.Herkunft{}, fehler
	}

	return a.eine(ctx,
		`SELECT art, quelle, quell_profil, uebernommen
		 FROM sammlung_herkunft WHERE exemplar_id = ?`, exemplarID)
}

// eine liest eine Herkunft, wenn es sie gibt.
func (a *Herkunftsablage) eine(ctx context.Context, abfrage string, kennung int64) (
	domain.Herkunft, error,
) {
	var (
		gelesen domain.Herkunft
		art     string
	)

	err := a.db.QueryRowContext(ctx, abfrage, kennung).Scan(
		&art, &gelesen.Quelle, &gelesen.QuellProfil, &gelesen.Uebernommen)

	if errors.Is(err, sql.ErrNoRows) {
		return domain.Herkunft{}, nil
	}

	if err != nil {
		return domain.Herkunft{}, uebersetze(err, "die Herkunft lesen")
	}

	gelesen.Art = domain.Herkunftsart(art)

	return gelesen, nil
}

// DeckherkunftSetzen schreibt sie — und schreibt eine vorhandene fort.
//
// **`quellDeck` steht neben der Herkunft und nicht darin**: [domain.Herkunft]
// gilt für ein Deck **und** für ein Exemplar, und ein Exemplar kommt aus
// keinem Deck. Ein Feld, das an der Hälfte der Fälle leer bleibt, wäre eine
// Einladung, es dort zu füllen.
func (a *Herkunftsablage) DeckherkunftSetzen(
	ctx context.Context, deckID int64, herkunft domain.Herkunft, quellDeck int64,
) error {
	if fehler := herkunft.Validate(); fehler != nil {
		return fehler
	}

	return imGeschaeft(ctx, a.db, "die Herkunft des Decks setzen", func(tx *sql.Tx) error {
		// **Nur an einem Deck des gezeigten Profils** (#1815).
		if fehler := deckImProfil(ctx, tx, deckID); fehler != nil {
			return fehler
		}

		return deckherkunftImZug(ctx, tx, deckID, herkunft, quellDeck)
	})
}

// deckherkunftImZug schreibt die Herkunft eines Decks in einer laufenden
// Transaktion — für `DeckherkunftSetzen` und für die Übernahme (#1820).
func deckherkunftImZug(
	ctx context.Context, tx *sql.Tx, deckID int64, herkunft domain.Herkunft, quellDeck int64,
) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO deck_herkunft
		   (deck_id, art, quelle, quell_profil, quell_deck, uebernommen)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT (deck_id) DO UPDATE SET
		   art = excluded.art, quelle = excluded.quelle,
		   quell_profil = excluded.quell_profil,
		   quell_deck = excluded.quell_deck,
		   uebernommen = excluded.uebernommen`,
		deckID, string(herkunft.Art), herkunft.Quelle, herkunft.QuellProfil,
		quellDeck, jetzt())

	return uebersetze(err, "die Herkunft des Decks setzen")
}

// ExemplarherkunftSetzen tut dasselbe für ein Exemplar.
func (a *Herkunftsablage) ExemplarherkunftSetzen(
	ctx context.Context, exemplarID int64, herkunft domain.Herkunft,
) error {
	if fehler := herkunft.Validate(); fehler != nil {
		return fehler
	}

	return imGeschaeft(ctx, a.db, "die Herkunft des Exemplars setzen", func(tx *sql.Tx) error {
		if fehler := exemplarImProfil(ctx, tx, exemplarID); fehler != nil {
			return fehler
		}

		_, err := tx.ExecContext(ctx,
			`INSERT INTO sammlung_herkunft
			   (exemplar_id, art, quelle, quell_profil, uebernommen)
			 VALUES (?, ?, ?, ?, ?)
			 ON CONFLICT (exemplar_id) DO UPDATE SET
			   art = excluded.art, quelle = excluded.quelle,
			   quell_profil = excluded.quell_profil,
			   uebernommen = excluded.uebernommen`,
			exemplarID, string(herkunft.Art), herkunft.Quelle, herkunft.QuellProfil,
			jetzt())

		return uebersetze(err, "die Herkunft des Exemplars setzen")
	})
}

// jetzt gibt den Zeitpunkt, den eine Übernahme trägt.
//
// **Gesetzt und nicht mitgegeben**: er sagt, wann **dieser** Bestand die
// Übernahme gesehen hat. Ein Zeitpunkt vom Aufrufer wäre eine Behauptung des
// Gebenden, und die prüft niemand nach (Vermerk 0071).
func jetzt() string { return time.Now().UTC().Format(time.RFC3339) }

// Deckentnahme liest ein Deck aus einem **anderen** Profil.
//
// # Die einzige Abfrage, die die Klemme aus Werkstück 2 absichtlich überschreitet
//
// **Und deshalb steht sie hier und nicht in `decks.go`.** Sie nennt das Profil
// ausdrücklich in der Signatur, statt es aus dem Kontext zu nehmen — wer den
// Namen liest, sieht sofort, dass hier ein fremdes gemeint ist.
//
// **Sie prüft die Erlaubnis nicht.** Das ist Sache von
// `service.Deckuebernahme`, ihres einzigen Aufrufers: eine Ablage entscheidet
// nicht, wer etwas darf (§2). Was sie leistet, ist, dass niemand aus Versehen
// über die Grenze liest — dafür müsste er diesen Namen tippen.
//
// **Ein eigener Empfänger und nicht `*Deckablage`**: eine Ablage, die diesen
// Ruf mitbrächte, hätte ihn an jeder Stelle, an der sie schon steht.
type Deckentnahme struct{ ablage *Deckablage }

// NeueDeckentnahme baut sie über einer Datenquelle.
func NeueDeckentnahme(db *sql.DB) *Deckentnahme {
	return &Deckentnahme{ablage: NeueDeckablage(db)}
}

// DeckAusProfil liest ein Deck, als stünde man in dem genannten Profil.
func (e *Deckentnahme) DeckAusProfil(
	ctx context.Context, profilID, deckID int64,
) (domain.Deck, error) {
	return e.ablage.Deck(MitProfil(ctx, profilID), deckID)
}

// DeckeAusProfil nennt die Decks des genannten Profils — ohne ihre Plätze.
//
// **Zur Wahl bei der Übernahme** (UX-08.AK-04, #1680): wer ein Deck holen
// will, soll es aus einer Liste wählen, statt seine Kennung zu tippen. Die
// Erlaubnis prüft auch hier `service.Deckuebernahme`, **bevor** gelesen wird.
func (e *Deckentnahme) DeckeAusProfil(
	ctx context.Context, profilID int64,
) ([]domain.Deck, error) {
	return e.ablage.Decks(MitProfil(ctx, profilID), "")
}
