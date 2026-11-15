package storage

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Uebernehmen legt ein Deck **samt Plätzen und Herkunft in einem Zug** an —
// #1820,
// Befund 6 des Zweitblicks über Phase 6.1.
//
// **Ein Anwendungsfall, eine Transaktion.** Bis hierher legte die Übernahme
// das Deck an und schrieb danach einzeln Plätze und Herkunft; scheiterte ein
// späterer Schritt, blieb ein halbes Deck stehen, und die Fläche meldete
// einen Fehler über einem Deck, das es gab. AGENTS.md §7 Regel 3 meint den
// **geglückten** Schreibweg — ein gescheiterter hinterlässt nichts.
//
// **Welche Plätze mitkommen, entscheidet der Dienst** (§2): ein Platz, der auf
// eine eigene Karte des anderen Profils zeigt, bleibt dort und ist ein Befund,
// kein Fehler. Hierher kommt nur, was geschrieben werden soll.
func (a *Deckablage) Uebernehmen(
	ctx context.Context, deck domain.Deck, plaetze []domain.Deckplatz,
	herkunft domain.Herkunft, quellDeck int64,
) (domain.Deck, error) {
	if fehler := herkunft.Validate(); fehler != nil {
		return domain.Deck{}, fehler
	}

	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Deck{}, profilfehler
	}

	var angelegt domain.Deck

	fehler := imGeschaeft(ctx, a.db, "das Deck übernehmen", func(zug *sql.Tx) error {
		var anlagefehler error

		angelegt, anlagefehler = deckEinfuegen(ctx, zug, deck, profilID)
		if anlagefehler != nil {
			return anlagefehler
		}

		for _, platz := range plaetze {
			// **Nur mit einer Karte aus dem gezeigten Profil** (#1815) — wie
			// bei `PlatzSetzen`; der Besitzwächter fand, dass die Übernahme
			// es sonst nur voraussetzte.
			if platz.KarteID != 0 {
				if fehler := karteImProfil(ctx, zug, platz.KarteID); fehler != nil {
					return fehler
				}
			}

			gesetzt, platzfehler := platzImZug(ctx, zug, angelegt.ID, platz)
			if platzfehler != nil {
				return platzfehler
			}

			angelegt.Plaetze = append(angelegt.Plaetze, gesetzt)
		}

		// **Die Herkunft zuletzt** — sie sagt „dieses Deck kam von dort",
		// und dafür muss es das Deck geben.
		return deckherkunftImZug(ctx, zug, angelegt.ID, herkunft, quellDeck)
	})
	if fehler != nil {
		return domain.Deck{}, fehler
	}

	// **Gezählt wie beim Lesen** — sonst kam die Antwort mit null Karten.
	angelegt.Zaehlen()

	return angelegt, nil
}

// deckEinfuegen schreibt die Zeile eines Decks — für `Anlegen` und für die
// Übernahme, damit die Anweisung **eine** ist.
func deckEinfuegen(
	ctx context.Context, ziel execer, deck domain.Deck, profilID int64,
) (domain.Deck, error) {
	deck.Normalisieren()

	ergebnis, err := ziel.ExecContext(ctx,
		`INSERT INTO decks (name, spiel, notiz, format_id, profil_id)
		 VALUES (?, ?, ?, ?, ?)`,
		deck.Name, deck.Spiel, deck.Notiz, formatspalte(deck.FormatID), profilID)
	if err != nil {
		return domain.Deck{}, uebersetze(err, "das Deck anlegen")
	}

	// **Ohne eigenen Zweig für die Kennung.** `uebersetze` gibt für `nil`
	// nichts zurück, und ein Zweig dafür wäre einer, den kein Lauf betreten
	// kann (§3) — dieselbe Form wie in `sammlung.go`.
	kennung, idfehler := ergebnis.LastInsertId()

	deck.ID = kennung
	deck.Fassung = 1

	return deck, uebersetze(idfehler, "die Kennung des Decks lesen")
}
