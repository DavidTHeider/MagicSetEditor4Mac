package storage

// **Ein Turnier anlegen — samt Format, Teilnehmenden und Verlauf** —
// DOM-04.AK-01 und AK-03
// (#1699).
//
// **Alles in einem Zug.** Eine Vorlage bringt Namen mit, eine Ausgabe den
// ganzen Verlauf; ein Turnier, dem die Hälfte fehlt, weil der dritte Name
// oder die zweite Runde scheiterte, wäre ein halber Stand — und wer es noch
// einmal versucht, stieße auf den vergebenen Namen des halben.

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Anlegen legt ein Turnier an — samt den Teilnehmenden, die es mitbringt.
//
// **Ob das Format zum Spiel passt, prüft der Dienst** (§2), wie beim Deck;
// hier steht nur, dass es in der Spalte landet.
func (a *Turnierablage) Anlegen(ctx context.Context, turnier domain.Turnier) (
	domain.Turnier, error,
) {
	// **Das Protokoll beginnt mit dem Anlegen** — DOM-04.AK-04 (#1699): erst
	// das Turnier, dann jeder mitgebrachte Name.
	return a.anlegenMit(ctx, turnier, func(angelegt domain.Turnier) []domain.Protokolleintrag {
		eintraege := make([]domain.Protokolleintrag, 0, 1+len(angelegt.Teilnehmer))
		eintraege = append(eintraege, domain.Protokolleintrag{Vorgang: domain.TurnierAngelegt, Angaben: angelegt.Name})

		for _, eines := range angelegt.Teilnehmer {
			eintraege = append(eintraege, domain.Protokolleintrag{
				Vorgang: domain.TurnierAngemeldet, Angaben: eines.Name,
			})
		}

		return eintraege
	})
}

// Einlesen legt ein Turnier **samt Verlauf** an — Teilnehmende mit ihren
// Listen, Runden, Ergebnisse, Abschluss (DOM-04.AK-03).
//
// **Die Kennungen der Teilnehmenden im Aufruf sind vorläufig**: die Paarungen
// verweisen über sie auf ihre Seiten, und der Speicher vergibt die echten.
// Das Protokoll beginnt mit einer Zeile: eingelesen.
func (a *Turnierablage) Einlesen(ctx context.Context, turnier domain.Turnier) (
	domain.Turnier, error,
) {
	return a.anlegenMit(ctx, turnier, func(angelegt domain.Turnier) []domain.Protokolleintrag {
		return []domain.Protokolleintrag{{Vorgang: domain.TurnierEingelesen, Angaben: angelegt.Name}}
	})
}

// anlegenMit ist der gemeinsame Rumpf: ein Zug, das Turnier samt allem, was
// es trägt, und die Zeilen des Protokolls, die der Weg dazu nennt.
func (a *Turnierablage) anlegenMit(
	ctx context.Context, turnier domain.Turnier,
	protokoll func(domain.Turnier) []domain.Protokolleintrag,
) (domain.Turnier, error) {
	// **Jede Wurzelabfrage klemmt auf ein Profil** (#1351, Vermerk 0066).
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Turnier{}, profilfehler
	}

	zug, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Turnier{}, uebersetze(err, "das Anlegen beginnen")
	}
	defer zug.Rollback() //nolint:errcheck // nach einem Commit ist es ein No-op

	angelegt, err := anlegenImZug(ctx, zug, turnier, profilID)
	if err != nil {
		return domain.Turnier{}, err
	}

	if fehler := protokollieren(ctx, zug, angelegt.ID, protokoll(angelegt)...); fehler != nil {
		return domain.Turnier{}, fehler
	}

	if err := zug.Commit(); err != nil {
		return domain.Turnier{}, uebersetze(err, "das Anlegen abschließen")
	}

	return angelegt, nil
}

// anlegenImZug schreibt Turnier, Teilnehmende und Verlauf in den Zug des
// Rufers.
func anlegenImZug(
	ctx context.Context, zug *sql.Tx, turnier domain.Turnier, profilID int64,
) (domain.Turnier, error) {
	ergebnis, err := zug.ExecContext(ctx,
		`INSERT INTO turniere (name, spiel, einzelspiel, notiz, format_id, abgeschlossen, profil_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		turnier.Name, turnier.Spiel, turnier.Einzelspiel, turnier.Notiz,
		kennungOderNull(turnier.FormatID), textOderNull(turnier.Abgeschlossen), profilID)
	if err != nil {
		return domain.Turnier{}, uebersetze(err, "das Turnier anlegen")
	}

	kennung, err := ergebnis.LastInsertId()
	if err != nil {
		return domain.Turnier{}, uebersetze(err, "die Kennung des Turniers lesen")
	}

	// **Eine neue Liste und nicht die des Aufrufers** — `meldeAnImZug` setzt
	// Kennung und Stelle, und die gehören dem angelegten Turnier.
	angemeldet := make([]domain.Teilnehmer, 0, len(turnier.Teilnehmer))
	echt := make(map[int64]int64, len(turnier.Teilnehmer))

	for _, teilnehmer := range turnier.Teilnehmer {
		// **Ein Deck nur aus diesem Profil** (#1815) — wie beim Anmelden.
		if teilnehmer.DeckID != 0 {
			if fehler := deckImProfil(ctx, zug, teilnehmer.DeckID); fehler != nil {
				return domain.Turnier{}, fehler
			}
		}

		vorlaeufig := teilnehmer.ID

		eines, meldefehler := meldeAnImZug(ctx, zug, kennung, teilnehmer)
		if meldefehler != nil {
			return domain.Turnier{}, meldefehler
		}

		echt[vorlaeufig] = eines.ID
		angemeldet = append(angemeldet, eines)
	}

	runden, err := verlaufImZug(ctx, zug, kennung, turnier.Runden, echt)
	if err != nil {
		return domain.Turnier{}, err
	}

	turnier.ID = kennung
	turnier.Fassung = 1
	turnier.Teilnehmer = angemeldet
	turnier.Runden = runden

	return turnier, nil
}

// verlaufImZug legt die Runden eines eingelesenen Turniers samt Ergebnissen.
func verlaufImZug(
	ctx context.Context, zug *sql.Tx, turnierID int64, runden []domain.Runde, echt map[int64]int64,
) ([]domain.Runde, error) {
	gelegt := make([]domain.Runde, 0, len(runden))

	for _, runde := range runden {
		paarungen, err := mitEchtenSeiten(runde.Paarungen, echt)
		if err != nil {
			return nil, err
		}

		eine, err := legeRundeImZug(ctx, zug, turnierID, runde.Nummer, paarungen)
		if err != nil {
			return nil, err
		}

		for _, paarung := range eine.Paarungen {
			if paarung.Ergebnis == nil {
				continue
			}

			if _, err := zug.ExecContext(ctx,
				`UPDATE turnierpaarungen SET spiele_a = ?, spiele_b = ?, unentschieden = ?
				  WHERE id = ?`,
				paarung.Ergebnis.SpieleA, paarung.Ergebnis.SpieleB, paarung.Ergebnis.Unentschieden,
				paarung.ID); err != nil {
				return nil, uebersetze(err, "ein Ergebnis einlesen")
			}
		}

		gelegt = append(gelegt, eine)
	}

	return gelegt, nil
}

// mitEchtenSeiten ersetzt die vorläufigen Kennungen einer Paarung durch die
// echten — **eine Seite, die es nicht gibt, ist ein Nein** und keine Paarung
// ins Leere.
func mitEchtenSeiten(paarungen []domain.Paarung, echt map[int64]int64) ([]domain.Paarung, error) {
	ersetzt := make([]domain.Paarung, 0, len(paarungen))

	for _, paarung := range paarungen {
		a, gibtA := echt[paarung.A]
		b, gibtB := echt[paarung.B]

		if !gibtA || (paarung.B != 0 && !gibtB) {
			return nil, fmt.Errorf(
				"%w: eine Paarung verweist auf eine Seite, die es nicht gibt", domain.ErrValidation)
		}

		paarung.A, paarung.B = a, b
		ersetzt = append(ersetzt, paarung)
	}

	return ersetzt, nil
}

// textOderNull macht aus einem leeren Text `NULL` — ein offenes Turnier hat
// kein Abschlussdatum, nicht ein leeres.
func textOderNull(text string) any {
	if text == "" {
		return nil
	}

	return text
}
