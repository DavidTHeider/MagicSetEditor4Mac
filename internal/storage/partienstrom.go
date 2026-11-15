package storage

// **Der Strom einer Partie** — Werkstück G1 von Phase 5.5
// (#1458),
// herausgelöst mit G5
// (#1462),
// als `partien.go` über die Zeilengrenze aus §2 ging.
//
// **Der Schnitt folgt der Frage**: dort steht, was eine Partie **ist** —
// Zustand, Sitze, Objekte —, hier, was ihr **geschehen** ist. Das ist nicht
// dieselbe Frage, auch wenn beides in derselben Datenbank liegt: der Zustand
// wird überschrieben, der Strom wächst und wird nie geändert.

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Befehlsergebnis gibt die Ereignisse eines schon angewendeten Befehls.
func (a *Partieablage) Befehlsergebnis(
	ctx context.Context, partieID, befehlID string,
) ([]domain.Ereignis, bool, error) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := partieImProfil(ctx, a.db, partieID); fehler != nil {
		return nil, false, fehler
	}

	var von, bis int64

	fehler := a.db.QueryRowContext(ctx,
		`SELECT von_folge, bis_folge FROM partiebefehle
		 WHERE partie_id = ? AND kennung = ?`, partieID, befehlID).
		Scan(&von, &bis)

	if errors.Is(fehler, sql.ErrNoRows) {
		return nil, false, nil
	}

	if fehler != nil {
		return nil, false, uebersetze(fehler, "den Befehl nachschlagen")
	}

	ereignisse, fehler := a.ereignisse(ctx, partieID, von, bis)

	return ereignisse, true, fehler
}

// Ereignisse gibt einen Ausschnitt des Stroms — beide Grenzen einschließlich.
//
// **`0, 0` heißt: alles.** So steht es am Vertrag in `service.Partiespeicher`,
// und hier wird es eingelöst. Der erste Anlauf setzte beide Grenzen wörtlich
// ein — `folge >= 0 AND folge <= 0` trifft nie etwas, denn die Folge fängt bei
// eins an. Die Rückschau war damit an **jeder** Fläche leer, und zwar **ohne
// Fehler**: eine Zusage, die nur im Kopf einer Schnittstelle steht, hält
// niemand (§10). Angestrichen hat es
// `internal/flaechen/partieverlauf_test.go`.
//
// **Nur die obere Grenze wird geöffnet**, nicht beide: `von` ist bei einem
// Ausschnitt schon die untere Schranke, und eine Null dort meint dasselbe wie
// eine Eins.
func (a *Partieablage) Ereignisse(
	ctx context.Context, partieID string, von, bis int64,
) ([]domain.Ereignis, error) {
	if fehler := partieImProfil(ctx, a.db, partieID); fehler != nil {
		return nil, fehler
	}

	return a.ereignisse(ctx, partieID, von, bis)
}

// ereignisse liest den Strom ohne erneute Frage nach dem Profil — für
// [Partieablage.Befehlsergebnis], das sie schon gestellt hat.
func (a *Partieablage) ereignisse(
	ctx context.Context, partieID string, von, bis int64,
) ([]domain.Ereignis, error) {
	if bis <= 0 {
		bis = math.MaxInt64
	}

	zeilen, fehler := a.db.QueryContext(ctx,
		`SELECT folge, art, oeffentlich, privat, kette
		 FROM partieereignisse
		 WHERE partie_id = ? AND folge >= ? AND folge <= ?
		 ORDER BY folge`, partieID, von, bis)
	if fehler != nil {
		return nil, uebersetze(fehler, "den Strom lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen ist gelesen

	gefunden := []domain.Ereignis{}

	for zeilen.Next() {
		ereignis := domain.Ereignis{PartieID: partieID}

		var oeffentlich, privat string

		if fehler := zeilen.Scan(&ereignis.Folge, &ereignis.Art,
			&oeffentlich, &privat, &ereignis.Kette); fehler != nil {
			return nil, uebersetze(fehler, "ein Ereignis lesen")
		}

		if fehler := json.Unmarshal(
			[]byte(oeffentlich), &ereignis.Oeffentlich); fehler != nil {
			return nil, domain.MitGrund(domain.ErrValidation, fehler,
				"ein Ereignis lesen")
		}

		if fehler := json.Unmarshal(
			[]byte(privat), &ereignis.Privat); fehler != nil {
			return nil, domain.MitGrund(domain.ErrValidation, fehler,
				"eine Ergänzung lesen")
		}

		gefunden = append(gefunden, ereignis)
	}

	return gefunden, uebersetze(zeilen.Err(), "den Strom lesen")
}

// LetzterBefehl gibt die Ereignisse des zuletzt angewendeten Befehls — P8
// (#2376).
//
// **Aus `partiebefehle` und nicht aus dem letzten Ereignis**: ein Befehl
// kann mehrere schreiben, und die Rücknahme muss wissen, ob es einer war.
func (a *Partieablage) LetzterBefehl(
	ctx context.Context, partieID string,
) ([]domain.Ereignis, error) {
	if fehler := partieImProfil(ctx, a.db, partieID); fehler != nil {
		return nil, fehler
	}

	var von, bis int64

	fehler := a.db.QueryRowContext(ctx,
		`SELECT von_folge, bis_folge FROM partiebefehle
		 WHERE partie_id = ? ORDER BY bis_folge DESC LIMIT 1`, partieID).
		Scan(&von, &bis)

	// **Eine Partie ohne Befehl ist kein Fehler**, sondern der Anfang.
	if errors.Is(fehler, sql.ErrNoRows) {
		return []domain.Ereignis{}, nil
	}

	if fehler != nil {
		return nil, uebersetze(fehler, "den letzten Befehl nachschlagen")
	}

	return a.ereignisse(ctx, partieID, von, bis)
}

// LetzteKette gibt den Abdruck und die Folge des letzten Ereignisses.
func (a *Partieablage) LetzteKette(
	ctx context.Context, partieID string,
) (abdruck string, folge int64, fehler error) {
	if besitzfehler := partieImProfil(ctx, a.db, partieID); besitzfehler != nil {
		return "", 0, besitzfehler
	}

	var (
		kette    sql.NullString
		gefunden sql.NullInt64
	)

	fehler = a.db.QueryRowContext(ctx,
		`SELECT kette, folge FROM partieereignisse
		 WHERE partie_id = ? ORDER BY folge DESC LIMIT 1`, partieID).
		Scan(&kette, &gefunden)

	// **Eine Partie ohne Ereignis ist kein Fehler**, sondern der Anfang.
	if errors.Is(fehler, sql.ErrNoRows) {
		return "", 0, nil
	}

	if fehler != nil {
		return "", 0, uebersetze(fehler, "die Kette lesen")
	}

	return kette.String, gefunden.Int64, nil
}
