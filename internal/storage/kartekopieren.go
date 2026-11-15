package storage

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Eine Karte kommt hinzu — leer oder als Kopie.
//
// **Eine eigene Datei**, weil `kartenschreiber.go` an der Zeilengrenze steht
// (§2). Der Schnitt folgt der Frage: dort steht, wie ein Feldwert geschrieben
// wird, hier, wie eine ganze Karte entsteht.

// KarteKopieren legt eine Kopie der Karte ans Ende ihres Sets.
//
// # Was mitkommt und wohin
//
// **ported.** MSE2 kennt kein „Duplizieren", sondern Kopieren und Einfügen über
// die Zwischenablage (`src/gui/control/card_list.cpp:146–172`, Fassung f9d9356
// vom 2021-07-06). Drei seiner Regeln gelten hier genauso:
//
//   - **Die ganze Karte reist mit.** Die Zwischenablage schreibt sie über
//     dieselbe Reflexion, die auch eine `.mse-set` schreibt
//     (`src/data/format/clipboard.cpp:44–79`) — jeder Feldwert, keine Auswahl.
//   - **Die Kopie kommt ans Ende**, nicht neben das Original:
//     `AddCardAction(ADD, …)` nimmt `pos = container.size()`
//     (`src/data/action/generic.hpp:54–56`).
//   - **Es ist ein Schritt.** Dieselbe Handlung, dieselbe Rücknahme.
//
// # Und eine Regel, die hier entfällt
//
// **adapted.** MSE2 muss beim Kopieren die Stiloptionen der Karte gegen ein
// Stylesheet lesen und setzt dafür vor dem Schreiben `card->stylesheet` um
// (`src/data/format/clipboard.cpp:63–82`). Hier liegen `has styling` und
// `styling data` als gewöhnliche Feldwerte in `card_values`; wer alle Werte
// kopiert, hat sie mitkopiert. Der Umweg entfällt **ersatzlos**, nicht
// versehentlich.
//
// Ebenso entfällt MSE2s Prüfung, ob die Karte aus demselben Spiel stammt
// (`clipboard.cpp:88–99`): eine Kopie im selben Set kann das nicht treffen.
func (k *Kartenschreiber) KarteKopieren(ctx context.Context, karteID int64) (int64, error) {
	tx, err := k.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, uebersetze(err, "die Karte %d kopieren", karteID)
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach Commit ein ErrTxDone

	// **Nur eine Karte des gezeigten Profils** (#1815).
	if fehler := karteImProfil(ctx, tx, karteID); fehler != nil {
		return 0, fehler
	}

	// **Erst die Karte, dann ihr Set.** `kartenstandLesen` macht aus einer
	// Karte, die es nicht gibt, das `ErrNotFound` — hier eine zweite Stelle
	// dafür zu bauen hieße, dieselbe Antwort zweimal zu geben.
	stand, err := kartenstandLesen(ctx, tx, karteID)
	if err != nil {
		return 0, err
	}

	var setID int64
	if fehler := tx.QueryRowContext(ctx,
		`SELECT set_id FROM cards WHERE id = ?`, karteID).Scan(&setID); fehler != nil {
		return 0, uebersetze(fehler, "das Set der Karte %d lesen", karteID)
	}

	kennung, fehler := k.karteEinreihen(ctx, tx, setID, stand.Felder)
	if fehler != nil {
		return 0, fehler
	}

	if fehler := tx.Commit(); fehler != nil {
		return 0, uebersetze(fehler, "die Karte %d kopieren", karteID)
	}

	return kennung, nil
}

// karteEinreihen hängt eine Karte ans Ende und schreibt ihren Verlaufsschritt.
//
// **Der Schritt gehört dazu, nicht daneben** — sonst gäbe es einen Weg, eine
// Karte anzulegen, ohne dass ⌘Z davon weiß
// (#626). In
// MSE2 gibt es diesen Weg nicht: `AddCardAction(Set&)` legt an **und** reiht
// ein (`src/data/action/set.cpp:20–22`, Fassung f9d9356 vom 2021-07-06).
//
// **Der Rückweg ist der leere Stand.** `GabEs: false` heißt für
// [ganzeKarteGehen] „entfernen"; die Rücknahme braucht damit keine eigene Art
// und kein zweites Stück Code.
func (k *Kartenschreiber) karteEinreihen(
	ctx context.Context, tx *sql.Tx, setID int64, felder []domain.Feldstand,
) (int64, error) {
	return k.karteEinreihenInGruppe(ctx, tx, setID, felder, "")
}

// karteEinreihenInGruppe tut dasselbe unter einer Gruppenmarke.
//
// **Ohne Marke bleibt es ein Schritt für sich** — der Fall der einzelnen
// Karte, und derselbe Vertrag wie bei [schreibeKartenverlaufInGruppe]. Mit
// Marke werden mehrere Karten zu **einem** ⌘Z
// (#1337).
func (k *Kartenschreiber) karteEinreihenInGruppe(
	ctx context.Context, tx *sql.Tx, setID int64,
	felder []domain.Feldstand, gruppe string,
) (int64, error) {
	nachher := domain.Kartenstand{GabEs: true, Felder: felder}

	var kennung int64
	if err := tx.QueryRowContext(ctx, karteAnlegen, setID, setID).
		Scan(&kennung, &nachher.Reihenfolge); err != nil {
		return 0, uebersetze(err, "eine Karte im Set %d anlegen", setID)
	}

	for i := range felder {
		feld := &felder[i]

		if err := feldEinsetzen(ctx, tx, kennung, feld); err != nil {
			return 0, err
		}
	}

	if err := schreibeKartenverlaufInGruppe(ctx, tx, kennung,
		domain.Kartenstand{}, nachher, k.urheber, gruppe); err != nil {
		return 0, err
	}

	return kennung, nil
}

// feldEinsetzen schreibt einen Feldwert samt seiner Stelle.
//
// **Die Stelle kommt mit**, weil sie für den Rundlauf zählt: `.mse-set` ist
// eine Textdatei, und die Reihenfolge der Zeilen einer Karte ist ihre
// Reihenfolge in der Datei (#48).
func feldEinsetzen(
	ctx context.Context, tx *sql.Tx, karteID int64, feld *domain.Feldstand,
) error {
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO card_values (card_id, feld, wert, reihenfolge) VALUES (?, ?, ?, ?)`,
		karteID, feld.Feld, feld.Wert, feld.Reihenfolge); err != nil {
		return uebersetze(err, "das Feld %q der Karte %d schreiben", feld.Feld, karteID)
	}

	return nil
}
