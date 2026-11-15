// **Schlüsselwörter pflegen** —
// #1062.
//
// Anlegen, ändern, löschen — für die des **Sets**. Die des Spiels kommen aus
// dem Paket, gehören keinem Set und sind unveränderlich (`Keyword::fixed`,
// `src/data/keyword.hpp`, Zeile 104, Fassung f9d9356 vom 2021-07-06).

package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Schluesselwortablage schreibt die Schlüsselwörter eines Sets.
type Schluesselwortablage struct {
	db *sql.DB

	// urheber sagt, wessen Änderung im Verlauf steht — `fenster`, `server`,
	// `kommandozeile`, `dienst`
	// (#1077).
	urheber domain.Urheber
}

// NeueSchluesselwortablage baut sie.
func NeueSchluesselwortablage(db *sql.DB) *Schluesselwortablage {
	return &Schluesselwortablage{db: db, urheber: domain.UrheberFenster}
}

// Als gibt eine Ablage, deren Änderungen unter einem anderen Namen im Verlauf
// stehen.
//
// **Ein zweiter Wert, kein umgestellter** — dieselbe Gestalt wie bei
// [Kartenschreiber.Als]: eine Ablage, die ihren Urheber wechselt, wechselte
// ihn für jeden, der sie schon hält.
func (a *Schluesselwortablage) Als(wer domain.Urheber) *Schluesselwortablage {
	zweite := *a
	zweite.urheber = wer

	return &zweite
}

// schluesselwortspalten ordnet die Feldnamen von außen den Spalten zu.
//
// **Eine Aufzählung und keine Zeichenkettenrechnung.** Der Name kommt aus einer
// Bindung; wer ihn ungeprüft in ein `UPDATE` setzte, hätte eine Spaltenwahl aus
// dem Fenster.
//
// **Die Namen sind die der Datei** — `keyword`, `match`, `reminder`, `rules`,
// `mode` heißen hier `wort`, `muster`, `erinnerung`, `regeln`, `modus`: die
// Spalten sind deutsch wie der übrige Bestand, und die Übersetzung steht an
// **einer** Stelle.
var schluesselwortspalten = map[string]string{
	"wort":       "wort",
	"muster":     "muster",
	"erinnerung": "erinnerung",
	"regeln":     "regeln",
	"modus":      "modus",
}

// SchluesselwortAnlegen hängt ein leeres Schlüsselwort ans Ende.
//
// **Leer in jedem Feld außer dem Modus** — so legt MSE eines an
// (`AddKeywordAction::AddKeywordAction`, `src/data/action/keyword.cpp`, Zeilen
// 20–32, Fassung f9d9356 vom 2021-07-06): es sucht den Modus mit
// `is default: true` und setzt sonst nichts.
//
// **Der Modus kommt von außen**, denn er steht im **Spielpaket**, und
// `internal/storage` kennt keine Pakete (§2).
//
// **Und der Verlauf bekommt einen Schritt** (#1077): ⌘Z nimmt das Anlegen
// zurück, ⌘⇧Z bringt es wieder.
func (a *Schluesselwortablage) SchluesselwortAnlegen(
	ctx context.Context, setID int64, modus string,
) (int64, error) {
	var kennung int64

	fehler := a.imZug(ctx, func(tx *sql.Tx) (
		vorher, nachher domain.Schluesselwortstand, fehler error,
	) {
		var stelle int64

		// **`RETURNING` gibt beides** — Kennung und Stelle. Die Stelle muss in
		// den Verlauf, damit ein zurückgenommenes Wiederherstellen sie
		// wiederfindet.
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO set_keywords (set_id, reihenfolge, modus)
			 VALUES (?, (SELECT coalesce(max(reihenfolge) + 1, 0)
			               FROM set_keywords WHERE set_id = ?), ?)
			 RETURNING id, reihenfolge`, setID, setID, modus).
			Scan(&kennung, &stelle); err != nil {
			return vorher, nachher, uebersetze(err, "ein Schlüsselwort anlegen")
		}

		return domain.Schluesselwortstand{ID: kennung},
			domain.Schluesselwortstand{
				ID: kennung, GabEs: true, Reihenfolge: stelle, Modus: modus,
			}, nil
	}, setID)

	return kennung, fehler
}

// imZug fährt eine Änderung samt ihrem Verlaufsschritt in **einer**
// Transaktion.
//
// **Beide Stände kommen aus der Arbeit**, nicht aus zwei Abfragen um sie
// herum. Der erste Anlauf las den Stand davor **und** den danach; die zweite
// Abfrage war ein Zweig, den kein Aufruf betreten kann — wenn die erste
// gelang, gelingt die zweite auch (dieselbe Tabelle, dieselbe Transaktion). §3
// sagt, was ein solcher Zweig wert ist.
//
// **Und jede der drei Änderungen kennt beide Stände ohnehin**: das Anlegen
// bekommt sie aus `RETURNING`, das Setzen aus dem Feld, das es ändert, das
// Löschen aus dem, was dastand.
func (a *Schluesselwortablage) imZug(
	ctx context.Context,
	arbeit func(*sql.Tx) (vorher, nachher domain.Schluesselwortstand, fehler error),
	setID int64,
) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "die Änderung beginnen")
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach einem Commit ohne Wirkung

	// **Nur in einem Set des gezeigten Profils** — alle drei Wege gehen durch
	// diesen Zug (#1815).
	if fehler := setImProfil(ctx, tx, setID); fehler != nil {
		return fehler
	}

	vorher, nachher, arbeitsfehler := arbeit(tx)
	if arbeitsfehler != nil {
		return arbeitsfehler
	}

	if verlauf := schreibeSchluesselwortverlauf(
		ctx, tx, setID, vorher, nachher, a.urheber); verlauf != nil {
		return verlauf
	}

	if abschluss := tx.Commit(); abschluss != nil {
		return uebersetze(abschluss, "die Änderung abschließen")
	}

	return nil
}

// SchluesselwortSetzen schreibt ein Feld eines Schlüsselworts.
//
// **Der vorige Stand reist in den Verlauf** — seit #1077, für die Rücknahme.
//
// **Die Setkennung steht in der Bedingung.** Eine Kennung aus einem anderen Set
// ist ein Zugriff über die Setgrenze hinweg, und die Oberfläche kann sich
// vertun (§7, Regel 1).
func (a *Schluesselwortablage) SchluesselwortSetzen(
	ctx context.Context, setID, wortID int64, feld, wert string,
) error {
	spalte, bekannt := schluesselwortspalten[feld]
	if !bekannt {
		return fmt.Errorf("%w: %q ist kein Feld eines Schlüsselworts",
			domain.ErrValidation, feld)
	}

	return a.imZug(ctx, func(tx *sql.Tx) (
		vorher, nachher domain.Schluesselwortstand, fehler error,
	) {
		vorher, fehler = schluesselwortstand(ctx, tx, setID, wortID)
		if fehler != nil {
			return vorher, nachher, fehler
		}

		if fremd := fremderWert(ctx, feld, feldwert(vorher, feld)); fremd != nil {
			return vorher, nachher, fremd
		}

		ergebnis, err := tx.ExecContext(ctx,
			//nolint:gosec // die Spalte kommt aus [schluesselwortspalten], nicht von außen
			`UPDATE set_keywords SET `+spalte+` = ? WHERE id = ? AND set_id = ?`,
			wert, wortID, setID)
		if err != nil {
			return vorher, nachher, uebersetze(err, "das Schlüsselwort %d schreiben", wortID)
		}

		if getrofffehler := getroffen(ergebnis, setID, wortID); getrofffehler != nil {
			return vorher, nachher, getrofffehler
		}

		return vorher, mitFeld(vorher, feld, wert), nil
	}, setID)
}

// mitFeld gibt den Stand, wie er nach dem Setzen **eines** Feldes dasteht.
//
// **Gerechnet und nicht noch einmal gelesen**: der Wert steht im Aufruf, und
// eine zweite Abfrage in derselben Transaktion gäbe dasselbe zurück — nur mit
// einem Fehlerweg mehr, den kein Aufruf betritt.
//
// **Ein unbekanntes Feld kommt hier nicht an**: [Schluesselwortablage.SchluesselwortSetzen]
// weist es ab, bevor geschrieben wird.
func mitFeld(stand domain.Schluesselwortstand, feld, wert string) domain.Schluesselwortstand {
	nachher := stand

	switch feld {
	case "wort":
		nachher.Wort = wert
	case "muster":
		nachher.Muster = wert
	case "erinnerung":
		nachher.Erinnerung = wert
	case "regeln":
		nachher.Regeln = wert
	case "modus":
		nachher.Modus = wert
	}

	return nachher
}

// getroffen macht aus „keine Zeile betroffen" ein `ErrNotFound`.
//
// **Ein Schreiben, das nichts trifft, ist ein Fehler und kein Erfolg.** Wer
// eine Kennung schickt, die es nicht gibt, hat einen Stand vor sich, den es
// nicht mehr gibt — und soll das erfahren.
func getroffen(ergebnis sql.Result, setID, wortID int64) error {
	zeilen, err := betroffen(ergebnis)
	if err != nil {
		return err
	}

	if zeilen == 0 {
		return fmt.Errorf("%w: das Set %d führt kein Schlüsselwort %d",
			domain.ErrNotFound, setID, wortID)
	}

	return nil
}

// SchluesselwortLoeschen nimmt eines weg.
//
// **Die Reihenfolge der übrigen bleibt, wie sie war.** Sie muss nur aufsteigend
// sein, nicht lückenlos: was zählt, ist welches zuerst auf einen Text passt.
// Ein Umnummerieren schriebe jede Zeile des Sets neu, um nichts zu ändern.
//
// **Und es kommt zurück** (#1077): ein zurückgenommenes Löschen bringt es an
// **seine** Stelle, nicht ans Ende — die Reihenfolge entscheidet, welches
// zuerst auf einen Text passt.
func (a *Schluesselwortablage) SchluesselwortLoeschen(
	ctx context.Context, setID, wortID int64,
) error {
	return a.imZug(ctx, func(tx *sql.Tx) (
		vorher, nachher domain.Schluesselwortstand, fehler error,
	) {
		vorher, fehler = schluesselwortstand(ctx, tx, setID, wortID)
		if fehler != nil {
			return vorher, nachher, fehler
		}

		ergebnis, err := tx.ExecContext(ctx,
			`DELETE FROM set_keywords WHERE id = ? AND set_id = ?`, wortID, setID)
		if err != nil {
			return vorher, nachher, uebersetze(err, "das Schlüsselwort %d löschen", wortID)
		}

		if getrofffehler := getroffen(ergebnis, setID, wortID); getrofffehler != nil {
			return vorher, nachher, getrofffehler
		}

		return vorher, domain.Schluesselwortstand{ID: wortID}, nil
	}, setID)
}
