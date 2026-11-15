// **Die Verschlüsselung, wie die Schichten sie sehen** — Werkstück 14 von
// Phase 4.12 (#1351),
// nach [0072](../../docs/entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md).
//
// **Ein Composition Root und keine Fachlogik** (§2): hier wird `storage` an
// die Schnittstelle gehängt, die `service` aufgestellt hat. Ein `if` über
// einen Domänenzustand steht hier nicht.

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Bestandschiffre ist die Verschlüsselung des Bestands.
//
// **Die Schnittstelle steht im Dienst, weil der Anwendungsfall sie braucht**
// (§2) — hier steht nur der Name, unter dem der Composition Root sie setzt.
type Bestandschiffre = service.Bestandschiffre

// chiffrenaht hängt `internal/storage` an diese Schnittstelle.
//
// # Die Verbindung wird **geholt** und nicht festgehalten
//
// **Weil sie wechselt.** Ein Tausch schließt sie und öffnet eine neue; wer sie
// beim Bauen einsteckt, hält danach einen Griff auf eine Datei, die es nicht
// mehr gibt. Beim Verschlüsseln fällt das erst beim **zweiten** Umschalten in
// einem Prozess auf — an- und wieder ausschalten —, und dann als „sql:
// database is closed" an einer Stelle, die damit nichts zu tun hat.
type chiffrenaht struct{ holen func() *sql.DB }

func (c chiffrenaht) IstVerschluesselt(pfad string) (bool, error) {
	return storage.IstVerschluesselt(pfad)
}

func (c chiffrenaht) NeuerSchluessel() string {
	return storage.NeuerSchluessel()
}

func (c chiffrenaht) Verschluesseln(
	ctx context.Context, ziel, schluessel string,
) error {
	return storage.Verschluesseln(ctx, c.holen(), ziel, schluessel)
}

func (c chiffrenaht) Entschluesseln(ctx context.Context, ziel string) error {
	return storage.Entschluesseln(ctx, c.holen(), ziel)
}

func (c chiffrenaht) SchluesselSchreiben(
	ctx context.Context, schluessel string,
) error {
	return storage.SchluesselSchreiben(ctx, schluessel)
}

func (c chiffrenaht) SchluesselLoeschen(ctx context.Context) error {
	return storage.SchluesselLoeschen(ctx)
}

// laufendeVerbindung gibt die Verbindung, die das Fenster gerade hält.
//
// **Eine Methode und kein Abschluss über die lokale Variable**, damit sie den
// Wechsel wirklich sieht: `a.db` wird bei jedem `Starten` neu gesetzt.
func (a *App) laufendeVerbindung() *sql.DB { return a.db }
