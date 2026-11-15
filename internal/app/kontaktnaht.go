package app

// **Wen dieses Profil kennt** — Werkstück N1 von Phase 5.4
// (#1454).
//
// **Ein Composition Root und keine Fachlogik** (§2): hier wird der Bestand an
// die Schnittstelle gehängt, die `service` aufgestellt hat. Ob eine
// Bestätigung gilt, entscheidet [service.Kontaktwerk] — und warum dort, steht
// bei ihm.

import (
	"context"
	"database/sql"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// kontaktnaht hängt `internal/storage` an [service.Kontaktvermerke].
//
// **Die Verbindung wird geholt und nicht festgehalten** — derselbe Grund wie
// bei [peervermerknaht]: ein Profilwechsel tauscht sie.
type kontaktnaht struct{ holen func() *sql.DB }

func (k kontaktnaht) ablage() *storage.Kontaktablage {
	return storage.NeueKontaktablage(k.holen())
}

func (k kontaktnaht) Liste(
	ctx context.Context,
) ([]domain.Peerkontakt, error) {
	return k.ablage().Liste(ctx)
}

func (k kontaktnaht) Lesen(
	ctx context.Context, kennung string,
) (domain.Peerkontakt, error) {
	return k.ablage().Lesen(ctx, kennung)
}

func (k kontaktnaht) Begegnet(
	ctx context.Context, kennung, fingerabdruck, name string, jetzt time.Time,
) error {
	return k.ablage().Begegnet(ctx, kennung, fingerabdruck, name, jetzt)
}

func (k kontaktnaht) Stand(
	ctx context.Context, kennung string, stand domain.Kontaktstand,
) error {
	return k.ablage().Stand(ctx, kennung, stand)
}

func (k kontaktnaht) Erlauben(
	ctx context.Context, kennung string, zweck domain.Zweck, jetzt time.Time,
) error {
	return k.ablage().Erlauben(ctx, kennung, zweck, jetzt)
}

func (k kontaktnaht) Entziehen(
	ctx context.Context, kennung string, zweck domain.Zweck,
) error {
	return k.ablage().Entziehen(ctx, kennung, zweck)
}

func (k kontaktnaht) AdresseSetzen(
	ctx context.Context, kennung, adresse string,
) error {
	return k.ablage().AdresseSetzen(ctx, kennung, adresse)
}

func (k kontaktnaht) NotizSetzen(
	ctx context.Context, kennung, notiz string,
) error {
	return k.ablage().NotizSetzen(ctx, kennung, notiz)
}

func (k kontaktnaht) AllesZurueck(ctx context.Context) error {
	return k.ablage().AllesZurueck(ctx)
}

// kontaktwerk baut das Werk für diesen Ruf.
func (a *App) kontaktwerk() *service.Kontaktwerk {
	return service.NeuesKontaktwerk(
		kontaktnaht{holen: a.laufendeVerbindung}, a.uhr)
}

func (k kontaktnaht) Angenommen(ctx context.Context, nonce string) (bool, error) {
	return k.ablage().Angenommen(ctx, nonce)
}

func (k kontaktnaht) Aufnehmen(
	ctx context.Context, aufnahme domain.Kontaktaufnahme, jetzt time.Time,
) error {
	return k.ablage().Aufnehmen(ctx, aufnahme, jetzt)
}

// Ausgestellt reicht an die Ablage (NET-01b).
func (k kontaktnaht) Ausgestellt(
	ctx context.Context, nonce string, zweck domain.Zweck, ablauf, jetzt time.Time,
) error {
	return k.ablage().Ausgestellt(ctx, nonce, zweck, ablauf, jetzt)
}

// Einloesen reicht an die Ablage (NET-01b).
func (k kontaktnaht) Einloesen(
	ctx context.Context, nonce string, kontakt domain.Peerkontakt, jetzt time.Time,
) error {
	return k.ablage().Einloesen(ctx, nonce, kontakt, jetzt)
}
