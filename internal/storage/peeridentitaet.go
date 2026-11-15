package storage

// **Der Vermerk über die eigene Peerkennung** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum die Kennung im Bestand steht und der Schlüssel nicht
//
// **Damit ein Verlust auffällt.** Der Grund steht ausführlich in der Wanderung
// `0061_peeridentitaet.sql`; kurz: der Schlüsselbund allein könnte nicht
// sagen, ob dieses Profil je eine Identität hatte.

import (
	"context"
	"database/sql"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Peerablage hält den Vermerk je Profil.
type Peerablage struct{ db *sql.DB }

// NeuePeerablage baut sie über einer Datenbank.
func NeuePeerablage(db *sql.DB) *Peerablage { return &Peerablage{db: db} }

// Lesen gibt den Vermerk dieses Profils — leer, wenn es keinen gibt.
//
// **Kein `ErrNotFound`**: „dieses Profil hatte noch nie eine Kennung" ist eine
// Auskunft und kein Fehlschlag. Wer sie als Fehler bekäme, müsste ihn beim
// ersten Start jedes Profils wegfangen.
func (p *Peerablage) Lesen(ctx context.Context) (domain.Peervermerk, error) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return domain.Peervermerk{}, fehler
	}

	var (
		kennung string
		seit    string
		vorher  string
	)

	err := p.db.QueryRowContext(ctx,
		`SELECT kennung, seit, vorher FROM peeridentitaeten WHERE profil_id = ?`,
		profilID).Scan(&kennung, &seit, &vorher)

	switch {
	case err == sql.ErrNoRows:
		return domain.Peervermerk{}, nil
	case err != nil:
		return domain.Peervermerk{}, uebersetze(err, "die Peerkennung lesen")
	}

	gestempelt, fehler := time.Parse(time.RFC3339Nano, seit)
	if fehler != nil {
		return domain.Peervermerk{}, domain.MitGrund(domain.ErrValidation,
			fehler, "den Zeitpunkt der Peerkennung lesen")
	}

	return domain.Peervermerk{
		Kennung: kennung, Seit: gestempelt, Vorher: vorher,
	}, nil
}

// Merken schreibt die Kennung dieses Profils — **ohne** eine Vorgängerin.
//
// **Für den ersten Schlüssel.** Wer eine bestehende Identität ablöst, nimmt
// [Peerablage.Rotieren]; der Unterschied ist nicht die Zeile, sondern die
// Zusage: hier wird nichts abgelöst, und deshalb kennt danach auch niemand
// eine Vorgängerin.
func (p *Peerablage) Merken(
	ctx context.Context, kennung string, jetzt time.Time,
) error {
	return p.schreibe(ctx, kennung, jetzt, false)
}

// Rotieren schreibt eine neue Kennung und hebt die alte nach `vorher`.
//
// **Ohne Vorgängerin ist es ein erstes Merken** und kein Fehler: wer einen
// Schlüssel verloren hat und einen neuen anlegt, rotiert — und es gibt
// nichts, was abgelöst würde.
func (p *Peerablage) Rotieren(
	ctx context.Context, kennung string, jetzt time.Time,
) error {
	return p.schreibe(ctx, kennung, jetzt, true)
}

// schreibe setzt die Kennung — und hebt die alte, wenn rotiert wird.
//
// **`excluded` und die alte Zeile in einem Ausdruck**: `ON CONFLICT DO UPDATE`
// sieht beide, und damit braucht die Rotation kein Lesen davor. Zwei Rufe
// hätten dazwischen einen Stand, in dem die alte Kennung fort ist und die
// neue noch nicht steht.
func (p *Peerablage) schreibe(
	ctx context.Context, kennung string, jetzt time.Time, rotiert bool,
) error {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return fehler
	}

	if kennung == "" {
		return domain.MitGrund(domain.ErrValidation, nil,
			"eine leere Peerkennung merken")
	}

	vorher := "''"
	if rotiert {
		vorher = "peeridentitaeten.kennung"
	}

	_, err := p.db.ExecContext(ctx, `
INSERT INTO peeridentitaeten (profil_id, kennung, seit, vorher)
VALUES (?, ?, ?, '')
ON CONFLICT (profil_id) DO UPDATE
   SET kennung = excluded.kennung,
       seit    = excluded.seit,
       vorher  = `+vorher, profilID, kennung, jetzt.Format(time.RFC3339Nano))
	if err != nil {
		return uebersetze(err, "die Peerkennung merken")
	}

	return nil
}
