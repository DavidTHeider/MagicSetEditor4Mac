package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Marken ist die Ablage der Marken an Karten
// (#1208).
//
// **Eine eigene Ablage und kein Feld am Kartenschreiber**: eine Marke ist kein
// Feldwert, geht in keine `.mse-set` und steht in keinem Verlauf der
// Feldänderungen. Wer sie dort führte, schriebe sie in jede weitergegebene
// Datei — und verlöre sie beim Rundlauf durch MSE2
// ([0053](../../docs/entscheidungen/0053-marken-liegen-neben-dem-set.md)).
type Marken struct {
	db *sql.DB
}

// NeueMarken gibt die Ablage über dieser Datenbank.
func NeueMarken(db *sql.DB) *Marken {
	return &Marken{db: db}
}

// Setzen hängt eine Marke an eine Karte.
//
// **Zweimal dieselbe ist kein Fehler, sondern nichts.** Wer eine Marke setzt,
// meint „sie soll dranstehen"; dass sie schon dranstand, ist kein Grund für
// eine Meldung.
func (m *Marken) Setzen(ctx context.Context, karteID int64, marke string) error {
	// **Nur an einer Karte des gezeigten Profils** — geprüft im selben Zug wie
	// das Schreiben (#1815).
	return imGeschaeft(ctx, m.db, "eine Marke setzen", func(tx *sql.Tx) error {
		if fehler := karteImProfil(ctx, tx, karteID); fehler != nil {
			return fehler
		}

		_, err := tx.ExecContext(ctx,
			`INSERT INTO karten_marken (card_id, marke) VALUES (?, ?)
			 ON CONFLICT (card_id, marke) DO NOTHING`,
			karteID, marke)

		return uebersetze(err, "eine Marke setzen")
	})
}

// Entfernen nimmt eine Marke wieder ab.
//
// **Eine Marke, die gar nicht dastand, ist ein Befund** — anders als beim
// Setzen. Wer entfernt, meint eine bestimmte; stand sie nicht da, hat der Ruf
// eine andere gemeint, und eine stille Zustimmung verbärge den Irrtum (§8).
func (m *Marken) Entfernen(ctx context.Context, karteID int64, marke string) error {
	return imGeschaeft(ctx, m.db, "eine Marke entfernen", func(tx *sql.Tx) error {
		if fehler := karteImProfil(ctx, tx, karteID); fehler != nil {
			return fehler
		}

		ergebnis, err := tx.ExecContext(ctx,
			`DELETE FROM karten_marken WHERE card_id = ? AND marke = ?`, karteID, marke)
		if err != nil {
			return uebersetze(err, "eine Marke entfernen")
		}

		// **Über [betroffen]** — die Übersetzung des Zählfehlers ist geteilt,
		// der Satz darunter nicht: „die Karte trägt keine Marke %q" ist eine
		// Auskunft über **diese** Handlung.
		getroffen, zaehlfehler := betroffen(ergebnis)
		if zaehlfehler != nil {
			return zaehlfehler
		}

		if getroffen == 0 {
			return fmt.Errorf("%w: die Karte %d trägt keine Marke %q",
				domain.ErrNotFound, karteID, marke)
		}

		return nil
	})
}

// AnKarte gibt die Marken einer Karte, alphabetisch.
func (m *Marken) AnKarte(ctx context.Context, karteID int64) ([]string, error) {
	if fehler := karteImProfil(ctx, m.db, karteID); fehler != nil {
		return nil, fehler
	}

	return m.woerter(ctx,
		`SELECT marke FROM karten_marken WHERE card_id = ? ORDER BY marke`, karteID)
}

// ImSet gibt die Marken, die in einem Set vorkommen — jede einmal.
//
// **Je Set und nicht über den ganzen Bestand**: die Auswahl steht neben der
// Kartenliste **eines** Sets, und eine Marke, die dort nichts trifft, wäre ein
// Eintrag, der immer eine leere Liste ergibt.
func (m *Marken) ImSet(ctx context.Context, setID int64) ([]string, error) {
	if fehler := setImProfil(ctx, m.db, setID); fehler != nil {
		return nil, fehler
	}

	return m.woerter(ctx,
		`SELECT DISTINCT k.marke FROM karten_marken k
		 JOIN cards c ON c.id = k.card_id
		 WHERE c.set_id = ? ORDER BY k.marke`, setID)
}

// JeKarte gibt die Marken aller Karten eines Sets — in **einem** Ruf.
//
// **Für die Kartenliste**: eine Abfrage je Zeile wären bei 3.788 Karten 3.788
// Abfragen, und die Liste ist der Weg zu jeder Karte.
func (m *Marken) JeKarte(ctx context.Context, setID int64) (map[int64][]string, error) {
	if fehler := setImProfil(ctx, m.db, setID); fehler != nil {
		return nil, fehler
	}

	return m.jeKarte(ctx,
		`SELECT k.card_id, k.marke FROM karten_marken k
		 JOIN cards c ON c.id = k.card_id
		 WHERE c.set_id = ? ORDER BY k.card_id, k.marke`, setID)
}

// jeKarte führt eine Abfrage aus, die Kartenkennung und Marke gibt.
//
// **Die Abfrage kommt als Parameter** — dieselbe Bauform wie
// `Sammlungsspeicher.lesen`, und aus demselben Grund: der `Scan`-Fehlerweg
// lässt sich sonst von niemandem herbeiführen, und dann fiele es nicht auf,
// wenn seine Meldung verloren ginge.
func (m *Marken) jeKarte(
	ctx context.Context, abfrage string, werte ...any,
) (map[int64][]string, error) {
	zeilen, err := m.db.QueryContext(ctx, abfrage, werte...)
	if err != nil {
		return nil, uebersetze(err, "die Marken des Sets lesen")
	}

	defer func() { _ = zeilen.Close() }()

	jeKarte := map[int64][]string{}

	for zeilen.Next() {
		var (
			karteID int64
			marke   string
		)

		if err := zeilen.Scan(&karteID, &marke); err != nil {
			return nil, uebersetze(err, "eine Marke lesen")
		}

		jeKarte[karteID] = append(jeKarte[karteID], marke)
	}

	return jeKarte, uebersetze(zeilen.Err(), "die Marken des Sets lesen")
}

// woerter führt eine Abfrage aus, die eine Spalte mit Marken gibt.
func (m *Marken) woerter(ctx context.Context, abfrage string, werte ...any) ([]string, error) {
	zeilen, err := m.db.QueryContext(ctx, abfrage, werte...)
	if err != nil {
		return nil, uebersetze(err, "Marken lesen")
	}

	defer func() { _ = zeilen.Close() }()

	marken := []string{}

	for zeilen.Next() {
		var marke string
		if err := zeilen.Scan(&marke); err != nil {
			return nil, uebersetze(err, "eine Marke lesen")
		}

		marken = append(marken, marke)
	}

	return marken, uebersetze(zeilen.Err(), "Marken lesen")
}
