package app

// **Womit sich dieses Programm ausweist** — Werkstück N1 von Phase 5.4
// (#1454).
//
// **Ein Composition Root und keine Fachlogik** (§2): hier werden Bestand und
// Schlüsselbund an die Schnittstellen gehängt, die `service` aufgestellt hat.
// Welcher der vier Stände gilt, entscheidet [service.Peerwerk] — und der
// Grund, warum das dort und nicht hier steht, ist der Satz des Issues: „nie
// still eine fremde Identität übernehmen".

import (
	"context"
	"database/sql"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Peerbund ist der Schlüsselbund, in dem die eigene Identität liegt.
//
// **Dieselbe Gestalt wie [Geheimnisleser] daneben, und trotzdem eine eigene**:
// dort wird nur gelesen, hier auch geschrieben — und ein Leser, der
// schriebe, wäre ein Name, der weniger verspricht als sein Rumpf hält (§10).
type Peerbund = service.Peergeheimnisse

// peervermerknaht hängt `internal/storage` an [service.Peervermerke].
//
// **Die Verbindung wird geholt und nicht festgehalten** — derselbe Grund wie
// bei [chiffrenaht]: ein Profilwechsel tauscht sie.
type peervermerknaht struct{ holen func() *sql.DB }

func (p peervermerknaht) Lesen(ctx context.Context) (domain.Peervermerk, error) {
	return storage.NeuePeerablage(p.holen()).Lesen(ctx)
}

func (p peervermerknaht) Merken(
	ctx context.Context, kennung string, jetzt time.Time,
) error {
	return storage.NeuePeerablage(p.holen()).Merken(ctx, kennung, jetzt)
}

func (p peervermerknaht) Rotieren(
	ctx context.Context, kennung string, jetzt time.Time,
) error {
	return storage.NeuePeerablage(p.holen()).Rotieren(ctx, kennung, jetzt)
}

// bundFuerDenPeer gibt den Schlüsselbund — den eingesetzten oder den echten.
func (a *App) bundFuerDenPeer() Peerbund {
	if a.peerbund != nil {
		return a.peerbund
	}

	// **`Anmeldebund` und nicht `NeuerBund("")`** — er achtet auf
	// `MSE_SCHLUESSELBUND`. Ohne diese Zeile schriebe **jeder** Lauf in den
	// Anmeldebund dieses Rechners, und das ist der Fehler, den
	// `internal/schluesselbund` in seinem Kopf benennt.
	return peerbundnaht{bund: schluesselbund.Anmeldebund()}
}

// MitPeerbund setzt den Schlüsselbund — für Läufe ohne echten (§3).
//
// **Ohne ihn schriebe jeder Lauf in den Anmeldebund dieses Rechners**, und
// das ist genau der Fehler, den `internal/schluesselbund` in seinem Kopf
// benennt.
func (a *App) MitPeerbund(bund Peerbund) *App {
	a.peerbund = bund

	return a
}

// peerbundnaht legt den Dienstnamen vor den Ruf.
//
// **Der Dienst steht in `service` und nicht hier** ([service.Peerdienst]):
// unter welchem Namen die eigene Identität im Bund liegt, ist eine Festlegung
// dieser Anwendung und keine des Betriebssystems.
type peerbundnaht struct{ bund *schluesselbund.Bund }

func (p peerbundnaht) Lies(ctx context.Context, konto string) (string, error) {
	return p.bund.Lies(ctx, service.Peerdienst, konto)
}

func (p peerbundnaht) Schreibe(ctx context.Context, konto, geheimnis string) error {
	return p.bund.Schreibe(ctx, service.Peerdienst, konto, geheimnis)
}

// peerwerk baut das Werk für diesen Ruf.
//
// **Je Ruf und nicht einmal**: Bestandsverbindung und Schlüsselbund können
// sich zwischen zwei Rufen ändern, und ein festgehaltenes Werk hielte einen
// Griff auf beides.
func (a *App) peerwerk() *service.Peerwerk {
	return service.NeuesPeerwerk(
		peervermerknaht{holen: a.laufendeVerbindung}, a.bundFuerDenPeer(),
		a.kontaktwerk(), a.uhr)
}
