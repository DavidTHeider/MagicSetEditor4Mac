package storage

// **Die eigenen Autoersetzungen eines Spiels** —
// #1341,
// Vermerk [0060](../../docs/entscheidungen/0060-eigene-autoersetzungen-liegen-im-bestand.md).

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Autoersetzungsablage liest und schreibt, was der Bestand an Regeln führt.
type Autoersetzungsablage struct{ db *sql.DB }

// NeueAutoersetzungsablage gibt die Ablage über diesem Bestand.
func NeueAutoersetzungsablage(db *sql.DB) *Autoersetzungsablage {
	return &Autoersetzungsablage{db: db}
}

// Regeln gibt die Zeilen eines Spiels — **nach dem Muster geordnet**.
//
// **Geordnet vom Bestand und nicht vom Aufrufer**: sonst sortierte jede
// Fläche für sich, und zwei von ihnen liefen auseinander. Dieselbe Regel wie
// bei [Listenablage.Spalten].
func (a *Autoersetzungsablage) Regeln(ctx context.Context, spiel string) (
	[]domain.Autoersetzungsregel, error,
) {
	if err := pruefeSpiel(spiel); err != nil {
		return nil, err
	}

	// **Jede Wurzelabfrage klemmt auf ein Profil** (#1351, Vermerk 0066).
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, err := a.db.QueryContext(ctx,
		`SELECT muster, ersatz, ganzes_wort, aktiv, eigen
		   FROM spiel_autoersetzungen
		  WHERE spiel = ? AND profil_id = ?
		  ORDER BY muster`, spiel, profilID)
	if err != nil {
		return nil, fmt.Errorf("die Autoersetzungen lesen: %w", err)
	}

	defer zeilen.Close() //nolint:errcheck // der Fehler steht in zeilen.Err()

	regeln := []domain.Autoersetzungsregel{}

	for zeilen.Next() {
		var regel domain.Autoersetzungsregel

		if err := zeilen.Scan(&regel.Muster, &regel.Ersatz,
			&regel.GanzesWort, &regel.Aktiv, &regel.Eigen); err != nil {
			return nil, fmt.Errorf("eine Autoersetzung lesen: %w", err)
		}

		regeln = append(regeln, regel)
	}

	// **Als Ausdruck und nicht als Verzweigung** — dieselbe Gestalt wie in
	// [Listenablage.Spalten]. Ein `if` darüber wäre ein Zweig, den kein Lauf
	// betreten kann: SQLite meldet den Fehler schon beim `Scan`.
	return regeln, uebersetze(zeilen.Err(), "die Autoersetzungen lesen")
}

// Setzen legt eine Regel an oder schreibt sie fort.
//
// **Ein `INSERT … ON CONFLICT`**, weil das Muster der Schlüssel ist: wer
// dieselbe Regel zweimal setzt, meint dieselbe Regel. Das ist MSEs
// Vergleich über `match` in einer Zeile SQL.
func (a *Autoersetzungsablage) Setzen(
	ctx context.Context, spiel string, regel domain.Autoersetzungsregel,
) error {
	if err := pruefeSpiel(spiel); err != nil {
		return err
	}

	if err := domain.PruefeAutoersetzung(regel.Muster); err != nil {
		return err
	}

	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	if _, err := a.db.ExecContext(ctx,
		`INSERT INTO spiel_autoersetzungen
		   (spiel, muster, ersatz, ganzes_wort, aktiv, eigen, profil_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (profil_id, spiel, muster) DO UPDATE SET
		   ersatz = excluded.ersatz,
		   ganzes_wort = excluded.ganzes_wort,
		   aktiv = excluded.aktiv,
		   eigen = excluded.eigen`,
		spiel, regel.Muster, regel.Ersatz,
		regel.GanzesWort, regel.Aktiv, regel.Eigen, profilID); err != nil {
		return fmt.Errorf("die Autoersetzung setzen: %w", err)
	}

	return nil
}

// Loeschen nimmt eine Regel heraus.
//
// **Eine Regel, die es nicht gibt, ist ein Befund** (§8): wer sie löschen
// wollte, hat einen anderen Stand vor sich als der Bestand.
func (a *Autoersetzungsablage) Loeschen(
	ctx context.Context, spiel, muster string,
) error {
	if err := pruefeSpiel(spiel); err != nil {
		return err
	}

	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	ergebnis, err := a.db.ExecContext(ctx,
		`DELETE FROM spiel_autoersetzungen
		  WHERE spiel = ? AND muster = ? AND profil_id = ?`,
		spiel, muster, profilID)
	if err != nil {
		return fmt.Errorf("die Autoersetzung löschen: %w", err)
	}

	// **Über [betroffen]** — der Zählfehler wird dort übersetzt, die Aussage
	// bleibt hier: „dieses Spiel führt keine eigene Regel" ist etwas anderes
	// als „das Exemplar gibt es nicht".
	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return zaehlfehler
	}

	if zeilen == 0 {
		return fmt.Errorf("%w: dieses Spiel führt keine eigene Regel %q",
			domain.ErrNotFound, muster)
	}

	return nil
}

// Zuruecksetzen nimmt **alle** eigenen Zeilen eines Spiels heraus.
//
// **MSEs „defaults"** (`AutoReplaceWindow::onDefault`,
// `src/gui/auto_replace_window.cpp`, Zeilen 234–238, Fassung f9d9356 vom
// 2021-07-06): der Schalter geht an,
// die Liste wird die des Spiels. Hier heißt das: der Bestand vergisst seine
// Abweichungen, und übrig bleibt, was das Paket mitbringt.
func (a *Autoersetzungsablage) Zuruecksetzen(ctx context.Context, spiel string) error {
	if err := pruefeSpiel(spiel); err != nil {
		return err
	}

	// **In einem Zug.**
	//
	// **MSE behandelt es als eine Bedienhandlung** (`AutoReplaceWindow::onDefault`,
	// Zeilen 234–238): der Schalter geht an, die Liste wird die des Spiels.
	// Zwei Hälften davon sind kein Zustand, den MSE kennt.
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	transaktion, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("das Zurücksetzen beginnen: %w", err)
	}

	defer func() { _ = transaktion.Rollback() }()

	if _, err := transaktion.ExecContext(ctx,
		`DELETE FROM spiel_autoersetzungen WHERE spiel = ? AND profil_id = ?`,
		spiel, profilID); err != nil {
		return fmt.Errorf("die Autoersetzungen zurücksetzen: %w", err)
	}

	if _, err := transaktion.ExecContext(ctx,
		`DELETE FROM spiel_autoersetzung_an WHERE spiel = ? AND profil_id = ?`,
		spiel, profilID); err != nil {
		return fmt.Errorf("den Schalter zurücksetzen: %w", err)
	}

	// **Ein Satz und kein Zweig.** Dieselbe Form wie in
	// `ersetzungschreiber.go`: ein eigener Fehlerzweig fürs Abschließen wäre
	// einer, den kein Lauf betreten kann (§3).
	return uebersetze(transaktion.Commit(), "das Zurücksetzen abschließen")
}

// Schalter sagt, ob die Autoersetzung für dieses Spiel greift.
//
// **Ohne Zeile gilt „an"** — MSEs `use_auto_replace(true)`
// (`src/data/settings.cpp`, Zeile 74, Fassung f9d9356 vom 2021-07-06). Es ist die Vorgabe, die niemanden
// überrascht: die Regeln greifen schon heute.
func (a *Autoersetzungsablage) Schalter(ctx context.Context, spiel string) (bool, error) {
	if err := pruefeSpiel(spiel); err != nil {
		return false, err
	}

	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return false, profilfehler
	}

	var an bool

	err := a.db.QueryRowContext(ctx,
		`SELECT an FROM spiel_autoersetzung_an
		  WHERE spiel = ? AND profil_id = ?`, spiel, profilID).Scan(&an)

	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}

	if err != nil {
		return false, fmt.Errorf("den Schalter lesen: %w", err)
	}

	return an, nil
}

// SchalterSetzen schreibt den Schalter fort.
func (a *Autoersetzungsablage) SchalterSetzen(
	ctx context.Context, spiel string, an bool,
) error {
	if err := pruefeSpiel(spiel); err != nil {
		return err
	}

	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	if _, err := a.db.ExecContext(ctx,
		`INSERT INTO spiel_autoersetzung_an (spiel, an, profil_id)
		 VALUES (?, ?, ?)
		 ON CONFLICT (profil_id, spiel) DO UPDATE SET an = excluded.an`,
		spiel, an, profilID); err != nil {
		return fmt.Errorf("den Schalter setzen: %w", err)
	}

	return nil
}
