// **Der Leser der Sammlung** — die eine Stelle, an der aus Zeilen Exemplare
// werden.
//
// **Eine eigene Datei, weil `sammlung.go` an der Zeilengrenze steht** (§2),
// und der Schnitt folgt der Frage: dort die Abfragen, hier das Bauen.

package storage

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// lesen führt eine Abfrage aus und baut die Exemplare.
func (s *Sammlungsspeicher) lesen(ctx context.Context, abfrage string, werte ...any) (
	[]*domain.Exemplar, error,
) {
	return s.lesenBei(ctx, s.db, abfrage, werte...)
}

// lesenBei ist dasselbe über einer **genannten** Abfragestelle.
//
// **Damit zwei Abfragen denselben Stand sehen können** — Werkstück O11 von
// Phase 5.2, Befund 13 des Zweitblicks: ein Fenster liest erst die Ordnung und
// dann die Zeilen, und ohne gemeinsames Geschäft beschreiben beide womöglich
// verschiedene Stände des Bestands.
func (s *Sammlungsspeicher) lesenBei(
	ctx context.Context, bei abfrager, abfrage string, werte ...any,
) ([]*domain.Exemplar, error) {
	zeilen, err := bei.QueryContext(ctx, abfrage, werte...)
	if err != nil {
		return nil, uebersetze(err, "die Sammlung lesen")
	}

	defer func() { _ = zeilen.Close() }()

	// **Eine leere Liste und kein `nil`.** Über die Bindungen käme sonst ein
	// `null` heraus, wo die Oberfläche eine Liste erwartet (§3).
	exemplare := []*domain.Exemplar{}

	for zeilen.Next() {
		exemplar := &domain.Exemplar{}

		var (
			art   string
			druck sql.NullInt64
		)

		if err := zeilen.Scan(&exemplar.ID, &exemplar.Fassung,
			&exemplar.Printing, &druck, &exemplar.Menge,
			&exemplar.Zustand, &exemplar.Sprache, &exemplar.Finish,
			&exemplar.Erwerbsdatum, &exemplar.Erwerbspreis,
			&exemplar.Erwerbswaehrung, &exemplar.Notiz,
			&exemplar.Name, &exemplar.Setcode, &exemplar.Nummer,
			&exemplar.Lagerort, &exemplar.Wunsch,
			&exemplar.Wunschrang, &exemplar.Wunschart,
			&art, &exemplar.Herkunft.Quelle, &exemplar.Herkunft.QuellProfil,
			&exemplar.Herkunft.Uebernommen); err != nil {
			return nil, uebersetze(err, "ein Exemplar lesen")
		}

		exemplar.Herkunft.Art = domain.Herkunftsart(art)
		exemplar.DruckID = druck.Int64

		exemplare = append(exemplare, exemplar)
	}

	return exemplare, uebersetze(zeilen.Err(), "die Sammlung lesen")
}
