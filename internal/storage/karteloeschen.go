package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KarteLoeschen entfernt eine Karte samt ihren Feldwerten.
//
// # Der Weg zurück reist mit
//
// **In derselben Transaktion**, und zwar der ganze Stand der Karte: Kennung,
// Reihenfolge und jeder Feldwert an seiner Stelle. `card_values` hängt an
// `ON DELETE CASCADE`; wer sie nicht mitschreibt, brächte eine leere Karte
// zurück und nennte das eine Rücknahme
// (#621).
//
// # Die Stelle kommt mit zurück — so hält es MSE2
//
// **ported.** MSE2 löscht eine Karte über dieselbe Handlung, mit der es eine
// anlegt, nur in der anderen Richtung: `AddCardAction(REMOVE, …)`
// (`src/data/action/set.cpp:24–41`, Fassung f9d9356 vom 2021-07-06). Sie hält
// die Karte **samt ihrer Stelle** fest — `GenericAddAction::Step{pos, item}`
// — und setzt sie bei der Rücknahme an `container.begin() + pos` wieder ein,
// nicht ans Ende (`src/data/action/generic.hpp:34–41, 57–65, 95–112`,
// dieselbe Fassung). Genau das tut `ganzeKarteGehen`: dieselbe Kennung,
// dieselbe Reihenfolge, dieselben Feldwerte.
//
// # Die Lücke bleibt — und hier weicht es ab
//
// **adapted.** MSE2 hält die Karten in einem `vector`; ein `erase` schiebt
// alles dahinter um eine Stelle vor, die Zählung bleibt dicht. Hier rücken die
// übrigen Karten **nicht** nach: `cards` trägt `UNIQUE (set_id, reihenfolge)`,
// und würden sie nachrücken, müsste die Rücknahme sie alle zurückschieben —
// dazwischen könnte jemand eine Karte anlegen, die den Platz einnimmt.
//
// Sichtbar ist der Unterschied nicht — **seit
// #666**: die
// Reihenfolge ist ein **Sortierschlüssel**, keine dichte Zählung, und die
// Liste zeigt ihre eigene Stelle.
//
// MSE2 speichert die Reihenfolge ohnehin nicht als Zahl, sondern als
// Reihenfolge der Karteneinträge in der Datei.
func (k *Kartenschreiber) KarteLoeschen(ctx context.Context, karteID int64) error {
	tx, err := k.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "die Karte %d löschen", karteID)
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach Commit ein ErrTxDone

	// **Nur eine Karte des gezeigten Profils** (#1815).
	if fehler := karteImProfil(ctx, tx, karteID); fehler != nil {
		return fehler
	}

	stand, err := kartenstandLesen(ctx, tx, karteID)
	if err != nil {
		return err
	}

	if err := schreibeKartenverlauf(ctx, tx, karteID, stand,
		domain.Kartenstand{}, k.urheber); err != nil {
		return err
	}

	if _, fehler := tx.ExecContext(ctx,
		`DELETE FROM cards WHERE id = ?`, karteID); fehler != nil {
		return uebersetze(fehler, "die Karte %d löschen", karteID)
	}

	if fehler := tx.Commit(); fehler != nil {
		return uebersetze(fehler, "die Karte %d löschen", karteID)
	}

	return nil
}

// kartenstandLesen holt die ganze Karte, bevor sie verschwindet.
//
// **`ErrNotFound`, wenn es sie nicht gibt** — ein stiller Erfolg auf einer
// Karte, die niemand hat, wäre eine Meldung über etwas, das nicht geschah (§8).
func kartenstandLesen(
	ctx context.Context, tx *sql.Tx, karteID int64,
) (domain.Kartenstand, error) {
	stand := domain.Kartenstand{GabEs: true}

	err := tx.QueryRowContext(ctx,
		`SELECT reihenfolge FROM cards WHERE id = ?`, karteID).
		Scan(&stand.Reihenfolge)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		// **Eine Karte, die es nicht gibt, ist ein Befund** — und zwar
		// derselbe, den jede andere Schicht kennt (§8). Ein stiller Erfolg
		// meldete etwas, das nicht geschah.
		return domain.Kartenstand{}, fmt.Errorf("%w: die Karte %d gibt es nicht",
			domain.ErrNotFound, karteID)

	case err != nil:
		return domain.Kartenstand{}, uebersetze(err, "die Karte %d lesen", karteID)
	}

	// **Die Herkünfte in einer eigenen Abfrage** (#1036). Ein Join hier
	// erweiterte die Spalten der Feldabfrage — und wer eine Abfrage um eine
	// Spalte erweitert, zieht die Zeilengestalten des Stolpertreibers mit;
	// derselbe Fehler ist in Phase 6 zweimal passiert. Beide laufen in
	// **derselben** Transaktion, der Stand ist damit einer.
	//
	// **Und es ist dieselbe Abfrage wie die des Lesers** — eine zweite
	// daneben liefe beim nächsten Schemaschritt auseinander.
	herkuenfte, herkunftfehler := herkuenfteVon(ctx, tx, karteID)
	if herkunftfehler != nil {
		return domain.Kartenstand{}, herkunftfehler
	}

	zeilen, err := tx.QueryContext(ctx,
		`SELECT feld, wert, reihenfolge FROM card_values
		  WHERE card_id = ? ORDER BY reihenfolge`, karteID)
	if err != nil {
		return domain.Kartenstand{}, uebersetze(err, "die Felder der Karte %d lesen", karteID)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen ist gelesen

	for zeilen.Next() {
		var feld domain.Feldstand
		if fehler := zeilen.Scan(&feld.Feld, &feld.Wert, &feld.Reihenfolge); fehler != nil {
			return domain.Kartenstand{}, uebersetze(fehler,
				"ein Feld der Karte %d lesen", karteID)
		}

		feld.Herkunft = herkuenfte[feld.Feld]
		stand.Felder = append(stand.Felder, feld)
	}

	if fehler := zeilen.Err(); fehler != nil {
		return domain.Kartenstand{}, uebersetze(fehler,
			"die Felder der Karte %d lesen", karteID)
	}

	return stand, nil
}

// schreibeKartenverlauf legt den Schritt an — **in der Transaktion des
// Aufrufers**, aus demselben Grund wie [schreibeVerlauf].
func schreibeKartenverlauf(
	ctx context.Context, tx *sql.Tx, karteID int64,
	vorher, nachher domain.Kartenstand, wer domain.Urheber,
) error {
	return schreibeKartenverlaufInGruppe(ctx, tx, karteID, vorher, nachher, wer, "")
}

// schreibeKartenverlaufInGruppe tut dasselbe unter einer Gruppenmarke.
//
// **Eine Gruppe macht aus mehreren Schritten einen** — ein ⌘Z holt sie
// zusammen zurück. Ohne Marke (die leere Zeichenkette) bleibt es ein Schritt
// für sich; das ist der Fall der einzelnen Karte.
func schreibeKartenverlaufInGruppe(
	ctx context.Context, tx *sql.Tx, karteID int64,
	vorher, nachher domain.Kartenstand, wer domain.Urheber, gruppe string,
) error {
	hin, _ := json.Marshal(nachher)    //nolint:errcheck // ein Verbund aus Text und Zahlen
	zurueck, _ := json.Marshal(vorher) //nolint:errcheck // dito

	_, err := tx.ExecContext(ctx,
		`INSERT INTO history
		        (set_id, karte_id, zeitpunkt, handlung, hinweg, rueckweg, urheber, gruppe)
		 SELECT set_id, ?, ?, ?, ?, ?, ?, ? FROM cards WHERE id = ?`,
		karteID, time.Now().UTC().Format(time.RFC3339Nano), domain.HandlungKarte,
		string(hin), string(zurueck), string(wer), gruppe, karteID)
	if err != nil {
		return uebersetze(err, "den Verlauf zur Karte %d schreiben", karteID)
	}

	return kuerzeVerlauf(ctx, tx, karteID)
}

// kartenstandAus liest einen Kartenstand aus seinem JSON.
func kartenstandAus(roh string) (*domain.Kartenstand, error) {
	if roh == "" {
		return &domain.Kartenstand{}, nil
	}

	var stand domain.Kartenstand
	if err := json.Unmarshal([]byte(roh), &stand); err != nil {
		return nil, uebersetze(err, "einen Kartenweg lesen")
	}

	return &stand, nil
}

// ganzeKarteGehen legt die Karte an oder entfernt sie — je nachdem, welcher
// Stand gilt.
//
// **Mit derselben Kennung.** Die Verlaufszeilen zeigen über `karte_id` auf
// sie; käme sie mit einer neuen zurück, hingen alle früheren Feldänderungen
// ins Leere. `AUTOINCREMENT` vergibt eine Kennung nie zweimal, die Lücke ist
// also frei.
func ganzeKarteGehen(
	ctx context.Context, tx *sql.Tx, karteID, setID int64, stand *domain.Kartenstand,
) error {
	if stand == nil || !stand.GabEs {
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM cards WHERE id = ?`, karteID); err != nil {
			return uebersetze(err, "die Karte %d entfernen", karteID)
		}

		return nil
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO cards (id, set_id, reihenfolge) VALUES (?, ?, ?)`,
		karteID, setID, stand.Reihenfolge); err != nil {
		return uebersetze(err, "die Karte %d zurückbringen", karteID)
	}

	for i := range stand.Felder {
		feld := &stand.Felder[i]

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO card_values (card_id, feld, wert, reihenfolge) VALUES (?, ?, ?, ?)`,
			karteID, feld.Feld, feld.Wert, feld.Reihenfolge); err != nil {
			return uebersetze(err, "das Feld %q der Karte %d zurückbringen",
				feld.Feld, karteID)
		}

		// **Und seine Herkunft** (#1036): `wert_herkunft` hängt über
		// `ON DELETE CASCADE` an `cards` und ist mit der Karte verschwunden.
		// Ein zurückgeholtes Scryfall-Bild stünde sonst ohne seine Nennung da.
		if feld.Herkunft == "" {
			continue
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO wert_herkunft (card_id, feld, herkunft) VALUES (?, ?, ?)`,
			karteID, feld.Feld, feld.Herkunft); err != nil {
			return uebersetze(err, "die Herkunft von %q der Karte %d zurückbringen",
				feld.Feld, karteID)
		}
	}

	return nil
}
