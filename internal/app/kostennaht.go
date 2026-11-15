package app

// **Die Naht zum Kostenjournal** — Werkstück O12 von Phase 5.2
// (#1439).
//
// **Die Verbindung wird geholt und nicht gehalten** — derselbe Grund wie bei
// `pflegenaht`: ein Zurückspielen tauscht sie, und ein Journal, das danach in
// die alte Datei schriebe, buchte auf einen Bestand, den es nicht mehr gibt.

import (
	"context"
	"database/sql"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// kostenablage bringt das Journal an die Schnittstelle des Dienstes.
//
// **Ein Griff je Ruf**: `storage.Kostenjournal` hält eine Verbindung, und die
// Naht baut es deshalb bei jedem Aufruf neu — das ist ein Zeiger und keine
// Arbeit.
type kostenablage struct{ holen func() *sql.DB }

func (k kostenablage) journal() *storage.Kostenjournal {
	return storage.NeuesKostenjournal(k.holen())
}

func (k kostenablage) BudgetSetzen(
	ctx context.Context, profilID int64, konto string,
	zeitraum domain.Zeitraum, betrag domain.Betrag, jetzt time.Time,
) error {
	return k.journal().BudgetSetzen(ctx, profilID, konto, zeitraum, betrag, jetzt)
}

func (k kostenablage) Reservieren(
	ctx context.Context, posten domain.Kostenposten,
	zeitraum domain.Zeitraum, jetzt time.Time,
) (domain.Kostenposten, error) {
	return k.journal().Reservieren(ctx, posten, zeitraum, jetzt)
}

func (k kostenablage) Abschliessen(
	ctx context.Context, profilID int64, kennung string, betrag domain.Betrag, jetzt time.Time,
) (domain.Kostenposten, error) {
	return k.journal().Abschliessen(ctx, profilID, kennung, betrag, jetzt)
}

func (k kostenablage) Freigeben(
	ctx context.Context, profilID int64, kennung string, jetzt time.Time,
) (domain.Kostenposten, error) {
	return k.journal().Freigeben(ctx, profilID, kennung, jetzt)
}

func (k kostenablage) AlsUnklarMerken(
	ctx context.Context, profilID int64, kennung, grund string, jetzt time.Time,
) (domain.Kostenposten, error) {
	return k.journal().AlsUnklarMerken(ctx, profilID, kennung, grund, jetzt)
}

func (k kostenablage) Stand(
	ctx context.Context, profilID int64, konto string,
	zeitraum domain.Zeitraum, jetzt time.Time,
) (domain.Haushaltsstand, error) {
	return k.journal().Stand(ctx, profilID, konto, zeitraum, jetzt)
}

func (k kostenablage) Posten(
	ctx context.Context, profilID int64, konto string,
	zeitraum domain.Zeitraum, jetzt time.Time, hoechstens int,
) ([]domain.Kostenposten, error) {
	return k.journal().Posten(ctx, profilID, konto, zeitraum, jetzt, hoechstens)
}

// KostenwacheUeber baut den Kostenwächter.
func KostenwacheUeber(
	holen func() *sql.DB, wache service.Ruecksetzungswache, uhr func() time.Time,
) *service.Kostenwache {
	return service.NeueKostenwache(kostenablage{holen: holen}, wache, uhr)
}

// kostenVerdrahten hängt das Journal ins Fenster.
func (a *App) kostenVerdrahten() {
	a.MitKosten(KostenwacheUeber(a.laufendeVerbindung, a.autorisierung, a.uhr))
}
