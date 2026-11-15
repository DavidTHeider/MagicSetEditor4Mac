package storage

// **Welcher Schritt als nächstes an der Reihe ist** — die andere Hälfte von
// `verlaufbewegen.go`, das sagt, *wie* ein Schritt geht.
//
// Getrennt, seit die gezielte Rücknahme dazukam
// (#1393):
// „welcher" ist seither eine eigene Frage mit zwei Antworten — der jüngste
// geltende Schritt des Sets, oder ein **benannter Vorgang**.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// naechsterSchritt sucht den Schritt, der als nächstes an der Reihe ist.
//
// **Ein Schritt, keine Zeile.** Die Abfrage sucht zuerst die maßgebliche Zeile
// und nimmt dann alle, die zu ihrem Schritt gehören — bei einer Einzeländerung
// ist das sie selbst, bei einer Massenänderung ihre ganze Gruppe. Die Rückgabe
// steht in aufsteigender Reihenfolge.
//
// # Und ein benannter Vorgang statt des jüngsten
//
// Ist `vorgang` gesetzt, sucht die innere Abfrage **innerhalb** dieser
// Gruppe. Das ist der ganze Unterschied zwischen ⌘Z am Fenster und einer
// Rücknahme über MCP, Kommandozeile oder REST: dort schreibt womöglich ein
// zweiter Prozess dazwischen, und „der jüngste" träfe dessen Änderung (#1393).
//
// **Der Zustand wird trotzdem geprüft.** Ein Vorgang, der schon
// zurückgenommen ist, kommt beim Zurücknehmen nicht noch einmal — die
// Bedingung `zurueckgenommen = ?` steht neben der Gruppe und nicht statt ihr.
func naechsterSchritt(
	ctx context.Context, tx *sql.Tx, setID int64, vorgang string, zurueck bool,
) ([]domain.Verlaufseintrag, error) {
	// **Zurück: der jüngste geltende. Vor: der älteste zurückgenommene.** So
	// bleibt die Reihenfolge eine Kette und keine Menge.
	gilt := 0
	if !zurueck {
		gilt = 1
	}

	werte := []any{setID, setID, gilt}
	if vorgang != "" {
		werte = append(werte, vorgang)
	}

	zeilen, err := tx.QueryContext(ctx, schrittabfrage(vorgang != "", zurueck), werte...)
	if err != nil {
		return nil, uebersetze(err, "den Verlauf lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen ist gelesen

	var schritt []domain.Verlaufseintrag

	for zeilen.Next() {
		eintrag, fehler := eintragAus(zeilen)
		if fehler != nil {
			return nil, fehler
		}

		schritt = append(schritt, eintrag)
	}

	if err := zeilen.Err(); err != nil {
		return nil, uebersetze(err, "den Verlauf lesen")
	}

	return schritt, nil
}

// schrittabfrage gibt eine der **vier** fertigen Abfragen.
//
// # Warum vier Konstanten statt einer Verkettung
//
// Die Abfrage variiert in zwei Achsen — Richtung und Vorgangsfilter —, und
// beide sind Konstanten. Zusammengesetzt am Aufruf sah `gosec` eine
// SQL-Verkettung (G202) und konnte nicht wissen, dass jeder Teil zur
// Übersetzungszeit feststeht. Ein `#nosec` daneben wäre die Antwort „glaub mir"
// gewesen; vier Konstanten sind die Antwort „sieh nach".
//
// **Der Vorgang selbst steht nie im Text**, sondern als Platzhalter — er ist
// der einzige Wert von außen, und er reist gebunden.
func schrittabfrage(nurDieser, zurueck bool) string {
	if nurDieser {
		if zurueck {
			return schrittZurueckMitVorgang
		}

		return schrittVorMitVorgang
	}

	if zurueck {
		return schrittZurueck
	}

	return schrittVor
}

// Die vier Abfragen. **Ein Rumpf, vier Enden** — was sich unterscheidet, sind
// die Sortierrichtung der inneren Abfrage und der Vorgangsfilter davor.
const (
	schrittrumpf = `SELECT ` + verlaufsspalten + `
	   FROM history
	  WHERE set_id = ?
	    AND ` + schrittkennung + ` = (
	      SELECT ` + schrittkennung + ` FROM history
	       WHERE set_id = ? AND zurueckgenommen = ?`

	vorgangsfilter = ` AND ` + schrittkennung + ` = ?`

	schrittZurueck = schrittrumpf + `
	       ORDER BY id DESC LIMIT 1)
	  ORDER BY id`

	schrittVor = schrittrumpf + `
	       ORDER BY id ASC LIMIT 1)
	  ORDER BY id`

	schrittZurueckMitVorgang = schrittrumpf + vorgangsfilter + `
	       ORDER BY id DESC LIMIT 1)
	  ORDER BY id`

	schrittVorMitVorgang = schrittrumpf + vorgangsfilter + `
	       ORDER BY id ASC LIMIT 1)
	  ORDER BY id`
)
