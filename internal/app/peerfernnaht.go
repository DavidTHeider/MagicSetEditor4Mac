package app

// **Die Naht zum Fernhorcher** — Werkstück N3 von Phase 5.4
// (#1456).
//
// # Warum eine eigene Datei
//
// **Weil sie ohne ein DERP-Relais im Internet nicht auszuführen ist.** Was
// hier steht, ist Verdrahtung: der Ruf an `fernleitung.Horche` und das
// Durchreichen seiner Methoden. Jede Entscheidung — der Lebenszyklus, der
// Konflikt beim zweiten Start, der Stand — steht in `peerfern.go` und wird
// gemessen.
//
// Der Eintrag in `.covignore` nennt genau das. Dieselbe Gestalt wie
// `internal/fernleitung/tailcatnaht.go` und `fensternaht.go`: die Naht ist
// eine Veränderliche, **damit** alles darüber ohne sie läuft.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/fernleitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// fernhorcher gibt die eingesetzte Naht — oder die echte.
func (a *App) fernhorcher() func(peer.Identitaet) (Fernhorcher, error) {
	if a.fernlage.horchen != nil {
		return a.fernlage.horchen
	}

	return func(ich peer.Identitaet) (Fernhorcher, error) {
		fernleitung.MitProxywahl(a.fernproxywahl())

		horcher, fehler := fernleitung.Horche(ich)
		if fehler != nil {
			return nil, fehler
		}

		return fernkiste{innen: horcher}, nil
	}
}

// fernkiste hängt `internal/fernleitung` an [Fernhorcher].
//
// **Dieselben zwei Zeilen wie `horcherkiste`, und aus demselben Grund** (§2):
// der Rückgabetyp ist ein `*leitung.Leitung`, die Schnittstelle will eine
// `service.Verbindung`, und keines der beiden Pakete darf das andere kennen.
type fernkiste struct{ innen *fernleitung.Horcher }

func (f fernkiste) Nimm(ctx context.Context) (service.Verbindung, error) {
	return f.innen.Nimm(ctx)
}

func (f fernkiste) Schliesse() error { return f.innen.Schliesse() }

func (f fernkiste) Adresse() string { return f.innen.Adresse() }

func (f fernkiste) Einladung() domain.Einladung { return f.innen.Einladung() }

func (f fernkiste) Wege() []fernleitung.Fernweg { return f.innen.Wege() }

func (f fernkiste) Angebunden() int { return f.innen.Angebunden() }

// MitFernwaehler setzt die Naht zum Wählen über ein fremdes Netz — für Läufe.
func (a *App) MitFernwaehler(
	statt func(
		ctx context.Context, einladung domain.Einladung,
		ich peer.Identitaet, erwartet peer.Kennung,
	) (service.Verbindung, error),
) *App {
	a.fernlage.waehlen = statt

	return a
}

// fernwaehler gibt die eingesetzte Naht — oder die echte.
func (a *App) fernwaehler() func(
	context.Context, domain.Einladung, peer.Identitaet, peer.Kennung,
) (service.Verbindung, error) {
	if a.fernlage.waehlen != nil {
		return a.fernlage.waehlen
	}

	return func(
		ctx context.Context, einladung domain.Einladung,
		ich peer.Identitaet, erwartet peer.Kennung,
	) (service.Verbindung, error) {
		fernleitung.MitProxywahl(a.fernproxywahl())

		return fernleitung.Waehle(ctx, einladung, ich, erwartet)
	}
}
