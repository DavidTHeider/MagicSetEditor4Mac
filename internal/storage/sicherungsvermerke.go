package storage

// **Wann der Bestand gesichert wurde** — M01
// (#2345),
// Wanderung `0098_sicherungsvermerke.sql`. Wer vermerkt, entscheidet der
// Aufrufer; was ein Vermerk ist, [domain.Sicherungsvermerk].

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// vermerkzeit ist die Form der Zeit in der Tafel — dieselbe wie
// `strftime('%Y-%m-%dT%H:%M:%fZ')` der übrigen Tafeln, damit sie sich als
// Zeichenkette ordnet wie die Zeit selbst.
const vermerkzeit = "2006-01-02T15:04:05.000Z"

// Sicherungsvermerke führt die Vermerke eines Bestands.
type Sicherungsvermerke struct {
	db *sql.DB
}

// NeueSicherungsvermerke gibt die Vermerke von db.
func NeueSicherungsvermerke(db *sql.DB) *Sicherungsvermerke {
	return &Sicherungsvermerke{db: db}
}

// Vermerken trägt einen Vermerk ein — wenn er die Prüfung der Domäne besteht.
func (s *Sicherungsvermerke) Vermerken(ctx context.Context, v domain.Sicherungsvermerk) error {
	if fehler := v.Pruefe(); fehler != nil {
		return fehler
	}

	_, fehler := s.db.ExecContext(ctx,
		`INSERT INTO sicherungsvermerke
		   (gesichert, anlass, flaeche, datei, groesse, pruefsumme, verschluesselt)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		v.Gesichert.UTC().Format(vermerkzeit), string(v.Anlass), string(v.Flaeche),
		v.Datei, v.Groesse, v.Pruefsumme, v.Verschluesselt)

	return uebersetze(fehler, "die Sicherung vermerken")
}

// Letzter gibt den jüngsten Vermerk einer Sicherung, die man wegtragen kann —
// und falsch, wenn es keinen gibt.
//
// **Unbekannt heißt unbekannt** (OPS01, #2347): ein Bestand, der nie über
// diese Fassung gesichert wurde, nennt keinen Zeitpunkt, statt einen zu
// erfinden — etwa den seiner Anlage.
//
// **Ohne die Sicherung vor einer Wanderung**: sie liegt neben dem Bestand und
// ohne Bilder, hilft gegen eine missglückte Wanderung und nicht gegen eine
// kaputte Platte. Zählte sie mit, stünde nach jeder neuen Fassung „frisch" da.
func (s *Sicherungsvermerke) Letzter(ctx context.Context) (domain.Sicherungsvermerk, bool, error) {
	vermerke, fehler := s.lesen(ctx, "den letzten Sicherungsvermerk lesen",
		`SELECT gesichert, anlass, flaeche, datei, groesse, pruefsumme, verschluesselt
		 FROM sicherungsvermerke
		 WHERE anlass IN (?, ?) ORDER BY gesichert DESC, id DESC LIMIT 1`,
		string(domain.SicherungVonHand), string(domain.SicherungNachPlan))
	if fehler != nil || len(vermerke) == 0 {
		return domain.Sicherungsvermerk{}, false, fehler
	}

	return vermerke[0], true, nil
}

// Vermerke gibt die Vermerke eines Anlasses, den jüngsten zuerst — für den
// Plan, der seine eigenen Dateien an ihnen wiedererkennt (OPS01, #2347).
func (s *Sicherungsvermerke) Vermerke(
	ctx context.Context, anlass domain.Sicherungsanlass,
) ([]domain.Sicherungsvermerk, error) {
	return s.lesen(ctx, "die Sicherungsvermerke lesen",
		`SELECT gesichert, anlass, flaeche, datei, groesse, pruefsumme, verschluesselt
		 FROM sicherungsvermerke
		 WHERE anlass = ? ORDER BY gesichert DESC, id DESC`, string(anlass))
}

// lesen holt die Vermerke einer Abfrage — **ganz ausgeschrieben je
// Aufrufer** und nicht aus Teilen zusammengesetzt: eine Abfrage, die aus
// Teilen entsteht, prüft niemand mehr als Ganzes.
func (s *Sicherungsvermerke) lesen(
	ctx context.Context, was, abfrage string, argumente ...any,
) ([]domain.Sicherungsvermerk, error) {
	zeilen, fehler := s.db.QueryContext(ctx, abfrage, argumente...)
	if fehler != nil {
		return nil, uebersetze(fehler, was)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // nur gelesen

	var vermerke []domain.Sicherungsvermerk

	for zeilen.Next() {
		var (
			v                          domain.Sicherungsvermerk
			gesichert, anlass, flaeche string
		)

		// **Lesen und Zeit auf einem Weg**: die Tafel ist `STRICT` und jede
		// Spalte `NOT NULL`, ein Fehler beim Lesen ließe sich nicht herstellen
		// — ein unlesbarer Zeitpunkt schon (Monat 13 besteht das `GLOB`).
		lesefehler := zeilen.Scan(&gesichert, &anlass, &flaeche,
			&v.Datei, &v.Groesse, &v.Pruefsumme, &v.Verschluesselt)

		zeit, zeitfehler := time.Parse(vermerkzeit, gesichert)
		if fehler := errors.Join(lesefehler, zeitfehler); fehler != nil {
			return nil, uebersetze(fehler, "einen Sicherungsvermerk lesen")
		}

		v.Gesichert, v.Anlass, v.Flaeche = zeit, domain.Sicherungsanlass(anlass), domain.Flaeche(flaeche)
		vermerke = append(vermerke, v)
	}

	return vermerke, uebersetze(zeilen.Err(), was)
}
