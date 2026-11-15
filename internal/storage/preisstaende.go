// Die **Preisstände** im Bestand
// (#1165).
//
// # Warum eine Zeile je Abruf und kein Feld
//
// Ein Preis altert, ein Preis hat eine Herkunft, und aus Momentaufnahmen wird
// ein Verlauf — ein Feld an einem Exemplar könnte keines der drei
// ([0052](../../docs/entscheidungen/0052-die-sammlung-ist-nicht-das-set.md)).

package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Preisspeicher liest und schreibt die Preisstände.
type Preisspeicher struct {
	db *sql.DB
}

// NeuerPreisspeicher baut ihn über einer Datenbank.
func NeuerPreisspeicher(db *sql.DB) *Preisspeicher {
	return &Preisspeicher{db: db}
}

// preisspalten sind die Spalten in der Reihenfolge, in der gelesen wird.
const preisspalten = `id, printing, druck_id, quelle, markt, art, waehrung, finish,
       zustand, sprache, lage, wert, beobachtet, erhoben`

// preisreihe sind die Spalten, die **eine Reihe** ausmachen (A3, #1442).
//
// **Als eigene Zeichenkette, weil drei Abfragen sie brauchen** — und weil eine
// vergessene Spalte darin zwei Reihen zu einer machte, ohne dass es auffiele.
const preisreihe = `quelle, markt, art, waehrung, finish, zustand, sprache`

// Eintragen legt die Preisstände **eines Abrufs** ab — alle oder keinen.
//
// # Warum als Zug und nicht einzeln
//
// Ein Abruf bringt bis zu sieben Preise. Der erste Anlauf schrieb sie
// nacheinander mit je einem `ExecContext`; scheiterte der vierte, blieben die
// ersten drei als neuer Stand stehen — ein halber Abruf, den niemand als
// solchen erkennt. Angestrichen vom Zweitblick über Phase 3.4.
//
// **Derselbe Abruf zweimal ist kein Fehler, sondern derselbe Stand.** Die
// Tabelle ist eindeutig über Druck, Quelle, Währung, Ausführung **und**
// Zeitpunkt; ein zweiter Eintrag in derselben Sekunde wird übergangen. Zwei
// Abrufe am selben Tag sind dagegen zwei Zeilen — der Verlauf soll sagen, wann
// sich etwas geändert hat.
func (p *Preisspeicher) Eintragen(
	ctx context.Context, staende ...domain.Preisstand,
) error {
	if len(staende) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "die Preisstände eintragen")
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach einem Commit gibt es nichts zurückzunehmen

	for _, stand := range staende {
		druck, druckfehler := druckZeile(ctx, tx, druckbezug{printing: stand.Printing})
		if druckfehler != nil {
			return druckfehler
		}

		if _, fehler := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO preisstaende
			   (printing, druck_id, quelle, markt, art, waehrung, finish,
			    zustand, sprache, lage, wert, beobachtet, erhoben)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			stand.Printing, druck, string(stand.Quelle), string(stand.Markt),
			string(stand.Art), stand.Waehrung, string(stand.Finish),
			stand.Zustand, stand.Sprache, string(stand.Lage), stand.Wert,
			zeitpunktOderNull(stand.Beobachtet),
			stand.Erhoben.Unix()); fehler != nil {
			return uebersetze(fehler, "einen Preisstand eintragen")
		}
	}

	return uebersetze(tx.Commit(), "die Preisstände eintragen")
}

// Neueste gibt den jüngsten Stand je Währung und Ausführung eines Drucks.
//
// # Warum das in SQL steht und nicht in Go
//
// „Der neueste je Gruppe" ist eine Frage an die Datenbank; wer alle Stände
// holt und in Go filtert, liest bei einem Jahr Verlauf ein paar hundert Zeilen
// für eine Antwort aus vieren. Der Index `preisstaende_neueste` steht genau
// dafür da.
func (p *Preisspeicher) Neueste(ctx context.Context, printing string) (
	[]domain.Preisstand, error,
) {
	return p.lesen(ctx,
		`SELECT `+preisspalten+`
		   FROM preisstaende
		  WHERE printing = ?
		    AND erhoben = (SELECT max(erhoben) FROM preisstaende AS neuer
		                    WHERE neuer.printing = preisstaende.printing`+
			gleicheReihe+`)
		  ORDER BY `+preisreihe, printing)
}

// NeuesteAlle gibt die jüngsten Stände **aller** Drucke.
//
// # Warum es sie neben [Preisspeicher.Neueste] gibt
//
// Weil das Portfolio sonst einmal je Exemplar fragt. Gemessen am 2026-08-30 an
// 5.000 Exemplaren **ohne einen einzigen Preis**: 89 ms für das
// Zusammenrechnen gegen 5 ms für das Laden derselben Zeilen — das
// Achtzehnfache, und es wächst mit dem Bestand
// (#1292).
//
// # Was das an Speicher kostet
//
// Diese Abfrage holt die Preistabelle in einem Zug. Sie ist **nicht** so groß
// wie ein Kartenkatalog: eingetragen wird nur, was jemand geholt hat, und je
// Druck sind es so viele Zeilen wie Währungen mal Ausführungen. Bei 5.000
// Drucken in zwei Währungen und zwei Ausführungen sind das 20.000 Zeilen —
// dieselbe Größenordnung wie die Sammlung selbst, die ohnehin ganz geladen
// wird.
//
// **Wo die Grenze läge**, steht bei der Sicht im Fenster
// (`frontend/src/seiten/sammlungssicht.ts`) und gilt hier genauso: wer eine
// Sammlung führt, die nicht in den Arbeitsspeicher passt, braucht eine Summe
// im SQL statt einer Schleife in Go.
func (p *Preisspeicher) NeuesteAlle(ctx context.Context) (
	[]domain.Preisstand, error,
) {
	return p.lesen(ctx,
		`SELECT `+preisspalten+`
		   FROM preisstaende
		  WHERE erhoben = (SELECT max(erhoben) FROM preisstaende AS neuer
		                    WHERE neuer.printing = preisstaende.printing`+
			gleicheReihe+`)
		  ORDER BY printing, `+preisreihe)
}

// Verlauf gibt alle Stände einer Reihe — ältester zuerst.
//
// **Ältester zuerst**, weil ein Verlauf von links nach rechts gelesen wird.
func (p *Preisspeicher) Verlauf(
	ctx context.Context, printing string, reihe domain.Preisstand,
) ([]domain.Preisstand, error) {
	return p.lesen(ctx,
		`SELECT `+preisspalten+`
		   FROM preisstaende
		  WHERE printing = ? AND quelle = ? AND markt = ? AND art = ?
		    AND waehrung = ? AND finish = ? AND zustand = ? AND sprache = ?
		  ORDER BY erhoben`,
		printing, string(reihe.Quelle), string(reihe.Markt), string(reihe.Art),
		reihe.Waehrung, string(reihe.Finish), reihe.Zustand, reihe.Sprache)
}

// gleicheReihe hält zwei Zeilen derselben Reihe zusammen.
//
// **Als Zeichenkette und nicht siebenmal getippt**: eine vergessene Spalte
// machte aus zwei Reihen eine, und der „neueste Preis" wäre dann der neueste
// von beiden (A3, #1442).
const gleicheReihe = `
		                      AND neuer.quelle   = preisstaende.quelle
		                      AND neuer.markt    = preisstaende.markt
		                      AND neuer.art      = preisstaende.art
		                      AND neuer.waehrung = preisstaende.waehrung
		                      AND neuer.finish   = preisstaende.finish
		                      AND neuer.zustand  = preisstaende.zustand
		                      AND neuer.sprache  = preisstaende.sprache`

// zeitpunktOderNull gibt Unix-Sekunden — null für einen Zeitpunkt, den es
// nicht gibt.
//
// **`time.Time{}.Unix()` ist nicht null**, sondern −62135596800. Wer das
// speicherte, schriebe „im Jahr 1" statt „kein Datum".
func zeitpunktOderNull(zeitpunkt time.Time) int64 {
	if zeitpunkt.IsZero() {
		return 0
	}

	return zeitpunkt.Unix()
}

// beobachtungAus liest, was [zeitpunktOderNull] geschrieben hat.
func beobachtungAus(sekunden int64) time.Time {
	if sekunden == 0 {
		return time.Time{}
	}

	return time.Unix(sekunden, 0).UTC()
}

// lesen führt eine Abfrage aus und baut die Stände.
func (p *Preisspeicher) lesen(ctx context.Context, abfrage string, werte ...any) (
	[]domain.Preisstand, error,
) {
	zeilen, err := p.db.QueryContext(ctx, abfrage, werte...)
	if err != nil {
		return nil, uebersetze(err, "Preisstände lesen")
	}

	defer func() { _ = zeilen.Close() }()

	staende := []domain.Preisstand{}

	for zeilen.Next() {
		var (
			stand      domain.Preisstand
			druck      sql.NullInt64
			beobachtet int64
			erhoben    int64
		)

		if err := zeilen.Scan(&stand.ID, &stand.Printing, &druck, &stand.Quelle,
			&stand.Markt, &stand.Art, &stand.Waehrung, &stand.Finish,
			&stand.Zustand, &stand.Sprache, &stand.Lage, &stand.Wert,
			&beobachtet, &erhoben); err != nil {
			return nil, uebersetze(err, "einen Preisstand lesen")
		}

		// **In UTC zurück und nicht in der Zone dieses Rechners.** Was
		// gespeichert ist, sind Sekunden seit 1970; eine Ortszeit daraus wäre
		// eine Angabe, die auf einem anderen Rechner anders lautet.
		stand.Erhoben = time.Unix(erhoben, 0).UTC()
		stand.Beobachtet = beobachtungAus(beobachtet)
		stand.DruckID = druck.Int64

		staende = append(staende, stand)
	}

	return staende, uebersetze(zeilen.Err(), "Preisstände lesen")
}
