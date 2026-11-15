package storage

// **Abmelden und Löschen am Turnier** — abgetrennt von `turniere.go`, als die
// Datei mit der Profilklemme (#1351) über die Zeilengrenze aus §2 kam.
//
// **Nach der Frage geschnitten, nicht nach der Zeilenzahl**: hier steht, wie
// ein Turnier **kleiner** wird — ein Teilnehmer geht, oder das Ganze geht.
// Was es aufbaut, steht nebenan.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Aussteigen meldet einen Teilnehmer ab — oder wieder an.
func (a *Turnierablage) Aussteigen(
	ctx context.Context, teilnehmerID int64, ausgestiegen bool,
) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	zug, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "das Abmelden beginnen")
	}
	defer zug.Rollback() //nolint:errcheck // nach einem Commit ist es ein No-op

	// **Nur in einem Turnier des gezeigten Profils** (#1815). Die Fassung
	// darunter klemmte schon — auf das Turnier, nicht auf den Teilnehmer:
	// sie schrieb nichts fort, und der Teilnehmer war trotzdem ausgestiegen.
	//
	// **Sie ist zugleich die Frage, ob es ihn gibt.** Bis dahin zählte das
	// Schreiben die getroffenen Zeilen; nach dieser Frage im selben Zug trifft
	// es immer genau eine.
	if fehler := teilnehmerImProfil(ctx, zug, teilnehmerID); fehler != nil {
		return fehler
	}

	if _, schreibfehler := zug.ExecContext(ctx,
		`UPDATE turnierteilnehmer SET ausgestiegen = ? WHERE id = ?`,
		ausgestiegen, teilnehmerID); schreibfehler != nil {
		return uebersetze(schreibfehler, "den Teilnehmer abmelden")
	}

	// **Der Ausstieg ist eine Änderung am Turnier**, und die Fassung zählt
	// sie mit — im selben Zug, sonst zählte sie manchmal.
	var (
		turnierID int64
		name      string
	)

	if lesefehler := zug.QueryRowContext(ctx,
		`SELECT turnier_id, name FROM turnierteilnehmer WHERE id = ?`, teilnehmerID,
	).Scan(&turnierID, &name); lesefehler != nil {
		return uebersetze(lesefehler, "das Turnier des Teilnehmers lesen")
	}

	eintrag := domain.Protokolleintrag{Vorgang: domain.TurnierAusgestiegen, Angaben: name}
	if !ausgestiegen {
		eintrag.Vorgang = domain.TurnierEingestiegen
	}

	if fortschrittfehler := fortschreiben(ctx, zug, turnierID, profilID, eintrag); fortschrittfehler != nil {
		return fortschrittfehler
	}

	return uebersetze(zug.Commit(), "das Abmelden abschließen")
}

// Loeschen nimmt ein Turnier samt seinem Verlauf heraus.
//
// **Der Verlauf geht mit** — über `ON DELETE CASCADE`. Ein Turnier ohne seine
// Runden wäre kein Turnier, sondern ein Name.
func (a *Turnierablage) Loeschen(ctx context.Context, kennung int64) error {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return profilfehler
	}

	ergebnis, err := a.db.ExecContext(ctx,
		`DELETE FROM turniere WHERE id = ? AND profil_id = ?`, kennung, profilID)
	if err != nil {
		return uebersetze(err, "das Turnier löschen")
	}

	zeilen, zaehlfehler := betroffen(ergebnis)
	if zaehlfehler != nil {
		return zaehlfehler
	}

	if zeilen == 0 {
		return fmt.Errorf(
			"%w: ein Turnier mit dieser Kennung gibt es nicht", domain.ErrNotFound)
	}

	return nil
}
