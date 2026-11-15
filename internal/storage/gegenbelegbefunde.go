package storage

// **Ein Gegenbeleg, der nicht passt, wird festgehalten** —
// #2500,
// Wanderung `0101_gegenbelegbefunde.sql`.

import (
	"context"
	"database/sql"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Abweisen hält einen abgewiesenen Gegenbeleg fest — **der jüngste gilt**.
//
// **Ersetzen und nicht anhängen**: ein Gegenüber, das immer neue Belege
// schickt, füllt sonst den Bestand. Was der vorige Befund sagte, steht im
// Protokoll des Dienstes, der abgewiesen hat.
//
// **Ohne eigenen Beleg kein Befund** — der Fremdschlüssel auf
// `handelsbelege` meldet „gibt es nicht", denn wogegen wäre geprüft worden?
func (a *Abschlussablage) Abweisen(
	ctx context.Context, befund domain.Gegenbelegbefund,
) error {
	if fehler := befund.Pruefe(); fehler != nil {
		return fehler
	}

	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	_, fehler = a.db.ExecContext(ctx,
		`INSERT INTO handelsgegenbelegbefunde
		   (profil_id, vorgang, von, abdruck, mangel, grund)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT (profil_id, vorgang) DO UPDATE SET
		   von = excluded.von, abdruck = excluded.abdruck,
		   mangel = excluded.mangel, grund = excluded.grund,
		   empfangen = datetime('now')`,
		profilID, befund.Vorgang, befund.Von, befund.Abdruck,
		string(befund.Mangel), befund.Grund)

	return uebersetze(fehler, "einen abgewiesenen Gegenbeleg festhalten")
}

// befundspalten sind die fünf Spalten des Befunds im Lesen des Belegs —
// **dasselbe Lesen**, aus demselben Grund wie beim Fremdbeleg.
type befundspalten struct {
	von, abdruck, mangel, grund, empfangen sql.NullString
}

// ziele gibt die Ziele für `Scan`, in der Reihenfolge der Abfrage.
func (s *befundspalten) ziele() []any {
	return []any{&s.von, &s.abdruck, &s.mangel, &s.grund, &s.empfangen}
}

// befund baut den Befund — `nil`, wenn keiner da ist.
func (s befundspalten) befund(vorgang string) (*domain.Gegenbelegbefund, error) {
	if !s.mangel.Valid {
		return nil, nil //nolint:nilnil // „keiner abgewiesen" ist kein Fehler
	}

	empfangen, fehler := time.Parse(time.DateTime, s.empfangen.String)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Uhr eines Befunds lesen")
	}

	return &domain.Gegenbelegbefund{
		Vorgang: vorgang, Von: s.von.String, Abdruck: s.abdruck.String,
		Mangel: domain.Gegenbelegmangel(s.mangel.String), Grund: s.grund.String,
		Empfangen: empfangen,
	}, nil
}
